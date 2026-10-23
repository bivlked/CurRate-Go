/**
 * Утилиты для CurRate-Go
 */

/**
 * Debounce функция для оптимизации вызовов
 * @param {Function} func - Функция для debounce
 * @param {number} wait - Время ожидания в миллисекундах
 * @returns {Function} Debounced функция
 */
function debounce(func, wait) {
    let timeout;
    return function executedFunction(...args) {
        const later = () => {
            clearTimeout(timeout);
            func(...args);
        };
        clearTimeout(timeout);
        timeout = setTimeout(later, wait);
    };
}

/**
 * Форматирует дату в формат DD.MM.YYYY
 * @param {Date} date - Дата для форматирования
 * @returns {string} Отформатированная дата
 */
function formatDate(date) {
    if (!(date instanceof Date) || isNaN(date)) {
        return '';
    }
    const day = String(date.getDate()).padStart(2, '0');
    const month = String(date.getMonth() + 1).padStart(2, '0');
    const year = date.getFullYear();
    return `${day}.${month}.${year}`;
}

/**
 * Парсит дату из формата DD.MM.YYYY
 * @param {string} dateStr - Строка с датой
 * @returns {Date|null} Дата или null при ошибке
 */
function parseDate(dateStr) {
    if (!dateStr || dateStr.length !== 10) {
        return null;
    }
    const parts = dateStr.split('.');
    if (parts.length !== 3) {
        return null;
    }
    const day = parseInt(parts[0], 10);
    const month = parseInt(parts[1], 10) - 1; // Месяц в JS начинается с 0
    const year = parseInt(parts[2], 10);
    
    if (isNaN(day) || isNaN(month) || isNaN(year)) {
        return null;
    }
    
    const date = new Date(year, month, day);
    // Проверка корректности даты
    if (date.getDate() !== day || date.getMonth() !== month || date.getFullYear() !== year) {
        return null;
    }
    
    return date;
}

/**
 * Валидирует формат даты DD.MM.YYYY
 * @param {string} dateStr - Строка с датой
 * @returns {boolean} true если формат корректный
 */
function isValidDateFormat(dateStr) {
    if (!dateStr) return false;
    const regex = /^\d{2}\.\d{2}\.\d{4}$/;
    if (!regex.test(dateStr)) return false;
    return parseDate(dateStr) !== null;
}

/**
 * Автоматически форматирует ввод даты (23042025 -> 23.04.2025)
 * @param {string} value - Введенное значение
 * @returns {string} Отформатированная дата
 */
function autoFormatDate(value) {
    // Убираем все нецифровые символы
    const digits = value.replace(/\D/g, '');
    
    if (digits.length === 0) return '';
    
    // Ограничиваем до 8 цифр (ДДММГГГГ)
    const limited = digits.slice(0, 8);
    
    // Форматируем: ДД.ММ.ГГГГ
    if (limited.length <= 2) {
        return limited;
    } else if (limited.length <= 4) {
        return `${limited.slice(0, 2)}.${limited.slice(2)}`;
    } else {
        return `${limited.slice(0, 2)}.${limited.slice(2, 4)}.${limited.slice(4)}`;
    }
}

/**
 * Валидирует ввод цифры на определенной позиции в дате
 * @param {string} currentValue - Текущее значение поля
 * @param {string} newChar - Новая вводимая цифра
 * @param {number} cursorPos - Позиция курсора (начало выделения)
 * @param {number} [selectionEnd] - Конец выделения; выделенный текст заменяется вводом
 * @returns {boolean} true если ввод допустим
 */
function isValidDateChar(currentValue, newChar, cursorPos, selectionEnd = cursorPos) {
    // Разрешаем только цифры
    if (!/^\d$/.test(newChar)) {
        return false;
    }

    // Выделенный фрагмент будет заменён: проверяем значение без него.
    // Иначе дату, выделенную при фокусе, нельзя перепечатать поверх
    if (selectionEnd > cursorPos) {
        currentValue = currentValue.slice(0, cursorPos) + currentValue.slice(selectionEnd);
    }
    
    // Убираем все нецифровые символы для анализа
    const digits = currentValue.replace(/\D/g, '');
    
    // Если уже 8 цифр, не разрешаем ввод
    if (digits.length >= 8) {
        return false;
    }
    
    // Определяем позицию в формате ДДММГГГГ
    let posInDigits = 0;
    let tempPos = 0;
    for (let i = 0; i < currentValue.length && tempPos < cursorPos; i++) {
        if (/\d/.test(currentValue[i])) {
            posInDigits++;
        }
        tempPos++;
    }
    
    // Валидация по позиции
    if (posInDigits === 0) {
        // Первая цифра дня: 0-3
        return newChar >= '0' && newChar <= '3';
    } else if (posInDigits === 1) {
        // Вторая цифра дня: зависит от первой
        const firstDay = digits[0] || '0';
        if (firstDay === '3') {
            return newChar >= '0' && newChar <= '1'; // 30-31
        }
        return true; // 01-29
    } else if (posInDigits === 2) {
        // Первая цифра месяца: 0-1
        return newChar >= '0' && newChar <= '1';
    } else if (posInDigits === 3) {
        // Вторая цифра месяца: зависит от первой
        const firstMonth = digits[2] || '0';
        if (firstMonth === '1') {
            return newChar >= '0' && newChar <= '2'; // 10-12
        }
        return true; // 01-09
    } else if (posInDigits === 4) {
        // Первая цифра года: 1-2 (для 19xx и 20xx)
        return newChar >= '1' && newChar <= '2';
    } else if (posInDigits === 5) {
        // Вторая цифра года: зависит от первой
        const firstYear = digits[4] || '1';
        if (firstYear === '1') {
            return newChar >= '9' && newChar <= '9'; // 19xx
        } else if (firstYear === '2') {
            return newChar >= '0' && newChar <= '0'; // 20xx
        }
        return true;
    }
    
    // Остальные цифры года: любые
    return true;
}

/**
 * Форматирует число с разделителями тысяч
 * @param {number} num - Число для форматирования
 * @param {number} decimals - Количество знаков после запятой
 * @returns {string} Отформатированное число
 */
function formatNumber(num, decimals = 2) {
    if (isNaN(num)) return '0';
    return num.toLocaleString('ru-RU', {
        minimumFractionDigits: decimals,
        maximumFractionDigits: decimals
    });
}

/**
 * Парсит сумму из строки (поддерживает различные форматы как в Go версии)
 * Поддерживает:
 * - "1000" - целое число
 * - "1 000" - с пробелами как разделитель тысяч
 * - "1,000.50" - американский формат (запятая - тысячи, точка - дробная часть)
 * - "1.000,50" - европейский формат (точка - тысячи, запятая - дробная часть)
 * - "1000.50" - простой формат с точкой
 * - "1000,50" - простой формат с запятой
 * @param {string} str - Строка с числом
 * @returns {number|null} Число или null при ошибке
 */
function parseAmount(str) {
    if (!str || str.trim() === '') return null;
    
    // Убираем пробелы
    let cleaned = str.trim().replace(/\s/g, '');
    
    if (cleaned === '') return null;
    
    // Проверяем, есть ли и точка, и запятая
    const hasDot = cleaned.includes('.');
    const hasComma = cleaned.includes(',');
    
    let normalized;
    
    if (hasDot && hasComma) {
        // Оба разделителя присутствуют - определяем формат
        const dotPos = cleaned.lastIndexOf('.');
        const commaPos = cleaned.lastIndexOf(',');
        
        if (dotPos > commaPos) {
            // Американский формат: 1,000.50 или 123,456,789.9
            // Точка после последней запятой - это дробная часть
            // Убираем все запятые (разделители тысяч), оставляем точку
            normalized = cleaned.replace(/,/g, '');
        } else {
            // Европейский формат: 1.000,50
            // Запятая после последней точки - это дробная часть
            // Убираем точки (разделители тысяч), заменяем запятую на точку
            normalized = cleaned.replace(/\./g, '').replace(',', '.');
        }
        // Группировка тысяч в целой части должна быть корректной:
        // "1,000.50" допустимо, "12,34.5" и "1..000,5" - нет
        const thousandsSep = dotPos > commaPos ? ',' : '.';
        const intPart = cleaned.slice(0, Math.max(dotPos, commaPos));
        if (!isValidGrouping(intPart, thousandsSep)) {
            return null;
        }
    } else if (hasComma) {
        // Только запятая
        const commaCount = (cleaned.match(/,/g) || []).length;
        if (commaCount === 1) {
            const parts = cleaned.split(',');
            // Тысячный разделитель: ровно 3 цифры после запятой И целая часть
            // из 1-3 цифр без ведущего нуля (1,000 → 1000). Ведущий ноль означает
            // десятичную дробь: 0,500 → 0.5, а не 500; 1234,567 - тоже дробь
            if (parts[1].length === 3 && /^[1-9]\d{0,2}$/.test(parts[0])) {
                normalized = cleaned.replace(',', '');
            } else {
                // Иначе — десятичный разделитель (1,5 → 1.5)
                normalized = cleaned.replace(',', '.');
            }
        } else {
            // Несколько запятых - разделители тысяч, но только в виде 1,234,567
            if (!isValidGrouping(cleaned, ',')) {
                return null;
            }
            normalized = cleaned.replace(/,/g, '');
        }
    } else if (hasDot) {
        // Только точка
        const dotCount = (cleaned.match(/\./g) || []).length;
        if (dotCount === 1) {
            const parts = cleaned.split('.');
            // Тысячный разделитель: ровно 3 цифры после точки И целая часть
            // из 1-3 цифр без ведущего нуля (1.000 → 1000). Ведущий ноль означает
            // десятичную дробь: 0.500 → 0.5, а не 500; 1234.567 - тоже дробь
            if (parts[1].length === 3 && /^[1-9]\d{0,2}$/.test(parts[0])) {
                normalized = cleaned.replace('.', '');
            } else {
                // Иначе — десятичный разделитель (1.5 → 1.5)
                normalized = cleaned;
            }
        } else {
            // Несколько точек - разделители тысяч, но только в виде 1.234.567
            if (!isValidGrouping(cleaned, '.')) {
                return null;
            }
            normalized = cleaned.replace(/\./g, '');
        }
    } else {
        // Нет разделителей - целое число
        normalized = cleaned;
    }
    
    // Проверяем, что нормализованная строка содержит только допустимые символы:
    // цифры, одна точка (дробная часть), опциональный ведущий минус
    if (!/^-?\d+(\.\d+)?$/.test(normalized)) {
        return null;
    }

    // Парсим результат
    const num = parseFloat(normalized);

    // Отклоняем NaN, Infinity, 0 и отрицательные числа (согласуется с бэкендом: amount > 0)
    if (!isFinite(num) || num <= 0) {
        return null;
    }

    return num;
}

/**
 * Проверяет группировку тысяч: 1-3 цифры, затем группы ровно по 3 цифры
 * ("1,234,567"). Строка без разделителя тоже допустима ("1234").
 * Отсекает опечатки вида "1..2" или "12,34,56", которые иначе молча
 * превратились бы в 12 и 123456
 * @param {string} intPart - Целая часть числа
 * @param {string} sep - Разделитель тысяч ('.' или ',')
 * @returns {boolean} true если группировка корректна
 */
function isValidGrouping(intPart, sep) {
    if (!intPart.includes(sep)) {
        return /^-?\d+$/.test(intPart);
    }
    const grouped = sep === '.' ? /^-?\d{1,3}(?:\.\d{3})+$/ : /^-?\d{1,3}(?:,\d{3})+$/;
    return grouped.test(intPart);
}

/**
 * Проверяет, является ли дата выходным днем (суббота или воскресенье)
 * @param {Date} date - Дата для проверки
 * @returns {boolean} true если выходной
 */
function isWeekend(date) {
    if (!(date instanceof Date) || isNaN(date)) {
        return false;
    }
    const day = date.getDay();
    return day === 0 || day === 6; // 0 = воскресенье, 6 = суббота
}

/**
 * Проверяет, является ли дата будущей
 * @param {Date} date - Дата для проверки
 * @returns {boolean} true если дата в будущем
 */
function isFutureDate(date) {
    if (!(date instanceof Date) || isNaN(date)) {
        return false;
    }
    const today = new Date();
    today.setHours(0, 0, 0, 0);
    const checkDate = new Date(date);
    checkDate.setHours(0, 0, 0, 0);
    return checkDate > today;
}

/**
 * Получает первый день месяца
 * @param {Date} date - Дата
 * @returns {Date} Первый день месяца
 */
function getFirstDayOfMonth(date) {
    return new Date(date.getFullYear(), date.getMonth(), 1);
}

/**
 * Получает последний день месяца
 * @param {Date} date - Дата
 * @returns {Date} Последний день месяца
 */
function getLastDayOfMonth(date) {
    return new Date(date.getFullYear(), date.getMonth() + 1, 0);
}

/**
 * Добавляет месяцы к дате
 * @param {Date} date - Исходная дата
 * @param {number} months - Количество месяцев (может быть отрицательным)
 * @returns {Date} Новая дата
 */
function addMonths(date, months) {
    const result = new Date(date);
    result.setMonth(result.getMonth() + months);
    return result;
}

/**
 * Копирует текст в буфер обмена
 * @param {string} text - Текст для копирования
 * @returns {Promise<boolean>} true при успешном копировании
 */
async function copyToClipboard(text) {
    try {
        await navigator.clipboard.writeText(text);
        return true;
    } catch (err) {
        // Fallback для старых браузеров
        const textArea = document.createElement('textarea');
        textArea.value = text;
        textArea.style.position = 'fixed';
        textArea.style.opacity = '0';
        document.body.appendChild(textArea);
        textArea.select();
        try {
            // execCommand сообщает о неудаче возвратом false, а не исключением
            return document.execCommand('copy');
        } catch (err) {
            return false;
        } finally {
            document.body.removeChild(textArea);
        }
    }
}


/**
 * Сравнивает курс с референсным. Курсы ЦБ для USD и EUR публикуются ровно
 * с 4 знаками, поэтому сравнение идёт в целых единицах 0,0001 ₽: разность
 * точная, без погрешности float (80,5381 - 79,9012 = 0,6369, а не 0,636899...)
 * @param {number} rate - Курс на выбранную дату
 * @param {number} refRate - Курс на референсную дату
 * @returns {{direction: 'up'|'down'|'equal', diff: number, percent: number}|null}
 *          diff и percent - модули; null, если сравнивать нечего
 */
function compareRates(rate, refRate) {
    if (!isFinite(rate) || !isFinite(refRate) || rate <= 0 || refRate <= 0) {
        return null;
    }
    const units = Math.round(rate * 1e4);
    const refUnits = Math.round(refRate * 1e4);
    const diffUnits = units - refUnits;
    let direction = 'equal';
    if (diffUnits > 0) direction = 'up';
    if (diffUnits < 0) direction = 'down';
    return {
        direction,
        diff: Math.abs(diffUnits) / 1e4,
        percent: Math.abs(diffUnits) / refUnits * 100
    };
}

/**
 * Форматирует процент изменения: 2 знака, а совсем малое - «менее 0,01%»,
 * чтобы ненулевая разница не выглядела как 0,00%
 * @param {number} percent - Неотрицательный процент
 * @returns {string}
 */
function formatPercent(percent) {
    if (percent > 0 && percent < 0.005) {
        return 'менее 0,01%';
    }
    return `${formatNumber(percent, 2)}%`;
}

/**
 * Текст и стиль строки сравнения курса с референсом
 * @param {ReturnType<typeof compareRates>} cmp - Результат compareRates
 * @returns {{className: string, text: string}}
 */
function describeRateComparison(cmp) {
    if (!cmp) {
        return { className: 'compare-none', text: 'Сравнение недоступно' };
    }
    if (cmp.direction === 'equal') {
        return { className: 'compare-equal', text: 'Совпадает с референсом' };
    }
    const where = cmp.direction === 'up' ? 'Выше' : 'Ниже';
    return {
        className: `compare-${cmp.direction}`,
        text: `${where} референса на ${formatNumber(cmp.diff, 4)} ₽ (${formatPercent(cmp.percent)})`
    };
}

/**
 * Текст и стиль строки результата: сумма по курсу на дату против суммы
 * по курсу на референсную дату. Разность приходит из Go (точная до копейки).
 * Текст в одну строку ширины окна; дата и курс референса - в подсказке
 * @param {number} diffRUB - Сумма по курсу минус сумма по референсу, ₽
 * @param {number} refAmountRUB - Сумма по референсному курсу, ₽
 * @returns {{className: string, text: string}}
 */
function describeAmountComparison(diffRUB, refAmountRUB) {
    const refAmount = `${formatNumber(refAmountRUB, 2)} ₽`;
    if (diffRUB === 0) {
        return { className: 'compare-equal', text: `Совпадает с суммой по референсу (${refAmount})` };
    }
    const more = diffRUB > 0;
    return {
        className: more ? 'compare-up' : 'compare-down',
        text: `На ${formatNumber(Math.abs(diffRUB), 2)} ₽ ${more ? 'больше' : 'меньше'} референса (${refAmount})`
    };
}
