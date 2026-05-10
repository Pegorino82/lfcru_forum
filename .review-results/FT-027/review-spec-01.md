# Spec Improve Loop — Review #01

- **Loop:** Spec Improve Loop
- **Artifact:** `/Users/evgenyshkryabin/study/lfcru_forum-FT-027/memory-bank/features/FT-027/feature.md` / `## How` + `## Verify`
- **Date:** 2026-05-07
- **Outcome:** revise

---

## Verdict Summary

| Check | Verdict | Details |
|---|---|---|
| A-1 | OK | Solution описывает конкретный подход: CSS Grid, карточный layout, адаптация handler |
| A-2 | MEDIUM | Trade-off назван неявно — нет явного описания отклонённой альтернативы |
| A-3 | OK | News модуль не имеет service-слоя (handler -> repo напрямую); feature не вводит бизнес-логику, только расширяет данные в запросе — допустимо |
| B-1 | OK | Все 5 путей существуют в репозитории (проверено через Glob) |
| B-2 | OK | `templates/news/list.html`, `templates/forum/index.html`, `templates/forum/section.html` — соответствуют `templates/<domain>/` |
| B-3 | OK | Feature не меняет static — CSS inline в шаблонах (CON-01), static не затрагивается |
| B-4 | OK | Все изменяемые поверхности перечислены |
| C-1 | OK | CTR-01 описывает расширение данных handler -> template, producer/consumer указаны |
| C-2 | MEDIUM | FM покрывает отсутствие изображения и пустой список, но нет FM для невалидного page number в пагинации |
| C-3 | OK | Нет ADR-зависимостей |
| D-1 | OK | Все REQ-01..05 прослеживаются к SC-01..05 |
| D-2 | OK | Каждый SC описывает наблюдаемый результат |
| D-3 | OK | Каждый SC связан с CHK через traceability matrix |
| D-4 | OK | Каждый CHK связан с EVID |
| E-1 | OK | Каждый CHK имеет Playwright-процедуру |
| E-2 | OK | Каждый EVID имеет конкретный path contract в `artifacts/ft-027/verify/` |
| E-3 | OK | Все UI-checks покрыты Playwright |
| E-4 | OK | Нет manual-only на основании HTMX/Alpine |
| E-5 | OK | Нет manual-only gaps |
| F-1 | OK | Feature использует только GET-запросы, CON-02 явно фиксирует это |
| F-2 | BLOCKER | Feature меняет пагинацию (DEC-02: page size 9) — нужен минимум 1 NEG-* для edge cases (невалидный page, page за пределами диапазона) |

---

## Findings

### Finding 1 — BLOCKER (F-2)

**Цитата:** Секция `## Verify` не содержит ни одного `NEG-*`.

**Норма:** `feature-flow.md` § Transition Gates, Draft -> Design Ready: "если deliverable нельзя принять без negative/edge coverage -> >= 1 `NEG-*`". `spec-improve-loop.md` § Exit Criteria: "если deliverable нельзя принять без negative coverage -> присутствует >= 1 `NEG-*`".

**Обоснование:** Feature изменяет page size пагинации (DEC-02) и расширяет данные handler (CTR-01). Edge cases пагинации (page=0, page=-1, page за пределами, page=NaN) влияют на корректность и должны быть покрыты. Без NEG-* невозможно подтвердить, что пагинация с новым page size обрабатывает граничные случаи.

**Исправление:**
1. Добавить в `## Verify` секцию `### Negative / Edge Scenarios` с минимум одним `NEG-*`:
   - `NEG-01` Открыть `/news?page=0` или `/news?page=-1` — сервер возвращает первую страницу (или 404), не падает с 500.
   - `NEG-02` Открыть `/news?page=999` (за пределами) — пустой список или редирект на последнюю страницу, не 500.
2. Добавить `NEG-*` в traceability matrix (связать с REQ-03).
3. Добавить CHK и EVID для NEG-сценариев (Playwright-проверка HTTP-статуса и поведения).

### Finding 2 — MEDIUM (A-2)

**Цитата:** Solution (строка 64): "Переработать шаблоны трёх страниц, взяв за визуальный эталон стиль главной..."

**Норма:** `spec-improve-loop.md` § Exit Criteria: "Solution описывает конкретный технический подход и главный trade-off".

**Обоснование:** Solution описывает подход (CSS Grid, inline styles, карточный layout), но не фиксирует явно trade-off или отклонённую альтернативу. Например: "CSS Grid выбран вместо Flexbox, потому что..." или "Inline styles (CON-01) вместо отдельного CSS-файла — потому что...".

**Исправление:** Добавить 1-2 предложения в Solution, явно назвав trade-off. Пример: "Альтернатива — вынести общие стили в shared CSS-файл; отклонена из-за CON-01 (все стили inline в шаблонах)."

### Finding 3 — MEDIUM (C-2)

**Цитата:** FM-01 (отсутствие изображения), FM-02 (пустой список разделов). Нет FM для невалидного page в пагинации.

**Норма:** `spec-improve-loop.md` § Exit Criteria: "если есть FM-* — покрыты критичные failure modes".

**Обоснование:** Пагинация с новым page size — изменённое поведение. Невалидные параметры (page=abc, page=-1) — предсказуемый failure mode. Хотя это пересекается с NEG-*, отдельный FM уточнил бы ожидаемое поведение handler при невалидном вводе.

**Исправление:** Добавить `FM-03`: "Невалидный параметр page (нечисловой, отрицательный, за пределами) -> handler использует page=1 по умолчанию."
