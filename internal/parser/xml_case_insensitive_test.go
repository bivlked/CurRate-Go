package parser

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/bivlked/currate-go/internal/models"
	"golang.org/x/text/encoding/charmap"
)

// TestParseXML_Windows1251CaseInsensitive проверяет обработку XML с разными регистрами windows-1251
func TestParseXML_Windows1251CaseInsensitive(t *testing.T) {
	testCases := []struct {
		name     string
		encoding string
	}{
		{
			name:     "windows-1251 в нижнем регистре",
			encoding: "windows-1251",
		},
		{
			name:     "Windows-1251 с заглавной буквой",
			encoding: "Windows-1251",
		},
		{
			name:     "WINDOWS-1251 в верхнем регистре",
			encoding: "WINDOWS-1251",
		},
		{
			name:     "WinDows-1251 смешанный регистр",
			encoding: "WinDows-1251",
		},
		{
			name:     "WiNdOwS-1251 чередующийся регистр",
			encoding: "WiNdOwS-1251",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Создаем XML с указанной кодировкой
			xmlTemplate := `<?xml version="1.0" encoding="%s"?>
<ValCurs Date="20.12.2025" name="Foreign Currency Market">
    <Valute ID="R01235">
        <NumCode>840</NumCode>
        <CharCode>USD</CharCode>
        <Nominal>1</Nominal>
        <Name>Доллар США</Name>
        <Value>80,7220</Value>
    </Valute>
</ValCurs>`

			xmlData := []byte(strings.Replace(xmlTemplate, "%s", tc.encoding, 1))

			// Кодируем в windows-1251
			encoder := charmap.Windows1251.NewEncoder()
			encodedXML, err := encoder.Bytes(xmlData)
			if err != nil {
				t.Fatalf("Не удалось закодировать XML в windows-1251: %v", err)
			}

			reader := bytes.NewReader(encodedXML)
			date := time.Date(2025, 12, 20, 0, 0, 0, 0, time.UTC)

			result, err := ParseXML(reader, date)
			if err != nil {
				t.Fatalf("ParseXML() error = %v, want nil", err)
			}

			if result == nil {
				t.Fatal("ParseXML() returned nil result")
			}

			// Проверяем, что USD найден
			usdRate, exists := result.Rates[models.USD]
			if !exists {
				t.Fatal("USD rate not found after windows-1251 decoding")
			}

			// Проверяем курс
			if usdRate.Rate != 80.7220 {
				t.Errorf("USD rate = %v, want 80.7220", usdRate.Rate)
			}
		})
	}
}
