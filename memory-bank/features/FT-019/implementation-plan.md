---
title: "FT-019: Implementation Plan"
doc_kind: feature
doc_function: derived
purpose: "Execution-план реализации FT-019. Фиксирует discovery context, шаги, риски и test strategy без переопределения canonical feature-фактов."
derived_from:
  - feature.md
status: active
audience: humans_and_agents
must_not_define:
  - ft_019_scope
  - ft_019_architecture
  - ft_019_acceptance_criteria
  - ft_019_blocker_state
---

# План имплементации

## Цель текущего плана

Добавить блок «Последний матч» на главную страницу: данные из football-data.org (FINISHED), кэш до kickoff следующего матча, позиция над блоком «Ближайший матч».

## Discovery Context / Reference Points

| Path / module | Current role | Why relevant | Reuse / mirror |
| --- | --- | --- | --- |
| `internal/football/client.go` | `MatchInfo`, `Client` struct (apiKey, httpClient, ttl, sync.Mutex, cached, fetchedAt), `NextMatch(ctx)`, `apiResponse` | Основной файл расширения: добавляем `LastMatchInfo`, `Goal`, новые поля в `Client`, метод `LastMatch()` | Паттерн кэша с mutex; паттерн `apiResponse`; `lookupVenue()` из `venues.go` |
| `internal/football/client_test.go` | Unit-тесты `NextMatch`: `httptest.NewServer`, fake JSON response, assert на поля | Зеркалим паттерн для `LastMatch` тестов | `httptest.NewServer` + `json.NewEncoder(w).Encode(resp)` |
| `internal/football/venues.go` | Статический справочник `stadium → {city, country}`, функция `lookupVenue(string)` | Переиспользуем для last match venue (тот же механизм) | `lookupVenue(homeTeamName)` — вызов идентичен `NextMatch` |
| `internal/home/handler.go` | `FootballSource` interface (`NextMatch`), `HomeData` struct, `ShowHome` handler | Расширяем интерфейс (`LastMatch`), `HomeData`, логику handler | DI-паттерн через интерфейс; `if h.footballClient != nil` guard |
| `internal/home/handler_test.go` | Integration tests (`//go:build integration`); mock `FootballSource`; `newTestServer()` | Расширяем mock и тест-кейсы для `LastFootballMatch` | Inline mock struct с методами `NextMatch` и `LastMatch` |
| `templates/home/index.html` | Секция `.home-match` с блоком «Ближайший матч»; CSS классы `.match-card`, `.match-type-home/away`, `.match-venue` | Добавляем новую секцию выше `.home-match`; переиспользуем CSS | Сохранить `{{if .LastFootballMatch}}...{{else}}...{{end}}` паттерн |
| `cmd/forum/main.go:110` | `football.NewClient(cfg.FootballAPIKey, 24*time.Hour)` — инстанциирование Client | TTL 24h здесь теряет смысл после введения динамического TTL; будем управлять TTL внутри `LastMatch()` | `NewClient` signature может не меняться — TTL используется только для `NextMatch` |

## Test Strategy

| Test surface | Canonical refs | Existing coverage | Planned automated coverage | Required local suites / commands | Required CI suites / jobs | Manual-only gap / justification | Manual-only approval ref |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `internal/football/client.go` — `LastMatch`, кэш, fallback | `REQ-08`, `REQ-09`, `SC-01`, `SC-02`, `SC-03`, `CHK-01`–`CHK-03` | Нет (новый метод) | Unit-тесты с `httptest.NewServer`: happy path (all fields), cache hit (no API call), no FINISHED matches (nil), API error (nil) | `docker run --rm -v $(pwd):/app -w /app golang:1.23-alpine go test ./internal/football/...` | — | — | none |
| `internal/home/handler.go` — `ShowHome` с `LastFootballMatch` | `REQ-01`–`REQ-07`, `REQ-10`, `SC-01`, `SC-02`, `CHK-01`, `CHK-02` | Есть: тесты с/без `NextFootballMatch` | Расширить mock: добавить `LastMatch()`; добавить тест-кейс с `LastMatchInfo`, тест с nil | `docker run --rm -v $(pwd):/app -w /app -e DATABASE_URL=... golang:1.23-alpine go test -tags integration ./internal/home/...` | — | — | none |
| `templates/home/index.html` — рендер блока «Последний матч» | `REQ-01`–`REQ-07`, `REQ-10`, `EC-01`, `EC-02`, `CHK-01`, `CHK-02` | Покрывается через handler integration тесты | Проверяется через handler тесты (assert на HTML body: счёт, голы, ссылка #) | вместе с `./internal/home/...` | — | Визуальная проверка layout и позиции в браузере | `AG-01` |

## Open Questions / Ambiguities

| OQ ID | Question | Why unresolved | Blocks | Default action / escalation owner |
| --- | --- | --- | --- | --- |
| `OQ-01` | Возвращает ли free tier `goals[]` с `scorer.name` и `minute` для FINISHED матчей? | `ASM-02` — предположение, не проверено до кодирования | `STEP-02` (парсинг response) | По умолчанию: добавляем парсинг `goals[]` как описано в `CTR-01`. Если поле отсутствует в реальном ответе — `Goals` будет `[]` (пустой slice), блок не падает, голы не отображаются. Эскалация к пользователю если `ASM-02` опровергается. |
| `OQ-02` | При `NextMatch == nil` (нет SCHEDULED матчей): как `LastMatch()` вычислит TTL? | `FM-03` задаёт fallback, но `nextKickoff` будет zero value | `STEP-02` — логика TTL | По умолчанию: если `c.nextKickoff.IsZero()` → TTL = 24ч. |

## Environment Contract

| Area | Contract | Used by | Failure symptom |
| --- | --- | --- | --- |
| Build | `docker run --rm -v $(pwd):/app -w /app golang:1.23-alpine go build ./...` | все STEP | `build failed` / import error |
| Test (unit) | `docker run --rm -v $(pwd):/app -w /app golang:1.23-alpine go test ./internal/football/...` | `STEP-02`, `STEP-04` | тесты не запускаются |
| Test (integration) | `docker run --rm -v $(pwd):/app -w /app -e DATABASE_URL=... golang:1.23-alpine go test -tags integration ./internal/home/...` | `STEP-04` | тесты скипаются или падают |
| API access | `FOOTBALL_DATA_API_KEY` в `.env.local`; заголовок `X-Auth-Token` | `STEP-02` | HTTP 401 от API |
| Network | Контейнер должен иметь outbound HTTPS к `api.football-data.org` | `STEP-02` | timeout / connection refused |

## Preconditions

| Precondition ID | Canonical ref | Required state | Used by steps | Blocks start |
| --- | --- | --- | --- | --- |
| `PRE-01` | `feature.md` status | `status: active`, `delivery_status: planned` | все | yes |
| `PRE-02` | `ASM-01` | Liverpool FC team ID = 64 подтверждён (проверено в FT-018) | `STEP-02` | no (уже подтверждено) |
| `PRE-03` | `CON-02`, `CTR-01` | `FOOTBALL_DATA_API_KEY` присутствует в `.env.local` (настроен в FT-018) | `STEP-02` | yes |

## Workstreams

| Workstream | Implements | Result | Owner | Dependencies |
| --- | --- | --- | --- | --- |
| `WS-1` Football package | `REQ-05`, `REQ-06`, `REQ-08`, `REQ-09`, `CTR-01`–`CTR-04`, `FM-01`–`FM-03` | `LastMatchInfo`, `Goal`, `LastMatch()` в `client.go`; расширенный `apiResponse` | agent | `PRE-03` |
| `WS-2` Handler + Template | `REQ-01`–`REQ-07`, `REQ-10`, `FM-01`, `FM-02` | Обновлённый `FootballSource`, `HomeData`, `ShowHome`, шаблон | agent | `WS-1` |
| `WS-3` Tests | `SC-01`–`SC-03`, `CHK-01`–`CHK-03` | Новые unit и integration тесты; все зелёные | agent | `WS-1`, `WS-2` |

## Approval Gates

| Approval Gate ID | Trigger | Applies to | Why approval is required | Approver / evidence |
| --- | --- | --- | --- | --- |
| `AG-01` | Визуальная проверка блока в браузере после `STEP-03` | `WS-2`, `CHK-01` | Стили, позиция блока и layout голов не покрываются автотестами | Пользователь (Evgeny); approval фиксируется комментарием в PR |

## Порядок работ

| Step ID | Actor | Implements | Goal | Touchpoints | Artifact | Verifies | Evidence IDs | Check command / procedure | Blocked by | Needs approval | Escalate if |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `STEP-01` | agent | `CTR-01`, `CTR-02`, `CTR-03`, `CTR-04`, `FM-01`–`FM-03` | Расширить `internal/football/client.go`: добавить `Goal{Scorer, Minute}`, `LastMatchInfo{...}`, поля `cachedLast`, `lastFetchedAt`, `nextKickoff` в `Client`; расширить `apiResponse` полями `Score` и `Goals`; добавить `fetchLast(ctx)` и `LastMatch(ctx)` | `internal/football/client.go` | Обновлённый `client.go` | `CHK-03` (partial) | — | `docker run ... go build ./internal/football/...` | `PRE-03` | none | build error / circular import |
| `STEP-02` | agent | `SC-01`–`SC-03`, `CHK-01`–`CHK-03` | Написать unit-тесты в `client_test.go`: `TestClient_LastMatch_HappyPath` (все поля), `TestClient_LastMatch_CacheHit` (нет второго вызова API), `TestClient_LastMatch_NoFinished` (nil), `TestClient_LastMatch_APIError` (nil) | `internal/football/client_test.go` | Обновлённый `client_test.go` | `CHK-03` | `EVID-03` (partial) | `docker run ... go test ./internal/football/...` | `STEP-01`, `OQ-01` | none | тест стабильно красный → не переходить к STEP-03 |
| `STEP-03` | agent | `REQ-01`–`REQ-07`, `REQ-10`, `FM-01`, `FM-02` | Обновить `FootballSource` интерфейс (добавить `LastMatch(ctx)`), `HomeData` (добавить `LastFootballMatch *football.LastMatchInfo`), `ShowHome` (вызов `LastMatch`); добавить секцию «Последний матч» в шаблон выше `.home-match` с CSS для счёта и голов | `internal/home/handler.go`, `templates/home/index.html` | Обновлённые файлы | `CHK-01`, `CHK-02` | `EVID-01`, `EVID-02` | Открыть главную в браузере | `STEP-01` | `AG-01` | шаблон не компилируется; блок не на месте |
| `STEP-04` | agent | `SC-01`, `SC-02`, integration coverage | Обновить mock в `handler_test.go`: добавить `LastMatch()` к mock-struct; добавить `TestHomeHandler_WithLastFootballMatch`, `TestHomeHandler_LastFootballMatchNil` | `internal/home/handler_test.go` | Обновлённый `handler_test.go` | `CHK-01`, `CHK-02` | — | `docker run ... go test -tags integration ./internal/home/...` | `STEP-03` | none | тест падает на assert HTML body |

## Parallelizable Work

- `PAR-01` `STEP-01` и `STEP-02` — можно запустить параллельно (разные файлы, общий write-surface только в `client.go` которое STEP-02 только читает через mock).
- `PAR-02` `STEP-03` и `STEP-04` **нельзя** распараллеливать — тесты зависят от финальных сигнатур handler и шаблона.

## Checkpoints

| Checkpoint ID | Refs | Condition | Evidence IDs |
| --- | --- | --- | --- |
| `CP-01` | `STEP-01`, `STEP-02` | `internal/football/` package собирается, unit-тесты `LastMatch` зелёные | `EVID-03` (partial) |
| `CP-02` | `STEP-03` | `go build ./...` чистый; главная страница рендерится с mock last match данными; блок виден над «Ближайшим матчем» | `EVID-01`, `EVID-02` |
| `CP-03` | `STEP-04` | Все integration тесты home и football пакетов зелёные | `EVID-03` |

## Execution Risks

| Risk ID | Risk | Impact | Mitigation | Trigger |
| --- | --- | --- | --- | --- |
| `ER-01` | Free tier не возвращает `goals[]` для всех соревнований | `Goals` пуст — блок показывается без голов | Graceful: `{{if .Goals}}...{{end}}` в шаблоне; лог warning при пустом goals | `goals: null` или `goals: []` в реальном API response |
| `ER-02` | `FootballSource` mock в handler_test.go не реализует новый метод `LastMatch()` | Compile error в тестах | `STEP-04` обновляет mock сразу после изменения интерфейса | `interface does not implement` в build output |
| `ER-03` | Динамический TTL = 0 если `nextKickoff` в прошлом | Бесконечные API calls | Clamp TTL: `if ttl <= 0 { ttl = 24h }` | cache always expired при отладке |

## Stop Conditions / Fallback

| Stop ID | Related refs | Trigger | Immediate action | Safe fallback state |
| --- | --- | --- | --- | --- |
| `STOP-01` | `PRE-03` | HTTP 401 постоянно | Остановить `STEP-01`; попросить пользователя проверить ключ | Блок скрыт (graceful degrade по `CON-02`) |
| `STOP-02` | `OQ-01` | `goals[]` отсутствует в API response для реальных матчей | Эскалация пользователю: убрать `REQ-06` или найти альтернативный endpoint | Блок показывается без голов |
| `STOP-03` | `ER-02` | `import cycle` в build | Остановить `STEP-03`; пересмотреть структуру пакетов | Откат до чистого build |

## Готово для приёмки

Все условия выполнены, когда:
- `CP-01`, `CP-02`, `CP-03` пройдены
- `AG-01` approval получен от пользователя (Evgeny)
- `EC-01`, `EC-02`, `EC-03` из `feature.md` истинны
- `EVID-01`, `EVID-02`, `EVID-03` заполнены конкретными carriers
