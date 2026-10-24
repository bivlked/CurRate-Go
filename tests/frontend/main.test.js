// Тесты превью курсов и строки сравнения из frontend/scripts/main.js:
// гонки ответов GetRate при смене дат и валюты.
// Запуск: node --test tests/frontend/main.test.js
// main.js работает с DOM, поэтому подставляем минимальный фейковый DOM
// и GetRate, ответами которого тест управляет вручную
const { test } = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const vm = require('node:vm');

const scripts = path.join(__dirname, '..', '..', 'frontend', 'scripts');
const utilsSrc = fs.readFileSync(path.join(scripts, 'utils.js'), 'utf8');
const mainSrc = fs.readFileSync(path.join(scripts, 'main.js'), 'utf8');

function fakeElement() {
    const classes = new Set();
    const attrs = {};
    return {
        value: '',
        textContent: '',
        title: '',
        get className() { return [...classes].join(' '); },
        set className(v) { classes.clear(); v.split(/\s+/).filter(Boolean).forEach(c => classes.add(c)); },
        classList: {
            add: (c) => classes.add(c),
            remove: (c) => classes.delete(c),
            contains: (c) => classes.has(c),
            toggle: (c, force) => { (force ?? !classes.has(c)) ? classes.add(c) : classes.delete(c); }
        },
        setAttribute: (k, v) => { attrs[k] = String(v); },
        getAttribute: (k) => attrs[k]
    };
}

function setup() {
    const ids = ['date-input', 'ref-date-input', 'rate-preview', 'rate-value', 'ref-rate-preview',
        'ref-rate-value', 'ref-compare', 'ref-compare-text', 'ref-compare-pct-wrap', 'ref-compare-pct'];
    const els = Object.fromEntries(ids.map(id => [id, fakeElement()]));
    const currency = { value: 'USD' };
    const calls = []; // ответы GetRate, которые тест разрешает вручную

    const document = {
        getElementById: (id) => els[id] || null,
        querySelector: (sel) => (sel === 'input[name="currency"]:checked' ? currency : null),
        addEventListener: () => {}
    };
    const ctx = vm.createContext({ console, setTimeout, clearTimeout, document, window: {} });
    vm.runInContext(utilsSrc, ctx);
    vm.runInContext(mainSrc, ctx);
    ctx.__getRate = (cur, date) => new Promise(resolve => calls.push({ cur, date, resolve }));
    vm.runInContext(`
        appInstance = { GetRate: (c, d) => __getRate(c, d) };
        mainPreview = createRatePreview({ inputId: 'date-input', boxId: 'rate-preview', valueId: 'rate-value', stateKey: 'main', what: 'Курс' });
        refPreview = createRatePreview({ inputId: 'ref-date-input', boxId: 'ref-rate-preview', valueId: 'ref-rate-value', stateKey: 'ref', what: 'Реф' });
    `, ctx);

    els['date-input'].value = '01.10.2026';
    els['ref-date-input'].value = '02.10.2026';

    const run = (code) => vm.runInContext(code, ctx);
    // строка сравнения целиком, как её видит пользователь: текст и процент в скобках
    const compare = () => els['ref-compare-text'].textContent +
        (els['ref-compare-pct-wrap'].classList.contains('hidden') ? '' : ` (${els['ref-compare-pct'].textContent})`);
    return { els, currency, calls, run, compare };
}

// debounce превью - 300 мс; ждём срабатывания и разбора промисов
const afterDebounce = () => new Promise(r => setTimeout(r, 350));
const flush = () => new Promise(r => setImmediate(r));

function answer(call, rate) {
    call.resolve({ success: true, rate, actualDate: call.date });
}

test('смена валюты: курс, пришедший первым, не сравнивается со старым курсом другой валюты', async () => {
    const { els, currency, calls, run, compare } = setup();

    run('mainPreview.request(); refPreview.request();');
    await afterDebounce();
    assert.equal(calls.length, 2);
    answer(calls[0], 80);
    answer(calls[1], 80);
    await flush();
    assert.equal(compare(), 'Совпадает с референсом');

    // USD -> EUR: курс на дату приходит раньше референсного
    currency.value = 'EUR';
    run('mainPreview.request(); refPreview.request();');
    await afterDebounce();
    const [eurMain, eurRef] = calls.slice(2);
    assert.equal(eurMain.cur, 'EUR');
    answer(eurMain, 90);
    await flush();
    // раньше здесь было «Выше референса на 10,0000 ₽ (12,50%)»: EUR против USD
    assert.equal(compare(), 'Совпадает с референсом', 'сравнение не пересчитывается, пока референс грузится');
    assert.ok(els['ref-compare'].classList.contains('loading'));
    assert.equal(els['ref-compare'].getAttribute('aria-busy'), 'true');

    answer(eurRef, 85);
    await flush();
    assert.equal(compare(), 'Выше референса на 5,0000 ₽ (5,88%)');
    assert.ok(!els['ref-compare'].classList.contains('loading'));
    assert.equal(els['ref-compare'].getAttribute('aria-busy'), 'false');
});

test('ответ на устаревший запрос отбрасывается', async () => {
    const { els, calls, run, compare } = setup();

    run('mainPreview.request(); refPreview.request();');
    await afterDebounce();
    answer(calls[1], 80); // референс

    // пока первый запрос курса в пути, дата меняется
    els['date-input'].value = '05.10.2026';
    run('mainPreview.request();');
    await afterDebounce();
    answer(calls[2], 82); // актуальный ответ на 05.10
    await flush();
    answer(calls[0], 999); // запоздалый ответ на 01.10
    await flush();

    assert.equal(els['rate-value'].textContent, '82,0000 ₽');
    assert.equal(compare(), 'Выше референса на 2,0000 ₽ (2,50%)');
    assert.equal(els['ref-compare-pct'].textContent, '2,50%');
});

test('недописанная дата гасит курс и сравнение, ответ в пути игнорируется', async () => {
    const { els, calls, run } = setup();

    run('mainPreview.request(); refPreview.request();');
    await afterDebounce();
    answer(calls[0], 80);
    answer(calls[1], 80);
    await flush();

    els['ref-date-input'].value = '03.10.2026';
    run('refPreview.request();');
    await afterDebounce();
    els['ref-date-input'].value = '03.10.20';
    run('refPreview.invalidate();');
    answer(calls[2], 70); // ответ на 03.10 пришёл после того, как дату начали править
    await flush();

    assert.equal(els['ref-rate-value'].textContent, '…');
    assert.equal(els['ref-compare-text'].textContent, 'Сравнение недоступно');
    assert.ok(els['ref-compare-pct-wrap'].classList.contains('hidden'), 'без процента скобки скрыты');
});

test('ошибка GetRate: «н/д» и «Сравнение недоступно»', async () => {
    const { els, calls, run } = setup();

    run('mainPreview.request(); refPreview.request();');
    await afterDebounce();
    answer(calls[0], 80);
    calls[1].resolve({ success: false, error: 'ЦБ РФ не публиковал курсы на выбранную дату' });
    await flush();

    assert.equal(els['ref-rate-value'].textContent, 'н/д');
    assert.equal(els['ref-rate-preview'].title, 'ЦБ РФ не публиковал курсы на выбранную дату');
    assert.equal(els['ref-compare-text'].textContent, 'Сравнение недоступно');
});

test('процент в скобках скрывается, когда курсы снова совпали', async () => {
    const { els, calls, run, compare } = setup();

    run('mainPreview.request(); refPreview.request();');
    await afterDebounce();
    answer(calls[0], 81);
    answer(calls[1], 80);
    await flush();
    assert.equal(compare(), 'Выше референса на 1,0000 ₽ (1,25%)');
    assert.ok(!els['ref-compare-pct-wrap'].classList.contains('hidden'));

    els['date-input'].value = '02.10.2026';
    run('mainPreview.request();');
    await afterDebounce();
    answer(calls[2], 80);
    await flush();
    assert.equal(compare(), 'Совпадает с референсом');
    assert.ok(els['ref-compare-pct-wrap'].classList.contains('hidden'));
});
