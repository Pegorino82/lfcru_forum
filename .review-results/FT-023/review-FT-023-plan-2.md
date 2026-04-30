# Ревью implementation-plan.md (FT-023)

Дата: 2026-04-29
Источник: `memory-bank/features/FT-023/implementation-plan.md`
Проверено против: `feature.md`, `coding-style.md`, `testing-policy.md`, `autonomy-boundaries.md`, `feature-flow.md`

---

## CRITICAL

---

### C-01: `status: draft` вместо `active`

Строка 8 плана: `status: draft`.

`feature-flow.md` gate Design Ready → Plan Ready: `implementation-plan.md` → `status: active`. Если план принят в Plan Ready, `draft` — нарушение gate condition.

---

### C-02: Расхождение с REQ-04 по "при рендеринге"

`feature.md` REQ-04:
> "бэкенд санитизирует HTML через allowlist-политику (bluemonday) **при сохранении и при рендеринге**"

План реализует sanitize **только при сохранении** (STEP-02). STEP-03 и STEP-04 делают `template.HTML(article.Content)` без повторной санитизации.

OQ-07 объясняет только отклонение от шаблонной функции `safeHTML`, но не адресует расхождение "при рендеринге". Необходимо одно из:
- исправить REQ-04 в feature.md (если double-sanitize не нужен, потому что sanitize-at-write делает данные already safe), или
- добавить sanitize при рендеринге, или
- добавить явный OQ с обоснованием, почему sanitize-at-write покрывает это требование.

Сейчас — молчаливое несоответствие canonical требованию без трассировки.

---

### C-03: OQ-07 фиксирует отклонение от Change Surface, но не инициирует обновление `feature.md`

`feature.md` Change Surface включает `templates/news/article.html` как code-изменение. OQ-07 документирует, что шаблон менять не нужно. Это изменение Change Surface (`feature.md` → How).

`feature-flow.md` Boundary Rule 6:
> "Если меняются scope, architecture, acceptance criteria или evidence contract, сначала обновляется `feature.md` или ADR, потом downstream-план."

OQ-07 содержит "Default action: ... `templates/news/article.html` не менять; feature.md не менять — реализационная деталь" — но Change Surface это часть `How` в feature.md, не реализационная деталь плана. Нужно либо обновить feature.md (убрать строку шаблона из Change Surface), либо явно обосновать, почему Boundary Rule 6 не применяется.

---

## HIGH

---

### H-01: Отсутствует `AG-*` — STEP-00 не формализован как approval gate

`feature-flow.md` Boundary Rule 9:
> "Для рискованных, необратимых или внешне-эффективных действий `implementation-plan.md` должен явно описывать human approval gates и не скрывать их внутри prose шага."

`feature-flow.md` Stable Identifiers: `AG-*` | approval gates for risky actions | `implementation-plan.md`.

STEP-00 описывает HARD STOP как prose, но не использует `AG-*`. Переход в `delivery_status: in_progress` — внешне-эффективное действие (меняет статус lifecycle, видимый в feature.md). Должен быть `AG-01` с явным approval ref.

---

### H-02: OQ-05 — follow-up по `architecture.md` не привязан ни к STEP-*, ни к тикету

OQ-05: "Обновить `architecture.md` после перевода ADR-007 в `accepted`; оформить как follow-up в PR."

Нет ни шага, ни формализованного follow-up action, ни owner. Это висячее обязательство. Если closure gate не требует этого явно — action потеряется.

---

## MEDIUM

---

### M-01: URL в `feature.md` CHK-01 неверный

`feature.md` CHK-01: "открыть `/articles/{id}/edit`"
STEP-08 плана: "открыть `/admin/articles/{id}/edit`"

Discovery Context подтверждает `internal/admin/articles_handler.go` — корректный URL `/admin/articles/{id}/edit`. В feature.md ошибка. По Boundary Rule 6 — исправление в feature.md (canonical source) первично.

---

### M-02: Двусмысленность в Test Strategy колонке Manual-only gap

Строка для `static/js/editor.js`:
> `— (автопилот по autonomy-boundaries.md § UI-верификация)`

Пояснение в скобках объясняет *почему* нет manual-only gap, но стоит в колонке самого gap. Лучше: прочерк без пояснения, или перенести пояснение в колонку "Planned automated coverage".

---

### M-03: STEP-05 подключает файл, который создаётся в STEP-06

STEP-05: "подключить `static/js/editor.js` через `<script type="module">`"
STEP-06: "Создать `static/js/editor.js`"

PAR-02 правильно указывает последовательность, но формулировка STEP-05 создаёт ложное впечатление, что файл уже должен существовать при выполнении шага. Уточнить: "подключить тег `<script>` (файл создаётся в STEP-06)".

---

## LOW

---

### L-01: OQ-01 — намеренная разница терминологии не обоснована явно

OQ-01 фиксирует расхождение (`articles.body` vs `news.content`) и правильно говорит "feature.md не менять". Но не объясняет намеренную природу разницы: feature.md использует продуктовый термин (`body`), план — технический (`Content`). Стоит добавить это пояснение в OQ-01 для будущих агентов/авторов.

---

## Итоговая таблица

| ID | Серьёзность | Что делать |
|---|---|---|
| C-01 | Critical | Изменить `status: draft` → `active` в plan frontmatter |
| C-02 | Critical | Добавить OQ с обоснованием, почему sanitize-only-at-write покрывает REQ-04 "при рендеринге", или обновить REQ-04 в feature.md |
| C-03 | Critical | Обновить `feature.md` Change Surface (убрать строку `templates/news/article.html`) или обосновать в плане почему Boundary Rule 6 не применяется |
| H-01 | High | Добавить `AG-01` для STEP-00 с approval ref |
| H-02 | High | Добавить STEP в closure или явный follow-up тикет для OQ-05 |
| M-01 | Medium | Исправить URL в `feature.md` CHK-01 (сначала feature.md, потом план) |
| M-02 | Medium | Убрать пояснение из колонки Manual-only gap или перенести его |
| M-03 | Medium | Уточнить формулировку STEP-05 |
| L-01 | Low | Добавить пояснение в OQ-01 о намеренной разнице терминологии |
