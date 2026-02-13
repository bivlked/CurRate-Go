/**
 * Календарь для выбора даты с выделением выходных
 */

let currentDate = new Date();
let selectedDate = null;

/**
 * Инициализация календаря
 */
function initCalendar() {
    const calendar = document.getElementById('calendar');
    
    if (!calendar) {
        console.error('Calendar element not found');
        return;
    }
    
    // Устанавливаем текущую дату по умолчанию
    const today = new Date();
    today.setHours(0, 0, 0, 0);
    currentDate = new Date(today.getFullYear(), today.getMonth(), 1);
    selectedDate = today;
    
    // Инициализация отображения
    renderCalendar();
}

/**
 * Рендерит календарь
 */
function renderCalendar() {
    const calendar = document.getElementById('calendar');
    if (!calendar) return;
    
    const year = currentDate.getFullYear();
    const month = currentDate.getMonth();
    
    // Заголовок календаря
    const monthNames = [
        'Январь', 'Февраль', 'Март', 'Апрель', 'Май', 'Июнь',
        'Июль', 'Август', 'Сентябрь', 'Октябрь', 'Ноябрь', 'Декабрь'
    ];
    
    const firstDay = getFirstDayOfMonth(currentDate);
    const lastDay = getLastDayOfMonth(currentDate);
    const firstDayWeekJS = firstDay.getDay(); // 0 = воскресенье, 1 = понедельник, ..., 6 = суббота
    // Конвертируем в европейский формат: понедельник = 0, воскресенье = 6
    const firstDayWeek = firstDayWeekJS === 0 ? 6 : firstDayWeekJS - 1;
    const daysInMonth = lastDay.getDate();
    
    // Названия дней недели (европейский формат: начинается с понедельника)
    const weekdays = ['Пн', 'Вт', 'Ср', 'Чт', 'Пт', 'Сб', 'Вс'];
    
    let html = `
        <div class="calendar-header">
            <button class="calendar-nav-btn" id="calendar-prev">‹</button>
            <div class="calendar-month-year">${monthNames[month]} ${year}</div>
            <button class="calendar-nav-btn" id="calendar-next">›</button>
        </div>
        <div class="calendar-weekdays">
    `;
    
    // Заголовки дней недели (индексы: 0=Пн, 1=Вт, 2=Ср, 3=Чт, 4=Пт, 5=Сб, 6=Вс)
    weekdays.forEach((day, index) => {
        const isWeekend = index === 5 || index === 6; // Суббота (5) или воскресенье (6)
        html += `<div class="calendar-weekday ${isWeekend ? 'weekend' : ''}">${day}</div>`;
    });
    
    html += '</div><div class="calendar-days">';
    
    // Пустые ячейки до первого дня месяца
    for (let i = 0; i < firstDayWeek; i++) {
        html += '<div class="calendar-day other-month"></div>';
    }
    
    // Дни месяца
    const today = new Date();
    today.setHours(0, 0, 0, 0);
    
    for (let day = 1; day <= daysInMonth; day++) {
        const date = new Date(year, month, day);
        const isWeekendDay = isWeekend(date);
        const isToday = date.getTime() === today.getTime();
        const isSelected = selectedDate && 
            date.getTime() === selectedDate.getTime();
        const isFuture = date > today;
        
        let classes = 'calendar-day';
        if (isWeekendDay) classes += ' weekend';
        if (isToday) classes += ' today';
        if (isSelected) classes += ' selected';
        if (isFuture) classes += ' disabled';
        
        html += `<div class="${classes}" data-date="${formatDate(date)}" ${isFuture ? '' : 'onclick="selectDate(this)"'}>${day}</div>`;
    }
    
    html += '</div>';
    
    calendar.innerHTML = html;
    
    // Обработчики навигации
    const prevBtn = document.getElementById('calendar-prev');
    const nextBtn = document.getElementById('calendar-next');
    
    if (prevBtn) {
        prevBtn.addEventListener('click', () => {
            currentDate = addMonths(currentDate, -1);
            renderCalendar();
        });
    }
    
    if (nextBtn) {
        nextBtn.addEventListener('click', () => {
            currentDate = addMonths(currentDate, 1);
            renderCalendar();
        });
    }
}

/**
 * Выбирает дату в календаре
 * @param {HTMLElement} element - Элемент дня календаря
 */
function selectDate(element) {
    const dateStr = element.getAttribute('data-date');
    if (!dateStr) return;
    
    const date = parseDate(dateStr);
    if (!date) return;
    
    selectedDate = date;
    
    // Обновляем поле ввода даты
    const dateInput = document.getElementById('date-input');
    if (dateInput) {
        dateInput.value = dateStr;
    }
    
    // Обновляем календарь
    renderCalendar();
    
    // Создаем событие для других обработчиков
    // Обработчик события в main.js вызовет updateRatePreview()
    const event = new CustomEvent('datechanged', { detail: { date: dateStr } });
    document.dispatchEvent(event);
}

/**
 * Устанавливает текущую дату в календаре
 * @param {Date} date - Дата для установки
 */
function setCalendarDate(date) {
    if (!(date instanceof Date) || isNaN(date)) {
        return;
    }
    
    currentDate = new Date(date.getFullYear(), date.getMonth(), 1);
    selectedDate = date;
    renderCalendar();
}

