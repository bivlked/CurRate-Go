// Package utils содержит вспомогательные утилиты
package utils

import (
	"fmt"
	"strconv"
	"strings"
)

// ParseAmount парсит строку с суммой в float64
// Поддерживает различные форматы:
// - "1000" - целое число
// - "1 000" - с пробелами как разделитель тысяч
// - "1,000.50" - американский формат (запятая - тысячи, точка - дробная часть)
// - "1.000,50" - европейский формат (точка - тысячи, запятая - дробная часть)
// - "1000.50" - простой формат с точкой
// - "1000,50" - простой формат с запятой
func ParseAmount(input string) (float64, error) {
	if input == "" {
		return 0, fmt.Errorf("пустая строка")
	}

	// Убираем пробелы
	cleaned := strings.ReplaceAll(input, " ", "")

	// Проверяем, есть ли и точка, и запятая
	hasDot := strings.Contains(cleaned, ".")
	hasComma := strings.Contains(cleaned, ",")

	var normalized string

	if hasDot && hasComma {
		// Оба разделителя присутствуют - определяем формат
		dotPos := strings.LastIndex(cleaned, ".")
		commaPos := strings.LastIndex(cleaned, ",")

		if dotPos > commaPos {
			// Американский формат: 1,000.50
			// Убираем запятые (разделители тысяч), оставляем точку
			normalized = strings.ReplaceAll(cleaned, ",", "")
		} else {
			// Европейский формат: 1.000,50
			// Убираем точки (разделители тысяч), заменяем запятую на точку
			normalized = strings.ReplaceAll(cleaned, ".", "")
			normalized = strings.ReplaceAll(normalized, ",", ".")
		}
	} else if hasComma {
		// Только запятая - может быть европейский формат или дробная часть
		commaCount := strings.Count(cleaned, ",")
		if commaCount == 1 {
			// Одна запятая - проверяем количество цифр после нее
			parts := strings.Split(cleaned, ",")
			if len(parts) == 2 && len(parts[1]) == 3 {
				// Ровно 3 цифры после запятой - это разделитель тысяч
				normalized = strings.ReplaceAll(cleaned, ",", "")
			} else {
				// Иначе это дробная часть
				normalized = strings.ReplaceAll(cleaned, ",", ".")
			}
		} else {
			// Несколько запятых - это разделители тысяч (не стандартно, но обработаем)
			normalized = strings.ReplaceAll(cleaned, ",", "")
		}
	} else if hasDot {
		// Только точка - американский формат или дробная часть
		dotCount := strings.Count(cleaned, ".")
		if dotCount == 1 {
			// Одна точка - проверяем количество цифр после нее
			parts := strings.Split(cleaned, ".")
			if len(parts) == 2 && len(parts[1]) == 3 {
				// Ровно 3 цифры после точки - это разделитель тысяч
				normalized = strings.ReplaceAll(cleaned, ".", "")
			} else {
				// Иначе это дробная часть - оставляем как есть
				normalized = cleaned
			}
		} else {
			// Несколько точек - это разделители тысяч
			normalized = strings.ReplaceAll(cleaned, ".", "")
		}
	} else {
		// Нет разделителей - целое число
		normalized = cleaned
	}

	// Парсим результат
	result, err := strconv.ParseFloat(normalized, 64)
	if err != nil {
		return 0, fmt.Errorf("не удалось распарсить число '%s': %w", input, err)
	}

	// Проверяем на отрицательные значения
	if result < 0 {
		return 0, fmt.Errorf("сумма не может быть отрицательной: %f", result)
	}

	return result, nil
}

// FormatAmount форматирует число в строку с разделителями тысяч
// Использует пробел как разделитель тысяч и точку как десятичный разделитель
func FormatAmount(amount float64, decimals int) string {
	// Форматируем с нужным количеством знаков после запятой
	formatted := fmt.Sprintf("%.*f", decimals, amount)

	// Разделяем на целую и дробную части
	parts := strings.Split(formatted, ".")
	intPart := parts[0]
	var decPart string
	if len(parts) > 1 {
		decPart = parts[1]
	}

	// Добавляем разделители тысяч к целой части
	var result strings.Builder
	for i, char := range intPart {
		if i > 0 && (len(intPart)-i)%3 == 0 {
			result.WriteRune(' ')
		}
		result.WriteRune(char)
	}

	// Добавляем дробную часть, если есть
	if decPart != "" {
		result.WriteRune('.')
		result.WriteString(decPart)
	}

	return result.String()
}
