# Style Guide: CurRate-Go Desktop GUI

**Версия:** 1.0  
**Дата создания:** 2025-12-22  
**Основан на:** Creative Phase GUI Design решения

---

## 🎨 Цветовая палитра

### Основные цвета

```css
/* Primary - Основной акцентный цвет (Windows 11 синий) */
--primary-color: #0078d4;
--primary-hover: #106ebe;
--primary-active: #005a9e;
--primary-light: #e3f2fd;

/* Success - Успешные операции */
--success-color: #107c10;
--success-bg: #e8f5e9;

/* Error - Ошибки и предупреждения */
--error-color: #d32f2f;
--error-bg: #ffebee;

/* Weekend - Выходные дни в календаре */
--weekend-color: #d32f2f;
--weekend-bg: #ffebee;
```

### Нейтральные цвета

```css
/* Фоны */
--bg-primary: #ffffff;           /* Основной фон окна */
--bg-secondary: #f5f5f5;         /* Фон карточек */
--bg-tertiary: #e8e8e8;           /* Фон неактивных элементов */

/* Текст */
--text-primary: #1f1f1f;          /* Основной текст */
--text-secondary: #605e5c;       /* Вторичный текст */
--text-disabled: #a19f9d;        /* Неактивный текст */

/* Границы */
--border-color: #e1dfdd;          /* Обычные границы */
--border-focus: #0078d4;          /* Граница при фокусе */
--border-error: #d32f2f;         /* Граница при ошибке */
```

### Тени

```css
--shadow-sm: 0 1px 2px rgba(0, 0, 0, 0.05);
--shadow-md: 0 2px 4px rgba(0, 0, 0, 0.1);
--shadow-lg: 0 4px 8px rgba(0, 0, 0, 0.15);
--shadow-xl: 0 8px 16px rgba(0, 0, 0, 0.2);
```

---

## 📝 Типографика

### Шрифты

```css
--font-family: 'Segoe UI', -apple-system, BlinkMacSystemFont, 'Roboto', 'Helvetica Neue', Arial, sans-serif;
```

**Использование:**
- Основной шрифт для всего интерфейса
- Segoe UI - нативный шрифт Windows 11
- Fallback на системные шрифты для совместимости

### Размеры шрифтов

```css
--font-size-xs: 12px;    /* Мелкие подписи */
--font-size-sm: 13px;    /* Вторичный текст */
--font-size-base: 14px;  /* Основной текст (по умолчанию) */
--font-size-lg: 16px;    /* Заголовки карточек */
--font-size-xl: 18px;    /* Крупные заголовки */
--font-size-2xl: 24px;   /* Результат конвертации */
```

### Веса шрифтов

```css
--font-weight-normal: 400;    /* Обычный текст */
--font-weight-medium: 500;    /* Акцентный текст */
--font-weight-semibold: 600;  /* Заголовки */
--font-weight-bold: 700;      /* Важные элементы */
```

### Высота строк

```css
--line-height-tight: 1.2;   /* Заголовки */
--line-height-normal: 1.5; /* Обычный текст */
--line-height-relaxed: 1.75; /* Длинный текст */
```

---

## 📐 Система отступов

### Базовый размер

```css
--spacing-unit: 4px;
```

### Размеры отступов

```css
--spacing-xs: 4px;    /* 1 unit */
--spacing-sm: 8px;    /* 2 units */
--spacing-md: 12px;   /* 3 units */
--spacing-lg: 16px;   /* 4 units */
--spacing-xl: 24px;   /* 6 units */
--spacing-2xl: 32px;  /* 8 units */
```

### Использование

- **Внутренние отступы карточек:** `16px` (--spacing-lg)
- **Отступы между карточками:** `16px` (--spacing-lg)
- **Отступы внутри элементов:** `12px` (--spacing-md)
- **Внешние отступы контейнера:** `24px` (--spacing-xl)

---

## 🎴 Компоненты

### Карточки (Cards)

```css
.card {
  background: var(--bg-secondary);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  padding: var(--spacing-lg);
  margin-bottom: var(--spacing-lg);
  box-shadow: var(--shadow-sm);
  transition: box-shadow 0.2s ease;
}

.card:hover {
  box-shadow: var(--shadow-md);
}

.card-label {
  font-size: var(--font-size-lg);
  font-weight: var(--font-weight-semibold);
  color: var(--text-primary);
  margin-bottom: var(--spacing-md);
}
```

**Размеры:**
- Border radius: `8px`
- Padding: `16px`
- Margin bottom: `16px`

### Кнопки

#### Primary Button (Конвертировать)

```css
.btn-primary {
  background: var(--primary-color);
  color: white;
  padding: 12px 24px;
  border: none;
  border-radius: 6px;
  font-size: var(--font-size-base);
  font-weight: var(--font-weight-medium);
  cursor: pointer;
  transition: all 0.2s ease;
}

.btn-primary:hover {
  background: var(--primary-hover);
  transform: translateY(-1px);
  box-shadow: var(--shadow-md);
}

.btn-primary:active {
  background: var(--primary-active);
  transform: translateY(0);
}

.btn-primary:disabled {
  background: var(--bg-tertiary);
  color: var(--text-disabled);
  cursor: not-allowed;
}
```

#### Secondary Button (Копировать)

```css
.btn-secondary {
  background: var(--bg-secondary);
  color: var(--text-primary);
  border: 1px solid var(--border-color);
  padding: 10px 20px;
  border-radius: 6px;
  font-size: var(--font-size-base);
  font-weight: var(--font-weight-medium);
  cursor: pointer;
  transition: all 0.2s ease;
}

.btn-secondary:hover {
  background: var(--bg-tertiary);
  border-color: var(--primary-color);
}
```

### Поля ввода

```css
.input {
  width: 100%;
  padding: 10px 12px;
  border: 1px solid var(--border-color);
  border-radius: 6px;
  font-size: var(--font-size-base);
  font-family: var(--font-family);
  background: var(--bg-primary);
  color: var(--text-primary);
  transition: border-color 0.2s ease, box-shadow 0.2s ease;
}

.input:focus {
  outline: none;
  border-color: var(--border-focus);
  box-shadow: 0 0 0 3px rgba(0, 120, 212, 0.1);
}

.input:invalid,
.input.error {
  border-color: var(--border-error);
}

.input:disabled {
  background: var(--bg-tertiary);
  color: var(--text-disabled);
  cursor: not-allowed;
}
```

### Радиокнопки

```css
.radio-group {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-md);
}

.radio-label {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
  cursor: pointer;
  padding: var(--spacing-sm);
  border-radius: 4px;
  transition: background 0.2s ease;
}

.radio-label:hover {
  background: var(--bg-tertiary);
}

.radio-label input[type="radio"] {
  width: 18px;
  height: 18px;
  cursor: pointer;
  accent-color: var(--primary-color);
}
```

---

## 🗓️ Календарь

### Выходные дни

```css
.calendar-weekend-header {
  color: var(--weekend-color);
  font-weight: var(--font-weight-semibold);
}

.calendar-weekend {
  color: var(--weekend-color);
  background-color: var(--weekend-bg);
  font-weight: var(--font-weight-medium);
}

.calendar-today {
  border: 2px solid var(--primary-color);
  border-radius: 4px;
}

.calendar-selected {
  background-color: var(--success-bg);
  color: var(--success-color);
  font-weight: var(--font-weight-semibold);
}
```

---

## 🎭 Состояния и анимации

### Transitions

```css
--transition-fast: 0.15s ease;
--transition-normal: 0.2s ease;
--transition-slow: 0.3s ease;
```

### Анимации

```css
/* Появление результата */
@keyframes fadeIn {
  from {
    opacity: 0;
    transform: translateY(-10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.result-card {
  animation: fadeIn 0.3s ease-in;
}

/* Индикатор загрузки */
@keyframes spin {
  to { transform: rotate(360deg); }
}

.loading {
  display: inline-block;
  width: 16px;
  height: 16px;
  border: 2px solid var(--primary-color);
  border-top-color: transparent;
  border-radius: 50%;
  animation: spin 0.6s linear infinite;
}
```

---

## 📱 Адаптивность

### Размеры окна

```css
/* Базовый размер */
--window-width: 800px;
--window-height: 650px;

/* Минимальные размеры */
--window-min-width: 600px;
--window-min-height: 500px;
```

### Breakpoints (для будущего расширения)

```css
@media (max-width: 600px) {
  .container {
    padding: var(--spacing-md);
  }
  
  .card {
    padding: var(--spacing-md);
  }
}
```

---

## ♿ Доступность

### ARIA атрибуты

```html
<!-- Кнопки -->
<button aria-label="Открыть календарь">📅</button>

<!-- Поля ввода -->
<input 
  type="text" 
  aria-label="Дата курса"
  aria-required="true"
  aria-invalid="false"
>

<!-- Сообщения об ошибках -->
<div role="alert" aria-live="polite">
  Дата не может быть в будущем
</div>
```

### Контрастность

- **Основной текст:** Минимум 4.5:1 (WCAG AA)
- **Вторичный текст:** Минимум 3:1 (WCAG AA для крупного текста)
- **Интерактивные элементы:** Минимум 3:1 (WCAG AA)

### Клавиатурная навигация

- Все интерактивные элементы доступны через Tab
- Enter активирует кнопки и подтверждает ввод
- Escape закрывает календарь
- Стрелки навигации в календаре

---

## 📐 Layout Grid

### Структура контейнера

```css
.container {
  max-width: var(--window-width);
  margin: 0 auto;
  padding: var(--spacing-xl);
  display: flex;
  flex-direction: column;
  gap: var(--spacing-lg);
}
```

### Карточки в layout

```css
.card {
  width: 100%;
  box-sizing: border-box;
}
```

---

## 🎯 Принципы использования

1. **Консистентность:** Всегда используйте переменные из этого style guide
2. **Иерархия:** Используйте размеры шрифтов и веса для создания визуальной иерархии
3. **Пространство:** Следуйте системе отступов для единообразия
4. **Цвета:** Используйте цвета только из палитры
5. **Анимации:** Применяйте плавные transitions для улучшения UX
6. **Доступность:** Всегда добавляйте ARIA атрибуты и поддерживайте клавиатурную навигацию

---

## 📚 Ссылки

- **Creative Phase Document:** `memory-bank/creative/creative-gui-design.md`
- **Wails Documentation:** https://wails.io/docs/
- **Windows 11 Design Guidelines:** Microsoft Fluent Design System

---

**Последнее обновление:** 2025-12-22

