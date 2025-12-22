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
 * Парсит сумму из строки (поддерживает запятую и точку как разделитель)
 * @param {string} str - Строка с числом
 * @returns {number|null} Число или null при ошибке
 */
function parseAmount(str) {
    if (!str || str.trim() === '') return null;
    
    // Заменяем запятую на точку для парсинга
    const normalized = str.trim().replace(',', '.');
    const num = parseFloat(normalized);
    
    if (isNaN(num) || num < 0) {
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

