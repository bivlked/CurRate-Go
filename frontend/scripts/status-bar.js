/**
 * Управление строкой состояния (Status Bar)
 */

/**
 * Показывает сообщение в строке состояния
 * @param {string} message - Текст сообщения
 * @param {string} type - Тип сообщения: 'success', 'error', 'warning', 'info'
 * @param {number} duration - Длительность отображения в мс (0 = не скрывать автоматически)
 */
function showStatus(message, type = 'info', duration = 3000) {
    const statusBar = document.getElementById('status-bar');
    const statusIcon = document.getElementById('status-icon');
    const statusMessage = document.getElementById('status-message');
    
    if (!statusBar || !statusIcon || !statusMessage) {
        console.error('Status bar elements not found');
        return;
    }
    
    // Устанавливаем иконку в зависимости от типа
    const icons = {
        success: '✅',
        error: '❌',
        warning: '⚠️',
        info: 'ℹ️'
    };
    
    statusIcon.textContent = icons[type] || icons.info;
    
    // Устанавливаем класс для стилизации
    statusMessage.className = `status-message ${type}`;
    statusMessage.textContent = message;
    
    // Показываем строку состояния
    statusBar.style.display = 'flex';
    
    // Автоматическое скрытие для success и info (если duration > 0)
    if ((type === 'success' || type === 'info') && duration > 0) {
        setTimeout(() => {
            // Проверяем, что сообщение не изменилось
            if (statusMessage.classList.contains(type)) {
                hideStatus();
            }
        }, duration);
    }
}

/**
 * Показывает сообщение об ошибке
 * @param {string} message - Текст ошибки
 */
function showError(message) {
    showStatus(message, 'error', 0); // Ошибки не скрываются автоматически
}

/**
 * Показывает сообщение об успехе
 * @param {string} message - Текст сообщения
 * @param {number} duration - Длительность отображения (по умолчанию 3 секунды)
 */
function showSuccess(message, duration = 3000) {
    showStatus(message, 'success', duration);
}

/**
 * Показывает информационное сообщение
 * @param {string} message - Текст сообщения
 * @param {number} duration - Длительность отображения (по умолчанию 3 секунды)
 */
function showInfo(message, duration = 3000) {
    showStatus(message, 'info', duration);
}

/**
 * Показывает предупреждение
 * @param {string} message - Текст предупреждения
 * @param {number} duration - Длительность отображения (по умолчанию 5 секунд)
 */
function showWarning(message, duration = 5000) {
    showStatus(message, 'warning', duration);
}

/**
 * Скрывает строку состояния
 */
function hideStatus() {
    const statusBar = document.getElementById('status-bar');
    if (statusBar) {
        statusBar.style.display = 'none';
    }
}

/**
 * Очищает сообщение (скрывает строку состояния)
 */
function clearStatus() {
    hideStatus();
}

