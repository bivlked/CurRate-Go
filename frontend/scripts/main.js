/**
 * Основная логика приложения CurRate-Go
 */

// Глобальные переменные
let appInstance = null;

/**
 * Последние полученные курсы для строки сравнения: { rate, actualDate } или null.
 * Обновляются только ответами на актуальные запросы (см. createRatePreview)
 */
const rateState = {
    main: null,
    ref: null
};

let mainPreview = null;
let refPreview = null;

/**
 * Инициализация приложения
 */
function initApp() {
    // Wails автоматически создает window.go.app.App при запуске
    // Ждем инициализации Wails (может быть асинхронной)
    function checkWailsBindings() {
        if (typeof window.go !== 'undefined' && window.go.app && window.go.app.App) {
            appInstance = window.go.app.App;

            mainPreview = createRatePreview({
                inputId: 'date-input',
                boxId: 'rate-preview',
                valueId: 'rate-value',
                stateKey: 'main',
                what: 'Курс ЦБ'
            });
            refPreview = createRatePreview({
                inputId: 'ref-date-input',
                boxId: 'ref-rate-preview',
                valueId: 'ref-rate-value',
                stateKey: 'ref',
                what: 'Курс ЦБ на референсную дату'
            });

            // Инициализация компонентов
            initCalendar(); // Сначала календарь
            initDateInput(); // Затем поля дат (запрашивают курсы)
            initReferenceInput();
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
            showError('Ошибка инициализации приложения. Перезапустите программу.');
        }
    }, 100);
}

/**
 * Подключает к полю маску и проверку даты ДД.ММ.ГГГГ: фильтр клавиш (с учётом
 * выделения), форматирование вставки, автоточки, проверка будущего.
 * Общий код для поля даты курса и поля референсной даты
 * @param {HTMLInputElement} input - Поле даты
 * @param {{onValid: function(string, Date): void, onInvalid: function(): void}} handlers
 *        onValid - дата полная и допустимая; onInvalid - пусто, не дописано или ошибка
 */
function bindDateInput(input, handlers) {
    // Обработчик keydown для валидации ввода
    input.addEventListener('keydown', (e) => {
        // Разрешаем специальные клавиши и любые сочетания с Ctrl/Cmd
        // (включая вставку Ctrl+V - её форматирует обработчик paste)
        if (e.key === 'Backspace' || e.key === 'Delete' || e.key === 'Tab' ||
            e.key === 'ArrowLeft' || e.key === 'ArrowRight' || e.key === 'ArrowUp' || e.key === 'ArrowDown' ||
            e.key === 'Home' || e.key === 'End' || e.key === 'Enter' || e.key === 'Escape' ||
            e.ctrlKey || e.metaKey) {
            return;
        }

        // Для остальных символов проверяем валидность
        if (e.key.length === 1) {
            const cursorPos = input.selectionStart || 0;
            const selectionEnd = input.selectionEnd ?? cursorPos;
            if (!isValidDateChar(input.value, e.key, cursorPos, selectionEnd)) {
                e.preventDefault();
            }
        }
    });

    // Обработчик paste для форматирования вставленного текста
    input.addEventListener('paste', (e) => {
        e.preventDefault();
        const pastedText = (e.clipboardData || window.clipboardData).getData('text');
        const digits = pastedText.replace(/\D/g, '').slice(0, 8); // Ограничиваем до 8 цифр
        if (digits.length > 0) {
            const formatted = autoFormatDate(digits);
            input.value = formatted;

            // Программная установка value не генерирует input: вызываем явно
            input.dispatchEvent(new Event('input', { bubbles: true }));

            setTimeout(() => {
                input.setSelectionRange(formatted.length, formatted.length);
            }, 0);
        }
    });

    // Обработчик input: автоформатирование + валидация
    input.addEventListener('input', () => {
        const value = input.value;
        const cursorPos = input.selectionStart || 0;

        // Автоформатирование даты
        if (!value.includes('.')) {
            const formatted = autoFormatDate(value);
            if (formatted !== value) {
                input.value = formatted;
                const newCursorPos = Math.min(cursorPos + (formatted.length - value.length), formatted.length);
                setTimeout(() => {
                    input.setSelectionRange(newCursorPos, newCursorPos);
                }, 0);
            }
        } else {
            const digits = value.replace(/\D/g, '');
            if (digits.length > 0 && digits.length <= 8) {
                const formatted = autoFormatDate(digits);
                if (formatted !== value) {
                    input.value = formatted;
                    setTimeout(() => {
                        const newCursorPos = Math.min(cursorPos, formatted.length);
                        input.setSelectionRange(newCursorPos, newCursorPos);
                    }, 0);
                }
            }
        }

        const dateStr = input.value.trim();

        // Неполная или неверная дата гасит превью сразу: иначе ответ на
        // запрос для прежней даты показал бы курс рядом с недописанной
        if (!dateStr || !isValidDateFormat(dateStr)) {
            if (dateStr.length === 10) {
                showError('Неверная дата. Используйте формат ДД.ММ.ГГГГ');
            }
            handlers.onInvalid();
            return;
        }

        const date = parseDate(dateStr);
        if (isFutureDate(date)) {
            showError('Дата не может быть в будущем');
            handlers.onInvalid();
            return;
        }

        handlers.onValid(dateStr, date);
        clearStatus();
    });

    // При фокусе выделяем дату целиком: её можно сразу перепечатать
    input.addEventListener('focus', () => {
        input.select();
    });
}

/**
 * Возвращает допустимую дату из поля или null (пусто, не дописано, будущее)
 * @param {string} inputId - id поля даты
 * @returns {string|null} Дата ДД.ММ.ГГГГ
 */
function readValidDate(inputId) {
    const input = document.getElementById(inputId);
    const dateStr = input ? input.value.trim() : '';
    if (!dateStr || !isValidDateFormat(dateStr)) {
        return null;
    }
    const date = parseDate(dateStr);
    if (!date || isFutureDate(date)) {
        return null;
    }
    return dateStr;
}

/**
 * Инициализация поля даты курса
 */
function initDateInput() {
    const dateInput = document.getElementById('date-input');
    if (!dateInput) return;

    // Устанавливаем сегодняшнюю дату по умолчанию
    const today = new Date();
    dateInput.value = formatDate(today);
    setCalendarDate(today);

    bindDateInput(dateInput, {
        onValid: (dateStr, date) => {
            setCalendarDate(date);
            mainPreview.request();
        },
        onInvalid: () => mainPreview.invalidate()
    });

    // Курс на сегодня при запуске
    mainPreview.request();
}

/**
 * Инициализация референсной даты: при запуске - сегодня,
 * кнопка «Сегодня» возвращает её к текущей дате
 */
function initReferenceInput() {
    const refInput = document.getElementById('ref-date-input');
    const todayBtn = document.getElementById('ref-today-btn');
    if (!refInput) return;

    bindDateInput(refInput, {
        onValid: (dateStr, date) => {
            setCalendarReference(date);
            refPreview.request();
        },
        onInvalid: () => {
            setCalendarReference(null);
            refPreview.invalidate();
        }
    });

    const setToday = () => {
        refInput.value = formatDate(new Date());
        refInput.dispatchEvent(new Event('input', { bubbles: true }));
    };

    if (todayBtn) {
        todayBtn.addEventListener('click', setToday);
    }

    setToday();
}

/**
 * Инициализация выбора валюты
 */
function initCurrencySelection() {
    const currencyInputs = document.querySelectorAll('input[name="currency"]');

    currencyInputs.forEach(input => {
        input.addEventListener('change', () => {
            updateAmountSuffix();
            mainPreview.request();
            refPreview.request();
        });
    });

    updateAmountSuffix();
}

/**
 * Знак выбранной валюты в поле суммы
 */
function updateAmountSuffix() {
    const suffix = document.getElementById('amount-suffix');
    const currencyInput = document.querySelector('input[name="currency"]:checked');
    if (suffix && currencyInput) {
        suffix.textContent = currencyInput.value === 'EUR' ? '€' : '$';
    }
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
    amountInput.addEventListener('keydown', (e) => {
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

    // Кнопка "Отправить звезду"
    if (sendStarBtn) {
        sendStarBtn.addEventListener('click', async () => {
            // Повторные клики во время отправки не шлют вторую звезду
            if (sendStarBtn.disabled) return;
            if (appInstance && typeof appInstance.SendStar === 'function') {
                sendStarBtn.disabled = true;
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
                } finally {
                    sendStarBtn.disabled = false;
                }
            } else {
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

    // Enter в поле суммы вызывает конвертацию в обход заблокированной кнопки:
    // второй параллельный запрос мог бы перезаписать результат устаревшим
    if (convertBtn.disabled) {
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

    // Референс необязателен: недопустимая референсная дата просто не участвует
    const refDate = readValidDate('ref-date-input') || '';

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
            date: dateStr,
            refDate: refDate
        });

        if (response.success) {
            renderResult(response);
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
 * Заполняет карточку результата
 * @param {object} response - ConvertResponse из Go
 */
function renderResult(response) {
    // Полная строка результата - для копирования (формат для документов не меняется)
    document.getElementById('result-text').textContent = response.result;

    const resultAmountEl = document.getElementById('result-amount');
    const resultMetaEl = document.getElementById('result-meta');
    if (resultAmountEl && resultMetaEl) {
        const amountRub = typeof response.targetAmountRUB === 'number' ? response.targetAmountRUB : NaN;
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

    // Сравнение с суммой по референсному курсу. Не показываем, если референс
    // недоступен или совпадает с датой курса (сравнивать нечего)
    const refEl = document.getElementById('result-ref');
    const refTextEl = document.getElementById('result-ref-text');
    if (!refEl || !refTextEl) return;

    if (!response.hasReference || response.refRequestedDate === response.requestedDate) {
        refEl.classList.add('hidden');
        if (response.refError) {
            console.warn('Reference conversion error:', response.refError);
        }
        return;
    }

    const view = describeAmountComparison(response.refDiffRUB, response.refAmountRUB);
    refEl.className = `compare result-ref ${view.className}`;
    refTextEl.textContent = view.text;
    refEl.title = `Референс ${response.refRequestedDate}: курс ${formatNumber(response.refRate, 4)} ₽ (курс ЦБ за ${response.refActualDate})`;
}

/**
 * Превью курса для поля даты: плашка с курсом, свой счётчик запросов и
 * состояние загрузки. Курс попадает в rateState только из ответа на
 * последний запрос; устаревшие ответы отбрасываются
 * @param {{inputId: string, boxId: string, valueId: string, stateKey: 'main'|'ref', what: string}} cfg
 * @returns {{request: function(): void, invalidate: function(): void}}
 */
function createRatePreview(cfg) {
    let requestId = 0;
    let pending = false;

    const box = () => document.getElementById(cfg.boxId);
    const value = () => document.getElementById(cfg.valueId);

    function setPending(isPending) {
        pending = isPending;
        const el = box();
        if (el) el.classList.toggle('loading', isPending);
        renderComparison();
    }

    function show(text, title) {
        const valueEl = value();
        const boxEl = box();
        if (valueEl) valueEl.textContent = text;
        if (boxEl) boxEl.title = title || '';
    }

    const fetchRate = debounce(async (id) => {
        const dateStr = readValidDate(cfg.inputId);
        const currencyInput = document.querySelector('input[name="currency"]:checked');
        if (id !== requestId) return;
        if (!dateStr || !currencyInput || !appInstance) {
            rateState[cfg.stateKey] = null;
            show('…');
            setPending(false);
            return;
        }

        try {
            const response = await appInstance.GetRate(currencyInput.value, dateStr);
            if (id !== requestId) return; // пока ждали, ввод изменился

            if (response.success) {
                rateState[cfg.stateKey] = { rate: response.rate, actualDate: response.actualDate };
                const title = response.actualDate && response.actualDate !== dateStr
                    ? `${cfg.what} за ${response.actualDate}: на ${dateStr} курс не устанавливался`
                    : `${cfg.what} на ${dateStr}`;
                show(`${formatNumber(response.rate, 4)} ₽`, title);
            } else {
                rateState[cfg.stateKey] = null;
                show('н/д', response.error || '');
                console.warn('Rate preview error:', response.error);
            }
        } catch (error) {
            if (id !== requestId) return;
            rateState[cfg.stateKey] = null;
            show('н/д');
            console.warn('Rate preview error:', error);
        }
        setPending(false);
    }, 300);

    return {
        // Номер запроса растёт сразу, а не после debounce: ответ, который ещё
        // летит для прежней даты или валюты, становится устаревшим немедленно
        request() {
            requestId++;
            setPending(true);
            fetchRate(requestId);
        },
        // Дата пуста, не дописана или в будущем: гасим курс и ответы в пути
        invalidate() {
            requestId++;
            rateState[cfg.stateKey] = null;
            show('…');
            setPending(false);
        },
        isPending: () => pending
    };
}

/**
 * Обновляет строку сравнения курса с референсом. Пока любой из курсов
 * загружается, прежнее сравнение приглушено, чтобы не сопоставлять курс
 * новой даты со старым
 */
function renderComparison() {
    const el = document.getElementById('ref-compare');
    const textEl = document.getElementById('ref-compare-text');
    if (!el || !textEl) return;

    const loading = (mainPreview && mainPreview.isPending()) || (refPreview && refPreview.isPending());
    const main = rateState.main;
    const ref = rateState.ref;
    const view = describeRateComparison(main && ref ? compareRates(main.rate, ref.rate) : null);

    el.className = `compare ${view.className}${loading ? ' loading' : ''}`;
    textEl.textContent = view.text;
}

// Инициализация при загрузке страницы
document.addEventListener('DOMContentLoaded', () => {
    initApp();
});

// Обработчик события изменения даты из календаря
document.addEventListener('datechanged', () => {
    if (mainPreview) mainPreview.request();
});
