package parser

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

// HTTP константы
const (
	// CBRURL - базовый URL XML API ЦБ РФ для получения курсов валют
	CBRURL = "https://www.cbr.ru/scripts/XML_daily.asp"

	// DefaultTimeout - таймаут для HTTP запросов
	DefaultTimeout = 10 * time.Second

	// MaxRetries - максимальное количество повторных попыток при ошибке
	MaxRetries = 3

	// BaseRetryDelay - базовая задержка для exponential backoff (1s, 2s, 4s)
	BaseRetryDelay = 1 * time.Second

	// UserAgent - User-Agent для HTTP запросов
	UserAgent = "CurRate-Go/2.0 (Windows; Go; XML)"
)

// Ошибки HTTP клиента
var (
	ErrHTTPFailed    = errors.New("HTTP request failed")
	ErrInvalidStatus = errors.New("invalid HTTP status code")
	ErrMaxRetries    = errors.New("max retries exceeded")
)

// fetchXML выполняет HTTP GET запрос с retry логикой и exponential backoff
// url - URL для запроса
// Возвращает io.ReadCloser с XML контентом (caller должен закрыть его)
func fetchXML(url string) (io.ReadCloser, error) {
	client := newHTTPClient()

	var lastErr error
	for attempt := 1; attempt <= MaxRetries; attempt++ {
		resp, err := doRequest(client, url)
		if err == nil {
			return resp.Body, nil
		}

		lastErr = err

		// Если это не последняя попытка, ждем с exponential backoff
		// Attempt 1: 1s, Attempt 2: 2s, Attempt 3: 4s (как в Python версии)
		if attempt < MaxRetries {
			delay := BaseRetryDelay * time.Duration(1<<uint(attempt-1))
			time.Sleep(delay)
		}
	}

	return nil, fmt.Errorf("%w after %d attempts: %v", ErrMaxRetries, MaxRetries, lastErr)
}

// doRequest выполняет одиночный HTTP запрос
// client - HTTP клиент
// url - URL для запроса
func doRequest(client *http.Client, url string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Устанавливаем User-Agent для идентификации
	req.Header.Set("User-Agent", UserAgent)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrHTTPFailed, err)
	}

	// Проверяем статус код
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("%w: %d %s", ErrInvalidStatus, resp.StatusCode, resp.Status)
	}

	return resp, nil
}

// newHTTPClient создает HTTP клиент с настройками
func newHTTPClient() *http.Client {
	return &http.Client{
		Timeout: DefaultTimeout,
		// Не следуем редиректам автоматически (для безопасности)
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 10 {
				return errors.New("too many redirects")
			}
			return nil
		},
	}
}

// buildURL строит URL для запроса курсов на определенную дату из XML API
// date - дата курсов
func buildURL(date time.Time) string {
	// Формат даты для XML API: DD/MM/YYYY (например, 20/12/2025)
	// Обратите внимание: слэш вместо точки, в отличие от HTML API
	dateStr := date.Format("02/01/2006")
	return fmt.Sprintf("%s?date_req=%s", CBRURL, dateStr)
}
