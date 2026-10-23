// Package converter предоставляет функциональность конвертации валют
package converter

import (
	"errors"
	"math"
	"time"
)

// Ошибки валидации
var (
	ErrInvalidAmount  = errors.New("сумма должна быть положительным числом")
	ErrAmountTooLarge = errors.New("сумма слишком велика")
	ErrDateInFuture   = errors.New("дата не может быть в будущем")
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
	if math.IsNaN(amount) || math.IsInf(amount, 0) || amount <= 0 {
		return ErrInvalidAmount
	}
	return nil
}

// ValidateDate проверяет корректность даты для получения курса
// Дата не может быть в будущем: календарная дата (год, месяц, день) в её
// собственной временной зоне сравнивается с сегодняшней локальной датой.
// Именно эта календарная дата уходит в запрос к ЦБ (buildURL), поэтому
// перевод в другую зону здесь недопустим: 24.10 00:00 UTC+14 в Москве ещё
// 23.10, но запрошен будет курс на 24.10
//
// Пример использования:
//
//	if err := ValidateDate(date); err != nil {
//	    return err
//	}
func ValidateDate(date time.Time) error {
	y, m, d := date.Date()
	ny, nm, nd := time.Now().Date()
	requested := time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
	today := time.Date(ny, nm, nd, 0, 0, 0, 0, time.UTC)
	if requested.After(today) {
		return ErrDateInFuture
	}
	return nil
}

func normalizeDate(date time.Time) time.Time {
	return time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
}
