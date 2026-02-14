# 🧪 Тестирование CurRate-Go

> **Руководство по тестированию, покрытие тестами и запуск тестов**

---

## Обзор

Проект имеет **test coverage >90%** с unit и интеграционными тестами.

---

## Покрытие по модулям

| Модуль | Покрытие | Тесты | Статус |
|--------|----------|-------|--------|
| `internal/models` | 100.0% | ✅ | Полное покрытие |
| `internal/converter` | 98.8% | ✅ | Почти полное покрытие |
| `internal/cache` | 97.8% | ✅ | Почти полное покрытие |
| `internal/parser` | 97.5% | ✅ | Почти полное покрытие |
| `internal/app` | 82.1% | ✅ | Хорошее покрытие |
| `internal/telegram` | 78.2% | ✅ | Хорошее покрытие |
| **Общее** | **>90%** | ✅ | Отличное покрытие |

*Актуально на 2026-02-14. Проверить: `go test -cover ./internal/...`*

---

## Запуск тестов

### Все тесты

```bash
# Запустить все тесты
go test ./...

# Запустить с подробным выводом
go test -v ./...

# Запустить только быстрые тесты (без интеграционных)
go test -short ./...
```

### С покрытием

```bash
# Запустить с покрытием
go test -coverprofile=coverage.out ./...

# Просмотреть покрытие в браузере
go tool cover -html=coverage.out

# Просмотреть покрытие в консоли
go tool cover -func=coverage.out
```

### Интеграционные тесты

```bash
# Запустить интеграционные тесты с реальным API ЦБ РФ
go test -v -tags=integration ./internal/parser

# Запустить все тесты включая интеграционные
go test -v -tags=integration ./internal/parser
```

**Примечание:** Интеграционные тесты делают реальные HTTP запросы к API ЦБ РФ. Используйте их осторожно, чтобы не превысить лимиты API.

### Benchmarks

```bash
# Запустить все benchmarks
go test -bench=. -benchmem ./...

# Запустить конкретный benchmark
go test -bench=BenchmarkLRUCache -benchmem ./internal/cache

# Запустить benchmarks с профилированием
go test -bench=. -cpuprofile=cpu.prof -memprofile=mem.prof ./...
```

---

## Типы тестов

### 1. Unit тесты

**Назначение:** Тестирование отдельных компонентов изолированно

**Примеры:**
- Тесты моделей данных (`internal/models`)
- Тесты кэша (`internal/cache`)
- Тесты конвертера (`internal/converter`)
- Тесты Telegram (`internal/telegram`)

**Особенности:**
- Быстрые (не требуют внешних зависимостей)
- Детерминированные (предсказуемые результаты)
- Изолированные (не зависят от других тестов)

### 2. Интеграционные тесты

**Назначение:** Тестирование взаимодействия с реальным API

**Примеры:**
- Тесты парсера с реальным CBR API (`internal/parser`)
- Тесты HTTP клиента с реальными запросами

**Особенности:**
- Требуют сетевого подключения
- Могут быть медленнее
- Используют реальные данные

**Теги:**
- Интеграционные тесты помечены тегом `integration`
- Запускаются только с флагом `-tags=integration`

### 3. Benchmarks

**Назначение:** Измерение производительности компонентов

**Примеры:**
- Benchmark LRU кэша
- Benchmark конвертера
- Benchmark парсера

**Особенности:**
- Измеряют время выполнения
- Измеряют использование памяти
- Помогают выявить узкие места

---

## Структура тестов

### Именование

- Тестовые файлы: `*_test.go`
- Тестовые функции: `Test*`
- Benchmark функции: `Benchmark*`
- Примеры: `Example*`

### Организация

```
internal/
├── app/
│   └── app_test.go
├── cache/
│   ├── lru_test.go
│   └── test_helpers_test.go
├── converter/
│   ├── converter_test.go
│   ├── getrate_test.go
│   └── test_helpers_test.go
├── models/
│   ├── currency_test.go
│   ├── rate_test.go
│   └── test_helpers_test.go
├── parser/
│   ├── cbr_integration_test.go
│   ├── cbr_test.go
│   ├── client_test.go
│   ├── parser_test.go
│   ├── test_helpers_test.go
│   ├── xml_additional_test.go
│   ├── xml_case_insensitive_test.go
│   ├── xml_nominal_edge_cases_test.go
│   ├── xml_ratedata_helpers_test.go
│   └── xml_test.go
└── telegram/
    ├── telegram_test.go
    └── userid_test.go
```

---

## Примеры тестов

### Unit тест

```go
func TestConverter_Convert(t *testing.T) {
    // Arrange
    rateProvider := &mockRateProvider{}
    conv := converter.NewConverter(rateProvider, nil)

    // Act
    ctx := context.Background()
    result, err := conv.Convert(ctx, 1000.0, models.USD, time.Now())

    // Assert
    if err != nil {
        t.Fatalf("Convert() error = %v, want nil", err)
    }
    if result == nil {
        t.Fatal("Convert() returned nil result")
    }
    if result.TargetAmount <= 0.0 {
        t.Errorf("Convert() result.TargetAmount = %v, want > 0", result.TargetAmount)
    }
}
```

### Интеграционный тест

```go
//go:build integration

func TestFetchRates_Integration(t *testing.T) {
    // Arrange
    date := time.Now()

    // Act
    rates, err := parser.FetchRates(date)

    // Assert
    if err != nil {
        t.Fatalf("FetchRates() error = %v, want nil", err)
    }
    if rates == nil {
        t.Fatal("FetchRates() returned nil")
    }
    if len(rates.Rates) == 0 {
        t.Error("FetchRates() returned empty rates map")
    }
}
```

### Benchmark

```go
func BenchmarkLRUCache_Get(b *testing.B) {
    cache := cache.NewLRUCache(100, 24*time.Hour)
    cache.Set(models.USD, time.Now(), 80.0, time.Now())

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        cache.Get(models.USD, time.Now())
    }
}
```

---

## Best Practices

### 1. Используйте табличные тесты

```go
func TestConverter_Convert(t *testing.T) {
    tests := []struct {
        name     string
        amount   float64
        currency models.Currency
        wantErr  bool
    }{
        {"valid USD", 1000.0, models.USD, false},
        {"valid EUR", 500.0, models.EUR, false},
        {"zero amount", 0.0, models.USD, true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // test implementation
        })
    }
}
```

### 2. Используйте моки для зависимостей

```go
type mockRateProvider struct{}

func (m *mockRateProvider) FetchRates(ctx context.Context, date time.Time) (*models.RateData, error) {
    rateData := models.NewRateData(date)
    rateData.AddRate(models.ExchangeRate{
        Currency: models.USD,
        Rate:     80.0,
        Nominal:  1,
        Date:     date,
    })
    return rateData, nil
}
```

### 3. Тестируйте граничные случаи

- Нулевые значения
- Отрицательные значения
- Максимальные значения
- Невалидные данные
- Ошибки сети

### 4. Проверяйте ошибки явно

```go
// Проверка отсутствия ошибок
if err != nil {
    t.Fatalf("unexpected error: %v", err)
}

// Проверка наличия результата
if result == nil {
    t.Fatal("result should not be nil")
}

// Проверка равенства
if actual != expected {
    t.Errorf("got %v, want %v", actual, expected)
}
```

---

## CI/CD

Тесты автоматически запускаются в GitHub Actions:

- **При каждом push** в main
- **При каждом Pull Request**
- **Матрица ОС:** Windows (`windows-latest`, без `-race`) + Ubuntu (`ubuntu-latest`, с `-race`)
- **Параллелизм:** `-p 4` для ускорения тестов
- **Покрытие:** артефакт `coverage.out` загружается для каждой ОС

См. `.github/workflows/test.yml` для деталей.

---

## Связанные документы

- **[PERFORMANCE.md](PERFORMANCE.md)** - Benchmarks и производительность
- **[ARCHITECTURE.md](ARCHITECTURE.md)** - Архитектура системы
- **[README.md](../README.md)** - Обзор проекта
- **[CONTRIBUTING.md](../CONTRIBUTING.md)** - Руководство для контрибьюторов

---

<div align="center">

[⬆ Наверх](#-тестирование-currate-go)

</div>

