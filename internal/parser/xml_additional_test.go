package parser

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/bivlked/currate-go/internal/models"
	"golang.org/x/text/encoding/charmap"
)

func TestParseXML_Windows1251Encoding(t *testing.T) {
	fallbackDate := testPastDateUTC()
	xmlDate := fallbackDate.AddDate(0, 0, -1)
	xmlDateStr := formatCBRDate(xmlDate)
	rawXML := fmt.Sprintf(`<?xml version="1.0" encoding="windows-1251"?>
<ValCurs Date="%s" name="Foreign Currency Market">
    <Valute ID="R01235">
        <NumCode>840</NumCode>
        <CharCode>USD</CharCode>
        <Nominal>1</Nominal>
        <Name>Доллар США</Name>
        <Value>80,7220</Value>
    </Valute>
</ValCurs>`, xmlDateStr)

	encoded, err := charmap.Windows1251.NewEncoder().Bytes([]byte(rawXML))
	if err != nil {
		t.Fatalf("Ошибка кодирования Windows-1251: %v", err)
	}

	data, err := ParseXML(bytes.NewReader(encoded), fallbackDate)
	if err != nil {
		t.Fatalf("ParseXML() ошибка: %v", err)
	}

	// Проверяем дату (сравниваем только календарную дату, игнорируя время)
	// ParseXML нормализует дату к началу дня, поэтому сравниваем только Year, Month, Day
	if data.Date.Year() != xmlDate.Year() || data.Date.Month() != xmlDate.Month() || data.Date.Day() != xmlDate.Day() {
		t.Fatalf("Дата курсов должна быть взята из XML, получено: %v (календарная дата %d.%d.%d), ожидалось календарная дата %d.%d.%d",
			data.Date, data.Date.Year(), data.Date.Month(), data.Date.Day(),
			xmlDate.Year(), xmlDate.Month(), xmlDate.Day())
	}

	usdRate, ok := data.Rates[models.USD]
	if !ok {
		t.Fatal("USD должен присутствовать в результатах")
	}
	if usdRate.Rate != 80.7220 {
		t.Errorf("USD курс: ожидалось 80.7220, получено %v", usdRate.Rate)
	}
}

func TestParseXML_SkipsUnsupportedAndInvalidRates(t *testing.T) {
	date := testPastDateUTC()
	dateStr := formatCBRDate(date)
	xmlData := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
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
        <Value>abc</Value>
    </Valute>
    <Valute ID="R01235">
        <NumCode>960</NumCode>
        <CharCode>XDR</CharCode>
        <Nominal>1</Nominal>
        <Name>СДР</Name>
        <Value>10,1234</Value>
    </Valute>
    <Valute ID="R01227">
        <NumCode>392</NumCode>
        <CharCode>JPY</CharCode>
        <Nominal>0</Nominal>
        <Name>Японская иена</Name>
        <Value>1,2345</Value>
    </Valute>
</ValCurs>`, dateStr)

	data, err := ParseXML(strings.NewReader(xmlData), date)
	if err != nil {
		t.Fatalf("ParseXML() ошибка: %v", err)
	}

	if len(data.Rates) != 1 {
		t.Fatalf("Ожидалась только 1 валюта после фильтрации, получено %d", len(data.Rates))
	}

	if _, ok := data.Rates[models.USD]; !ok {
		t.Fatal("USD должен остаться после фильтрации")
	}
}

func TestParseXML_NoRates(t *testing.T) {
	date := testPastDateUTC()
	dateStr := formatCBRDate(date)
	xmlData := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<ValCurs Date="%s" name="Foreign Currency Market"></ValCurs>`, dateStr)

	_, err := ParseXML(strings.NewReader(xmlData), date)
	if !errors.Is(err, ErrNoXMLRates) {
		t.Fatalf("Ожидалась ошибка ErrNoXMLRates, получено: %v", err)
	}
}
