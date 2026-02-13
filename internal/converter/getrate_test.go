package converter

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/bivlked/currate-go/internal/models"
)

// TestConverter_GetRate тестирует метод GetRate для получения курса без форматирования
func TestConverter_GetRate(t *testing.T) {
	date := time.Date(2025, 12, 20, 0, 0, 0, 0, time.UTC)

	mockProvider := &MockRateProvider{
		rateData: &models.RateData{
			Date: date,
			Rates: map[models.Currency]models.ExchangeRate{
				models.USD: {
					Currency: models.USD,
					Rate:     80.7220,
					Nominal:  1,
					Date:     date,
				},
				models.EUR: {
					Currency: models.EUR,
					Rate:     88.1234,
					Nominal:  1,
					Date:     date,
				},
			},
		},
	}

	cache := NewMockCache()
	converter := NewConverter(mockProvider, cache)

	tests := []struct {
		name     string
		currency models.Currency
		date     time.Time
		want     float64
		wantErr  bool
	}{
		{
			name:     "USD курс",
			currency: models.USD,
			date:     date,
			want:     80.7220,
			wantErr:  false,
		},
		{
			name:     "EUR курс",
			currency: models.EUR,
			date:     date,
			want:     88.1234,
			wantErr:  false,
		},
		{
			name:     "RUB всегда возвращает 1.0",
			currency: models.RUB,
			date:     date,
			want:     1.0,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := converter.GetRate(context.Background(), tt.currency, tt.date)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetRate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("GetRate() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestConverter_GetRate_ValidationErrors тестирует валидацию в GetRate
func TestConverter_GetRate_ValidationErrors(t *testing.T) {
	date := time.Date(2025, 12, 20, 0, 0, 0, 0, time.UTC)

	mockProvider := &MockRateProvider{
		rateData: &models.RateData{
			Date: date,
			Rates: map[models.Currency]models.ExchangeRate{
				models.USD: {
					Currency: models.USD,
					Rate:     80.7220,
					Nominal:  1,
					Date:     date,
				},
			},
		},
	}
	cache := NewMockCache()
	converter := NewConverter(mockProvider, cache)

	futureDate := time.Now().AddDate(0, 0, 1)

	tests := []struct {
		name     string
		currency models.Currency
		date     time.Time
		wantErr  error
	}{
		{
			name:     "Неподдерживаемая валюта",
			currency: models.Currency("GBP"),
			date:     date,
			wantErr:  models.ErrUnsupportedCurrency,
		},
		{
			name:     "Дата в будущем",
			currency: models.USD,
			date:     futureDate,
			wantErr:  ErrDateInFuture,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := converter.GetRate(context.Background(), tt.currency, tt.date)
			if err == nil {
				t.Fatal("Ожидалась ошибка, но её нет")
			}

			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Ожидалась ошибка %v, получена %v", tt.wantErr, err)
			}
		})
	}
}

// TestConverter_GetRate_UsesCache тестирует использование кэша в GetRate
func TestConverter_GetRate_UsesCache(t *testing.T) {
	date := time.Date(2025, 12, 20, 0, 0, 0, 0, time.UTC)

	mockProvider := &MockRateProvider{
		rateData: &models.RateData{
			Date: date,
			Rates: map[models.Currency]models.ExchangeRate{
				models.USD: {
					Currency: models.USD,
					Rate:     80.7220,
					Nominal:  1,
					Date:     date,
				},
			},
		},
	}
	cache := NewMockCache()
	converter := NewConverter(mockProvider, cache)

	// Первый вызов - должен запросить у provider
	rate1, err := converter.GetRate(context.Background(), models.USD, date)
	if err != nil {
		t.Fatalf("GetRate() error = %v", err)
	}

	// Проверяем, что provider был вызван
	if mockProvider.callCount == 0 {
		t.Error("Provider не был вызван при первом запросе")
	}

	// Второй вызов - должен использовать кэш
	initialCallCount := mockProvider.callCount
	rate2, err := converter.GetRate(context.Background(), models.USD, date)
	if err != nil {
		t.Fatalf("GetRate() error = %v", err)
	}

	// Проверяем, что provider не был вызван повторно
	if mockProvider.callCount > initialCallCount {
		t.Error("Provider был вызван повторно, хотя должен использоваться кэш")
	}

	// Проверяем, что курс одинаковый
	if rate1 != rate2 {
		t.Errorf("Курсы не совпадают: rate1 = %v, rate2 = %v", rate1, rate2)
	}
}
