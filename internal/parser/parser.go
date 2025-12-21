// Package parser предоставляет функциональность для парсинга курсов валют с сайта ЦБ РФ
package parser

import (
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/bivlked/currate-go/internal/models"
)

// Ошибки парсинга
var (
	ErrInvalidHTML         = errors.New("invalid HTML structure")
	ErrNoRates             = errors.New("no exchange rates found")
	ErrInvalidRate         = errors.New("invalid rate format")
	ErrInvalidNominal      = errors.New("invalid nominal format")
	ErrUnsupportedCurrency = errors.New("unsupported currency code")
)

// ParseHTML парсит HTML страницу ЦБ РФ и возвращает данные о курсах валют
// r - io.Reader с HTML контентом
// date - дата курсов (передается, так как не всегда можно надежно извлечь из HTML)
func ParseHTML(r io.Reader, date time.Time) (*models.RateData, error) {
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	rates := make(map[models.Currency]models.ExchangeRate)

	// Ищем таблицу с классом "data"
	// Структура: <table class="data"><tbody><tr><td>...</td></tr></tbody></table>
	found := false
	doc.Find("table.data tbody tr").Each(func(i int, row *goquery.Selection) {
		// Пропускаем заголовок (первая строка с <th>)
		if row.Find("th").Length() > 0 {
			return // это заголовок таблицы
		}

		cells := row.Find("td")
		if cells.Length() < 5 {
			return // недостаточно ячеек, пропускаем
		}

		// Извлекаем данные из ячеек:
		// 0: Цифр. код
		// 1: Букв. код (например, "USD")
		// 2: Единиц (номинал, например, "1" или "100")
		// 3: Валюта (название)
		// 4: Курс (например, "80,7220")

		currencyCode := strings.TrimSpace(cells.Eq(1).Text())
		nominalStr := strings.TrimSpace(cells.Eq(2).Text())
		rateStr := strings.TrimSpace(cells.Eq(4).Text())

		// Парсим данные
		currency, err := parseCurrency(currencyCode)
		if err != nil {
			// Пропускаем неподдерживаемые валюты (например, SDR)
			return
		}

		nominal, err := parseNominal(nominalStr)
		if err != nil {
			return // пропускаем строки с некорректным номиналом
		}

		rate, err := parseRate(rateStr)
		if err != nil {
			return // пропускаем строки с некорректным курсом
		}

		// Добавляем в результат
		rates[currency] = models.ExchangeRate{
			Currency: currency,
			Rate:     rate,
			Nominal:  nominal,
			Date:     date,
		}
		found = true
	})

	if !found || len(rates) == 0 {
		return nil, ErrNoRates
	}

	return &models.RateData{
		Date:  date,
		Rates: rates,
	}, nil
}

// parseRate парсит строку курса в формате "80,7220" (с запятой как десятичным разделителем)
// и возвращает float64
func parseRate(s string) (float64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, ErrInvalidRate
	}

	// Заменяем запятую на точку (ЦБ РФ использует запятую как десятичный разделитель)
	s = strings.ReplaceAll(s, ",", ".")

	rate, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, fmt.Errorf("%w: %s", ErrInvalidRate, s)
	}

	if rate <= 0 {
		return 0, fmt.Errorf("%w: rate must be positive", ErrInvalidRate)
	}

	return rate, nil
}

// parseNominal парсит строку номинала в формат int
// Примеры: "1", "10", "100", "1000"
func parseNominal(s string) (int, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, ErrInvalidNominal
	}

	// Удаляем возможные пробелы в числе (например, "10 000")
	s = strings.ReplaceAll(s, " ", "")

	nominal, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("%w: %s", ErrInvalidNominal, s)
	}

	if nominal <= 0 {
		return 0, fmt.Errorf("%w: nominal must be positive", ErrInvalidNominal)
	}

	return nominal, nil
}

// parseCurrency конвертирует трехбуквенный код валюты в тип Currency
// Возвращает ошибку для неподдерживаемых валют
func parseCurrency(code string) (models.Currency, error) {
	code = strings.ToUpper(strings.TrimSpace(code))

	currency := models.Currency(code)
	if err := currency.Validate(); err != nil {
		return "", fmt.Errorf("%w: %s", ErrUnsupportedCurrency, code)
	}

	return currency, nil
}
