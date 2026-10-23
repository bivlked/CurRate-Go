// Тесты чистых функций frontend/scripts/utils.js
// Запуск: node --test tests/frontend/
// utils.js - обычный браузерный скрипт с глобальными функциями, поэтому
// загружаем его в изолированный vm-контекст, а не через require
const { test } = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const vm = require('node:vm');

const source = fs.readFileSync(path.join(__dirname, '..', '..', 'frontend', 'scripts', 'utils.js'), 'utf8');

function loadUtils(globals = {}) {
    const context = vm.createContext({ ...globals });
    vm.runInContext(source, context);
    return context;
}

const u = loadUtils();

test('parseAmount: поддерживаемые форматы', () => {
    const cases = [
        ['1000', 1000],
        ['1 000', 1000],
        ['1,000.50', 1000.5],
        ['1.000,50', 1000.5],
        ['1000.50', 1000.5],
        ['1000,50', 1000.5],
        ['1,5', 1.5],
        ['1,000', 1000],
        ['0,500', 0.5],
        ['0.500', 0.5],
        ['1,234,567', 1234567],
        ['1.234.567', 1234567],
        ['1.234.567,89', 1234567.89],
        ['123,456,789.9', 123456789.9],
        // после 4+ цифр одиночный разделитель - дробная часть, а не тысячи
        ['1234,567', 1234.567],
        ['1234.567', 1234.567],
        ['-1,000', null],
    ];
    for (const [input, want] of cases) {
        assert.equal(u.parseAmount(input), want, input);
    }
});

test('parseAmount: опечатки и неверная группировка отвергаются', () => {
    for (const input of ['1..2', '12,34,56', '1,2,3', '12,34.5', '1..000,5', '1,000.000.5', '', '   ', 'abc', '-5', '0', '1e5', '1.2.3']) {
        assert.equal(u.parseAmount(input), null, JSON.stringify(input));
    }
});

test('isValidGrouping', () => {
    assert.equal(u.isValidGrouping('1234', ','), true);
    assert.equal(u.isValidGrouping('1,234', ','), true);
    assert.equal(u.isValidGrouping('12,345,678', ','), true);
    assert.equal(u.isValidGrouping('1234,567', ','), false);
    assert.equal(u.isValidGrouping('1,23', ','), false);
    assert.equal(u.isValidGrouping('1.234', '.'), true);
    assert.equal(u.isValidGrouping('1..234', '.'), false);
});

test('formatNumber: половина вверх по десятичной записи, как на backend', () => {
    // Intl округляет кратчайшую десятичную запись числа (1.005, а не
    // двоичное 1.00499...) половиной вверх - так же, как converter.formatNumber
    assert.equal(u.formatNumber(1.005, 2), '1,01');
    assert.equal(u.formatNumber(2.675, 2), '2,68');
    assert.equal(u.formatNumber(1.4649999999999999, 2), '1,46');
    assert.equal(u.formatNumber(807.23, 2), '807,23');
    assert.equal(u.formatNumber(80.7225, 4), '80,7225');
    assert.equal(u.formatNumber(NaN, 2), '0');
    // Разделитель тысяч в ru-RU - неразрывный пробел
    assert.equal(u.formatNumber(80722, 2).replace(/\s/g, ' '), '80 722,00');
});

test('isValidDateChar: выделенную дату можно перепечатать', () => {
    // При фокусе поле выделяет всю дату; ввод заменяет выделение
    assert.equal(u.isValidDateChar('23.10.2026', '1', 0, 10), true);
    // Без выделения восьмая цифра уже введена - ввод запрещён
    assert.equal(u.isValidDateChar('23.10.2026', '1', 10, 10), false);
    // Выделение заменяется, и проверка идёт по новой позиции: первая цифра дня 0-3
    assert.equal(u.isValidDateChar('23.10.2026', '5', 0, 10), false);
});

test('isValidDateChar: правила позиций', () => {
    assert.equal(u.isValidDateChar('', '4', 0), false); // день 40+
    assert.equal(u.isValidDateChar('3', '2', 1), false); // 32
    assert.equal(u.isValidDateChar('3', '1', 1), true); // 31
    assert.equal(u.isValidDateChar('12.', '2', 3), false); // месяц 2x
    assert.equal(u.isValidDateChar('12.1', '3', 4), false); // месяц 13
    assert.equal(u.isValidDateChar('12.12.', '3', 6), false); // год 3xxx
    assert.equal(u.isValidDateChar('', 'a', 0), false);
});

test('autoFormatDate и parseDate', () => {
    assert.equal(u.autoFormatDate('23102026'), '23.10.2026');
    assert.equal(u.autoFormatDate('2310'), '23.10');
    assert.equal(u.autoFormatDate('231020261'), '23.10.2026');
    const d = u.parseDate('29.02.2024');
    assert.equal(d.getDate(), 29);
    assert.equal(u.parseDate('29.02.2025'), null); // не високосный
    assert.equal(u.parseDate('31.04.2025'), null);
    assert.equal(u.isValidDateFormat('1.1.2025'), false);
});

function fakeClipboardEnv(execCommandResult) {
    const removed = [];
    const body = {
        appendChild() {},
        removeChild(node) { removed.push(node); },
    };
    return {
        removed,
        globals: {
            navigator: { clipboard: { writeText: () => Promise.reject(new Error('denied')) } },
            document: {
                body,
                createElement: () => ({ style: {}, select() {} }),
                execCommand: () => {
                    if (execCommandResult instanceof Error) throw execCommandResult;
                    return execCommandResult;
                },
            },
        },
    };
}

test('copyToClipboard: fallback сообщает о неудаче и убирает textarea', async () => {
    for (const [result, want] of [[true, true], [false, false], [new Error('boom'), false]]) {
        const env = fakeClipboardEnv(result);
        const ctx = loadUtils(env.globals);
        assert.equal(await ctx.copyToClipboard('x'), want, String(result));
        assert.equal(env.removed.length, 1, 'textarea должна быть удалена');
    }
});

test('compareRates: точная разность в единицах 0,0001', () => {
    // float: 80.5381 - 79.9012 = 0.6368999999999971
    const up = u.compareRates(80.5381, 79.9012);
    assert.equal(up.direction, 'up');
    assert.equal(up.diff, 0.6369);
    assert.ok(Math.abs(up.percent - 0.6369 / 79.9012 * 100) < 1e-9);

    const down = u.compareRates(79.9012, 80.5381);
    assert.equal(down.direction, 'down');
    assert.equal(down.diff, 0.6369);

    const eq = u.compareRates(80.5381, 80.5381);
    assert.equal(eq.direction, 'equal');
    assert.equal(eq.diff, 0);
    assert.equal(eq.percent, 0);
});

test('compareRates: нечего сравнивать - null', () => {
    for (const [a, b] of [[NaN, 80], [80, NaN], [0, 80], [80, 0], [-1, 80], [Infinity, 80]]) {
        assert.equal(u.compareRates(a, b), null, `${a} vs ${b}`);
    }
});

test('formatPercent: 2 знака, малое - «менее 0,01%»', () => {
    assert.equal(u.formatPercent(0.797), '0,80%');
    assert.equal(u.formatPercent(12.345), '12,35%');
    assert.equal(u.formatPercent(0), '0,00%');
    assert.equal(u.formatPercent(0.0012), 'менее 0,01%');
});

// Объекты из vm-контекста имеют чужой прототип: сравниваем содержимое
const plain = (o) => JSON.parse(JSON.stringify(o));

test('describeRateComparison: текст и цвет строки сравнения', () => {
    assert.deepEqual(plain(u.describeRateComparison(null)), { className: 'compare-none', text: 'Сравнение недоступно' });
    assert.deepEqual(plain(u.describeRateComparison(u.compareRates(80, 80))), { className: 'compare-equal', text: 'Совпадает с референсом' });

    const up = u.describeRateComparison(u.compareRates(80.5381, 79.9012));
    assert.equal(up.className, 'compare-up');
    assert.equal(up.text, 'Выше референса на 0,6369 ₽ (0,80%)');

    const down = u.describeRateComparison(u.compareRates(79.9012, 80.5381));
    assert.equal(down.className, 'compare-down');
    assert.equal(down.text, 'Ниже референса на 0,6369 ₽ (0,79%)');

    // разница в одну десятитысячную не выглядит как 0,00%
    assert.equal(u.describeRateComparison(u.compareRates(80.0001, 80)).text, 'Выше референса на 0,0001 ₽ (менее 0,01%)');
});

test('describeAmountComparison: строка результата', () => {
    const nbsp = (s) => s.replace(/\s/g, ' ');
    const more = u.describeAmountComparison(636.9, 79901.2);
    assert.equal(more.className, 'compare-up');
    assert.equal(nbsp(more.text), 'На 636,90 ₽ больше референса (79 901,20 ₽)');

    const less = u.describeAmountComparison(-4.62, 810);
    assert.equal(less.className, 'compare-down');
    assert.equal(nbsp(less.text), 'На 4,62 ₽ меньше референса (810,00 ₽)');

    const eq = u.describeAmountComparison(0, 810);
    assert.equal(eq.className, 'compare-equal');
    assert.equal(nbsp(eq.text), 'Совпадает с суммой по референсу (810,00 ₽)');
});
