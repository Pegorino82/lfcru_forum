---
title: "FT-020: Implementation Plan"
doc_kind: feature
doc_function: derived
purpose: "Execution-план реализации FT-020. Фиксирует discovery context, шаги, риски и test strategy без переопределения canonical feature-фактов."
derived_from:
  - feature.md
status: archived
audience: humans_and_agents
must_not_define:
  - ft_020_scope
  - ft_020_architecture
  - ft_020_acceptance_criteria
  - ft_020_blocker_state
---

# План имплементации

## Цель текущего плана

Добавить блок таблицы АПЛ на главную страницу: новый тип `StandingsEntry`, метод `Standings(ctx)` в `football.Client` с TTL по дням недели и событийным сбросом, расширение `HomeData` и `FootballSource`, шаблон + CSS, перегруппировка колонок лейаута.

## Discovery Context / Reference Points

| Path / module | Current role | Why relevant | Reuse / mirror |
| --- | --- | --- | --- |
| `internal/football/client.go` | Кэшируемые запросы к football-data.org; типы `MatchInfo`, `LastMatchInfo`; `Client` со `sync.Mutex`, `baseURL`, `httpClient` | Сюда добавляются `StandingsEntry`, метод `Standings(ctx)`, кэш standings | Повторить паттерн `mu.Lock`/`defer mu.Unlock`, graceful nil return при ошибке, инъекция `baseURL` и `httpClient` для тестов |
| `internal/football/client_test.go` | Unit-тесты через `httptest.NewServer`; инъекция `c.baseURL = srv.URL` и `c.httpClient` | Паттерн для новых unit-тестов standings | Повторить `httptest.NewServer` + `c.baseURL = srv.URL` |
| `internal/home/handler.go` | `FootballSource` интерфейс (`NextMatch`, `LastMatch`); `HomeData` struct; `ShowHome` handler | Нужно добавить `Standings(ctx)` в интерфейс и `Standings []football.StandingsEntry` в `HomeData` | Повторить паттерн nil-guard `if h.footballClient != nil`, `slog.Warn` при ошибке |
| `internal/home/handler_test.go` | Integration-тесты handler с mock `FootballSource` | Нужно обновить mock и добавить тесты с `Standings` | Повторить структуру mock + table-driven tests |
| `templates/home/index.html` | CSS grid (`home-grid`); секции `home-news`, `home-last-match`, `home-match`, `home-forum` | Нужно добавить `home-standings`, перегруппировать grid, переместить `home-forum` | Стиль секций (`section h2`, `match-card`); медиа-запрос `@media (max-width: 768px)` |

### Текущий лейаут и цель

Текущий grid (2 колонки):
- Col 1: `home-news` (rows 1–2) + `home-forum` (row 3, full-width `1/-1`)
- Col 2: `home-last-match` (row 1) + `home-match` (row 2)

Целевой grid:
- Col 1: `home-news` (rows 1–2) + `home-forum` (row 3)
- Col 2: `home-last-match` (row 1) + `home-match` (row 2) + `home-standings` (row 3)

### Standings API response shape

```json
GET /v4/competitions/PL/standings
{
  "standings": [{
    "stage": "REGULAR_SEASON",
    "type": "TOTAL",
    "table": [{
      "position": 1,
      "team": { "id": 64, "name": "Liverpool FC", "crest": "https://..." },
      "playedGames": 32,
      "goalsFor": 70,
      "goalsAgainst": 28,
      "goalDifference": 42,
      "points": 79
    }, ...]
  }]
}
```

### Событийный сброс кэша standings

`LastMatch()` при cache miss вызывает `fetchLast()`. Если новый ответ содержит матч с `MatchDate` ≠ `cachedLast.MatchDate` (прокси для "новый матч завершился") — это сигнал для сброса standings кэша. Реализуется через поле `lastKnownMatchDate time.Time` в `Client` и сравнение после `fetchLast`.

## Test Strategy

| Test surface | Canonical refs | Existing coverage | Planned automated coverage | Required local suites / commands | Required CI suites | Manual-only gap / justification | Manual-only approval ref |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `football.Client.Standings` — happy path | `REQ-07`, `SC-01`, `CHK-01` | нет | Unit-тест `TestClient_Standings_HappyPath` через `httptest.NewServer` | `docker run --rm -v "$(pwd)":/app -w /app golang:1.23-alpine go test ./internal/football/...` | нет CI | — | none |
| `football.Client.Standings` — cache hit | `REQ-08`, `SC-03`, `CHK-03` | нет | Unit-тест `TestClient_Standings_CacheHit`: два вызова, callCount == 1 | см. выше | нет CI | — | none |
| `football.Client.Standings` — TTL weekday/weekend | `REQ-08`, `CTR-03` | нет | Unit-тест `TestClient_Standings_WeekdayTTL` и `TestClient_Standings_WeekendTTL` с mock time | см. выше | нет CI | — | none |
| `football.Client.Standings` — invalidation on new LastMatch | `REQ-08`, `SC-04`, `CHK-04` | нет | Unit-тест `TestClient_Standings_InvalidatedOnNewLastMatch` | см. выше | нет CI | — | none |
| `football.Client.Standings` — empty API key / API error | `FM-01`, `FM-04`, `SC-02`, `CHK-02` | нет | Unit-тесты `TestClient_Standings_EmptyAPIKey`, `TestClient_Standings_APIError` | см. выше | нет CI | — | none |
| `HomeHandler.ShowHome` с standings | `REQ-05`, `REQ-06`, `SC-01` | нет | Integration-тест с mock `FootballSource.Standings` | `docker run ... --network lfcru_forum_default -e DATABASE_URL=... golang:1.23-alpine go test ./internal/home/...` | нет CI | — | none |
| Визуальный лейаут и CSS-зоны | `REQ-03`, `EC-02`, `CHK-01` | нет | — | — | нет CI | Ручная проверка в браузере: цветовые зоны, иконки, расположение форум-блока | `AG-01` |

## Open Questions / Ambiguities

Нет открытых вопросов — discovery context достаточен для старта.

## Environment Contract

| Area | Contract | Used by | Failure symptom |
| --- | --- | --- | --- |
| Go-тесты (unit) | `docker run --rm -v "$(pwd)":/app -w /app golang:1.23-alpine go test ./...` | `CHK-03`, `CHK-04`, все unit | `cannot find package` или `no such file` |
| Go-тесты (integration) | Docker Compose `lfcru_forum_default` network + `DATABASE_URL` | `STEP-04` integration-тесты | `dial tcp: connection refused` |
| API key | `FOOTBALL_DATA_API_KEY` задан в `.env.local` | `STEP-01` локальная ручная проверка | Блок не отображается — это expected FM-04 |
| Тестовый сервер | `httptest.NewServer` + `c.baseURL = srv.URL` | все unit-тесты standings | Запросы уходят в реальный API вместо mock |

## Preconditions

| Precondition ID | Canonical ref | Required state | Used by steps | Blocks start |
| --- | --- | --- | --- | --- |
| `PRE-01` | `ASM-01`, `CON-01` | football-data.org `/v4/competitions/PL/standings` доступен для free tier (подтверждено существующими FT-018/019 + документацией API) | `STEP-01`, `STEP-03` | no — можно начать с unit-тестов без реального API |
| `PRE-02` | `CON-02` | `FOOTBALL_DATA_API_KEY` задан в `.env.local` | `STEP-05` (ручная проверка) | no — unit-тесты не требуют реального ключа |
| `PRE-03` | feature.md `status: active` | `feature.md` Design Ready — подтверждено (2026-04-27) | все шаги | yes |

## Workstreams

| Workstream | Implements | Result | Owner | Dependencies |
| --- | --- | --- | --- | --- |
| `WS-1` | `REQ-07`, `REQ-08`, `CTR-01`, `CTR-02`, `CTR-03` | `StandingsEntry` тип + `Standings(ctx)` метод в `Client` с кэшем и тестами | agent | `PRE-03` |
| `WS-2` | `REQ-01`..`REQ-06`, `CTR-02` | `FootballSource` + `HomeData` расширены; `ShowHome` вызывает `Standings` | agent | `WS-1` завершён |
| `WS-3` | `REQ-01`..`REQ-06`, `EC-01`..`EC-05` | шаблон `index.html`: таблица + цвета + иконки + перегруппировка grid | agent | `WS-2` завершён |

## Approval Gates

| Approval Gate ID | Trigger | Applies to | Why approval is required | Approver / evidence |
| --- | --- | --- | --- | --- |
| `AG-01` | После `STEP-05`: ручная визуальная проверка завершена | `CHK-01`, `EC-01`..`EC-05`, `SC-01` | Цветовые зоны, иконки и перегруппировка форум-блока нельзя верифицировать автоматически | Evgeny; скриншот → `artifacts/ft-020/verify/chk-01/` |

## Порядок работ

| Step ID | Actor | Implements | Goal | Touchpoints | Artifact | Verifies | Evidence IDs | Check command / procedure | Blocked by | Needs approval | Escalate if |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `STEP-01` | agent | `REQ-07`, `CTR-01`, `CTR-02` | Добавить `StandingsEntry` и `standingsAPIResponse`; добавить кэш-поля в `Client`; реализовать `fetchStandings(ctx)` | `internal/football/client.go` | Новый тип + приватный fetch-метод | — | — | компилируется | `PRE-03` | none | API shape отличается от `CTR-01` |
| `STEP-02` | agent | `REQ-08`, `CTR-03` | Реализовать `Standings(ctx)` с TTL-логикой (weekday/weekend) и событийным сбросом при новом LastMatch | `internal/football/client.go` | Публичный метод `Standings(ctx)` | `CHK-03`, `CHK-04` | `EVID-03`, `EVID-04` | `docker run --rm -v "$(pwd)":/app -w /app golang:1.23-alpine go test ./internal/football/...` | `STEP-01` | none | Тесты красные после 2 итераций |
| `STEP-03` | agent | `REQ-07`..`REQ-08` | Написать unit-тесты: happy path, cache hit, weekday/weekend TTL, invalidation on new LastMatch, empty key, API error | `internal/football/client_test.go` | 6+ unit-тестов standings | `CHK-03`, `CHK-04` | `EVID-03`, `EVID-04` | `docker run --rm -v "$(pwd)":/app -w /app golang:1.23-alpine go test ./internal/football/... -v` | `STEP-02` | none | Любой тест красный |
| `STEP-04` | agent | `REQ-05`, `CTR-02` | Добавить `Standings(ctx)` в `FootballSource` интерфейс; добавить `Standings []football.StandingsEntry` в `HomeData`; вызвать `Standings(ctx)` в `ShowHome` | `internal/home/handler.go`, `internal/home/handler_test.go` | Расширенный интерфейс + handler + integration-тест | `SC-01`, `SC-02` | — | `docker run ... golang:1.23-alpine go test ./internal/home/...` | `STEP-01` | none | Mock интерфейс не компилируется |
| `STEP-05` | agent | `REQ-01`..`REQ-06`, `EC-01`..`EC-05` | Добавить блок таблицы в шаблон (`home-standings`); CSS цветовые зоны; иконки через `<img>`; перегруппировать grid (`home-forum` → col 1 row 3, `home-standings` → col 2 row 3) | `templates/home/index.html` | Обновлённый шаблон | `CHK-01`, `CHK-02` | `EVID-01`, `EVID-02` | Визуальная проверка в браузере | `STEP-04` | `AG-01` | Grid ломается на мобильном |
| `STEP-06` | agent | все | Запустить полный test suite; зафиксировать commit | worktree | Зелёный тест suite + коммит | `CHK-03`, `CHK-04` | `EVID-03`, `EVID-04` | `docker run --rm -v "$(pwd)":/app -w /app golang:1.23-alpine go test ./...` | `STEP-03`, `STEP-04`, `STEP-05` | none | Красные тесты после `STEP-03` |

## Parallelizable Work

- `PAR-01` `STEP-03` (тесты football) и `STEP-04` (handler расширение) можно начать параллельно после завершения `STEP-01` и `STEP-02` — разные файлы, нет конфликта.
- `PAR-02` `STEP-05` (шаблон) зависит от `STEP-04` (интерфейс должен компилироваться с `Standings`).

## Checkpoints

| Checkpoint ID | Refs | Condition | Evidence IDs |
| --- | --- | --- | --- |
| `CP-01` | `STEP-01`, `STEP-02`, `STEP-03` | Все unit-тесты `./internal/football/...` зелёные; `Standings(ctx)` компилируется | `EVID-03`, `EVID-04` |
| `CP-02` | `STEP-04` | `FootballSource` интерфейс, `HomeData`, `ShowHome` компилируются; integration-тесты home handler зелёные | — |
| `CP-03` | `STEP-05`, `AG-01` | Ручная проверка одобрена Evgeny: таблица, цвета, иконки, форум-блок на месте | `EVID-01` |
| `CP-04` | `STEP-06` | `go test ./...` зелёный; commit создан в worktree | `EVID-03`, `EVID-04` |

## Execution Risks

| Risk ID | Risk | Impact | Mitigation | Trigger |
| --- | --- | --- | --- | --- |
| `ER-01` | football-data.org free tier не возвращает standings или изменил структуру `standings[0].table` | `STEP-01` заблокирован, `CTR-01` неверен | Проверить реальный ответ API через `curl` с `FOOTBALL_DATA_API_KEY` до написания типов | Компилятор падает при json decode или `table` пуст |
| `ER-02` | Событийный сброс по `MatchDate` не работает если API возвращает тот же матч с другим временем (DST) | Standings кэш не сбрасывается | Использовать match ID вместо MatchDate если API возвращает `id` в matches response | `TestClient_Standings_InvalidatedOnNewLastMatch` красный |
| `ER-03` | Иконки команд заблокированы CSP или не загружаются из внешнего CDN | Битые img теги | `alt`-атрибут с названием команды; `onerror` в шаблоне скрывает `<img>` (опционально) | Ручная проверка CHK-01 |

## Stop Conditions / Fallback

| Stop ID | Related refs | Trigger | Immediate action | Safe fallback state |
| --- | --- | --- | --- | --- |
| `STOP-01` | `ER-01`, `CTR-01` | API не возвращает `standings[0].table` для PL free tier | Остановить `STEP-01`, поднять upstream вопрос в `feature.md` | Функциональность не реализуется; `Standings()` возвращает nil |
| `STOP-02` | `ER-02` | `TestClient_Standings_InvalidatedOnNewLastMatch` красный после рефактора на match ID | Проверить реальный ответ API на наличие поля `id` в матчах | Откатиться к MatchDate-based сравнению с комментарием о ER-02 |

## Готово для приемки

- Все unit-тесты `./internal/football/...` зелёные (cp-01)
- Все integration-тесты `./internal/home/...` зелёные (cp-02)
- `AG-01` approved Evgeny со скриншотом (cp-03)
- `go test ./...` зелёный (cp-04)
- `feature.md` → `delivery_status: in_progress` при старте, → `done` после приемки
- PR переведён из draft в ready for review
