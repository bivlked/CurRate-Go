// Package converter предоставляет функциональность конвертации валют
package converter

import (
	"errors"
	"math"
	"time"
)

// Ошибки валидации
var (
	ErrInvalidAmount = errors.New("сумма должна быть положительным числом")
	ErrDateInFuture  = errors.New("дата не может быть в будущем")
)

// ValidateAmount проверяет корректность суммы для конвертации
// Сумма должна быть положительным числом (> 0)
//
// Пример использования:
//
//	if err := ValidateAmount(amount); err != nil {
//	    return err
//	}
func ValidateAmount(amount float64) error {
	if math.IsNaN(amount) || math.IsInf(amount, 0) {
		return ErrInvalidAmount
	}
	if amount <= 0 {
		return ErrInvalidAmount
	}
	return nil
}

// ValidateDate проверяет корректность даты для получения курса
// Дата не может быть в будущем (больше текущего времени)
//
// Пример использования:
//
//	if err := ValidateDate(date); err != nil {
//	    return err
//	}
func ValidateDate(date time.Time) error {
	normalized := normalizeDate(date)
	nowInDateLocation := time.Now().In(date.Location())
	if normalized.After(normalizeDate(nowInDateLocation)) {
		return ErrDateInFuture
	}
	return nil
}

func normalizeDate(date time.Time) time.Time {
	return time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
}
