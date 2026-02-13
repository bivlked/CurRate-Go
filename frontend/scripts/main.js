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
            initAboutButton();
            
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
    
    // Обработчик paste для форматирования вставленного текста
    dateInput.addEventListener('paste', (e) => {
        e.preventDefault();
        const pastedText = (e.clipboardData || window.clipboardData).getData('text');
        const digits = pastedText.replace(/\D/g, '').slice(0, 8); // Ограничиваем до 8 цифр
        if (digits.length > 0) {
            const formatted = autoFormatDate(digits);
            dateInput.value = formatted;

            // Явно вызываем событие input для запуска валидации и превью курса
            // Это необходимо, потому что программная установка value не генерирует событие автоматически
            dateInput.dispatchEvent(new Event('input', { bubbles: true }));

            setTimeout(() => {
                dateInput.setSelectionRange(formatted.length, formatted.length);
            }, 0);
        }
    });
    
    // Обработчик input: автоформатирование + валидация + preview курса
    dateInput.addEventListener('input', (e) => {
        const value = e.target.value;
        const cursorPos = e.target.selectionStart || 0;

        // Автоформатирование даты
        if (!value.includes('.')) {
            const formatted = autoFormatDate(value);
            if (formatted !== value) {
                e.target.value = formatted;
                const newCursorPos = Math.min(cursorPos + (formatted.length - value.length), formatted.length);
                setTimeout(() => {
                    e.target.setSelectionRange(newCursorPos, newCursorPos);
                }, 0);
            }
        } else {
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

        // Валидация и обновление preview курса
        const dateStr = dateInput.value.trim();

        if (!dateStr) {
            hideRatePreview();
            return;
        }

        if (!isValidDateFormat(dateStr)) {
            if (dateStr.length === 10) {
                showError('Неверный формат даты. Используйте формат ДД.ММ.ГГГГ');
                hideRatePreview();
            }
            return;
        }

        const date = parseDate(dateStr);
        if (!date) {
            if (dateStr.length === 10) {
                showError('Неверная дата');
                hideRatePreview();
            }
            return;
        }

        if (isFutureDate(date)) {
            showError('Дата не может быть в будущем');
            hideRatePreview();
            return;
        }

        if (typeof setCalendarDate === 'function') {
            setCalendarDate(date);
        }
        updateRatePreview();
        clearStatus();
    });
    
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
 * Инициализация кнопки "О программе" и модального окна
 */
function initAboutButton() {
    const aboutBtn = document.getElementById('about-btn');
    const aboutModal = document.getElementById('about-modal');
    const closeBtn = aboutModal?.querySelector('.about-modal-close');
    const sendStarBtn = document.getElementById('send-star-btn');

    if (!aboutBtn || !aboutModal) return;

    // Открытие модального окна
    aboutBtn.addEventListener('click', () => {
        aboutModal.showModal();
    });

    // Закрытие по кнопке ×
    if (closeBtn) {
        closeBtn.addEventListener('click', () => {
            aboutModal.close();
        });
    }

    // Закрытие по клику на backdrop
    aboutModal.addEventListener('click', (e) => {
        if (e.target === aboutModal) {
            aboutModal.close();
        }
    });

    // Закрытие по Escape (встроено в dialog, но добавим для надёжности)
    aboutModal.addEventListener('keydown', (e) => {
        if (e.key === 'Escape') {
            aboutModal.close();
        }
    });

    // Кнопка "Отправить звезду"
    if (sendStarBtn) {
        sendStarBtn.addEventListener('click', async () => {
            if (appInstance && typeof appInstance.SendStar === 'function') {
                try {
                    const response = await appInstance.SendStar();
                    if (response.success) {
                        showSuccess('Спасибо за звезду! ⭐', 3000);
                        aboutModal.close();
                    } else {
                        showError(response.error || 'Не удалось отправить звезду');
                    }
                } catch (error) {
                    console.error('SendStar error:', error);
                    showError('Произошла ошибка при отправке');
                }
            } else {
                // Заглушка, пока функционал не реализован
                showInfo('Функция будет доступна в следующей версии', 3000);
            }
        });
    }
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
            // Храним полную строку результата для копирования
            resultText.textContent = response.result;

            // Заполняем улучшенный UI результата (если элементы доступны)
            const resultAmountEl = document.getElementById('result-amount');
            const resultMetaEl = document.getElementById('result-meta');

            if (resultAmountEl && resultMetaEl) {
                const amountRub = typeof response.targetAmountRUB === 'number'
                    ? response.targetAmountRUB
                    : NaN;
                const rate = typeof response.rate === 'number' ? response.rate : NaN;
                const srcAmount = typeof response.sourceAmount === 'number' ? response.sourceAmount : NaN;
                const symbol = response.currencySymbol || '';
                const requestedDate = response.requestedDate || '';
                const actualDate = response.actualDate || '';

                resultAmountEl.textContent = `${formatNumber(amountRub, 2)} ₽`;

                const dateLine = (actualDate && requestedDate && actualDate !== requestedDate)
                    ? `Курс фактически за ${actualDate} (запрошено ${requestedDate})`
                    : (actualDate ? `Курс за ${actualDate}` : '');

                resultMetaEl.textContent = `${dateLine}${dateLine ? ' · ' : ''}${symbol}${formatNumber(srcAmount, 2)} по курсу ${formatNumber(rate, 4)} ₽`;
            }

            resultCard.classList.remove('hidden');
            showSuccess('Конвертация выполнена успешно', 2000);
        } else {
            showError(response.error || 'Ошибка конвертации');
            resultCard.classList.add('hidden');
        }
    } catch (error) {
        console.error('Convert error:', error);
        showError(error.message || 'Произошла ошибка при конвертации');
        resultCard.classList.add('hidden');
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
 * Скрывает live preview курса (показывает прочерк)
 */
function hideRatePreview() {
    const ratePreview = document.getElementById('rate-preview');
    const rateValue = document.getElementById('rate-value');
    
    if (ratePreview && rateValue) {
        // Не скрываем элемент, показываем прочерк
        rateValue.textContent = '—';
        // Элемент остается видимым (display: flex из CSS)
    }
}

// Инициализация при загрузке страницы
document.addEventListener('DOMContentLoaded', () => {
    initApp();
});

// Обработчик события изменения даты из календаря
document.addEventListener('datechanged', (e) => {
    updateRatePreview();
});
