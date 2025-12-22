package app

import (
	"context"
	"fmt"
	"time"

	"github.com/bivlked/currate-go/internal/converter"
	"github.com/bivlked/currate-go/internal/models"
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
	Result  string `json:"result"`  // Отформатированный результат (если success=true)
	Error   string `json:"error"`   // Сообщение об ошибке (если success=false)
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
		Result:  result.FormattedStr,
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

	// Используем конвертер для получения курса (конвертируем 1 единицу валюты)
	// Это использует кэш и все механизмы конвертера
	result, err := a.converter.Convert(1.0, currency, date)
	if err != nil {
		return RateResponse{
			Success: false,
			Error:   translateError(err),
		}
	}

	// Извлекаем курс из результата конвертации
	rate := result.Rate

	return RateResponse{
		Success: true,
		Rate:    rate,
	}
}

// parseDate парсит дату из формата "DD.MM.YYYY"
func parseDate(dateStr string) (time.Time, error) {
	layout := "02.01.2006"
	date, err := time.Parse(layout, dateStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("неверный формат даты: %w", err)
	}
	return date, nil
}

// translateError преобразует ошибку в понятное сообщение на русском языке
func translateError(err error) string {
	if err == nil {
		return ""
	}

	errStr := err.Error()

	// Переводим известные ошибки
	switch {
	case errStr == "источник курсов не задан":
		return "Ошибка конфигурации: источник курсов не настроен"
	case errStr == "сумма должна быть положительным числом":
		return "Сумма должна быть положительным числом"
	case errStr == "дата не может быть в будущем":
		return "Дата не может быть в будущем"
	case errStr == "неподдерживаемая валюта":
		return "Неподдерживаемая валюта. Поддерживаются только USD, EUR и RUB"
	default:
		// Для неизвестных ошибок возвращаем оригинальное сообщение
		// или общее сообщение, если оно слишком техническое
		if len(errStr) > 100 {
			return "Произошла ошибка при выполнении операции. Попробуйте еще раз."
		}
		return errStr
	}
}

