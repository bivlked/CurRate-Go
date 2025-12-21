package parser

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/bivlked/currate-go/internal/models"
)

// TestParseXML_Success тестирует успешный парсинг валидного XML
func TestParseXML_Success(t *testing.T) {
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
        <Value>88,1234</Value>
    </Valute>
</ValCurs>`

	reader := strings.NewReader(xmlData)
	date := time.Date(2025, 12, 20, 0, 0, 0, 0, time.UTC)

	result, err := ParseXML(reader, date)
	if err != nil {
		t.Fatalf("ParseXML() error = %v, want nil", err)
	}

	if result == nil {
		t.Fatal("ParseXML() returned nil result")
	}

	// Проверяем дату
	if !result.Date.Equal(date) {
		t.Errorf("result.Date = %v, want %v", result.Date, date)
	}

	// Проверяем количество валют
	if len(result.Rates) != 2 {
		t.Errorf("len(result.Rates) = %d, want 2", len(result.Rates))
	}

	// Проверяем USD
	usdRate, ok := result.Rates[models.USD]
	if !ok {
		t.Fatal("USD rate not found")
	}
	if usdRate.Currency != models.USD {
		t.Errorf("USD currency = %v, want %v", usdRate.Currency, models.USD)
	}
	if usdRate.Rate != 80.7220 {
		t.Errorf("USD rate = %f, want %f", usdRate.Rate, 80.7220)
	}
	if usdRate.Nominal != 1 {
		t.Errorf("USD nominal = %d, want 1", usdRate.Nominal)
	}

	// Проверяем EUR
	eurRate, ok := result.Rates[models.EUR]
	if !ok {
		t.Fatal("EUR rate not found")
	}
	if eurRate.Rate != 88.1234 {
		t.Errorf("EUR rate = %f, want %f", eurRate.Rate, 88.1234)
	}
}

// TestParseXML_DifferentNominals тестирует обработку различных значений Nominal
func TestParseXML_DifferentNominals(t *testing.T) {
	tests := []struct {
		name    string
		nominal int
		value   string
		want    float64
	}{
		{
			name:    "Nominal 1 (USD)",
			nominal: 1,
			value:   "80,7220",
			want:    80.7220,
		},
		{
			name:    "Nominal 10 (DKK)",
			nominal: 10,
			value:   "118,3456",
			want:    118.3456,
		},
		{
			name:    "Nominal 100 (HUF)",
			nominal: 100,
			value:   "24,4161",
			want:    24.4161,
		},
		{
			name:    "Nominal 10000 (VND)",
			nominal: 10000,
			value:   "32,0988",
			want:    32.0988,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			xmlData := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<ValCurs Date="20.12.2025" name="Foreign Currency Market">
    <Valute ID="R01235">
        <NumCode>840</NumCode>
        <CharCode>USD</CharCode>
        <Nominal>%d</Nominal>
        <Name>Test Currency</Name>
        <Value>%s</Value>
    </Valute>
</ValCurs>`, tt.nominal, tt.value)

			reader := strings.NewReader(xmlData)
			date := time.Date(2025, 12, 20, 0, 0, 0, 0, time.UTC)

			result, err := ParseXML(reader, date)
			if err != nil {
				t.Fatalf("ParseXML() error = %v, want nil", err)
			}

			usdRate, ok := result.Rates[models.USD]
			if !ok {
				t.Fatal("USD rate not found")
			}

			if usdRate.Nominal != tt.nominal {
				t.Errorf("nominal = %d, want %d", usdRate.Nominal, tt.nominal)
			}

			if usdRate.Rate != tt.want {
				t.Errorf("rate = %f, want %f", usdRate.Rate, tt.want)
			}
		})
	}
}

// TestParseXML_InvalidXML тестирует обработку некорректного XML
func TestParseXML_InvalidXML(t *testing.T) {
	tests := []struct {
		name    string
		xmlData string
		wantErr error
	}{
		{
			name:    "Invalid XML syntax",
			xmlData: `<ValCurs><Valute>unclosed`,
			wantErr: ErrInvalidXML,
		},
		{
			name:    "Empty XML",
			xmlData: ``,
			wantErr: ErrInvalidXML,
		},
		{
			name: "No valutes",
			xmlData: `<?xml version="1.0" encoding="UTF-8"?>
<ValCurs Date="20.12.2025" name="Foreign Currency Market">
</ValCurs>`,
			wantErr: ErrNoXMLRates,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.xmlData)
			date := time.Date(2025, 12, 20, 0, 0, 0, 0, time.UTC)

			_, err := ParseXML(reader, date)
			if err == nil {
				t.Fatal("ParseXML() error = nil, want error")
			}

			// Проверяем тип ошибки
			if !strings.Contains(err.Error(), tt.wantErr.Error()) {
				t.Errorf("ParseXML() error = %v, want error containing %v", err, tt.wantErr)
			}
		})
	}
}

// TestParseXML_UnsupportedCurrency тестирует пропуск неподдерживаемых валют
func TestParseXML_UnsupportedCurrency(t *testing.T) {
	xmlData := `<?xml version="1.0" encoding="UTF-8"?>
<ValCurs Date="20.12.2025" name="Foreign Currency Market">
    <Valute ID="R01589">
        <NumCode>960</NumCode>
        <CharCode>XDR</CharCode>
        <Nominal>1</Nominal>
        <Name>СДР (специальные права заимствования)</Name>
        <Value>108,1234</Value>
    </Valute>
    <Valute ID="R01235">
        <NumCode>840</NumCode>
        <CharCode>USD</CharCode>
        <Nominal>1</Nominal>
        <Name>Доллар США</Name>
        <Value>80,7220</Value>
    </Valute>
</ValCurs>`

	reader := strings.NewReader(xmlData)
	date := time.Date(2025, 12, 20, 0, 0, 0, 0, time.UTC)

	result, err := ParseXML(reader, date)
	if err != nil {
		t.Fatalf("ParseXML() error = %v, want nil", err)
	}

	// XDR должен быть пропущен
	if _, ok := result.Rates[models.Currency("XDR")]; ok {
		t.Error("XDR should be skipped")
	}

	// USD должен присутствовать
	if _, ok := result.Rates[models.USD]; !ok {
		t.Error("USD should be present")
	}

	// Должна быть только 1 валюта (USD)
	if len(result.Rates) != 1 {
		t.Errorf("len(result.Rates) = %d, want 1", len(result.Rates))
	}
}

// TestParseXMLValue тестирует функцию parseXMLValue
func TestParseXMLValue(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    float64
		wantErr bool
	}{
		{
			name:    "Valid value with comma",
			input:   "80,7220",
			want:    80.7220,
			wantErr: false,
		},
		{
			name:    "Valid value with dot",
			input:   "80.7220",
			want:    80.7220,
			wantErr: false,
		},
		{
			name:    "Valid value with spaces",
			input:   "  80,7220  ",
			want:    80.7220,
			wantErr: false,
		},
		{
			name:    "Empty string",
			input:   "",
			want:    0,
			wantErr: true,
		},
		{
			name:    "Invalid format",
			input:   "abc",
			want:    0,
			wantErr: true,
		},
		{
			name:    "Negative value",
			input:   "-80,7220",
			want:    0,
			wantErr: true,
		},
		{
			name:    "Zero value",
			input:   "0",
			want:    0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseXMLValue(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseXMLValue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("parseXMLValue() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestParseXML_InvalidNominal тестирует обработку некорректного Nominal
func TestParseXML_InvalidNominal(t *testing.T) {
	xmlData := `<?xml version="1.0" encoding="UTF-8"?>
<ValCurs Date="20.12.2025" name="Foreign Currency Market">
    <Valute ID="R01235">
        <NumCode>840</NumCode>
        <CharCode>USD</CharCode>
        <Nominal>0</Nominal>
        <Name>Доллар США</Name>
        <Value>80,7220</Value>
    </Valute>
</ValCurs>`

	reader := strings.NewReader(xmlData)
	date := time.Date(2025, 12, 20, 0, 0, 0, 0, time.UTC)

	result, err := ParseXML(reader, date)

	// Должна быть ошибка, так как нет валидных валют
	if err == nil {
		t.Fatal("ParseXML() error = nil, want error")
	}

	if result != nil {
		t.Errorf("ParseXML() result should be nil when no valid currencies")
	}
}
