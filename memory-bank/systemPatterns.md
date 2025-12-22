# Memory Bank: System Patterns

## Architecture Pattern
**Model-View-Controller (MVC) with layered separation**

## Design Principles
- **Separation of Concerns** - Each module handles its own functionality
- **Dependency Injection** - Dependencies passed through constructors
- **Interface-based Design** - Interfaces for testability and flexibility
- **Error Handling** - Comprehensive error handling with clear messages

## Module Structure
```
internal/
├── models/         # Data models (Currency, ExchangeRate)
├── parser/         # XML API parsing (CBR)
│   ├── xml.go      # XML parser (encoding/xml + windows-1251)
│   ├── client.go   # HTTP client with retry
│   └── cbr.go      # Public API
├── converter/      # Business logic
│   ├── converter.go    # Main converter
│   ├── validator.go    # Input validation
│   └── formatter.go    # Result formatting
└── pkg/utils/      # Utilities (number parsing, formatting)
```

## Key Patterns
- **LRU Cache Pattern** - Caching with TTL for performance
- **Retry Pattern** - Exponential backoff for network resilience
- **Factory Pattern** - Constructor functions for object creation
- **Strategy Pattern** - Pluggable conversion strategies

## Error Handling Strategy
- Comprehensive error wrapping
- Clear error messages for users
- Logging for debugging
- Graceful degradation

## Testing Strategy
- Unit tests for all modules
- Integration tests with real API
- Benchmark tests for performance
- >90% code coverage target

