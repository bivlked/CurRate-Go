# ДЕТАЛЬНЫЙ ПЛАН РАЗРАБОТКИ GUI НА WAILS V2
## Проект: CurRate Go Rewrite

**Версия:** 1.0
**Дата:** 22 декабря 2025
**Основа:** Wails v2.11.0 + Vanilla JavaScript
**Автор:** Claude (Anthropic)

---

## 1. ОБЗОР ТРЕБОВАНИЙ

### 1.1. Функциональные элементы интерфейса

На основе анализа Python-версии (tkinter) необходимо реализовать:

✅ **Обязательные компоненты:**
1. **Поле ввода даты (текстовое)** - для ручного ввода в формате DD.MM.YYYY
2. **Календарь (визуальный)** - для выбора даты кликом
3. **Радиокнопки выбора валюты** - USD / EUR
4. **Поле ввода суммы** - числовой ввод
5. **Кнопка "Конвертировать"** - запуск конвертации
6. **Кнопка "Копировать в буфер"** - копирование результата
7. **Метка результата** - отображение отформатированного результата

✨ **Особые требования:**
- **Выходные дни (Сб/Вс) визуально выделены** в календаре (красный цвет)
- Асинхронная обработка (не блокирует UI)
- Блокировка кнопок во время запроса
- Активация кнопки копирования только после успешного результата

### 1.2. Размеры окна

- **Минимальный размер:** 340x455 (как в Python-версии)
- **Стартовый размер:** 400x600 (удобнее для календаря)
- **Изменяемость:** Да (пользователь может расширить)

---

## 2. АРХИТЕКТУРА РЕШЕНИЯ

### 2.1. Выбранный стек

```
┌─────────────────────────────────────────┐
│  FRONTEND: Vanilla JavaScript           │
│  - HTML5 (структура)                    │
│  - CSS3 (стили + календарь)             │
│  - JavaScript ES6+ (логика)             │
└──────────────┬──────────────────────────┘
               │
               │ Wails Runtime
               │ (автоматические bindings)
               ▼
┌─────────────────────────────────────────┐
│  BACKEND: Go                            │
│  - internal/app/app.go (адаптер)        │
│  - internal/converter/* (логика)        │
│  - internal/parser/* (API ЦБ РФ)        │
│  - internal/cache/* (кэширование)       │
└─────────────────────────────────────────┘
```

**Обоснование Vanilla JS:**
- Простота задачи (7 элементов UI)
- Нет сложного state management
- Минимальный размер exe (~12-15 MB)
- Легко поддерживать и расширять

### 2.2. Структура проекта

```
CurRate-Go-Rewrite/
│
├── cmd/
│   └── currate/
│       └── main.go                  # Точка входа Wails
│
├── internal/
│   ├── app/
│   │   ├── app.go                   # App struct с методами для frontend
│   │   └── app_test.go              # Тесты
│   │
│   ├── converter/                   # ✅ Уже реализовано (Этап 5)
│   ├── parser/                      # ✅ Уже реализовано (Этап 3)
│   ├── cache/                       # ✅ Уже реализовано (Этап 4)
│   └── models/                      # ✅ Уже реализовано (Этап 2)
│
├── frontend/
│   ├── src/
│   │   ├── index.html               # Главная разметка
│   │   ├── styles/
│   │   │   ├── main.css             # Основные стили
│   │   │   └── calendar.css         # Стили календаря
│   │   ├── scripts/
│   │   │   ├── main.js              # Главная логика
│   │   │   ├── calendar.js          # Логика календаря
│   │   │   └── utils.js             # Вспомогательные функции
│   │   └── assets/
│   │       └── icon.png             # Иконка приложения
│   └── dist/                        # Build output (генерируется)
│
├── build/
│   ├── windows/
│   │   └── icon.ico                 # Windows иконка
│   └── appicon.png                  # Общая иконка
│
├── wails.json                       # Конфигурация Wails
└── go.mod
```

---

## 3. ДЕТАЛЬНЫЙ ДИЗАЙН КАЛЕНДАРЯ

### 3.1. Требования к календарю

1. **Отображение месяца** - текущий месяц + навигация
2. **Выделение выходных** - Суббота и Воскресенье красным цветом
3. **Выделение выбранной даты** - визуально отличается
4. **Выделение сегодняшней даты** - рамка или фон
5. **Синхронизация с полем ввода** - двусторонняя
6. **Навигация** - кнопки "предыдущий/следующий месяц"

### 3.2. HTML структура календаря

```html
<div class="calendar-container">
    <!-- Заголовок календаря -->
    <div class="calendar-header">
        <button id="prevMonth" class="nav-btn">&lt;</button>
        <div id="currentMonth" class="month-title">Декабрь 2025</div>
        <button id="nextMonth" class="nav-btn">&gt;</button>
    </div>

    <!-- Названия дней недели -->
    <div class="calendar-weekdays">
        <div class="weekday">Пн</div>
        <div class="weekday">Вт</div>
        <div class="weekday">Ср</div>
        <div class="weekday">Чт</div>
        <div class="weekday">Пт</div>
        <div class="weekday weekend">Сб</div>
        <div class="weekday weekend">Вс</div>
    </div>

    <!-- Сетка дней -->
    <div id="calendarDays" class="calendar-days">
        <!-- Генерируется JavaScript -->
    </div>
</div>
```

### 3.3. CSS стили календаря

```css
/* Контейнер календаря */
.calendar-container {
    background: #fff;
    border: 1px solid #ddd;
    border-radius: 8px;
    padding: 15px;
    margin: 15px 0;
}

/* Заголовок с навигацией */
.calendar-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 15px;
}

.month-title {
    font-size: 16px;
    font-weight: 600;
    color: #333;
}

.nav-btn {
    background: #f0f0f0;
    border: 1px solid #ddd;
    border-radius: 4px;
    width: 32px;
    height: 32px;
    cursor: pointer;
    font-size: 18px;
    transition: background 0.2s;
}

.nav-btn:hover {
    background: #e0e0e0;
}

/* Названия дней недели */
.calendar-weekdays {
    display: grid;
    grid-template-columns: repeat(7, 1fr);
    gap: 5px;
    margin-bottom: 10px;
}

.weekday {
    text-align: center;
    font-size: 12px;
    font-weight: 600;
    color: #666;
    padding: 5px;
}

.weekday.weekend {
    color: #d32f2f; /* Красный для Сб/Вс */
}

/* Сетка дней */
.calendar-days {
    display: grid;
    grid-template-columns: repeat(7, 1fr);
    gap: 5px;
}

.calendar-day {
    aspect-ratio: 1;
    display: flex;
    align-items: center;
    justify-content: center;
    border-radius: 4px;
    cursor: pointer;
    font-size: 14px;
    transition: all 0.2s;
    border: 2px solid transparent;
}

/* Обычный день */
.calendar-day.active {
    background: #f5f5f5;
    color: #333;
}

.calendar-day.active:hover {
    background: #e3f2fd;
    border-color: #2196f3;
}

/* Выходные дни (Сб/Вс) */
.calendar-day.weekend {
    color: #d32f2f; /* Красный текст */
    background: #ffebee; /* Светло-красный фон */
}

.calendar-day.weekend:hover {
    background: #ffcdd2;
    border-color: #d32f2f;
}

/* Сегодняшняя дата */
.calendar-day.today {
    border: 2px solid #4caf50;
    font-weight: bold;
}

/* Выбранная дата */
.calendar-day.selected {
    background: #2196f3;
    color: white;
    font-weight: bold;
}

.calendar-day.selected:hover {
    background: #1976d2;
}

/* Неактивные дни (из других месяцев) */
.calendar-day.inactive {
    color: #ccc;
    cursor: default;
    background: transparent;
}

.calendar-day.inactive:hover {
    background: transparent;
    border-color: transparent;
}
```

### 3.4. JavaScript логика календаря

```javascript
// calendar.js

export class Calendar {
    constructor(containerId, onDateSelect) {
        this.container = document.getElementById(containerId);
        this.onDateSelect = onDateSelect;
        this.currentDate = new Date();
        this.selectedDate = new Date();

        this.init();
    }

    init() {
        this.render();
        this.attachEventListeners();
    }

    render() {
        const year = this.currentDate.getFullYear();
        const month = this.currentDate.getMonth();

        // Обновить заголовок
        this.updateHeader(year, month);

        // Отрендерить дни
        this.renderDays(year, month);
    }

    updateHeader(year, month) {
        const monthNames = [
            'Январь', 'Февраль', 'Март', 'Апрель', 'Май', 'Июнь',
            'Июль', 'Август', 'Сентябрь', 'Октябрь', 'Ноябрь', 'Декабрь'
        ];

        const title = document.getElementById('currentMonth');
        title.textContent = `${monthNames[month]} ${year}`;
    }

    renderDays(year, month) {
        const daysContainer = document.getElementById('calendarDays');
        daysContainer.innerHTML = '';

        // Первый день месяца
        const firstDay = new Date(year, month, 1);
        // Последний день месяца
        const lastDay = new Date(year, month + 1, 0);

        // Начинаем с понедельника (1 = Пн, 0 = Вс)
        let startDay = firstDay.getDay();
        startDay = startDay === 0 ? 6 : startDay - 1;

        // Дни предыдущего месяца
        const prevMonthLastDay = new Date(year, month, 0).getDate();
        for (let i = startDay - 1; i >= 0; i--) {
            const day = prevMonthLastDay - i;
            this.createDayElement(day, false, -1);
        }

        // Дни текущего месяца
        for (let day = 1; day <= lastDay.getDate(); day++) {
            const date = new Date(year, month, day);
            const dayOfWeek = date.getDay();
            const isWeekend = (dayOfWeek === 0 || dayOfWeek === 6);
            const isToday = this.isSameDay(date, new Date());
            const isSelected = this.isSameDay(date, this.selectedDate);

            this.createDayElement(day, true, 0, isWeekend, isToday, isSelected, date);
        }

        // Дни следующего месяца
        const remainingCells = 42 - (startDay + lastDay.getDate());
        for (let day = 1; day <= remainingCells; day++) {
            this.createDayElement(day, false, 1);
        }
    }

    createDayElement(day, isActive, monthOffset, isWeekend = false, isToday = false, isSelected = false, date = null) {
        const dayEl = document.createElement('div');
        dayEl.className = 'calendar-day';
        dayEl.textContent = day;

        if (!isActive) {
            dayEl.classList.add('inactive');
        } else {
            dayEl.classList.add('active');

            if (isWeekend) {
                dayEl.classList.add('weekend');
            }

            if (isToday) {
                dayEl.classList.add('today');
            }

            if (isSelected) {
                dayEl.classList.add('selected');
            }

            // Обработчик клика
            dayEl.addEventListener('click', () => {
                if (date) {
                    this.selectDate(date);
                }
            });
        }

        document.getElementById('calendarDays').appendChild(dayEl);
    }

    selectDate(date) {
        this.selectedDate = date;
        this.render();

        // Callback для обновления поля ввода
        if (this.onDateSelect) {
            this.onDateSelect(date);
        }
    }

    setDate(date) {
        this.selectedDate = date;
        this.currentDate = new Date(date);
        this.render();
    }

    prevMonth() {
        this.currentDate.setMonth(this.currentDate.getMonth() - 1);
        this.render();
    }

    nextMonth() {
        this.currentDate.setMonth(this.currentDate.getMonth() + 1);
        this.render();
    }

    attachEventListeners() {
        document.getElementById('prevMonth').addEventListener('click', () => {
            this.prevMonth();
        });

        document.getElementById('nextMonth').addEventListener('click', () => {
            this.nextMonth();
        });
    }

    isSameDay(date1, date2) {
        return date1.getFullYear() === date2.getFullYear() &&
               date1.getMonth() === date2.getMonth() &&
               date1.getDate() === date2.getDate();
    }
}
```

---

## 4. ПОЛНАЯ СТРУКТУРА HTML

```html
<!DOCTYPE html>
<html lang="ru">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Конвертер валют</title>
    <link rel="stylesheet" href="styles/main.css">
    <link rel="stylesheet" href="styles/calendar.css">
</head>
<body>
    <div class="container">
        <!-- Заголовок -->
        <h1>Конвертер валют</h1>
        <p class="subtitle">(с) BiV 2025 г.</p>

        <!-- Блок даты -->
        <div class="form-section">
            <label for="dateInput" class="form-label">Дата курса:</label>
            <input
                type="text"
                id="dateInput"
                placeholder="ДД.ММ.ГГГГ"
                maxlength="10"
                class="form-input"
            >
        </div>

        <!-- Календарь -->
        <div class="calendar-container">
            <div class="calendar-header">
                <button id="prevMonth" class="nav-btn">&lt;</button>
                <div id="currentMonth" class="month-title"></div>
                <button id="nextMonth" class="nav-btn">&gt;</button>
            </div>

            <div class="calendar-weekdays">
                <div class="weekday">Пн</div>
                <div class="weekday">Вт</div>
                <div class="weekday">Ср</div>
                <div class="weekday">Чт</div>
                <div class="weekday">Пт</div>
                <div class="weekday weekend">Сб</div>
                <div class="weekday weekend">Вс</div>
            </div>

            <div id="calendarDays" class="calendar-days"></div>
        </div>

        <!-- Выбор валюты -->
        <div class="form-section">
            <label class="form-label">Выберите валюту:</label>
            <div class="radio-group">
                <label class="radio-label">
                    <input type="radio" name="currency" value="USD" checked>
                    <span>Доллар США (USD)</span>
                </label>
                <label class="radio-label">
                    <input type="radio" name="currency" value="EUR">
                    <span>Евро (EUR)</span>
                </label>
            </div>
        </div>

        <!-- Ввод суммы -->
        <div class="form-section">
            <label for="amountInput" class="form-label">Введите сумму:</label>
            <input
                type="text"
                id="amountInput"
                placeholder="0.00"
                class="form-input"
            >
        </div>

        <!-- Кнопки -->
        <div class="button-group">
            <button id="convertBtn" class="btn btn-primary">
                Конвертировать
            </button>
            <button id="copyBtn" class="btn btn-secondary" disabled>
                Копировать в буфер
            </button>
        </div>

        <!-- Результат -->
        <div id="resultSection" class="result-section" style="display: none;">
            <div id="resultText" class="result-text"></div>
        </div>

        <!-- Ошибки -->
        <div id="errorSection" class="error-section" style="display: none;">
            <div id="errorText" class="error-text"></div>
        </div>

        <!-- Индикатор загрузки -->
        <div id="loadingSection" class="loading-section" style="display: none;">
            <div class="spinner"></div>
            <div class="loading-text">Получаю курс...</div>
        </div>
    </div>

    <script src="scripts/main.js" type="module"></script>
</body>
</html>
```

---

## 5. GO BACKEND (App struct)

### 5.1. internal/app/app.go

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

// NewApp создает новый экземпляр App с инициализированными зависимостями
func NewApp() *App {
	// Инициализация кэша (100 элементов, TTL 24 часа)
	cacheInstance := cache.NewLRUCache(100, 24*time.Hour)

	// Converter использует parser.FetchRates напрямую
	currencyConverter := converter.NewConverter(parser.FetchRates, cacheInstance)

	return &App{
		converter: currencyConverter,
	}
}

// Startup вызывается при старте приложения
func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
}

// ConvertRequest структура запроса на конвертацию из JavaScript
type ConvertRequest struct {
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
	Date     string  `json:"date"` // Формат: "DD.MM.YYYY"
}

// Convert метод для вызова из JavaScript
// Принимает запрос, валидирует, вызывает converter и возвращает результат
func (a *App) Convert(req ConvertRequest) (*models.ConversionResult, error) {
	// Парсинг даты из формата DD.MM.YYYY
	date, err := time.Parse("02.01.2006", req.Date)
	if err != nil {
		return nil, fmt.Errorf("некорректный формат даты. Используйте ДД.ММ.ГГГГ: %w", err)
	}

	// Парсинг валюты
	currency := models.Currency(req.Currency)

	// Вызов конвертера
	result, err := a.converter.Convert(req.Amount, currency, date)
	if err != nil {
		return nil, fmt.Errorf("ошибка конвертации: %w", err)
	}

	return result, nil
}

// GetTodayDate возвращает сегодняшнюю дату в формате DD.MM.YYYY
func (a *App) GetTodayDate() string {
	return time.Now().Format("02.01.2006")
}
```

### 5.2. cmd/currate/main.go

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
	// Создание экземпляра приложения
	application := app.NewApp()

	// Запуск Wails приложения
	err := wails.Run(&options.App{
		Title:  "Конвертер валют (с) BiV 2025 г.",
		Width:  400,
		Height: 650,
		MinWidth:  340,
		MinHeight: 600,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 255, G: 255, B: 255, A: 255},
		OnStartup:        application.Startup,
		Bind: []interface{}{
			application,
		},
		Windows: &options.Windows{
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

---

## 6. JAVASCRIPT ГЛАВНАЯ ЛОГИКА

### 6.1. scripts/main.js

```javascript
import { Calendar } from './calendar.js';
import { formatDate, parseDate, formatAmount } from './utils.js';
import { Convert, GetTodayDate } from '../../wailsjs/go/app/App.js';

// Глобальные переменные
let calendar = null;
let lastResult = '';

// Инициализация при загрузке DOM
document.addEventListener('DOMContentLoaded', async () => {
    await init();
});

async function init() {
    // Получить сегодняшнюю дату из Go
    const today = await GetTodayDate();
    document.getElementById('dateInput').value = today;

    // Инициализация календаря
    calendar = new Calendar('calendarDays', onCalendarDateSelect);

    // Установка обработчиков событий
    setupEventListeners();
}

function setupEventListeners() {
    // Кнопка конвертации
    document.getElementById('convertBtn').addEventListener('click', handleConvert);

    // Кнопка копирования
    document.getElementById('copyBtn').addEventListener('click', handleCopy);

    // Enter на поле суммы
    document.getElementById('amountInput').addEventListener('keypress', (e) => {
        if (e.key === 'Enter') {
            handleConvert();
        }
    });

    // Синхронизация поля даты с календарем
    document.getElementById('dateInput').addEventListener('change', (e) => {
        const dateStr = e.target.value;
        const date = parseDate(dateStr);
        if (date) {
            calendar.setDate(date);
        }
    });

    // Форматирование поля суммы при вводе
    document.getElementById('amountInput').addEventListener('input', (e) => {
        // Разрешаем только цифры, точку и запятую
        e.target.value = e.target.value.replace(/[^\d.,]/g, '');
    });
}

function onCalendarDateSelect(date) {
    // Обновить поле ввода даты
    const formatted = formatDate(date);
    document.getElementById('dateInput').value = formatted;
}

async function handleConvert() {
    // Получить данные из формы
    const dateStr = document.getElementById('dateInput').value.trim();
    const currency = document.querySelector('input[name="currency"]:checked').value;
    const amountStr = document.getElementById('amountInput').value.trim();

    // Скрыть предыдущие сообщения
    hideResult();
    hideError();

    // Базовая валидация
    if (!dateStr) {
        showError('Пожалуйста, укажите дату');
        return;
    }

    if (!amountStr) {
        showError('Пожалуйста, введите сумму');
        return;
    }

    const amount = parseFloat(amountStr.replace(',', '.'));
    if (isNaN(amount) || amount <= 0) {
        showError('Сумма должна быть положительным числом');
        return;
    }

    // Показать загрузку
    showLoading();
    disableButtons();

    try {
        // Вызов Go метода
        const result = await Convert({
            amount: amount,
            currency: currency,
            date: dateStr
        });

        // Отобразить результат
        showResult(result.FormattedStr);
        lastResult = result.FormattedStr;

        // Активировать кнопку копирования
        document.getElementById('copyBtn').disabled = false;

    } catch (error) {
        showError(error.toString());
        lastResult = '';
        document.getElementById('copyBtn').disabled = true;
    } finally {
        hideLoading();
        enableButtons();
    }
}

function handleCopy() {
    if (!lastResult) {
        showError('Нет результата для копирования');
        return;
    }

    navigator.clipboard.writeText(lastResult)
        .then(() => {
            // Показать уведомление
            showNotification('Результат скопирован в буфер обмена');
        })
        .catch((err) => {
            showError('Не удалось скопировать: ' + err);
        });
}

function showResult(text) {
    const section = document.getElementById('resultSection');
    const textEl = document.getElementById('resultText');
    textEl.textContent = text;
    section.style.display = 'block';
}

function hideResult() {
    document.getElementById('resultSection').style.display = 'none';
}

function showError(message) {
    const section = document.getElementById('errorSection');
    const textEl = document.getElementById('errorText');
    textEl.textContent = 'Ошибка: ' + message;
    section.style.display = 'block';

    // Автоскрытие через 5 секунд
    setTimeout(() => {
        hideError();
    }, 5000);
}

function hideError() {
    document.getElementById('errorSection').style.display = 'none';
}

function showLoading() {
    document.getElementById('loadingSection').style.display = 'flex';
}

function hideLoading() {
    document.getElementById('loadingSection').style.display = 'none';
}

function disableButtons() {
    document.getElementById('convertBtn').disabled = true;
}

function enableButtons() {
    document.getElementById('convertBtn').disabled = false;
}

function showNotification(message) {
    // Простое всплывающее уведомление
    const notification = document.createElement('div');
    notification.className = 'notification';
    notification.textContent = message;
    document.body.appendChild(notification);

    setTimeout(() => {
        notification.classList.add('show');
    }, 10);

    setTimeout(() => {
        notification.classList.remove('show');
        setTimeout(() => {
            notification.remove();
        }, 300);
    }, 2000);
}
```

### 6.2. scripts/utils.js

```javascript
// Форматирование даты в DD.MM.YYYY
export function formatDate(date) {
    const day = String(date.getDate()).padStart(2, '0');
    const month = String(date.getMonth() + 1).padStart(2, '0');
    const year = date.getFullYear();
    return `${day}.${month}.${year}`;
}

// Парсинг даты из DD.MM.YYYY
export function parseDate(dateStr) {
    const parts = dateStr.split('.');
    if (parts.length !== 3) return null;

    const day = parseInt(parts[0], 10);
    const month = parseInt(parts[1], 10) - 1;
    const year = parseInt(parts[2], 10);

    const date = new Date(year, month, day);
    if (isNaN(date.getTime())) return null;

    return date;
}

// Форматирование суммы
export function formatAmount(amount) {
    return amount.toLocaleString('ru-RU', {
        minimumFractionDigits: 2,
        maximumFractionDigits: 2
    });
}
```

---

## 7. ЭТАПЫ РЕАЛИЗАЦИИ

### Этап 6.1: Подготовка окружения
- [ ] Установить Wails CLI (`go install github.com/wailsapp/wails/v2/cmd/wails@latest`)
- [ ] Проверить версию: `wails doctor`
- [ ] Инициализировать Wails проект: `wails init`

### Этап 6.2: Создание базовой структуры
- [ ] Создать `cmd/currate/main.go`
- [ ] Создать `internal/app/app.go`
- [ ] Создать `frontend/src/index.html`
- [ ] Создать `wails.json` конфигурацию

### Этап 6.3: Реализация календаря
- [ ] Создать `frontend/src/scripts/calendar.js`
- [ ] Создать `frontend/src/styles/calendar.css`
- [ ] Реализовать логику выделения выходных
- [ ] Добавить навигацию по месяцам
- [ ] Тестирование календаря

### Этап 6.4: Основная логика UI
- [ ] Создать `frontend/src/scripts/main.js`
- [ ] Создать `frontend/src/scripts/utils.js`
- [ ] Создать `frontend/src/styles/main.css`
- [ ] Реализовать обработчики событий

### Этап 6.5: Интеграция с backend
- [ ] Настроить Wails bindings
- [ ] Протестировать вызовы Go из JavaScript
- [ ] Реализовать обработку ошибок
- [ ] Добавить индикатор загрузки

### Этап 6.6: Тестирование и полировка
- [ ] Ручное тестирование всех функций
- [ ] Проверка на разных размерах окна
- [ ] Тестирование копирования в буфер
- [ ] Проверка обработки ошибок

### Этап 6.7: Сборка
- [ ] Запустить `wails dev` для разработки
- [ ] Собрать релиз: `wails build -clean`
- [ ] Протестировать exe файл
- [ ] (Опционально) Сжать с UPX

---

## 8. ОЖИДАЕМЫЙ РЕЗУЛЬТАТ

### 8.1. Размер приложения
- **Без оптимизаций:** ~15-18 MB
- **С `-ldflags "-s -w"`:** ~12-15 MB
- **С UPX сжатием:** ~8-10 MB ✅

### 8.2. Производительность
- **Запуск приложения:** < 1 секунда
- **Время конвертации:** 1-3 секунды (зависит от API ЦБ РФ)
- **Отзывчивость UI:** Мгновенная (асинхронная обработка)

### 8.3. Функциональность
✅ Все требования Python-версии
✅ Выходные дни выделены в календаре
✅ Асинхронная обработка
✅ Копирование в буфер обмена
✅ Обработка ошибок
✅ Современный UI

---

## 9. ИТОГОВАЯ ОЦЕНКА ПЛАНА

| Критерий | Оценка |
|----------|--------|
| **Соответствие требованиям** | ✅ 100% |
| **Техническая реализуемость** | ✅ Высокая |
| **Сложность реализации** | ⭐⭐⭐ Средняя |
| **Время разработки** | 2-3 дня |
| **Поддерживаемость** | ✅ Отличная |
| **Размер exe** | ✅ <= 10 MB (с UPX) |

---

**ПЛАН ГОТОВ К РЕАЛИЗАЦИИ** ✅

**Следующий шаг:** Начать реализацию с Этапа 6.1
