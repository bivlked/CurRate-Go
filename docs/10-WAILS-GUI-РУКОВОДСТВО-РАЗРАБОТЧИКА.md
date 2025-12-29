# CurRate-Go - Руководство разработчика Desktop GUI

> **Версия документа:** 1.1
> **Дата создания:** 2025-12-22
> **Обновлено:** 2025-12-24
> **Статус:** Актуально (соответствует реализации)
> **Целевая аудитория:** Разработчики, контрибьюторы

---

## 📖 Содержание

1. [Введение](#введение)
2. [Требования для разработки](#требования-для-разработки)
3. [Установка окружения](#установка-окружения)
4. [Структура проекта Wails](#структура-проекта-wails)
5. [Запуск в режиме разработки](#запуск-в-режиме-разработки)
6. [Сборка production версии](#сборка-production-версии)
7. [Отладка приложения](#отладка-приложения)
8. [Модификация кода](#модификация-кода)
9. [Тестирование GUI](#тестирование-gui)
10. [Архитектурные решения](#архитектурные-решения)
11. [Best Practices](#best-practices)
12. [Troubleshooting](#troubleshooting)
13. [CI/CD](#cicd)

---

## Введение

Это техническое руководство для разработчиков, работающих над desktop GUI версией CurRate-Go, построенной на Wails v2.11.0.

### Что такое Wails

**Wails** — это фреймворк для создания desktop приложений с использованием Go для backend и HTML/CSS/JavaScript для frontend. Wails использует WebView2 (на Windows) для рендеринга UI, что даёт:

- ✅ Нативную производительность Go backend
- ✅ Современный HTML/CSS/JS frontend без Electron
- ✅ Маленький размер бинарника (~8-10 МБ)
- ✅ Автоматическое связывание Go ↔ JavaScript
- ✅ Кросс-платформенность (Windows, macOS, Linux)

### Архитектура приложения

```
┌─────────────────────────────────────────┐
│          Frontend (WebView2)            │
│     HTML/CSS/JS (Vanilla, no React)     │
│  - index.html, main.js, calendar.js     │
│  - Wails Runtime API для вызова Go      │
└─────────────────┬───────────────────────┘
                  │ Wails Bindings
                  │ (автогенерация)
┌─────────────────▼───────────────────────┐
│          Backend (Go)                   │
│  - app.go: App struct с методами        │
│  - Converter, Parser, Cache layers      │
│  - CBR XML API интеграция               │
└─────────────────────────────────────────┘
```

### Технологический стек

| Компонент | Технология | Версия |
|-----------|------------|--------|
| **Framework** | Wails | v2.11.0 |
| **Backend** | Go | 1.25.5 |
| **Frontend** | Vanilla JS | ES6+ |
| **UI Rendering** | WebView2 | Latest (Windows) |
| **Build Tool** | Wails CLI | v2.11.0 |
| **Package Manager** | Go modules | - |

---

## Требования для разработки

### Минимальные требования

| Компонент | Требование |
|-----------|------------|
| **Go** | 1.25.5 |
| **OS** | Windows 10 (22H2+) / Windows 11 |
| **IDE** | VS Code / GoLand / любой с Go поддержкой |
| **Git** | 2.30+ |

**Примечание:** Node.js и npm **не требуются**, так как фронтенд использует статический vanilla JavaScript без сборки.

### Рекомендуемые инструменты

- **VS Code** с расширениями:
  - Go (golang.go)
  - Wails Snippets (если доступно)
  - HTML CSS Support
  - JavaScript (ES6) code snippets
- **Git Bash** или **PowerShell Core 7+**
- **WebView2 Runtime** (для Windows 10)

### Проверка установленных инструментов

```bash
# Проверить версию Go
go version
# Ожидается: go version go1.25.5

# Проверить Git
git --version
# Ожидается: git version 2.30.x или выше

**Примечание:** Node.js и npm не требуются, так как фронтенд - статический vanilla JavaScript.
```

---

## Установка окружения

### Шаг 1: Клонировать репозиторий

```bash
# Клонировать проект
git clone https://github.com/bivlked/CurRate-Go.git
cd CurRate-Go

# Переключиться на ветку разработки GUI (если есть)
git checkout gui-development

# Установить Go зависимости
go mod download
```

### Шаг 2: Установить Wails CLI

**Важно:** Wails CLI — это основной инструмент для разработки и сборки.

```bash
# Установить последнюю версию Wails
go install github.com/wailsapp/wails/v2/cmd/wails@latest

# Проверить установку
wails version
# Ожидается: v2.11.0 или выше
```

**Добавить Wails в PATH (если нужно):**

- Wails устанавливается в `$GOPATH/bin` (обычно `%USERPROFILE%\go\bin`)
- Убедитесь, что эта папка в PATH
- Windows: `setx PATH "%PATH%;%USERPROFILE%\go\bin"`
- PowerShell: `$env:PATH += ";$env:USERPROFILE\go\bin"`

### Шаг 3: Проверить окружение

```bash
# Запустить диагностику Wails
wails doctor
```

**Ожидаемый вывод:**

```
Wails CLI v2.11.0

Scanning system - Please wait (this may take a long time)...

# System
OS:           Windows 11 x64
Version:      22H2 (Build: 22621)
ID:           windows
Go Version:   go1.25.5
Platform:     windows
Architecture: amd64

# Wails
Version:      v2.11.0

# Dependencies
Dependency            Package Name    Status      Version
WebView2              N/A             Installed   120.0.2210.144
npm                   N/A             Installed   8.19.3
*upx                  N/A             Available   -

# Diagnosis
Your system is ready for Wails development!
```

**Если есть ошибки:**
- **WebView2 не установлен:** скачайте https://go.microsoft.com/fwlink/p/?LinkId=2124703
- **upx не найден:** опционально для сжатия бинарника (см. ниже)

**Примечание:** npm и Node.js не требуются, так как фронтенд использует статический vanilla JavaScript.

### Шаг 4: Установить UPX (опционально, для сжатия)

**UPX** (Ultimate Packer for eXecutables) уменьшает размер бинарника на 30-40%.

**Windows:**

```bash
# С помощью Scoop (рекомендуется)
scoop install upx

# Или скачать вручную
# https://upx.github.io/
# Распаковать в любую папку и добавить в PATH
```

**Проверка:**

```bash
upx --version
# Ожидается: UPX 4.x.x
```

### Шаг 5: Установить дополнительные инструменты (опционально)

**GolangCI-Lint** для статического анализа:

```bash
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Проверка
golangci-lint --version
```

---

## Структура проекта Wails

После создания GUI проекта структура будет следующей:

```
CurRate-Go/
├── main_gui.go                  # Entry point (GUI приложение)
│
├── internal/                    # Бизнес-логика
│   ├── models/                  # Модели данных
│   ├── parser/                  # XML парсер ЦБ РФ
│   ├── cache/                   # LRU кэш
│   └── converter/               # Конвертер валют
│
├── pkg/utils/                   # Утилиты
│
├── internal/app/               # 🆕 GUI код (Wails)
│   └── app.go                   # 🆕 Backend: App struct, Go ↔ JS bindings
│
├── main_gui.go                  # 🆕 Entry point для GUI версии
│
├── frontend/                    # 🆕 Frontend код
│   ├── index.html               # 🆕 Главный HTML
│   ├── scripts/
│   │   ├── main.js              # 🆕 Основная логика JS
│   │   ├── calendar.js          # 🆕 Календарь с выделением выходных
│   │   └── utils.js             # 🆕 Утилиты (форматирование, валидация)
│   ├── styles/
│   │   ├── main.css             # 🆕 Основные стили
│   │   └── calendar.css         # 🆕 Стили календаря
│   └── wailsjs/                 # 🆕 Автогенерируемые Wails bindings
│
├── build/                       # 🆕 Конфигурация сборки
│   ├── windows/
│   │   ├── icon.ico             # 🆕 Иконка для Windows
│   │   └── wails.exe.manifest   # 🆕 Манифест приложения
│   ├── darwin/                  # macOS (если будет сборка)
│   └── linux/                   # Linux (если будет сборка)
│
├── wails.json                   # 🆕 Конфигурация Wails
├── go.mod
├── go.sum
├── .gitignore
├── README.md
└── docs/
    ├── 01-ТЕХНИЧЕСКОЕ-ЗАДАНИЕ.md
    ├── ...
    ├── 08-WAILS-GUI-ПЛАН.md
    ├── 09-WAILS-GUI-РУКОВОДСТВО-ПОЛЬЗОВАТЕЛЯ.md    # Этот документ
    └── 10-WAILS-GUI-РУКОВОДСТВО-РАЗРАБОТЧИКА.md    # 🔥 Вы здесь
```

### Ключевые файлы

#### `wails.json` — Конфигурация проекта

```json
{
  "$schema": "https://wails.io/schemas/config.v2.json",
  "name": "CurRate-Go",
  "outputfilename": "CurRate",
  "frontend:install": "",
  "frontend:build": "",
  "assetdir": "frontend",
  "wailsjsdir": "frontend/wailsjs",
  "author": {
    "name": "Ivan Bondarev",
    "email": "your-email@example.com"
  },
  "info": {
    "companyName": "New Digital Technologies Ltd.",
    "productName": "CurRate-Go",
    "productVersion": "1.0.0",
    "copyright": "© 2025 Ivan Bondarev",
    "comments": "Currency converter based on CBR RF official rates"
  }
}
```

#### `main_gui.go` — Entry point GUI приложения

```go
package main

import (
	"context"
	"embed"
	"log"
	"time"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/windows"

	"github.com/bivlked/currate-go/internal/cache"
	"github.com/bivlked/currate-go/internal/converter"
	"github.com/bivlked/currate-go/internal/parser"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// Создать конвертер (бизнес-логика)
	cacheStorage := cache.NewLRUCache(100, 24*time.Hour)
	conv := converter.NewConverter(parser.FetchRates, cacheStorage)

	// Создать App instance (GUI backend)
	app := NewApp(conv)

	// Запустить Wails приложение
	err := wails.Run(&options.App{
		Title:  "CurRate-Go - Конвертер валют",
		Width:  800,
		Height: 650,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 255, G: 255, B: 255, A: 1},
		OnStartup:        app.startup,
		Bind: []interface{}{
			app, // Привязать методы App к JS
		},
		Windows: &windows.Options{
			WebviewIsTransparent: false,
			WindowIsTranslucent:  false,
			DisableWindowIcon:    false,
		},
	})

	if err != nil {
		log.Fatal("Ошибка запуска приложения:", err)
	}
}
```

#### `internal/app/app.go` — Backend структура

```go
package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/bivlked/currate-go/internal/converter"
	"github.com/bivlked/currate-go/internal/models"
)

// App struct — основной backend для GUI
type App struct {
	ctx       context.Context
	converter *converter.Converter
}

// NewApp создаёт новый экземпляр App
func NewApp(conv *converter.Converter) *App {
	return &App{
		converter: conv,
	}
}

// startup вызывается при запуске приложения
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// ConvertRequest — запрос на конвертацию из JS
type ConvertRequest struct {
	Amount   float64 `json:"amount"`   // Сумма для конвертации
	Currency string  `json:"currency"` // "USD" или "EUR"
	Date     string  `json:"date"`     // "DD.MM.YYYY"
}

// ConvertResponse — ответ на конвертацию для JS
type ConvertResponse struct {
	Success   bool    `json:"success"`
	Result    string  `json:"result"`    // Форматированная строка
	RubAmount float64 `json:"rubAmount"` // Сумма в рублях
	Rate      float64 `json:"rate"`      // Курс
	Error     string  `json:"error"`     // Сообщение об ошибке (если есть)
}

// Convert выполняет конвертацию валюты
// Вызывается из JavaScript через Wails bindings
func (a *App) Convert(req ConvertRequest) ConvertResponse {
	// Валидация и парсинг даты
	date, err := time.Parse("02.01.2006", req.Date)
	if err != nil {
		return ConvertResponse{
			Success: false,
			Error:   "Некорректный формат даты. Используйте ДД.ММ.ГГГГ",
		}
	}

	// Парсинг валюты
	currency, err := models.ParseCurrency(req.Currency)
	if err != nil {
		return ConvertResponse{
			Success: false,
			Error:   fmt.Sprintf("Неподдерживаемая валюта: %s", req.Currency),
		}
	}

	// Конвертация через существующую бизнес-логику
	result, err := a.converter.Convert(req.Amount, currency, models.RUB, date)
	if err != nil {
		return ConvertResponse{
			Success: false,
			Error:   handleError(err),
		}
	}

	// Успешный результат
	return ConvertResponse{
		Success: true,
		Result:  result.FormattedStr,
	}
}

// handleError обрабатывает ошибки и возвращает понятное сообщение
func handleError(err error) string {
	switch {
	case errors.Is(err, converter.ErrInvalidAmount):
		return "Сумма должна быть больше нуля"
	case errors.Is(err, converter.ErrDateInFuture):
		return "Дата не может быть в будущем"
	case errors.Is(err, models.ErrUnsupportedCurrency):
		return "Валюта не поддерживается"
	default:
		return fmt.Sprintf("Ошибка: %s", err.Error())
	}
}
```

#### `frontend/scripts/main.js` — Frontend логика

```javascript
// Глобальные переменные
let calendar = null;

// Инициализация при загрузке DOM
document.addEventListener('DOMContentLoaded', () => {
    initApp();
});

async function initApp() {
    // Установить текущую дату (используя JavaScript Date)
    const today = new Date();
    const todayDate = formatDate(today);
    document.getElementById('date-input').value = todayDate;

    // Инициализировать календарь
    calendar = new Calendar('calendar-container', onDateSelect);

    // Установить обработчики событий
    setupEventListeners();
    
    // Загрузить курс валюты на текущую дату
    await loadRateForDate(todayDate);
}

function setupEventListeners() {
    // Кнопка конвертации
    document.getElementById('convert-btn').addEventListener('click', handleConvert);

    // Enter в поле суммы
    document.getElementById('amount-input').addEventListener('keypress', (e) => {
        if (e.key === 'Enter') handleConvert();
    });

    // Кнопка календаря
    document.getElementById('calendar-btn').addEventListener('click', toggleCalendar);

    // Кнопка копирования
    document.getElementById('copy-btn').addEventListener('click', copyToClipboard);

    // Валидация суммы в реальном времени
    document.getElementById('amount-input').addEventListener('input', validateAmount);
}

async function handleConvert() {
    const amount = parseFloat(document.getElementById('amount-input').value);
    const currency = document.querySelector('input[name="currency"]:checked').value;
    const date = document.getElementById('date-input').value;

    // Валидация
    if (!amount || amount <= 0) {
        showError('Введите корректную сумму');
        return;
    }

    // Показать состояние загрузки
    setLoadingState(true);

    try {
        // Вызвать Go метод через Wails bindings
        const response = await window.go.app.App.Convert({
            amount: amount,
            currency: currency,
            date: date
        });

        if (response.success) {
            // Показать результат
            document.getElementById('result-text').textContent = response.result;
            document.getElementById('copy-btn').disabled = false;
            clearError();
        } else {
            // Показать ошибку
            showError(response.error);
        }
    } catch (error) {
        showError('Ошибка подключения к backend');
    } finally {
        setLoadingState(false);
    }
}

function setLoadingState(loading) {
    const btn = document.getElementById('convert-btn');
    if (loading) {
        btn.disabled = true;
        btn.textContent = 'Загрузка...';
    } else {
        btn.disabled = false;
        btn.textContent = 'Конвертировать';
    }
}

function showError(message) {
    const errorEl = document.getElementById('error-message');
    errorEl.textContent = message;
    errorEl.style.display = 'block';
}

function clearError() {
    document.getElementById('error-message').style.display = 'none';
}

function copyToClipboard() {
    const result = document.getElementById('result-text').textContent;
    navigator.clipboard.writeText(result).then(() => {
        // Показать индикацию успешного копирования
        const btn = document.getElementById('copy-btn');
        const originalText = btn.textContent;
        btn.textContent = '✓ Скопировано!';
        setTimeout(() => {
            btn.textContent = originalText;
        }, 2000);
    });
}

function toggleCalendar() {
    calendar.toggle();
}

function onDateSelect(date) {
    document.getElementById('date-input').value = formatDate(date);
}

function formatDate(date) {
    const day = String(date.getDate()).padStart(2, '0');
    const month = String(date.getMonth() + 1).padStart(2, '0');
    const year = date.getFullYear();
    return `${day}.${month}.${year}`;
}

function validateAmount() {
    const input = document.getElementById('amount-input');
    const value = parseFloat(input.value);

    if (isNaN(value) || value <= 0) {
        input.classList.add('invalid');
    } else {
        input.classList.remove('invalid');
    }
}
```

---

## Запуск в режиме разработки

Режим разработки (`wails dev`) — это главный инструмент для разработки GUI. Он обеспечивает:

- 🔄 **Hot Reload** — изменения в frontend применяются мгновенно
- 🐛 **DevTools** — встроенные инструменты отладки браузера
- 📝 **Логирование** — подробные логи Go backend в консоли
- ⚡ **Быстрый старт** — автоматическая перекомпиляция Go кода

### Команда запуска

```bash
# Перейти в папку проекта
cd CurRate-Go

# Запустить в режиме разработки
wails dev
```

**Что произойдёт:**

1. Wails скомпилирует Go код
2. Запустит встроенный веб-сервер для frontend
3. Откроет окно приложения с DevTools
4. Начнёт отслеживать изменения в файлах

### Вывод в консоли

```
DEB | [ExternalAssetHandler] Serving assets from frontend directory
DEB | [Wails] Using DevServer URL: http://localhost:34115
DEB | [Wails] Launching application in Dev mode...
DEB | [App] Application started
```

### Hot Reload

**Frontend изменения (HTML/CSS/JS):**
- Сохраните файл в `frontend/scripts/`
- Браузер автоматически перезагрузится
- Изменения видны мгновенно

**Backend изменения (Go):**
- Сохраните файл в `internal/app/`
- Wails автоматически перекомпилирует
- Приложение перезапустится (~2-3 секунды)

### Доступ к DevTools

**Открыть DevTools:**
- **Правый клик** в окне приложения → "Inspect" (Проверить)
- Или нажмите **F12**

**Возможности DevTools:**
- Console — логи JavaScript, ошибки
- Network — запросы к backend (Go методы)
- Elements — инспекция HTML/CSS
- Sources — отладка JS с breakpoints
- Performance — профилирование производительности

### Логирование

**В Go коде:**

```go
import "log"

func (a *App) Convert(req ConvertRequest) ConvertResponse {
    log.Printf("Convert called: amount=%.2f, currency=%s, date=%s",
        req.Amount, req.Currency, req.Date)

    // ... логика конвертации
}
```

**Логи появятся в терминале**, где запущен `wails dev`.

**В JavaScript:**

```javascript
console.log('Convert button clicked');
console.error('Validation failed:', error);
```

**Логи появятся в DevTools Console**.

### Остановка

- Нажмите **Ctrl+C** в терминале
- Или закройте окно приложения

---

## Сборка production версии

Production сборка создаёт оптимизированный бинарник для конечных пользователей.

### Автоматизированная сборка (рекомендуется)

Для удобства создан скрипт `build.bat`, который автоматизирует процесс сборки:

```bash
# Development build (без оптимизации)
build.bat dev

# Production build (оптимизированный, без UPX)
build.bat prod

# Production build с UPX сжатием (минимальный размер)
build.bat prod-upx
```

**Преимущества build.bat:**
- ✅ Автоматическая проверка результата сборки
- ✅ Отображение размера собранного файла
- ✅ Опциональное UPX сжатие с проверкой наличия UPX
- ✅ Подробный вывод информации о процессе сборки
- ✅ Обработка ошибок

### Ручная сборка

Если вы предпочитаете использовать команды напрямую:

```bash
# Сборка с настройками по умолчанию
wails build

# Production сборка с оптимизацией
wails build -clean -ldflags "-s -w"

# Production сборка с UPX сжатием
wails build -clean -upx -ldflags "-s -w"

# Сборка с отладочной информацией
wails build -debug

# Сборка для конкретной платформы
wails build -platform windows/amd64
```

### Параметры сборки

| Флаг | Описание | Пример |
|------|----------|--------|
| `-clean` | Очистить build кэш перед сборкой | `wails build -clean` |
| `-upx` | Сжать бинарник с помощью UPX | `wails build -upx` |
| `-debug` | Включить режим отладки | `wails build -debug` |
| `-ldflags` | Передать флаги компоновщику | `wails build -ldflags "-s -w"` |
| `-platform` | Целевая платформа | `wails build -platform windows/amd64` |
| `-o` | Имя выходного файла | `wails build -o CurRate.exe` |

### Процесс сборки

```
Wails CLI v2.11.0
• Building application...
  - Minifying frontend assets...
  - Compiling Go code...
  - Embedding assets...
  - Linking binary...
  - Compressing with UPX... (если -upx)
• Build complete!
  Binary: build/bin/CurRate.exe
  Size: 8.2 MB
```

### Результаты сборки

```
build/
└── bin/
    ├── CurRate.exe         # Основной бинарник (Windows)
    └── ...
```

### Размеры бинарника

| Режим | Размер | Примечание |
|-------|--------|------------|
| **Без оптимизации** | ~12-15 МБ | `wails build` |
| **С -ldflags "-s -w"** | ~10-12 МБ | Удаление символов отладки |
| **С UPX сжатием** | ~8-10 МБ | `wails build -upx` |

**Рекомендация для релиза:**

```bash
# Используйте build.bat для автоматизации
build.bat prod-upx

# Или вручную:
wails build -clean -upx -ldflags "-s -w"
```

**Примечание:** Для UPX сжатия необходимо установить UPX:
- Chocolatey: `choco install upx`
- Или скачать вручную: https://upx.github.io/

### Создание установщика (Windows)

Для создания установщика (`.msi` или `.exe`) можно использовать **NSIS** или **Inno Setup**.

**Пример с Inno Setup:**

1. Установить Inno Setup: https://jrsoftware.org/isdl.php
2. Создать скрипт `installer.iss`:

```ini
[Setup]
AppName=CurRate-Go
AppVersion=1.0.0
DefaultDirName={pf}\CurRate-Go
DefaultGroupName=CurRate-Go
OutputDir=build\installer
OutputBaseFilename=CurRate-Setup-v1.0.0
Compression=lzma2
SolidCompression=yes

[Files]
Source: "build\bin\CurRate.exe"; DestDir: "{app}"; Flags: ignoreversion

[Icons]
Name: "{group}\CurRate-Go"; Filename: "{app}\CurRate.exe"
Name: "{commondesktop}\CurRate-Go"; Filename: "{app}\CurRate.exe"

[Run]
Filename: "{app}\CurRate.exe"; Description: "Launch CurRate-Go"; Flags: postinstall nowait
```

3. Скомпилировать:

```bash
iscc installer.iss
```

---

## Отладка приложения

### Отладка Go backend

**1. Логирование:**

```go
import "log"

log.Printf("Debug: received request %+v", req)
log.Println("Converting currency...")
```

**2. Breakpoints в IDE:**

- **VS Code:** используйте расширение Go и конфигурацию launch.json
- **GoLand:** встроенный отладчик

**3. Delve (Go debugger):**

```bash
# Установка
go install github.com/go-delve/delve/cmd/dlv@latest

# Запуск с отладчиком
dlv debug main_gui.go
```

### Отладка JavaScript frontend

**1. Console.log:**

```javascript
console.log('Amount:', amount);
console.error('Error:', error);
console.table(response);
```

**2. Breakpoints в DevTools:**

- Откройте DevTools (F12)
- Перейдите на вкладку **Sources**
- Найдите файл `main.js`
- Кликните на номер строки для установки breakpoint
- Код остановится на этой строке

**3. Network inspection:**

- Откройте DevTools → Network
- Все вызовы Go методов через Wails bindings видны как XHR запросы

### Отладка Wails bindings

Если вызовы Go методов из JS не работают:

**1. Проверить, что методы экспортированы (начинаются с заглавной буквы):**

```go
// ✅ Правильно
func (a *App) Convert(req ConvertRequest) ConvertResponse { ... }

// ❌ Неправильно (не экспортирован)
func (a *App) convert(req ConvertRequest) ConvertResponse { ... }
```

**2. Проверить, что App привязан в `main.go`:**

```go
Bind: []interface{}{
    app, // Должен быть здесь
},
```

**3. Проверить имя вызова в JS:**

```javascript
// Правильное имя: window.go.<пакет>.<Структура>.<Метод>
await window.go.app.App.Convert(...)
```

**4. Проверить типы данных:**

- Go использует JSON для сериализации
- JS объекты автоматически конвертируются
- Убедитесь, что структуры имеют JSON tags

### Логи Wails Runtime

Для включения подробных логов Wails:

```bash
# Установить переменную окружения
set WAILS_LOGLEVEL=Debug

# Запустить
wails dev
```

---

## Модификация кода

### Добавление нового Go метода

**Шаг 1:** Добавить метод в `internal/app/app.go`:

```go
// GetRate получает курс валюты на указанную дату (для live preview)
func (a *App) GetRate(currencyStr string, dateStr string) RateResponse {
    // Парсинг валюты
    currency, err := models.ParseCurrency(currencyStr)
    if err != nil {
        return RateResponse{
            Success: false,
            Error:   fmt.Sprintf("Неподдерживаемая валюта: %s", currencyStr),
        }
    }

    // Парсинг даты
    date, err := parseDate(dateStr)
    if err != nil {
        return RateResponse{
            Success: false,
            Error:   fmt.Sprintf("Неверный формат даты: %s", dateStr),
        }
    }

    // Используем оптимизированный метод GetRate
    rate, err := a.converter.GetRate(currency, date)
    if err != nil {
        return RateResponse{
            Success: false,
            Error:   translateError(err),
        }
    }

    return RateResponse{
        Success: true,
        Rate:    rate,
    }
}
```

**Шаг 2:** Использовать в JavaScript:

```javascript
async function fetchRate() {
    const currency = 'USD';
    const date = '22.12.2025';

    try {
        const response = await window.go.app.App.GetRate(currency, date);
        if (response.success) {
            console.log(`Курс ${currency}: ${response.rate}`);
        } else {
            console.error('Ошибка получения курса:', response.error);
        }
    } catch (error) {
        console.error('Ошибка вызова метода:', error);
    }
}
```

**Шаг 3:** Перезапустить `wails dev` для перекомпиляции.

### Изменение UI

**HTML (`frontend/index.html`):**

```html
<!-- Добавить новый элемент -->
<div class="info-block">
    <p id="last-update">Последнее обновление: <span id="last-update-time"></span></p>
</div>
```

**CSS (`frontend/styles/main.css`):**

```css
.info-block {
    margin-top: 10px;
    padding: 10px;
    background: #f0f0f0;
    border-radius: 5px;
    text-align: center;
    font-size: 12px;
    color: #666;
}
```

**JavaScript (`frontend/scripts/main.js`):**

```javascript
function updateLastUpdateTime() {
    const now = new Date();
    const formatted = now.toLocaleString('ru-RU');
    document.getElementById('last-update-time').textContent = formatted;
}

// Вызвать при загрузке
updateLastUpdateTime();
```

**Результат:** Изменения применятся автоматически в `wails dev` режиме.

### Добавление новой валюты

**Шаг 1:** Обновить `internal/models/currency.go`:

```go
const (
    USD Currency = "USD" // Доллар США
    EUR Currency = "EUR" // Евро
    GBP Currency = "GBP" // 🆕 Фунт стерлингов
    CNY Currency = "CNY" // 🆕 Китайский юань
)

func (c Currency) Validate() error {
    switch c {
    case USD, EUR, GBP, CNY: // Добавить новые
        return nil
    default:
        return fmt.Errorf("%w: %s", ErrUnsupportedCurrency, c)
    }
}
```

**Шаг 2:** Обновить UI (`frontend/index.html`):

```html
<div class="currency-selection">
    <label><input type="radio" name="currency" value="USD" checked> USD</label>
    <label><input type="radio" name="currency" value="EUR"> EUR</label>
    <label><input type="radio" name="currency" value="GBP"> GBP</label>
    <label><input type="radio" name="currency" value="CNY"> CNY</label>
</div>
```

**Шаг 3:** Обновить парсер `internal/parser/parser.go`:

```go
func parseCurrency(code string) (models.Currency, error) {
    switch strings.TrimSpace(strings.ToUpper(code)) {
    case "USD", "840":
        return models.USD, nil
    case "EUR", "978":
        return models.EUR, nil
    case "GBP", "826": // 🆕
        return models.GBP, nil
    case "CNY", "156": // 🆕
        return models.CNY, nil
    default:
        return "", fmt.Errorf("unknown currency code: %s", code)
    }
}
```

**Шаг 4:** Добавить тесты.

---

## Тестирование GUI

### Unit тесты (Go backend)

Существующие тесты уже покрывают бизнес-логику (96% coverage). Для GUI нужно добавить тесты `internal/app/app.go`:

**Создать `internal/app/app_test.go`:**

```go
package main

import (
	"testing"
	"time"

	"github.com/bivlked/currate-go/internal/cache"
	"github.com/bivlked/currate-go/internal/converter"
	"github.com/bivlked/currate-go/internal/models"
	"github.com/bivlked/currate-go/internal/parser"
)

func TestApp_Convert_Success(t *testing.T) {
	// Setup
	cacheStorage := cache.NewLRUCache(10, 1*time.Hour)
	conv := converter.NewConverter(parser.FetchRates, cacheStorage)
	app := NewApp(conv)

	// Test
	req := ConvertRequest{
		Amount:   1000,
		Currency: "USD",
		Date:     time.Now().Format("02.01.2006"),
	}

	resp := app.Convert(req)

	// Assert
	if !resp.Success {
		t.Errorf("Expected success=true, got error: %s", resp.Error)
	}
	if resp.RubAmount <= 0 {
		t.Errorf("Expected positive RubAmount, got: %.2f", resp.RubAmount)
	}
}

func TestApp_Convert_InvalidDate(t *testing.T) {
	cacheStorage := cache.NewLRUCache(10, 1*time.Hour)
	conv := converter.NewConverter(parser.FetchRates, cacheStorage)
	app := NewApp(conv)

	req := ConvertRequest{
		Amount:   1000,
		Currency: "USD",
		Date:     "invalid-date",
	}

	resp := app.Convert(req)

	if resp.Success {
		t.Error("Expected success=false for invalid date")
	}
	if resp.Error == "" {
		t.Error("Expected error message for invalid date")
	}
}

func TestApp_GetTodayDate(t *testing.T) {
	app := NewApp(nil)
	date := app.GetTodayDate()

	// Проверить формат DD.MM.YYYY
	_, err := time.Parse("02.01.2006", date)
	if err != nil {
		t.Errorf("GetTodayDate returned invalid format: %s", date)
	}
}
```

**Запуск:**

```bash
go test ./internal/app/...
```

### E2E тесты (Frontend)

Для тестирования frontend можно использовать **Playwright** или **Selenium**.

**Пример с Playwright:**

```bash
# Примечание: Для E2E тестов требуется Node.js и npm
# В текущем проекте E2E тесты не используются, так как фронтенд - статический vanilla JS
# Если планируется добавить E2E тесты, установите:
# npm install --save-dev @playwright/test
```

**Создать `frontend/tests/e2e.spec.js`:**

```javascript
const { test, expect } = require('@playwright/test');

test('convert USD to RUB', async ({ page }) => {
    await page.goto('http://localhost:34115');

    // Ввести сумму
    await page.fill('#amount-input', '1000');

    // Выбрать USD
    await page.check('input[value="USD"]');

    // Нажать конвертировать
    await page.click('#convert-btn');

    // Дождаться результата
    await page.waitForSelector('#result-text', { timeout: 5000 });

    // Проверить, что результат содержит "руб."
    const result = await page.textContent('#result-text');
    expect(result).toContain('руб.');
});
```

**Запуск:**

```bash
npx playwright test
```

### Ручное тестирование (Checklist)

- [ ] Приложение запускается без ошибок
- [ ] Текущая дата устанавливается автоматически
- [ ] Календарь открывается по клику на 📅
- [ ] Выходные дни (Сб, Вс) выделены красным
- [ ] Можно ввести дату вручную в формате DD.MM.YYYY
- [ ] Выбор USD/EUR работает
- [ ] Ввод суммы валидируется (только числа > 0)
- [ ] Кнопка "Конвертировать" работает
- [ ] Результат отображается корректно
- [ ] Кнопка "Копировать в буфер" копирует результат
- [ ] Ошибки валидации показываются красным
- [ ] Ошибки сети обрабатываются корректно
- [ ] Повторные запросы используют кэш (мгновенный результат)

---

## Архитектурные решения

### Почему Vanilla JS вместо React/Vue?

**Решение:** Использовать Vanilla JavaScript без фреймворков.

**Причины:**

1. **Размер бинарника:**
   - Vanilla JS: ~8-10 МБ
   - React: ~18-25 МБ (из-за бандла)

2. **Простота проекта:**
   - Всего 7 UI элементов
   - Нет сложного state management
   - Нет роутинга

3. **Производительность:**
   - Нет виртуального DOM overhead
   - Прямая манипуляция DOM быстрее для простых UI

4. **Легкость обучения:**
   - Контрибьюторам не нужно знать React/Vue
   - Стандартный JavaScript API

**Когда переходить на фреймворк:**
- Если UI станет намного сложнее (>20 компонентов)
- Если появится необходимость в state management
- Если потребуется роутинг между страницами

### Структура кода: Монорепо vs Раздельные репозитории

**Решение:** Монорепо — GUI и CLI в одном репозитории.

**Причины:**

1. **Общая бизнес-логика:**
   - GUI и CLI используют одни и те же `internal/` пакеты
   - Нет дублирования кода

2. **Единый процесс релиза:**
   - Версия CLI = версия GUI
   - Одновременное тестирование

3. **Упрощённая разработка:**
   - Один `go.mod`, один набор зависимостей
   - Изменения в `internal/` сразу доступны обоим

**Структура:**

```
CurRate-Go/
├── main_gui.go       # GUI entry point
└── internal/         # Общая бизнес-логика
    └── app/          # GUI backend (App struct)
```

### Кэширование на уровне GUI

**Решение:** Не добавлять дополнительный кэш в JavaScript, использовать только Go backend кэш.

**Причины:**

1. **Избежать дублирования:**
   - LRU кэш уже есть в Go backend
   - Два кэша = сложность синхронизации

2. **Память:**
   - Go кэш живёт весь lifetime приложения
   - JS кэш занимал бы дополнительную память

3. **Простота:**
   - Один источник истины (Go backend)

**Как это работает:**

```
JavaScript → window.go.app.App.Convert()
              ↓
          Go App.Convert()
              ↓
          Converter.Convert()
              ↓
   Проверка LRU кэша → возврат из кэша (мгновенно)
   Если нет → запрос к CBR API → сохранение в кэш
```

---

## Best Practices

### Go Backend

**1. Всегда экспортируйте методы для Wails:**

```go
// ✅ Правильно
func (a *App) Convert(...) { }

// ❌ Неправильно
func (a *App) convert(...) { }
```

**2. Используйте JSON tags для структур:**

```go
type ConvertRequest struct {
    Amount   float64 `json:"amount"`
    Currency string  `json:"currency"`
    Date     string  `json:"date"`
}
```

**3. Обрабатывайте ошибки:**

```go
func (a *App) Convert(req ConvertRequest) ConvertResponse {
    if err != nil {
        return ConvertResponse{
            Success: false,
            Error:   handleError(err),
        }
    }
    // ...
}
```

**4. Логируйте важные операции:**

```go
log.Printf("Convert: %.2f %s on %s", req.Amount, req.Currency, req.Date)
```

### JavaScript Frontend

**1. Используйте async/await для Go вызовов:**

```javascript
// ✅ Правильно
async function handleConvert() {
    const response = await window.go.app.App.Convert(...);
}

// ❌ Неправильно (промисы без обработки)
function handleConvert() {
    window.go.app.App.Convert(...); // Результат игнорируется
}
```

**2. Всегда обрабатывайте ошибки:**

```javascript
try {
    const response = await window.go.app.App.Convert(...);
    if (!response.success) {
        showError(response.error);
    }
} catch (error) {
    showError('Ошибка подключения');
}
```

**3. Валидируйте ввод на фронтенде:**

```javascript
if (!amount || amount <= 0) {
    showError('Введите корректную сумму');
    return;
}
```

**4. Используйте функции для переиспользования:**

```javascript
function formatDate(date) {
    // Одна функция для форматирования
}

function parseDate(str) {
    // Одна функция для парсинга
}
```

### CSS

**1. Используйте CSS переменные для цветов:**

```css
:root {
    --primary-color: #2196F3;
    --error-color: #d32f2f;
    --weekend-color: #d32f2f;
    --weekend-bg: #ffebee;
}

.calendar-day.weekend {
    color: var(--weekend-color);
    background: var(--weekend-bg);
}
```

**2. Группируйте связанные стили:**

```css
/* Календарь */
.calendar-container { ... }
.calendar-header { ... }
.calendar-day { ... }
```

**3. Используйте понятные имена классов:**

```css
/* ✅ Хорошо */
.calendar-day.weekend { }
.error-message { }

/* ❌ Плохо */
.cd.we { }
.em { }
```

---

## Troubleshooting

### Проблема: `wails: command not found`

**Причина:** Wails CLI не добавлен в PATH.

**Решение:**

```bash
# Найти путь к Go bin
go env GOPATH
# Вывод: C:\Users\<user>\go

# Добавить в PATH
setx PATH "%PATH%;%USERPROFILE%\go\bin"

# Перезапустить терминал
```

### Проблема: `WebView2 not found`

**Причина:** WebView2 Runtime не установлен (Windows 10).

**Решение:**

- Скачать: https://go.microsoft.com/fwlink/p/?LinkId=2124703
- Установить
- Перезапустить `wails dev`

### Проблема: Методы Go не вызываются из JS

**Причины и решения:**

1. **Метод не экспортирован:**
   ```go
   // Должен начинаться с заглавной буквы
   func (a *App) Convert(...) { } // ✅
   ```

2. **App не привязан в main.go:**
   ```go
   Bind: []interface{}{
       app, // Должен быть здесь
   },
   ```

3. **Неправильное имя в JS:**
   ```javascript
   // Правильно
   window.go.app.App.Convert(...)

   // Неправильно
   window.go.App.Convert(...) // Пропущен пакет
   ```

4. **Приложение не перезапущено после изменения Go:**
   - Остановите `wails dev` (Ctrl+C)
   - Запустите снова

### Проблема: Hot Reload не работает

**Решение:**

```bash
# Полностью очистить кэш
wails dev -clean

# Если не помогает
rm -rf build/
rm -rf frontend/node_modules/
# Node.js и npm не требуются для текущего проекта
wails dev
```

### Проблема: Большой размер бинарника

**Решение:**

```bash
# Примечание: Удаление node_modules не требуется, так как фронтенд не использует npm
# Использовать все оптимизации
wails build -clean -upx -ldflags "-s -w"

# Проверить размер
ls -lh build/bin/CurRate.exe
```

**Ожидаемые размеры:**
- Без оптимизаций: ~12-15 МБ
- С оптимизациями: ~8-10 МБ

---

## CI/CD

### GitHub Actions для автоматической сборки

> ⚠️ **Примечание:** Раздел CI/CD описывает рекомендуемую конфигурацию. Реальные workflow-файлы могут отсутствовать в репозитории. Для добавления CI/CD следуйте примерам ниже.

**Создать `.github/workflows/build.yml`:**

```yaml
name: Build GUI

on:
  push:
    branches: [ main, gui-development ]
  pull_request:
    branches: [ main ]

jobs:
  build:
    runs-on: windows-latest

    steps:
      - name: Checkout code
        uses: actions/checkout@v3

      - name: Setup Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.25.5'

      - name: Install Wails
        run: go install github.com/wailsapp/wails/v2/cmd/wails@latest

      - name: Install dependencies
        run: |
          go mod download
          # Примечание: npm install не требуется, так как фронтенд - статический vanilla JS

      - name: Build
        run: wails build -clean

      - name: Upload artifact
        uses: actions/upload-artifact@v3
        with:
          name: CurRate-Windows-amd64
          path: build/bin/CurRate.exe
```

### Автоматический релиз

**Создать `.github/workflows/release.yml`:**

```yaml
name: Release

on:
  push:
    tags:
      - 'v*'

jobs:
  release:
    runs-on: windows-latest

    steps:
      - name: Checkout
        uses: actions/checkout@v3

      - name: Setup Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.25.5'

      - name: Install Wails
        run: go install github.com/wailsapp/wails/v2/cmd/wails@latest

      - name: Build
        run: wails build -clean -upx -ldflags "-s -w"

      - name: Create Release
        uses: softprops/action-gh-release@v1
        with:
          files: build/bin/CurRate.exe
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

**Использование:**

```bash
# Создать тег
git tag v1.0.0
git push origin v1.0.0

# GitHub Actions автоматически:
# 1. Соберёт приложение
# 2. Создаст релиз
# 3. Загрузит бинарник
```

---

## Заключение

Это руководство охватывает все аспекты разработки desktop GUI для CurRate-Go с использованием Wails v2.

### Быстрый старт для нового разработчика

1. **Установить инструменты:** Go 1.25.5, Wails CLI
   - **Примечание:** Node.js и npm не требуются, так как фронтенд - статический vanilla JavaScript
2. **Клонировать репозиторий:** `git clone ...`
3. **Запустить в dev режиме:** `wails dev`
4. **Начать разработку:** изменять файлы в `internal/app/` и `frontend/`
5. **Тестировать:** `go test ./...` для backend, E2E для frontend
6. **Собрать:** `wails build -upx`

### Полезные ресурсы

- 📚 **Документация Wails:** https://wails.io/docs/
- 💬 **Discord сообщество:** https://discord.gg/wails
- 🐛 **GitHub Issues:** https://github.com/wailsapp/wails/issues
- 📖 **Примеры Wails приложений:** https://github.com/wailsapp/awesome-wails

### Контрибьютинг

Вопросы, баги и предложения:
- 🐛 **GitHub Issues:** https://github.com/bivlked/CurRate-Go/issues
- 💡 **Discussions:** https://github.com/bivlked/CurRate-Go/discussions

---

**Спасибо за разработку CurRate-Go!**

*Документ подготовлен: 2025-12-22*
*Версия: 1.0*
*Статус: Реализовано/Актуально*
