# CurRate-Go - API Документация Desktop GUI

> **Версия документа:** 1.1
> **Дата создания:** 2025-12-22
> **Обновлено:** 2025-12-24
> **Статус:** Актуально (соответствует реализации)
> **Целевая аудитория:** Разработчики, интеграторы

---

## 📖 Содержание

1. [Введение](#введение)
2. [Go Backend API](#go-backend-api)
3. [JavaScript Frontend API](#javascript-frontend-api)
4. [Модели данных (DTO)](#модели-данных-dto)
5. [Обработка ошибок](#обработка-ошибок)
6. [Примеры интеграции](#примеры-интеграции)
7. [Типы данных и валидация](#типы-данных-и-валидация)

---

## Введение

Этот документ описывает все публичные API методы для desktop GUI приложения CurRate-Go, построенного на Wails v2.

### Архитектура API

```
┌──────────────────────────────────────────┐
│         JavaScript Frontend              │
│   - Вызов через window.go.app.App.*     │
│   - Автоматическая JSON сериализация     │
└───────────────┬──────────────────────────┘
                │ Wails Runtime Bridge
                │ (автогенерация биндингов)
┌───────────────▼──────────────────────────┐
│          Go Backend (App struct)         │
│   - Методы экспортированы (Uppercase)    │
│   - Возвращают JSON-сериализуемые типы   │
└──────────────────────────────────────────┘
```

### Принципы API

1. **Типобезопасность:** Все параметры и возвращаемые значения строго типизированы
2. **JSON сериализация:** Автоматическая конвертация Go ↔ JSON ↔ JavaScript
3. **Обработка ошибок:** Ошибки возвращаются в структурированном виде (не через panic)
4. **Валидация:** Все входные данные валидируются на стороне backend
5. **Асинхронность:** Все Go методы вызываются асинхронно из JavaScript (Promise-based)

---

## Go Backend API

Backend API реализован в `internal/app/app.go` через структуру `App`.

### Структура App

```go
// App struct — основной контейнер для GUI backend
type App struct {
	converter *converter.Converter
}
```

**Инициализация:**

```go
func NewApp(conv *converter.Converter) *App
```

**Параметры:**
- `conv` — экземпляр конвертера валют (из `internal/converter`)

**Возвращает:**
- Указатель на `*App`

---

### Метод: `Convert`

Выполняет конвертацию валюты в рубли на заданную дату.

#### Сигнатура

```go
func (a *App) Convert(req ConvertRequest) ConvertResponse
```

#### Параметры

**ConvertRequest** — структура запроса:

```go
type ConvertRequest struct {
	Amount   float64 `json:"amount"`   // Сумма для конвертации (> 0)
	Currency string  `json:"currency"` // Код валюты: "USD", "EUR" или "RUB" (для API)
	Date     string  `json:"date"`     // Дата в формате "DD.MM.YYYY"
	RefDate  string  `json:"refDate"`  // Референсная дата "DD.MM.YYYY" (необязательно)
}

**Примечание о валютах:**
- **UI (пользовательский интерфейс):** поддерживает только USD и EUR
- **Backend API:** поддерживает USD, EUR и RUB
- RUB не включен в UI, так как конвертация RUB→RUB не имеет практического смысла
- При вызове API с RUB, backend возвращает курс 1.0
```

| Поле | Тип | Обязательное | Описание | Пример |
|------|-----|--------------|----------|--------|
| `amount` | `float64` | ✅ | Сумма для конвертации | `1000.50` |
| `currency` | `string` | ✅ | Код валюты (USD, EUR для UI; USD, EUR, RUB для API) | `"USD"` |
| `date` | `string` | ✅ | Дата в российском формате | `"22.12.2025"` |
| `refDate` | `string` | ❌ | Референсная дата: ту же сумму пересчитать по курсу на неё и вернуть разность | `"01.12.2025"` |

#### Возвращаемое значение

**ConvertResponse** — структура ответа:

```go
type ConvertResponse struct {
	Success bool   `json:"success"` // true если успех, false если ошибка
	Result  string `json:"result"`  // Форматированная строка результата (если success=true)
	Error   string `json:"error"`   // Сообщение об ошибке (если success=false)

	// Поля для интерфейса
	SourceAmount    float64 `json:"sourceAmount"`
	TargetAmountRUB float64 `json:"targetAmountRUB"` // округлено до копеек (половина вверх)
	Rate            float64 `json:"rate"`
	Currency        string  `json:"currency"`
	CurrencySymbol  string  `json:"currencySymbol"`
	RequestedDate   string  `json:"requestedDate"`
	ActualDate      string  `json:"actualDate"` // на выходные ЦБ отдаёт курс прошлого рабочего дня

	// Референс (если refDate задан). Ошибка референса не отменяет основной
	// результат: success=true, hasReference=false, текст в refError
	HasReference     bool    `json:"hasReference"`
	RefRate          float64 `json:"refRate"`
	RefAmountRUB     float64 `json:"refAmountRUB"`
	RefDiffRUB       float64 `json:"refDiffRUB"` // targetAmountRUB - refAmountRUB, точно до копейки
	RefRequestedDate string  `json:"refRequestedDate"`
	RefActualDate    string  `json:"refActualDate"`
	RefError         string  `json:"refError"`
}
```

| Поле | Тип | Всегда присутствует | Описание |
|------|-----|---------------------|----------|
| `success` | `bool` | ✅ | Флаг успешности операции |
| `result` | `string` | Только если `success=true` | Форматированная строка: `"80 722,00 руб. ($1 000,00 по курсу 80,7220)"` |
| `error` | `string` | Только если `success=false` | Понятное сообщение об ошибке |

#### Примеры

**Пример 1: Успешная конвертация USD**

Запрос:
```json
{
  "amount": 1000,
  "currency": "USD",
  "date": "22.12.2025"
}
```

Ответ:
```json
{
  "success": true,
  "result": "80 722,00 руб. ($1 000,00 по курсу 80,7220)",
  "error": ""
}
```

**Пример 2: Успешная конвертация EUR**

Запрос:
```json
{
  "amount": 500,
  "currency": "EUR",
  "date": "20.12.2025"
}
```

Ответ:
```json
{
  "success": true,
  "result": "43 225,50 руб. (€500,00 по курсу 86,4510)",
  "error": ""
}
```

**Пример 3: Ошибка — некорректная дата**

Запрос:
```json
{
  "amount": 1000,
  "currency": "USD",
  "date": "invalid-date"
}
```

Ответ:
```json
{
  "success": false,
  "result": "",
  "error": "Неверный формат даты: invalid-date. Используйте формат ДД.ММ.ГГГГ"
}
```

**Пример 4: Ошибка — дата в будущем**

Запрос:
```json
{
  "amount": 1000,
  "currency": "USD",
  "date": "31.12.2030"
}
```

Ответ:
```json
{
  "success": false,
  "result": "",
  "error": "Дата не может быть в будущем"
}
```

**Пример 5: Ошибка — некорректная сумма**

Запрос:
```json
{
  "amount": -100,
  "currency": "USD",
  "date": "22.12.2025"
}
```

Ответ:
```json
{
  "success": false,
  "result": "",
  "error": "Сумма должна быть положительным числом"
}
```

**Пример 6: Ошибка — неподдерживаемая валюта**

Запрос:
```json
{
  "amount": 1000,
  "currency": "GBP",
  "date": "22.12.2025"
}
```

Ответ:
```json
{
  "success": false,
  "result": "",
  "error": "Неподдерживаемая валюта: GBP"
}
```

#### Алгоритм работы

```
1. Валидация даты (формат DD.MM.YYYY)
   └─ Ошибка → return { success: false, error: "..." }

2. Парсинг валюты (USD/EUR)
   └─ Ошибка → return { success: false, error: "..." }

3. Вызов converter.Convert(amount, currency, date)
   3.1. Проверка кэша
        └─ Есть в кэше → использовать кэшированный курс
        └─ Нет в кэше → запрос к CBR XML API
             └─ Ошибка сети → return { success: false, error: "..." }
   3.2. Расчёт: rubAmount = amount * rate
   3.3. Форматирование результата

4. Return { success: true, result: "..." }
```

#### Производительность

- **С кэшем (hit):** ~1-2 мс (мгновенно)
- **Без кэша (miss):** ~100-500 мс (зависит от скорости интернета)
- **Кэш TTL:** 24 часа

#### Безопасность

- ✅ Валидация всех входных данных
- ✅ Защита от SQL injection (не применимо, нет БД)
- ✅ Защита от XSS (автоматическая через Wails)
- ✅ Нет раскрытия внутренних ошибок (только понятные сообщения)

---

### Метод: `GetRate`

Получает курс валюты на указанную дату (для live preview). Вызывается из JavaScript при изменении даты для автоматического отображения курса.

#### Сигнатура

```go
func (a *App) GetRate(currencyStr string, dateStr string) RateResponse
```

#### Параметры

- `currencyStr` (string) — код валюты: "USD", "EUR" или "RUB"
- `dateStr` (string) — дата в формате "DD.MM.YYYY"

#### Возвращаемое значение

**RateResponse** — структура ответа:

```go
type RateResponse struct {
	Success    bool    `json:"success"`    // Успешность операции
	Rate       float64 `json:"rate"`       // Курс валюты (если success=true)
	ActualDate string  `json:"actualDate"` // Фактическая дата курса "DD.MM.YYYY"
	Error      string  `json:"error"`      // Сообщение об ошибке (если success=false)
}
```

#### Примеры

**Запрос:**
```javascript
const response = await window.go.app.App.GetRate("USD", "22.12.2025");
```

**Ответ (успех):**
```json
{
  "success": true,
  "rate": 80.7220,
  "error": ""
}
```

**Ответ (ошибка):**
```json
{
  "success": false,
  "rate": 0,
  "error": "Дата не может быть в будущем"
}
```

#### Использование

Метод используется для live preview курса при изменении даты:

```javascript
document.getElementById('date-input').addEventListener('change', async (e) => {
    const currency = document.querySelector('input[name="currency"]:checked').value;
    const date = e.target.value;
    
    const response = await window.go.app.App.GetRate(currency, date);
    if (response.success) {
        document.getElementById('rate-preview').textContent = response.rate.toFixed(4);
    }
});
```

#### Производительность

- Использует оптимизированный метод `converter.GetRate()` без форматирования
- Кэширование автоматическое (24 часа TTL)
- С кэшем: ~1-2 мс, без кэша: ~100-500 мс

---

## JavaScript Frontend API

Frontend API реализован в `frontend/scripts/` и состоит из четырёх основных модулей:

1. **main.js** — основная логика приложения
2. **calendar.js** — календарь с выделением выходных
3. **status-bar.js** — управление строкой состояния
4. **utils.js** — вспомогательные функции

---

### Модуль: main.js

#### Функция: `initApp()`

Инициализация приложения при загрузке DOM.

**Сигнатура:**

```javascript
async function initApp()
```

**Что делает:**

1. Устанавливает текущую дату в поле ввода (используя JavaScript Date)
2. Инициализирует календарь
3. Устанавливает обработчики событий
4. Загружает курс валюты на текущую дату через `GetRate()`

**Использование:**

```javascript
document.addEventListener('DOMContentLoaded', () => {
    initApp();
});
```

---

#### Функция: `handleConvert()`

Выполняет конвертацию валюты.

**Сигнатура:**

```javascript
async function handleConvert()
```

**Что делает:**

1. Считывает значения из UI (amount, currency, date)
2. Валидирует сумму
3. Вызывает `window.go.app.App.Convert()`
4. Показывает результат или ошибку

**Пример вызова:**

```javascript
document.getElementById('convert-btn').addEventListener('click', handleConvert);
```

**Логика:**

```javascript
async function handleConvert() {
    const amount = parseFloat(document.getElementById('amount-input').value);
    const currency = document.querySelector('input[name="currency"]:checked').value;
    const date = document.getElementById('date-input').value;

    // Валидация
    if (!amount || amount <= 0) {
        showError('Введите корректную сумму');
        return;
    }

    // Состояние загрузки
    setLoadingState(true);

    try {
        // Вызов Go метода
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
```

---

#### Функция: `copyToClipboard()`

Копирует результат конвертации в буфер обмена.

**Сигнатура:**

```javascript
function copyToClipboard()
```

**Использование:**

```javascript
document.getElementById('copy-btn').addEventListener('click', copyToClipboard);
```

**Реализация:**

```javascript
function copyToClipboard() {
    const result = document.getElementById('result-text').textContent;

    navigator.clipboard.writeText(result).then(() => {
        // Индикация успеха
        const btn = document.getElementById('copy-btn');
        const originalText = btn.textContent;
        btn.textContent = '✓ Скопировано!';

        setTimeout(() => {
            btn.textContent = originalText;
        }, 2000);
    }).catch(err => {
        console.error('Ошибка копирования:', err);
        showError('Не удалось скопировать в буфер обмена');
    });
}
```

**Совместимость:**

- ✅ Работает в современных браузерах (WebView2)
- ✅ Требует HTTPS или localhost (WebView2 автоматически обеспечивает)
- ❌ Не работает в старых браузерах (Internet Explorer)

---

#### Функция: `setLoadingState(loading)`

Управляет состоянием загрузки UI.

**Сигнатура:**

```javascript
function setLoadingState(loading)
```

**Параметры:**

- `loading` (boolean) — `true` для активации состояния загрузки, `false` для деактивации

**Реализация:**

```javascript
function setLoadingState(loading) {
    const btn = document.getElementById('convert-btn');

    if (loading) {
        btn.disabled = true;
        btn.textContent = 'Загрузка...';
        btn.classList.add('loading');
    } else {
        btn.disabled = false;
        btn.textContent = 'Конвертировать';
        btn.classList.remove('loading');
    }
}
```

---

#### Функция: `showError(message)`

Показывает сообщение об ошибке.

**Сигнатура:**

```javascript
function showError(message)
```

**Параметры:**

- `message` (string) — текст сообщения об ошибке

**Реализация:**

```javascript
function showError(message) {
    const errorEl = document.getElementById('error-message');
    errorEl.textContent = message;
    errorEl.style.display = 'block';
    errorEl.classList.add('shake'); // Анимация

    setTimeout(() => {
        errorEl.classList.remove('shake');
    }, 500);
}
```

---

#### Функция: `clearError()`

Скрывает сообщение об ошибке.

**Сигнатура:**

```javascript
function clearError()
```

**Реализация:**

```javascript
function clearError() {
    const errorEl = document.getElementById('error-message');
    errorEl.style.display = 'none';
    errorEl.textContent = '';
}
```

---

### Модуль: calendar.js

#### Класс: `Calendar`

Календарь с визуальным выделением выходных дней.

**Конструктор:**

```javascript
class Calendar {
    constructor(containerId, onDateSelect) {
        this.container = document.getElementById(containerId);
        this.onDateSelect = onDateSelect; // Callback при выборе даты
        this.currentDate = new Date();
        this.selectedDate = new Date();
        this.init();
    }
}
```

**Параметры:**

- `containerId` (string) — ID элемента для вставки календаря
- `onDateSelect` (function) — Callback функция, вызываемая при выборе даты

**Пример использования:**

```javascript
const calendar = new Calendar('calendar-container', (date) => {
    console.log('Выбрана дата:', date);
    document.getElementById('date-input').value = formatDate(date);
});
```

---

#### Метод: `toggle()`

Переключает видимость календаря.

**Сигнатура:**

```javascript
toggle()
```

**Использование:**

```javascript
document.getElementById('calendar-btn').addEventListener('click', () => {
    calendar.toggle();
});
```

---

#### Метод: `renderDays(year, month)`

Отрисовывает дни месяца с выделением выходных.

**Сигнатура:**

```javascript
renderDays(year, month)
```

**Параметры:**

- `year` (number) — год (например, 2025)
- `month` (number) — месяц (0-11, где 0 = январь)

**Ключевая логика:**

```javascript
renderDays(year, month) {
    const firstDay = new Date(year, month, 1);
    const lastDay = new Date(year, month + 1, 0);
    const daysContainer = this.container.querySelector('.calendar-days');
    daysContainer.innerHTML = '';

    // Пустые ячейки до первого дня месяца
    const startDayOfWeek = firstDay.getDay();
    const offset = startDayOfWeek === 0 ? 6 : startDayOfWeek - 1;

    for (let i = 0; i < offset; i++) {
        const emptyCell = document.createElement('div');
        emptyCell.className = 'calendar-day empty';
        daysContainer.appendChild(emptyCell);
    }

    // Дни месяца
    for (let day = 1; day <= lastDay.getDate(); day++) {
        const date = new Date(year, month, day);
        const dayOfWeek = date.getDay();

        // ВАЖНО: Определение выходных (суббота = 6, воскресенье = 0)
        const isWeekend = (dayOfWeek === 0 || dayOfWeek === 6);

        const dayCell = document.createElement('div');
        dayCell.className = 'calendar-day';
        dayCell.textContent = day;

        // Добавить класс weekend для выделения красным
        if (isWeekend) {
            dayCell.classList.add('weekend');
        }

        // Сегодняшний день
        if (this.isToday(date)) {
            dayCell.classList.add('today');
        }

        // Выбранный день
        if (this.isSameDay(date, this.selectedDate)) {
            dayCell.classList.add('selected');
        }

        // Обработчик клика
        dayCell.addEventListener('click', () => {
            this.selectDate(date);
        });

        daysContainer.appendChild(dayCell);
    }
}
```

**CSS для выделения выходных:**

```css
/* Выходные дни в заголовке (Сб, Вс) */
.calendar-weekday.weekend {
    color: #d32f2f; /* Красный */
    font-weight: bold;
}

/* Выходные дни в календаре */
.calendar-day.weekend {
    color: #d32f2f;       /* Красный текст */
    background: #ffebee;  /* Светло-красный фон */
}

/* Выбранный день */
.calendar-day.selected {
    background: #4CAF50; /* Зелёный */
    color: white;
}

/* Текущий день */
.calendar-day.today {
    border: 2px solid #2196F3; /* Синяя рамка */
}
```

---

### Модуль: utils.js

Вспомогательные функции для форматирования и валидации.

#### Функция: `formatDate(date)`

Форматирует Date объект в строку DD.MM.YYYY.

**Сигнатура:**

```javascript
function formatDate(date)
```

**Параметры:**

- `date` (Date) — объект даты

**Возвращает:**

- `string` — форматированная дата в формате `DD.MM.YYYY`

**Реализация:**

```javascript
function formatDate(date) {
    const day = String(date.getDate()).padStart(2, '0');
    const month = String(date.getMonth() + 1).padStart(2, '0');
    const year = date.getFullYear();
    return `${day}.${month}.${year}`;
}
```

**Примеры:**

```javascript
formatDate(new Date(2025, 11, 22)); // "22.12.2025"
formatDate(new Date(2025, 0, 1));   // "01.01.2025"
```

---

#### Функция: `parseDate(dateString)`

Парсит строку DD.MM.YYYY в Date объект.

**Сигнатура:**

```javascript
function parseDate(dateString)
```

**Параметры:**

- `dateString` (string) — дата в формате `DD.MM.YYYY`

**Возвращает:**

- `Date` — объект даты или `null` если формат некорректен

**Реализация:**

```javascript
function parseDate(dateString) {
    const regex = /^(\d{2})\.(\d{2})\.(\d{4})$/;
    const match = dateString.match(regex);

    if (!match) {
        return null; // Некорректный формат
    }

    const day = parseInt(match[1], 10);
    const month = parseInt(match[2], 10) - 1; // Месяцы с 0
    const year = parseInt(match[3], 10);

    const date = new Date(year, month, day);

    // Валидация (проверка, что дата существует)
    if (date.getFullYear() !== year ||
        date.getMonth() !== month ||
        date.getDate() !== day) {
        return null; // Несуществующая дата (например, 31.02.2025)
    }

    return date;
}
```

**Примеры:**

```javascript
parseDate("22.12.2025");  // Date(2025, 11, 22)
parseDate("31.02.2025");  // null (несуществующая дата)
parseDate("2025-12-22");  // null (неправильный формат)
```

---

#### Функция: `validateAmount(value)`

Валидирует сумму для конвертации.

**Сигнатура:**

```javascript
function validateAmount(value)
```

**Параметры:**

- `value` (any) — значение для валидации

**Возвращает:**

- `object` — `{ valid: boolean, error: string }`

**Реализация:**

```javascript
function validateAmount(value) {
    const num = parseFloat(value);

    if (isNaN(num)) {
        return { valid: false, error: 'Введите число' };
    }

    if (num <= 0) {
        return { valid: false, error: 'Сумма должна быть больше нуля' };
    }

    if (num > 999999999.99) {
        return { valid: false, error: 'Сумма слишком велика' };
    }

    return { valid: true, error: '' };
}
```

**Примеры:**

```javascript
validateAmount(1000);      // { valid: true, error: '' }
validateAmount(-100);      // { valid: false, error: 'Сумма должна быть больше нуля' }
validateAmount("abc");     // { valid: false, error: 'Введите число' }
validateAmount(0);         // { valid: false, error: 'Сумма должна быть больше нуля' }
```

---

## Модели данных (DTO)

### ConvertRequest

Запрос на конвертацию валюты.

**Go структура:**

```go
type ConvertRequest struct {
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
	Date     string  `json:"date"`
}
```

**JSON пример:**

```json
{
  "amount": 1000.50,
  "currency": "USD",
  "date": "22.12.2025"
}
```

**JavaScript объект:**

```javascript
const request = {
    amount: 1000.50,
    currency: "USD",
    date: "22.12.2025"
};
```

**Валидация:**

| Поле | Правило | Ошибка |
|------|---------|--------|
| `amount` | > 0 | "Сумма должна быть больше нуля" |
| `amount` | float64 | "Некорректное число" |
| `currency` | "USD" или "EUR" | "Неподдерживаемая валюта" |
| `date` | Формат DD.MM.YYYY | "Некорректный формат даты" |
| `date` | Не в будущем | "Дата не может быть в будущем" |
| `date` | >= 01.01.1992 | "Курсы доступны с 1992 года" |

---

### ConvertResponse

Ответ на запрос конвертации.

**Go структура:**

```go
type ConvertResponse struct {
	Success bool   `json:"success"`
	Result  string `json:"result"`
	Error   string `json:"error"`
}
```

**JSON пример (успех):**

```json
{
  "success": true,
  "result": "80 722,00 руб. ($1 000,00 по курсу 80,7220)",
  "error": ""
}
```

**JSON пример (ошибка):**

```json
{
  "success": false,
  "result": "",
  "error": "Дата не может быть в будущем"
}
```

**JavaScript обработка:**

```javascript
const response = await window.go.app.App.Convert(request);

if (response.success) {
    console.log('Результат:', response.result);
} else {
    console.error('Ошибка:', response.error);
}
```

---

## Обработка ошибок

### Типы ошибок

| Код ошибки | Сообщение | Причина |
|------------|-----------|---------|
| `ERR_INVALID_DATE` | "Некорректный формат даты. Используйте ДД.ММ.ГГГГ" | Неправильный формат даты |
| `ERR_DATE_FUTURE` | "Дата не может быть в будущем" | Дата больше сегодняшней |
| `ERR_INVALID_AMOUNT` | "Сумма должна быть больше нуля" | amount <= 0 |
| `ERR_UNSUPPORTED_CURRENCY` | "Неподдерживаемая валюта: XXX" | Валюта не USD/EUR/RUB (для API) или не USD/EUR (для UI) |
| `ERR_NETWORK` | "Ошибка загрузки курсов. Проверьте интернет-соединение" | Нет доступа к CBR API |
| `ERR_RATE_NOT_FOUND` | "Курс валюты на эту дату не найден" | ЦБ РФ не предоставляет курс |

### Обработка в JavaScript

**Структурированная обработка:**

```javascript
async function handleConvert() {
    try {
        const response = await window.go.app.App.Convert(request);

        if (!response.success) {
            // Обработать ошибку из backend
            switch (true) {
                case response.error.includes('формат даты'):
                    showError('❌ ' + response.error);
                    highlightField('date-input');
                    break;
                case response.error.includes('Сумма'):
                    showError('❌ ' + response.error);
                    highlightField('amount-input');
                    break;
                case response.error.includes('интернет'):
                    showError('🌐 ' + response.error);
                    showRetryButton();
                    break;
                default:
                    showError('❌ ' + response.error);
            }
        } else {
            // Успешный результат
            displayResult(response);
        }
    } catch (error) {
        // Обработать ошибку вызова (например, backend недоступен)
        console.error('Fatal error:', error);
        showError('💥 Критическая ошибка. Перезапустите приложение.');
    }
}
```

---

## Примеры интеграции

### Пример 1: Базовая конвертация

```javascript
async function convertUSD() {
    const request = {
        amount: 1000,
        currency: "USD",
        date: "22.12.2025"
    };

    const response = await window.go.app.App.Convert(request);

    if (response.success) {
        console.log(response.result);
        // "80 722,00 руб. ($1 000,00 по курсу 80,7220)"
    } else {
        console.error(response.error);
    }
}
```

### Пример 2: Массовая конвертация

```javascript
async function convertMultiple(amounts) {
    const results = [];

    for (const amount of amounts) {
        const response = await window.go.app.App.Convert({
            amount: amount,
            currency: "USD",
            date: "22.12.2025"
        });

        if (response.success) {
            results.push({
                amount: amount,
                result: response.result
            });
        }
    }

    return results;
}

// Использование
const amounts = [100, 500, 1000, 5000];
const results = await convertMultiple(amounts);
console.table(results);
```

### Пример 3: Динамический выбор даты через календарь

```javascript
// Инициализация календаря
const calendar = new Calendar('calendar-container', (selectedDate) => {
    // Callback при выборе даты
    const formatted = formatDate(selectedDate);
    document.getElementById('date-input').value = formatted;

    // Автоматически триггерить конвертацию при выборе даты
    handleConvert();
});

// Открыть календарь
document.getElementById('calendar-btn').addEventListener('click', () => {
    calendar.toggle();
});
```

### Пример 4: Валидация в реальном времени

```javascript
document.getElementById('amount-input').addEventListener('input', (e) => {
    const value = e.target.value;
    const validation = validateAmount(value);

    if (validation.valid) {
        e.target.classList.remove('invalid');
        clearError();
    } else {
        e.target.classList.add('invalid');
        showError(validation.error);
    }
});
```

---

## Типы данных и валидация

### Поддерживаемые валюты

| Код | Название | Символ | NumCode (ЦБ РФ) |
|-----|----------|--------|-----------------|
| USD | Доллар США | $ | 840 |
| EUR | Евро | € | 978 |

**Планируется в будущих версиях:**

- GBP — Фунт стерлингов (826)
- CNY — Китайский юань (156)
- JPY — Японская иена (392)

### Формат даты

**Формат:** `DD.MM.YYYY`

**Примеры корректных дат:**

- `01.01.2025`
- `22.12.2025`
- `31.12.2024`

**Примеры некорректных дат:**

- `2025-12-22` (ISO формат)
- `22/12/2025` (слэши вместо точек)
- `1.1.25` (короткий формат)
- `31.02.2025` (несуществующая дата)

### Формат суммы

**Тип:** `float64`

**Диапазон:** `0.01` — `999,999,999.99`

**Примеры корректных сумм:**

- `100`
- `1000.50`
- `0.99`
- `1000000`

**Примеры некорректных сумм:**

- `0` (должна быть > 0)
- `-100` (отрицательная)
- `"abc"` (не число)

---

## Заключение

Эта документация описывает все публичные API методы для CurRate-Go Desktop GUI. Используйте её как справочник при разработке, интеграции или расширении функциональности приложения.

### Ключевые моменты

1. **Все Go методы асинхронны** — используйте `async/await` в JavaScript
2. **Валидация происходит на backend** — не полагайтесь только на frontend валидацию
3. **Ошибки всегда в структурированном виде** — проверяйте `response.success`
4. **Кэширование автоматическое** — повторные запросы мгновенны
5. **Выходные дни выделяются красным** — важная особенность календаря

### Полезные ссылки

- 📚 **Wails Bindings:** https://wails.io/docs/howdoesitwork#binding
- 📖 **Wails Runtime:** https://wails.io/docs/reference/runtime/intro
- 🐛 **GitHub Issues:** https://github.com/bivlked/CurRate-Go/issues

---

**Спасибо за использование CurRate-Go API!**

*Документ подготовлен: 2025-12-22*  
*Обновлено: 2025-12-24*  
*Версия: 1.1*  
*Статус: Актуально (соответствует реализации)*
