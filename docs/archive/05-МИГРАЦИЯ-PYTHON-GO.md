# РУКОВОДСТВО ПО МИГРАЦИИ
## Python → Go: Сопоставление кода и концепций

**Версия:** 1.0
**Дата:** 19 декабря 2025

---

## 1. ВВЕДЕНИЕ

Этот документ содержит сопоставление кода Python (текущей реализации CurRate) с кодом Go (новой реализации). Цель - помочь разработчикам понять, как концепции Python переносятся на Go.

---

## 2. СТРУКТУРА ПРОЕКТА

### 2.1. Сопоставление файлов

| Python | Go | Назначение |
|--------|-----|-----------|
| `src/currate/main.py` | `main_gui.go` | Точка входа GUI |
| `src/currate/gui.py` | `internal/app/app.go` | GUI Backend (Wails) |
| `src/currate/currency_converter.py` | `internal/converter/converter.go` | Конвертация |
| `src/currate/cbr_parser.py` | `internal/parser/cbr.go` | Парсинг ЦБ РФ |
| `src/currate/cache.py` | `internal/cache/lru.go` | Кэширование |
| - | `internal/models/currency.go` | Модели данных |
| - | `internal/converter/validator.go` | Валидация |
| - | `internal/converter/formatter.go` | Форматирование |

---

## 3. ОСНОВНЫЕ КОНЦЕПЦИИ

### 3.1. Импорты

**Python:**
```python
import tkinter as tk
from typing import Optional, Tuple
from datetime import datetime
import requests
from bs4 import BeautifulSoup
```

**Go:**
```go
import (
    "time"
    "net/http"

    "github.com/wailsapp/wails/v2"
    "encoding/xml" // Стандартная библиотека для XML парсинга
)
```

**Отличия:**
- Go использует explicit импорты (без wildcard `*`)
- Go группирует импорты: стандартная библиотека → пустая строка → внешние пакеты
- Go не требует псевдонимов для большинства пакетов

---

### 3.2. Типы данных

| Python | Go | Примечания |
|--------|-----|-----------|
| `int` | `int`, `int64`, `int32` | Go различает размеры |
| `float` | `float64`, `float32` | Явное указание precision |
| `str` | `string` | |
| `bool` | `bool` | |
| `list[T]` | `[]T` | Слайсы в Go |
| `dict[K, V]` | `map[K]V` | |
| `tuple` | Struct или несколько return values | |
| `Optional[T]` | `*T` или возврат `(T, error)` | |
| `None` | `nil` | |

---

### 3.3. Классы vs Структуры

**Python:**
```python
class CurrencyConverter:
    def __init__(self, use_cache: bool = True):
        self._use_cache = use_cache
        self._cache = get_cache() if use_cache else None

    def convert(self, amount: float, from_currency: str, date: str) -> Tuple[Optional[float], Optional[float], Optional[str]]:
        # ...
        return result, rate, None
```

**Go:**
```go
type Converter struct {
    useCache bool
    cache    *cache.LRUCache
}

func NewConverter(useCache bool, cache *cache.LRUCache) *Converter {
    return &Converter{
        useCache: useCache,
        cache:    cache,
    }
}

func (c *Converter) Convert(amount float64, currency Currency, date time.Time) (*ConversionResult, error) {
    // ...
    return &ConversionResult{...}, nil
}
```

**Отличия:**
- Go использует функции-конструкторы `NewXxx()` вместо `__init__`
- Методы в Go: `func (receiver Type) MethodName()`
- Вместо multiple return values Go использует struct или `(value, error)`
- Приватные поля в Go: начинаются с lowercase (`useCache`)
- Публичные поля/методы: начинаются с uppercase (`Convert`)

---

## 4. ДЕТАЛЬНОЕ СОПОСТАВЛЕНИЕ МОДУЛЕЙ

### 4.1. Модуль: Models (Currency, Rate)

**Python (неявно):**
```python
# Валюты определены как строки
SUPPORTED_CURRENCIES = ['USD', 'EUR']

# Результат возвращается как tuple
def convert(...) -> Tuple[Optional[float], Optional[float], Optional[str]]:
    return result, rate, None
```

**Go (явно):**
```go
// models/currency.go
type Currency string

const (
    USD Currency = "USD"
    EUR Currency = "EUR"
)

func (c Currency) Validate() error {
    switch c {
    case USD, EUR:
        return nil
    default:
        return ErrUnsupportedCurrency
    }
}

func (c Currency) Symbol() string {
    return map[Currency]string{
        USD: "$",
        EUR: "€",
    }[c]
}

// models/rate.go
type ConversionResult struct {
    Amount       float64
    Currency     Currency
    Rate         float64
    ResultRUB    float64
    FormattedStr string
}
```

**Преимущества Go подхода:**
- Type-safety: невозможно передать некорректную валюту
- Методы на типе: `currency.Validate()`, `currency.Symbol()`
- Явная структура результата вместо tuple

---

### 4.2. Модуль: HTTP клиент и парсер

#### 4.2.1. Создание HTTP клиента

**Python:**
```python
def create_session_with_retry() -> requests.Session:
    session = requests.Session()

    retry_strategy = Retry(
        total=3,
        backoff_factor=1,
        status_forcelist=[429, 500, 502, 503, 504],
        allowed_methods=["GET"]
    )

    adapter = HTTPAdapter(max_retries=retry_strategy)
    session.mount("http://", adapter)
    session.mount("https://", adapter)

    return session

_session: Optional[requests.Session] = None

def get_session() -> requests.Session:
    global _session
    if _session is None:
        _session = create_session_with_retry()
    return _session
```

**Go:**
```go
// parser/client.go
type HTTPClient struct {
    client *http.Client
}

func NewHTTPClient(timeout time.Duration) *HTTPClient {
    return &HTTPClient{
        client: &http.Client{
            Timeout: timeout,
            Transport: &http.Transport{
                MaxIdleConns:       10,
                IdleConnTimeout:    30 * time.Second,
                DisableCompression: false,
            },
        },
    }
}

func (c *HTTPClient) Get(url string) (*http.Response, error) {
    return c.client.Get(url)
}
```

**Отличия:**
- Go не использует глобальные переменные (передаем через DI)
- Retry-логика реализуется вручную в CBRParser
- Go явно указывает timeout

#### 4.2.2. Парсинг HTML

**Python:**
```python
def get_currency_rate(currency: str, date: str, timeout: int = 10) -> float:
    url = f"https://cbr.ru/currency_base/daily/?UniDbQuery.Posted=True&UniDbQuery.To={date}"

    session = get_session()
    response = session.get(url, timeout=timeout)
    response.raise_for_status()

    soup = BeautifulSoup(response.text, 'html.parser')
    table = soup.find('table', class_='data')

    for row in table.find_all('tr')[1:]:
        cols = row.find_all('td')
        if len(cols) < 5:
            continue

        currency_code = cols[1].text.strip()
        if currency_code == currency:
            nominal = int(cols[2].text.strip())
            rate_str = cols[4].text.strip().replace(',', '.')
            rate = float(rate_str)
            return rate / nominal

    raise CBRParseError(f"Валюта {currency} не найдена")
```

**Go:**
```go
// parser/cbr.go
func (p *CBRParser) GetRate(currency Currency, date time.Time) (float64, error) {
    url := p.buildURL(date)

    // Retry логика
    for i := 0; i < p.retry.MaxAttempts; i++ {
        rate, err := p.fetchRate(url, currency)
        if err == nil {
            return rate, nil
        }

        if i < p.retry.MaxAttempts-1 {
            time.Sleep(p.retry.BackoffDuration(i))
        }
    }

    return 0, fmt.Errorf("failed after %d attempts", p.retry.MaxAttempts)
}

func (p *CBRParser) fetchRate(url string, currency Currency) (float64, error) {
    resp, err := p.client.Get(url)
    if err != nil {
        return 0, &NetworkError{Err: err}
    }
    defer resp.Body.Close()

    if resp.StatusCode != 200 {
        return 0, &HTTPError{StatusCode: resp.StatusCode}
    }

    doc, err := goquery.NewDocumentFromReader(resp.Body)
    if err != nil {
        return 0, &ParseError{Err: err}
    }

    return p.parseTable(doc, currency)
}

func (p *CBRParser) parseTable(doc *goquery.Document, currency Currency) (float64, error) {
    table := doc.Find("table.data")
    if table.Length() == 0 {
        return 0, &ParseError{Message: "table.data not found"}
    }

    var rate float64
    var nominal int
    found := false

    table.Find("tr").Each(func(i int, row *goquery.Selection) {
        if found {
            return
        }

        cols := row.Find("td")
        if cols.Length() < 5 {
            return
        }

        currencyCode := strings.TrimSpace(cols.Eq(1).Text())
        if currencyCode != string(currency) {
            return
        }

        nominalStr := strings.TrimSpace(cols.Eq(2).Text())
        nominal, _ = strconv.Atoi(nominalStr)

        rateStr := strings.TrimSpace(cols.Eq(4).Text())
        rateStr = strings.ReplaceAll(rateStr, ",", ".")
        rate, _ = strconv.ParseFloat(rateStr, 64)

        found = true
    })

    if !found {
        return 0, &CurrencyNotFoundError{Currency: currency}
    }

    return rate / float64(nominal), nil
}
```

**Отличия:**
- Go использует `goquery` (аналог BeautifulSoup)
- Go использует callback-функции для итерации: `table.Find("tr").Each(func(...) {...})`
- Go явно возвращает ошибки через `error` интерфейс
- Retry-логика реализована вручную

---

### 4.3. Модуль: Кэш

**Python:**
```python
class CurrencyCache:
    def __init__(self, max_size: int = 100, ttl_hours: int = 24):
        self._cache: OrderedDict[Tuple[str, str], Tuple[float, datetime]] = OrderedDict()
        self._max_size = max_size
        self._ttl = timedelta(hours=ttl_hours)

    def get(self, currency: str, date: str) -> Optional[float]:
        key = (currency, date)

        if key not in self._cache:
            return None

        rate, cached_at = self._cache.pop(key)

        if datetime.now() - cached_at > self._ttl:
            return None

        self._cache[key] = (rate, cached_at)
        return rate

    def set(self, currency: str, date: str, rate: float) -> None:
        key = (currency, date)

        if key in self._cache:
            self._cache.pop(key)
        elif len(self._cache) >= self._max_size:
            self.cleanup_expired()
            if len(self._cache) >= self._max_size:
                self._cache.popitem(last=False)

        self._cache[key] = (rate, datetime.now())
```

**Go:**
```go
// cache/lru.go
type LRUCache struct {
    mu       sync.RWMutex
    cache    map[string]*list.Element
    lru      *list.List
    maxSize  int
    ttl      time.Duration
}

type Entry struct {
    key       string
    rate      float64
    timestamp time.Time
}

func NewLRUCache(maxSize int, ttl time.Duration) *LRUCache {
    return &LRUCache{
        cache:   make(map[string]*list.Element),
        lru:     list.New(),
        maxSize: maxSize,
        ttl:     ttl,
    }
}

func (c *LRUCache) Get(currency Currency, date time.Time) (float64, bool) {
    c.mu.Lock()
    defer c.mu.Unlock()

    key := c.makeKey(currency, date)
    elem, exists := c.cache[key]
    if !exists {
        return 0, false
    }

    entry := elem.Value.(*Entry)

    if time.Since(entry.timestamp) > c.ttl {
        c.lru.Remove(elem)
        delete(c.cache, key)
        return 0, false
    }

    c.lru.MoveToBack(elem)
    return entry.rate, true
}

func (c *LRUCache) Set(currency Currency, date time.Time, rate float64) {
    c.mu.Lock()
    defer c.mu.Unlock()

    key := c.makeKey(currency, date)

    if elem, exists := c.cache[key]; exists {
        entry := elem.Value.(*Entry)
        entry.rate = rate
        entry.timestamp = time.Now()
        c.lru.MoveToBack(elem)
        return
    }

    if c.lru.Len() >= c.maxSize {
        oldest := c.lru.Front()
        if oldest != nil {
            c.lru.Remove(oldest)
            delete(c.cache, oldest.Value.(*Entry).key)
        }
    }

    entry := &Entry{
        key:       key,
        rate:      rate,
        timestamp: time.Now(),
    }
    elem := c.lru.PushBack(entry)
    c.cache[key] = elem
}
```

**Отличия:**
- Go требует explicit thread-safety: `sync.RWMutex`
- Python `OrderedDict` → Go `map + container/list`
- Go использует указатели (`*Entry`) для efficiency
- Go возвращает `(value, bool)` вместо `Optional[T]`

---

### 4.4. Модуль: Конвертер

**Python:**
```python
class CurrencyConverter:
    def convert(self, amount: float, from_currency: str, date: str) -> Tuple[Optional[float], Optional[float], Optional[str]]:
        # Валидация
        if amount <= 0:
            return None, None, "Сумма должна быть положительным числом"

        # Получение курса
        rate = None
        if self._use_cache and self._cache is not None:
            rate = self._cache.get(from_currency, date)

        if rate is None:
            try:
                rate = get_currency_rate(from_currency, date)
                if self._use_cache and self._cache is not None:
                    self._cache.set(from_currency, date, rate)
            except CBRParserError as e:
                return None, None, e.get_user_message()

        # Конвертация
        result = amount * rate
        return result, rate, None

    @staticmethod
    def format_result(amount: float, rate: float, currency: str) -> str:
        result_in_rub = amount * rate
        currency_symbol = "$" if currency == "USD" else "€"

        result_str = f"{result_in_rub:,.2f} руб. ({currency_symbol}{amount:,.2f} по курсу {rate:,.4f})"
        result_str = result_str.replace(',', ' ').replace('.', ',')
        return result_str
```

**Go:**
```go
// converter/converter.go
func (c *Converter) Convert(amount float64, currency Currency, date time.Time) (*ConversionResult, error) {
    // Валидация
    if err := ValidateAmount(amount); err != nil {
        return nil, err
    }

    // Получение курса
    rate, found := c.cache.Get(currency, date)
    if !found {
        var err error
        rate, err = c.parser.GetRate(currency, date)
        if err != nil {
            return nil, err
        }
        c.cache.Set(currency, date, rate)
    }

    // Конвертация
    resultRUB := amount * rate

    // Форматирование
    formatted := FormatResult(amount, rate, currency, resultRUB)

    return &ConversionResult{
        Amount:       amount,
        Currency:     currency,
        Rate:         rate,
        ResultRUB:    resultRUB,
        FormattedStr: formatted,
    }, nil
}

// converter/formatter.go
func FormatResult(amount, rate float64, currency Currency, resultRUB float64) string {
    resultStr := formatNumber(resultRUB)
    amountStr := formatNumber(amount)
    rateStr := fmt.Sprintf("%.4f", rate)
    rateStr = strings.ReplaceAll(rateStr, ".", ",")

    symbol := currency.Symbol()

    return fmt.Sprintf("%s руб. (%s%s по курсу %s)",
        resultStr, symbol, amountStr, rateStr)
}

func formatNumber(num float64) string {
    str := fmt.Sprintf("%.2f", num)
    str = strings.ReplaceAll(str, ".", ",")

    parts := strings.Split(str, ",")
    intPart := addThousandsSeparator(parts[0])

    return intPart + "," + parts[1]
}
```

**Отличия:**
- Go возвращает struct + error вместо tuple
- Go использует dependency injection (parser, cache передаются в конструктор)
- Форматирование вынесено в отдельный файл

---

### 4.5. Модуль: GUI

#### 4.5.1. Создание окна

**Python (Tkinter):**
```python
class CurrencyConverterApp:
    def __init__(self, root: tk.Tk):
        self.root = root
        self.converter = CurrencyConverter(use_cache=True)

        self.root.title("Конвертер валют (с) BiV 2024 г.")
        self.root.minsize(340, 455)

        self._create_widgets()

    def _create_widgets(self) -> None:
        self._create_date_widgets()
        self._create_calendar()
        self._create_currency_widgets()
        self._create_amount_widgets()
        self._create_buttons()
        self._create_result_label()
```

**Go (Walk):**
```go
// gui/window.go
type AppWindow struct {
    *walk.MainWindow
    converter     *converter.Converter

    dateEdit      *walk.DateEdit
    usdRadio      *walk.RadioButton
    eurRadio      *walk.RadioButton
    amountEdit    *walk.LineEdit
    convertBtn    *walk.PushButton
    copyBtn       *walk.PushButton
    resultLabel   *walk.Label
}

func Run(conv *converter.Converter) error {
    app := &AppWindow{converter: conv}

    return MainWindow{
        AssignTo: &app.MainWindow,
        Title:    "Конвертер валют (с) BiV 2024 г.",
        MinSize:  Size{Width: 340, Height: 455},
        Size:     Size{Width: 340, Height: 455},
        Layout:   VBox{},
        Children: []Widget{
            Label{Text: "Дата:"},
            DateEdit{
                AssignTo: &app.dateEdit,
                Format:   "02.01.2006",
            },
            Label{Text: "Выберите валюту:"},
            RadioButtonGroup{
                DataMember: "Currency",
                Buttons: []RadioButton{
                    {Text: "USD", Value: "USD", AssignTo: &app.usdRadio},
                    {Text: "EUR", Value: "EUR", AssignTo: &app.eurRadio},
                },
            },
            // ... остальные виджеты
        },
    }.Create()
}
```

**Отличия:**
- Walk использует декларативный DSL
- Tkinter: императивный стиль (`.pack()`, `.grid()`)
- Walk: все виджеты описываются в одной структуре
- Walk использует `AssignTo` для получения ссылок на виджеты

#### 4.5.2. Обработка событий

**Python:**
```python
def _create_buttons(self) -> None:
    self.convert_button = ttk.Button(
        self.root,
        text="Конвертировать",
        command=self._on_convert
    )
    self.convert_button.pack(pady=10)

def _on_convert(self) -> None:
    date = self.date_entry.get()
    currency = self.currency_var.get().strip().upper()
    amount_str = self.amount_entry.get()

    amount = CurrencyConverter.parse_amount(amount_str)
    if amount is None:
        self._show_error("Некорректное значение суммы")
        return

    # Фоновый запрос
    worker = threading.Thread(
        target=self._perform_conversion,
        args=(amount, currency, date),
        daemon=True
    )
    worker.start()
```

**Go:**
```go
// gui/window.go (в декларации)
PushButton{
    AssignTo: &app.convertBtn,
    Text:     "Конвертировать",
    OnClicked: app.onConvert,
}

// gui/callbacks.go
func (app *AppWindow) onConvert() {
    date := app.dateEdit.Date()
    currency := app.getSelectedCurrency()
    amountStr := app.amountEdit.Text()

    amount, err := utils.ParseAmount(amountStr)
    if err != nil {
        app.showError("Некорректное значение суммы")
        return
    }

    // Фоновый запрос
    go app.performConversion(amount, currency, date)
}

func (app *AppWindow) performConversion(amount float64, currency Currency, date time.Time) {
    result, err := app.converter.Convert(amount, currency, date)

    // Обновление GUI из goroutine
    app.Synchronize(func() {
        if err != nil {
            app.showError(err.Error())
            return
        }
        app.resultLabel.SetText(result.FormattedStr)
        app.copyBtn.SetEnabled(true)
    })
}
```

**Отличия:**
- Python использует `threading.Thread` для фоновых задач
- Go использует `goroutine` (просто `go func()`)
- Walk требует `Synchronize()` для обновления GUI из goroutine
- Tkinter требует `after()` или `queue` для межпоточной коммуникации

---

## 5. ОБРАБОТКА ОШИБОК

### 5.1. Python подход

**Python:**
```python
try:
    rate = get_currency_rate(currency, date)
except CBRConnectionError as e:
    return None, None, e.get_user_message()
except CBRParseError as e:
    return None, None, e.get_user_message()
except Exception as e:
    return None, None, f"Неожиданная ошибка: {e}"
```

### 5.2. Go подход

**Go:**
```go
rate, err := p.parser.GetRate(currency, date)
if err != nil {
    var netErr *NetworkError
    var parseErr *ParseError

    if errors.As(err, &netErr) {
        return nil, errors.New(netErr.UserMessage())
    } else if errors.As(err, &parseErr) {
        return nil, errors.New(parseErr.UserMessage())
    }

    return nil, fmt.Errorf("неожиданная ошибка: %w", err)
}
```

**Отличия:**
- Go не использует exceptions
- Ошибки возвращаются явно через `error`
- Type assertion для проверки типа ошибки: `errors.As()`
- Wrapping ошибок: `fmt.Errorf("... %w", err)`

---

## 6. ТЕСТИРОВАНИЕ

### 6.1. Unit-тесты

**Python (pytest):**
```python
def test_parse_amount_simple():
    assert CurrencyConverter.parse_amount("1000") == 1000.0

def test_parse_amount_with_spaces():
    assert CurrencyConverter.parse_amount("1 000") == 1000.0

def test_parse_amount_invalid():
    assert CurrencyConverter.parse_amount("abc") is None
```

**Go (testing):**
```go
func TestParseAmount(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    float64
        wantErr bool
    }{
        {"simple", "1000", 1000.0, false},
        {"with spaces", "1 000", 1000.0, false},
        {"invalid", "abc", 0, true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := ParseAmount(tt.input)

            if (err != nil) != tt.wantErr {
                t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
            }

            if got != tt.want {
                t.Errorf("got = %v, want %v", got, tt.want)
            }
        })
    }
}
```

**Отличия:**
- Go использует table-driven tests
- Go не требует внешних библиотек для тестов (встроенный `testing`)
- Запуск: `go test ./...`

---

## 7. CHEAT SHEET: ЧАСТЫЕ ПАТТЕРНЫ

### 7.1. Null/None значения

**Python:**
```python
def get_value() -> Optional[int]:
    if condition:
        return None
    return 42

value = get_value()
if value is not None:
    print(value)
```

**Go:**
```go
func getValue() (int, bool) {
    if condition {
        return 0, false
    }
    return 42, true
}

value, ok := getValue()
if ok {
    fmt.Println(value)
}
```

### 7.2. String formatting

**Python:**
```python
s = f"Amount: {amount:.2f}, Currency: {currency}"
```

**Go:**
```go
s := fmt.Sprintf("Amount: %.2f, Currency: %s", amount, currency)
```

### 7.3. Списки/массивы

**Python:**
```python
items = [1, 2, 3, 4, 5]
for item in items:
    print(item)

filtered = [x for x in items if x > 2]
```

**Go:**
```go
items := []int{1, 2, 3, 4, 5}
for _, item := range items {
    fmt.Println(item)
}

var filtered []int
for _, x := range items {
    if x > 2 {
        filtered = append(filtered, x)
    }
}
```

### 7.4. Словари/мапы

**Python:**
```python
m = {"key": "value"}
value = m.get("key")  # None if not exists

for key, value in m.items():
    print(key, value)
```

**Go:**
```go
m := map[string]string{"key": "value"}
value, ok := m["key"]  // ok=false if not exists

for key, value := range m {
    fmt.Println(key, value)
}
```

### 7.5. Context managers

**Python:**
```python
with open("file.txt") as f:
    content = f.read()
```

**Go:**
```go
f, err := os.Open("file.txt")
if err != nil {
    return err
}
defer f.Close()

content, err := io.ReadAll(f)
```

---

## 8. ПОЛЕЗНЫЕ РЕСУРСЫ

### 8.1. Для изучения Go
- **A Tour of Go:** https://go.dev/tour/
- **Go by Example:** https://gobyexample.com/
- **Effective Go:** https://go.dev/doc/effective_go

### 8.2. Библиотеки
- **Walk (GUI):** https://github.com/lxn/walk
- **goquery (HTML):** https://github.com/PuerkitoBio/goquery
- **Go standard library:** https://pkg.go.dev/std

### 8.3. Для Python разработчиков
- **Python to Go:** https://github.com/golang/go/wiki/FromPythonToGo
- **Comparison:** https://yourbasic.org/golang/go-vs-python/

---

**Конец документа**

**Подготовлено:** Ivan Bondarev (BiV)
**Дата:** 19.12.2025
