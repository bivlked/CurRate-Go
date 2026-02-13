package app

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/bivlked/currate-go/internal/converter"
	"github.com/bivlked/currate-go/internal/models"
)

// mockRateProvider - мок для RateProvider
type mockRateProvider struct {
	rateData *models.RateData
	err      error
}

func (m *mockRateProvider) FetchRates(_ context.Context, date time.Time) (*models.RateData, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.rateData, nil
}

// mockCacheStorage - мок для CacheStorage
type mockCacheStorage struct {
	data map[string]struct {
		rate       float64
		actualDate time.Time
	}
}

func newMockCache() *mockCacheStorage {
	return &mockCacheStorage{
		data: make(map[string]struct {
			rate       float64
			actualDate time.Time
		}),
	}
}

func (m *mockCacheStorage) Get(currency models.Currency, date time.Time) (float64, time.Time, bool) {
	key := string(currency) + ":" + date.Format("2006-01-02")
	entry, exists := m.data[key]
	if !exists {
		return 0, time.Time{}, false
	}
	return entry.rate, entry.actualDate, true
}

func (m *mockCacheStorage) Set(currency models.Currency, requestedDate time.Time, rate float64, actualDate time.Time) {
	key := string(currency) + ":" + requestedDate.Format("2006-01-02")
	m.data[key] = struct {
		rate       float64
		actualDate time.Time
	}{
		rate:       rate,
		actualDate: actualDate,
	}
}

func (m *mockCacheStorage) Clear() {
	m.data = make(map[string]struct {
		rate       float64
		actualDate time.Time
	})
}

// createTestConverter создает Converter с моками для тестирования
func createTestConverter(rateData *models.RateData, rateError error, cacheRate float64, cacheFound bool) *converter.Converter {
	mockProvider := &mockRateProvider{
		rateData: rateData,
		err:      rateError,
	}
	mockCache := newMockCache()
	if cacheFound && cacheRate > 0 {
		date := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
		mockCache.Set(models.USD, date, cacheRate, date)
	}
	return converter.NewConverter(mockProvider, mockCache)
}

func TestNewApp(t *testing.T) {
	date := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	rateData := &models.RateData{
		Date: date,
		Rates: map[models.Currency]models.ExchangeRate{
			models.USD: {
				Currency: models.USD,
				Rate:     80.0,
				Nominal:  1,
				Date:     date,
			},
		},
	}
	conv := createTestConverter(rateData, nil, 0, false)
	app := NewApp(conv)

	if app == nil {
		t.Fatal("NewApp() returned nil")
	}

	if app.converter == nil {
		t.Error("NewApp() converter is nil")
	}
}

func TestApp_Startup(t *testing.T) {
	date := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	rateData := &models.RateData{
		Date: date,
		Rates: map[models.Currency]models.ExchangeRate{
			models.USD: {
				Currency: models.USD,
				Rate:     80.0,
				Nominal:  1,
				Date:     date,
			},
		},
	}
	conv := createTestConverter(rateData, nil, 0, false)
	app := NewApp(conv)

	// Startup не должен паниковать
	ctx := context.Background()
	app.Startup(ctx)
}

func TestApp_Convert_Success(t *testing.T) {
	date := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	rateData := &models.RateData{
		Date: date,
		Rates: map[models.Currency]models.ExchangeRate{
			models.USD: {
				Currency: models.USD,
				Rate:     80.0,
				Nominal:  1,
				Date:     date,
			},
		},
	}
	conv := createTestConverter(rateData, nil, 0, false)
	app := NewApp(conv)
	app.Startup(context.Background())

	req := ConvertRequest{
		Amount:   100,
		Currency: "USD",
		Date:     "15.01.2024",
	}

	result := app.Convert(req)

	if !result.Success {
		t.Errorf("Convert() Success = false, want true. Error: %q", result.Error)
	}

	if result.Error != "" {
		t.Errorf("Convert() Error = %q, want empty", result.Error)
	}

	if !strings.Contains(result.Result, "8 000") {
		t.Errorf("Convert() Result = %q, want to contain '8 000'", result.Result)
	}
}

func TestApp_Convert_InvalidCurrency(t *testing.T) {
	date := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	rateData := &models.RateData{
		Date:  date,
		Rates: map[models.Currency]models.ExchangeRate{},
	}
	conv := createTestConverter(rateData, nil, 0, false)
	app := NewApp(conv)
	app.Startup(context.Background())

	req := ConvertRequest{
		Amount:   100,
		Currency: "GBP",
		Date:     "15.01.2024",
	}

	result := app.Convert(req)

	if result.Success {
		t.Errorf("Convert() Success = true, want false")
	}

	if result.Error == "" {
		t.Error("Convert() Error is empty, want error message")
	}

	if !strings.Contains(result.Error, "Неподдерживаемая валюта") {
		t.Errorf("Convert() Error = %q, want to contain 'Неподдерживаемая валюта'", result.Error)
	}
}

func TestApp_Convert_InvalidDate(t *testing.T) {
	date := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	rateData := &models.RateData{
		Date:  date,
		Rates: map[models.Currency]models.ExchangeRate{},
	}
	conv := createTestConverter(rateData, nil, 0, false)
	app := NewApp(conv)
	app.Startup(context.Background())

	req := ConvertRequest{
		Amount:   100,
		Currency: "USD",
		Date:     "invalid-date",
	}

	result := app.Convert(req)

	if result.Success {
		t.Errorf("Convert() Success = true, want false")
	}

	if result.Error == "" {
		t.Error("Convert() Error is empty, want error message")
	}

	if !strings.Contains(result.Error, "Неверный формат даты") {
		t.Errorf("Convert() Error = %q, want to contain 'Неверный формат даты'", result.Error)
	}
}

func TestApp_Convert_ConverterError(t *testing.T) {
	// Тест с отрицательной суммой должна вызвать ErrInvalidAmount
	date := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	rateData := &models.RateData{
		Date: date,
		Rates: map[models.Currency]models.ExchangeRate{
			models.USD: {
				Currency: models.USD,
				Rate:     80.0,
				Nominal:  1,
				Date:     date,
			},
		},
	}
	conv := createTestConverter(rateData, nil, 0, false)
	app := NewApp(conv)
	app.Startup(context.Background())

	req := ConvertRequest{
		Amount:   -100, // Отрицательная сумма должна вызвать ошибку
		Currency: "USD",
		Date:     "15.01.2024",
	}

	result := app.Convert(req)

	if result.Success {
		t.Errorf("Convert() Success = true, want false")
	}

	if result.Error == "" {
		t.Error("Convert() Error is empty, want error message")
	}

	if !strings.Contains(result.Error, "Сумма должна быть положительным числом") {
		t.Errorf("Convert() Error = %q, want to contain 'Сумма должна быть положительным числом'", result.Error)
	}
}

func TestApp_GetRate_Success(t *testing.T) {
	date := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	rateData := &models.RateData{
		Date: date,
		Rates: map[models.Currency]models.ExchangeRate{
			models.USD: {
				Currency: models.USD,
				Rate:     80.5,
				Nominal:  1,
				Date:     date,
			},
		},
	}
	conv := createTestConverter(rateData, nil, 0, false)
	app := NewApp(conv)
	app.Startup(context.Background())

	result := app.GetRate("USD", "15.01.2024")

	if !result.Success {
		t.Errorf("GetRate() Success = false, want true. Error: %q", result.Error)
	}

	if result.Error != "" {
		t.Errorf("GetRate() Error = %q, want empty", result.Error)
	}

	if result.Rate != 80.5 {
		t.Errorf("GetRate() Rate = %v, want 80.5", result.Rate)
	}
}

func TestApp_GetRate_RUB(t *testing.T) {
	date := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	rateData := &models.RateData{
		Date:  date,
		Rates: map[models.Currency]models.ExchangeRate{},
	}
	conv := createTestConverter(rateData, nil, 0, false)
	app := NewApp(conv)
	app.Startup(context.Background())

	result := app.GetRate("RUB", "15.01.2024")

	if !result.Success {
		t.Errorf("GetRate() Success = false, want true")
	}

	if result.Error != "" {
		t.Errorf("GetRate() Error = %q, want empty", result.Error)
	}

	if result.Rate != 1.0 {
		t.Errorf("GetRate() Rate = %v, want 1.0", result.Rate)
	}
}

func TestApp_GetRate_InvalidCurrency(t *testing.T) {
	date := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	rateData := &models.RateData{
		Date:  date,
		Rates: map[models.Currency]models.ExchangeRate{},
	}
	conv := createTestConverter(rateData, nil, 0, false)
	app := NewApp(conv)
	app.Startup(context.Background())

	result := app.GetRate("GBP", "15.01.2024")

	if result.Success {
		t.Errorf("GetRate() Success = true, want false")
	}

	if result.Error == "" {
		t.Error("GetRate() Error is empty, want error message")
	}

	if !strings.Contains(result.Error, "Неподдерживаемая валюта") {
		t.Errorf("GetRate() Error = %q, want to contain 'Неподдерживаемая валюта'", result.Error)
	}
}

func TestApp_GetRate_InvalidDate(t *testing.T) {
	date := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	rateData := &models.RateData{
		Date:  date,
		Rates: map[models.Currency]models.ExchangeRate{},
	}
	conv := createTestConverter(rateData, nil, 0, false)
	app := NewApp(conv)
	app.Startup(context.Background())

	result := app.GetRate("USD", "invalid-date")

	if result.Success {
		t.Errorf("GetRate() Success = true, want false")
	}

	if result.Error == "" {
		t.Error("GetRate() Error is empty, want error message")
	}

	if !strings.Contains(result.Error, "Неверный формат даты") {
		t.Errorf("GetRate() Error = %q, want to contain 'Неверный формат даты'", result.Error)
	}
}

func TestApp_GetRate_ConverterError(t *testing.T) {
	// Тест с датой в будущем должна вызвать ErrDateInFuture
	date := time.Now().AddDate(0, 0, 1) // Завтра
	rateData := &models.RateData{
		Date: date,
		Rates: map[models.Currency]models.ExchangeRate{
			models.USD: {
				Currency: models.USD,
				Rate:     80.0,
				Nominal:  1,
				Date:     date,
			},
		},
	}
	conv := createTestConverter(rateData, nil, 0, false)
	app := NewApp(conv)
	app.Startup(context.Background())

	// Используем дату в будущем
	futureDateStr := date.Format("02.01.2006")
	result := app.GetRate("USD", futureDateStr)

	if result.Success {
		t.Errorf("GetRate() Success = true, want false")
	}

	if result.Error == "" {
		t.Error("GetRate() Error is empty, want error message")
	}

	if !strings.Contains(result.Error, "Дата не может быть в будущем") {
		t.Errorf("GetRate() Error = %q, want to contain 'Дата не может быть в будущем'", result.Error)
	}
}

func TestParseDate_Success(t *testing.T) {
	dateStr := "15.01.2024"
	date, err := parseDate(dateStr)

	if err != nil {
		t.Fatalf("parseDate() error = %v, want nil", err)
	}

	expectedDate := time.Date(2024, 1, 15, 0, 0, 0, 0, time.Local)
	if !date.Equal(expectedDate) {
		t.Errorf("parseDate() date = %v, want %v", date, expectedDate)
	}
}

func TestParseDate_InvalidFormat(t *testing.T) {
	tests := []struct {
		name    string
		dateStr string
	}{
		{
			name:    "Invalid format",
			dateStr: "2024-01-15",
		},
		{
			name:    "Empty string",
			dateStr: "",
		},
		{
			name:    "Wrong separator",
			dateStr: "15/01/2024",
		},
		{
			name:    "Invalid day",
			dateStr: "32.01.2024",
		},
		{
			name:    "Invalid month",
			dateStr: "15.13.2024",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseDate(tt.dateStr)
			if err == nil {
				t.Errorf("parseDate() error = nil, want error for %q", tt.dateStr)
			}

			if !strings.Contains(err.Error(), "неверный формат даты") {
				t.Errorf("parseDate() error = %q, want to contain 'неверный формат даты'", err.Error())
			}
		})
	}
}

func TestTranslateError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{
			name: "ErrNilRateProvider - прямая ошибка",
			err:  converter.ErrNilRateProvider,
			want: "Ошибка конфигурации: источник курсов не настроен",
		},
		{
			name: "ErrInvalidAmount - прямая ошибка",
			err:  converter.ErrInvalidAmount,
			want: "Сумма должна быть положительным числом",
		},
		{
			name: "ErrDateInFuture - прямая ошибка",
			err:  converter.ErrDateInFuture,
			want: "Дата не может быть в будущем",
		},
		{
			name: "ErrUnsupportedCurrency - прямая ошибка",
			err:  models.ErrUnsupportedCurrency,
			want: "Неподдерживаемая валюта. Поддерживаются только USD, EUR и RUB",
		},
		{
			name: "Неизвестная ошибка - короткое сообщение",
			err:  errors.New("network error"),
			want: "network error",
		},
		{
			name: "Неизвестная ошибка - длинное сообщение (обрезается)",
			err:  errors.New("very long error message that exceeds 100 characters limit and should be replaced with generic message"),
			want: "Произошла ошибка при выполнении операции. Попробуйте еще раз.",
		},
		{
			name: "nil ошибка",
			err:  nil,
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := translateError(tt.err)
			if got != tt.want {
				t.Errorf("translateError() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestTranslateError_WrappedErrors(t *testing.T) {
	// Тест для обёрнутых ошибок через fmt.Errorf с %w
	tests := []struct {
		name string
		err  error
		want string
	}{
		{
			name: "fmt.Errorf обёртывает ErrInvalidAmount",
			err:  fmt.Errorf("validation failed: %w", converter.ErrInvalidAmount),
			want: "Сумма должна быть положительным числом",
		},
		{
			name: "fmt.Errorf обёртывает ErrDateInFuture",
			err:  fmt.Errorf("date check: %w", converter.ErrDateInFuture),
			want: "Дата не может быть в будущем",
		},
		{
			name: "fmt.Errorf обёртывает ErrNilRateProvider",
			err:  fmt.Errorf("config error: %w", converter.ErrNilRateProvider),
			want: "Ошибка конфигурации: источник курсов не настроен",
		},
		{
			name: "fmt.Errorf обёртывает ErrUnsupportedCurrency",
			err:  fmt.Errorf("currency check: %w", models.ErrUnsupportedCurrency),
			want: "Неподдерживаемая валюта. Поддерживаются только USD, EUR и RUB",
		},
		{
			name: "Множественные обёртки",
			err:  fmt.Errorf("outer: %w", fmt.Errorf("middle: %w", converter.ErrInvalidAmount)),
			want: "Сумма должна быть положительным числом",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := translateError(tt.err)
			if got != tt.want {
				t.Errorf("translateError() = %q, want %q", got, tt.want)
			}
		})
	}
}
