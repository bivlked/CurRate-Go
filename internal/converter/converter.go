package converter

import (
	"errors"
	"fmt"
	"time"

	"github.com/bivlked/currate-go/internal/models"
)

// RateProvider - интерфейс для получения курсов валют
// Позволяет использовать моки для тестирования
type RateProvider interface {
	// FetchRates получает курсы валют на указанную дату
	// Возвращает RateData с картой курсов или ошибку
	FetchRates(date time.Time) (*models.RateData, error)
}

// CacheStorage - интерфейс для кэширования курсов
// Позволяет использовать моки для тестирования
type CacheStorage interface {
	// Get получает курс из кэша
	// Возвращает (rate, actualDate, found), где actualDate - фактическая дата курса из XML
	Get(currency models.Currency, date time.Time) (float64, time.Time, bool)

	// Set сохраняет курс в кэш
	// requestedDate - запрошенная дата (используется как ключ кэша)
	// actualDate - фактическая дата курса из XML (сохраняется в Entry)
	Set(currency models.Currency, requestedDate time.Time, rate float64, actualDate time.Time)

	// Clear очищает весь кэш
	Clear()
}

// Ошибки конвертера
var (
	ErrNilRateProvider = errors.New("источник курсов не задан")
)

// Converter - конвертер валют с кэшированием
type Converter struct {
	provider RateProvider
	cache    CacheStorage
}

// NewConverter создает новый конвертер валют
// provider - источник курсов валют (обычно parser.CBRParser)
// cache - хранилище кэша (обычно cache.LRUCache)
//
// Пример использования:
//
//	cacheInstance := cache.NewLRUCache(100, 24*time.Hour)
//	provider := parser.NewCBRParser(httpClient)
//	converter := converter.NewConverter(provider, cacheInstance)
func NewConverter(provider RateProvider, cache CacheStorage) *Converter {
	if cache == nil {
		cache = noopCache{}
	}

	return &Converter{
		provider: provider,
		cache:    cache,
	}
}

// Convert конвертирует сумму в указанной валюте в рубли на заданную дату
// amount - сумма для конвертации
// currency - валюта (USD, EUR)
// date - дата курса
//
// # Возвращает ConversionResult с отформатированным результатом или ошибку
//
// Алгоритм:
// 1. Валидация входных данных
// 2. Проверка кэша
// 3. Если нет в кэше - получение курса через provider
// 4. Конвертация и форматирование
//
// Пример использования:
//
//	date := time.Date(2025, 12, 20, 0, 0, 0, 0, time.UTC)
//	result, err := converter.Convert(1000, models.USD, date)
//	if err != nil {
//	    return err
//	}
//	fmt.Println(result.FormattedStr)
//	// Вывод: "80 722,00 руб. ($1 000,00 по курсу 80,7220)"
func (c *Converter) Convert(amount float64, currency models.Currency, date time.Time) (*models.ConversionResult, error) {
	if c.provider == nil {
		return nil, ErrNilRateProvider
	}

	normalizedDate := normalizeDate(date)

	// Валидация входных данных
	if err := ValidateAmount(amount); err != nil {
		return nil, err
	}

	if err := currency.Validate(); err != nil {
		return nil, err
	}

	if err := ValidateDate(normalizedDate); err != nil {
		return nil, err
	}

	// Получаем курс через внутренний метод (использует кэш и provider)
	rate, actualDate, err := c.getRateInternal(currency, normalizedDate)
	if err != nil {
		return nil, err
	}

	if currency == models.RUB {
		resultRUB := amount
		formatted := FormatResult(amount, 1, currency, resultRUB)

		return &models.ConversionResult{
			SourceCurrency: currency,
			TargetCurrency: models.RUB,
			SourceAmount:   amount,
			TargetAmount:   resultRUB,
			Rate:           1,
			Date:           normalizedDate, // Для RUB используем запрошенную дату
			FormattedStr:   formatted,
		}, nil
	}

	// Конвертация
	resultRUB := amount * rate

	// Форматирование
	formatted := FormatResult(amount, rate, currency, resultRUB)

	return &models.ConversionResult{
		SourceCurrency: currency,
		TargetCurrency: models.RUB,
		SourceAmount:   amount,
		TargetAmount:   resultRUB,
		Rate:           rate,
		Date:           actualDate, // Используем фактическую дату из XML
		FormattedStr:   formatted,
	}, nil
}

// getRateInternal получает курс валюты на указанную дату без форматирования
// Возвращает (rate, actualDate, error), где actualDate - фактическая дата из XML
// Это внутренний метод, который используется как в Convert, так и в GetRate
// для избежания дублирования кода
func (c *Converter) getRateInternal(currency models.Currency, normalizedDate time.Time) (float64, time.Time, error) {
	if c.provider == nil {
		return 0, time.Time{}, ErrNilRateProvider
	}

	// Для RUB всегда возвращаем 1.0
	if currency == models.RUB {
		return 1.0, normalizedDate, nil
	}

	// Получение курса (сначала проверяем кэш по запрошенной дате)
	// Ключ кэша - запрошенная дата, но в Entry хранится фактическая дата
	rate, actualDate, found := c.cache.Get(currency, normalizedDate)
	if !found {
		// Курса нет в кэше - получаем через provider
		rateData, err := c.provider.FetchRates(normalizedDate)
		if err != nil {
			return 0, time.Time{}, fmt.Errorf("failed to fetch rates: %w", err)
		}

		// Используем фактическую дату из XML
		actualDate = normalizeDate(rateData.Date)

		// Извлекаем курс для нужной валюты
		exchangeRate, exists := rateData.Rates[currency]
		if !exists {
			return 0, time.Time{}, fmt.Errorf("currency %s not found in rates", currency)
		}

		rate = exchangeRate.Rate
		if exchangeRate.Nominal > 1 {
			rate = rate / float64(exchangeRate.Nominal)
		}

		// Сохраняем в кэш дважды для максимальной эффективности:
		// 1) По запрошенной дате - чтобы последующие запросы на ту же дату попадали в кэш
		c.cache.Set(currency, normalizedDate, rate, actualDate)

		// 2) По фактической дате - чтобы избежать повторных сетевых запросов
		//    Например: запрос на воскресенье вернет пятницу, затем запрос на пятницу
		//    найдет данные в кэше без нового обращения к API ЦБ РФ
		if !actualDate.Equal(normalizedDate) {
			c.cache.Set(currency, actualDate, rate, actualDate)
		}

		return rate, actualDate, nil
	}

	// Курс найден в кэше - возвращаем с фактической датой из кэша
	return rate, actualDate, nil
}

// GetRate получает курс валюты на указанную дату без форматирования
// Используется для live preview в GUI, где не нужна полная конвертация
// Возвращает только числовой курс без создания ConversionResult
func (c *Converter) GetRate(currency models.Currency, date time.Time) (float64, error) {
	if c.provider == nil {
		return 0, ErrNilRateProvider
	}

	normalizedDate := normalizeDate(date)

	// Валидация входных данных
	if err := currency.Validate(); err != nil {
		return 0, err
	}

	if err := ValidateDate(normalizedDate); err != nil {
		return 0, err
	}

	// Используем внутренний метод для получения курса
	// Для live preview фактическая дата не нужна, поэтому игнорируем её
	rate, _, err := c.getRateInternal(currency, normalizedDate)
	return rate, err
}

type noopCache struct{}

func (noopCache) Get(currency models.Currency, date time.Time) (float64, time.Time, bool) {
	return 0, time.Time{}, false
}

func (noopCache) Set(currency models.Currency, requestedDate time.Time, rate float64, actualDate time.Time) {
}

func (noopCache) Clear() {}
