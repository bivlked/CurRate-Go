package telegram

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGetOrCreateUserID_NewFile(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("APPDATA", tmpDir)

	userID, err := GetOrCreateUserID()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if userID == "" {
		t.Fatal("expected non-empty userID")
	}

	// Проверяем формат UUID v4 (8-4-4-4-12)
	parts := strings.Split(userID, "-")
	if len(parts) != 5 {
		t.Errorf("expected UUID format (5 parts), got %d parts: %s", len(parts), userID)
	}
	if len(parts[0]) != 8 || len(parts[1]) != 4 || len(parts[2]) != 4 || len(parts[3]) != 4 || len(parts[4]) != 12 {
		t.Errorf("unexpected UUID part lengths: %s", userID)
	}

	// Проверяем что файл создан
	userFile := filepath.Join(tmpDir, "CurRate", "user.json")
	if _, err := os.Stat(userFile); os.IsNotExist(err) {
		t.Fatal("expected user.json to be created")
	}
}

func TestGetOrCreateUserID_ExistingFile(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("APPDATA", tmpDir)

	// Создаём директорию и файл заранее
	appDir := filepath.Join(tmpDir, "CurRate")
	if err := os.MkdirAll(appDir, 0755); err != nil {
		t.Fatal(err)
	}

	existingID := "aaaaaaaa-bbbb-4ccc-8ddd-eeeeeeeeeeee"
	data, _ := json.MarshalIndent(UserData{UserID: existingID}, "", "  ")
	if err := os.WriteFile(filepath.Join(appDir, "user.json"), data, 0644); err != nil {
		t.Fatal(err)
	}

	userID, err := GetOrCreateUserID()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if userID != existingID {
		t.Errorf("expected %s, got %s", existingID, userID)
	}
}

func TestGetOrCreateUserID_InvalidContent(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("APPDATA", tmpDir)

	// Создаём файл с невалидным содержимым
	appDir := filepath.Join(tmpDir, "CurRate")
	if err := os.MkdirAll(appDir, 0755); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(appDir, "user.json"), []byte("not-json"), 0644); err != nil {
		t.Fatal(err)
	}

	userID, err := GetOrCreateUserID()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Должен сгенерировать новый ID
	if userID == "" {
		t.Fatal("expected non-empty userID")
	}
}

func TestGetOrCreateUserID_EmptyUserID(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("APPDATA", tmpDir)

	// Создаём файл с пустым user_id
	appDir := filepath.Join(tmpDir, "CurRate")
	if err := os.MkdirAll(appDir, 0755); err != nil {
		t.Fatal(err)
	}

	data, _ := json.MarshalIndent(UserData{UserID: ""}, "", "  ")
	if err := os.WriteFile(filepath.Join(appDir, "user.json"), data, 0644); err != nil {
		t.Fatal(err)
	}

	userID, err := GetOrCreateUserID()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if userID == "" {
		t.Fatal("expected new userID to be generated for empty user_id")
	}
}

func TestGetOrCreateUserID_Idempotent(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("APPDATA", tmpDir)

	id1, err := GetOrCreateUserID()
	if err != nil {
		t.Fatalf("first call error: %v", err)
	}

	id2, err := GetOrCreateUserID()
	if err != nil {
		t.Fatalf("second call error: %v", err)
	}

	if id1 != id2 {
		t.Errorf("expected same ID on repeated calls, got %s and %s", id1, id2)
	}
}

func TestGenerateUUID(t *testing.T) {
	uuid, err := generateUUID()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Проверяем формат UUID v4
	parts := strings.Split(uuid, "-")
	if len(parts) != 5 {
		t.Fatalf("expected 5 parts, got %d: %s", len(parts), uuid)
	}

	// Проверяем версию (4) — третья группа начинается с '4'
	if parts[2][0] != '4' {
		t.Errorf("expected version 4 (third group starts with '4'), got %s", parts[2])
	}

	// Проверяем вариант RFC 4122 — четвёртая группа начинается с '8', '9', 'a', 'b'
	firstChar := parts[3][0]
	if firstChar != '8' && firstChar != '9' && firstChar != 'a' && firstChar != 'b' {
		t.Errorf("expected RFC 4122 variant (fourth group starts with 8/9/a/b), got %c", firstChar)
	}
}
