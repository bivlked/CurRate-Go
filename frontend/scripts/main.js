/**
 * Основная логика приложения CurRate-Go
 */

// Глобальные переменные
let appInstance = null;

/**
 * Инициализация приложения
 */
function initApp() {
    // Проверяем доступность Wails bindings
    if (typeof window.go === 'undefined' || !window.go.app || !window.go.app.App) {
        console.error('Wails bindings not found. Make sure to run "wails dev" or build the application.');
        showError('Ошибка инициализации: Wails bindings не найдены');
        return;
    }
    
    // Используем правильный путь к Wails bindings
    // Wails автоматически создает window.go.app.App при запуске
    // Проверяем доступность после небольшой задержки (Wails инициализируется асинхронно)
    if (!window.go || !window.go.app || !window.go.app.App) {
        // Ждем инициализации Wails
        setTimeout(() => {
            if (window.go && window.go.app && window.go.app.App) {
                appInstance = window.go.app.App;
                // Повторно инициализируем компоненты
                initCalendar();
                initDateInput();
                initCurrencySelection();
                initAmountInput();
                initConvertButton();
                initCopyButton();
                showInfo('Готов к работе', 2000);
            } else {
                showError('Ошибка: Wails bindings не инициализированы');
            }
        }, 100);
        return;
    }
    
    appInstance = window.go.app.App;
    
    // Инициализация компонентов
    initCalendar();
    initDateInput();
    initCurrencySelection();
    initAmountInput();
    initConvertButton();
    initCopyButton();
    
    // Показываем начальное сообщение
    showInfo('Готов к работе', 2000);
}

/**
 * Инициализация поля ввода даты
 */
function initDateInput() {
    const dateInput = document.getElementById('date-input');
    if (!dateInput) return;
    
    // Устанавливаем сегодняшнюю дату по умолчанию
    const today = new Date();
    dateInput.value = formatDate(today);
    selectedDate = today;
    
    // Обработчик ввода даты
    dateInput.addEventListener('input', debounce(() => {
        const dateStr = dateInput.value.trim();
        
        if (!dateStr) {
            hideRatePreview();
            return;
        }
        
        if (!isValidDateFormat(dateStr)) {
            showError('Неверный формат даты. Используйте формат ДД.ММ.ГГГГ');
            hideRatePreview();
            return;
        }
        
        const date = parseDate(dateStr);
        if (!date) {
            showError('Неверная дата');
            hideRatePreview();
            return;
        }
        
        if (isFutureDate(date)) {
            showError('Дата не может быть в будущем');
            hideRatePreview();
            return;
        }
        
        selectedDate = date;
        setCalendarDate(date);
        updateRatePreview();
        clearStatus();
    }, 300));
    
    // Обработчик фокуса
    dateInput.addEventListener('focus', () => {
        dateInput.select();
    });
}

/**
 * Инициализация выбора валюты
 */
function initCurrencySelection() {
    const currencyInputs = document.querySelectorAll('input[name="currency"]');
    
    currencyInputs.forEach(input => {
        input.addEventListener('change', () => {
            updateRatePreview();
        });
    });
}

/**
 * Инициализация поля ввода суммы
 */
function initAmountInput() {
    const amountInput = document.getElementById('amount-input');
    if (!amountInput) return;
    
    // Обработчик ввода
    amountInput.addEventListener('input', () => {
        const value = amountInput.value.trim();
        
        if (value && parseAmount(value) === null) {
            amountInput.classList.add('error');
        } else {
            amountInput.classList.remove('error');
        }
    });
    
    // Обработчик Enter
    amountInput.addEventListener('keypress', (e) => {
        if (e.key === 'Enter') {
            performConvert();
        }
    });
}

/**
 * Инициализация кнопки конвертации
 */
function initConvertButton() {
    const convertBtn = document.getElementById('convert-btn');
    if (!convertBtn) return;
    
    convertBtn.addEventListener('click', performConvert);
}

/**
 * Инициализация кнопки копирования
 */
function initCopyButton() {
    const copyBtn = document.getElementById('copy-btn');
    if (!copyBtn) return;
    
    copyBtn.addEventListener('click', async () => {
        const resultText = document.getElementById('result-text');
        if (!resultText || !resultText.textContent) {
            showWarning('Нет результата для копирования');
            return;
        }
        
        const success = await copyToClipboard(resultText.textContent);
        if (success) {
            showSuccess('Результат скопирован в буфер обмена', 2000);
        } else {
            showError('Не удалось скопировать в буфер обмена');
        }
    });
}

/**
 * Выполняет конвертацию валюты
 */
async function performConvert() {
    const amountInput = document.getElementById('amount-input');
    const dateInput = document.getElementById('date-input');
    const currencyInput = document.querySelector('input[name="currency"]:checked');
    const convertBtn = document.getElementById('convert-btn');
    const resultCard = document.getElementById('result-card');
    const resultText = document.getElementById('result-text');
    
    if (!amountInput || !dateInput || !currencyInput || !convertBtn || !resultCard || !resultText) {
        showError('Ошибка: не найдены необходимые элементы интерфейса');
        return;
    }
    
    // Валидация суммы
    const amount = parseAmount(amountInput.value);
    if (amount === null || amount <= 0) {
        showError('Введите корректную сумму (положительное число)');
        amountInput.focus();
        return;
    }
    
    // Валидация даты
    const dateStr = dateInput.value.trim();
    if (!dateStr || !isValidDateFormat(dateStr)) {
        showError('Введите корректную дату в формате ДД.ММ.ГГГГ');
        dateInput.focus();
        return;
    }
    
    const date = parseDate(dateStr);
    if (!date || isFutureDate(date)) {
        showError('Дата не может быть в будущем');
        dateInput.focus();
        return;
    }
    
    // Получаем валюту
    const currency = currencyInput.value;
    
    // Блокируем кнопку
    convertBtn.disabled = true;
    convertBtn.textContent = 'Конвертация...';
    clearStatus();
    
    try {
        // Вызываем метод Go через Wails bindings
        const response = await appInstance.Convert({
            amount: amount,
            currency: currency,
            date: dateStr
        });
        
        if (response.success) {
            // Показываем результат
            resultText.textContent = response.result;
            resultCard.style.display = 'block';
            showSuccess('Конвертация выполнена успешно', 2000);
        } else {
            showError(response.error || 'Ошибка конвертации');
            resultCard.style.display = 'none';
        }
    } catch (error) {
        console.error('Convert error:', error);
        showError(error.message || 'Произошла ошибка при конвертации');
        resultCard.style.display = 'none';
    } finally {
        // Разблокируем кнопку
        convertBtn.disabled = false;
        convertBtn.textContent = 'Конвертировать';
    }
}

/**
 * Обновляет live preview курса
 */
const updateRatePreview = debounce(async () => {
    const dateInput = document.getElementById('date-input');
    const currencyInput = document.querySelector('input[name="currency"]:checked');
    const ratePreview = document.getElementById('rate-preview');
    const rateValue = document.getElementById('rate-value');
    
    if (!dateInput || !currencyInput || !ratePreview || !rateValue) {
        return;
    }
    
    const dateStr = dateInput.value.trim();
    if (!dateStr || !isValidDateFormat(dateStr)) {
        hideRatePreview();
        return;
    }
    
    const date = parseDate(dateStr);
    if (!date || isFutureDate(date)) {
        hideRatePreview();
        return;
    }
    
    const currency = currencyInput.value;
    
    try {
        const response = await appInstance.GetRate(currency, dateStr);
        
        if (response.success) {
            rateValue.textContent = `${formatNumber(response.rate, 4)} ₽`;
            ratePreview.style.display = 'flex';
        } else {
            hideRatePreview();
            if (response.error) {
                // Не показываем ошибку в status bar для live preview, только скрываем preview
                console.warn('Rate preview error:', response.error);
            }
        }
    } catch (error) {
        hideRatePreview();
        console.warn('Rate preview error:', error);
    }
}, 300);

/**
 * Скрывает live preview курса
 */
function hideRatePreview() {
    const ratePreview = document.getElementById('rate-preview');
    if (ratePreview) {
        ratePreview.style.display = 'none';
    }
}

/**
 * Получает выбранную валюту
 * @returns {string|null} Код валюты или null
 */
function getSelectedCurrency() {
    const currencyInput = document.querySelector('input[name="currency"]:checked');
    return currencyInput ? currencyInput.value : null;
}

/**
 * Получает выбранную дату
 * @returns {string|null} Дата в формате DD.MM.YYYY или null
 */
function getSelectedDate() {
    const dateInput = document.getElementById('date-input');
    if (!dateInput || !dateInput.value) {
        return null;
    }
    const dateStr = dateInput.value.trim();
    if (!isValidDateFormat(dateStr)) {
        return null;
    }
    return dateStr;
}

// Инициализация при загрузке страницы
document.addEventListener('DOMContentLoaded', () => {
    initApp();
});

// Обработчик события изменения даты из календаря
document.addEventListener('datechanged', (e) => {
    updateRatePreview();
});

