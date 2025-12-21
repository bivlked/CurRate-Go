package utils

import (
	"math"
	"testing"
)

func TestParseAmount(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    float64
		wantErr bool
	}{
		// Простые форматы
		{
			name:    "Целое число",
			input:   "1000",
			want:    1000.0,
			wantErr: false,
		},
		{
			name:    "Число с точкой",
			input:   "1000.50",
			want:    1000.50,
			wantErr: false,
		},
		{
			name:    "Число с запятой",
			input:   "1000,50",
			want:    1000.50,
			wantErr: false,
		},
		// С пробелами
		{
			name:    "Число с пробелами",
			input:   "1 000",
			want:    1000.0,
			wantErr: false,
		},
		{
			name:    "Число с пробелами и точкой",
			input:   "1 000.50",
			want:    1000.50,
			wantErr: false,
		},
		{
			name:    "Число с пробелами и запятой",
			input:   "1 000,50",
			want:    1000.50,
			wantErr: false,
		},
		// Американский формат (запятая - тысячи, точка - дробная часть)
		{
			name:    "Американский формат целое",
			input:   "1,000",
			want:    1000.0,
			wantErr: false,
		},
		{
			name:    "Американский формат дробное",
			input:   "1,000.50",
			want:    1000.50,
			wantErr: false,
		},
		{
			name:    "Американский формат большое число",
			input:   "1,234,567.89",
			want:    1234567.89,
			wantErr: false,
		},
		// Европейский формат (точка - тысячи, запятая - дробная часть)
		{
			name:    "Европейский формат целое",
			input:   "1.000",
			want:    1000.0,
			wantErr: false,
		},
		{
			name:    "Европейский формат дробное",
			input:   "1.000,50",
			want:    1000.50,
			wantErr: false,
		},
		{
			name:    "Европейский формат большое число",
			input:   "1.234.567,89",
			want:    1234567.89,
			wantErr: false,
		},
		// Малые числа
		{
			name:    "Малое число с точкой",
			input:   "0.01",
			want:    0.01,
			wantErr: false,
		},
		{
			name:    "Малое число с запятой",
			input:   "0,01",
			want:    0.01,
			wantErr: false,
		},
		// Граничные случаи
		{
			name:    "Ноль",
			input:   "0",
			want:    0.0,
			wantErr: false,
		},
		{
			name:    "Ноль с точкой",
			input:   "0.00",
			want:    0.0,
			wantErr: false,
		},
		// Ошибочные случаи
		{
			name:    "Пустая строка",
			input:   "",
			want:    0,
			wantErr: true,
		},
		{
			name:    "Только пробелы",
			input:   "   ",
			want:    0,
			wantErr: true,
		},
		{
			name:    "Буквы",
			input:   "abc",
			want:    0,
			wantErr: true,
		},
		{
			name:    "Отрицательное число",
			input:   "-100",
			want:    0,
			wantErr: true,
		},
		{
			name:    "Спецсимволы",
			input:   "$1000",
			want:    0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseAmount(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseAmount() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && math.Abs(got-tt.want) > 0.0001 {
				t.Errorf("ParseAmount() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFormatAmount(t *testing.T) {
	tests := []struct {
		name     string
		amount   float64
		decimals int
		want     string
	}{
		{
			name:     "Целое число без десятичных",
			amount:   1000.0,
			decimals: 0,
			want:     "1 000",
		},
		{
			name:     "Число с двумя десятичными",
			amount:   1000.50,
			decimals: 2,
			want:     "1 000.50",
		},
		{
			name:     "Большое число",
			amount:   1234567.89,
			decimals: 2,
			want:     "1 234 567.89",
		},
		{
			name:     "Малое число",
			amount:   99.99,
			decimals: 2,
			want:     "99.99",
		},
		{
			name:     "Ноль",
			amount:   0.0,
			decimals: 2,
			want:     "0.00",
		},
		{
			name:     "Число без дробной части",
			amount:   1000.0,
			decimals: 2,
			want:     "1 000.00",
		},
		{
			name:     "Округление",
			amount:   1000.556,
			decimals: 2,
			want:     "1 000.56",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FormatAmount(tt.amount, tt.decimals); got != tt.want {
				t.Errorf("FormatAmount() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Бенчмарки
func BenchmarkParseAmountSimple(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = ParseAmount("1000.50")
	}
}

func BenchmarkParseAmountComplex(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = ParseAmount("1,234,567.89")
	}
}

func BenchmarkFormatAmount(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = FormatAmount(1234567.89, 2)
	}
}
