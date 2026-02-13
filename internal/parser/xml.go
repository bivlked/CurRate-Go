// Package parser предоставляет функциональность для парсинга XML курсов валют с сайта ЦБ РФ
package parser

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"

	"github.com/bivlked/currate-go/internal/models"
	"golang.org/x/text/encoding/charmap"
)

// windows1251Regex используется для case-insensitive замены декларации кодировки
// Компилируется один раз при загрузке пакета для оптимизации производительности
var windows1251Regex = regexp.MustCompile(`(?i)windows-1251`)

// Ошибки XML парсинга
var (
	ErrInvalidXML     = errors.New("invalid XML structure")
	ErrNoXMLRates     = errors.New("no exchange rates found in XML")
	ErrInvalidXMLRate = errors.New("invalid rate format in XML")
	ErrXMLTooLarge    = errors.New("XML response exceeds size limit")
)

// maxXMLSize ограничивает максимальный размер XML ответа от ЦБ РФ (4 MB)
const maxXMLSize = 4 << 20

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
	Nominal  string `xml:"Nominal"` // Строка для обработки некорректных значений без падения всего парсинга
	Name     string `xml:"Name"`
	Value    string `xml:"Value"` // Строка, так как ЦБ использует запятую
}

// ParseXML парсит XML ответ ЦБ РФ и возвращает данные о курсах валют
// r - io.Reader с XML контентом (может быть в кодировке windows-1251)
// date - дата курсов (используется для установки даты в ExchangeRate)
func ParseXML(r io.Reader, date time.Time) (*models.RateData, error) {
	// Читаем весь XML в память с ограничением размера
	// Читаем maxXMLSize+1 байт: если прочитано больше лимита — явная ошибка
	r = io.LimitReader(r, maxXMLSize+1)
	xmlData, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read XML: %w", err)
	}
	if int64(len(xmlData)) > maxXMLSize {
		return nil, fmt.Errorf("%w: received %d bytes (limit %d)", ErrXMLTooLarge, len(xmlData), maxXMLSize)
	}

	// Если XML содержит декларацию windows-1251 (case-insensitive), конвертируем в UTF-8
	// Используем regex с флагом (?i) для case-insensitive проверки без копирования данных
	if windows1251Regex.Match(xmlData) {
		decoder := charmap.Windows1251.NewDecoder()
		xmlData, err = decoder.Bytes(xmlData)
		if err != nil {
			return nil, fmt.Errorf("failed to decode windows-1251: %w", err)
		}
		// Заменяем декларацию кодировки на UTF-8 (case-insensitive)
		// Используем regex для полностью регистронезависимой замены
		// Это обрабатывает все возможные варианты регистра (windows-1251, Windows-1251, WINDOWS-1251, WinDows-1251 и т.д.)
		xmlData = windows1251Regex.ReplaceAll(xmlData, []byte("UTF-8"))
	}

	// Декодируем XML в структуру
	var valCurs ValCurs
	if err := xml.Unmarshal(xmlData, &valCurs); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidXML, err)
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

	// Создаём RateData через NewRateData для единообразия
	rateData := models.NewRateData(parsedDate)

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

		// Парсим номинал с обработкой ошибок
		// Теперь valute.Nominal - string, что позволяет обработать некорректные значения
		// без падения всего парсинга XML
		nominal, err := parseNominal(valute.Nominal)
		if err != nil {
			// Пропускаем валюту с некорректным номиналом
			continue
		}

		// Добавляем курс через AddRate для единообразия
		exchangeRate := models.ExchangeRate{
			Currency: currency,
			Rate:     rate,
			Nominal:  nominal,
			Date:     parsedDate,
		}
		rateData.AddRate(exchangeRate)
	}

	if len(rateData.Rates) == 0 {
		return nil, ErrNoXMLRates
	}

	return rateData, nil
}

// parseXMLValue парсит строку значения из XML в формате "80,7220" (с запятой)
// и возвращает float64
var parseRateFunc = parseRate

func parseXMLValue(s string) (float64, error) {
	rate, err := parseRateFunc(s)
	if err == nil {
		return rate, nil
	}
	if errors.Is(err, ErrInvalidRate) {
		return 0, fmt.Errorf("%w: %s", ErrInvalidXMLRate, strings.TrimSpace(s))
	}

	return 0, err
}
