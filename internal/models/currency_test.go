package models

import "testing"

func TestCurrencyValidate(t *testing.T) {
	tests := []struct {
		name    string
		curr    Currency
		wantErr bool
	}{
		{
			name:    "USD валидна",
			curr:    USD,
			wantErr: false,
		},
		{
			name:    "EUR валидна",
			curr:    EUR,
			wantErr: false,
		},
		{
			name:    "RUB валидна",
			curr:    RUB,
			wantErr: false,
		},
		{
			name:    "Неизвестная валюта",
			curr:    Currency("GBP"),
			wantErr: true,
		},
		{
			name:    "Пустая строка",
			curr:    Currency(""),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.curr.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Currency.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCurrencySymbol(t *testing.T) {
	tests := []struct {
		name string
		curr Currency
		want string
	}{
		{
			name: "Символ USD",
			curr: USD,
			want: "$",
		},
		{
			name: "Символ EUR",
			curr: EUR,
			want: "€",
		},
		{
			name: "Символ RUB",
			curr: RUB,
			want: "₽",
		},
		{
			name: "Неизвестная валюта возвращает код",
			curr: Currency("GBP"),
			want: "GBP",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.curr.Symbol(); got != tt.want {
				t.Errorf("Currency.Symbol() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCurrencyString(t *testing.T) {
	tests := []struct {
		name string
		curr Currency
		want string
	}{
		{
			name: "Строка USD",
			curr: USD,
			want: "USD",
		},
		{
			name: "Строка EUR",
			curr: EUR,
			want: "EUR",
		},
		{
			name: "Строка RUB",
			curr: RUB,
			want: "RUB",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.curr.String(); got != tt.want {
				t.Errorf("Currency.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCurrencyName(t *testing.T) {
	tests := []struct {
		name string
		curr Currency
		want string
	}{
		{
			name: "Название USD",
			curr: USD,
			want: "Доллар США",
		},
		{
			name: "Название EUR",
			curr: EUR,
			want: "Евро",
		},
		{
			name: "Название RUB",
			curr: RUB,
			want: "Российский рубль",
		},
		{
			name: "Неизвестная валюта возвращает код",
			curr: Currency("GBP"),
			want: "GBP",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.curr.Name(); got != tt.want {
				t.Errorf("Currency.Name() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseCurrency(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		want      Currency
		wantError bool
	}{
		{
			name:      "Поддерживаемая валюта USD",
			input:     "USD",
			want:      USD,
			wantError: false,
		},
		{
			name:      "Поддерживаемая валюта EUR",
			input:     "EUR",
			want:      EUR,
			wantError: false,
		},
		{
			name:      "Валюта в нижнем регистре",
			input:     "usd",
			want:      USD,
			wantError: false,
		},
		{
			name:      "Валюта с пробелами",
			input:     " eur ",
			want:      EUR,
			wantError: false,
		},
		{
			name:      "Валюта смешанного регистра",
			input:     "Usd",
			want:      USD,
			wantError: false,
		},
		{
			name:      "Неподдерживаемая валюта",
			input:     "GBP",
			want:      "",
			wantError: true,
		},
		{
			name:      "Пустая строка",
			input:     "",
			want:      "",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseCurrency(tt.input)
			if (err != nil) != tt.wantError {
				t.Fatalf("ParseCurrency() error = %v, wantError %v", err, tt.wantError)
			}
			if !tt.wantError && got != tt.want {
				t.Errorf("ParseCurrency() = %v, want %v", got, tt.want)
			}
		})
	}
}
