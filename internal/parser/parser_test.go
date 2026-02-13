package parser

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/bivlked/currate-go/internal/models"
)

func TestParseRate(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    float64
		wantErr error
	}{
		{
			name:    "Стандартный формат с запятой",
			input:   "80,7220",
			want:    80.7220,
			wantErr: nil,
		},
		{
			name:    "Формат с точкой",
			input:   "80.7220",
			want:    80.7220,
			wantErr: nil,
		},
		{
			name:    "Целое число",
			input:   "100",
			want:    100.0,
			wantErr: nil,
		},
		{
			name:    "С пробелами",
			input:   "  94,5120  ",
			want:    94.5120,
			wantErr: nil,
		},
		{
			name:    "Пустая строка",
			input:   "",
			want:    0,
			wantErr: ErrInvalidRate,
		},
		{
			name:    "Некорректный формат",
			input:   "abc",
			want:    0,
			wantErr: ErrInvalidRate,
		},
		{
			name:    "Отрицательное значение",
			input:   "-50,5",
			want:    0,
			wantErr: ErrInvalidRate,
		},
		{
			name:    "Ноль",
			input:   "0",
			want:    0,
			wantErr: ErrInvalidRate,
		},
		{
			name:    "Infinity",
			input:   "Inf",
			want:    0,
			wantErr: ErrInvalidRate,
		},
		{
			name:    "Negative Infinity",
			input:   "-Inf",
			want:    0,
			wantErr: ErrInvalidRate,
		},
		{
			name:    "NaN",
			input:   "NaN",
			want:    0,
			wantErr: ErrInvalidRate,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseRate(tt.input)

			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("Ожидалась ошибка %v, ошибка не получена", tt.wantErr)
				}
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("Ожидалась ошибка %v, получена %v", tt.wantErr, err)
				}
			} else {
				if err != nil {
					t.Fatalf("Неожиданная ошибка: %v", err)
				}
				if got != tt.want {
					t.Errorf("Результат: ожидалось %v, получено %v", tt.want, got)
				}
			}
		})
	}
}

func TestParseNominal(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    int
		wantErr error
	}{
		{
			name:    "Один",
			input:   "1",
			want:    1,
			wantErr: nil,
		},
		{
			name:    "Сто",
			input:   "100",
			want:    100,
			wantErr: nil,
		},
		{
			name:    "Тысяча",
			input:   "1000",
			want:    1000,
			wantErr: nil,
		},
		{
			name:    "С пробелами",
			input:   "  10  ",
			want:    10,
			wantErr: nil,
		},
		{
			name:    "С пробелами в числе (10 000)",
			input:   "10 000",
			want:    10000,
			wantErr: nil,
		},
		{
			name:    "Пустая строка",
			input:   "",
			want:    0,
			wantErr: ErrInvalidNominal,
		},
		{
			name:    "Некорректный формат",
			input:   "abc",
			want:    0,
			wantErr: ErrInvalidNominal,
		},
		{
			name:    "Отрицательное значение",
			input:   "-100",
			want:    0,
			wantErr: ErrInvalidNominal,
		},
		{
			name:    "Ноль",
			input:   "0",
			want:    0,
			wantErr: ErrInvalidNominal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseNominal(tt.input)

			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("Ожидалась ошибка %v, ошибка не получена", tt.wantErr)
				}
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("Ожидалась ошибка %v, получена %v", tt.wantErr, err)
				}
			} else {
				if err != nil {
					t.Fatalf("Неожиданная ошибка: %v", err)
				}
				if got != tt.want {
					t.Errorf("Результат: ожидалось %v, получено %v", tt.want, got)
				}
			}
		})
	}
}

func TestParseCurrency(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    models.Currency
		wantErr error
	}{
		{
			name:    "USD",
			input:   "USD",
			want:    models.USD,
			wantErr: nil,
		},
		{
			name:    "EUR",
			input:   "EUR",
			want:    models.EUR,
			wantErr: nil,
		},
		{
			name:    "RUB",
			input:   "RUB",
			want:    models.RUB,
			wantErr: nil,
		},
		{
			name:    "Lowercase USD",
			input:   "usd",
			want:    models.USD,
			wantErr: nil,
		},
		{
			name:    "С пробелами",
			input:   "  EUR  ",
			want:    models.EUR,
			wantErr: nil,
		},
		{
			name:    "Неподдерживаемая валюта",
			input:   "XDR",
			want:    "",
			wantErr: ErrUnsupportedCurrency,
		},
		{
			name:    "Пустая строка",
			input:   "",
			want:    "",
			wantErr: ErrUnsupportedCurrency,
		},
		{
			name:    "Некорректный код",
			input:   "INVALID",
			want:    "",
			wantErr: ErrUnsupportedCurrency,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseCurrency(tt.input)

			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("Ожидалась ошибка %v, ошибка не получена", tt.wantErr)
				}
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("Ожидалась ошибка %v, получена %v", tt.wantErr, err)
				}
			} else {
				if err != nil {
					t.Fatalf("Неожиданная ошибка: %v", err)
				}
				if got != tt.want {
					t.Errorf("Результат: ожидалось %v, получено %v", tt.want, got)
				}
			}
		})
	}
}

// Бенчмарк для ParseXML
func BenchmarkParseXML(b *testing.B) {
	testDate := testPastDateUTC()
	dateStr := formatCBRDate(testDate)
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
        <Value>94,5120</Value>
    </Valute>
</ValCurs>`, dateStr)
	reader := strings.NewReader(xmlData)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reader.Reset(xmlData)
		_, _ = ParseXML(reader, testDate)
	}
}
