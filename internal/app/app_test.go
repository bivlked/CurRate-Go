package app

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/bivlked/currate-go/internal/converter"
	"github.com/bivlked/currate-go/internal/models"
)

// MockConverter - мок для converter.Converter
type MockConverter struct {
	convertResult *models.ConversionResult
	convertError  error
	getRateResult float64
	getRateError  error
}

func (m *MockConverter) Convert(amount float64, currency models.Currency, date time.Time) (*models.ConversionResult, error) {
	if m.convertError != nil {
		return nil, m.convertError
	}
	return m.convertResult, nil
}

func (m *MockConverter) GetRate(currency models.Currency, date time.Time) (float64, error) {
	if m.getRateError != nil {
		return 0, m.getRateError
	}
	return m.getRateResult, nil
}

func TestTranslateError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		want     string
		wantWrap bool // Нужно ли обернуть ошибку для теста
	}{
		{
			name: "ErrNilRateProvider - прямая ошибка",
			err:  converter.ErrNilRateProvider,
			want: "Ошибка конфигурации: источник курсов не настроен",
		},
		{
			name:     "ErrNilRateProvider - обёрнутая ошибка",
			err:      fmt.Errorf("wrapped: %w", converter.ErrNilRateProvider),
			wantWrap: false, // Уже обёрнута правильно
			want:     "Ошибка конфигурации: источник курсов не настроен",
		},
		{
			name: "ErrInvalidAmount - прямая ошибка",
			err:  converter.ErrInvalidAmount,
			want: "Сумма должна быть положительным числом",
		},
		{
			name:     "ErrInvalidAmount - обёрнутая ошибка",
			err:      fmt.Errorf("validation failed: %w", converter.ErrInvalidAmount),
			wantWrap: false, // Уже обёрнута правильно
			want:     "Сумма должна быть положительным числом",
		},
		{
			name: "ErrDateInFuture - прямая ошибка",
			err:  converter.ErrDateInFuture,
			want: "Дата не может быть в будущем",
		},
		{
			name:     "ErrDateInFuture - обёрнутая ошибка",
			err:      fmt.Errorf("date validation: %w", converter.ErrDateInFuture),
			wantWrap: false, // Уже обёрнута правильно
			want:     "Дата не может быть в будущем",
		},
		{
			name: "ErrUnsupportedCurrency - прямая ошибка",
			err:  models.ErrUnsupportedCurrency,
			want: "Неподдерживаемая валюта. Поддерживаются только USD, EUR и RUB",
		},
		{
			name:     "ErrUnsupportedCurrency - обёрнутая ошибка",
			err:      fmt.Errorf("currency error: %w", models.ErrUnsupportedCurrency),
			wantWrap: false, // Уже обёрнута правильно
			want:     "Неподдерживаемая валюта. Поддерживаются только USD, EUR и RUB",
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
		{
			name:     "Множественные обёртки",
			err:      fmt.Errorf("outer: %w", fmt.Errorf("middle: %w", converter.ErrInvalidAmount)),
			wantWrap: false, // Уже обёрнута правильно
			want:     "Сумма должна быть положительным числом",
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

