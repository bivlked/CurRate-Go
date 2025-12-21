// +build integration

package parser

import (
	"testing"
	"time"

	"github.com/bivlked/currate-go/internal/models"
)

// Интеграционные тесты с реальным XML API ЦБ РФ
// Запускаются с флагом: go test -tags=integration ./internal/parser
// Требуют доступа к интернету

func TestFetchRatesIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Пропускаем интеграционный тест в кратком режиме")
	}

	t.Run("Реальный запрос к XML API ЦБ РФ", func(t *testing.T) {
		// Используем недавнюю дату (не сегодняшнюю, так как данные могут обновляться)
		date := time.Date(2025, 12, 20, 0, 0, 0, 0, time.UTC)
		data, err := FetchRates(date)

		if err != nil {
			t.Fatalf("Ошибка получения курсов: %v", err)
		}

		if data == nil {
			t.Fatal("Data не должна быть nil")
		}

		if len(data.Rates) == 0 {
			t.Error("Ожидались валюты в результате")
		}

		// Проверяем наличие основных валют
		usd, hasUSD := data.Rates[models.USD]
		if !hasUSD {
			t.Error("USD должен присутствовать в результатах")
		} else {
			// Проверяем, что курс в разумных пределах (не ноль, не отрицательный)
			if usd.Rate <= 0 {
				t.Errorf("USD курс должен быть положительным, получено: %.4f", usd.Rate)
			}
			if usd.Nominal != 1 {
				t.Errorf("USD nominal должен быть 1, получено: %d", usd.Nominal)
			}
			t.Logf("USD курс: %.4f (номинал: %d)", usd.Rate, usd.Nominal)
		}

		eur, hasEUR := data.Rates[models.EUR]
		if !hasEUR {
			t.Error("EUR должен присутствовать в результатах")
		} else {
			if eur.Rate <= 0 {
				t.Errorf("EUR курс должен быть положительным, получено: %.4f", eur.Rate)
			}
			if eur.Nominal != 1 {
				t.Errorf("EUR nominal должен быть 1, получено: %d", eur.Nominal)
			}
			t.Logf("EUR курс: %.4f (номинал: %d)", eur.Rate, eur.Nominal)
		}

		// Проверяем, что есть хотя бы одна валюта с номиналом > 1
		foundMultiNominal := false
		for code, rate := range data.Rates {
			if rate.Nominal > 1 {
				foundMultiNominal = true
				t.Logf("Валюта с номиналом > 1: %s курс: %.4f (номинал: %d)", code, rate.Rate, rate.Nominal)
				break
			}
		}
		if !foundMultiNominal {
			t.Log("Внимание: не найдено валют с номиналом > 1")
		}

		t.Logf("Всего получено валют: %d", len(data.Rates))
		t.Logf("Дата курсов: %v", data.Date)
	})

	t.Run("Запрос с устаревшей датой", func(t *testing.T) {
		// Проверяем, что API работает и с более старыми датами
		date := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
		data, err := FetchRates(date)

		if err != nil {
			t.Fatalf("Ошибка получения курсов для старой даты: %v", err)
		}

		if data == nil {
			t.Fatal("Data не должна быть nil")
		}

		if len(data.Rates) == 0 {
			t.Error("Ожидались валюты в результате для старой даты")
		}

		t.Logf("Получено валют для даты %v: %d", date, len(data.Rates))
	})
}

func TestFetchLatestRatesIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Пропускаем интеграционный тест в кратком режиме")
	}

	t.Run("Реальный запрос последних курсов", func(t *testing.T) {
		data, err := FetchLatestRates()

		if err != nil {
			t.Fatalf("Ошибка получения последних курсов: %v", err)
		}

		if data == nil {
			t.Fatal("Data не должна быть nil")
		}

		if len(data.Rates) == 0 {
			t.Error("Ожидались валюты в результате")
		}

		// Проверяем, что дата не в далеком прошлом
		now := time.Now()
		daysDiff := now.Sub(data.Date).Hours() / 24
		if daysDiff > 7 {
			t.Logf("Внимание: Дата курсов более 7 дней назад: %v", data.Date)
		}

		t.Logf("Дата курсов: %v", data.Date)
		t.Logf("Получено валют: %d", len(data.Rates))

		// Проверяем основные валюты
		if usd, ok := data.Rates[models.USD]; ok {
			t.Logf("USD: %.4f", usd.Rate)
		}
		if eur, ok := data.Rates[models.EUR]; ok {
			t.Logf("EUR: %.4f", eur.Rate)
		}
	})
}

func TestXMLAPIResponseFormat(t *testing.T) {
	if testing.Short() {
		t.Skip("Пропускаем интеграционный тест в кратком режиме")
	}

	t.Run("Проверка формата XML ответа", func(t *testing.T) {
		date := time.Date(2025, 12, 20, 0, 0, 0, 0, time.UTC)

		// Строим URL и выполняем запрос вручную, чтобы проверить формат
		url := buildURL(date)
		t.Logf("URL запроса: %s", url)

		body, err := fetchXML(url)
		if err != nil {
			t.Fatalf("Ошибка получения XML: %v", err)
		}
		defer body.Close()

		// Парсим XML
		data, err := ParseXML(body, date)
		if err != nil {
			t.Fatalf("Ошибка парсинга XML: %v", err)
		}

		// Проверяем, что получены данные
		if len(data.Rates) == 0 {
			t.Fatal("Не получено ни одной валюты из XML")
		}

		// Выводим информацию о первых 5 валютах для проверки
		count := 0
		for code, rate := range data.Rates {
			if count >= 5 {
				break
			}
			t.Logf("%s: %.4f (номинал: %d)", code, rate.Rate, rate.Nominal)
			count++
		}

		t.Logf("Всего валют в XML: %d", len(data.Rates))
	})
}
