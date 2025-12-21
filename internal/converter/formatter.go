package converter

import (
	"fmt"
	"strings"

	"github.com/bivlked/currate-go/internal/models"
)

// FormatResult форматирует результат конвертации в читаемую строку
// Формат: "80 722,00 руб. ($1000.00 по курсу 80,7220)"
//
// Пример использования:
//
//	formatted := FormatResult(1000.0, 80.7220, models.USD, 80722.0)
//	// Результат: "80 722,00 руб. ($1 000,00 по курсу 80,7220)"
func FormatResult(amount, rate float64, currency models.Currency, resultRUB float64) string {
	// Форматирование с разделителями тысяч и запятой
	resultStr := formatNumber(resultRUB)
	amountStr := formatNumber(amount)

	// Форматируем курс: 4 знака после запятой
	rateStr := fmt.Sprintf("%.4f", rate)
	rateStr = strings.ReplaceAll(rateStr, ".", ",")

	symbol := currency.Symbol()

	return fmt.Sprintf("%s руб. (%s%s по курсу %s)",
		resultStr, symbol, amountStr, rateStr)
}

// formatNumber форматирует число с разделителями тысяч (пробел) и запятой
// Примеры:
//   - 1000.5 → "1 000,50"
//   - 80722.0 → "80 722,00"
//   - 123456789.12 → "123 456 789,12"
func formatNumber(num float64) string {
	// Форматируем с 2 знаками после запятой
	str := fmt.Sprintf("%.2f", num)

	// Заменяем точку на запятую
	str = strings.ReplaceAll(str, ".", ",")

	// Разделение на целую и дробную части
	parts := strings.Split(str, ",")
	intPart := parts[0]
	decPart := parts[1]

	// Добавление разделителей тысяч
	intPart = addThousandsSeparator(intPart)

	return intPart + "," + decPart
}

// addThousandsSeparator добавляет пробелы как разделители тысяч
// Примеры:
//   - "1000" → "1 000"
//   - "80722" → "80 722"
//   - "123456789" → "123 456 789"
//   - "500" → "500"
func addThousandsSeparator(s string) string {
	if len(s) <= 3 {
		return s
	}

	var result strings.Builder

	for i, c := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			result.WriteString(" ")
		}
		result.WriteRune(c)
	}

	return result.String()
}
