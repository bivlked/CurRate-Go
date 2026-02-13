package converter

import (
	"context"
	"errors"
	"math"
	"testing"
	"time"

	"github.com/bivlked/currate-go/internal/models"
)

// Моки для тестирования

// MockRateProvider - мок для RateProvider
type MockRateProvider struct {
	rateData  *models.RateData
	err       error
	callCount int
}

func (m *MockRateProvider) FetchRates(_ context.Context, date time.Time) (*models.RateData, error) {
	m.callCount++
	if m.err != nil {
		return nil, m.err
	}
	return m.rateData, nil
}

// MockCacheStorage - мок для CacheStorage
type MockCacheStorage struct {
	data map[string]struct {
		rate       float64
		actualDate time.Time
	}
}

func NewMockCache() *MockCacheStorage {
	return &MockCacheStorage{
		data: make(map[string]struct {
			rate       float64
			actualDate time.Time
		}),
	}
}

func (m *MockCacheStorage) Get(currency models.Currency, date time.Time) (float64, time.Time, bool) {
	key := string(currency) + ":" + date.Format("2006-01-02")
	entry, exists := m.data[key]
	if !exists {
		return 0, time.Time{}, false
	}
	return entry.rate, entry.actualDate, true
}

func (m *MockCacheStorage) Set(currency models.Currency, requestedDate time.Time, rate float64, actualDate time.Time) {
	key := string(currency) + ":" + requestedDate.Format("2006-01-02")
	m.data[key] = struct {
		rate       float64
		actualDate time.Time
	}{
		rate:       rate,
		actualDate: actualDate,
	}
}

func (m *MockCacheStorage) Clear() {
	m.data = make(map[string]struct {
		rate       float64
		actualDate time.Time
	})
}

// Тесты для проблемы 4: Rate date - использование фактической даты из XML

// TestConverter_Convert_UsesActualDateFromXML проверяет, что Convert использует фактическую дату из XML
func TestConverter_Convert_UsesActualDateFromXML(t *testing.T) {
	// Запрошенная дата - понедельник
	requestedDate := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC) // Понедельник

	// Фактическая дата из XML - предыдущий рабочий день (пятница)
	actualDateFromXML := time.Date(2024, 1, 12, 0, 0, 0, 0, time.UTC) // Пятница

	mockProvider := &MockRateProvider{
		rateData: &models.RateData{
			Date: actualDateFromXML, // Фактическая дата из XML (предыдущий рабочий день)
			Rates: map[models.Currency]models.ExchangeRate{
				models.USD: {
					Currency: models.USD,
					Rate:     80.0,
					Nominal:  1,
					Date:     actualDateFromXML,
				},
			},
		},
	}

	cache := NewMockCache()
	converter := NewConverter(mockProvider, cache)

	result, err := converter.Convert(context.Background(), 100, models.USD, requestedDate)
	if err != nil {
		t.Fatalf("Convert() error = %v, want nil", err)
	}

	if result == nil {
		t.Fatal("Convert() returned nil result")
	}

	// Проверяем, что используется фактическая дата из XML, а не запрошенная
	if !result.Date.Equal(actualDateFromXML) {
		t.Errorf("Convert() result.Date = %v, want %v (actual date from XML)", result.Date, actualDateFromXML)
	}

	if result.Date.Equal(requestedDate) {
		t.Errorf("Convert() result.Date should not equal requestedDate %v, but got %v", requestedDate, result.Date)
	}
}

// TestConverter_Convert_UsesActualDateFromCache проверяет, что Convert использует фактическую дату из кэша
func TestConverter_Convert_UsesActualDateFromCache(t *testing.T) {
	requestedDate := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)
	actualDateFromXML := time.Date(2024, 1, 12, 0, 0, 0, 0, time.UTC)

	mockProvider := &MockRateProvider{
		rateData: &models.RateData{
			Date: actualDateFromXML,
			Rates: map[models.Currency]models.ExchangeRate{
				models.USD: {
					Currency: models.USD,
					Rate:     80.0,
					Nominal:  1,
					Date:     actualDateFromXML,
				},
			},
		},
	}

	cache := NewMockCache()
	converter := NewConverter(mockProvider, cache)

	// Первый вызов - получаем из provider и кэшируем
	result1, err := converter.Convert(context.Background(), 100, models.USD, requestedDate)
	if err != nil {
		t.Fatalf("First Convert() error = %v, want nil", err)
	}

	if !result1.Date.Equal(actualDateFromXML) {
		t.Errorf("First Convert() result.Date = %v, want %v", result1.Date, actualDateFromXML)
	}

	// Второй вызов - должен использовать кэш с фактической датой
	result2, err := converter.Convert(context.Background(), 200, models.USD, requestedDate)
	if err != nil {
		t.Fatalf("Second Convert() error = %v, want nil", err)
	}

	// Проверяем, что используется фактическая дата из кэша
	if !result2.Date.Equal(actualDateFromXML) {
		t.Errorf("Second Convert() result.Date = %v, want %v (from cache)", result2.Date, actualDateFromXML)
	}

	// Проверяем, что provider был вызван только один раз (второй раз использован кэш)
	if mockProvider.callCount != 1 {
		t.Errorf("MockRateProvider.FetchRates() callCount = %d, want 1 (second call should use cache)", mockProvider.callCount)
	}
}

// TestConverter_GetRateInternal_UsesActualDate проверяет, что getRateInternal возвращает фактическую дату
func TestConverter_GetRateInternal_UsesActualDate(t *testing.T) {
	requestedDate := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)
	actualDateFromXML := time.Date(2024, 1, 12, 0, 0, 0, 0, time.UTC)

	mockProvider := &MockRateProvider{
		rateData: &models.RateData{
			Date: actualDateFromXML,
			Rates: map[models.Currency]models.ExchangeRate{
				models.USD: {
					Currency: models.USD,
					Rate:     80.0,
					Nominal:  1,
					Date:     actualDateFromXML,
				},
			},
		},
	}

	cache := NewMockCache()
	converter := NewConverter(mockProvider, cache)

	// Используем рефлексию или тестируем через публичный метод GetRate
	// Но getRateInternal - приватный метод, поэтому тестируем через Convert
	// и проверяем, что actualDate используется корректно

	result, err := converter.Convert(context.Background(), 100, models.USD, requestedDate)
	if err != nil {
		t.Fatalf("Convert() error = %v, want nil", err)
	}

	// Проверяем, что дата в результате - фактическая дата из XML
	if !result.Date.Equal(actualDateFromXML) {
		t.Errorf("Convert() result.Date = %v, want %v (actual date from XML)", result.Date, actualDateFromXML)
	}
}

// Тесты для validator.go

func TestValidateAmount(t *testing.T) {
	tests := []struct {
		name    string
		amount  float64
		wantErr bool
	}{
		{
			name:    "Положительная сумма",
			amount:  1000.0,
			wantErr: false,
		},
		{
			name:    "Маленькая положительная сумма",
			amount:  0.01,
			wantErr: false,
		},
		{
			name:    "Большая сумма",
			amount:  1000000.0,
			wantErr: false,
		},
		{
			name:    "Ноль - ошибка",
			amount:  0.0,
			wantErr: true,
		},
		{
			name:    "Отрицательная сумма - ошибка",
			amount:  -100.0,
			wantErr: true,
		},
		{
			name:    "Очень маленькая отрицательная - ошибка",
			amount:  -0.01,
			wantErr: true,
		},
		{
			name:    "NaN - ошибка",
			amount:  math.NaN(),
			wantErr: true,
		},
		{
			name:    "Положительная бесконечность - ошибка",
			amount:  math.Inf(1),
			wantErr: true,
		},
		{
			name:    "Отрицательная бесконечность - ошибка",
			amount:  math.Inf(-1),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAmount(tt.amount)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateAmount() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr && !errors.Is(err, ErrInvalidAmount) {
				t.Errorf("ValidateAmount() ожидалась ошибка ErrInvalidAmount, получена %v", err)
			}
		})
	}
}

func TestValidateDate(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name    string
		date    time.Time
		wantErr bool
	}{
		{
			name:    "Прошлая дата",
			date:    now.AddDate(0, 0, -10),
			wantErr: false,
		},
		{
			name:    "Дата год назад",
			date:    now.AddDate(-1, 0, 0),
			wantErr: false,
		},
		{
			name:    "Текущая дата (приблизительно)",
			date:    now,
			wantErr: false,
		},
		{
			name:    "Дата в будущем - ошибка",
			date:    now.AddDate(0, 0, 1),
			wantErr: true,
		},
		{
			name:    "Дата через год - ошибка",
			date:    now.AddDate(1, 0, 0),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDate(tt.date)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateDate() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr && !errors.Is(err, ErrDateInFuture) {
				t.Errorf("ValidateDate() ожидалась ошибка ErrDateInFuture, получена %v", err)
			}
		})
	}
}

func TestValidateDateUTCInputWithNonUTCLocal(t *testing.T) {
	originalLocal := time.Local
	time.Local = time.FixedZone("UTC+10", 10*60*60)
	t.Cleanup(func() {
		time.Local = originalLocal
	})

	localNow := time.Now().In(time.Local)
	date := time.Date(localNow.Year(), localNow.Month(), localNow.Day(), 0, 0, 0, 0, time.UTC)

	if err := ValidateDate(date); err != nil {
		t.Fatalf("ValidateDate() неожиданная ошибка для даты в UTC при локальной зоне, отличной от UTC: %v", err)
	}
}

// Тесты для formatter.go

func TestAddThousandsSeparator(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Без разделителей (до 3 цифр)",
			input:    "500",
			expected: "500",
		},
		{
			name:     "Одна тысяча",
			input:    "1000",
			expected: "1 000",
		},
		{
			name:     "Несколько тысяч",
			input:    "80722",
			expected: "80 722",
		},
		{
			name:     "Миллион",
			input:    "1000000",
			expected: "1 000 000",
		},
		{
			name:     "Сложное число",
			input:    "123456789",
			expected: "123 456 789",
		},
		{
			name:     "Двузначное число",
			input:    "99",
			expected: "99",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := addThousandsSeparator(tt.input)
			if got != tt.expected {
				t.Errorf("addThousandsSeparator() = %s, expected %s", got, tt.expected)
			}
		})
	}
}

func TestFormatNumber(t *testing.T) {
	tests := []struct {
		name     string
		number   float64
		expected string
	}{
		{
			name:     "Целое число",
			number:   1000.0,
			expected: "1 000,00",
		},
		{
			name:     "Число с копейками",
			number:   1000.50,
			expected: "1 000,50",
		},
		{
			name:     "Большое число",
			number:   80722.0,
			expected: "80 722,00",
		},
		{
			name:     "Миллион",
			number:   1000000.0,
			expected: "1 000 000,00",
		},
		{
			name:     "Число с округлением",
			number:   999.999,
			expected: "1 000,00", // Округляется до 1000.00
		},
		{
			name:     "Маленькое число",
			number:   100.5,
			expected: "100,50",
		},
		{
			name:     "Ноль",
			number:   0.0,
			expected: "0,00",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatNumber(tt.number)
			if got != tt.expected {
				t.Errorf("formatNumber() = %s, expected %s", got, tt.expected)
			}
		})
	}
}

func TestFormatResult(t *testing.T) {
	tests := []struct {
		name      string
		amount    float64
		rate      float64
		currency  models.Currency
		resultRUB float64
		expected  string
	}{
		{
			name:      "USD конвертация",
			amount:    1000.0,
			rate:      80.7220,
			currency:  models.USD,
			resultRUB: 80722.0,
			expected:  "80 722,00 руб. ($1 000,00 по курсу 80,7220)",
		},
		{
			name:      "EUR конвертация",
			amount:    500.0,
			rate:      94.5120,
			currency:  models.EUR,
			resultRUB: 47256.0,
			expected:  "47 256,00 руб. (€500,00 по курсу 94,5120)",
		},
		{
			name:      "Маленькая сумма USD",
			amount:    100.0,
			rate:      80.0,
			currency:  models.USD,
			resultRUB: 8000.0,
			expected:  "8 000,00 руб. ($100,00 по курсу 80,0000)",
		},
		{
			name:      "Дробная сумма",
			amount:    100.50,
			rate:      80.0,
			currency:  models.USD,
			resultRUB: 8040.0,
			expected:  "8 040,00 руб. ($100,50 по курсу 80,0000)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatResult(tt.amount, tt.rate, tt.currency, tt.resultRUB)
			if got != tt.expected {
				t.Errorf("FormatResult() = %s, expected %s", got, tt.expected)
			}
		})
	}
}

// Тесты для converter.go

func TestNewConverter(t *testing.T) {
	provider := &MockRateProvider{}
	cache := NewMockCache()

	converter := NewConverter(provider, cache)

	if converter == nil {
		t.Fatal("Converter не должен быть nil")
	}

	if converter.provider != provider {
		t.Error("Provider не установлен корректно")
	}

	if converter.cache != cache {
		t.Error("Cache не установлен корректно")
	}
}

func TestNewConverter_NilCacheUsesNoop(t *testing.T) {
	provider := &MockRateProvider{}

	converter := NewConverter(provider, nil)

	if converter == nil {
		t.Fatal("Converter не должен быть nil")
	}

	if converter.cache == nil {
		t.Fatal("Cache не должен быть nil, ожидался noopCache")
	}

	// Проверяем, что noopCache работает корректно
	date := testPastDateUTC()

	// Get должен возвращать false
	rate, actualDate, found := converter.cache.Get(models.USD, date)
	if found {
		t.Error("noopCache.Get() должен возвращать false")
	}
	if rate != 0 {
		t.Errorf("noopCache.Get() должен возвращать rate=0, получено %v", rate)
	}
	if !actualDate.IsZero() {
		t.Errorf("noopCache.Get() должен возвращать нулевую дату, получено %v", actualDate)
	}

	// Set не должен вызывать ошибок
	converter.cache.Set(models.USD, date, 80.0, date)

	// После Set Get все равно должен возвращать false (noop cache)
	_, _, found = converter.cache.Get(models.USD, date)
	if found {
		t.Error("noopCache.Get() после Set должен все равно возвращать false")
	}

	// Clear не должен вызывать ошибок
	converter.cache.Clear()
}

func TestConverter_Convert_NilProvider(t *testing.T) {
	cache := NewMockCache()
	converter := NewConverter(nil, cache)

	_, err := converter.Convert(context.Background(), 100, models.USD, time.Now())
	if err == nil {
		t.Fatal("Ожидалась ошибка при отсутствии источника курсов")
	}

	if !errors.Is(err, ErrNilRateProvider) {
		t.Fatalf("Ожидалась ошибка ErrNilRateProvider, получено: %v", err)
	}
}

func TestConverter_Convert_Success(t *testing.T) {
	date := testPastDateUTC()

	// Настройка мок provider
	mockProvider := &MockRateProvider{
		rateData: &models.RateData{
			Date: date,
			Rates: map[models.Currency]models.ExchangeRate{
				models.USD: {
					Currency: models.USD,
					Rate:     80.7220,
					Nominal:  1,
					Date:     date,
				},
				models.EUR: {
					Currency: models.EUR,
					Rate:     94.5120,
					Nominal:  1,
					Date:     date,
				},
			},
		},
	}

	cache := NewMockCache()
	converter := NewConverter(mockProvider, cache)

	// Тест USD
	result, err := converter.Convert(context.Background(), 1000, models.USD, date)
	if err != nil {
		t.Fatalf("Неожиданная ошибка: %v", err)
	}

	if result == nil {
		t.Fatal("Result не должен быть nil")
	}

	if result.SourceAmount != 1000 {
		t.Errorf("SourceAmount: ожидалось 1000, получено %v", result.SourceAmount)
	}

	if result.SourceCurrency != models.USD {
		t.Errorf("SourceCurrency: ожидалось USD, получено %v", result.SourceCurrency)
	}

	if result.TargetCurrency != models.RUB {
		t.Errorf("TargetCurrency: ожидалось RUB, получено %v", result.TargetCurrency)
	}

	if result.Rate != 80.7220 {
		t.Errorf("Rate: ожидалось 80.7220, получено %v", result.Rate)
	}

	if result.TargetAmount != 80722.0 {
		t.Errorf("TargetAmount: ожидалось 80722.0, получено %v", result.TargetAmount)
	}

	expectedFormatted := "80 722,00 руб. ($1 000,00 по курсу 80,7220)"
	if result.FormattedStr != expectedFormatted {
		t.Errorf("FormattedStr: ожидалось %s, получено %s", expectedFormatted, result.FormattedStr)
	}

	// Проверяем что курс сохранен в кэше
	cachedRate, _, found := cache.Get(models.USD, date)
	if !found {
		t.Error("Курс должен быть сохранен в кэше")
	}
	if cachedRate != 80.7220 {
		t.Errorf("Cached rate: ожидалось 80.7220, получено %v", cachedRate)
	}
}

func TestConverter_Convert_CacheHit(t *testing.T) {
	date := testPastDateUTC()

	// Provider который не должен быть вызван
	mockProvider := &MockRateProvider{
		err: errors.New("provider should not be called"),
	}

	cache := NewMockCache()
	// Предварительно заполняем кэш
	cache.Set(models.USD, date, 85.0, date) // requestedDate и actualDate одинаковы

	converter := NewConverter(mockProvider, cache)

	result, err := converter.Convert(context.Background(), 1000, models.USD, date)
	if err != nil {
		t.Fatalf("Неожиданная ошибка: %v", err)
	}

	// Проверяем что использован курс из кэша
	if result.Rate != 85.0 {
		t.Errorf("Rate: ожидалось 85.0 (из кэша), получено %v", result.Rate)
	}

	if result.TargetAmount != 85000.0 {
		t.Errorf("TargetAmount: ожидалось 85000.0, получено %v", result.TargetAmount)
	}
}

func TestConverter_Convert_RUB(t *testing.T) {
	date := testPastDateUTC()

	mockProvider := &MockRateProvider{
		err: errors.New("provider should not be called"),
	}

	cache := NewMockCache()
	converter := NewConverter(mockProvider, cache)

	result, err := converter.Convert(context.Background(), 1000, models.RUB, date)
	if err != nil {
		t.Fatalf("Неожиданная ошибка: %v", err)
	}

	if result.Rate != 1 {
		t.Errorf("Rate: ожидалось 1, получено %v", result.Rate)
	}

	if result.TargetAmount != 1000 {
		t.Errorf("TargetAmount: ожидалось 1000, получено %v", result.TargetAmount)
	}

	expectedFormatted := "1 000,00 руб. (₽1 000,00 по курсу 1,0000)"
	if result.FormattedStr != expectedFormatted {
		t.Errorf("FormattedStr: ожидалось %s, получено %s", expectedFormatted, result.FormattedStr)
	}
}

func TestConverter_Convert_ValidationErrors(t *testing.T) {
	date := testPastDateUTC()
	futureDate := time.Now().AddDate(0, 0, 1)

	mockProvider := &MockRateProvider{}
	cache := NewMockCache()
	converter := NewConverter(mockProvider, cache)

	tests := []struct {
		name     string
		amount   float64
		currency models.Currency
		date     time.Time
		wantErr  error
	}{
		{
			name:     "Отрицательная сумма",
			amount:   -100,
			currency: models.USD,
			date:     date,
			wantErr:  ErrInvalidAmount,
		},
		{
			name:     "Нулевая сумма",
			amount:   0,
			currency: models.USD,
			date:     date,
			wantErr:  ErrInvalidAmount,
		},
		{
			name:     "Неподдерживаемая валюта",
			amount:   1000,
			currency: models.Currency("GBP"),
			date:     date,
			wantErr:  models.ErrUnsupportedCurrency,
		},
		{
			name:     "Дата в будущем",
			amount:   1000,
			currency: models.USD,
			date:     futureDate,
			wantErr:  ErrDateInFuture,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := converter.Convert(context.Background(), tt.amount, tt.currency, tt.date)
			if err == nil {
				t.Fatal("Ожидалась ошибка, но её нет")
			}

			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Ожидалась ошибка %v, получена %v", tt.wantErr, err)
			}
		})
	}
}

func TestConverter_Convert_ProviderError(t *testing.T) {
	date := testPastDateUTC()

	// Provider который возвращает ошибку
	mockProvider := &MockRateProvider{
		err: errors.New("network error"),
	}

	cache := NewMockCache()
	converter := NewConverter(mockProvider, cache)

	_, err := converter.Convert(context.Background(), 1000, models.USD, date)
	if err == nil {
		t.Fatal("Ожидалась ошибка от provider")
	}

	// Проверяем что ошибка обернута
	if !errors.Is(err, mockProvider.err) {
		t.Errorf("Ошибка должна содержать ошибку provider: %v", err)
	}
}

func TestConverter_Convert_CurrencyNotFound(t *testing.T) {
	date := testPastDateUTC()

	// Provider возвращает данные, но без USD
	mockProvider := &MockRateProvider{
		rateData: &models.RateData{
			Date: date,
			Rates: map[models.Currency]models.ExchangeRate{
				models.EUR: {
					Currency: models.EUR,
					Rate:     94.5120,
					Nominal:  1,
					Date:     date,
				},
			},
		},
	}

	cache := NewMockCache()
	converter := NewConverter(mockProvider, cache)

	_, err := converter.Convert(context.Background(), 1000, models.USD, date)
	if err == nil {
		t.Fatal("Ожидалась ошибка 'currency not found'")
	}
}

func TestConverter_Convert_MultipleConversions(t *testing.T) {
	date := testPastDateUTC()

	mockProvider := &MockRateProvider{
		rateData: &models.RateData{
			Date: date,
			Rates: map[models.Currency]models.ExchangeRate{
				models.USD: {Currency: models.USD, Rate: 80.0, Nominal: 1, Date: date},
				models.EUR: {Currency: models.EUR, Rate: 90.0, Nominal: 1, Date: date},
			},
		},
	}

	cache := NewMockCache()
	converter := NewConverter(mockProvider, cache)

	// Первая конвертация USD
	result1, err := converter.Convert(context.Background(), 1000, models.USD, date)
	if err != nil {
		t.Fatalf("Ошибка первой конвертации: %v", err)
	}
	if result1.TargetAmount != 80000.0 {
		t.Errorf("USD результат: ожидалось 80000, получено %v", result1.TargetAmount)
	}

	// Вторая конвертация EUR
	result2, err := converter.Convert(context.Background(), 500, models.EUR, date)
	if err != nil {
		t.Fatalf("Ошибка второй конвертации: %v", err)
	}
	if result2.TargetAmount != 45000.0 {
		t.Errorf("EUR результат: ожидалось 45000, получено %v", result2.TargetAmount)
	}

	// Третья конвертация USD (должна использовать кэш)
	result3, err := converter.Convert(context.Background(), 2000, models.USD, date)
	if err != nil {
		t.Fatalf("Ошибка третьей конвертации: %v", err)
	}
	if result3.TargetAmount != 160000.0 {
		t.Errorf("USD результат: ожидалось 160000, получено %v", result3.TargetAmount)
	}
}

func TestConverter_Convert_NormalizesNominalRate(t *testing.T) {
	date := testPastDateUTC()

	mockProvider := &MockRateProvider{
		rateData: &models.RateData{
			Date: date,
			Rates: map[models.Currency]models.ExchangeRate{
				models.USD: {
					Currency: models.USD,
					Rate:     100.0,
					Nominal:  10,
					Date:     date,
				},
			},
		},
	}

	cache := NewMockCache()
	converter := NewConverter(mockProvider, cache)

	result, err := converter.Convert(context.Background(), 10, models.USD, date)
	if err != nil {
		t.Fatalf("Неожиданная ошибка: %v", err)
	}

	if result.Rate != 10.0 {
		t.Errorf("Rate: ожидалось 10.0 (100/10), получено %v", result.Rate)
	}

	if result.TargetAmount != 100.0 {
		t.Errorf("TargetAmount: ожидалось 100.0, получено %v", result.TargetAmount)
	}

	cachedRate, _, found := cache.Get(models.USD, date)
	if !found {
		t.Fatal("Ожидался сохраненный курс в кэше")
	}
	if cachedRate != 10.0 {
		t.Errorf("Cached rate: ожидалось 10.0, получено %v", cachedRate)
	}
}

func TestConverter_Convert_NormalizesDateForCache(t *testing.T) {
	baseDate := testPastDateUTC()
	date := time.Date(baseDate.Year(), baseDate.Month(), baseDate.Day(), 15, 30, 0, 0, time.UTC)
	sameDayLater := time.Date(baseDate.Year(), baseDate.Month(), baseDate.Day(), 22, 45, 0, 0, time.UTC)

	mockProvider := &MockRateProvider{
		rateData: &models.RateData{
			Date: date,
			Rates: map[models.Currency]models.ExchangeRate{
				models.USD: {
					Currency: models.USD,
					Rate:     80.0,
					Nominal:  1,
					Date:     date,
				},
			},
		},
	}

	cache := NewMockCache()
	converter := NewConverter(mockProvider, cache)

	_, err := converter.Convert(context.Background(), 100, models.USD, date)
	if err != nil {
		t.Fatalf("Неожиданная ошибка первой конвертации: %v", err)
	}

	_, err = converter.Convert(context.Background(), 200, models.USD, sameDayLater)
	if err != nil {
		t.Fatalf("Неожиданная ошибка второй конвертации: %v", err)
	}

	if mockProvider.callCount != 1 {
		t.Errorf("Ожидался один вызов провайдера благодаря нормализации даты, получено %d", mockProvider.callCount)
	}
}

func TestConverter_GetRate_NilProvider(t *testing.T) {
	cache := NewMockCache()
	converter := NewConverter(nil, cache)

	_, err := converter.GetRate(context.Background(), models.USD, time.Now())
	if err == nil {
		t.Fatal("Ожидалась ошибка при отсутствии источника курсов")
	}

	if !errors.Is(err, ErrNilRateProvider) {
		t.Fatalf("Ожидалась ошибка ErrNilRateProvider, получено: %v", err)
	}
}
