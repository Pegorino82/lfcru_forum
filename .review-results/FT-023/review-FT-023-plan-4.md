# Ревью implementation-plan.md (FT-023)

**Дата:** 2026-04-29
**Документ:** `memory-bank/features/FT-023/implementation-plan.md`
**Критерии:** консистентность, отсутствие двусмысленности, соответствие memory-bank

**Итог:** 4 замечания уровня HIGH, 3 — MEDIUM, 3 — LOW

---

## HIGH

### H-01: `status: draft` нарушает план-Ready гейт feature-flow.md

**Проблема.** `feature-flow.md` § Design Ready → Plan Ready: предикат — `implementation-plan.md → status: active`. Это предикат гейта Plan Ready, то есть должен быть выставлен при подтверждении плана человеком. Текущий `status: draft` означает, что документ ещё не прошёл Plan Ready. Но STEP-00 трактует смену статуса как действие Execution-старта ("до первого коммита с кодом"), что смешивает два гейта (Plan Ready и Execution).

**Следствие.** Агент, следующий плану буквально, войдёт в Execution с `status: draft`, что технически нарушает gate predicate.

**Рекомендация.** STEP-00 должен разделить действия по гейтам явно:
- `implementation-plan.md → status: active` — выполняется при подтверждении человеком (Plan Ready gate, до Execution).
- `feature.md → delivery_status: in_progress` — выполняется в начале Execution.

Или зафиксировать в STEP-00 явное объяснение: "status: active устанавливается здесь как первое действие Execution — до первого коммита; это эквивалентно Plan Ready confirmation т.к. никакой работы ещё не выполнено."

---

### H-02: Allowlist не покрывает `<p style="text-align:...">` — TipTap TextAlign применяет стиль к `<p>`, не к `<div>`

**Проблема.** CTR-02 в feature.md и STEP-02 описывают allowlist с `<div style="text-align:*">`. TipTap extension `@tiptap/extension-text-align` применяет `style="text-align: ..."` к paragraph-ноду (`<p>`), не к `<div>`. Если bluemonday allowlist разрешает `style` только для `<div>`, выравнивание будет обрезаться при сохранении.

**Следствие.** SC-01 ("center-выравнивание") будет систематически терять форматирование — ER-02 сценарий без шанса на исправление в рамках описанного allowlist. CHK-01 будет красным.

**Рекомендация.** STEP-02 должен явно включить `<p style>` (или более широко — параграфные элементы) в allowlist. Это также требует уточнения CTR-02 в feature.md (boundary rule 6: feature.md обновляется первым).

---

### H-03: `TestShowArticle_HTMLBody` — двусмысленное утверждение теста

**Проблема.** STEP-07: "GET `/news/{id}` → assert `ContentHTML` содержит ожидаемый HTML". `ContentHTML` — имя поля Go-структуры (`ArticleData.ContentHTML template.HTML`), оно не присутствует в HTTP-ответе. В HTTP-ответе — рендеренный HTML шаблона. Формулировка смешивает уровень структуры данных и уровень HTTP-ответа.

**Следствие.** Агент может написать тест, который проверяет не то: либо попытается извлечь поле из JSON (которого нет), либо неправильно организует assertion.

**Рекомендация.** Переформулировать: "GET `/news/{id}` → тело HTTP-ответа содержит ожидаемые HTML-теги (например, `<strong>`, `<h2>`)", или "ответ handler содержит `ContentHTML: template.HTML(...)` — проверить через unit-тест на уровне data struct без HTTP-запроса."

---

### H-04: Сборка тестов — отсутствует build tag для `TestShowArticle_HTMLBody`

**Проблема.** STEP-07 добавляет `TestArticlesHandler_XSSSanitization` в `articles_handler_test.go` (файл с `//go:build integration`) — это integration-тест. Параллельно добавляется `TestShowArticle_HTMLBody` в `internal/news/handler_test.go` — build tag не указан. Test Strategy говорит "Required local suites: Unit (`go test ./...`)", что подразумевает unit-тест без тега. Но если `handler_test.go` тоже помечен `//go:build integration` (что типично для handler-тестов, требующих БД), тест не запустится локально без флага `-tags integration`.

**Следствие.** CP-03 ("Unit-тесты зелёные локально") может быть недостижимым, если тест фактически integration.

**Рекомендация.** STEP-07 должен явно указать: либо тест unit (мок-репозиторий) и объяснить как мокировать, либо integration (с `//go:build integration`) и скорректировать Test Strategy — "Required local suites: —; Required CI: integration".

---

## MEDIUM

### M-01: Preview change в feature.md Change Surface описан неполно — тихое расширение scope

**Проблема.** `feature.md` § Change Surface для `internal/admin/articles_handler.go`: "Добавить bluemonday-санитизацию поля content при save (Create/Update)". STEP-04 добавляет замену `RenderMarkdown → template.HTML` в `Preview` — это изменение поведения, не упомянутое явно в Change Surface feature.md.

**Feature-flow § Boundary Rules 6:** "Если меняются scope... сначала обновляется `feature.md`".

**Следствие.** Изменение preview-рендеринга фактически расширяет Change Surface без обновления canonical документа.

**Рекомендация.** Добавить в STEP-04 явную отсылку к тому, что это следствие из feature.md § Solution ("Рендеринг статьи — `template.HTML(article.Content)`"), или добавить уточнение в Change Surface feature.md перед Execution.

---

### M-02: ER-02, ER-03, ER-04 не имеют соответствующих STOP-условий

**Проблема.** Секция `Stop Conditions` покрывает только ER-01 (CDN недоступен → STOP-01) и CHK-01/CHK-02 (3 итерации → STOP-02). ER-02 (bluemonday обрезает нужные теги), ER-03 (Playwright не может взаимодействовать с TipTap), ER-04 (HTMX response parsing) — не имеют stop conditions или escalation threshold.

**Следствие.** Агент не знает, когда прекратить итерации при блокере ER-02/ER-03/ER-04 — может войти в бесконечный цикл (нарушение `autonomy-boundaries.md` § Правило эскалации: "2–3 итерации → остановись и предложи вернуться").

**Рекомендация.** Добавить STOP-03: "ER-02 (allowlist) не устранён за 2 итерации → зафиксировать блокер и эскалировать к human". Аналогично для ER-03 и ER-04 или включить их в единый STOP-03.

---

### M-03: STEP-07b — реквизиты `e2e_admin` не специфицированы

**Проблема.** STEP-07b: "Добавить `e2e_admin` (role=admin) в `e2e/global-setup.ts`" — не указан пароль или credential-паттерн. STEP-08 предполагает "войти как e2e_admin", но без пароля тест написать невозможно.

**Следствие.** Агент вынужден угадывать или изобретать credential-схему — риск divergence с существующим паттерном `e2e_user` из `global-setup.ts`.

**Рекомендация.** Добавить в STEP-07b: "использовать тот же паттерн credentials, что у `e2e_user` в `global-setup.ts` — например `e2e_admin` / `e2e_admin_password`; зеркалировать полностью".

---

## LOW

### L-01: STEP-00 — конфликт описания: "Plan Ready gate" внутри Execution HARD STOP

**Проблема.** STEP-00 в колонке `Implements` пишет "Plan Ready→Execution" и "(Plan Ready gate)" в скобках рядом с action. Это создаёт путаницу: HARD STOP описывает действия Execution, но "Plan Ready gate" — предшествующий гейт. Агент может интерпретировать это как "сначала выполнить гейт Plan Ready как часть Execution" вместо "убедиться, что гейт был пройден ранее".

**Рекомендация.** Переименовать скобку: не "(Plan Ready gate)" а "(подтверждение перехода — если не выполнено ранее)".

---

### L-02: WS-2 "Dependencies: STEP-02 (hidden input pattern)" — неясная зависимость

**Проблема.** В таблице Workstreams WS-2 указано: "Dependencies: STEP-02 (hidden input pattern)". Что именно из STEP-02 является паттерном для WS-2 — неочевидно. STEP-02 — это sanitization в Go-хендлере, а "hidden input pattern" относится к STEP-05 (шаблон).

**Рекомендация.** Уточнить: "STEP-05 (hidden input `name=\"content\"` определяет, что sync из TipTap → hidden field должен быть реализован в editor.js до submit)" или убрать пояснение в скобках, оставив просто "STEP-05".

---

### L-03: OQ-04 и ER-01 описывают один риск дважды без явной связи

**Проблема.** OQ-04 ("TipTap ESM CDN доступность") имеет статус "Гипотетическая проблема" и "Default action: Проверить на STEP-05". Колонка `Blocks` — "—". Фактически вопрос не разрешён, а deferred. При этом ER-01 корректно описывает тот же риск и имеет STOP-01. Создаётся дублирование: один и тот же риск описан дважды (OQ-04 и ER-01) с разным статусом.

**Рекомендация.** Либо закрыть OQ-04 с явной ссылкой "escalation pathway зафиксирован в ER-01 / STOP-01", либо удалить OQ-04 и оставить только ER-01.

---

## Итоговая таблица

| ID | Уровень | Суть |
|---|---|---|
| H-01 | HIGH | `status: draft` нарушает gate predicate Plan Ready; STEP-00 смешивает два гейта |
| H-02 | HIGH | Allowlist не покрывает `<p style>` — TipTap TextAlign использует `<p>`, не `<div>` |
| H-03 | HIGH | `TestShowArticle_HTMLBody` — assertion на `ContentHTML` двусмысленна (struct поле vs HTTP body) |
| H-04 | HIGH | Build tag для `TestShowArticle_HTMLBody` не указан; тип теста (unit/integration) не определён |
| M-01 | MEDIUM | Preview change не отражён в feature.md Change Surface — тихое расширение scope |
| M-02 | MEDIUM | ER-02/ER-03/ER-04 без STOP-условий; риск бесконечных итераций |
| M-03 | MEDIUM | `e2e_admin` credentials не специфицированы в STEP-07b |
| L-01 | LOW | STEP-00 пишет "(Plan Ready gate)" внутри Execution HARD STOP — терминологическая путаница |
| L-02 | LOW | WS-2 "Dependencies: STEP-02 (hidden input pattern)" — неверная или непонятная ссылка |
| L-03 | LOW | OQ-04 и ER-01 описывают один риск дважды без явной связи между ними |

**Блокеры для перехода в Execution:** H-01 (gate integrity), H-02 (рискует сломать SC-01), H-04 (рискует сломать CP-03). H-03 и M-* можно устранить параллельно с Execution, но лучше до старта.
