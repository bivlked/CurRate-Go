package telegram

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Переменные для внедрения при сборке через ldflags
// Пример: go build -ldflags "-X github.com/bivlked/CurRate-Go/internal/telegram.botToken=YOUR_TOKEN"
var (
	// botToken - токен Telegram бота (внедряется при сборке)
	botToken = ""
	// chatID - ID чата для уведомлений (внедряется при сборке)
	chatID = ""
)

const (
	// API URL
	telegramAPI = "https://api.telegram.org/bot"
)

// Client представляет клиент для работы с Telegram API
type Client struct {
	httpClient *http.Client
}

// NewClient создает новый клиент Telegram
func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// sendMessageRequest - структура запроса для отправки сообщения
type sendMessageRequest struct {
	ChatID    string `json:"chat_id"`
	Text      string `json:"text"`
	ParseMode string `json:"parse_mode,omitempty"`
}

// IsConfigured проверяет, настроена ли интеграция с Telegram
func IsConfigured() bool {
	return botToken != "" && chatID != ""
}

// SendStar отправляет уведомление о "звезде" в Telegram
func (c *Client) SendStar(userID string, appVersion string) error {
	// Проверяем, настроена ли интеграция
	if !IsConfigured() {
		return fmt.Errorf("Telegram интеграция не настроена")
	}

	// Формируем текст сообщения
	timestamp := time.Now().Format("02.01.2006 15:04:05")
	text := fmt.Sprintf("⭐ *Новая звезда!*\n\n"+
		"📱 Приложение: Конвертер валют\n"+
		"📦 Версия: %s\n"+
		"👤 ID пользователя: `%s`\n"+
		"🕐 Время: %s",
		appVersion, userID, timestamp)

	// Создаем запрос
	reqBody := sendMessageRequest{
		ChatID:    chatID,
		Text:      text,
		ParseMode: "Markdown",
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("ошибка сериализации: %w", err)
	}

	// Отправляем запрос
	url := telegramAPI + botToken + "/sendMessage"
	resp, err := c.httpClient.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("ошибка отправки: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Telegram API вернул статус: %d", resp.StatusCode)
	}

	return nil
}
