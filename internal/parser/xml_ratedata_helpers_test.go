package parser

import (
	"strings"
	"testing"
	"time"

	"github.com/bivlked/currate-go/internal/models"
)

// TestParseXML_UsesNewRateData проверяет, что ParseXML использует NewRateData в runtime
func TestParseXML_UsesNewRateData(t *testing.T) {
	date := testPastDateUTC()
	dateStr := formatCBRDate(date)
	xmlData := `<?xml version="1.0" encoding="UTF-8"?>
<ValCurs Date="` + dateStr + `" name="Foreign Currency Market">
    <Valute ID="R01235">
        <NumCode>840</NumCode>
        <CharCode>USD</CharCode>
        <Nominal>1</Nominal>
        <Name>Доллар США</Name>
        <Value>80,7220</Value>
    </Valute>
</ValCurs>`

	reader := strings.NewReader(xmlData)
	result, err := ParseXML(reader, date)
	if err != nil {
		t.Fatalf("ParseXML() error = %v, want nil", err)
	}

	if result == nil {
		t.Fatal("ParseXML() returned nil result")
	}

	// Проверяем, что RateData создан корректно (через NewRateData)
	if result.Rates == nil {
		t.Error("ParseXML() result.Rates is nil, should be initialized by NewRateData")
	}

	// Проверяем, что дата установлена корректно
	// ParseXML использует дату из XML (DD.MM.YYYY), которая нормализуется к полуночи
	expectedDate := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	if !result.Date.Equal(expectedDate) {
		t.Errorf("ParseXML() result.Date = %v, want %v", result.Date, expectedDate)
	}
}

// TestParseXML_UsesAddRate проверяет, что ParseXML использует AddRate в runtime
func TestParseXML_UsesAddRate(t *testing.T) {
	date := testPastDateUTC()
	dateStr := formatCBRDate(date)
	xmlData := `<?xml version="1.0" encoding="UTF-8"?>
<ValCurs Date="` + dateStr + `" name="Foreign Currency Market">
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
        <Value>90,1234</Value>
    </Valute>
</ValCurs>`

	reader := strings.NewReader(xmlData)
	result, err := ParseXML(reader, date)
	if err != nil {
		t.Fatalf("ParseXML() error = %v, want nil", err)
	}

	if result == nil {
		t.Fatal("ParseXML() returned nil result")
	}

	// Проверяем, что курсы добавлены через AddRate
	if len(result.Rates) != 2 {
		t.Errorf("ParseXML() result.Rates len = %d, want 2", len(result.Rates))
	}

	// Проверяем, что USD добавлен корректно
	usdRate, exists := result.Rates[models.USD]
	if !exists {
		t.Error("ParseXML() USD rate not found, AddRate should have added it")
	}
	if usdRate.Rate != 80.7220 {
		t.Errorf("ParseXML() USD rate = %v, want 80.7220", usdRate.Rate)
	}

	// Проверяем, что EUR добавлен корректно
	eurRate, exists := result.Rates[models.EUR]
	if !exists {
		t.Error("ParseXML() EUR rate not found, AddRate should have added it")
	}
	if eurRate.Rate != 90.1234 {
		t.Errorf("ParseXML() EUR rate = %v, want 90.1234", eurRate.Rate)
	}
}

// TestParseXML_RateDataHelpersIntegration проверяет интеграцию NewRateData и AddRate в runtime
func TestParseXML_RateDataHelpersIntegration(t *testing.T) {
	date := testPastDateUTC()
	dateStr := formatCBRDate(date)
	xmlData := `<?xml version="1.0" encoding="UTF-8"?>
<ValCurs Date="` + dateStr + `" name="Foreign Currency Market">
    <Valute ID="R01235">
        <NumCode>840</NumCode>
        <CharCode>USD</CharCode>
        <Nominal>1</Nominal>
        <Name>Доллар США</Name>
        <Value>80,7220</Value>
    </Valute>
</ValCurs>`

	reader := strings.NewReader(xmlData)
	result, err := ParseXML(reader, date)
	if err != nil {
		t.Fatalf("ParseXML() error = %v, want nil", err)
	}

	// Проверяем, что можно использовать GetRate для получения курса
	// Это проверяет, что структура RateData создана корректно через NewRateData
	usdRate, found := result.GetRate(models.USD)
	if !found {
		t.Error("GetRate() USD not found, but should be added via AddRate")
	}
	if usdRate.Rate != 80.7220 {
		t.Errorf("GetRate() USD rate = %v, want 80.7220", usdRate.Rate)
	}
	if usdRate.Currency != models.USD {
		t.Errorf("GetRate() USD currency = %v, want USD", usdRate.Currency)
	}
	if usdRate.Nominal != 1 {
		t.Errorf("GetRate() USD nominal = %d, want 1", usdRate.Nominal)
	}
	// ParseXML использует дату из XML (DD.MM.YYYY), которая нормализуется к полуночи
	expectedDate := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	if !usdRate.Date.Equal(expectedDate) {
		t.Errorf("GetRate() USD date = %v, want %v", usdRate.Date, expectedDate)
	}
}
