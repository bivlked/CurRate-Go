package telegram

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// UserData хранит данные пользователя
type UserData struct {
	UserID string `json:"user_id"`
}

// GetOrCreateUserID получает или создает уникальный ID пользователя
// ID сохраняется в %APPDATA%/CurRate/user.json
func GetOrCreateUserID() (string, error) {
	// Получаем путь к APPDATA
	appData := os.Getenv("APPDATA")
	if appData == "" {
		// Fallback для не-Windows систем
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("не удалось определить домашнюю директорию: %w", err)
		}
		appData = filepath.Join(homeDir, ".config")
	}

	// Создаем директорию приложения
	appDir := filepath.Join(appData, "CurRate")
	if err := os.MkdirAll(appDir, 0755); err != nil {
		return "", fmt.Errorf("не удалось создать директорию: %w", err)
	}

	// Путь к файлу с данными пользователя
	userFile := filepath.Join(appDir, "user.json")

	// Пробуем прочитать существующий ID
	if data, err := os.ReadFile(userFile); err == nil {
		var userData UserData
		if err := json.Unmarshal(data, &userData); err == nil && userData.UserID != "" {
			return userData.UserID, nil
		}
	}

	// Генерируем новый UUID
	userID, err := generateUUID()
	if err != nil {
		return "", fmt.Errorf("не удалось сгенерировать UUID: %w", err)
	}

	// Сохраняем в файл
	userData := UserData{UserID: userID}
	data, err := json.MarshalIndent(userData, "", "  ")
	if err != nil {
		return "", fmt.Errorf("не удалось сериализовать данные: %w", err)
	}

	if err := os.WriteFile(userFile, data, 0644); err != nil {
		return "", fmt.Errorf("не удалось сохранить данные: %w", err)
	}

	return userID, nil
}

// generateUUID генерирует UUID v4
func generateUUID() (string, error) {
	uuid := make([]byte, 16)
	_, err := rand.Read(uuid)
	if err != nil {
		return "", err
	}

	// Устанавливаем версию (4) и вариант (RFC 4122)
	uuid[6] = (uuid[6] & 0x0f) | 0x40 // версия 4
	uuid[8] = (uuid[8] & 0x3f) | 0x80 // вариант RFC 4122

	// Форматируем как стандартный UUID
	return fmt.Sprintf("%s-%s-%s-%s-%s",
		hex.EncodeToString(uuid[0:4]),
		hex.EncodeToString(uuid[4:6]),
		hex.EncodeToString(uuid[6:8]),
		hex.EncodeToString(uuid[8:10]),
		hex.EncodeToString(uuid[10:16])), nil
}
