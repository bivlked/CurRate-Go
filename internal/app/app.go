package app

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/bivlked/currate-go/internal/converter"
	"github.com/bivlked/currate-go/internal/models"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App - основной backend для GUI приложения
// Предоставляет методы для взаимодействия с frontend через Wails bindings
type App struct {
	ctx       context.Context
	converter *converter.Converter
}

// NewApp создает новый экземпляр App
func NewApp(conv *converter.Converter) *App {
	return &App{
		converter: conv,
	}
}

// Startup вызывается при запуске приложения
func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
}

// ConvertRequest - запрос на конвертацию из JavaScript
type ConvertRequest struct {
	Amount   float64 `json:"amount"`   // Сумма для конвертации
	Currency string  `json:"currency"` // "USD", "EUR" или "RUB"
	Date     string  `json:"date"`     // "DD.MM.YYYY"
}

// ConvertResponse - ответ на конвертацию для JavaScript
type ConvertResponse struct {
	Success bool   `json:"success"` // Успешность операции
	Result  string `json:"result"`  // Отформатированный результат (совместимость)
	Error   string `json:"error"`   // Сообщение об ошибке (если success=false)

	// Дополнительные поля для более богатого UI
	SourceAmount    float64 `json:"sourceAmount"`
	TargetAmountRUB float64 `json:"targetAmountRUB"`
	Rate            float64 `json:"rate"`
	Currency        string  `json:"currency"`
	CurrencySymbol  string  `json:"currencySymbol"`
	RequestedDate   string  `json:"requestedDate"`
	ActualDate      string  `json:"actualDate"`
}

// RateResponse - ответ для получения курса (live preview)
type RateResponse struct {
	Success bool    `json:"success"` // Успешность операции
	Rate    float64 `json:"rate"`    // Курс валюты (если success=true)
	Error   string  `json:"error"`   // Сообщение об ошибке (если success=false)
}

// Convert конвертирует валюту
// Вызывается из JavaScript для выполнения конвертации
func (a *App) Convert(req ConvertRequest) ConvertResponse {
	// Парсим валюту
	currency, err := models.ParseCurrency(req.Currency)
	if err != nil {
		return ConvertResponse{
			Success: false,
			Error:   fmt.Sprintf("Неподдерживаемая валюта: %s", req.Currency),
		}
	}

	// Парсим дату (формат DD.MM.YYYY)
	date, err := parseDate(req.Date)
	if err != nil {
		return ConvertResponse{
			Success: false,
			Error:   fmt.Sprintf("Неверный формат даты: %s. Используйте формат ДД.ММ.ГГГГ", req.Date),
		}
	}

	// Выполняем конвертацию
	result, err := a.converter.Convert(req.Amount, currency, date)
	if err != nil {
		// Преобразуем ошибку в понятное сообщение на русском
		return ConvertResponse{
			Success: false,
			Error:   translateError(err),
		}
	}

	return ConvertResponse{
		Success: true,
		Result:          result.FormattedStr,
		SourceAmount:    result.SourceAmount,
		TargetAmountRUB: result.TargetAmount,
		Rate:            result.Rate,
		Currency:        string(result.SourceCurrency),
		CurrencySymbol:  result.SourceCurrency.Symbol(),
		RequestedDate:   req.Date,
		ActualDate:      result.Date.Format("02.01.2006"),
	}
}

// GetRate получает курс валюты на указанную дату (для live preview)
// Вызывается из JavaScript при изменении даты для автоматического отображения курса
func (a *App) GetRate(currencyStr string, dateStr string) RateResponse {
	// Парсим валюту
	currency, err := models.ParseCurrency(currencyStr)
	if err != nil {
		return RateResponse{
			Success: false,
			Error:   fmt.Sprintf("Неподдерживаемая валюта: %s", currencyStr),
		}
	}

	// Парсим дату
	date, err := parseDate(dateStr)
	if err != nil {
		return RateResponse{
			Success: false,
			Error:   fmt.Sprintf("Неверный формат даты: %s", dateStr),
		}
	}

	// Для RUB всегда возвращаем 1.0
	if currency == models.RUB {
		return RateResponse{
			Success: true,
			Rate:    1.0,
		}
	}

	// Используем оптимизированный метод GetRate для получения курса без форматирования
	// Это избегает лишних вычислений и аллокаций для live preview
	rate, err := a.converter.GetRate(currency, date)
	if err != nil {
		return RateResponse{
			Success: false,
			Error:   translateError(err),
		}
	}

	return RateResponse{
		Success: true,
		Rate:    rate,
	}
}

// parseDate парсит дату из формата "DD.MM.YYYY"
// Использует time.ParseInLocation для сохранения локальной календарной даты
// Это предотвращает сдвиг дня для пользователей в отрицательных временных зонах
func parseDate(dateStr string) (time.Time, error) {
	layout := "02.01.2006"
	// Парсим дату в локальной временной зоне для сохранения календарной даты
	date, err := time.ParseInLocation(layout, dateStr, time.Local)
	if err != nil {
		return time.Time{}, fmt.Errorf("неверный формат даты: %w", err)
	}
	return date, nil
}

// translateError преобразует ошибку в понятное сообщение на русском языке
// Использует errors.Is для распознавания базовых ошибок, даже если они обёрнуты
func translateError(err error) string {
	if err == nil {
		return ""
	}

	// Проверяем базовые ошибки через errors.Is (работает с обёрнутыми ошибками)
	switch {
	case errors.Is(err, converter.ErrNilRateProvider):
		return "Ошибка конфигурации: источник курсов не настроен"
	case errors.Is(err, converter.ErrInvalidAmount):
		return "Сумма должна быть положительным числом"
	case errors.Is(err, converter.ErrDateInFuture):
		return "Дата не может быть в будущем"
	case errors.Is(err, models.ErrUnsupportedCurrency):
		return "Неподдерживаемая валюта. Поддерживаются только USD, EUR и RUB"
	default:
		// Для неизвестных ошибок возвращаем оригинальное сообщение
		// или общее сообщение, если оно слишком техническое
		errStr := err.Error()
		if len(errStr) > 100 {
			return "Произошла ошибка при выполнении операции. Попробуйте еще раз."
		}
		return errStr
	}
}

// ShowAbout показывает информацию о программе
func (a *App) ShowAbout() {
	_, _ = runtime.MessageDialog(a.ctx, runtime.MessageDialogOptions{
		Type:    runtime.InfoDialog,
		Title:   "О программе",
		Message: "Конвертер валют\n\nВерсия: 1.0.0\n\nКонвертирует доллары и евро в рубли по курсу ЦБ РФ на выбранную дату.\n\n© 2025 BiV\n\nРазработано с использованием:\n• Wails v2\n• Go 1.21+\n• WebView2",
	})
}

