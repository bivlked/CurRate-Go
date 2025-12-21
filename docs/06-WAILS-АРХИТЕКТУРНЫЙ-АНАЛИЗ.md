# АРХИТЕКТУРНЫЙ АНАЛИЗ WAILS
## Проект: CurRate Go Rewrite

**Версия:** 1.0
**Дата:** 20 декабря 2025
**Wails версия:** v2.11.0 (стабильная)

---

## 1. ОБЗОР WAILS

### 1.1. Что такое Wails?

**Wails** - современный фреймворк для создания desktop приложений с использованием:
- **Backend**: Go (бизнес-логика, производительность)
- **Frontend**: HTML/CSS/JavaScript (или React/Vue/Svelte)
- **WebView**: Нативный WebView OS (WebView2 на Windows)

### 1.2. Преимущества vs Walk

| Критерий | Walk | Wails v2.11.0 |
|----------|------|---------------|
| **Активность** | ❌ Последний коммит ~2021 | ✅ Активная разработка (релиз 08.11.2025) |
| **Документация** | ⚠️ Минимальная | ✅ Обширная (Context7: 1918 примеров) |
| **UI технология** | Win32 нативные контролы | HTML/CSS/JS (современно) |
| **Размер exe** | ✅ 2-5 MB | ⚠️ 10-20 MB |
| **Гибкость UI** | ⚠️ Ограничена виджетами | ✅ Полная свобода веб-технологий |
| **Community** | ⚠️ Маленькое | ✅ Большое (7k stars) |
| **Cross-platform** | ❌ Только Windows | ✅ Windows/macOS/Linux |

**Вывод**: Wails - современнее, надежнее, с лучшей поддержкой. Недостаток только в размере exe.

---

## 2. АРХИТЕКТУРА WAILS ПРИЛОЖЕНИЯ

### 2.1. Общая схема

```
┌─────────────────────────────────────────────┐
│         FRONTEND (WebView)                  │
│    HTML/CSS/JavaScript                      │
│    - UI компоненты                          │
│    - Event handlers                         │
│    - Вызовы Go методов через JS bindings   │
└──────────────┬──────────────────────────────┘
               │ JS → Go Bridge
               │ (автоматически генерируется)
               ▼
┌─────────────────────────────────────────────┐
│         GO BACKEND                          │
│    - App struct с методами                  │
│    - Бизнес-логика (Converter)              │
│    - Работа с API, файлами, БД             │
└──────────────┬──────────────────────────────┘
               │
               ▼
     ┌─────────────────┐
     │  Native WebView  │
     │  (WebView2 Win)  │
     └─────────────────┘
```

### 2.2. Binding механизм

**Go структура:**
```go
type App struct {
    ctx       context.Context
    converter *converter.Converter
}

// Этот метод будет доступен из JavaScript
func (a *App) Convert(amount float64, currency string, date string) (*models.ConversionResult, error) {
    // Бизнес-логика
}
```

**JavaScript (авто-генерируется):**
```javascript
import { Convert } from "../wailsjs/go/main/App";

// Вызов Go метода как Promise
Convert(1000, "USD", "2025-12-20")
    .then(result => {
        console.log(result);
    })
    .catch(error => {
        console.error(error);
    });
```

---

## 3. ВАРИАНТЫ АРХИТЕКТУРЫ FRONTEND

Рассмотрим 3 варианта frontend решений.

### 📊 Вариант 1: Vanilla JavaScript (HTML/CSS/JS)

**Структура:**
```
frontend/
├── index.html      (~100 строк)
├── main.css        (~200 строк)
└── main.js         (~150 строк)
```

**Пример кода:**
```html
<!-- index.html -->
<!DOCTYPE html>
<html>
<head>
    <link rel="stylesheet" href="main.css">
</head>
<body>
    <div class="container">
        <h1>Конвертер валют</h1>

        <label>Дата:</label>
        <input type="date" id="dateInput">

        <label>Валюта:</label>
        <select id="currencySelect">
            <option value="USD">Доллар США ($)</option>
            <option value="EUR">Евро (€)</option>
        </select>

        <label>Сумма:</label>
        <input type="number" id="amountInput">

        <button id="convertBtn">Конвертировать</button>
        <button id="copyBtn">Копировать</button>

        <div id="result"></div>
    </div>
    <script src="main.js" type="module"></script>
</body>
</html>
```

```javascript
// main.js
import { Convert } from "../wailsjs/go/main/App";

document.getElementById('convertBtn').addEventListener('click', async () => {
    const date = document.getElementById('dateInput').value;
    const currency = document.getElementById('currencySelect').value;
    const amount = parseFloat(document.getElementById('amountInput').value);

    try {
        const result = await Convert(amount, currency, date);
        document.getElementById('result').textContent = result.FormattedStr;
    } catch (error) {
        alert('Ошибка: ' + error);
    }
});
```

**Преимущества:**
- ✅ Простота - нет фреймворков, нет сборщиков
- ✅ Маленький размер - минимум зависимостей
- ✅ Быстрая разработка для простого UI
- ✅ Легко понять и поддерживать
- ✅ Не нужен Node.js для разработки (опционально)

**Недостатки:**
- ❌ Ручное управление DOM
- ❌ Нет реактивности
- ❌ Сложнее масштабировать при росте проекта

**Размер итогового exe:** ~12-15 MB

**Оценка:** 8/10 (**Рекомендуется для нашего проекта**)

---

### 📊 Вариант 2: Svelte (современный легкий фреймворк)

**Структура:**
```
frontend/
├── src/
│   ├── App.svelte           (~150 строк)
│   ├── components/
│   │   ├── DatePicker.svelte
│   │   ├── CurrencySelect.svelte
│   │   └── ResultDisplay.svelte
│   └── main.js
├── package.json
└── vite.config.js
```

**Пример кода:**
```svelte
<!-- App.svelte -->
<script>
  import { Convert } from "../wailsjs/go/main/App";

  let date = new Date().toISOString().split('T')[0];
  let currency = "USD";
  let amount = 0;
  let result = "";

  async function handleConvert() {
    try {
      const res = await Convert(amount, currency, date);
      result = res.FormattedStr;
    } catch (error) {
      alert('Ошибка: ' + error);
    }
  }
</script>

<div class="container">
  <h1>Конвертер валют</h1>

  <input type="date" bind:value={date}>

  <select bind:value={currency}>
    <option value="USD">Доллар США ($)</option>
    <option value="EUR">Евро (€)</option>
  </select>

  <input type="number" bind:value={amount}>

  <button on:click={handleConvert}>Конвертировать</button>

  {#if result}
    <div class="result">{result}</div>
  {/if}
</div>
```

**Преимущества:**
- ✅ Реактивность из коробки
- ✅ Компилируется в ванильный JS (меньше runtime)
- ✅ Легче чем React/Vue
- ✅ Хорошая производительность

**Недостатки:**
- ❌ Нужен Node.js и npm
- ❌ Build процесс (Vite)
- ❌ Больше зависимостей
- ❌ Избыточно для простого UI

**Размер итогового exe:** ~15-18 MB

**Оценка:** 6/10 (хорошо, но избыточно для нашего случая)

---

### 📊 Вариант 3: React + TypeScript (максимальная мощь)

**Структура:**
```
frontend/
├── src/
│   ├── App.tsx
│   ├── components/
│   │   ├── DatePicker.tsx
│   │   ├── CurrencySelect.tsx
│   │   ├── AmountInput.tsx
│   │   └── ResultDisplay.tsx
│   ├── hooks/
│   │   └── useConverter.ts
│   └── main.tsx
├── package.json
└── tsconfig.json
```

**Преимущества:**
- ✅ Type safety (TypeScript)
- ✅ Огромная экосистема
- ✅ Легко найти разработчиков
- ✅ Компонентная архитектура

**Недостатки:**
- ❌ Самый тяжелый вариант
- ❌ Больше всего зависимостей
- ❌ Сложнее setup
- ❌ **ИЗБЫТОЧНО для простого приложения с 5 полями**

**Размер итогового exe:** ~18-25 MB

**Оценка:** 4/10 (over-engineering для нашего случая)

---

## 4. СРАВНИТЕЛЬНАЯ ТАБЛИЦА

| Критерий | Vanilla JS | Svelte | React+TS |
|----------|-----------|--------|----------|
| **Простота** | ⭐⭐⭐⭐⭐ | ⭐⭐⭐ | ⭐⭐ |
| **Размер exe** | ⭐⭐⭐⭐⭐ (12-15 MB) | ⭐⭐⭐⭐ (15-18 MB) | ⭐⭐ (18-25 MB) |
| **Скорость разработки** | ⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐ |
| **Масштабируемость** | ⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ |
| **Поддержка** | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ |
| **Нужен Node.js** | ❌ (опционально) | ✅ | ✅ |
| **Build сложность** | ⭐⭐⭐⭐⭐ (минимум) | ⭐⭐⭐ | ⭐⭐ |

---

## 5. ВЫБРАННОЕ РЕШЕНИЕ

### ✅ **Вариант 1: Vanilla JavaScript**

**Обоснование:**

1. **Простота задачи**: Наше приложение имеет всего 5 UI элементов:
   - DateEdit (дата)
   - RadioButton группа (валюта)
   - LineEdit (сумма)
   - 2 кнопки (конвертировать, копировать)
   - Label (результат)

2. **Критерии проекта**:
   - ✅ Размер exe <= 10 MB (с Vanilla JS: 12-15 MB - приемлемо)
   - ✅ Standalone exe без установки
   - ✅ Простота поддержки

3. **Нет необходимости в фреймворке**:
   - Нет сложного state management
   - Нет динамических списков
   - Нет роутинга
   - Минимум интерактивности

4. **Следование плану**:
   - План предполагал простой GUI
   - Vanilla JS - самый простой вариант

---

## 6. ДЕТАЛЬНАЯ АРХИТЕКТУРА РЕШЕНИЯ

### 6.1. Файловая структура

```
CurRate-Go-Rewrite/
│
├── cmd/
│   └── currate/
│       └── main.go                    # Точка входа Wails приложения
│
├── internal/
│   ├── app/
│   │   └── app.go                     # App struct с методами для frontend
│   ├── converter/                     # (уже реализовано на Этапе 5)
│   ├── parser/                        # (уже реализовано на Этапе 3)
│   ├── cache/                         # (уже реализовано на Этапе 4)
│   └── models/                        # (уже реализовано на Этапе 2)
│
├── frontend/
│   ├── dist/                          # Build output (генерируется)
│   ├── index.html                     # Главный HTML файл
│   ├── main.css                       # Стили
│   └── main.js                        # JavaScript логика
│
├── build/
│   └── windows/                       # Windows ресурсы (иконки и т.д.)
│
├── wails.json                         # Конфигурация Wails проекта
└── go.mod                             # Go зависимости
```

### 6.2. Архитектурные слои

```
┌─────────────────────────────────────────────┐
│         PRESENTATION LAYER (Frontend)       │
│    index.html + main.css + main.js          │
│    - UI рендеринг                           │
│    - Валидация ввода (базовая)             │
│    - Event handling                         │
└──────────────┬──────────────────────────────┘
               │
               │ JavaScript → Go Bindings
               │
┌──────────────▼──────────────────────────────┐
│         APPLICATION LAYER (App)             │
│    internal/app/app.go                      │
│    - Методы для frontend                    │
│    - Парсинг и валидация параметров        │
│    - Обработка ошибок                       │
└──────────────┬──────────────────────────────┘
               │
               │ Dependency Injection
               │
┌──────────────▼──────────────────────────────┐
│         BUSINESS LOGIC LAYER                │
│    internal/converter/converter.go          │
│    - Конвертация валют                      │
│    - Форматирование результата              │
└────────┬─────────────────┬──────────────────┘
         │                 │
         ▼                 ▼
   ┌─────────┐       ┌──────────┐
   │  CACHE  │       │  PARSER  │
   │  LAYER  │       │  LAYER   │
   └─────────┘       └────┬─────┘
                          │
                          ▼
                     ┌─────────┐
                     │   CBR   │
                     │   API   │
                     └─────────┘
```

### 6.3. Ключевые файлы

#### cmd/currate/main.go
```go
package main

import (
    "embed"
    "log"

    "github.com/wailsapp/wails/v2"
    "github.com/wailsapp/wails/v2/pkg/options"
    "github.com/wailsapp/wails/v2/pkg/options/assetserver"
    "github.com/wailsapp/wails/v2/pkg/options/windows"

    "github.com/bivlked/currate-go/internal/app"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
    // Инициализация компонентов
    application := app.NewApp()

    // Запуск Wails приложения
    err := wails.Run(&options.App{
        Title:  "Конвертер валют (с) BiV 2024 г.",
        Width:  340,
        Height: 455,
        AssetServer: &assetserver.Options{
            Assets: assets,
        },
        BackgroundColour: &options.RGBA{R: 255, G: 255, B: 255, A: 255},
        OnStartup:        application.Startup,
        Bind: []interface{}{
            application,
        },
        Windows: &windows.Options{
            WebviewIsTransparent: false,
            WindowIsTranslucent:  false,
            DisableWindowIcon:    false,
        },
    })

    if err != nil {
        log.Fatal(err)
    }
}
```

#### internal/app/app.go
```go
package app

import (
    "context"
    "fmt"
    "time"

    "github.com/bivlked/currate-go/internal/cache"
    "github.com/bivlked/currate-go/internal/converter"
    "github.com/bivlked/currate-go/internal/models"
    "github.com/bivlked/currate-go/internal/parser"
)

// App структура для Wails приложения
type App struct {
    ctx       context.Context
    converter *converter.Converter
}

// NewApp создает новый App
func NewApp() *App {
    // Инициализация компонентов
    cacheInstance := cache.NewLRUCache(100, 24*time.Hour)
    httpClient := parser.NewHTTPClient(10 * time.Second)
    cbrParser := parser.NewCBRParser(httpClient)
    currencyConverter := converter.NewConverter(cbrParser, cacheInstance)

    return &App{
        converter: currencyConverter,
    }
}

// Startup вызывается при старте приложения
func (a *App) Startup(ctx context.Context) {
    a.ctx = ctx
}

// ConvertRequest структура запроса на конвертацию
type ConvertRequest struct {
    Amount   float64 `json:"amount"`
    Currency string  `json:"currency"`
    Date     string  `json:"date"` // "2025-12-20"
}

// Convert метод для вызова из JavaScript
func (a *App) Convert(req ConvertRequest) (*models.ConversionResult, error) {
    // Парсинг даты
    date, err := time.Parse("2006-01-02", req.Date)
    if err != nil {
        return nil, fmt.Errorf("некорректный формат даты: %w", err)
    }

    // Парсинг валюты
    currency := models.Currency(req.Currency)

    // Вызов конвертера
    result, err := a.converter.Convert(req.Amount, currency, date)
    if err != nil {
        return nil, err
    }

    return result, nil
}

// GetTodayDate возвращает сегодняшнюю дату в формате YYYY-MM-DD
func (a *App) GetTodayDate() string {
    return time.Now().Format("2006-01-02")
}
```

#### frontend/index.html
```html
<!DOCTYPE html>
<html lang="ru">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Конвертер валют</title>
    <link rel="stylesheet" href="main.css">
</head>
<body>
    <div class="container">
        <h1>Конвертер валют</h1>

        <div class="form-group">
            <label for="date">Дата курса:</label>
            <input type="date" id="date" required>
        </div>

        <div class="form-group">
            <label>Валюта:</label>
            <div class="radio-group">
                <label>
                    <input type="radio" name="currency" value="USD" checked>
                    Доллар США ($)
                </label>
                <label>
                    <input type="radio" name="currency" value="EUR">
                    Евро (€)
                </label>
            </div>
        </div>

        <div class="form-group">
            <label for="amount">Сумма:</label>
            <input type="number" id="amount" min="0" step="0.01" required>
        </div>

        <div class="button-group">
            <button id="convertBtn" class="btn btn-primary">Конвертировать</button>
            <button id="copyBtn" class="btn btn-secondary">Копировать</button>
        </div>

        <div id="result" class="result"></div>
        <div id="error" class="error"></div>
    </div>

    <script src="main.js" type="module"></script>
</body>
</html>
```

#### frontend/main.js
```javascript
import { Convert, GetTodayDate } from "../wailsjs/go/app/App";

// Инициализация
let lastResult = "";

document.addEventListener('DOMContentLoaded', async () => {
    // Установить сегодняшнюю дату
    const today = await GetTodayDate();
    document.getElementById('date').value = today;

    // Обработчики событий
    document.getElementById('convertBtn').addEventListener('click', handleConvert);
    document.getElementById('copyBtn').addEventListener('click', handleCopy);

    // Enter для конвертации
    document.getElementById('amount').addEventListener('keypress', (e) => {
        if (e.key === 'Enter') handleConvert();
    });
});

async function handleConvert() {
    const date = document.getElementById('date').value;
    const currency = document.querySelector('input[name="currency"]:checked').value;
    const amount = parseFloat(document.getElementById('amount').value);

    // Очистка предыдущих сообщений
    document.getElementById('result').textContent = '';
    document.getElementById('error').textContent = '';

    // Базовая валидация
    if (!date || !amount || amount <= 0) {
        showError('Пожалуйста, заполните все поля корректно');
        return;
    }

    try {
        // Вызов Go метода
        const result = await Convert({
            amount: amount,
            currency: currency,
            date: date
        });

        // Отображение результата
        lastResult = result.FormattedStr;
        document.getElementById('result').textContent = lastResult;

    } catch (error) {
        showError(error);
    }
}

function handleCopy() {
    if (!lastResult) {
        showError('Нет результата для копирования');
        return;
    }

    navigator.clipboard.writeText(lastResult)
        .then(() => {
            alert('Результат скопирован в буфер обмена');
        })
        .catch((err) => {
            showError('Не удалось скопировать: ' + err);
        });
}

function showError(message) {
    document.getElementById('error').textContent = message;
    setTimeout(() => {
        document.getElementById('error').textContent = '';
    }, 5000);
}
```

#### frontend/main.css
```css
* {
    margin: 0;
    padding: 0;
    box-sizing: border-box;
}

body {
    font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
    background: #f5f5f5;
}

.container {
    max-width: 340px;
    margin: 0 auto;
    padding: 20px;
    background: white;
    min-height: 100vh;
}

h1 {
    font-size: 20px;
    margin-bottom: 20px;
    color: #333;
    text-align: center;
}

.form-group {
    margin-bottom: 15px;
}

label {
    display: block;
    margin-bottom: 5px;
    font-weight: 500;
    color: #555;
}

input[type="date"],
input[type="number"] {
    width: 100%;
    padding: 8px;
    border: 1px solid #ddd;
    border-radius: 4px;
    font-size: 14px;
}

.radio-group {
    display: flex;
    flex-direction: column;
    gap: 8px;
}

.radio-group label {
    display: flex;
    align-items: center;
    font-weight: normal;
}

.radio-group input[type="radio"] {
    margin-right: 8px;
}

.button-group {
    display: flex;
    gap: 10px;
    margin-top: 20px;
}

.btn {
    flex: 1;
    padding: 10px;
    border: none;
    border-radius: 4px;
    font-size: 14px;
    cursor: pointer;
    transition: background 0.3s;
}

.btn-primary {
    background: #007bff;
    color: white;
}

.btn-primary:hover {
    background: #0056b3;
}

.btn-secondary {
    background: #6c757d;
    color: white;
}

.btn-secondary:hover {
    background: #545b62;
}

.result {
    margin-top: 20px;
    padding: 15px;
    background: #d4edda;
    border: 1px solid #c3e6cb;
    border-radius: 4px;
    color: #155724;
    font-weight: bold;
    text-align: center;
    min-height: 50px;
}

.error {
    margin-top: 10px;
    padding: 10px;
    background: #f8d7da;
    border: 1px solid #f5c6cb;
    border-radius: 4px;
    color: #721c24;
    font-size: 13px;
}
```

---

## 7. ТЕХНОЛОГИЧЕСКИЕ РЕШЕНИЯ

### 7.1. Асинхронность

**Проблема**: HTTP запрос к ЦБ РФ может занять 1-3 секунды

**Решение**: Go обрабатывает запросы асинхронно автоматически
- JavaScript вызывает Go метод через Promise
- Go выполняет HTTP запрос в отдельной горутине (автоматически)
- UI остается отзывчивым

**Код:**
```javascript
// Автоматически асинхронный вызов
const result = await Convert({...}); // UI не блокируется
```

### 7.2. Обработка ошибок

**Frontend → Go:**
```javascript
try {
    const result = await Convert({...});
    // Успех
} catch (error) {
    // Go вернул ошибку
    showError(error);
}
```

**Go обработка:**
```go
func (a *App) Convert(req ConvertRequest) (*models.ConversionResult, error) {
    // Валидация
    if req.Amount <= 0 {
        return nil, errors.New("сумма должна быть больше 0")
    }

    // Бизнес-логика
    result, err := a.converter.Convert(...)
    if err != nil {
        return nil, fmt.Errorf("ошибка конвертации: %w", err)
    }

    return result, nil
}
```

### 7.3. Копирование в буфер обмена

**Решение**: Используем Web Clipboard API
```javascript
navigator.clipboard.writeText(lastResult)
    .then(() => alert('Скопировано'))
    .catch(err => showError(err));
```

**Преимущества:**
- ✅ Нативная поддержка в WebView2
- ✅ Не нужны дополнительные библиотеки
- ✅ Работает во всех современных браузерах

---

## 8. ИНТЕГРАЦИЯ С СУЩЕСТВУЮЩИМ КОДОМ

### 8.1. Уже реализованные модули

✅ **Этап 2**: `internal/models` - модели данных
✅ **Этап 3**: `internal/parser` - парсинг ЦБ РФ
✅ **Этап 4**: `internal/cache` - LRU кэш
✅ **Этап 5**: `internal/converter` - бизнес-логика

### 8.2. Новые модули для Wails

🆕 **Этап 6**:
- `cmd/currate/main.go` - точка входа Wails
- `internal/app/app.go` - адаптер для frontend
- `frontend/` - UI слой

### 8.3. Схема интеграции

```
Frontend (JS)
    │
    └─> App.Convert(req)          [internal/app/app.go]
            │
            └─> converter.Convert() [internal/converter/converter.go]
                    │
                    ├─> cache.Get()     [internal/cache/lru.go]
                    │
                    └─> parser.GetRate() [internal/parser/cbr.go]
                            │
                            └─> HTTP → cbr.ru
```

**Ключевое преимущество**:
- ✅ **НЕ нужно переписывать** существующий код
- ✅ Создаем только тонкий слой адаптера (`app.go`)
- ✅ Вся бизнес-логика остается без изменений

---

## 9. СБОРКА И DEPLOYMENT

### 9.1. Команды

**Разработка:**
```bash
wails dev
# Запускает приложение с hot-reload
```

**Production build:**
```bash
wails build
# Создает оптимизированный exe в build/bin/
```

**С дополнительными флагами:**
```bash
wails build -clean -upx -ldflags "-s -w"
# -clean: очистить build директорию
# -upx: сжать exe через UPX (~30-40% меньше)
# -ldflags: оптимизация Go бинарника
```

### 9.2. Ожидаемый размер

| Вариант | Размер |
|---------|--------|
| Без оптимизаций | ~15-18 MB |
| С -ldflags "-s -w" | ~12-15 MB |
| С UPX компрессией | ~8-10 MB |

**Цель плана**: <= 10 MB ✅ (достижимо с UPX)

---

## 10. ПРЕИМУЩЕСТВА ВЫБРАННОГО РЕШЕНИЯ

### 10.1. Технические

1. ✅ **Современная технология** (Wails v2.11.0)
2. ✅ **Активная поддержка** (релиз 08.11.2025)
3. ✅ **Отличная документация** (1918 примеров)
4. ✅ **Простота UI** (Vanilla JS для 5 полей)
5. ✅ **Минимальные зависимости**
6. ✅ **Асинхронность из коробки**
7. ✅ **Интеграция с существующим кодом**

### 10.2. Бизнес

1. ✅ **Быстрая разработка** (~1-2 дня на UI)
2. ✅ **Легкая поддержка** (простой код)
3. ✅ **Размер <= 10 MB** (с UPX)
4. ✅ **Cross-platform потенциал** (macOS, Linux в будущем)

---

## 11. РИСКИ И МИТИГАЦИЯ

| Риск | Вероятность | Влияние | Митигация |
|------|-------------|---------|-----------|
| WebView2 не установлен на Windows | Средняя | Высокое | Wails автоматически проверяет и предлагает установку |
| Размер exe > 10 MB | Низкая | Среднее | UPX компрессия (~8-10 MB) |
| Проблемы с CORS в WebView | Очень низкая | Низкое | Wails управляет WebView, нет CORS |
| Сложности с clipboard API | Очень низкая | Низкое | Нативная поддержка в WebView2 |

---

## 12. ИТОГОВОЕ РЕШЕНИЕ

### ✅ **Выбрано: Wails v2.11.0 + Vanilla JavaScript**

**Обоснование:**
1. Современная альтернатива Walk
2. Отличная документация и поддержка
3. Простой UI не требует фреймворков
4. Интеграция с существующим кодом без переписывания
5. Достижим размер <= 10 MB
6. Потенциал cross-platform в будущем

**Следующие шаги:**
1. Обновить техническую документацию (стек, архитектура, план)
2. Установить Wails CLI
3. Создать Proof of Concept
4. Реализовать интеграцию

---

**Конец документа**

**Автор:** Claude (Anthropic)
**Дата:** 20.12.2025
**Версия:** 1.0
