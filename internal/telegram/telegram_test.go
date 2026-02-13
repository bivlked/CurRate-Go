package telegram

import (
	"net/http"
	"net/http/httptest"
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

// mockTransport реализует http.RoundTripper для тестирования HTTP-клиента
type mockTransport struct {
	handler func(req *http.Request) (*http.Response, error)
}

func (m *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return m.handler(req)
}
