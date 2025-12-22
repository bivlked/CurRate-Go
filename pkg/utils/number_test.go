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
		{
			name:    "Плюс перед числом",
			input:   "+2500",
			want:    2500.0,
			wantErr: false,
		},
		{
			name:    "Лидирующие нули",
			input:   "000123",
			want:    123.0,
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
		{
			name:    "Табуляция в числе",
			input:   "1\t000",
			want:    0,
			wantErr: true,
		},
		// Edge cases для улучшения покрытия
		{
			name:    "Несколько запятых (разделители тысяч)",
			input:   "1,234,567",
			want:    1234567.0,
			wantErr: false,
		},
		{
			name:    "Несколько точек (разделители тысяч)",
			input:   "1.234.567",
			want:    1234567.0,
			wantErr: false,
		},
		{
			name:    "Запятая с 3 цифрами после (разделитель тысяч)",
			input:   "1,234",
			want:    1234.0,
			wantErr: false,
		},
		{
			name:    "Точка с 3 цифрами после (разделитель тысяч)",
			input:   "1.234",
			want:    1234.0,
			wantErr: false,
		},
		{
			name:    "Запятая с 2 цифрами после (дробная часть)",
			input:   "1,23",
			want:    1.23,
			wantErr: false,
		},
		{
			name:    "Точка с 2 цифрами после (дробная часть)",
			input:   "1.23",
			want:    1.23,
			wantErr: false,
		},
		{
			name:    "Запятая с 4 цифрами после (нестандартный формат)",
			input:   "1,2345",
			want:    1.2345,
			wantErr: false,
		},
		{
			name:    "Точка с 4 цифрами после (нестандартный формат)",
			input:   "1.2345",
			want:    1.2345,
			wantErr: false,
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
		{
			name:     "Без дробной части (decimals=0)",
			amount:   1000.0,
			decimals: 0,
			want:     "1 000",
		},
		{
			name:     "С дробной частью (decimals=0, но есть дробь - округление)",
			amount:   1000.5,
			decimals: 0,
			want:     "1 000", // fmt.Sprintf("%.0f", 1000.5) использует "round to even", результат "1000"
		},
		{
			name:     "С дробной частью меньше 0.5 (decimals=0)",
			amount:   1000.4,
			decimals: 0,
			want:     "1 000",
		},
		{
			name:     "Малое число без разделителей тысяч",
			amount:   99.99,
			decimals: 2,
			want:     "99.99",
		},
		{
			name:     "Очень большое число",
			amount:   1234567890.12,
			decimals: 2,
			want:     "1 234 567 890.12",
		},
		{
			name:     "Одна цифра",
			amount:   5.0,
			decimals: 0,
			want:     "5",
		},
		{
			name:     "Две цифры",
			amount:   50.0,
			decimals: 0,
			want:     "50",
		},
		{
			name:     "Три цифры",
			amount:   500.0,
			decimals: 0,
			want:     "500",
		},
		{
			name:     "Четыре цифры",
			amount:   5000.0,
			decimals: 0,
			want:     "5 000",
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
