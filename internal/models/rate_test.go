package models

import "testing"

func TestNewRateData(t *testing.T) {
	date := testPastDateUTC()
	rd := NewRateData(date)

	if rd == nil {
		t.Fatal("NewRateData() вернула nil")
	}

	if !rd.Date.Equal(date) {
		t.Errorf("NewRateData() дата = %v, ожидается %v", rd.Date, date)
	}

	if rd.Rates == nil {
		t.Error("NewRateData() карта курсов не инициализирована")
	}

	if len(rd.Rates) != 0 {
		t.Errorf("NewRateData() карта курсов не пуста: len = %d", len(rd.Rates))
	}
}

func TestRateDataAddRate(t *testing.T) {
	date := testPastDateUTC()
	rd := NewRateData(date)

	rate := ExchangeRate{
		Currency: USD,
		Rate:     95.50,
		Nominal:  1,
		Date:     date,
	}

	rd.AddRate(rate)

	if len(rd.Rates) != 1 {
		t.Errorf("AddRate() не добавил курс: len = %d", len(rd.Rates))
	}

	storedRate, ok := rd.Rates[USD]
	if !ok {
		t.Error("AddRate() курс USD не найден")
	}

	if storedRate.Rate != rate.Rate {
		t.Errorf("AddRate() курс = %v, ожидается %v", storedRate.Rate, rate.Rate)
	}
}

func TestRateDataGetRate(t *testing.T) {
	date := testPastDateUTC()
	rd := NewRateData(date)

	usdRate := ExchangeRate{
		Currency: USD,
		Rate:     95.50,
		Nominal:  1,
		Date:     date,
	}

	eurRate := ExchangeRate{
		Currency: EUR,
		Rate:     105.20,
		Nominal:  1,
		Date:     date,
	}

	rd.AddRate(usdRate)
	rd.AddRate(eurRate)

	// Тест получения существующего курса
	t.Run("Получить USD курс", func(t *testing.T) {
		rate, ok := rd.GetRate(USD)
		if !ok {
			t.Error("GetRate() не нашел USD курс")
		}
		if rate.Rate != usdRate.Rate {
			t.Errorf("GetRate() курс = %v, ожидается %v", rate.Rate, usdRate.Rate)
		}
	})

	// Тест получения несуществующего курса
	t.Run("Получить несуществующий курс", func(t *testing.T) {
		_, ok := rd.GetRate(Currency("GBP"))
		if ok {
			t.Error("GetRate() нашел несуществующий курс GBP")
		}
	})
}

func TestExchangeRateStruct(t *testing.T) {
	date := testPastDateUTC()

	rate := ExchangeRate{
		Currency: USD,
		Rate:     95.50,
		Nominal:  1,
		Date:     date,
	}

	if rate.Currency != USD {
		t.Errorf("ExchangeRate.Currency = %v, ожидается %v", rate.Currency, USD)
	}

	if rate.Rate != 95.50 {
		t.Errorf("ExchangeRate.Rate = %v, ожидается %v", rate.Rate, 95.50)
	}

	if rate.Nominal != 1 {
		t.Errorf("ExchangeRate.Nominal = %v, ожидается %v", rate.Nominal, 1)
	}

	if !rate.Date.Equal(date) {
		t.Errorf("ExchangeRate.Date = %v, ожидается %v", rate.Date, date)
	}
}

func TestConversionResultStruct(t *testing.T) {
	date := testPastDateUTC()

	result := ConversionResult{
		SourceCurrency: USD,
		TargetCurrency: RUB,
		SourceAmount:   100.0,
		TargetAmount:   9550.0,
		Rate:           95.50,
		Date:           date,
	}

	if result.SourceCurrency != USD {
		t.Errorf("ConversionResult.SourceCurrency = %v, ожидается %v", result.SourceCurrency, USD)
	}

	if result.TargetCurrency != RUB {
		t.Errorf("ConversionResult.TargetCurrency = %v, ожидается %v", result.TargetCurrency, RUB)
	}

	if result.SourceAmount != 100.0 {
		t.Errorf("ConversionResult.SourceAmount = %v, ожидается %v", result.SourceAmount, 100.0)
	}

	if result.TargetAmount != 9550.0 {
		t.Errorf("ConversionResult.TargetAmount = %v, ожидается %v", result.TargetAmount, 9550.0)
	}

	if result.Rate != 95.50 {
		t.Errorf("ConversionResult.Rate = %v, ожидается %v", result.Rate, 95.50)
	}

	if !result.Date.Equal(date) {
		t.Errorf("ConversionResult.Date = %v, ожидается %v", result.Date, date)
	}
}
