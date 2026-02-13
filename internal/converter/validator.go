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
	if math.IsNaN(amount) || math.IsInf(amount, 0) || amount <= 0 {
		return ErrInvalidAmount
	}
	return nil
}

// ValidateDate проверяет корректность даты для получения курса
// Дата не может быть в будущем (больше текущего времени)
// Сравнение выполняется по календарным датам (год, месяц, день) в локальной временной зоне
// для корректной работы с разными временными зонами
//
// Пример использования:
//
//	if err := ValidateDate(date); err != nil {
//	    return err
//	}
func ValidateDate(date time.Time) error {
	// Нормализуем входную дату (только дата, без времени) в её исходной временной зоне
	normalized := normalizeDate(date)

	// Конвертируем нормализованную дату в локальную временную зону и извлекаем календарную дату
	normalizedInLocal := normalized.In(time.Local)
	dateYear, dateMonth, dateDay := normalizedInLocal.Date()

	// Получаем текущее время в локальной временной зоне и извлекаем календарную дату
	nowLocal := time.Now()
	nowYear, nowMonth, nowDay := nowLocal.Date()

	// Сравниваем календарные даты (год, месяц, день) в локальной временной зоне
	// Это гарантирует корректное сравнение независимо от временной зоны входной даты
	if dateYear > nowYear ||
		(dateYear == nowYear && dateMonth > nowMonth) ||
		(dateYear == nowYear && dateMonth == nowMonth && dateDay > nowDay) {
		return ErrDateInFuture
	}
	return nil
}

func normalizeDate(date time.Time) time.Time {
	return time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
}
