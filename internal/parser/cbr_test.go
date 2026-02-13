package parser

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/bivlked/currate-go/internal/models"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}

func setTestHTTPClientFactory(t *testing.T, rt http.RoundTripper) {
	t.Helper()
	originalClient := defaultHTTPClient
	defaultHTTPClient = &http.Client{Transport: rt}
	t.Cleanup(func() {
		defaultHTTPClient = originalClient
	})
}

func newResponse(req *http.Request, statusCode int, body string) *http.Response {
	statusText := http.StatusText(statusCode)
	status := fmt.Sprintf("%d %s", statusCode, statusText)
	if statusText == "" {
		status = fmt.Sprintf("%d", statusCode)
	}

	return &http.Response{
		StatusCode: statusCode,
		Status:     status,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(body)),
		Request:    req,
	}
}

// Unit тесты с мок-сервером
func TestFetchRates(t *testing.T) {
	testDate := testPastDateUTC()
	dateStr := formatCBRDate(testDate)
	// Мок XML с реального API ЦБ РФ (упрощенная версия)
	mockXML := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<ValCurs Date="%s" name="Foreign Currency Market">
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
</ValCurs>`, dateStr)

	t.Run("Успешное получение курсов", func(t *testing.T) {
		setTestHTTPClientFactory(t, roundTripFunc(func(req *http.Request) (*http.Response, error) {
			return newResponse(req, http.StatusOK, mockXML), nil
		}))

		data, err := fetchRatesFromURL("http://example.test", testDate)
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
		setTestHTTPClientFactory(t, roundTripFunc(func(req *http.Request) (*http.Response, error) {
			return newResponse(req, http.StatusInternalServerError, ""), nil
		}))

		data, err := fetchRatesFromURL("http://example.test", testDate)
		if err == nil {
			t.Fatal("Ожидалась ошибка для статуса 500")
		}

		if data != nil {
			t.Error("Data должна быть nil при ошибке")
		}
	})

	t.Run("Невалидный XML", func(t *testing.T) {
		setTestHTTPClientFactory(t, roundTripFunc(func(req *http.Request) (*http.Response, error) {
			return newResponse(req, http.StatusOK, "<html><body>Not XML response</body></html>"), nil
		}))

		data, err := fetchRatesFromURL("http://example.test", testDate)
		if err == nil {
			t.Fatal("Ожидалась ошибка для невалидного XML")
		}

		if data != nil {
			t.Error("Data должна быть nil при ошибке")
		}
	})

	t.Run("Ошибка парсинга XML", func(t *testing.T) {
		setTestHTTPClientFactory(t, roundTripFunc(func(req *http.Request) (*http.Response, error) {
			return newResponse(req, http.StatusOK, fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<ValCurs Date="%s" name="Foreign Currency Market">
</ValCurs>`, dateStr)), nil
		}))

		data, err := fetchRatesFromURL("http://example.test", testDate)
		if err == nil {
			t.Fatal("Ожидалась ошибка для XML без валют")
		}

		if data != nil {
			t.Error("Data должна быть nil при ошибке")
		}
	})
}

// TestFetchRates_WithBuildURL проверяет обработку ошибки парсинга XML
func TestFetchRates_ParseError(t *testing.T) {
	testDate := testPastDateUTC()
	dateStr := formatCBRDate(testDate)
	setTestHTTPClientFactory(t, roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return newResponse(req, http.StatusOK, fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<ValCurs Date="%s" name="Foreign Currency Market">
</ValCurs>`, dateStr)), nil
	}))

	data, err := fetchRatesFromURL("http://example.test", testDate)
	if err == nil {
		t.Fatal("Ожидалась ошибка для XML без валют")
	}

	if data != nil {
		t.Error("Data должна быть nil при ошибке")
	}

	// Проверяем, что ошибка содержит информацию о парсинге
	if !strings.Contains(err.Error(), "parse") && !strings.Contains(err.Error(), "rates") {
		t.Errorf("Ошибка должна содержать информацию о парсинге или rates, получена: %v", err)
	}
}

func TestFetchRates_UsesBuildURLAndParsesResponse(t *testing.T) {
	mockXML := `<?xml version="1.0" encoding="UTF-8"?>
<ValCurs Date="20.12.2025" name="Foreign Currency Market">
    <Valute ID="R01235">
        <NumCode>840</NumCode>
        <CharCode>USD</CharCode>
        <Nominal>1</Nominal>
        <Name>Доллар США</Name>
        <Value>80,7220</Value>
    </Valute>
</ValCurs>`

	oldTransport := http.DefaultTransport
	t.Cleanup(func() { http.DefaultTransport = oldTransport })

	var capturedQuery url.Values
	http.DefaultTransport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.Method != http.MethodGet {
			t.Fatalf("ожидался GET, получен %s", req.Method)
		}
		if req.URL.Host != "www.cbr.ru" {
			t.Fatalf("ожидался запрос к www.cbr.ru, получен %s", req.URL.Host)
		}
		capturedQuery = req.URL.Query()
		return &http.Response{
			StatusCode: http.StatusOK,
			Status:     "200 OK",
			Body:       io.NopCloser(strings.NewReader(mockXML)),
			Header:     make(http.Header),
		}, nil
	})

	requestDate := time.Date(2025, 12, 20, 0, 0, 0, 0, time.UTC)
	data, err := FetchRates(requestDate)
	if err != nil {
		t.Fatalf("FetchRates() error = %v", err)
	}

	if capturedQuery.Get("date_req") != requestDate.Format("02/01/2006") {
		t.Fatalf("date_req = %s, want %s", capturedQuery.Get("date_req"), requestDate.Format("02/01/2006"))
	}

	if data == nil {
		t.Fatal("FetchRates() result is nil")
	}

	usd, ok := data.Rates[models.USD]
	if !ok {
		t.Fatal("USD курс не найден")
	}
	if usd.Rate != 80.7220 {
		t.Errorf("USD курс = %v, want 80.7220", usd.Rate)
	}
}

func TestFetchLatestRates_UsesCurrentDateAndParsesXMLDate(t *testing.T) {
	mockXML := `<?xml version="1.0" encoding="UTF-8"?>
<ValCurs Date="19.12.2025" name="Foreign Currency Market">
    <Valute ID="R01239">
        <NumCode>978</NumCode>
        <CharCode>EUR</CharCode>
        <Nominal>1</Nominal>
        <Name>Евро</Name>
        <Value>94,5120</Value>
    </Valute>
</ValCurs>`

	oldTransport := http.DefaultTransport
	t.Cleanup(func() { http.DefaultTransport = oldTransport })

	http.DefaultTransport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.URL.Query().Get("date_req") == "" {
			t.Fatal("date_req не должен быть пустым")
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Status:     "200 OK",
			Body:       io.NopCloser(strings.NewReader(mockXML)),
			Header:     make(http.Header),
		}, nil
	})

	data, err := FetchLatestRates()
	if err != nil {
		t.Fatalf("FetchLatestRates() error = %v", err)
	}

	if data == nil {
		t.Fatal("FetchLatestRates() result is nil")
	}

	// Проверяем календарную дату (год, месяц, день) независимо от временной зоны
	// ParseXML использует time.ParseInLocation с локальной зоной, поэтому сравниваем только календарные даты
	expectedDate := time.Date(2025, 12, 19, 0, 0, 0, 0, time.UTC)
	if data.Date.Year() != expectedDate.Year() ||
		data.Date.Month() != expectedDate.Month() ||
		data.Date.Day() != expectedDate.Day() {
		t.Errorf("data.Date = %v (календарная дата %d.%d.%d), want календарная дата %d.%d.%d",
			data.Date, data.Date.Year(), data.Date.Month(), data.Date.Day(),
			expectedDate.Year(), expectedDate.Month(), expectedDate.Day())
	}

	if _, ok := data.Rates[models.EUR]; !ok {
		t.Fatal("EUR курс не найден")
	}
}
