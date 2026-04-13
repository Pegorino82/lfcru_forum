---
title: "FT-006: Implementation Plan"
doc_kind: feature
doc_function: derived
purpose: "Execution-план реализации FT-006 (список новостей с пагинацией). Фиксирует discovery context, шаги, test strategy без переопределения canonical feature-фактов."
derived_from:
  - feature.md
status: active
audience: humans_and_agents
must_not_define:
  - ft_006_scope
  - ft_006_architecture
  - ft_006_acceptance_criteria
  - ft_006_blocker_state
---

# План имплементации FT-006

## Цель

Реализовать `GET /news` — список опубликованных новостей с offset-пагинацией (20/стр.), сортировка `published_at DESC`. Новый шаблон, метод репозитория, handler. Все тесты зелёные.

## Discovery Context / Reference Points

| Path / module | Current role | Why relevant | Reuse / mirror |
| --- | --- | --- | --- |
| `migrations/004_create_news.sql` | Определяет таблицу `news` | Схема уже есть: `id, title, is_published, published_at` | Запрос только к этой таблице |
| `internal/news/repo.go` | `LatestPublished(ctx, limit)`, `GetPublishedByID(ctx, id)` | Паттерн запроса: WHERE is_published=true ORDER BY published_at DESC | Добавить `ListPublished(ctx, limit, offset)` в тот же файл |
| `internal/news/handler.go` | `Handler`, `ShowArticle`, `RegisterRoutes` | Паттерн: struct Handler, HTMX check, `RenderPartial` для partial | Добавить `ShowList`, `ListData`, регистрацию `GET /news` |
| `internal/news/repo_test.go` | Интеграционные тесты репо | Паттерны: `setupPool`, `insertUser`, `cleanNews` | Добавить тест-кейсы `TestListPublished_*` в тот же файл |
| `internal/news/handler_test.go` | Интеграционные тесты handler | Паттерны: `testDB`, `newTestServer`, `doGet`, `insertNews` | Добавить `TestShowList_*` в тот же файл |
| `templates/news/article.html` | Шаблон статьи | Паттерн: `{{template "templates/layouts/base.html" .}}`, inline CSS | Создать `list.html` по тому же паттерну |
| `templates/layouts/base.html` | Базовый layout | `{{define "content"}}` block | Использовать тот же механизм |

## Test Strategy

| Test surface | Canonical refs | Existing coverage | Planned automated coverage | Required local suites | Manual-only gap | Manual-only approval ref |
| --- | --- | --- | --- | --- | --- | --- |
| `Repo.ListPublished` | `REQ-01`, `REQ-02`, `SC-01`, `SC-02`, `SC-05` | Нет | `TestListPublished_Empty`, `TestListPublished_Pagination`, `TestListPublished_ExcludesDrafts`, `TestListPublished_SortedDesc` | `go test -tags integration -p 1 ./internal/news/...` | Нет | none |
| `Handler.ShowList` | `REQ-01`..`REQ-04`, `SC-01`..`SC-05` | Нет | `TestShowList_OK`, `TestShowList_Page2`, `TestShowList_InvalidPage`, `TestShowList_NoDrafts`, `TestShowList_HTMXPartial` | `go test -tags integration -p 1 ./internal/news/...` | Нет | none |

## Open Questions

| OQ-ID | Question | Why unresolved | Blocks | Default action |
| --- | --- | --- | --- | --- |
| `OQ-01` | Offset или cursor пагинация? | Выбор влияет на API контракт и сложность реализации | `STEP-01` | Offset-based: достаточно для фан-сайта. DEC-01 в feature.md зафиксировал это решение. |
| `OQ-02` | Page size из конфига или хардкод? | Конфиг-системы для page size нет | `STEP-01` | Хардкод `const pageSize = 20` в handler. NS-05 в feature.md. |
| `OQ-03` | Роут `/news` уже существует? | Нужно проверить регистрацию | `STEP-02` | Нет — `RegisterRoutes` регистрирует только `/news/:id` и `/news/:id/comments`. |

## Environment Contract

| Area | Contract | Used by | Failure symptom |
| --- | --- | --- | --- |
| test | `docker compose -f docker-compose.dev.yml run --rm app go test -tags integration -p 1 ./internal/news/...` | `CHK-01` | FAIL или panic в тестах |
| build | `docker compose -f docker-compose.dev.yml run --rm app go build ./...` | `STEP-05` | Компиляционные ошибки |
| format | `docker compose -f docker-compose.dev.yml run --rm app gofmt -w` | `STEP-05` | Diff в файлах после format |

## Preconditions

| PRE-ID | Canonical ref | Required state | Used by steps | Blocks start |
| --- | --- | --- | --- | --- |
| `PRE-01` | `ASM-01` | Таблица `news` существует, миграции применены | `STEP-01`, `STEP-03` | yes |
| `PRE-02` | `ASM-02` | Роут `/news/:id` существует и не меняется | `STEP-02` | no |
| `PRE-03` | `CON-01` | Docker и docker-compose доступны | `STEP-04`, `STEP-05` | yes |

## Порядок работ

| Step ID | Actor | Implements | Goal | Touchpoints | Artifact | Verifies | Evidence IDs | Check command | Blocked by | Needs approval | Escalate if |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `STEP-01` | agent | `REQ-01`, `REQ-02`, `CTR-02` | Добавить `ListPublished(ctx, limit, offset int) ([]News, int64, error)` в `repo.go` | `internal/news/repo.go` | Метод с COUNT+SELECT | `SC-01`, `SC-02`, `SC-05` | `EVID-01` | compile | `PRE-01` | none | Ошибка SQL — остановиться, исправить запрос |
| `STEP-02` | agent | `REQ-01`, `REQ-03`, `REQ-04`, `CTR-01` | Добавить `ShowList` в `handler.go`, `ListData` struct, зарегистрировать `GET /news` | `internal/news/handler.go` | Handler method + route | `SC-01`..`SC-04` | `EVID-01` | compile | `STEP-01` | none | Конфликт роутов — остановиться, проверить регистрацию |
| `STEP-03` | agent | `REQ-03`, `CON-02` | Создать `templates/news/list.html` | `templates/news/list.html` | HTML шаблон | `SC-01`, `SC-02`, `SC-04` | `EVID-01` | visual inspect | `STEP-02` | none | Template parse error — остановиться, исправить |
| `STEP-04` | agent | `REQ-01`, `REQ-02` | Добавить тесты `TestListPublished_*` в `repo_test.go` | `internal/news/repo_test.go` | Тест-кейсы | `CHK-01` | `EVID-01` | `go test -tags integration -p 1 ./internal/news/...` | `STEP-01` | none | FAIL — остановиться, исправить repo или тест |
| `STEP-05` | agent | `REQ-01`..`REQ-04` | Добавить тесты `TestShowList_*` в `handler_test.go` | `internal/news/handler_test.go` | Тест-кейсы | `CHK-01` | `EVID-01` | `go test -tags integration -p 1 ./internal/news/...` | `STEP-02`, `STEP-03`, `STEP-04` | none | FAIL — остановиться, исправить handler/шаблон/тест |
| `STEP-06` | agent | — | gofmt всех изменённых Go-файлов, запуск финальных тестов | все изменённые .go файлы | чистый diff | `EC-05` | `EVID-01` | `go test -tags integration -p 1 ./internal/news/...` | `STEP-05` | none | FAIL — не коммитить |

## Checkpoints

| CP-ID | Refs | Condition | Evidence IDs |
| --- | --- | --- | --- |
| `CP-01` | `STEP-01`..`STEP-03` | Код компилируется, шаблон парсится | stdout build |
| `CP-02` | `STEP-04`..`STEP-06` | Все тесты зелёные | stdout go test |

## Execution Risks

| ER-ID | Risk | Impact | Mitigation | Trigger |
| --- | --- | --- | --- | --- |
| `ER-01` | `cleanArticleData` чистит только `LIKE 'test-%'` — seed-данные могут мешать `TestListPublished_Empty` | Ложный FAIL | Именовать тестовые новости с префиксом `test-` | Тест видит лишние записи |
| `ER-02` | Конфликт имён test-хелперов между `repo_test.go` и `handler_test.go` | Ошибка компиляции | Оба файла в одном пакете `news_test`; общие хелперы (`testDB`, `insertNews`, `cleanArticleData`) уже в `handler_test.go` | Ошибка "already declared" |

## Stop Conditions

| STOP-ID | Related refs | Trigger | Immediate action | Safe fallback state |
| --- | --- | --- | --- | --- |
| `STOP-01` | `ER-02` | Конфликт имён при компиляции тестов | Вынести общие хелперы в `testhelpers_test.go` или проверить пакет | Не добавлять дублирующие хелперы |
| `STOP-02` | `FM-03` | Тесты падают с DB error | Проверить `DATABASE_URL`, запустить `goose up` вручную | Остановиться до стабилизации DB |

## Готово для приемки

Все `CHK-01` зелёные, `gofmt` применён, `delivery_status: done` в `feature.md`, `implementation-plan.md` → `status: archived`.
