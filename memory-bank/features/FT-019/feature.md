---
title: "FT-019: Блок «Последний матч» на главной странице"
doc_kind: feature
doc_function: canonical
purpose: "Отображать последний сыгранный матч первой команды ЛФК на главной странице: соперник, дата/время (UTC), стадион, H/A, финальный счёт, ссылка-заглушка на форум. Блок располагается над блоком «Ближайший матч». Данные — football-data.org API с кэшем до начала следующего матча."
derived_from:
  - ../../domain/problem.md
status: active
delivery_status: done
audience: humans_and_agents
trello: https://trello.com/c/vWjzMaXc
must_not_define:
  - implementation_sequence
---

# FT-019: Блок «Последний матч» на главной странице

## What

### Problem

На главной странице lfc.ru реализован блок «Ближайший матч» (FT-018), но результат последней игры нигде не отображается. Болельщики не могут быстро узнать счёт последнего матча, кто забил и на какой минуте — не уходя с главной страницы.

### Outcome

| Metric ID | Metric | Baseline | Target | Measurement method |
| --- | --- | --- | --- | --- |
| `MET-01` | Блок показывает актуальные данные о последнем сыгранном матче | Блок отсутствует | Данные отображаются при наличии завершённого матча | Ручная проверка на главной странице |

### Scope

- `REQ-01` Отображать имя команды-соперника в последнем завершённом матче.
- `REQ-02` Отображать дату и время матча в формате UTC/GMT.
- `REQ-03` Отображать стадион: название, город, страна.
- `REQ-04` Отображать тип матча: домашний (H) или выездной (A).
- `REQ-05` Отображать финальный счёт (например, `2:1`).
- `REQ-06` Отображать ссылку на тему матча на форуме как заглушку `href="#"`.
- `REQ-07` Получать данные через football-data.org API (Liverpool FC, team ID 64, `status=FINISHED&limit=1`).
- `REQ-08` Кэшировать ответ API до начала следующего матча (TTL = время до kickoff следующего матча; fallback TTL = 24ч если следующий матч неизвестен).
- `REQ-09` Блок отображается над блоком «Ближайший матч» (`home-match`) в шаблоне главной страницы.

### Non-Scope

- `NS-01` Счёт по тайму — только финальный.
- `NS-02` Автоматическое создание темы на форуме для матча.
- `NS-03` Статистика матча, составы, форма команд.
- `NS-04` Real-time или SSE-обновления блока.
- `NS-05` Матчи молодёжной или женской команды.
- `NS-06` Турнирная таблица или статистика сезона.
- `NS-07` Пенальти или удары по воротам.
- `NS-08` Авторы голов с минутами — football-data.org free tier не возвращает поле `goals[]` для завершённых матчей (подтверждено 2026-04-26, OQ-01 закрыт как неверная предпосылка).

### Constraints / Assumptions

- `ASM-01` Liverpool FC имеет стабильный team ID = 64 на football-data.org (подтверждено в FT-018).
- `CON-01` football-data.org free tier: rate limit 10 req/min; поле `venue` содержит только название стадиона — город и страна берутся из справочника `venues.go` (реализован в FT-018).
- `CON-02` Env var `FOOTBALL_DATA_API_KEY` уже задан и используется пакетом `football` (FT-018); отсутствие ключа → graceful degrade (блок скрывается).
- `FM-01` API недоступен или возвращает ошибку — показывать данные из кэша, если кэш не истёк; если кэш пуст — скрыть блок без вывода ошибки пользователю.
- `FM-02` Нет завершённых матчей (начало сезона) — скрыть блок.
- `FM-03` Следующий матч неизвестен (нет SCHEDULED матчей) — TTL кэша = 24ч как fallback.

## How

### Solution

На сервере при рендере главной страницы запрашивается последний завершённый матч ЛФК через football-data.org API (`/v4/teams/64/matches?status=FINISHED&limit=1`). Ответ кэшируется in-memory с TTL, равным времени до kickoff следующего матча (взятого из уже кэшированного результата `NextMatch`); при отсутствии следующего матча TTL = 24ч. Шаблон главной страницы получает новое поле `LastFootballMatch` и рендерит блок выше существующего «Ближайший матч», либо скрывает блок при nil.

### Change Surface

| Surface | Type | Why it changes |
| --- | --- | --- |
| `internal/football/client.go` | code | Новый тип `LastMatchInfo`; расширение `apiResponse` полем `score`; хранение `nextKickoff time.Time` в `Client`; новый метод `LastMatch(ctx)` с динамическим TTL |
| `internal/football/client_test.go` | code | Новые unit-тесты для `LastMatch`: happy path, cache hit, нет данных |
| `internal/home/handler.go` | code | `FootballSource` интерфейс: добавить `LastMatch(ctx)`; `HomeData`: добавить `LastFootballMatch *football.LastMatchInfo`; `ShowHome`: вызов `LastMatch()` |
| `internal/home/handler_test.go` | code | Новые integration test cases с `LastFootballMatch` |
| `templates/home/index.html` | code | Новая секция «Последний матч» над `.home-match`; CSS для счёта |

### Flow

1. HTTP-запрос на главную поступает в `HomeHandler`.
2. Handler вызывает `footballClient.NextMatch(ctx)` — результат попадает в `HomeData.NextFootballMatch`; `Client` сохраняет `nextKickoff` из ответа.
3. Handler вызывает `footballClient.LastMatch(ctx)` — кэш проверяется: TTL = `nextKickoff - now` (или 24ч).
4. При cache miss — запрос к football-data.org, результат кэшируется.
5. При ошибке API или пустом ответе — `LastMatch` возвращает nil.
6. Шаблон: если `LastFootballMatch != nil` — рендерит блок над «Ближайшим матчем»; иначе скрывает блок.

### Contracts

| Contract ID | Input / Output | Producer / Consumer | Notes |
| --- | --- | --- | --- |
| `CTR-01` | `GET /v4/teams/64/matches?status=FINISHED&limit=1` → JSON с полями `utcDate`, `homeTeam`, `awayTeam`, `venue`, `score.fullTime` | football-data.org / `football` пакет | Авторизация через `X-Auth-Token: $FOOTBALL_DATA_API_KEY`; поле `goals[]` free tier не возвращает |
| `CTR-02` | Struct `LastMatchInfo{Opponent, MatchDate, Stadium, City, Country, IsHome, Tournament, HomeScore, AwayScore, ForumURL}` | `football` пакет → `HomeHandler` | `ForumURL` всегда `"#"` в текущей реализации |
| `CTR-03` | TTL = `max(0, nextKickoff.Sub(now))` или 24ч | `Client.nextKickoff` → `LastMatch()` cache | `nextKickoff` обновляется при каждом успешном `NextMatch` fetch |

### Failure Modes

- `FM-01` API недоступен или ошибка ≥ 500 — возврат из кэша (если не истёк), иначе nil.
- `FM-02` Нет завершённых матчей (`[]`) — возврат nil, блок скрыт.
- `FM-03` Следующий матч неизвестен — TTL = 24ч как fallback.

## Verify

### Exit Criteria

- `EC-01` Блок «Последний матч» отображается на главной странице над блоком «Ближайший матч» и содержит: имя соперника, дату/время UTC, стадион, город, страну, метку H/A, финальный счёт, ссылку `href="#"`.
- `EC-02` При недоступности API или отсутствии завершённых матчей — блок не отображается, страница рендерится без ошибок.
- `EC-03` Повторный запрос главной страницы в пределах TTL не делает новый запрос к API (кэш работает).

### Traceability matrix

| Requirement ID | Design refs | Acceptance refs | Checks | Evidence IDs |
| --- | --- | --- | --- | --- |
| `REQ-01` | `CTR-02` | `EC-01`, `SC-01` | `CHK-01` | `EVID-01` |
| `REQ-02` | `CTR-02` | `EC-01`, `SC-01` | `CHK-01` | `EVID-01` |
| `REQ-03` | `CON-01`, `CTR-02` | `EC-01`, `SC-01` | `CHK-01` | `EVID-01` |
| `REQ-04` | `CTR-02` | `EC-01`, `SC-01` | `CHK-01` | `EVID-01` |
| `REQ-05` | `CTR-01`, `CTR-02` | `EC-01`, `SC-01` | `CHK-01` | `EVID-01` |
| `REQ-06` | `CTR-02` | `EC-01`, `SC-01` | `CHK-01` | `EVID-01` |
| `REQ-07` | `CTR-01`, `CON-02` | `EC-02`, `SC-02` | `CHK-02` | `EVID-02` |
| `REQ-08` | `CTR-03`, `FM-03` | `EC-03`, `SC-03` | `CHK-03` | `EVID-03` |
| `REQ-09` | | `EC-01`, `SC-01` | `CHK-01` | `EVID-01` |

### Acceptance Scenarios

- `SC-01` Последний матч сыгран, API доступен. Когда пользователь открывает главную страницу — блок «Последний матч» расположен над «Ближайшим матчем» и отображает: имя соперника, дату/время UTC, стадион, город, страну, H/A, финальный счёт, ссылку `#`.
- `SC-02` API недоступен, кэш пуст. Главная страница рендерится без ошибок; блок «Последний матч» не отображается.
- `SC-03` Второй запрос главной страницы в пределах TTL. К API запрос не уходит; блок отображает те же данные.

### Checks

| Check ID | Covers | How to check | Expected result | Evidence path |
| --- | --- | --- | --- | --- |
| `CHK-01` | `EC-01`, `SC-01` | Открыть главную страницу при действующем `FOOTBALL_DATA_API_KEY` и наличии завершённого матча | Блок «Последний матч» над «Ближайшим матчем»; содержит: соперник, дата/время, стадион, город, страна, H/A, счёт, ссылка `#` | `artifacts/ft-019/verify/chk-01/` |
| `CHK-02` | `REQ-08`, `EC-02`, `SC-02` | Запустить с отсутствующим/невалидным `FOOTBALL_DATA_API_KEY` | Блок «Последний матч» не отображается; страница рендерится без паники/500 | `artifacts/ft-019/verify/chk-02/` |
| `CHK-03` | `REQ-09`, `EC-03`, `SC-03` | Запустить сервер с DEBUG-логированием в `football` пакете. Открыть главную страницу дважды с интервалом < TTL. Найти записи `api.football-data.org` в stdout — должна быть ровно одна за оба запроса. | Ровно одна запись `api.football-data.org` в логах за два запроса | `artifacts/ft-019/verify/chk-03/` |

### Test matrix

| Check ID | Evidence IDs | Evidence path |
| --- | --- | --- |
| `CHK-01` | `EVID-01` | `artifacts/ft-019/verify/chk-01/` |
| `CHK-02` | `EVID-02` | `artifacts/ft-019/verify/chk-02/` |
| `CHK-03` | `EVID-03` | `artifacts/ft-019/verify/chk-03/` |

### Evidence

- `EVID-01` Визуальная проверка в браузере — блок «Последний матч» (Crystal Palace FC, 25.04.2026, 3:1) расположен над «Ближайшим матчем»; все поля отображаются корректно. Подтверждено Evgeny, 2026-04-26, AG-01.
- `EVID-02` Визуальная проверка в браузере — CHK-02 manual-only gap; graceful degrade подтверждён unit-тестом `TestClient_LastMatch_APIError` (nil при ошибке API) и integration-тестом `TestHomeHandler_LastFootballMatchNil_ShowsEmptyState`.
- `EVID-03` Unit-тест `TestClient_LastMatch_CacheHit` (`internal/football/client_test.go`): callCount == 1 после двух вызовов в пределах TTL — зелёный. 2026-04-26.

### Eval evidence

- `EVID-04` Eval Draft → Design Ready — accept. 2026-04-26. self-check.
- `EVID-05` Eval Execution → Done — accept. 2026-04-26. EVID-01..03 verified, CHK-01/03 pass; CHK-02 покрыт unit+integration тестами; AG-01 approved (Evgeny); все тесты зелёные; NS-08 зафиксирован (goals[] не возвращается free tier).

### Evidence contract

| Evidence ID | Artifact | Producer | Path contract | Reused by checks |
| --- | --- | --- | --- | --- |
| `EVID-01` | Скриншот главной страницы | human / verify-runner | `artifacts/ft-019/verify/chk-01/` | `CHK-01` |
| `EVID-02` | Скриншот главной страницы (нет блока) | human / verify-runner | `artifacts/ft-019/verify/chk-02/` | `CHK-02` |
| `EVID-03` | Фрагмент server log + unit-тест | human / verify-runner | `artifacts/ft-019/verify/chk-03/` | `CHK-03` |
