---
title: "FT-027: Implementation Plan"
doc_kind: feature
doc_function: derived
purpose: "Execution-план реализации FT-027. Фиксирует discovery context, шаги, риски и test strategy без переопределения canonical feature-фактов."
derived_from:
  - feature.md
status: active
audience: humans_and_agents
must_not_define:
  - ft_027_scope
  - ft_027_architecture
  - ft_027_acceptance_criteria
  - ft_027_blocker_state
---

# План имплементации

При переходе Plan Ready → Execution: `feature.md → delivery_status: in_progress`, `implementation-plan.md → status: active`. Выполняется как `PRE-04` до первого коммита с кодом.

## Цель текущего плана

Привести страницы `/news`, `/forum`, `/forum/sections/X` к единому визуальному стилю с главной страницей: карточный layout, адаптивная сетка новостей, единая типографика и цветовая схема.

## Discovery Context / Reference Points

| Path / module | Current role | Why relevant | Reuse / mirror |
| --- | --- | --- | --- |
| `internal/news/handler.go:66-101` | ShowList — пагинированный список новостей (pageSize=20, ListData с []News) | Handler нужно адаптировать: сменить тип Items на []HomeNewsItem, pageSize на 9 | Паттерн пагинации (page validation, totalPages, HasPrev/HasNext) сохранить |
| `internal/news/repo.go:84-116` | ListPublished — SELECT id, title, published_at | Возвращает только 3 поля — нужно расширить до cover_image, content, comment_count | Зеркалить SQL из LatestPublishedForHome (строки 44-82): JOIN article_images, subquery comment_count |
| `internal/news/repo.go:44-82` | LatestPublishedForHome — полный набор полей с image и comment_count | Эталонный SQL-запрос, который нужно адаптировать для пагинированного списка | Скопировать JOIN и subquery, добавить COUNT(*) для total и LIMIT/OFFSET |
| `internal/news/model.go:38-56` | HomeNewsItem struct + ExcerptText() | Структура с нужными полями (ID, Title, Content, CoverImageURL, CommentCount, PublishedAt) | Использовать HomeNewsItem вместо News в ListData.Items |
| `templates/home/index.html:6-49` | CSS стили карточек на главной (.news-item, .news-hero, .news-placeholder, .news-body) | Эталонный дизайн карточек | Скопировать стили и HTML-структуру |
| `templates/news/list.html` | Текущий шаблон новостей — плоский список | Полная переработка | Сохранить пагинацию (.pagination), заменить .news-item на карточки |
| `templates/forum/index.html` | Список разделов — section-card с flex | Стилистическая переработка: full-width, увеличить padding | Сохранить структуру .section-card, .section-info |
| `templates/forum/section.html` | Темы раздела — topic-row с flex | Стилистическая унификация | Сохранить breadcrumbs, topic-row структуру |
| `templates/layouts/base.html:46` | `.container { max-width: 960px }` | Общий контейнер для всех страниц | Не менять (ASM-02) |

## Test Strategy

| Test surface | Canonical refs | Existing coverage | Planned automated coverage | Required local suites | Required CI suites | Manual-only gap | Approval ref |
| --- | --- | --- | --- | --- | --- | --- | --- |
| News list visual layout | `REQ-01`, `SC-01`, `CHK-01` | Нет | Playwright E2E: element assertions на карточки, изображения, анонсы | `npx playwright test` | E2E job | — | none |
| News grid responsiveness | `REQ-02`, `SC-02`, `CHK-02` | Нет | Playwright E2E: screenshots at 3 viewports (400/800/1280px), grid column count assertions | `npx playwright test` | E2E job | — | none |
| News pagination | `REQ-03`, `SC-03`, `NEG-01`, `NEG-02`, `CHK-03`, `CHK-05` | Нет | Playwright E2E: navigate pages, test edge cases (page=0, -1, 999) | `npx playwright test` | E2E job | — | none |
| Forum sections layout | `REQ-04`, `SC-04`, `CHK-01` | Нет | Playwright E2E: element assertions на full-width карточки | `npx playwright test` | E2E job | — | none |
| Forum topics layout | `REQ-05`, `SC-05`, `CHK-01` | Нет | Playwright E2E: element assertions на стиль тем | `npx playwright test` | E2E job | — | none |
| JS console errors | `EC-04`, `CHK-04` | Нет | Playwright E2E: console error listener | `npx playwright test` | E2E job | — | none |

> Go unit-тесты на repo не добавляются: изменение ListPublished — расширение SELECT-запроса без новой бизнес-логики. Корректность SQL верифицируется E2E тестами (карточки рендерятся с данными) и integration тестами в CI.

## Open Questions / Ambiguities

| OQ ID | Question | Why unresolved | Blocks | Default action |
| --- | --- | --- | --- | --- |
| `OQ-01` | Нужно ли менять сигнатуру `ListPublished()` или создать новый метод? | Существующий метод возвращает `[]News`, а нужно `[]HomeNewsItem`. Связан с `ER-01`. | `STEP-01` | Изменить сигнатуру `ListPublished()` на возврат `[]HomeNewsItem` — текущий вызывающий код (handler) единственный. При срабатывании `ER-01` → `STOP-01`. |

## Environment Contract

| Area | Contract | Used by | Failure symptom |
| --- | --- | --- | --- |
| setup | `docker compose -f docker-compose.dev.yml up -d` — app + postgres running | Все STEP-* | Страницы не загружаются на localhost:8080 |
| unit test (Go) | `docker run --rm -v "$(pwd)":/app -w /app -v lfcru_gomod:/root/go/pkg/mod golang:1.23-alpine go test ./...` | CP-01, Closure | Unit test failures |
| E2E setup | `docker compose -f docker-compose.dev.yml up -d && docker compose -f docker-compose.e2e.yml up -d && npm install && npx playwright install chromium` | STEP-06, CP-03 | E2E контейнер не стартует или Playwright не находит браузер |
| E2E run | `npx playwright test` | STEP-06, CP-03 | Playwright test failures |

## Preconditions

| Precondition ID | Canonical ref | Required state | Used by steps | Blocks start |
| --- | --- | --- | --- | --- |
| `PRE-01` | `ASM-01` | HomeNewsItem struct содержит CoverImageURL, Content, CommentCount + ExcerptText() | `STEP-01`, `STEP-02` | yes — подтверждено при grounding |
| `PRE-02` | `ASM-02` | `.container` = 960px max-width | `STEP-03`, `STEP-04`, `STEP-05` | yes — подтверждено при grounding |
| `PRE-03` | `DEC-01`, `DEC-02` | Breakpoints (640/1024px) и page size (9) согласованы | `STEP-01`, `STEP-03` | yes — подтверждено в feature.md |
| `PRE-04` | `feature-flow.md` § Plan Ready → Execution | `feature.md → delivery_status: in_progress`, `implementation-plan.md → status: active` | Все STEP-* | yes — выполняется до первого коммита с кодом |

## Workstreams

| Workstream | Implements | Result | Owner | Dependencies |
| --- | --- | --- | --- | --- |
| `WS-1` | `REQ-01`, `REQ-02`, `REQ-03`, `CTR-01` | Карточный layout новостей с адаптивной сеткой и пагинацией | agent | — |
| `WS-2` | `REQ-04` | Full-width карточки разделов форума | agent | — |
| `WS-3` | `REQ-05` | Унифицированные темы раздела | agent | — |

## Approval Gates

Нет — фича не содержит рискованных или необратимых действий (только UI/template изменения).

## Порядок работ

| Step ID | Actor | Implements | Goal | Touchpoints | Artifact | Verifies | Evidence IDs | Check command | Blocked by | Needs approval | Escalate if |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `STEP-01` | agent | `REQ-01`, `CTR-01` | Расширить repo: ListPublished возвращает HomeNewsItem с image, content, comment_count. Обновить integration-тесты `TestListPublished_*` в `repo_test.go` (scan в HomeNewsItem вместо News) | `internal/news/repo.go`, `internal/news/repo_test.go` | Обновлённый ListPublished() + обновлённые тесты | `CHK-E01` | — | `go vet ./internal/news/...` (внутри golang container) | `PRE-01`, `PRE-04` | none | SQL не компилируется или pgx scan ошибка |
| `STEP-02` | agent | `REQ-01`, `REQ-03`, `CTR-01` | Адаптировать handler: ListData.Items → []HomeNewsItem, pageSize → 9. Обновить unit-тесты `TestShowList_*` в `handler_test.go` (mock возвращает []HomeNewsItem) | `internal/news/handler.go`, `internal/news/handler_test.go` | Обновлённый ShowList() + обновлённые тесты | `CHK-E02` | — | `go vet ./internal/news/...` (внутри golang container) | `STEP-01` | none | Несовместимость с template |
| `STEP-03` | agent | `REQ-01`, `REQ-02`, `REQ-03` | Переписать шаблон новостей: карточки с grid + пагинация | `templates/news/list.html` | Обновлённый шаблон | `CHK-E03` | — | Визуальная проверка localhost:8080/news (Playwright screenshot) | `STEP-02` | none | Карточки не рендерятся |
| `STEP-04` | agent | `REQ-04` | Переработать шаблон разделов форума: full-width карточки | `templates/forum/index.html` | Обновлённый шаблон | `CHK-E04` | — | Визуальная проверка localhost:8080/forum (Playwright screenshot) | — | none | — |
| `STEP-05` | agent | `REQ-05` | Унифицировать стиль тем раздела | `templates/forum/section.html` | Обновлённый шаблон | `CHK-E05` | — | Визуальная проверка localhost:8080/forum/sections/X (Playwright screenshot) | — | none | — |
| `STEP-06` | agent | `SC-01`..`SC-05`, `NEG-01`, `NEG-02` | Написать Playwright E2E тесты (spec file: `e2e/sections-design/unified-design.spec.ts`) | `e2e/sections-design/` | E2E spec файл (TypeScript) | `CHK-E06` | — | `npx playwright test e2e/sections-design/` | `STEP-03`, `STEP-04`, `STEP-05` | none | Тесты красные после 2 fix-итераций |
| `STEP-07` | agent | — | Simplify review: проверить код на premature abstractions, дублирование, dead code | Все изменённые файлы | — | `CHK-E07` | — | Code review pass | `STEP-06` | none | — |

> `CHK-E01`..`CHK-E07` — execution-level checks плана. Acceptance-level `CHK-01`..`CHK-05` из `feature.md` верифицируются в E2E тестах (STEP-06) и при closure.

## Parallelizable Work

- `PAR-01` STEP-04 и STEP-05 (форум) могут выполняться параллельно с STEP-01..03 (новости) — разные файлы, нет общего write-surface.
- `PAR-02` STEP-06 (E2E тесты) блокирован STEP-03..05 — требуется готовый UI.
- `PAR-03` STEP-07 (Simplify review) блокирован STEP-06 — выполняется после тестов.

## Checkpoints

| Checkpoint ID | Refs | Condition | Evidence IDs |
| --- | --- | --- | --- |
| `CP-01` | `STEP-01`, `STEP-02` | Repo и handler компилируются, go vet проходит, ListPublished возвращает HomeNewsItem | — |
| `CP-02` | `STEP-03`, `STEP-04`, `STEP-05` | Все 3 шаблона обновлены, страницы рендерятся без ошибок на localhost | — |
| `CP-03` | `STEP-06`, `STEP-07` | E2E тесты написаны и проходят локально, simplify review пройден | `EVID-01`..`EVID-05` |

## Execution Risks

| Risk ID | Risk | Impact | Mitigation | Trigger |
| --- | --- | --- | --- | --- |
| `ER-01` | ListPublished() используется в других местах кроме handler | Сломается другой вызывающий код | Grep по `ListPublished` перед изменением — при grounding найден только один вызов в handler. При срабатывании → `STOP-01` | Компиляция падает после STEP-01 |
| `ER-02` | У большинства новостей нет cover_image → сетка из placeholder'ов | Визуально непривлекательный результат | Placeholder уже стилизован (gradient LFC red) — приемлемо визуально | SC-01 не проходит визуальную проверку |

## Stop Conditions / Fallback

| Stop ID | Related refs | Trigger | Immediate action | Safe fallback state |
| --- | --- | --- | --- | --- |
| `STOP-01` | `ER-01`, `OQ-01` | ListPublished используется из других модулей | Создать новый метод ListPublishedCards вместо изменения существующего | Оба метода сосуществуют |
| `STOP-02` | `ER-02` | Визуальная проверка SC-01 не проходит после 2 итераций | Эскалировать к человеку для дизайн-решения | Шаблон с placeholder'ами |

## Готово для приемки

- Все STEP-* выполнены (включая STEP-07 Simplify Review)
- E2E тесты (`npx playwright test`) зелёные локально
- Unit-тесты зелёные (команда из ops/development.md)
- CI зелёный (Lint + Go Tests + E2E)
- Страницы `/news`, `/forum`, `/forum/sections/X` визуально соответствуют стилю главной
