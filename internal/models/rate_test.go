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
