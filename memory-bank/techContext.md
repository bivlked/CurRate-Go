# Memory Bank: Technical Context

## Technology Stack

### Core Technologies
- **Go 1.21+** - Programming language
- **encoding/xml** - XML parsing (standard library)
- **golang.org/x/text** - Windows-1251 encoding support

### GUI Framework (In Development)
- **Wails v2.11.0** - Desktop UI framework
- **WebView2** - Native rendering (Windows 11)
- **Vanilla JS** - Frontend without frameworks

### Utilities
- **atotto/clipboard** - Clipboard operations

## Development Environment
- **OS**: Windows 10/11
- **Shell**: PowerShell
- **Build Tool**: Go build
- **Testing**: go test

## Dependencies
```go
module github.com/bivlked/currate-go

go 1.25.5

require golang.org/x/text v0.32.0
```

## Performance Characteristics
- **Request Time**: ~5-10ms (XML API)
- **Cache Hit Rate**: High (LRU with 24h TTL)
- **Memory Usage**: Low (efficient Go data structures)
- **Binary Size**: ~8-10 MB (with UPX compression)

## API Integration
- **CBR XML API**: https://www.cbr.ru/scripts/XML_daily.asp
- **Encoding**: Windows-1251 → UTF-8 conversion
- **Retry Logic**: Exponential backoff (1s, 2s, 4s)
- **Timeout**: Configurable HTTP timeout

## Build Configuration
- **Target OS**: Windows
- **Architecture**: AMD64
- **Compression**: UPX (optional)
- **Output**: Standalone .exe file

## Testing Infrastructure
- **Unit Tests**: go test
- **Integration Tests**: go test -tags=integration
- **Benchmarks**: go test -bench=.
- **Coverage**: go test -coverprofile=coverage.out

