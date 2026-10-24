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
	"github.com/bivlked/currate-go/internal/parser"
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

func TestNewApp_NilConverter(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("NewApp(nil) should panic")
		}
		msg, ok := r.(string)
		if !ok || !strings.Contains(msg, "converter must not be nil") {
			t.Errorf("unexpected panic message: %v", r)
		}
	}()
	NewApp(nil)
}

func TestConvert_BeforeStartup(t *testing.T) {
	date := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	rateData := &models.RateData{
		Date: date,
		Rates: map[models.Currency]models.ExchangeRate{
			models.USD: {Currency: models.USD, Rate: 80.0, Nominal: 1, Date: date},
		},
	}
	conv := createTestConverter(rateData, nil, 0, false)
	app := NewApp(conv)
	// Не вызываем Startup — a.ctx == nil

	result := app.Convert(ConvertRequest{Amount: 100, Currency: "USD", Date: "15.01.2024"})
	if result.Success {
		t.Fatal("Convert() before Startup should return Success=false")
	}
	if !strings.Contains(result.Error, "не инициализировано") {
		t.Errorf("Convert() Error = %q, want to contain 'не инициализировано'", result.Error)
	}
}

func TestGetRate_BeforeStartup(t *testing.T) {
	date := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	rateData := &models.RateData{
		Date: date,
		Rates: map[models.Currency]models.ExchangeRate{
			models.USD: {Currency: models.USD, Rate: 80.0, Nominal: 1, Date: date},
		},
	}
	conv := createTestConverter(rateData, nil, 0, false)
	app := NewApp(conv)
	// Не вызываем Startup — a.ctx == nil

	result := app.GetRate("USD", "15.01.2024")
	if result.Success {
		t.Fatal("GetRate() before Startup should return Success=false")
	}
	if !strings.Contains(result.Error, "не инициализировано") {
		t.Errorf("GetRate() Error = %q, want to contain 'не инициализировано'", result.Error)
	}
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

	type ctxKey struct{}
	first := context.WithValue(context.Background(), ctxKey{}, "first")
	app.Startup(first)
	if app.ctx != first {
		t.Fatal("Startup() не сохранил контекст")
	}

	// Повторный вызов (Startup забинден и доступен из JS) не затирает контекст
	app.Startup(context.Background())
	if app.ctx != first {
		t.Error("повторный Startup() затёр рабочий контекст")
	}
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

// RUB не должен обходить проверку даты: Convert отвергает будущую дату,
// и превью курса обязано вести себя так же
func TestApp_GetRate_RUBFutureDate(t *testing.T) {
	conv := createTestConverter(nil, nil, 0, false)
	app := NewApp(conv)
	app.Startup(context.Background())

	future := time.Now().AddDate(0, 0, 2).Format("02.01.2006")
	result := app.GetRate("RUB", future)

	if result.Success {
		t.Fatalf("GetRate(RUB, %s) Success = true, want false", future)
	}
	if result.Error != "Дата не может быть в будущем" {
		t.Errorf("GetRate() Error = %q", result.Error)
	}
}

// В тестовой сборке токены не внедрены: SendStar сразу отвечает понятной
// ошибкой и не ходит в сеть
func TestApp_SendStar_NotConfigured(t *testing.T) {
	app := NewApp(createTestConverter(nil, nil, 0, false))

	result := app.SendStar()

	if result.Success {
		t.Fatal("SendStar() Success = true без токенов")
	}
	if result.Error != "Функция доступна только в release-версии приложения." {
		t.Errorf("SendStar() Error = %q", result.Error)
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
			name: "ErrAmountTooLarge",
			err:  converter.ErrAmountTooLarge,
			want: "Сумма слишком велика",
		},
		{
			// EUR до 1999 года: ответ ЦБ есть, но курса валюты в нём нет
			name: "ErrRateNotFound",
			err:  fmt.Errorf("%w: EUR", converter.ErrRateNotFound),
			want: "ЦБ РФ не публиковал курс этой валюты на выбранную дату",
		},
		{
			// 01.01.1990: ЦБ отвечает пустым <ValCurs>, это не сетевая ошибка
			name: "Пустой ответ ЦБ внутри ErrFetchRates",
			err:  fmt.Errorf("%w: %w", converter.ErrFetchRates, fmt.Errorf("parse: %w", parser.ErrNoXMLRates)),
			want: "ЦБ РФ не публиковал курсы на выбранную дату",
		},
		{
			name: "Некорректный XML внутри ErrFetchRates",
			err:  fmt.Errorf("%w: %w", converter.ErrFetchRates, parser.ErrInvalidXML),
			want: "ЦБ РФ вернул некорректный ответ. Попробуйте позже.",
		},
		{
			// ответ без валидной даты курсов - не проблема связи
			name: "Невалидная дата XML внутри ErrFetchRates",
			err:  fmt.Errorf("%w: %w", converter.ErrFetchRates, parser.ErrInvalidXMLDate),
			want: "ЦБ РФ вернул некорректный ответ. Попробуйте позже.",
		},
		{
			name: "Сетевая ошибка внутри ErrFetchRates",
			err:  fmt.Errorf("%w: %w", converter.ErrFetchRates, parser.ErrMaxRetries),
			want: "Не удалось получить курсы ЦБ РФ. Проверьте подключение к интернету.",
		},
		{
			// технический текст на английском в интерфейс не попадает
			name: "Неизвестная короткая ошибка",
			err:  errors.New("network error"),
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

// datedRateProvider отдаёт курс USD в зависимости от запрошенной даты;
// для дат, которых нет в карте, возвращает курс последней известной даты
// не позже запрошенной (как ЦБ на выходные)
type datedRateProvider struct {
	rates map[string]float64 // "2006-01-02" -> курс
}

func (p *datedRateProvider) FetchRates(_ context.Context, date time.Time) (*models.RateData, error) {
	for d := date; d.After(date.AddDate(0, 0, -10)); d = d.AddDate(0, 0, -1) {
		if rate, ok := p.rates[d.Format("2006-01-02")]; ok {
			rd := models.NewRateData(d)
			rd.AddRate(models.ExchangeRate{Currency: models.USD, Rate: rate, Nominal: 1, Date: d})
			return rd, nil
		}
	}
	return nil, errors.New("no rates")
}

func newDatedApp(rates map[string]float64) *App {
	app := NewApp(converter.NewConverter(&datedRateProvider{rates: rates}, nil))
	app.Startup(context.Background())
	return app
}

func TestApp_Convert_WithReference(t *testing.T) {
	app := newDatedApp(map[string]float64{
		"2026-10-23": 80.5381,
		"2026-10-01": 79.9012,
	})

	resp := app.Convert(ConvertRequest{Amount: 1000, Currency: "USD", Date: "23.10.2026", RefDate: "01.10.2026"})

	if !resp.Success || !resp.HasReference {
		t.Fatalf("Convert() = %+v, want success with reference", resp)
	}
	if resp.TargetAmountRUB != 80538.1 || resp.RefAmountRUB != 79901.2 {
		t.Errorf("суммы = %v / %v, want 80538.1 / 79901.2", resp.TargetAmountRUB, resp.RefAmountRUB)
	}
	// float64: 80538.1 - 79901.2 = 636.9000000000087
	if resp.RefDiffRUB != 636.9 {
		t.Errorf("RefDiffRUB = %v, want 636.9", resp.RefDiffRUB)
	}
	if resp.RefRate != 79.9012 || resp.RefActualDate != "01.10.2026" || resp.RefRequestedDate != "01.10.2026" {
		t.Errorf("референс = %v %q %q", resp.RefRate, resp.RefActualDate, resp.RefRequestedDate)
	}
	if resp.Result != "80 538,10 руб. ($1 000,00 по курсу 80,5381)" {
		t.Errorf("Result = %q: строка для буфера обмена не должна меняться", resp.Result)
	}
}

func TestApp_Convert_ReferenceOnWeekend(t *testing.T) {
	// 04.10.2026 - воскресенье: курс референса берётся за пятницу 02.10
	app := newDatedApp(map[string]float64{
		"2026-10-23": 80.5381,
		"2026-10-02": 81.0000,
	})

	resp := app.Convert(ConvertRequest{Amount: 10, Currency: "USD", Date: "23.10.2026", RefDate: "04.10.2026"})

	if !resp.HasReference || resp.RefActualDate != "02.10.2026" || resp.RefRequestedDate != "04.10.2026" {
		t.Fatalf("референс = %+v, want фактическая дата 02.10.2026", resp)
	}
	if resp.RefDiffRUB != -4.62 { // 805.38 - 810.00
		t.Errorf("RefDiffRUB = %v, want -4.62", resp.RefDiffRUB)
	}
}

// Сбой референса не должен ломать основную конвертацию
func TestApp_Convert_ReferenceErrors(t *testing.T) {
	app := newDatedApp(map[string]float64{"2026-10-23": 80.5381})
	future := time.Now().AddDate(0, 0, 3).Format("02.01.2006")

	tests := []struct {
		name    string
		refDate string
		wantErr string
	}{
		{"Неверный формат", "1.10.2026", "Неверный формат референсной даты: 1.10.2026"},
		{"Будущая дата", future, "Дата не может быть в будущем"},
		{"Сбой получения курса", "01.01.2000", "Не удалось получить курсы ЦБ РФ. Проверьте подключение к интернету."},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := app.Convert(ConvertRequest{Amount: 1, Currency: "USD", Date: "23.10.2026", RefDate: tt.refDate})
			if !resp.Success || resp.TargetAmountRUB != 80.54 {
				t.Fatalf("основная конвертация сломалась: %+v", resp)
			}
			if resp.HasReference || resp.RefError != tt.wantErr {
				t.Errorf("HasReference=%v RefError=%q, want false %q", resp.HasReference, resp.RefError, tt.wantErr)
			}
		})
	}
}

func TestApp_Convert_NoReference(t *testing.T) {
	app := newDatedApp(map[string]float64{"2026-10-23": 80.5381})

	resp := app.Convert(ConvertRequest{Amount: 1, Currency: "USD", Date: "23.10.2026"})

	if !resp.Success || resp.HasReference || resp.RefError != "" || resp.RefRequestedDate != "" {
		t.Errorf("без refDate поля референса должны быть пустыми: %+v", resp)
	}
}

func TestApp_GetRate_ActualDate(t *testing.T) {
	app := newDatedApp(map[string]float64{"2026-10-02": 81.0})

	resp := app.GetRate("USD", "04.10.2026") // воскресенье

	if !resp.Success || resp.Rate != 81.0 || resp.ActualDate != "02.10.2026" {
		t.Errorf("GetRate() = %+v, want 81 за 02.10.2026", resp)
	}
}
