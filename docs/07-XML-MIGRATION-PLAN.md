# ПЛАН МИГРАЦИИ НА XML ПАРСИНГ
## Проект: CurRate Go Rewrite

**Версия:** 1.0
**Дата:** 21 декабря 2025
**Автор:** Ivan Bondarev (BiV)
**Основа:** Python проект CurRate v3.0.0

---

## 📋 СОДЕРЖАНИЕ

1. [Обзор изменений](#1-обзор-изменений)
2. [Анализ Python v3.0.0](#2-анализ-python-v300)
3. [XML API ЦБ РФ](#3-xml-api-цб-рф)
4. [Архитектура Go парсера](#4-архитектура-go-парсера)
5. [План реализации](#5-план-реализации)
6. [Тестирование](#6-тестирование)
7. [Обновление документации](#7-обновление-документации)

---

## 1. ОБЗОР ИЗМЕНЕНИЙ

### 1.1. Мотивация миграции

**Почему переходим на XML API:**

✅ **Официальный API:** XML эндпоинт - это официальный способ получения данных от ЦБ РФ
✅ **Стабильность:** XML структура редко меняется, в отличие от HTML верстки
✅ **Производительность:** XML парсинг быстрее и эффективнее HTML парсинга
✅ **Надежность:** Меньше зависимостей (не требуется сторонний парсер)
✅ **Простота:** `encoding/xml` - встроенная библиотека Go

### 1.2. Ключевые изменения

| Аспект | Было (HTML) | Стало (XML) |
|--------|-------------|-------------|
| **URL** | `https://www.cbr.ru/currency_base/daily/?UniDbQuery.Posted=True&UniDbQuery.To=DD.MM.YYYY` | `https://www.cbr.ru/scripts/XML_daily.asp?date_req=DD/MM/YYYY` |
| **Формат даты** | `DD.MM.YYYY` | `DD/MM/YYYY` (замена `.` → `/`) |
| **Парсер** | goquery (jQuery-подобный) | encoding/xml (стандартная библиотека) |
| **Зависимости** | +1 внешняя библиотека | Только стандартная библиотека |
| **Производительность** | ~50-100 мс парсинг | ~5-10 мс парсинг |
| **Размер ответа** | ~80-100 КБ (HTML) | ~15-20 КБ (XML) |

---

## 2. АНАЛИЗ PYTHON V3.0.0

### 2.1. Структура Python кода

**Файл:** `src/currate/cbr_parser.py` (253 строки)

**Ключевые функции:**

```python
def get_currency_rate(currency: str, date: str, timeout: int = 10) -> float:
    """
    Получает курс валюты с сайта ЦБ РФ на указанную дату.

    Args:
        currency: Код валюты (USD, EUR)
        date: Дата в формате DD.MM.YYYY
        timeout: Таймаут запроса в секундах

    Returns:
        float: Курс валюты за 1 единицу

    Raises:
        CBRConnectionError: При ошибке соединения
        CBRParseError: При ошибке парсинга
    """
```

### 2.2. Улучшения из Python v3.0.0

**1. Retry-стратегия:**
```python
retry_strategy = Retry(
    total=3,                # 3 попытки
    backoff_factor=1,       # Задержка: 1, 2, 4 секунды
    status_forcelist=[429, 500, 502, 503, 504],
    allowed_methods=["GET"]
)
```

**2. Потокобезопасность:**
```python
_session_lock = threading.Lock()     # Для session инициализации
_request_lock = threading.Lock()     # Для HTTP запросов

with _request_lock:
    response = session.get(url, timeout=timeout)
```

**3. Глобальная HTTP сессия:**
```python
_session: Optional[requests.Session] = None

def get_session() -> requests.Session:
    global _session
    if _session is None:
        with _session_lock:
            if _session is None:  # Double-check locking
                _session = create_session_with_retry()
    return _session
```

**4. Нормализация входных данных:**
```python
currency = currency.strip().upper()
date = date.strip()
```

**5. Иерархия ошибок:**
```python
CBRParserError
├── CBRConnectionError (timeout, connection errors)
└── CBRParseError (parsing errors, currency not found)
```

### 2.3. Логика парсинга XML

```python
# 1. Конвертация формата даты
api_date = date.replace('.', '/')  # DD.MM.YYYY → DD/MM/YYYY

# 2. Формирование URL
url = f"https://www.cbr.ru/scripts/XML_daily.asp?date_req={api_date}"

# 3. Парсинг XML
root = ET.fromstring(response.content)

# 4. Поиск валюты
for valute in root.findall('Valute'):
    char_code_elem = valute.find('CharCode')
    if char_code_elem.text.strip() == currency:
        # 5. Извлечение номинала и курса
        nominal = int(valute.find('Nominal').text.strip())
        value_str = valute.find('Value').text.strip().replace(',', '.')
        value = float(value_str)

        # 6. Расчет курса за 1 единицу
        return value / nominal
```

---

## 3. XML API ЦБ РФ

### 3.1. Структура XML ответа

**URL:** `https://www.cbr.ru/scripts/XML_daily.asp?date_req=21/12/2025`

**Пример ответа:**

```xml
<?xml version="1.0" encoding="windows-1251"?>
<ValCurs Date="20.12.2025" name="Foreign Currency Market">
    <Valute ID="R01235">
        <NumCode>840</NumCode>
        <CharCode>USD</CharCode>
        <Nominal>1</Nominal>
        <Name>Доллар США</Name>
        <Value>80,7220</Value>
        <VunitRate>80,722</VunitRate>
    </Valute>
    <Valute ID="R01239">
        <NumCode>978</NumCode>
        <CharCode>EUR</CharCode>
        <Nominal>1</Nominal>
        <Name>Евро</Name>
        <Value>94,5120</Value>
        <VunitRate>94,512</VunitRate>
    </Valute>
    <Valute ID="R01135">
        <NumCode>348</NumCode>
        <CharCode>HUF</CharCode>
        <Nominal>100</Nominal>
        <Name>Форинтов</Name>
        <Value>24,4161</Value>
        <VunitRate>0,244161</VunitRate>
    </Valute>
</ValCurs>
```

### 3.2. Ключевые элементы

| Элемент | Описание | Пример |
|---------|----------|--------|
| `<ValCurs>` | Корневой элемент | `Date="20.12.2025"` |
| `<Valute>` | Данные по валюте | `ID="R01235"` |
| `<NumCode>` | Числовой код валюты (ISO 4217) | `840` (USD) |
| `<CharCode>` | **Буквенный код валюты** | `USD`, `EUR` |
| `<Nominal>` | **Номинал (за сколько единиц указан курс)** | `1`, `100` |
| `<Name>` | Название валюты на русском | `Доллар США` |
| `<Value>` | **Курс за номинал (запятая!)** | `80,7220` |
| `<VunitRate>` | Курс за 1 единицу | `80,722` |

### 3.3. Важные особенности

**1. Кодировка:** `windows-1251` (НЕ UTF-8!)

**2. Разделитель дробной части:** Запятая `,` (НЕ точка `.`)
```
80,7220  →  Нужно заменить на  →  80.7220
```

**3. Номинал (Nominal):**
- Для большинства валют: `1` (USD, EUR, RUB)
- Для некоторых: `10`, `100`, `1000`, `10000` (HUF, VND и др.)
- **Расчет курса за 1 единицу:** `rate = Value / Nominal`

**Пример:**
```
100 HUF = 24,4161 RUB
Курс за 1 HUF = 24,4161 / 100 = 0,244161 RUB
```

**4. Формат даты в URL:** `DD/MM/YYYY` (slash, НЕ dot!)

---

## 4. АРХИТЕКТУРА GO ПАРСЕРА

### 4.1. Структуры данных

```go
// XML структуры для unmarshal
type ValCurs struct {
    XMLName xml.Name `xml:"ValCurs"`
    Date    string   `xml:"Date,attr"`
    Valutes []Valute `xml:"Valute"`
}

type Valute struct {
    ID       string  `xml:"ID,attr"`
    NumCode  string  `xml:"NumCode"`
    CharCode string  `xml:"CharCode"`
    Nominal  int     `xml:"Nominal"`
    Name     string  `xml:"Name"`
    Value    string  `xml:"Value"`      // Строка, т.к. запятая
}
```

### 4.2. Основная функция парсинга

```go
package parser

import (
    "encoding/xml"
    "fmt"
    "io"
    "net/http"
    "strconv"
    "strings"
    "time"
)

// FetchRate получает курс валюты с XML API ЦБ РФ
func FetchRate(currency string, date time.Time) (float64, error) {
    // 1. Нормализация входных данных
    currency = strings.TrimSpace(strings.ToUpper(currency))

    // 2. Конвертация формата даты: DD.MM.YYYY → DD/MM/YYYY
    dateStr := date.Format("02/01/2006")  // Go формат для DD/MM/YYYY

    // 3. Формирование URL
    url := fmt.Sprintf(
        "https://www.cbr.ru/scripts/XML_daily.asp?date_req=%s",
        dateStr,
    )

    // 4. HTTP запрос
    resp, err := http.Get(url)
    if err != nil {
        return 0, fmt.Errorf("failed to fetch rates: %w", err)
    }
    defer resp.Body.Close()

    // 5. Чтение body
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return 0, fmt.Errorf("failed to read response: %w", err)
    }

    // 6. Парсинг XML
    var valCurs ValCurs
    if err := xml.Unmarshal(body, &valCurs); err != nil {
        return 0, fmt.Errorf("failed to parse XML: %w", err)
    }

    // 7. Поиск валюты
    for _, valute := range valCurs.Valutes {
        if strings.TrimSpace(valute.CharCode) == currency {
            // 8. Парсинг значения (заменяем запятую на точку)
            valueStr := strings.ReplaceAll(valute.Value, ",", ".")
            value, err := strconv.ParseFloat(valueStr, 64)
            if err != nil {
                return 0, fmt.Errorf("failed to parse value: %w", err)
            }

            // 9. Расчет курса за 1 единицу
            rate := value / float64(valute.Nominal)
            return rate, nil
        }
    }

    // 10. Валюта не найдена
    return 0, fmt.Errorf("currency %s not found", currency)
}
```

### 4.3. Retry логика (опционально)

**Вариант 1: Простой retry (рекомендуется для начала)**

```go
func fetchWithRetry(url string, maxRetries int) (*http.Response, error) {
    var resp *http.Response
    var err error

    for i := 0; i < maxRetries; i++ {
        resp, err = http.Get(url)
        if err == nil && resp.StatusCode == http.StatusOK {
            return resp, nil
        }

        if i < maxRetries-1 {
            // Экспоненциальная задержка: 1s, 2s, 4s
            time.Sleep(time.Duration(1<<i) * time.Second)
        }
    }

    return nil, fmt.Errorf("failed after %d retries: %w", maxRetries, err)
}
```

**Вариант 2: HTTP Client с настройками (production)**

```go
var (
    httpClient *http.Client
    once       sync.Once
)

func getHTTPClient() *http.Client {
    once.Do(func() {
        httpClient = &http.Client{
            Timeout: 10 * time.Second,
            Transport: &http.Transport{
                MaxIdleConns:        10,
                MaxIdleConnsPerHost: 2,
                IdleConnTimeout:     30 * time.Second,
                DisableCompression:  false,
            },
        }
    })
    return httpClient
}
```

### 4.4. Обработка ошибок

```go
package parser

import "errors"

var (
    ErrInvalidCurrency = errors.New("invalid currency code")
    ErrCurrencyNotFound = errors.New("currency not found in response")
    ErrInvalidDate     = errors.New("invalid date")
    ErrNetworkFailure  = errors.New("network request failed")
    ErrParseFailure    = errors.New("XML parsing failed")
)

// ParseError содержит детали ошибки парсинга
type ParseError struct {
    Currency string
    Date     time.Time
    Err      error
}

func (e *ParseError) Error() string {
    return fmt.Sprintf(
        "failed to parse rate for %s on %s: %v",
        e.Currency,
        e.Date.Format("02.01.2006"),
        e.Err,
    )
}

func (e *ParseError) Unwrap() error {
    return e.Err
}
```

---

## 5. ПЛАН РЕАЛИЗАЦИИ

### 5.1. Этап 1: Подготовка (1-2 часа)

**Задачи:**

1. ✅ **Изучить Python v3.0.0** - Анализ изменений
2. ✅ **Изучить XML API ЦБ РФ** - Структура и особенности
3. ✅ **Спланировать архитектуру** - Этот документ
4. ⏳ **Обновить документацию** - Технологический стек, архитектура

**Результат:** Полное понимание задачи и архитектуры

### 5.2. Этап 2: Обновление parser пакета (2-3 часа)

**Файлы для изменения:**

1. **internal/parser/cbr.go** - Основной парсер
   - Изменить URL на XML эндпоинт
   - Реализовать XML парсинг вместо HTML
   - Добавить конвертацию формата даты
   - Добавить обработку номинала

2. **internal/parser/client.go** - HTTP клиент
   - Добавить retry логику
   - Настроить таймауты
   - Connection pooling

3. **internal/parser/parser.go** - Интерфейс (если есть)
   - Обновить сигнатуры при необходимости

**Детальный план для cbr.go:**

```go
// Шаг 1: Добавить XML структуры
type ValCurs struct { ... }
type Valute struct { ... }

// Шаг 2: Изменить URL константу
const (
    // OLD: cbrURLFormat = "https://www.cbr.ru/currency_base/daily/..."
    cbrURLFormat = "https://www.cbr.ru/scripts/XML_daily.asp?date_req=%s"
)

// Шаг 3: Добавить функцию конвертации даты
func formatDateForAPI(date time.Time) string {
    return date.Format("02/01/2006")  // DD/MM/YYYY
}

// Шаг 4: Переписать FetchRates() для XML парсинга
func (c *Client) FetchRates(date time.Time) (*models.RateData, error) {
    // Реализация с xml.Unmarshal
}

// Шаг 5: Добавить parseValue() для обработки запятой
func parseValue(valueStr string) (float64, error) {
    cleaned := strings.ReplaceAll(valueStr, ",", ".")
    return strconv.ParseFloat(cleaned, 64)
}
```

### 5.3. Этап 3: Обновление models (30 минут)

**Файл:** `internal/models/rate.go`

**Возможные изменения:**

```go
// Если нужны дополнительные поля из XML
type ExchangeRate struct {
    Currency models.Currency
    Rate     float64
    Nominal  int       // НОВОЕ: номинал из XML
    NumCode  string    // НОВОЕ: числовой код ISO 4217
}
```

**Вероятно, изменения не требуются**, т.к. текущая структура уже подходит.

### 5.4. Этап 4: Обновление тестов (2-3 часа)

**Файлы:**

1. **internal/parser/cbr_test.go**
   - Обновить mock данные (XML вместо HTML)
   - Добавить тесты для обработки номинала
   - Тесты для кодировки windows-1251
   - Тесты для формата даты DD/MM/YYYY

2. **internal/parser/client_test.go**
   - Тесты retry логики
   - Тесты таймаутов

**Пример mock XML данных:**

```go
const mockXMLResponse = `<?xml version="1.0" encoding="windows-1251"?>
<ValCurs Date="20.12.2025" name="Foreign Currency Market">
    <Valute ID="R01235">
        <NumCode>840</NumCode>
        <CharCode>USD</CharCode>
        <Nominal>1</Nominal>
        <Name>Доллар США</Name>
        <Value>80,7220</Value>
        <VunitRate>80,722</VunitRate>
    </Valute>
    <Valute ID="R01239">
        <NumCode>978</NumCode>
        <CharCode>EUR</CharCode>
        <Nominal>1</Nominal>
        <Name>Евро</Name>
        <Value>94,5120</Value>
        <VunitRate>94,512</VunitRate>
    </Valute>
</ValCurs>`
```

**Граничные случаи для тестирования:**

| Случай | Входные данные | Ожидаемый результат |
|--------|---------------|---------------------|
| **Nominal = 1** | USD, 1, 80,7220 | 80.7220 |
| **Nominal = 100** | HUF, 100, 24,4161 | 0.244161 |
| **Nominal = 10000** | VND, 10000, 32,0988 | 0.00320988 |
| **Запятая в Value** | "80,7220" | 80.7220 (парсинг OK) |
| **Валюта не найдена** | "XXX" | error |
| **Некорректный XML** | "invalid xml" | parse error |
| **Network timeout** | timeout | network error |

### 5.5. Этап 5: Интеграционные тесты (1-2 часа)

**Задачи:**

1. **Тест полного цикла:**
   - Converter → Parser → XML API → обратно

2. **Сравнение с Python:**
   - Запустить Python v3.0.0 и Go версию
   - Сравнить результаты для одной и той же даты
   - Убедиться в идентичности

3. **Тест production API:**
   - Реальные запросы к ЦБ РФ
   - Проверка актуальных данных

**Пример интеграционного теста:**

```go
func TestIntegration_RealAPI(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test in short mode")
    }

    // Создаем парсер
    client := parser.NewClient()

    // Запрашиваем сегодняшний курс USD
    date := time.Now()
    rateData, err := client.FetchRates(date)
    require.NoError(t, err)

    // Проверяем, что USD есть в ответе
    usdRate, exists := rateData.Rates[models.USD]
    require.True(t, exists, "USD should be present")

    // Проверяем разумность курса (50-150 рублей за доллар)
    assert.Greater(t, usdRate.Rate, 50.0)
    assert.Less(t, usdRate.Rate, 150.0)

    t.Logf("USD rate: %.4f RUB", usdRate.Rate)
}
```

### 5.6. Этап 6: Обновление документации (1 час)

**Файлы для обновления:**

1. **docs/02-ТЕХНОЛОГИЧЕСКИЙ-СТЕК.md**
   - Убрать goquery из зависимостей
   - Добавить encoding/xml (стандартная библиотека)
   - Обновить раздел "Основные библиотеки"

2. **docs/03-АРХИТЕКТУРНЫЙ-ДИЗАЙН.md**
   - Обновить схему parser модуля
   - Описать XML структуры
   - Обновить примеры кода

3. **docs/04-ПЛАН-РАЗРАБОТКИ.md**
   - Отметить Этап 4 (Parser) как завершенный
   - Обновить статус этапов

4. **CHANGELOG.md**
   - Добавить запись о переходе на XML API

**Пример записи в CHANGELOG.md:**

```markdown
## [Unreleased]

### Changed

- **BREAKING:** Переход на официальный XML API ЦБ РФ вместо HTML парсинга
  - URL изменен: `XML_daily.asp` вместо `currency_base/daily`
  - Формат даты: `DD/MM/YYYY` вместо `DD.MM.YYYY`
  - Удалена зависимость `github.com/PuerkitoBio/goquery`
  - Используется встроенная библиотека `encoding/xml`

### Added

- Retry логика для HTTP запросов (3 попытки с экспоненциальной задержкой)
- Обработка номинала валюты (Nominal) из XML
- Потокобезопасный HTTP клиент с connection pooling

### Improved

- Производительность парсинга улучшена в 5-10 раз
- Размер ответа уменьшен с ~100 КБ до ~20 КБ
- Стабильность: XML структура стабильнее HTML
```

---

## 6. ТЕСТИРОВАНИЕ

### 6.1. Unit тесты

**Coverage target:** 100% для parser пакета

**Критические функции для тестирования:**

1. ✅ XML парсинг (unmarshal)
2. ✅ Конвертация формата даты
3. ✅ Обработка запятой в Value
4. ✅ Расчет курса с учетом Nominal
5. ✅ Обработка ошибок (сеть, парсинг, не найдено)
6. ✅ Retry логика

### 6.2. Table-driven тесты

```go
func TestParseValue(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected float64
        wantErr  bool
    }{
        {
            name:     "simple value with comma",
            input:    "80,7220",
            expected: 80.7220,
            wantErr:  false,
        },
        {
            name:     "value with comma and spaces",
            input:    " 94,5120 ",
            expected: 94.5120,
            wantErr:  false,
        },
        {
            name:     "small value",
            input:    "0,244161",
            expected: 0.244161,
            wantErr:  false,
        },
        {
            name:     "invalid value",
            input:    "invalid",
            expected: 0,
            wantErr:  true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := parseValue(tt.input)
            if tt.wantErr {
                assert.Error(t, err)
                return
            }
            require.NoError(t, err)
            assert.InDelta(t, tt.expected, got, 0.000001)
        })
    }
}
```

### 6.3. Benchmark тесты

```go
func BenchmarkXMLParsing(b *testing.B) {
    xmlData := []byte(mockXMLResponse)

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        var valCurs ValCurs
        _ = xml.Unmarshal(xmlData, &valCurs)
    }
}

// Ожидаемая производительность:
// BenchmarkXMLParsing-8   50000   ~25000 ns/op   (0.025 ms)
// vs HTML парсинг:        5000    ~200000 ns/op  (0.2 ms)
// Ускорение: ~8x
```

---

## 7. ОБНОВЛЕНИЕ ДОКУМЕНТАЦИИ

### 7.1. Checklist обновления

- [ ] **docs/02-ТЕХНОЛОГИЧЕСКИЙ-СТЕК.md**
  - [ ] Убрать goquery из раздела "Основные библиотеки"
  - [ ] Добавить encoding/xml в раздел "Стандартная библиотека Go"
  - [ ] Обновить go.mod пример (удалить goquery)
  - [ ] Обновить раздел "Итоговый список зависимостей"

- [ ] **docs/03-АРХИТЕКТУРНЫЙ-ДИЗАЙН.md**
  - [ ] Обновить диаграмму parser модуля
  - [ ] Добавить XML структуры в описание
  - [ ] Обновить примеры кода парсинга
  - [ ] Добавить описание retry логики

- [ ] **docs/04-ПЛАН-РАЗРАБОТКИ.md**
  - [ ] Отметить "Этап 4: HTTP Parser" как completed
  - [ ] Добавить подзадачу "Миграция на XML API"
  - [ ] Обновить оценки времени

- [ ] **CHANGELOG.md**
  - [ ] Добавить раздел [Unreleased]
  - [ ] Описать breaking changes
  - [ ] Перечислить улучшения

- [ ] **README.md** (если будет создан)
  - [ ] Указать использование XML API
  - [ ] Обновить примеры

### 7.2. Пример обновления технологического стека

**Было (в разделе "Основные библиотеки"):**

```markdown
### 3.2. HTML парсинг

**goquery - jQuery-like HTML парсер**

- **Репозиторий:** https://github.com/PuerkitoBio/goquery
- **Установка:** `go get github.com/PuerkitoBio/goquery`
- **Использование:** Парсинг HTML таблицы с курсами ЦБ РФ
```

**Стало (в разделе "Стандартная библиотека Go"):**

```markdown
### 4.5. Работа с XML

**encoding/xml (стандартная библиотека)**

```go
import "encoding/xml"
```

**Преимущества:**
- ✅ Встроенный в Go
- ✅ Быстрый и эффективный
- ✅ Поддержка struct tags для маппинга
- ✅ Автоматическая обработка кодировок

**Использование в проекте:**
```go
type ValCurs struct {
    XMLName xml.Name `xml:"ValCurs"`
    Date    string   `xml:"Date,attr"`
    Valutes []Valute `xml:"Valute"`
}

var valCurs ValCurs
xml.Unmarshal(data, &valCurs)
```
```

---

## 8. КОНТРОЛЬНЫЕ ТОЧКИ

### 8.1. Критерии готовности

**Этап считается завершенным, если:**

✅ Все тесты проходят (coverage ≥ 100% для parser пакета)
✅ Интеграционный тест с реальным API проходит
✅ Результаты идентичны Python v3.0.0
✅ Документация обновлена
✅ CHANGELOG.md актуализирован
✅ Code review пройден (линтеры, форматирование)

### 8.2. Rollback план

**Если миграция не удалась:**

1. **Откатить изменения в parser/**
   ```bash
   git checkout HEAD~1 internal/parser/
   ```

2. **Восстановить goquery зависимость**
   ```bash
   go get github.com/PuerkitoBio/goquery
   ```

3. **Вернуть старые тесты**
   ```bash
   git checkout HEAD~1 internal/parser/*_test.go
   ```

4. **Создать issue** с описанием проблемы

---

## 9. РИСКИ И МИТИГАЦИИ

| Риск | Вероятность | Влияние | Митигация |
|------|-------------|---------|-----------|
| **ЦБ изменит структуру XML** | Низкая | Высокое | Мониторинг API, версионирование парсера |
| **Кодировка windows-1251 проблемы** | Средняя | Среднее | Явная обработка кодировки, тесты |
| **Запятая вместо точки** | Высокая | Низкое | Функция parseValue() с тестами |
| **Номинал не учтен** | Средняя | Высокое | Тесты на HUF, VND и другие валюты |
| **Network failures** | Высокая | Среднее | Retry логика, таймауты |

---

## 10. ИТОГОВАЯ ОЦЕНКА ВРЕМЕНИ

| Этап | Задачи | Время | Статус |
|------|--------|-------|--------|
| **1. Подготовка** | Анализ, планирование | 1-2 ч | ✅ Завершено |
| **2. Parser** | Реализация XML парсинга | 2-3 ч | ⏳ Ожидание |
| **3. Models** | Обновление структур | 0.5 ч | ⏳ Ожидание |
| **4. Tests** | Unit + интеграционные | 2-3 ч | ⏳ Ожидание |
| **5. Integration** | Тесты с реальным API | 1-2 ч | ⏳ Ожидание |
| **6. Documentation** | Обновление docs | 1 ч | ⏳ Ожидание |
| **Итого** | | **7-11 часов** | |

**Рекомендация:** Разбить на 2-3 рабочих сессии по 3-4 часа.

---

## 11. СЛЕДУЮЩИЕ ШАГИ

**После завершения миграции на XML:**

1. ✅ **Удалить goquery зависимость**
   ```bash
   go mod tidy
   ```

2. ✅ **Обновить go.mod и go.sum**
   ```bash
   go mod download
   ```

3. ✅ **Запустить все тесты**
   ```bash
   go test ./...
   go test -race ./...
   go test -cover ./...
   ```

4. ✅ **Проверить производительность**
   ```bash
   go test -bench=. ./internal/parser/
   ```

5. ✅ **Создать commit**
   ```bash
   git add .
   git commit -m "feat: migrate to CBR XML API

   - Replace HTML parsing with official XML API
   - Remove goquery dependency, use encoding/xml
   - Add retry logic and connection pooling
   - Handle currency nominal correctly
   - 5-10x faster parsing performance

   BREAKING CHANGE: API URL changed from currency_base/daily to XML_daily.asp"
   ```

6. ✅ **Создать tag**
   ```bash
   git tag -a v0.2.0 -m "XML API migration"
   ```

---

**Конец документа**

---

**Подготовлено:** Ivan Bondarev (BiV)
**Дата:** 21.12.2025
**Версия:** 1.0
