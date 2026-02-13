package parser

import (
	"context"
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
	UserAgent = "CurRate-Go/1.2 (Windows; Go; XML)"
)

// Ошибки HTTP клиента
var (
	ErrHTTPFailed    = errors.New("HTTP request failed")
	ErrInvalidStatus = errors.New("invalid HTTP status code")
	ErrMaxRetries    = errors.New("max retries exceeded")
)

// sleepWithContext ожидает указанную длительность с поддержкой отмены через контекст
var sleepWithContext = defaultSleepWithContext

func defaultSleepWithContext(ctx context.Context, d time.Duration) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(d):
		return nil
	}
}

// Глобальный HTTP-клиент для переиспользования соединений (HTTP keep-alive)
// Создается один раз при инициализации пакета и используется для всех запросов
// Это значительно повышает производительность за счет:
// - Переиспользования TCP соединений (keep-alive)
// - Снижения накладных расходов на создание нового transport для каждого запроса
// - Эффективного использования пула соединений
var defaultHTTPClient = newHTTPClient()

// fetchXML выполняет HTTP GET запрос с retry логикой и exponential backoff
// ctx - контекст для отмены запросов и backoff ожидания
// url - URL для запроса
// Возвращает io.ReadCloser с XML контентом (caller должен закрыть его)
func fetchXML(ctx context.Context, url string) (io.ReadCloser, error) {
	// Используем глобальный HTTP-клиент для переиспользования соединений
	// Это значительно улучшает производительность при частых запросах
	client := defaultHTTPClient

	var lastErr error
	var attempts int
	for attempt := 1; attempt <= MaxRetries; attempt++ {
		attempts = attempt
		resp, err := doRequest(ctx, client, url)
		if err == nil {
			return resp.Body, nil
		}

		lastErr = err

		// Если контекст отменён — выходим немедленно
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		// Не повторяем запрос при ошибках клиента (4xx) — повторный запрос вернет тот же результат
		// Возвращаем оригинальную ошибку без обёртки ErrMaxRetries
		if errors.Is(err, ErrInvalidStatus) {
			return nil, lastErr
		}

		// Если это не последняя попытка, ждем с exponential backoff
		// Attempt 1: 1s, Attempt 2: 2s, Attempt 3: 4s (как в Python версии)
		if attempt < MaxRetries {
			delay := BaseRetryDelay * time.Duration(1<<uint(attempt-1))
			if err := sleepWithContext(ctx, delay); err != nil {
				return nil, err
			}
		}
	}

	return nil, fmt.Errorf("%w after %d attempts: %w", ErrMaxRetries, attempts, lastErr)
}

// doRequest выполняет одиночный HTTP запрос
// ctx - контекст для отмены запроса
// client - HTTP клиент
// url - URL для запроса
func doRequest(ctx context.Context, client *http.Client, url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Устанавливаем User-Agent для идентификации
	req.Header.Set("User-Agent", UserAgent)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrHTTPFailed, err)
	}

	// Проверяем статус код
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		// Разделяем клиентские (4xx) и серверные (5xx) ошибки:
		// 4xx — не ретраим (проблема в запросе, повторный вызов не поможет)
		// 5xx и прочие — ретраим (временная проблема сервера)
		if resp.StatusCode >= 400 && resp.StatusCode < 500 {
			return nil, fmt.Errorf("%w: %d %s", ErrInvalidStatus, resp.StatusCode, resp.Status)
		}
		return nil, fmt.Errorf("%w: %d %s", ErrHTTPFailed, resp.StatusCode, resp.Status)
	}

	return resp, nil
}

// newHTTPClient создает HTTP клиент с настройками
func newHTTPClient() *http.Client {
	return &http.Client{
		Timeout: DefaultTimeout,
		// Разрешаем автоматические редиректы с лимитом на 10 редиректов
		// Это защищает от бесконечных циклов, но позволяет обрабатывать нормальные редиректы
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 10 {
				return errors.New("too many redirects")
			}
			// Возвращаем nil для продолжения автоматической обработки редиректа
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
