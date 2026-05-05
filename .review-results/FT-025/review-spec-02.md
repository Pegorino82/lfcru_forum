---
loop: Spec Improve Loop
iteration: 2
artifact: memory-bank/features/FT-025/feature.md / ## How + ## Verify
date: 2026-05-04
outcome: accept
---

# Review: Spec Improve Loop — FT-025 / Итерация 2

## Outcome: accept

EVID-SP-01: Spec loop — accept. 2026-05-04. improve-loop.sh / evaluator agent

---

## Проверка замечаний итерации 1

### Замечание 1 (MEDIUM / A-2) — устранено

В `## How / Solution` добавлен явный trade-off (строки 62–64):
> **Trade-off:** JOIN в SQL vs lazy HTMX-fetch аватарки по клику. Выбран JOIN: аватарка нужна при первом рендере страницы…

Добавлен `DEC-01` с явным решением и отклонённой альтернативой. **FIXED.**

### Замечание 2 (MEDIUM / B-4) — устранено

В Change Surface добавлена строка `internal/tmpl/renderer.go` с явным описанием: проверить при grounding, изменений не требуется (ASM-02). **FIXED.**

### Замечание 3 (LOW / E-2) — устранено

Evidence path конкретизирован до `e2e/test-report/index.html` с именами тестов для каждого EVID-01..04. **FIXED.**

---

## Полный проход проверок (итерация 2)

### A. Solution

| Проверка | Вердикт | Примечание |
| --- | --- | --- |
| A-1 Solution конкретен, не повторяет REQ-* | OK | Описан технический механизм: шаблонное изменение + JOIN + расширение моделей |
| A-2 Trade-off или отклонённая альтернатива | OK | JOIN vs lazy HTMX-fetch, DEC-01 присутствует |
| A-3 Бизнес-логика в Service-слое | OK | Read-path без бизнес-логики; JOIN в repo, рендер в шаблоне — соответствует архитектуре |

### B. Change Surface

| Проверка | Вердикт | Примечание |
| --- | --- | --- |
| B-1 Пути существуют | OK | Все 8 путей проверены: `templates/layouts/base.html`, `templates/forum/partials/post.html`, `templates/news/article.html`, `internal/forum/model.go`, `internal/forum/repo.go`, `internal/comment/model.go`, `internal/comment/repo.go`, `internal/tmpl/renderer.go` — все существуют |
| B-2 Пути шаблонов: `templates/<domain>/` | OK | layouts/base.html, forum/partials/post.html, news/article.html — соответствуют |
| B-3 Пути статики | OK | Статика не затрагивается |
| B-4 REQ-* поверхности покрыты Change Surface | OK | renderer.go добавлен; все поверхности по REQ-01..04 явно перечислены |

### C. Contracts и Failure Modes

| Проверка | Вердикт | Примечание |
| --- | --- | --- |
| C-1 CTR-* для изменённых contracts | OK | Нет новых API/event/schema/env contracts; CON-02 фиксирует расширение модели |
| C-2 Критичные failure modes | OK | FM-01 (NULL avatar_url → инициалы); html/template auto-escaping исключает XSS; FM-02 явно выносится за scope |
| C-3 ADR с proposed | OK | Нет внешних ADR-зависимостей |

### D. Traceability

| Проверка | Вердикт | Примечание |
| --- | --- | --- |
| D-1 REQ-* → SC-* через matrix | OK | REQ-01→SC-01, REQ-02→SC-02, REQ-03→SC-03, REQ-04→SC-04 |
| D-2 SC-* описывают наблюдаемый результат | OK | Каждый SC-* содержит конкретный UI-результат (видна аватарка/инициалы, открывается модалка) |
| D-3 SC-* → CHK-* | OK | SC-01→CHK-01, SC-02→CHK-02, SC-03→CHK-03, SC-04→CHK-04 |
| D-4 CHK-* → EVID-* | OK | CHK-01→EVID-01, CHK-02→EVID-02, CHK-03→EVID-03, CHK-04→EVID-04 |

### E. Checks и Evidence

| Проверка | Вердикт | Примечание |
| --- | --- | --- |
| E-1 CHK-* с командой/процедурой | OK | Каждый CHK-* содержит Playwright-команду с testid и assert |
| E-2 EVID-* с конкретным path contract | OK | `e2e/test-report/index.html` + имя спека + имя теста для каждого EVID-* |
| E-3 UI-изменения покрыты Playwright | OK | 4 Playwright-спека для 3 поверхностей + fallback-сценарий |
| E-4 HTMX/Alpine не обоснование manual-only | OK | Нет manual-only; UI-проверки автоматизированы |
| E-5 Manual-only gap | OK | Отсутствует |

### F. Системные ограничения

| Проверка | Вердикт | Примечание |
| --- | --- | --- |
| F-1 CSRF для POST/PUT/DELETE | OK | Только GET-запросы (modal open); PCON-02 не нарушается |
| F-2 NEG-* | OK | NS-03 явно исключает гостей; SC-04 покрывает NULL-fallback; отдельный NEG-* не требуется |

---

## Итог

Все три замечания итерации 1 устранены. Все 18 проверок — OK. Блокеров нет.
