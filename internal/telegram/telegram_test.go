package telegram

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestIsConfigured_Empty(t *testing.T) {
	// Сохраняем оригинальные значения
	origToken := botToken
	origChat := chatID
	defer func() {
		botToken = origToken
		chatID = origChat
	}()

	botToken = ""
	chatID = ""

	if IsConfigured() {
		t.Error("expected IsConfigured() = false when both are empty")
	}
}

func TestIsConfigured_OnlyToken(t *testing.T) {
	origToken := botToken
	origChat := chatID
	defer func() {
		botToken = origToken
		chatID = origChat
	}()

	botToken = "some-token"
	chatID = ""

	if IsConfigured() {
		t.Error("expected IsConfigured() = false when chatID is empty")
	}
}

func TestIsConfigured_OnlyChatID(t *testing.T) {
	origToken := botToken
	origChat := chatID
	defer func() {
		botToken = origToken
		chatID = origChat
	}()

	botToken = ""
	chatID = "12345"

	if IsConfigured() {
		t.Error("expected IsConfigured() = false when botToken is empty")
	}
}

func TestIsConfigured_BothSet(t *testing.T) {
	origToken := botToken
	origChat := chatID
	defer func() {
		botToken = origToken
		chatID = origChat
	}()

	botToken = "some-token"
	chatID = "12345"

	if !IsConfigured() {
		t.Error("expected IsConfigured() = true when both are set")
	}
}

func TestSendStar_NotConfigured(t *testing.T) {
	origToken := botToken
	origChat := chatID
	defer func() {
		botToken = origToken
		chatID = origChat
	}()

	botToken = ""
	chatID = ""

	client := NewClient()
	err := client.SendStar("user-123", "1.0.0")
	if err == nil {
		t.Fatal("expected error when not configured")
	}
}

func TestSendStar_Success(t *testing.T) {
	origToken := botToken
	origChat := chatID
	defer func() {
		botToken = origToken
		chatID = origChat
	}()

	botToken = "test-token"
	chatID = "12345"

	// Перехватываем HTTP-запросы через mockTransport,
	// т.к. telegramAPI — константа и не может быть подменена
	transport := &mockTransport{
		handler: func(req *http.Request) (*http.Response, error) {
			recorder := httptest.NewRecorder()
			recorder.WriteHeader(http.StatusOK)
			recorder.WriteString(`{"ok":true}`)
			return recorder.Result(), nil
		},
	}

	client := &Client{
		httpClient: &http.Client{Transport: transport},
	}

	err := client.SendStar("user-123", "1.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSendStar_HTTPError(t *testing.T) {
	origToken := botToken
	origChat := chatID
	defer func() {
		botToken = origToken
		chatID = origChat
	}()

	botToken = "test-token"
	chatID = "12345"

	transport := &mockTransport{
		handler: func(req *http.Request) (*http.Response, error) {
			recorder := httptest.NewRecorder()
			recorder.WriteHeader(http.StatusForbidden)
			recorder.WriteString(`{"ok":false,"description":"Forbidden: bot was blocked"}`)
			return recorder.Result(), nil
		},
	}

	client := &Client{
		httpClient: &http.Client{Transport: transport},
	}

	err := client.SendStar("user-123", "1.0.0")
	if err == nil {
		t.Fatal("expected error on HTTP 403")
	}
	if !strings.Contains(err.Error(), "403") {
		t.Errorf("error should contain status code 403, got: %v", err)
	}
	if !strings.Contains(err.Error(), "Forbidden") {
		t.Errorf("error should contain response body text, got: %v", err)
	}
}

func TestSendStar_TransportErrorDoesNotLeakToken(t *testing.T) {
	origToken := botToken
	origChat := chatID
	defer func() {
		botToken = origToken
		chatID = origChat
	}()

	botToken = "SECRET-BOT-TOKEN-12345"
	chatID = "12345"

	// Транспорт возвращает ошибку - http.Client оборачивает её в *url.Error,
	// который содержит полный URL запроса (включая botToken)
	transport := &mockTransport{
		handler: func(req *http.Request) (*http.Response, error) {
			return nil, errors.New("dial tcp: connection refused")
		},
	}

	client := &Client{
		httpClient: &http.Client{Transport: transport},
	}

	err := client.SendStar("user-123", "1.0.0")
	if err == nil {
		t.Fatal("expected error on transport failure")
	}
	if strings.Contains(err.Error(), botToken) {
		t.Errorf("error must not contain botToken, got: %v", err)
	}
	if !strings.Contains(err.Error(), "connection refused") {
		t.Errorf("error should preserve underlying cause, got: %v", err)
	}
}

func TestSanitizeSendError(t *testing.T) {
	origToken := botToken
	defer func() { botToken = origToken }()
	botToken = "SECRET-TOKEN"

	tests := []struct {
		name string
		err  error
	}{
		{
			name: "url.Error с токеном в URL",
			err: &url.Error{
				Op:  "Post",
				URL: telegramAPI + botToken + "/sendMessage",
				Err: errors.New("dial tcp: timeout"),
			},
		},
		{
			name: "вложенный url.Error с токеном",
			err: &url.Error{
				Op:  "Post",
				URL: "https://example.com",
				Err: &url.Error{
					Op:  "Get",
					URL: telegramAPI + botToken + "/sendMessage",
					Err: errors.New("stopped after 10 redirects"),
				},
			},
		},
		{
			name: "токен в тексте не-url ошибки",
			err:  errors.New("request to " + telegramAPI + botToken + " failed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sanitizeSendError(tt.err)
			if got == nil {
				t.Fatal("sanitizeSendError() = nil, want error")
			}
			if strings.Contains(got.Error(), botToken) {
				t.Errorf("sanitized error still contains token: %v", got)
			}
		})
	}
}

// mockTransport реализует http.RoundTripper для тестирования HTTP-клиента
type mockTransport struct {
	handler func(req *http.Request) (*http.Response, error)
}

func (m *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return m.handler(req)
}
