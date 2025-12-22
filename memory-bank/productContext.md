# Memory Bank: Product Context

## Product Description
**CurRate-Go** is a standalone Windows desktop application for currency conversion (USD, EUR) to rubles using official Central Bank of Russia exchange rates for a specified date.

## Target Users
- Windows 11 users
- Users needing quick currency conversion
- Users requiring official CBR exchange rates

## Key Features
- 🎨 **Intuitive Interface** - Simple and clear design in one window
- 📅 **Visual Calendar** - Date selection with automatic weekend highlighting (red)
- ⌨️ **Manual Date Input** - Alternative to calendar for quick input
- 💱 **Currency Selection** - Radio buttons for USD/EUR
- 📋 **Copy to Clipboard** - One-click result copying
- ⚡ **Instant Results** - Thanks to LRU cache
- 💾 **Compact Size** - Only ~8-10 MB (with UPX compression)

## User Workflow
1. Select date (via calendar or manual input)
2. Choose currency (USD or EUR)
3. Enter amount
4. Click "Convert"
5. View result: `"80 722,00 руб. ($1 000,00 по курсу 80,7220)"`
6. Copy result to clipboard if needed

## Success Criteria
- Standalone executable (no Python/runtime dependencies)
- One-click launch
- Compact size (5-10 MB)
- Native Windows interface
- Fast performance (<10ms per request)

