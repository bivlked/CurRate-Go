/**
 * Основная логика приложения CurRate-Go
 */

// Глобальные переменные
let appInstance = null;

/**
 * Инициализация приложения
 */
function initApp() {
    // Wails автоматически создает window.go.app.App при запуске
    // Ждем инициализации Wails (может быть асинхронной)
    function checkWailsBindings() {
        if (typeof window.go !== 'undefined' && window.go.app && window.go.app.App) {
            appInstance = window.go.app.App;
            
            // Инициализация компонентов
            initCalendar(); // Сначала календарь
            initDateInput(); // Затем поле даты (вызывает updateRatePreview)
            initCurrencySelection();
            initAmountInput();
            initConvertButton();
            initCopyButton();
            
            // Показываем начальное сообщение
            showInfo('Готов к работе', 2000);
            return true;
        }
        return false;
    }
    
    // Проверяем сразу
    if (checkWailsBindings()) {
        return;
    }
    
    // Если не готово, ждем с повторными попытками
    let attempts = 0;
    const maxAttempts = 50; // 5 секунд максимум
    
    const checkInterval = setInterval(() => {
        attempts++;
        if (checkWailsBindings()) {
            clearInterval(checkInterval);
        } else if (attempts >= maxAttempts) {
            clearInterval(checkInterval);
            console.error('Wails bindings not found after', attempts * 100, 'ms');
            showError('Ошибка инициализации: Wails bindings не найдены. Убедитесь, что приложение запущено через "wails dev"');
        }
    }, 100);
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
    
    // Устанавливаем дату в календаре (если он уже инициализирован)
    if (typeof setCalendarDate === 'function') {
        setCalendarDate(today);
    }
    
    // Показываем курс на сегодня при запуске
    // Используем setTimeout для гарантии, что appInstance готов и календарь инициализирован
    setTimeout(() => {
        if (appInstance) {
            updateRatePreview();
        }
    }, 200);
    
    // Обработчик keydown для валидации ввода
    dateInput.addEventListener('keydown', (e) => {
        // Разрешаем специальные клавиши
        if (e.key === 'Backspace' || e.key === 'Delete' || e.key === 'Tab' || 
            e.key === 'ArrowLeft' || e.key === 'ArrowRight' || e.key === 'ArrowUp' || e.key === 'ArrowDown' ||
            e.key === 'Home' || e.key === 'End' || e.ctrlKey || e.metaKey) {
            return;
        }
        
        // Разрешаем вставку (Ctrl+V, Cmd+V)
        if ((e.ctrlKey || e.metaKey) && e.key === 'v') {
            return;
        }
        
        // Для остальных символов проверяем валидность
        if (e.key.length === 1) {
            const cursorPos = dateInput.selectionStart || 0;
            if (!isValidDateChar(dateInput.value, e.key, cursorPos)) {
                e.preventDefault();
                return false;
            }
        }
    });
    
    // Обработчик input для автоформатирования
    dateInput.addEventListener('input', (e) => {
        const value = e.target.value;
        const cursorPos = e.target.selectionStart || 0;
        
        // Если ввод сплошняком (без точек), форматируем автоматически
        if (!value.includes('.')) {
            const formatted = autoFormatDate(value);
            if (formatted !== value) {
                e.target.value = formatted;
                // Восстанавливаем позицию курсора
                const newCursorPos = Math.min(cursorPos + (formatted.length - value.length), formatted.length);
                setTimeout(() => {
                    e.target.setSelectionRange(newCursorPos, newCursorPos);
                }, 0);
            }
        } else {
            // Если есть точки, проверяем формат и исправляем при необходимости
            const digits = value.replace(/\D/g, '');
            if (digits.length > 0 && digits.length <= 8) {
                const formatted = autoFormatDate(digits);
                if (formatted !== value) {
                    e.target.value = formatted;
                    setTimeout(() => {
                        const newCursorPos = Math.min(cursorPos, formatted.length);
                        e.target.setSelectionRange(newCursorPos, newCursorPos);
                    }, 0);
                }
            }
        }
    });
    
    // Обработчик ввода даты (валидация и обновление)
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
        
        // Обновляем календарь (если он инициализирован)
        if (typeof setCalendarDate === 'function') {
            setCalendarDate(date);
        }
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
            // ratePreview всегда видим (display: flex в CSS), не нужно менять display
        } else {
            hideRatePreview(); // Показывает '—' вместо скрытия
            if (response.error) {
                // Не показываем ошибку в status bar для live preview, только показываем прочерк
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

