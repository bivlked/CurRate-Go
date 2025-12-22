// Package parser предоставляет функциональность для парсинга XML курсов валют с сайта ЦБ РФ
package parser

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/bivlked/currate-go/internal/models"
	"golang.org/x/text/encoding/charmap"
)

// Ошибки XML парсинга
var (
	ErrInvalidXML     = errors.New("invalid XML structure")
	ErrNoXMLRates     = errors.New("no exchange rates found in XML")
	ErrInvalidXMLRate = errors.New("invalid rate format in XML")
)

// ValCurs представляет корневой элемент XML ответа ЦБ РФ
// Пример: <ValCurs Date="20.12.2025" name="Foreign Currency Market">
type ValCurs struct {
	XMLName xml.Name `xml:"ValCurs"`
	Date    string   `xml:"Date,attr"`
	Name    string   `xml:"name,attr"`
	Valutes []Valute `xml:"Valute"`
}

// Valute представляет информацию об одной валюте в XML ответе
// Пример:
//
//	<Valute ID="R01235">
//	    <NumCode>840</NumCode>
//	    <CharCode>USD</CharCode>
//	    <Nominal>1</Nominal>
//	    <Name>Доллар США</Name>
//	    <Value>80,7220</Value>
//	</Valute>
type Valute struct {
	ID       string `xml:"ID,attr"`
	NumCode  string `xml:"NumCode"`
	CharCode string `xml:"CharCode"`
	Nominal  int    `xml:"Nominal"`
	Name     string `xml:"Name"`
	Value    string `xml:"Value"` // Строка, так как ЦБ использует запятую
}

// ParseXML парсит XML ответ ЦБ РФ и возвращает данные о курсах валют
// r - io.Reader с XML контентом (может быть в кодировке windows-1251)
// date - дата курсов (используется для установки даты в ExchangeRate)
func ParseXML(r io.Reader, date time.Time) (*models.RateData, error) {
	// Читаем весь XML в память
	xmlData, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read XML: %w", err)
	}

	// Если XML содержит декларацию windows-1251, конвертируем в UTF-8
	if bytes.Contains(xmlData, []byte("windows-1251")) {
		decoder := charmap.Windows1251.NewDecoder()
		xmlData, err = decoder.Bytes(xmlData)
		if err != nil {
			return nil, fmt.Errorf("failed to decode windows-1251: %w", err)
		}
		// Заменяем декларацию кодировки на UTF-8
		xmlData = bytes.Replace(xmlData, []byte("windows-1251"), []byte("UTF-8"), 1)
	}

	// Декодируем XML в структуру
	var valCurs ValCurs
	if err := xml.Unmarshal(xmlData, &valCurs); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidXML, err)
	}

	// Проверяем, что есть данные о валютах
	if len(valCurs.Valutes) == 0 {
		return nil, ErrNoXMLRates
	}

	// Определяем дату курсов из XML (ЦБ может вернуть прошлый рабочий день)
	parsedDate := date
	if valCurs.Date != "" {
		if parsed, err := time.ParseInLocation("02.01.2006", valCurs.Date, date.Location()); err == nil {
			parsedDate = parsed
		}
	}

	// Конвертируем в нашу структуру данных
	rates := make(map[models.Currency]models.ExchangeRate)

	for _, valute := range valCurs.Valutes {
		// Парсим код валюты
		currency, err := parseCurrency(valute.CharCode)
		if err != nil {
			// Пропускаем неподдерживаемые валюты (например, SDR)
			continue
		}

		// Парсим значение курса (с запятой)
		rate, err := parseXMLValue(valute.Value)
		if err != nil {
			// Пропускаем валюты с некорректным значением
			continue
		}

		// Валидируем номинал
		nominal, err := parseNominal(strconv.Itoa(valute.Nominal))
		if err != nil {
			continue
		}

		// Сохраняем результат
		rates[currency] = models.ExchangeRate{
			Currency: currency,
			Rate:     rate,
			Nominal:  nominal,
			Date:     parsedDate,
		}
	}

	if len(rates) == 0 {
		return nil, ErrNoXMLRates
	}

	return &models.RateData{
		Date:  parsedDate,
		Rates: rates,
	}, nil
}

// parseXMLValue парсит строку значения из XML в формате "80,7220" (с запятой)
// и возвращает float64
func parseXMLValue(s string) (float64, error) {
	rate, err := parseRate(s)
	if err == nil {
		return rate, nil
	}
	if errors.Is(err, ErrInvalidRate) {
		return 0, fmt.Errorf("%w: %s", ErrInvalidXMLRate, strings.TrimSpace(s))
	}

	return 0, err
}
