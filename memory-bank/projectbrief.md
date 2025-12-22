# Memory Bank: Project Brief

## Project Overview
**CurRate-Go** - High-performance currency converter written in Go that fetches actual exchange rates from the official Central Bank of Russia XML API.

## Key Characteristics
- **Language**: Go 1.21+
- **Performance**: ~5-10ms per request (10x faster than HTML parsing)
- **Reliability**: Official XML API instead of HTML scraping
- **Caching**: LRU cache with TTL to minimize requests
- **Retry Logic**: Exponential backoff (1s, 2s, 4s) for fault tolerance
- **Encoding Support**: Automatic conversion windows-1251 → UTF-8
- **Test Coverage**: 100% unit + integration tests
- **UI**: Desktop application on Wails v2 (in development)

## Architecture
- **GUI Layer**: Wails v2 (HTML/CSS/JS → Go Backend)
- **Business Logic**: Converter, Validator, Formatter
- **Data Access**: XML Parser, HTTP Client
- **Caching**: LRU Cache (100 items, 24h TTL, thread-safe)

## Current Status
- ✅ Backend completed (models, parser, converter, cache)
- ✅ Test coverage >90% (target: >=95%)
- ✅ GUI Design completed (CREATIVE phase)
- 🚧 GUI Implementation in progress (planned)

## Development Standards
- **Test Coverage:** >= 95% (обязательно)
- **Documentation Language:** Русский (основной), английский (где уместно)
- **Dependency Check:** Context7 перед использованием
- **Git Sync:** Локальное состояние = GitHub состояние
- **Quality:** Полная, понятная, красивая, информативная документация без "воды"

## Technology Stack
- **Core**: Go 1.21+, encoding/xml, golang.org/x/text
- **GUI**: Wails v2.11.0, WebView2, Vanilla JS
- **Utilities**: atotto/clipboard

