package parser

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/bivlked/currate-go/internal/models"
)

// Тестовый HTML фрагмент, взятый с реального сайта ЦБ РФ
const testHTML = `
<!DOCTYPE html>
<html>
<body>
	<table class="data">
		<tbody>
			<tr>
				<th>Цифр. код</th>
				<th>Букв. код</th>
				<th>Единиц</th>
				<th>Валюта</th>
				<th>Курс</th>
			</tr>
			<tr>
				<td>840</td>
				<td>USD</td>
				<td>1</td>
				<td>Доллар США</td>
				<td>80,7220</td>
			</tr>
			<tr>
				<td>978</td>
				<td>EUR</td>
				<td>1</td>
				<td>Евро</td>
				<td>94,5120</td>
			</tr>
			<tr>
				<td>392</td>
				<td>JPY</td>
				<td>100</td>
				<td>Иен</td>
				<td>51,8346</td>
			</tr>
			<tr>
				<td>960</td>
				<td>XDR</td>
				<td>1</td>
				<td>СДР</td>
				<td>110,3454</td>
			</tr>
		</tbody>
	</table>
</body>
</html>
`

// Пустой HTML без таблицы
const emptyHTML = `
<!DOCTYPE html>
<html>
<body>
	<p>No data</p>
</body>
</html>
`

// HTML с некорректными данными
const invalidHTML = `
<!DOCTYPE html>
<html>
<body>
	<table class="data">
		<tbody>
			<tr>
				<th>Цифр. код</th>
				<th>Букв. код</th>
				<th>Единиц</th>
				<th>Валюта</th>
				<th>Курс</th>
			</tr>
			<tr>
				<td>840</td>
				<td>USD</td>
				<td>invalid</td>
				<td>Доллар США</td>
				<td>abc</td>
			</tr>
		</tbody>
	</table>
</body>
</html>
`

func TestParseHTML(t *testing.T) {
	testDate := time.Date(2025, 12, 20, 0, 0, 0, 0, time.UTC)

	t.Run("Успешный парсинг валидного HTML", func(t *testing.T) {
		reader := strings.NewReader(testHTML)
		data, err := ParseHTML(reader, testDate)

		if err != nil {
			t.Fatalf("Ожидался успешный парсинг, получена ошибка: %v", err)
		}

		if data == nil {
			t.Fatal("Ожидались данные, получен nil")
		}

		if !data.Date.Equal(testDate) {
			t.Errorf("Дата: ожидалось %v, получено %v", testDate, data.Date)
		}

		// Проверяем USD
		usd, ok := data.Rates[models.USD]
		if !ok {
			t.Fatal("USD не найден в результатах")
		}
		if usd.Rate != 80.7220 {
			t.Errorf("USD курс: ожидалось 80.7220, получено %v", usd.Rate)
		}
		if usd.Nominal != 1 {
			t.Errorf("USD номинал: ожидалось 1, получено %v", usd.Nominal)
		}

		// Проверяем EUR
		eur, ok := data.Rates[models.EUR]
		if !ok {
			t.Fatal("EUR не найден в результатах")
		}
		if eur.Rate != 94.5120 {
			t.Errorf("EUR курс: ожидалось 94.5120, получено %v", eur.Rate)
		}

		// Проверяем что XDR (неподдерживаемая валюта) пропущен
		_, exists := data.Rates[models.Currency("XDR")]
		if exists {
			t.Error("XDR не должен присутствовать в результатах (неподдерживаемая валюта)")
		}

		// Проверяем количество валют (USD, EUR из поддерживаемых)
		expectedCount := 2
		if len(data.Rates) != expectedCount {
			t.Errorf("Количество валют: ожидалось %d, получено %d", expectedCount, len(data.Rates))
		}
	})

	t.Run("Пустой HTML без таблицы", func(t *testing.T) {
		reader := strings.NewReader(emptyHTML)
		data, err := ParseHTML(reader, testDate)

		if err == nil {
			t.Fatal("Ожидалась ошибка для пустого HTML")
		}

		if !errors.Is(err, ErrNoRates) {
			t.Errorf("Ожидалась ошибка ErrNoRates, получена: %v", err)
		}

		if data != nil {
			t.Error("Для ошибки ожидался nil data")
		}
	})

	t.Run("Некорректный HTML", func(t *testing.T) {
		reader := strings.NewReader(invalidHTML)
		// Все строки с некорректными данными должны быть пропущены
		data, err := ParseHTML(reader, testDate)

		// Должна быть ошибка, так как все строки некорректны
		if err == nil {
			t.Fatal("Ожидалась ошибка для некорректного HTML")
		}

		if !errors.Is(err, ErrNoRates) {
			t.Errorf("Ожидалась ошибка ErrNoRates, получена: %v", err)
		}

		if data != nil {
			t.Error("Для ошибки ожидался nil data")
		}
	})

	t.Run("Невалидный HTML (не HTML)", func(t *testing.T) {
		reader := strings.NewReader("not html at all")
		data, err := ParseHTML(reader, testDate)

		// goquery обычно не падает на невалидном HTML, но может вернуть пустой документ
		// В нашем случае должна быть ошибка ErrNoRates
		if err == nil {
			t.Fatal("Ожидалась ошибка для невалидного HTML")
		}

		if data != nil {
			t.Error("Для ошибки ожидался nil data")
		}
	})
}

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

// Бенчмарк для ParseHTML
func BenchmarkParseHTML(b *testing.B) {
	reader := strings.NewReader(testHTML)
	testDate := time.Date(2025, 12, 20, 0, 0, 0, 0, time.UTC)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reader.Reset(testHTML)
		_, _ = ParseHTML(reader, testDate)
	}
}
