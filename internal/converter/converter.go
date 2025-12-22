package converter

import (
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
	Get(currency models.Currency, date time.Time) (float64, bool)

	// Set сохраняет курс в кэш
	Set(currency models.Currency, date time.Time, rate float64)

	// Clear очищает весь кэш
	Clear()
}

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

	if currency == models.RUB {
		resultRUB := amount
		formatted := FormatResult(amount, 1, currency, resultRUB)

		return &models.ConversionResult{
			SourceCurrency: currency,
			TargetCurrency: models.RUB,
			SourceAmount:   amount,
			TargetAmount:   resultRUB,
			Rate:           1,
			Date:           normalizedDate,
			FormattedStr:   formatted,
		}, nil
	}

	// Получение курса (сначала проверяем кэш)
	rate, found := c.cache.Get(currency, normalizedDate)
	if !found {
		// Курса нет в кэше - получаем через provider
		rateData, err := c.provider.FetchRates(normalizedDate)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch rates: %w", err)
		}

		// Извлекаем курс для нужной валюты
		exchangeRate, exists := rateData.Rates[currency]
		if !exists {
			return nil, fmt.Errorf("currency %s not found in rates", currency)
		}

		rate = exchangeRate.Rate
		if exchangeRate.Nominal > 1 {
			rate = rate / float64(exchangeRate.Nominal)
		}

		// Сохраняем в кэш
		c.cache.Set(currency, normalizedDate, rate)
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
		Date:           normalizedDate,
		FormattedStr:   formatted,
	}, nil
}
