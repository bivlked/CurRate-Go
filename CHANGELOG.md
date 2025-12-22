# История изменений / Changelog

Все важные изменения в проекте CurRate Go Rewrite будут задокументированы в этом файле.

Формат основан на [Keep a Changelog](https://keepachangelog.com/ru/1.0.0/).

---

## [Unreleased]

### Добавлено (Added)
- **[2025-12-22] Улучшения команды (версия v0.5.1)**
  - `internal/converter/converter.go` - улучшена обработка ошибок
    - Добавлена проверка на nil-провайдер (ErrNilRateProvider)
    - Добавлен noopCache - заглушка для nil cache
    - Оптимизация: RUB → RUB конвертация без вызова API
    - Нормализация дат через normalizeDate()
  - `internal/parser/xml_additional_test.go` - новый файл с расширенными тестами
    - Тест обработки windows-1251 кодировки
    - Тест фильтрации невалидных и неподдерживаемых валют
    - Тест обработки пустого XML
  - `internal/converter/converter_test.go` - расширено покрытие тестами
    - Тесты для nil-провайдера и nil-cache
    - Тесты для RUB → RUB конвертации
    - Дополнительные edge cases
  - `internal/parser/xml_test.go` - расширены тесты XML парсера
    - Тесты для различных номиналов (1, 10, 100, 10000)
    - Тесты обработки запятых в значениях
    - Тесты пропуска неподдерживаемых валют
  - `pkg/utils/number_test.go` - добавлены тесты для ParseAmount
    - Тесты для строк с пробелами
    - Тесты для edge cases

### Изменено (Changed)
- **[2025-12-22] Рефакторинг и улучшения**
  - `internal/parser/parser.go` - удален устаревший HTML парсер
    - Удалены функции: ParseHTML и связанные (78 строк)
    - Оставлены только общие функции: parseRate, parseNominal, parseCurrency
  - `internal/parser/xml.go` - улучшен XML парсер
    - parseXMLValue теперь использует общую функцию parseRate()
    - Добавлено использование parseNominal для валидации
    - Улучшена обработка даты из XML атрибута
  - `internal/parser/client.go` - улучшена обработка HTTP редиректов
    - Добавлена корректная обработка HTTP статусов
  - `internal/parser/client_test.go` - обновлены тесты HTTP клиента
    - Исправлены моки для новой логики
  - `internal/converter/validator.go` - улучшена валидация дат
    - Учет временных зон (timezone-aware validation)
  - `go.sum` - очищены неиспользуемые зависимости (-82 строки)

### Безопасность (Security)
- Улучшена обработка nil-значений для предотвращения panic
- Добавлена валидация всех входных параметров

### Покрытие тестами
- **Общее покрытие: 96.0%** (превышает требование 90%)
  - internal/cache: 100.0%
  - internal/converter: 98.4%
  - internal/models: 100.0%
  - internal/parser: 92.3%
  - pkg/utils: 96.1%

---

### Добавлено (Added)
- **[2025-12-20] Этап 2: Базовые модели данных (завершен)**
  - `internal/models/currency.go` - тип Currency с константами USD, EUR, RUB
  - `internal/models/rate.go` - ExchangeRate, ConversionResult, RateData
  - `pkg/utils/number.go` - ParseAmount и FormatAmount для работы с суммами
  - Unit-тесты с покрытием 97.1%

- **[2025-12-20] Этап 3: HTTP клиент и парсер ЦБ РФ (завершен)**
  - `internal/parser/parser.go` - парсинг HTML таблицы курсов валют
    - Функции: ParseHTML, parseRate, parseNominal, parseCurrency
    - Поддержка формата с запятой как десятичным разделителем
  - `internal/parser/client.go` - HTTP клиент с retry логикой
    - Retry до 3 раз с задержкой 1 секунда
    - Timeout 10 секунд
    - User-Agent: "CurRate-Go/1.0 (Windows; Go)"
  - `internal/parser/cbr.go` - публичный API фасад
    - FetchRates(date) - получение курсов на дату
    - FetchLatestRates() - получение последних курсов
  - Comprehensive unit-тесты с мок HTTP сервером (покрытие 91.3%)
  - 8 файлов: 3 модуля + 3 тестовых файла

- **[2025-12-20] Этап 4: LRU Cache (завершен)**
  - `internal/cache/lru.go` - потокобезопасный LRU кэш с TTL
    - Структуры: LRUCache, Entry
    - Методы: NewLRUCache, Get, Set, Clear, Size, makeKey
    - Использование sync.RWMutex для thread-safety
    - Использование container/list для LRU алгоритма
    - TTL проверка (24 часа по умолчанию)
    - Автоматическое вытеснение при переполнении
  - `internal/cache/lru_test.go` - comprehensive unit-тесты
    - Базовая функциональность (Set/Get/Update)
    - LRU вытеснение (переполнение, обновление порядка)
    - TTL истечение (удаление, актуальность, обновление timestamp)
    - Thread-safety (100 goroutines, конкурентный доступ)
    - Clear и makeKey
    - Бенчмарки (Set, Get, Concurrent)
  - Покрытие тестами: **100.0%** (превышает требование 90%)
  - 2 файла: 1 модуль + 1 тестовый файл

- **[2025-12-20] Этап 5: Конвертер валют (завершен)**
  - `internal/converter/validator.go` - валидация входных данных
    - ValidateAmount - проверка суммы (должна быть > 0)
    - ValidateDate - проверка даты (не в будущем)
    - Ошибки: ErrInvalidAmount, ErrDateInFuture
  - `internal/converter/formatter.go` - форматирование результата
    - FormatResult - создание строки вида "80 722,00 руб. ($1 000,00 по курсу 80,7220)"
    - formatNumber - форматирование числа с пробелами и запятой
    - addThousandsSeparator - добавление разделителей тысяч
  - `internal/converter/converter.go` - бизнес-логика конвертации
    - Интерфейсы: RateProvider, CacheStorage (для тестируемости)
    - Структура Converter с интеграцией Parser и Cache
    - Метод Convert - полный цикл конвертации с валидацией
    - Cache-first стратегия (сначала проверка кэша, затем запрос)
  - `internal/converter/converter_test.go` - comprehensive unit-тесты с моками
    - Моки: MockRateProvider, MockCacheStorage
    - Тесты валидации (amount, date, currency)
    - Тесты форматирования (разделители, запятая, валюты)
    - Тесты конвертации (success, cache hit/miss, errors)
    - Тесты ошибок (provider errors, currency not found)
    - Множественные конвертации (проверка кэширования)
  - Обновлена `internal/models/rate.go`:
    - Добавлено поле FormattedStr в ConversionResult
  - Обновлена `internal/models/currency.go`:
    - Добавлена ошибка ErrUnsupportedCurrency
    - Обновлен метод Validate() для использования обернутых ошибок
  - Покрытие тестами: **100.0%** (превышает требование 85%)
  - 4 файла: 3 модуля + 1 тестовый файл

- **[2025-12-21] Миграция парсера на XML API ЦБ РФ (завершена)**
  - `internal/parser/xml.go` - новый XML парсер
    - Структуры: ValCurs, Valute
    - ParseXML - парсинг XML ответа ЦБ РФ
    - parseXMLValue - парсинг значений с запятой
    - Поддержка кодировки windows-1251 (автоконвертация в UTF-8)
  - `internal/parser/client.go` - обновлен для XML API
    - Новый URL: `https://www.cbr.ru/scripts/XML_daily.asp`
    - Exponential backoff: 1s, 2s, 4s (как в Python v3.0.0)
    - Формат даты изменен: DD.MM.YYYY → DD/MM/YYYY (слэш вместо точки)
    - UserAgent обновлен: "CurRate-Go/2.0 (Windows; Go; XML)"
  - `internal/parser/cbr.go` - обновлен для использования ParseXML
  - `internal/parser/xml_test.go` - comprehensive unit-тесты
    - Тесты успешного парсинга
    - Тесты разных значений Nominal (1, 10, 100, 10000)
    - Тесты обработки запятой в значениях
    - Тесты невалидного XML
    - Тесты пропуска неподдерживаемых валют
  - `internal/parser/cbr_integration_test.go` - интеграционные тесты
    - Тесты с реальным XML API ЦБ РФ
    - Проверка формата ответа
    - Проверка работы с разными датами
  - Обновлены существующие тесты:
    - `internal/parser/cbr_test.go` - мок XML вместо HTML
    - `internal/parser/client_test.go` - новые URLs для XML API
  - Зависимости:
    - Добавлено: `golang.org/x/text/encoding/charmap` v0.32.0 (для windows-1251)
    - Deprecated: `github.com/PuerkitoBio/goquery` (оставлено для обратной совместимости)
  - Производительность:
    - ~5-10 мс вместо ~50-100 мс (улучшение в 10 раз)
    - Использование стандартной библиотеки encoding/xml
  - Документация:
    - `docs/07-XML-MIGRATION-PLAN.md` - подробный план миграции
  - Покрытие тестами: сохранено на уровне >90%

### Изменено (Changed)
- **[2025-12-20] Обновление зависимостей до актуальных стабильных версий:**
  - `github.com/PuerkitoBio/goquery`: v1.8.1 → **v1.11.0** (выпущена 16 ноября 2025)
  - `github.com/andybalholm/cascadia`: v1.3.1 → **v1.3.3** (косвенная зависимость от goquery)
  - `golang.org/x/net`: v0.7.0 → **v0.47.0** (косвенная зависимость от goquery)
  - `golang.org/x/sys`: v0.5.0 → **v0.39.0** (косвенная зависимость)
  - `github.com/lxn/walk`: v0.0.0-20210112085537-c389da54e794 (без изменений, последняя стабильная)
  - `github.com/atotto/clipboard`: v0.1.4 (без изменений, актуальная версия)
- Обновлена документация:
  - `docs/02-ТЕХНОЛОГИЧЕСКИЙ-СТЕК.md` - версия 1.1
  - `SUMMARY.md` - обновлены версии зависимостей
  - `README.md` - обновлена дата

---

## [2025-12-19] - Инициализация проекта

### Добавлено
- Структура проекта Go
- Git репозиторий
- Полный пакет документации (5 файлов):
  - `docs/01-ТЕХНИЧЕСКОЕ-ЗАДАНИЕ.md`
  - `docs/02-ТЕХНОЛОГИЧЕСКИЙ-СТЕК.md`
  - `docs/03-АРХИТЕКТУРНЫЙ-ДИЗАЙН.md`
  - `docs/04-ПЛАН-РАЗРАБОТКИ.md`
  - `docs/05-МИГРАЦИЯ-PYTHON-GO.md`
- Конфигурационные файлы:
  - `.gitignore`
  - `.golangci.yml`
  - `LICENSE` (MIT)
  - `go.mod`
- Папки для модулей:
  - `cmd/currate/` - точка входа
  - `internal/` - внутренние пакеты
  - `pkg/` - публичные пакеты

---

## Типы изменений

- `Добавлено` (Added) - новый функционал
- `Изменено` (Changed) - изменения в существующем функционале
- `Устарело` (Deprecated) - функционал, который скоро будет удален
- `Удалено` (Removed) - удаленный функционал
- `Исправлено` (Fixed) - исправления ошибок
- `Безопасность` (Security) - исправления уязвимостей
