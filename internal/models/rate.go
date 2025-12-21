package models

import "time"

// ExchangeRate представляет курс валюты ЦБ РФ
type ExchangeRate struct {
	Currency Currency  // Валюта
	Rate     float64   // Курс к рублю
	Nominal  int       // Номинал (количество единиц валюты)
	Date     time.Time // Дата курса
}

// ConversionResult представляет результат конвертации валюты
type ConversionResult struct {
	SourceCurrency Currency  // Исходная валюта
	TargetCurrency Currency  // Целевая валюта (всегда RUB)
	SourceAmount   float64   // Исходная сумма
	TargetAmount   float64   // Результат конвертации
	Rate           float64   // Использованный курс
	Date           time.Time // Дата курса
	FormattedStr   string    // Отформатированная строка для отображения
}

// RateData представляет полные данные о курсах валют на определенную дату
type RateData struct {
	Date  time.Time                 // Дата курса
	Rates map[Currency]ExchangeRate // Курсы валют (ключ - валюта)
}

// NewRateData создает новый RateData с инициализированной картой
func NewRateData(date time.Time) *RateData {
	return &RateData{
		Date:  date,
		Rates: make(map[Currency]ExchangeRate),
	}
}

// AddRate добавляет курс валюты в RateData
func (rd *RateData) AddRate(rate ExchangeRate) {
	rd.Rates[rate.Currency] = rate
}

// GetRate возвращает курс валюты, если он есть
func (rd *RateData) GetRate(currency Currency) (ExchangeRate, bool) {
	rate, ok := rd.Rates[currency]
	return rate, ok
}
