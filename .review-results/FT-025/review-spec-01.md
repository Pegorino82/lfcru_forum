---
loop: Spec Improve Loop
artifact: memory-bank/features/FT-025/feature.md / ## How + ## Verify
date: 2026-05-04
outcome: revise
---

# Review: Spec Improve Loop — FT-025

## Outcome: revise

---

## Замечания

### 1. [MEDIUM] A-2 — Отсутствует явный trade-off или отклонённая альтернатива

**Цитата из ## How / Solution:**
> Для header: чисто шаблонное изменение `base.html`… Для форума и комментариев: добавляем поле `AuthorAvatarURL *string` в `PostView` и `CommentView`, расширяем SQL-запросы LEFT JOIN с `users`, обновляем шаблоны.

**Норма:** `feature-flow.md` § «Boundary Rules» п. 3: `feature.md` фиксирует design decisions; `feature-flow.md` § «Stable Identifiers» предусматривает `DEC-*` для blocking decisions. Для `large.md` Solution обязан называть главный trade-off или отклонённую альтернативу — это требование проверки A-2 настоящего цикла.

**Исправление:** добавить в Solution один из вариантов:
- явно указать отклонённую альтернативу (например: lazy-load аватарок через отдельный HTMX-запрос отклонён в пользу JOIN, чтобы избежать N+1 клиентских запросов);
- или добавить `DEC-01` с trade-off: "JOIN vs lazy HTMX fetch — выбран JOIN для атомарности рендера".

---

### 2. [MEDIUM] B-4 — `internal/tmpl/renderer.go` отсутствует в Change Surface при наличии `ASM-02`

**Цитата из ## How / Constraints:**
> `ASM-02` FuncMap функции `avatarColor` и `avatarInitials` уже зарегистрированы в `tmpl/renderer.go` и доступны во всех шаблонах.

**Цитата из ## How / Change Surface:** файл `internal/tmpl/renderer.go` отсутствует в таблице.

**Норма:** `feature-flow.md` § «Boundary Rules» п. 1–2: Change Surface обязан отражать все поверхности, которые фактически изменятся по REQ-*. Если ASM-02 не подтвердится при grounding и FuncMap придётся дополнять — это изменение окажется вне Change Surface без фиксации.

**Исправление:** добавить строку в Change Surface:

| `internal/tmpl/renderer.go` | code | Проверить наличие `avatarColor`/`avatarInitials` в FuncMap (ASM-02); при отсутствии — зарегистрировать |

Альтернатива: если ASM-02 будет верифицирован при grounding и файл точно не изменится — добавить `NT-01` (do-not-touch) с явной ссылкой на результат grounding.

---

### 3. [LOW] E-2 — Evidence path неспецифичен

**Цитата из ## Verify / Evidence contract:**
> `EVID-01..04`: Path contract — `e2e/test-report/`

**Норма:** `testing-policy.md` § «E2E-тесты (Playwright)»: артефакты — `e2e/test-results/` (скриншоты при падении), `e2e/test-report/` (HTML-отчёт). Требование E-2 настоящего цикла: каждый `EVID-*` имеет конкретный path contract (не "где-нибудь").

**Исправление:** уточнить Path contract для каждого EVID-*. Например:
- `EVID-01`: `e2e/test-report/index.html` → тест `avatar in nav`
- `EVID-02`: `e2e/test-report/index.html` → тест `avatar in forum post`

Минимально приемлемо: добавить имя спек-файла (например `e2e/profile/avatar.spec.ts`) как часть path contract.

---

## Проверки без замечаний

| Проверка | Вердикт |
| --- | --- |
| A-1 Solution конкретен, не повторяет REQ-* | OK |
| A-3 Бизнес-логика в Service-слое | OK (read-path, бизнес-логики нет) |
| B-1 Пути Change Surface существуют | OK (все 7 путей проверены) |
| B-2 Пути шаблонов соответствуют `templates/<domain>/` | OK |
| B-3 Пути статики соответствуют `static/js/`, `static/css/` | OK (static не затрагивается) |
| C-1 CTR-* для изменённых contracts | OK (нет новых contracts) |
| C-2 Failure modes: auth, data corruption, XSS | OK (html/template auto-escaping; FM-01 покрывает NULL) |
| C-3 ADR с proposed status | OK (нет зависимостей) |
| D-1 REQ-* → SC-* через traceability matrix | OK |
| D-2 SC-* описывают наблюдаемый результат | OK |
| D-3 SC-* → CHK-* | OK |
| D-4 CHK-* → EVID-* | OK |
| E-1 CHK-* с конкретной процедурой | OK |
| E-3 UI-изменения покрыты Playwright | OK |
| E-4 HTMX/Alpine.js не обоснование manual-only | OK |
| E-5 Manual-only gaps | OK (нет manual-only) |
| F-1 CSRF для POST/PUT/DELETE | OK (только GET modal) |
| F-2 NEG-* при необходимости | OK (NS-03 исключает гостей; SC-04 покрывает NULL-fallback) |
