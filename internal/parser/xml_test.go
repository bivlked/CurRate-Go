package parser

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/bivlked/currate-go/internal/models"
)

// TestParseXML_Success тестирует успешный парсинг валидного XML
func TestParseXML_Success(t *testing.T) {
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
        <Value>88,1234</Value>
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

	// Проверяем дату (сравниваем только календарную дату, игнорируя время)
	// ParseXML нормализует дату к началу дня, поэтому сравниваем только Year, Month, Day
	if result.Date.Year() != date.Year() || result.Date.Month() != date.Month() || result.Date.Day() != date.Day() {
		t.Errorf("result.Date = %v (календарная дата %d.%d.%d), want календарная дата %d.%d.%d",
			result.Date, result.Date.Year(), result.Date.Month(), result.Date.Day(),
			date.Year(), date.Month(), date.Day())
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
			date := testPastDateUTC()
			dateStr := formatCBRDate(date)
			xmlData := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<ValCurs Date="%s" name="Foreign Currency Market">
    <Valute ID="R01235">
        <NumCode>840</NumCode>
        <CharCode>USD</CharCode>
        <Nominal>%d</Nominal>
        <Name>Test Currency</Name>
        <Value>%s</Value>
    </Valute>
			</ValCurs>`, dateStr, tt.nominal, tt.value)

			reader := strings.NewReader(xmlData)
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
<ValCurs Date="%s" name="Foreign Currency Market">
</ValCurs>`,
			wantErr: ErrNoXMLRates,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			date := testPastDateUTC()
			dateStr := formatCBRDate(date)
			xmlData := fmt.Sprintf(tt.xmlData, dateStr)
			reader := strings.NewReader(xmlData)

			_, err := ParseXML(reader, date)
			if err == nil {
				t.Fatal("ParseXML() error = nil, want error")
			}

			// Проверяем тип ошибки через errors.Is (sentinel errors обёрнуты через %w)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("ParseXML() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

// TestParseXML_UnsupportedCurrency тестирует пропуск неподдерживаемых валют
func TestParseXML_UnsupportedCurrency(t *testing.T) {
	date := testPastDateUTC()
	dateStr := formatCBRDate(date)
	xmlData := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<ValCurs Date="%s" name="Foreign Currency Market">
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
</ValCurs>`, dateStr)

	reader := strings.NewReader(xmlData)

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

// TestParseXML_UsesXMLDate проверяет, что дата берется из XML, если она указана
func TestParseXML_UsesXMLDate(t *testing.T) {
	fallbackDate := testPastDateUTC()
	xmlDate := fallbackDate.AddDate(0, 0, -1)
	xmlDateStr := formatCBRDate(xmlDate)
	xmlData := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<ValCurs Date="%s" name="Foreign Currency Market">
    <Valute ID="R01235">
        <NumCode>840</NumCode>
        <CharCode>USD</CharCode>
        <Nominal>1</Nominal>
        <Name>Доллар США</Name>
        <Value>80,7220</Value>
    </Valute>
</ValCurs>`, xmlDateStr)

	reader := strings.NewReader(xmlData)

	result, err := ParseXML(reader, fallbackDate)
	if err != nil {
		t.Fatalf("ParseXML() error = %v, want nil", err)
	}

	// Проверяем дату (сравниваем только календарную дату, игнорируя время)
	// ParseXML нормализует дату к началу дня, поэтому сравниваем только Year, Month, Day
	if result.Date.Year() != xmlDate.Year() || result.Date.Month() != xmlDate.Month() || result.Date.Day() != xmlDate.Day() {
		t.Errorf("result.Date = %v (календарная дата %d.%d.%d), want календарная дата %d.%d.%d",
			result.Date, result.Date.Year(), result.Date.Month(), result.Date.Day(),
			xmlDate.Year(), xmlDate.Month(), xmlDate.Day())
	}
}

// TestParseXML_NoSupportedRates проверяет, что при отсутствии валидных курсов возвращается ошибка
func TestParseXML_NoSupportedRates(t *testing.T) {
	date := testPastDateUTC()
	dateStr := formatCBRDate(date)
	xmlData := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<ValCurs Date="%s" name="Foreign Currency Market">
    <Valute ID="R01589">
        <NumCode>960</NumCode>
        <CharCode>XDR</CharCode>
        <Nominal>1</Nominal>
        <Name>Special Drawing Rights</Name>
        <Value>12,3456</Value>
    </Valute>
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
        <Value>invalid</Value>
    </Valute>
</ValCurs>`, dateStr)

	reader := strings.NewReader(xmlData)

	_, err := ParseXML(reader, date)
	if err == nil {
		t.Fatal("Ожидалась ошибка ErrNoXMLRates")
	}

	if !errors.Is(err, ErrNoXMLRates) {
		t.Errorf("Ожидалась ошибка ErrNoXMLRates, получена %v", err)
	}
}

func TestParseXMLValue_InvalidRate(t *testing.T) {
	_, err := parseXMLValue("abc")
	if err == nil {
		t.Fatal("Ожидалась ошибка для некорректного курса")
	}

	if !errors.Is(err, ErrInvalidXMLRate) {
		t.Errorf("Ожидалась ошибка ErrInvalidXMLRate, получена %v", err)
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
</ValCurs>`, dateStr)

	reader := strings.NewReader(xmlData)

	result, err := ParseXML(reader, date)

	// Должна быть ошибка, так как нет валидных валют
	if err == nil {
		t.Fatal("ParseXML() error = nil, want error")
	}

	if result != nil {
		t.Errorf("ParseXML() result should be nil when no valid currencies")
	}
}

// TestParseXML_InvalidDateInXML проверяет fallback на переданную дату при некорректной дате в XML
func TestParseXML_InvalidDateInXML(t *testing.T) {
	xmlData := `<?xml version="1.0" encoding="UTF-8"?>
<ValCurs Date="invalid-date" name="Foreign Currency Market">
    <Valute ID="R01235">
        <NumCode>840</NumCode>
        <CharCode>USD</CharCode>
        <Nominal>1</Nominal>
        <Name>Доллар США</Name>
        <Value>80,7220</Value>
    </Valute>
</ValCurs>`

	reader := strings.NewReader(xmlData)
	fallbackDate := testPastDateUTC()

	result, err := ParseXML(reader, fallbackDate)
	if err != nil {
		t.Fatalf("ParseXML() error = %v, want nil", err)
	}

	// Должна использоваться fallback дата, так как дата в XML некорректна
	if !result.Date.Equal(fallbackDate) {
		t.Errorf("result.Date = %v, want %v (fallback date)", result.Date, fallbackDate)
	}
}

// TestParseXMLValue_NonErrInvalidRateError проверяет обработку ошибки, не являющейся ErrInvalidRate
func TestParseXMLValue_NonErrInvalidRateError(t *testing.T) {
	originalParseRate := parseRateFunc
	t.Cleanup(func() {
		parseRateFunc = originalParseRate
	})

	sentinelErr := errors.New("sentinel parse rate error")
	parseRateFunc = func(string) (float64, error) {
		return 0, sentinelErr
	}

	_, err := parseXMLValue("80,7220")
	if err == nil {
		t.Fatal("parseXMLValue() error = nil, want sentinel error")
	}
	if !errors.Is(err, sentinelErr) {
		t.Fatalf("parseXMLValue() error = %v, want %v", err, sentinelErr)
	}
}

// TestParseXML_ReadError проверяет обработку ошибки чтения XML
func TestParseXML_ReadError(t *testing.T) {
	// Создаем reader, который возвращает ошибку при чтении
	errorReader := &errorReader{err: fmt.Errorf("read error")}
	date := testPastDateUTC()

	_, err := ParseXML(errorReader, date)
	if err == nil {
		t.Fatal("ParseXML() error = nil, want error")
	}

	if !strings.Contains(err.Error(), "failed to read XML") {
		t.Errorf("ParseXML() error = %v, want error containing 'failed to read XML'", err)
	}
}

// errorReader - io.Reader, который всегда возвращает ошибку
type errorReader struct {
	err error
}

func (e *errorReader) Read(p []byte) (n int, err error) {
	return 0, e.err
}

// TestParseXML_OversizedResponse проверяет, что XML больше maxXMLSize возвращает ErrXMLTooLarge
func TestParseXML_OversizedResponse(t *testing.T) {
	date := testPastDateUTC()
	dateStr := formatCBRDate(date)

	// Создаём валидный XML-заголовок
	header := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<ValCurs Date="%s" name="Foreign Currency Market">
    <Valute ID="R01235">
        <NumCode>840</NumCode>
        <CharCode>USD</CharCode>
        <Nominal>1</Nominal>
        <Name>Доллар США</Name>
        <Value>80,7220</Value>
    </Valute>`, dateStr)

	// Добавляем padding чтобы превысить maxXMLSize (4 MB)
	padding := strings.Repeat(" ", maxXMLSize)
	footer := `</ValCurs>`
	xmlData := header + padding + footer

	reader := strings.NewReader(xmlData)
	_, err := ParseXML(reader, date)
	if err == nil {
		t.Fatal("ParseXML() error = nil, want error for oversized XML")
	}
	if !errors.Is(err, ErrXMLTooLarge) {
		t.Errorf("ParseXML() error = %v, want ErrXMLTooLarge", err)
	}
}
