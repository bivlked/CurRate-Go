package parser

import (
	"bytes"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/bivlked/currate-go/internal/models"
	"golang.org/x/text/encoding/charmap"
)

func TestParseXML_Windows1251Encoding(t *testing.T) {
	rawXML := `<?xml version="1.0" encoding="windows-1251"?>
<ValCurs Date="20.12.2025" name="Foreign Currency Market">
    <Valute ID="R01235">
        <NumCode>840</NumCode>
        <CharCode>USD</CharCode>
        <Nominal>1</Nominal>
        <Name>Доллар США</Name>
        <Value>80,7220</Value>
    </Valute>
</ValCurs>`

	encoded, err := charmap.Windows1251.NewEncoder().Bytes([]byte(rawXML))
	if err != nil {
		t.Fatalf("Ошибка кодирования Windows-1251: %v", err)
	}

	date := time.Date(2025, 12, 21, 10, 0, 0, 0, time.UTC)
	data, err := ParseXML(bytes.NewReader(encoded), date)
	if err != nil {
		t.Fatalf("ParseXML() ошибка: %v", err)
	}

	if data.Date.Day() != 20 || data.Date.Month() != time.December || data.Date.Year() != 2025 {
		t.Fatalf("Дата курсов должна быть взята из XML, получено: %v", data.Date)
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
	xmlData := `<?xml version="1.0" encoding="UTF-8"?>
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
</ValCurs>`

	data, err := ParseXML(strings.NewReader(xmlData), time.Date(2025, 12, 20, 0, 0, 0, 0, time.UTC))
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
	xmlData := `<?xml version="1.0" encoding="UTF-8"?>
<ValCurs Date="20.12.2025" name="Foreign Currency Market"></ValCurs>`

	_, err := ParseXML(strings.NewReader(xmlData), time.Date(2025, 12, 20, 0, 0, 0, 0, time.UTC))
	if !errors.Is(err, ErrNoXMLRates) {
		t.Fatalf("Ожидалась ошибка ErrNoXMLRates, получено: %v", err)
	}
}
