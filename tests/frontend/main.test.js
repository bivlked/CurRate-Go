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
        'ref-rate-value', 'ref-compare', 'ref-compare-text', 'ref-compare-pct-wrap', 'ref-compare-pct',
        'amount-input', 'convert-btn', 'result-card', 'result-text', 'result-amount', 'result-meta',
        'result-ref', 'result-ref-text'];
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
        mainPreview = createRatePreview({ inputId: 'date-input', boxId: 'rate-preview', valueId: 'rate-value', stateKey: 'main', dateLabel: '' });
        refPreview = createRatePreview({ inputId: 'ref-date-input', boxId: 'ref-rate-preview', valueId: 'ref-rate-value', stateKey: 'ref', dateLabel: 'референсную дату ' });
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

test('подсказка у курса: дата и курс прошлого рабочего дня на выходной', async () => {
    const { els, calls, run } = setup();

    run('mainPreview.request(); refPreview.request();');
    await afterDebounce();
    answer(calls[0], 80);
    calls[1].resolve({ success: true, rate: 80, actualDate: '01.10.2026' });
    await flush();

    assert.equal(els['rate-preview'].title, 'Курс ЦБ на 01.10.2026');
    assert.equal(els['ref-rate-preview'].title,
        'Курс ЦБ за 01.10.2026: на референсную дату 02.10.2026 курс не устанавливался');

    els['ref-date-input'].value = '01.10.2026';
    run('refPreview.request();');
    await afterDebounce();
    answer(calls[2], 80);
    await flush();
    assert.equal(els['ref-rate-preview'].title, 'Курс ЦБ на референсную дату 01.10.2026');
});

// Ответ Convert с референсом: 1000 USD, курс 81,5381, референс 80,9012
function convertResponse(extra) {
    return Object.assign({
        success: true, result: '81 538,10 руб. ($1 000,00 по курсу 81,5381)',
        targetAmountRUB: 81538.1, rate: 81.5381, sourceAmount: 1000, currencySymbol: '$',
        requestedDate: '05.10.2026', actualDate: '03.10.2026',
        hasReference: true, refRate: 80.9012, refAmountRUB: 80901.2, refDiffRUB: 636.9,
        refRequestedDate: '04.10.2026', refActualDate: '03.10.2026', refError: ''
    }, extra);
}

test('renderResult: строка сравнения суммы, её подсказка и скрытие', () => {
    const { els, run } = setup();
    const render = (resp) => { run(`renderResult(${JSON.stringify(resp)})`); };
    const ref = els['result-ref'];
    const nbsp = (str) => str.replace(/\s/g, ' ');

    render(convertResponse());
    assert.ok(!ref.classList.contains('hidden'));
    assert.ok(ref.classList.contains('compare-up'));
    assert.equal(nbsp(els['result-ref-text'].textContent), 'На 636,90 ₽ больше референса (80 901,20 ₽)');
    assert.equal(ref.title, 'Референс 04.10.2026: курс 80,9012 ₽ (курс ЦБ за 03.10.2026)');
    assert.equal(els['result-text'].textContent, '81 538,10 руб. ($1 000,00 по курсу 81,5381)');

    // референс совпадает с датой курса: сравнивать нечего
    render(convertResponse({ refRequestedDate: '05.10.2026' }));
    assert.ok(ref.classList.contains('hidden'));

    // референс задан, но не посчитался: строка не пропадает молча
    render(convertResponse());
    render(convertResponse({ hasReference: false, refError: 'Не удалось получить курсы ЦБ РФ' }));
    assert.ok(!ref.classList.contains('hidden'));
    assert.ok(ref.classList.contains('compare-none'));
    assert.equal(els['result-ref-text'].textContent, 'Сравнение с референсом недоступно');
    assert.equal(ref.title, 'Не удалось получить курсы ЦБ РФ');

    // референс не передавался вовсе
    render(convertResponse({ hasReference: false, refRequestedDate: '' }));
    assert.ok(ref.classList.contains('hidden'));
});

test('performConvert: недописанная референсная дата не уходит в Convert', async () => {
    const { els, run } = setup();
    // строка состояния (status-bar.js) здесь не нужна
    run(`
        showError = () => {}; showSuccess = () => {}; clearStatus = () => {};
        sentReqs = [];
        appInstance.Convert = (req) => { sentReqs.push(req); return Promise.resolve({ success: false, error: 'x' }); };
    `);
    els['amount-input'].value = '1000';
    els['date-input'].value = '05.10.2026';

    els['ref-date-input'].value = '04.10.20';
    await run('performConvert()');
    els['ref-date-input'].value = '04.10.2026';
    await run('performConvert()');

    const sent = run('sentReqs');
    assert.equal(sent.length, 2);
    assert.equal(sent[0].refDate, '');
    assert.equal(sent[1].refDate, '04.10.2026');
});
