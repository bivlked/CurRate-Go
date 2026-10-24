package converter

import (
	"fmt"
	"math/big"
	"strconv"
	"strings"

	"github.com/bivlked/currate-go/internal/models"
)

// FormatResult форматирует результат конвертации в читаемую строку
// Формат: "80 722,00 руб. ($1 000,00 по курсу 80,7220)"
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
	rateStr := strings.ReplaceAll(decimalRat(rate).FloatString(4), ".", ",")

	symbol := currency.Symbol()

	return fmt.Sprintf("%s руб. (%s%s по курсу %s)",
		resultStr, symbol, amountStr, rateStr)
}

// decimalRat возвращает точное десятичное значение x по его кратчайшему
// представлению: 80.7225 -> 807225/10000, а не двоичное 80.72249999999999659...
// Так округление идёт по тому числу, которое ввёл пользователь или прислал ЦБ.
// x должен быть конечным (NaN и Inf отсекаются валидацией раньше)
func decimalRat(x float64) *big.Rat {
	r, ok := new(big.Rat).SetString(strconv.FormatFloat(x, 'f', -1, 64))
	if !ok {
		panic(fmt.Sprintf("converter: non-finite value %v", x))
	}
	return r
}

// multiplyKopecks перемножает сумму и курс точно и округляет до копеек
// (половина копейки - вверх): 10 * 80.7225 = 807.225 -> 807.23.
// Умножение float64 дало бы 807.2249999... и потерю копейки
func multiplyKopecks(amount, rate float64) float64 {
	product := new(big.Rat).Mul(decimalRat(amount), decimalRat(rate))
	// FloatString(2) округляет половину от нуля; Float64 даёт ближайший
	// float64 (для слишком большого произведения - ±Inf)
	rounded, _ := new(big.Rat).SetString(product.FloatString(2))
	result, _ := rounded.Float64()
	return result
}

// SubtractKopecks возвращает разность двух сумм в рублях (a - b), посчитанную
// десятично и округлённую до копеек: 80538.10 - 79901.20 = 636.90 ровно,
// а float64 дал бы 636.9000000000087
func SubtractKopecks(a, b float64) float64 {
	diff := new(big.Rat).Sub(decimalRat(a), decimalRat(b))
	rounded, _ := new(big.Rat).SetString(diff.FloatString(2))
	result, _ := rounded.Float64()
	return result
}

// divideNominal делит курс за nominal единиц на nominal десятично и
// возвращает ближайший float64: его кратчайшая запись совпадает с точным
// частным (7.00025), поэтому decimalRat дальше восстановит его без потерь
func divideNominal(rate float64, nominal int) float64 {
	q, _ := new(big.Rat).Quo(decimalRat(rate), big.NewRat(int64(nominal), 1)).Float64()
	return q
}

// formatNumber форматирует число с разделителями тысяч (пробел) и запятой
// Округление до 2 знаков - половина вверх по десятичному значению
// Примеры:
//   - 1000.5 → "1 000,50"
//   - 80722.0 → "80 722,00"
//   - 123456789.12 → "123 456 789,12"
func formatNumber(num float64) string {
	// big.Rat.FloatString округляет половину от нуля: 1.005 -> "1.01"
	str := decimalRat(num).FloatString(2)

	// Разделение на целую и дробную части
	intPart, decPart, _ := strings.Cut(str, ".")

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
