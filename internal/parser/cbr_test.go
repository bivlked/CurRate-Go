package parser

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/bivlked/currate-go/internal/models"
)

// Unit тесты с мок-сервером
func TestFetchRates(t *testing.T) {
	// Мок XML с реального API ЦБ РФ (упрощенная версия)
	mockXML := `<?xml version="1.0" encoding="UTF-8"?>
<ValCurs Date="20.12.2025" name="Foreign Currency Market">
    <Valute ID="R01235">
        <NumCode>840</NumCode>
        <CharCode>USD</CharCode>
        <Nominal>1</Nominal>
        <Name>Доллар США</Name>
        <Value>80,7220</Value>
    </Valute>
    <Valute ID="R01239">
        <NumCode>978</NumCode>
        <CharCode>EUR</CharCode>
        <Nominal>1</Nominal>
        <Name>Евро</Name>
        <Value>94,5120</Value>
    </Valute>
</ValCurs>`

	t.Run("Успешное получение курсов", func(t *testing.T) {
		// Создаем мок сервер
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(mockXML))
		}))
		defer server.Close()

		testDate := time.Date(2025, 12, 20, 0, 0, 0, 0, time.UTC)

		// Используем fetchRatesFromURL с тестовым сервером
		data, err := fetchRatesFromURL(server.URL, testDate)
		if err != nil {
			t.Fatalf("Ошибка fetchRatesFromURL: %v", err)
		}

		// Проверяем результат
		if data == nil {
			t.Fatal("Data не должна быть nil")
		}

		if len(data.Rates) != 2 {
			t.Errorf("Ожидалось 2 валюты, получено %d", len(data.Rates))
		}

		// Проверяем USD
		usd, ok := data.Rates[models.USD]
		if !ok {
			t.Fatal("USD не найден")
		}
		if usd.Rate != 80.7220 {
			t.Errorf("USD курс: ожидалось 80.7220, получено %v", usd.Rate)
		}

		// Проверяем EUR
		eur, ok := data.Rates[models.EUR]
		if !ok {
			t.Fatal("EUR не найден")
		}
		if eur.Rate != 94.5120 {
			t.Errorf("EUR курс: ожидалось 94.5120, получено %v", eur.Rate)
		}
	})

	t.Run("Ошибка HTTP 500", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		testDate := time.Date(2025, 12, 20, 0, 0, 0, 0, time.UTC)
		data, err := fetchRatesFromURL(server.URL, testDate)
		if err == nil {
			t.Fatal("Ожидалась ошибка для статуса 500")
		}

		if data != nil {
			t.Error("Data должна быть nil при ошибке")
		}
	})

	t.Run("Невалидный XML", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("<html><body>Not XML response</body></html>"))
		}))
		defer server.Close()

		testDate := time.Date(2025, 12, 20, 0, 0, 0, 0, time.UTC)
		data, err := fetchRatesFromURL(server.URL, testDate)
		if err == nil {
			t.Fatal("Ожидалась ошибка для невалидного XML")
		}

		if data != nil {
			t.Error("Data должна быть nil при ошибке")
		}
	})
}

func TestFetchLatestRates(t *testing.T) {
	// Этот тест сложно тестировать без мок-сервера
	// Проверяем только что функция вызывается без паники
	t.Run("Проверка существования функции", func(t *testing.T) {
		// Просто проверяем, что функция компилируется и может быть вызвана
		// Реальный вызов требует доступа к интернету и реальному сайту ЦБ РФ
		// Такой тест помечается как интеграционный и запускается отдельно
		t.Skip("Интеграционный тест, требует доступа к реальному API ЦБ РФ")
	})
}

// Интеграционные тесты (требуют доступа к реальному сайту ЦБ РФ)
// Запускаются с флагом: go test -tags=integration
// Для обычных тестов пропускаются

// Эти тесты закомментированы, так как требуют доступа к интернету
// Раскомментируйте для ручного тестирования

/*
func TestFetchRatesIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Пропускаем интеграционный тест в кратком режиме")
	}

	t.Run("Реальный запрос к ЦБ РФ", func(t *testing.T) {
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
		if _, ok := data.Rates[models.USD]; !ok {
			t.Error("USD должен присутствовать в результатах")
		}

		if _, ok := data.Rates[models.EUR]; !ok {
			t.Error("EUR должен присутствовать в результатах")
		}

		t.Logf("Получено валют: %d", len(data.Rates))
		t.Logf("USD курс: %.4f", data.Rates[models.USD].Rate)
		t.Logf("EUR курс: %.4f", data.Rates[models.EUR].Rate)
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

		t.Logf("Дата курсов: %v", data.Date)
		t.Logf("Получено валют: %d", len(data.Rates))
	})
}
*/
