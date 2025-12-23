// Package models содержит основные модели данных приложения
package models

import (
	"errors"
	"fmt"
	"strings"
)

// Ошибки валидации валют
var (
	ErrUnsupportedCurrency = errors.New("неподдерживаемая валюта")
)

// Currency представляет тип валюты
type Currency string

// Поддерживаемые валюты
const (
	USD Currency = "USD" // Доллар США
	EUR Currency = "EUR" // Евро
	RUB Currency = "RUB" // Российский рубль
)

// Validate проверяет, является ли валюта поддерживаемой
func (c Currency) Validate() error {
	switch c {
	case USD, EUR, RUB:
		return nil
	default:
		return fmt.Errorf("%w: %s", ErrUnsupportedCurrency, c)
	}
}

// Symbol возвращает символ валюты
func (c Currency) Symbol() string {
	switch c {
	case USD:
		return "$"
	case EUR:
		return "€"
	case RUB:
		return "₽"
	default:
		return string(c)
	}
}

// String возвращает строковое представление валюты
func (c Currency) String() string {
	return string(c)
}

// Name возвращает полное название валюты на русском
func (c Currency) Name() string {
	switch c {
	case USD:
		return "Доллар США"
	case EUR:
		return "Евро"
	case RUB:
		return "Российский рубль"
	default:
		return string(c)
	}
}

// ParseCurrency парсит строку в Currency и валидирует её
// Нормализует ввод: убирает пробелы и приводит к верхнему регистру
// Возвращает ошибку, если валюта не поддерживается
//
// Примеры:
//   - "usd" -> USD
//   - " eur " -> EUR
//   - "Usd" -> USD
func ParseCurrency(s string) (Currency, error) {
	// Нормализуем ввод: убираем пробелы и приводим к верхнему регистру
	normalized := strings.ToUpper(strings.TrimSpace(s))
	currency := Currency(normalized)
	if err := currency.Validate(); err != nil {
		return "", err
	}
	return currency, nil
}
