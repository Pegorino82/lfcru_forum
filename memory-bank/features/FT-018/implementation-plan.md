---
title: "FT-018: Implementation Plan"
doc_kind: feature
doc_function: derived
purpose: "Execution-план реализации FT-018. Фиксирует discovery context, шаги, риски и test strategy без переопределения canonical feature-фактов."
derived_from:
  - feature.md
status: active
audience: humans_and_agents
must_not_define:
  - ft_018_scope
  - ft_018_architecture
  - ft_018_acceptance_criteria
  - ft_018_blocker_state
---

# План имплементации

## Цель текущего плана

Наполнить блок «Ближайший матч» на главной странице данными из football-data.org API: соперник, дата/время UTC, стадион, город, страна, тип матча (H/A). Кэш in-memory TTL ≥ 24ч.

## Discovery Context / Reference Points

| Path / module | Current role | Why relevant | Reuse / mirror |
| --- | --- | --- | --- |
| `internal/home/handler.go` | Хендлер главной страницы; принимает `newsRepo`, `matchRepo`, `topicRepo`; собирает `HomeData` и рендерит шаблон | Сюда добавляем `FootballSource` интерфейс и новое поле `NextFootballMatch` | Паттерн dependency injection: новый источник данных передаётся через конструктор |
| `internal/home/handler.go:HomeData` | Struct, передаваемый в шаблон: `{User, CSRFToken, News, NextMatch, Topics}` | Расширяем: добавляем `NextFootballMatch *football.MatchInfo` | Структура данных для шаблона — следовать существующему паттерну |
| `templates/home/index.html:47-62` | HTML-блок «Ближайший матч»; рендерит `{{.NextMatch.Opponent}}`, `{{.NextMatch.MatchDate}}`, `{{.NextMatch.Tournament}}` | Обновляем для показа `NextFootballMatch` с новыми полями | Сохранить существующую структуру `{{if .NextFootballMatch}}...{{else}}<empty-state>{{end}}` |
| `internal/config/config.go` | `Config` struct + `Load()` + helpers `getEnv`, `getBool`, `getInt`, `getDuration` | Добавляем `FootballAPIKey string` | Паттерн: `FootballAPIKey: getEnv("FOOTBALL_DATA_API_KEY", "")` |
| `internal/forum/hub.go` | SSE Hub: in-memory `map` + `sync.RWMutex` | Единственный пример in-memory state с mutex в кодовой базе | Паттерн: `sync.Mutex` + struct fields для cached value + timestamp |
| `internal/match/` | `match.Match` struct + `match.Repo.NextUpcoming()` — admin-managed матчи из PostgreSQL | **Не трогаем** — параллельный admin flow; `HomeData.NextMatch` остаётся `*match.Match` | — |
| `internal/home/handler_test.go` | Integration tests; `//go:build integration`; shared `pgxpool`; `httptest`; `newTestServer()` factory | Расширяем: добавляем mock `FootballSource`; тест с данными и тест с пустым ответом | Паттерн `newTestServer()` + `doGet()` |
| `cmd/forum/main.go:109` | `home.NewHandler(newsRepo, matchRepo, topicRepo)` — точка инстанциирования хендлера | Обновляем: передаём `footballClient` как новый аргумент | — |

## Test Strategy

| Test surface | Canonical refs | Existing coverage | Planned automated coverage | Required local suites | Required CI suites | Manual-only gap / justification | Manual-only approval ref |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `internal/football/client.go` — NextMatch, кэш, fallback | `REQ-05`, `REQ-06`, `SC-01`, `SC-02`, `SC-03`, `CHK-01`–`CHK-03` | Нет (новый пакет) | Unit-тесты с `httptest.NewServer` (mock API): happy path, кэш hit, API error → nil | `docker run golang:1.23-alpine go test ./internal/football/...` | — | — | none |
| `internal/football/venues.go` — lookupVenue | `REQ-03`, `CON-01` | Нет (новый файл) | Unit-тест: известный venue → ожидаемый city/country; unknown venue → fallback | вместе с `./internal/football/...` | — | — | none |
| `internal/home/handler.go` — ShowHome с football data | `REQ-01`–`REQ-04`, `SC-01`, `SC-02`, `CHK-01`, `CHK-02` | Есть: `TestHomeHandler_EmptyState`, `TestHomeHandler_WithData` | Расширить: `TestHomeHandler_WithFootballMatch` (mock возвращает MatchInfo), `TestHomeHandler_FootballClientError` (mock возвращает nil) | `docker run golang:1.23-alpine go test -tags integration ./internal/home/...` | — | — | none |
| `templates/home/index.html` — рендер блока | `REQ-01`–`REQ-04`, `EC-01`, `EC-02`, `CHK-01`, `CHK-02` | Покрывается handler integration тестами | Проверяется через handler тесты (assert на HTML body) | вместе с `./internal/home/...` | — | Визуальная проверка в браузере (layout, стили) | `AG-01` |

## Open Questions / Ambiguities

| OQ ID | Question | Why unresolved | Blocks | Default action / escalation owner |
| --- | --- | --- | --- | --- |
| `OQ-01` | Возвращает ли `/v4/teams/64/matches?status=SCHEDULED&limit=1` поле `venue` в ответе free tier? | Не проверено до старта кодирования | `STEP-03` (парсинг API response) | По умолчанию: предполагаем, что `venue` есть (документация football-data.org обещает). При отсутствии — lookupVenue получает пустую строку и возвращает city/country как `""`. Эскалация к пользователю если поле отсутствует для > 50% матчей. |
| `OQ-02` | Какие competitions включить в запрос к API? Free tier покрывает PL, UCL, FA Cup, EFL Cup. | Нет явного требования в feature.md | `STEP-03` — параметр запроса | По умолчанию: не фильтруем по competition (вернёт ближайший матч из любой лиги). Если free tier требует явного competition filter — добавить `competitions=PL,CL,FAC,ELC`. |

## Environment Contract

| Area | Contract | Used by | Failure symptom |
| --- | --- | --- | --- |
| Build | Go через Docker: `docker run --rm -v $(pwd):/app -w /app golang:1.23-alpine go build ./...` | все STEP | `build failed` / import error |
| Test | `docker run --rm -v $(pwd):/app -w /app -e DATABASE_URL=... golang:1.23-alpine go test -tags integration ./...` | `STEP-06` | тесты не запускаются или падают |
| API access | `FOOTBALL_DATA_API_KEY` установлен в `.env.local`; `X-Auth-Token` заголовок в запросах к `https://api.football-data.org/v4/` | `STEP-03`, `STEP-04` | HTTP 401 от API |
| Network | Контейнер должен иметь доступ к `api.football-data.org` (outbound HTTPS) | `STEP-03` | `connection refused` / timeout |

## Preconditions

| Precondition ID | Canonical ref | Required state | Used by steps | Blocks start |
| --- | --- | --- | --- | --- |
| `PRE-01` | `feature.md` status | `status: active`, `delivery_status: planned` | все | yes |
| `PRE-02` | `CON-02`, `CTR-03` | API key получен на football-data.org (free tier регистрация) | `STEP-01`, `STEP-03` | yes |
| `PRE-03` | `ASM-01` | Liverpool FC team ID = 64 подтверждён в football-data.org | `STEP-03` | yes (быстрая проверка: `curl "https://api.football-data.org/v4/teams/64" -H "X-Auth-Token: $KEY"`) |

## Workstreams

| Workstream | Implements | Result | Owner | Dependencies |
| --- | --- | --- | --- | --- |
| `WS-1` Config | `CON-02`, `CTR-03` | `FootballAPIKey` в `Config`, `FOOTBALL_DATA_API_KEY` в `.env.example` и `.env.local` | agent | `PRE-02` |
| `WS-2` Football package | `REQ-05`, `REQ-06`, `CTR-01`, `CTR-02`, `FM-01`, `FM-02` | `internal/football/`: client, cache, venues | agent | `WS-1` |
| `WS-3` Handler + Template | `REQ-01`–`REQ-04`, `FM-01`, `FM-02` | Обновлённый `HomeHandler`, обновлённый шаблон | agent | `WS-2` |
| `WS-4` Tests | `SC-01`–`SC-03`, `CHK-01`–`CHK-03` | Новые unit и integration тесты; все зелёные | agent | `WS-2`, `WS-3` |
| `WS-5` Docs | `CON-02` | `memory-bank/ops/config.md` + `.env.example` обновлены | agent | `WS-1` |

## Approval Gates

| Approval Gate ID | Trigger | Applies to | Why approval is required | Approver / evidence |
| --- | --- | --- | --- | --- |
| `AG-01` | Визуальная проверка блока в браузере после `STEP-05` | `WS-3`, `CHK-01` | Стили и layout не покрываются автотестами; пользователь должен подтвердить что блок выглядит корректно | Пользователь (Evgeny); approval фиксируется комментарием в PR |

## Порядок работ

| Step ID | Actor | Implements | Goal | Touchpoints | Artifact | Verifies | Evidence IDs | Check command / procedure | Blocked by | Needs approval | Escalate if |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `STEP-01` | agent | `CON-02`, `CTR-03` | Добавить `FootballAPIKey string` в `Config` и `Load()`; добавить в `.env.example` | `internal/config/config.go`, `.env.example` | Обновлённый `config.go` | `CHK-02` (частично) | — | `docker run ... go build ./internal/config/...` | `PRE-02` | none | build error |
| `STEP-02` | agent | `REQ-03`, `CON-01` | Создать `internal/football/venues.go`: статический map `venueToCity` (≥ 20 стадионов АПЛ + UCL) и `lookupVenue(venue string) (city, country string)` | `internal/football/venues.go` (новый файл) | `venues.go` | `CHK-01` (partial) | — | `docker run ... go test ./internal/football/...` | none | none | не хватает стадиона для матча → добавить вручную |
| `STEP-03` | agent | `REQ-05`, `REQ-06`, `CTR-01`, `CTR-02`, `FM-01`, `FM-02` | Создать `internal/football/client.go`: `MatchInfo` struct, `Client` struct (apiKey, httpClient, ttl, sync.Mutex, cached, fetchedAt), `NewClient(apiKey, ttl)`, `NextMatch(ctx) (*MatchInfo, error)` с in-memory кэшем | `internal/football/client.go` (новый файл) | `client.go` | `CHK-02`, `CHK-03` | — | `docker run ... go build ./internal/football/...` | `STEP-01`, `STEP-02`, `OQ-01` | none | API возвращает 401 постоянно → эскалация на `PRE-02`; venue всегда пустой → эскалация `OQ-01` |
| `STEP-04` | agent | `REQ-01`–`REQ-04`, `FM-01`, `FM-02` | Добавить `FootballSource` интерфейс в `internal/home/handler.go`; добавить `NextFootballMatch *football.MatchInfo` в `HomeData`; обновить конструктор `NewHandler` и `ShowHome`; обновить `cmd/forum/main.go` | `internal/home/handler.go`, `cmd/forum/main.go` | Обновлённые файлы | `CHK-01`, `CHK-02` | — | `docker run ... go build ./...` | `STEP-03` | none | circular import / build error → пересмотр пакетной структуры |
| `STEP-05` | agent | `REQ-01`–`REQ-04`, `EC-01`, `EC-02` | Обновить `templates/home/index.html`: блок «Ближайший матч» рендерит `{{.NextFootballMatch}}` с полями Opponent, MatchDate (UTC), Stadium, City, Country, IsHome (H/A); empty-state при nil | `templates/home/index.html` | Обновлённый шаблон | `CHK-01`, `CHK-02` | `EVID-01`, `EVID-02` (ручная) | Открыть главную страницу в браузере | `STEP-04` | `AG-01` | шаблон не компилируется → fix template syntax |
| `STEP-06` | agent | `SC-01`–`SC-03`, `CHK-01`–`CHK-03` | Написать тесты: `internal/football/client_test.go` (unit: happy path, cache hit, API error); расширить `internal/home/handler_test.go` (integration: с MatchInfo, с nil) | `internal/football/client_test.go`, `internal/home/handler_test.go` | Новые/обновлённые тест-файлы | `CHK-01`–`CHK-03` | `EVID-03` | `docker run ... go test -tags integration ./internal/football/... ./internal/home/...` | `STEP-04`, `STEP-05` | none | тест стабильно красный → не переходить к CP-03 |
| `STEP-07` | agent | `CON-02` | Обновить `memory-bank/ops/config.md`: добавить `FOOTBALL_DATA_API_KEY`; добавить в `.env.example` | `memory-bank/ops/config.md`, `.env.example` | Обновлённые doc-файлы | — | — | review diff | `STEP-01` | none | — |

## Parallelizable Work

- `PAR-01` `STEP-01` и `STEP-02` независимы — можно выполнять параллельно (разные файлы, нет общего write-surface).
- `PAR-02` `STEP-07` можно начать сразу после `STEP-01` (docs-only, не блокирует код).
- `PAR-03` `STEP-04` и `STEP-05` **нельзя** распараллеливать — шаблон зависит от финального `HomeData`.

## Checkpoints

| Checkpoint ID | Refs | Condition | Evidence IDs |
| --- | --- | --- | --- |
| `CP-01` | `STEP-01`, `STEP-02`, `STEP-03` | `internal/football/` пакет собирается и unit-тесты клиента зелёные | `EVID-03` (partial) |
| `CP-02` | `STEP-04`, `STEP-05` | `go build ./...` чистый; главная страница рендерится с mock-данными | `EVID-01`, `EVID-02` |
| `CP-03` | `STEP-06` | Все integration тесты home и football пакетов зелёные | `EVID-03` |

## Execution Risks

| Risk ID | Risk | Impact | Mitigation | Trigger |
| --- | --- | --- | --- | --- |
| `ER-01` | Football-data.org free tier не возвращает `venue` для некоторых матчей | Stadium/City/Country будут пустыми для части матчей | `lookupVenue("")` возвращает `""` gracefully; template скрывает пустые поля | venue = "" в API response |
| `ER-02` | Free tier требует явного `competitions` фильтра, без него возвращает 403 | `NextMatch` всегда nil | Добавить `competitions=PL,CL,FAC,ELC` в query string | HTTP 403 от API |
| `ER-03` | Team ID 64 изменился или команда не найдена | HTTP 404; `NextMatch` nil | Проверить `PRE-03` до STEP-03; при 404 — эскалация | HTTP 404 при тестовом curl |
| `ER-04` | Circular import: `football` пакет импортирует что-то из `home` | Build error | `football.MatchInfo` — независимый struct без зависимостей от других internal пакетов | `import cycle` в build output |

## Stop Conditions / Fallback

| Stop ID | Related refs | Trigger | Immediate action | Safe fallback state |
| --- | --- | --- | --- | --- |
| `STOP-01` | `PRE-02`, `CTR-03` | HTTP 401 постоянно (ключ невалиден) | Остановить `STEP-03`; попросить пользователя проверить ключ | Блок скрыт (graceful degrade по `CON-02`) |
| `STOP-02` | `OQ-01` | Venue всегда пустая строка для > 50% тестовых запросов | Эскалация: обсудить с пользователем — убрать город/страну из блока или добавить обходной путь | Шаблон показывает блок без города/страны |
| `STOP-03` | `ER-04` | `import cycle` в build | Остановить `STEP-04`; пересмотреть структуру пакетов | Откат `STEP-04` до чистого build |

## Eval Evidence

- `EVID-DR-PR` Eval Design Ready → Plan Ready — accept. 2026-04-26. evaluator agent.

## Готово для приёмки

Все условия выполнены, когда:
- `CP-01`, `CP-02`, `CP-03` пройдены
- `AG-01` approval получен от пользователя
- `EC-01`, `EC-02`, `EC-03` из `feature.md` истинны
- `EVID-01`, `EVID-02`, `EVID-03` заполнены конкретными carriers
