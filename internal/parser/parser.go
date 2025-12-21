// Package parser предоставляет функциональность для парсинга курсов валют с сайта ЦБ РФ
package parser

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/bivlked/currate-go/internal/models"
)

// Ошибки парсинга
var (
	ErrInvalidRate         = errors.New("invalid rate format")
	ErrInvalidNominal      = errors.New("invalid nominal format")
	ErrUnsupportedCurrency = errors.New("unsupported currency code")
)

// parseRate парсит строку курса в формате "80,7220" (с запятой как десятичным разделителем)
// и возвращает float64
func parseRate(s string) (float64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, ErrInvalidRate
	}

	// Заменяем запятую на точку (ЦБ РФ использует запятую как десятичный разделитель)
	s = strings.ReplaceAll(s, ",", ".")

	rate, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, fmt.Errorf("%w: %s", ErrInvalidRate, s)
	}

	if rate <= 0 {
		return 0, fmt.Errorf("%w: rate must be positive", ErrInvalidRate)
	}

	return rate, nil
}

// parseNominal парсит строку номинала в формат int
// Примеры: "1", "10", "100", "1000"
func parseNominal(s string) (int, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, ErrInvalidNominal
	}

	// Удаляем возможные пробелы в числе (например, "10 000")
	s = strings.ReplaceAll(s, " ", "")

	nominal, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("%w: %s", ErrInvalidNominal, s)
	}

	if nominal <= 0 {
		return 0, fmt.Errorf("%w: nominal must be positive", ErrInvalidNominal)
	}

	return nominal, nil
}

// parseCurrency конвертирует трехбуквенный код валюты в тип Currency
// Возвращает ошибку для неподдерживаемых валют
func parseCurrency(code string) (models.Currency, error) {
	code = strings.ToUpper(strings.TrimSpace(code))

	currency := models.Currency(code)
	if err := currency.Validate(); err != nil {
		return "", fmt.Errorf("%w: %s", ErrUnsupportedCurrency, code)
	}

	return currency, nil
}
