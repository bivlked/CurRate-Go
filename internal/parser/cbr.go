// Package parser предоставляет функциональность для получения и парсинга курсов валют с сайта ЦБ РФ
package parser

import (
	"context"
	"fmt"
	"time"

	"github.com/bivlked/currate-go/internal/models"
)

// FetchRates получает курсы валют с сайта ЦБ РФ на указанную дату
// ctx - контекст для отмены запроса
// date - дата, на которую нужно получить курсы валют
// Возвращает *models.RateData с курсами валют или ошибку
//
// Пример использования:
//
//	date := time.Date(2025, 12, 20, 0, 0, 0, 0, time.UTC)
//	rates, err := parser.FetchRates(ctx, date)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("USD курс: %.4f\n", rates.Rates[models.USD].Rate)
func FetchRates(ctx context.Context, date time.Time) (*models.RateData, error) {
	// Строим URL для запроса с датой
	url := buildURL(date)
	return fetchRatesFromURL(ctx, url, date)
}

// fetchRatesFromURL - внутренняя функция для получения курсов с произвольного URL
// Используется для тестирования и внутри FetchRates
func fetchRatesFromURL(ctx context.Context, url string, date time.Time) (*models.RateData, error) {
	// Выполняем HTTP запрос с retry логикой и exponential backoff
	body, err := fetchXML(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch rates from CBR: %w", err)
	}
	defer body.Close()

	// Парсим XML в структуру данных
	data, err := ParseXML(body, date)
	if err != nil {
		return nil, fmt.Errorf("failed to parse CBR XML rates: %w", err)
	}

	return data, nil
}

// FetchLatestRates получает последние актуальные курсы валют с сайта ЦБ РФ
// ctx - контекст для отмены запроса
// Использует текущую дату (время.Now()) для запроса
// Возвращает *models.RateData с курсами валют или ошибку
//
// Пример использования:
//
//	rates, err := parser.FetchLatestRates(ctx)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("EUR курс: %.4f\n", rates.Rates[models.EUR].Rate)
func FetchLatestRates(ctx context.Context) (*models.RateData, error) {
	// Используем текущую дату
	now := time.Now()
	return FetchRates(ctx, now)
}
