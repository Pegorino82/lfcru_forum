# Ревью implementation-plan.md (FT-023)

**Дата:** 2026-04-29
**Документ:** `memory-bank/features/FT-023/implementation-plan.md`
**Проверено на соответствие:** `feature.md`, `architecture.md`, `coding-style.md`, `testing-policy.md`, `feature-flow.md`, `autonomy-boundaries.md`

---

## Статус: требует исправлений

---

## Значимые проблемы (блокеры Plan Ready)

### 1. ER-04 конфликтует с ASM-01 без разрешения

**Где:** ER-04, STEP-06

`feature.md` ASM-01 гласит: "upload-endpoint используется без изменений". ER-04 и STEP-06 предлагают "парсить URL из img-тега **или** добавить `data-url` атрибут в `image_item.html`". Второй вариант — это изменение существующего шаблона ответа upload-endpoint, что прямо нарушает ASM-01.

**Что нужно:** Зафиксировать единственный выбранный подход. Если `data-url` — обновить ASM-01 в `feature.md` до начала исполнения (по правилу feature-flow § Boundary Rules 6: "сначала обновляется feature.md, потом план"). Если парсинг img-тега — убрать вариант с `image_item.html` как вводящий в заблуждение.

---

### 2. Тест для `news/handler.go` описан в Test Strategy, но не привязан ни к одному STEP

**Где:** Test Strategy (строка `news/handler.go — рендеринг HTML`), Порядок работ

Test Strategy декларирует "Добавить тест: статья с HTML body → GET `/news/{id}` → `ContentHTML` содержит ожидаемый HTML", но ни STEP-07, ни другие шаги этого не предусматривают. STEP-07 покрывает только `articles_handler_test.go`. Тест выпадет при исполнении — либо не будет написан, либо будет добавлен ad hoc.

**Что нужно:** Добавить явный sub-task в STEP-07 или выделить STEP-07b.

---

### 3. STEP-07 объединяет две несвязанные задачи

**Где:** STEP-07

Один шаг одновременно покрывает "добавить XSS-тест в `articles_handler_test.go`" и "добавить `e2e_admin` в `e2e/global-setup.ts`". Это разные файлы, разные concerns, разная проверка. При блокере в одной задаче статус шага становится неопределённым.

**Что нужно:** Разделить на STEP-07 (unit-тест XSS) и STEP-07b (e2e_admin в global-setup), или явно обосновать, почему они атомарны.

---

### 4. STEP-08/09 не описывают seeding тестовых данных

**Где:** STEP-08, STEP-09

Playwright тест открывает `/admin/articles/{id}/edit` — но откуда берётся `{id}`? `testing-policy.md` § E2E гласит: "тестовые данные вставляются с фиксированными ID через `OVERRIDING SYSTEM VALUE`; teardown чистит их по тому же ID". В плане этот паттерн не упомянут. Отсутствует указание на создание тестовой статьи в global-setup (или в самом тесте) и её teardown.

**Что нужно:** Добавить в STEP-07 (или STEP-08) явное описание seeding/teardown тестовой статьи по паттерну `testing-policy.md`.

---

## Средние проблемы

### 5. Колонка `Implements` смешивает разные классы идентификаторов

**Где:** Таблица «Порядок работ»

`feature-flow.md` § Stable Identifiers чётко разграничивает `REQ-*` (scope), `SC-*` (acceptance scenarios), `NEG-*` (negative test cases). В колонке `Implements` используются все три вперемешку (например, STEP-07: `CHK-03`, `SC-03`, `NEG-01`). По `feature-flow.md` § Traceability Contract п.3: "implementation-plan.md ссылается на canonical IDs из feature.md в колонках **Implements**, Verifies и Evidence IDs". Логичнее `SC-*`/`NEG-*` относить в отдельную колонку `Verifies` — либо договориться о смешанном использовании и зафиксировать это явно.

**Что нужно:** Либо добавить колонку `Verifies` в таблицу шагов для `SC-*`/`NEG-*`, либо добавить пояснение в преамбуле таблицы, что `Implements` покрывает оба класса.

---

### 6. OQ-07 описывает отклонение от `feature.md` Flow, но не предлагает обновить feature.md

**Где:** OQ-07

`feature.md` § Flow, строка 6: "`{{ .Body | safeHTML }}`" — шаблонная функция. OQ-07 корректно фиксирует, что план реализует через `template.HTML(...)` в handler. Но `feature-flow.md` § Boundary Rules 6 требует: "если меняются ... acceptance criteria — сначала обновляется feature.md, потом план". Реализационная деталь (`safeHTML` vs `template.HTML`) здесь не является scope-изменением, но расхождение в `feature.md § Flow` создаёт ложное ожидание для будущего читателя.

**Что нужно:** Либо добавить в Action OQ-07 явное "feature.md § Flow не обновляется — реализационная деталь, не scope", либо запланировать правку `feature.md § Flow` до closure (STEP-11).

---

### 7. `implementation-plan.md` имеет `status: draft`, момент смены не зафиксирован

**Где:** Frontmatter, строка 8

По `feature-flow.md`, Plan Ready gate требует `implementation-plan.md → status: active`. Текущий статус `draft`. STEP-00 описывает только смену `feature.md → delivery_status: in_progress`, но не момент смены статуса самого плана.

**Что нужно:** Зафиксировать явно: статус плана меняется на `active` при переводе в Plan Ready (до STEP-00). Добавить это как подпункт STEP-00 или в преамбулу плана.

---

## Незначительные замечания

### 8. STEP-05 содержит избыточную ссылку на STEP-06

**Где:** STEP-05, колонка Goal

Формулировка "(сам файл создаётся в STEP-06)" дублирует PAR-02. Убрать для краткости.

---

### 9. WS-2 зависимость не конкретизирована

**Где:** Таблица Workstreams, WS-2

"Dependencies: WS-1 (hidden input pattern)" — неясно, полная ли это зависимость (нужен завершённый WS-1) или только STEP-02.

**Что нужно:** Конкретизировать: "depends on STEP-02".

---

## Что консистентно и корректно

- **Environment Contract** полностью соответствует `testing-policy.md`: integration-тесты только в CI, флаг `-p 1`, правильные команды Docker.
- **OQ-02 (пустой TipTap → `<p></p>`)** — корректно задокументировано как known behavior без блокера.
- **OQ-06 (sanitize в handler вместо Service)** — обоснованное отклонение от layer stack с правильной трассировкой.
- **Allowlist в STEP-02** совпадает с `CTR-02` из `feature.md`.
- **PRE-01/PRE-02** корректно ссылаются на `ASM-*` из `feature.md`.
- **CP-* / ER-* / STOP-*** структурно валидны; STOP-* не требуют `AG-*` (это stop conditions, не manual-only gaps).
- **Нет AG-*** — корректно, т.к. все Manual-only gaps в Test Strategy = `—`.
- **coding-style.md** конфликта нет: OQ-07 объясняет, почему `template.HTML()` допустим (sanitize-at-write).

---

## Сводная таблица

| # | Находка | Приоритет |
|---|---------|-----------|
| 1 | ER-04 vs ASM-01: неразрешённый выбор подхода | Блокер |
| 2 | Тест для `news/handler.go` не привязан к STEP | Блокер |
| 3 | STEP-07 объединяет несвязанные задачи | Блокер |
| 4 | STEP-08/09 не описывают seeding тестовых данных | Блокер |
| 5 | `Implements` смешивает REQ-* / SC-* / NEG-* | Средний |
| 6 | OQ-07 не указывает судьбу расхождения в feature.md Flow | Средний |
| 7 | status: draft — момент смены не зафиксирован в плане | Средний |
| 8 | Избыточная ссылка на STEP-06 в STEP-05 | Незначительный |
| 9 | WS-2 зависимость не конкретизирована | Незначительный |
