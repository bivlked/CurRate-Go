package parser

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestFetchXML(t *testing.T) {
	t.Run("Успешный HTTP запрос", func(t *testing.T) {
		// Создаем мок сервер
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Проверяем User-Agent
			userAgent := r.Header.Get("User-Agent")
			if !strings.Contains(userAgent, "CurRate-Go") {
				t.Errorf("Ожидался User-Agent 'CurRate-Go', получен: %s", userAgent)
			}

			w.WriteHeader(http.StatusOK)
			w.Write([]byte("<html><body>Test</body></html>"))
		}))
		defer server.Close()

		body, err := fetchXML(server.URL)
		if err != nil {
			t.Fatalf("Неожиданная ошибка: %v", err)
		}
		defer body.Close()

		content, err := io.ReadAll(body)
		if err != nil {
			t.Fatalf("Ошибка чтения body: %v", err)
		}

		expected := "<html><body>Test</body></html>"
		if string(content) != expected {
			t.Errorf("Контент: ожидалось %s, получено %s", expected, string(content))
		}
	})

	t.Run("Сервер возвращает 404", func(t *testing.T) {
		originalSleep := sleepFunc
		sleepFunc = func(time.Duration) {}
		t.Cleanup(func() {
			sleepFunc = originalSleep
		})

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer server.Close()

		body, err := fetchXML(server.URL)
		if err == nil {
			body.Close()
			t.Fatal("Ожидалась ошибка для статуса 404")
		}

		// После retry получаем ErrMaxRetries (внешняя ошибка)
		if !errors.Is(err, ErrMaxRetries) {
			t.Errorf("Ожидалась ошибка ErrMaxRetries, получена: %v", err)
		}

		// Проверяем, что в сообщении есть упоминание статуса
		errMsg := err.Error()
		if !strings.Contains(errMsg, "404") {
			t.Errorf("Ожидалось упоминание статуса 404 в ошибке: %s", errMsg)
		}
	})

	t.Run("Сервер недоступен (retry logic)", func(t *testing.T) {
		originalSleep := sleepFunc
		sleepFunc = func(time.Duration) {}
		t.Cleanup(func() {
			sleepFunc = originalSleep
		})

		// Используем невалидный URL
		body, err := fetchXML("http://localhost:99999")
		if err == nil {
			body.Close()
			t.Fatal("Ожидалась ошибка для недоступного сервера")
		}

		if !errors.Is(err, ErrMaxRetries) {
			t.Errorf("Ожидалась ошибка ErrMaxRetries, получена: %v", err)
		}
	})

	t.Run("Сервер восстанавливается после нескольких попыток", func(t *testing.T) {
		originalSleep := sleepFunc
		sleepFunc = func(time.Duration) {}
		t.Cleanup(func() {
			sleepFunc = originalSleep
		})

		attemptCount := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			attemptCount++
			if attemptCount < 3 {
				// Первые 2 попытки - ошибка
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			// 3я попытка - успех
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Success after retry"))
		}))
		defer server.Close()

		body, err := fetchXML(server.URL)
		if err != nil {
			t.Fatalf("Неожиданная ошибка: %v", err)
		}
		defer body.Close()

		if attemptCount != 3 {
			t.Errorf("Ожидалось 3 попытки, выполнено: %d", attemptCount)
		}
	})
}

func TestDoRequest(t *testing.T) {
	t.Run("Успешный запрос", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
		}))
		defer server.Close()

		client := newHTTPClient()
		resp, err := doRequest(client, server.URL)
		if err != nil {
			t.Fatalf("Неожиданная ошибка: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Статус код: ожидалось %d, получено %d", http.StatusOK, resp.StatusCode)
		}
	})

	t.Run("Невалидный URL", func(t *testing.T) {
		client := newHTTPClient()
		resp, err := doRequest(client, "://invalid-url")
		if err == nil {
			resp.Body.Close()
			t.Fatal("Ожидалась ошибка для невалидного URL")
		}
	})

	t.Run("User-Agent установлен", func(t *testing.T) {
		var receivedUserAgent string
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			receivedUserAgent = r.Header.Get("User-Agent")
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		client := newHTTPClient()
		resp, err := doRequest(client, server.URL)
		if err != nil {
			t.Fatalf("Неожиданная ошибка: %v", err)
		}
		defer resp.Body.Close()

		if receivedUserAgent != UserAgent {
			t.Errorf("User-Agent: ожидалось %s, получено %s", UserAgent, receivedUserAgent)
		}
	})
}

func TestNewHTTPClient(t *testing.T) {
	client := newHTTPClient()

	if client == nil {
		t.Fatal("Клиент не должен быть nil")
	}

	if client.Timeout != DefaultTimeout {
		t.Errorf("Timeout: ожидалось %v, получено %v", DefaultTimeout, client.Timeout)
	}

	if client.CheckRedirect == nil {
		t.Error("CheckRedirect не должен быть nil")
	}
}

func TestBuildURL(t *testing.T) {
	tests := []struct {
		name     string
		date     time.Time
		expected string
	}{
		{
			name:     "Дата 20 декабря 2025",
			date:     time.Date(2025, 12, 20, 0, 0, 0, 0, time.UTC),
			expected: "https://www.cbr.ru/scripts/XML_daily.asp?date_req=20/12/2025",
		},
		{
			name:     "Дата 1 января 2024",
			date:     time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			expected: "https://www.cbr.ru/scripts/XML_daily.asp?date_req=01/01/2024",
		},
		{
			name:     "Дата 15 мая 2023",
			date:     time.Date(2023, 5, 15, 10, 30, 0, 0, time.UTC),
			expected: "https://www.cbr.ru/scripts/XML_daily.asp?date_req=15/05/2023",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildURL(tt.date)
			if got != tt.expected {
				t.Errorf("URL: ожидалось %s, получено %s", tt.expected, got)
			}
		})
	}
}

// Тест redirect логики
func TestHTTPClientRedirects(t *testing.T) {
	t.Run("Остановка на первом редиректе", func(t *testing.T) {
		redirectCount := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			redirectCount++
			http.Redirect(w, r, "/redirect", http.StatusFound)
		}))
		defer server.Close()

		client := newHTTPClient()
		resp, err := doRequest(client, server.URL)
		if err == nil {
			resp.Body.Close()
			t.Fatal("Ожидалась ошибка для редиректа")
		}

		if redirectCount != 1 {
			t.Errorf("Ожидался один запрос без следования редиректу, получено %d", redirectCount)
		}
	})

	t.Run("Too many redirects (10+)", func(t *testing.T) {
		client := newHTTPClient()
		
		// Создаем запрос, который будет иметь 10+ редиректов
		// Для этого используем мок, который симулирует 10 редиректов
		req, _ := http.NewRequest("GET", "http://example.com", nil)
		via := make([]*http.Request, 10)
		for i := 0; i < 10; i++ {
			via[i], _ = http.NewRequest("GET", "http://example.com", nil)
		}
		
		err := client.CheckRedirect(req, via)
		if err == nil {
			t.Fatal("Ожидалась ошибка 'too many redirects' для 10+ редиректов")
		}
		
		if err.Error() != "too many redirects" {
			t.Errorf("Ожидалась ошибка 'too many redirects', получена: %v", err)
		}
	})

	t.Run("Less than 10 redirects (returns ErrUseLastResponse)", func(t *testing.T) {
		client := newHTTPClient()
		
		// Создаем запрос с менее чем 10 редиректами
		req, _ := http.NewRequest("GET", "http://example.com", nil)
		via := make([]*http.Request, 5)
		for i := 0; i < 5; i++ {
			via[i], _ = http.NewRequest("GET", "http://example.com", nil)
		}
		
		err := client.CheckRedirect(req, via)
		if err != http.ErrUseLastResponse {
			t.Errorf("Ожидалась ошибка http.ErrUseLastResponse для <10 редиректов, получена: %v", err)
		}
	})
}

// Бенчмарк для doRequest
func BenchmarkDoRequest(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))
	defer server.Close()

	client := newHTTPClient()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resp, err := doRequest(client, server.URL)
		if err != nil {
			b.Fatal(err)
		}
		io.ReadAll(resp.Body)
		resp.Body.Close()
	}
}
