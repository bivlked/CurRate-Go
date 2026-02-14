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
 * @param {number} cursorPos - Позиция курсора
 * @returns {boolean} true если ввод допустим
 */
function isValidDateChar(currentValue, newChar, cursorPos) {
    // Разрешаем только цифры
    if (!/^\d$/.test(newChar)) {
        return false;
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
            // Американский формат: 1,000.50 или 123,456,78.9
            // Точка после последней запятой - это дробная часть
            // Убираем все запятые (разделители тысяч), оставляем точку
            normalized = cleaned.replace(/,/g, '');
        } else {
            // Европейский формат: 1.000,50
            // Запятая после последней точки - это дробная часть
            // Убираем точки (разделители тысяч), заменяем запятую на точку
            normalized = cleaned.replace(/\./g, '').replace(',', '.');
        }
    } else if (hasComma) {
        // Только запятая
        const commaCount = (cleaned.match(/,/g) || []).length;
        if (commaCount === 1) {
            const parts = cleaned.split(',');
            if (parts[1].length === 3 && parts[0].length >= 1) {
                // Ровно 3 цифры после запятой — тысячный разделитель (1,000 → 1000)
                normalized = cleaned.replace(',', '');
            } else {
                // Иначе — десятичный разделитель (1,5 → 1.5)
                normalized = cleaned.replace(',', '.');
            }
        } else {
            // Несколько запятых — разделители тысяч
            normalized = cleaned.replace(/,/g, '');
        }
    } else if (hasDot) {
        // Только точка
        const dotCount = (cleaned.match(/\./g) || []).length;
        if (dotCount === 1) {
            const parts = cleaned.split('.');
            if (parts[1].length === 3 && parts[0].length >= 1) {
                // Ровно 3 цифры после точки — тысячный разделитель (1.000 → 1000)
                normalized = cleaned.replace('.', '');
            } else {
                // Иначе — десятичный разделитель (1.5 → 1.5)
                normalized = cleaned;
            }
        } else {
            // Несколько точек — разделители тысяч
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
 * @returns {Promise<void>}
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
            document.execCommand('copy');
            document.body.removeChild(textArea);
            return true;
        } catch (err) {
            document.body.removeChild(textArea);
            return false;
        }
    }
}

