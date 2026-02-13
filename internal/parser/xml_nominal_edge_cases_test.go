package parser

import (
	"fmt"
	"strings"
	"testing"

	"github.com/bivlked/currate-go/internal/models"
)

// TestParseXML_InvalidNominalValues проверяет обработку различных некорректных значений номинала
// Проблема 7: XML Nominal - обработка некорректных значений
func TestParseXML_InvalidNominalValues(t *testing.T) {
	date := testPastDateUTC()
	dateStr := formatCBRDate(date)

	tests := []struct {
		name          string
		nominalValue  string
		shouldSkip    bool // Должна ли валюта быть пропущена
		expectedRates int  // Ожидаемое количество валидных курсов
		expectedError bool // Ожидается ли ошибка (если нет валидных валют)
	}{
		{
			name:          "Номинал = 0 (некорректный)",
			nominalValue:  "0",
			shouldSkip:    true,
			expectedRates: 0,
			expectedError: true, // Нет валидных валют
		},
		{
			name:          "Номинал = отрицательное число",
			nominalValue:  "-1",
			shouldSkip:    true,
			expectedRates: 0,
			expectedError: true,
		},
		{
			name:          "Номинал = пустая строка",
			nominalValue:  "",
			shouldSkip:    true,
			expectedRates: 0,
			expectedError: true,
		},
		{
			name:          "Номинал = не число",
			nominalValue:  "abc",
			shouldSkip:    true,
			expectedRates: 0,
			expectedError: true,
		},
		{
			name:          "Номинал = дробное число",
			nominalValue:  "1.5",
			shouldSkip:    true,
			expectedRates: 0,
			expectedError: true,
		},
		{
			name:          "Номинал = валидный (1)",
			nominalValue:  "1",
			shouldSkip:    false,
			expectedRates: 1,
			expectedError: false,
		},
		{
			name:          "Номинал = валидный (10)",
			nominalValue:  "10",
			shouldSkip:    false,
			expectedRates: 1,
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			xmlData := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<ValCurs Date="%s" name="Foreign Currency Market">
    <Valute ID="R01235">
        <NumCode>840</NumCode>
        <CharCode>USD</CharCode>
        <Nominal>%s</Nominal>
        <Name>Доллар США</Name>
        <Value>80,7220</Value>
    </Valute>
</ValCurs>`, dateStr, tt.nominalValue)

			reader := strings.NewReader(xmlData)
			result, err := ParseXML(reader, date)

			if tt.expectedError {
				if err == nil {
					t.Errorf("ParseXML() error = nil, want error (no valid currencies)")
				}
				if result != nil {
					t.Errorf("ParseXML() result should be nil when no valid currencies")
				}
			} else {
				if err != nil {
					t.Errorf("ParseXML() error = %v, want nil", err)
				}
				if result == nil {
					t.Fatal("ParseXML() returned nil result")
				}
				if len(result.Rates) != tt.expectedRates {
					t.Errorf("ParseXML() result.Rates len = %d, want %d", len(result.Rates), tt.expectedRates)
				}
			}
		})
	}
}

// TestParseXML_MixedValidAndInvalidNominals проверяет, что валидные валюты обрабатываются,
// а невалидные пропускаются
func TestParseXML_MixedValidAndInvalidNominals(t *testing.T) {
	date := testPastDateUTC()
	dateStr := formatCBRDate(date)

	xmlData := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<ValCurs Date="%s" name="Foreign Currency Market">
    <Valute ID="R01235">
        <NumCode>840</NumCode>
        <CharCode>USD</CharCode>
        <Nominal>0</Nominal>
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
    <Valute ID="R01236">
        <NumCode>826</NumCode>
        <CharCode>XXX</CharCode>
        <Nominal>-1</Nominal>
        <Name>Неподдерживаемая валюта</Name>
        <Value>100,5678</Value>
    </Valute>
</ValCurs>`, dateStr)

	reader := strings.NewReader(xmlData)
	result, err := ParseXML(reader, date)

	if err != nil {
		t.Fatalf("ParseXML() error = %v, want nil", err)
	}

	if result == nil {
		t.Fatal("ParseXML() returned nil result")
	}

	// Должна быть только одна валидная валюта (EUR с номиналом 1)
	if len(result.Rates) != 1 {
		t.Errorf("ParseXML() result.Rates len = %d, want 1 (only EUR should be valid)", len(result.Rates))
	}

	// Проверяем, что EUR присутствует
	eurRate, exists := result.Rates[models.EUR]
	if !exists {
		t.Error("ParseXML() EUR rate not found, but should be valid")
	}
	if eurRate.Rate != 90.1234 {
		t.Errorf("ParseXML() EUR rate = %v, want 90.1234", eurRate.Rate)
	}
	if eurRate.Nominal != 1 {
		t.Errorf("ParseXML() EUR nominal = %d, want 1", eurRate.Nominal)
	}

	// Проверяем, что USD отсутствует (некорректный номинал)
	if _, exists := result.Rates[models.USD]; exists {
		t.Error("ParseXML() USD rate should not be present (invalid nominal 0)")
	}
}

// TestParseXML_NominalWithWhitespace проверяет обработку номинала с пробелами
func TestParseXML_NominalWithWhitespace(t *testing.T) {
	date := testPastDateUTC()
	dateStr := formatCBRDate(date)

	xmlData := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<ValCurs Date="%s" name="Foreign Currency Market">
    <Valute ID="R01235">
        <NumCode>840</NumCode>
        <CharCode>USD</CharCode>
        <Nominal>  1  </Nominal>
        <Name>Доллар США</Name>
        <Value>80,7220</Value>
    </Valute>
</ValCurs>`, dateStr)

	reader := strings.NewReader(xmlData)
	result, err := ParseXML(reader, date)

	if err != nil {
		t.Fatalf("ParseXML() error = %v, want nil", err)
	}

	if result == nil {
		t.Fatal("ParseXML() returned nil result")
	}

	// Номинал с пробелами должен быть обработан (trimmed)
	usdRate, exists := result.Rates[models.USD]
	if !exists {
		t.Error("ParseXML() USD rate not found")
	}
	if usdRate.Nominal != 1 {
		t.Errorf("ParseXML() USD nominal = %d, want 1 (whitespace should be trimmed)", usdRate.Nominal)
	}
}
