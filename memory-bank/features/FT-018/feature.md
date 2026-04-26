---
title: "FT-018: Блок «Ближайший матч» на главной странице"
doc_kind: feature
doc_function: canonical
purpose: "Отображать следующий предстоящий матч первой команды ЛФК на главной странице: соперник, дата/время (UTC), стадион, тип матча (H/A). Данные — football-data.org API с кэшем TTL ≥ 24ч."
derived_from:
  - ../../domain/problem.md
status: active
delivery_status: in_progress
audience: humans_and_agents
trello: https://trello.com/c/Yh5wlot4
github_issue: https://github.com/Pegorino82/lfcru_forum/issues/6
must_not_define:
  - implementation_sequence
---

# FT-018: Блок «Ближайший матч» на главной странице

## What

### Problem

На главной странице сайта lfc.ru есть блок «Ближайший матч», но он полностью пустой — никакой информации о предстоящем матче не отображается. Болельщики не могут быстро узнать, когда и против кого следующая игра.

### Outcome

| Metric ID | Metric | Baseline | Target | Measurement method |
| --- | --- | --- | --- | --- |
| `MET-01` | Блок показывает актуальные данные о следующем матче | Блок пустой | Данные отображаются при наличии запланированного матча | Ручная проверка на главной странице |

### Scope

- `REQ-01` Отображать имя команды-соперника.
- `REQ-02` Отображать дату и время матча в формате UTC/GMT.
- `REQ-03` Отображать стадион: название, город, страна.
- `REQ-04` Отображать тип матча: домашний (H) или выездной (A).
- `REQ-05` Получать данные через football-data.org API (Liverpool FC, team ID 64).
- `REQ-06` Кэшировать ответ API с TTL ≥ 24 часов.

### Non-Scope

- `NS-01` Результат матча и списки голов — только предстоящие матчи.
- `NS-02` Ссылка на тему матча на форуме.
- `NS-03` Статистика матча, составы, форма команд.
- `NS-04` Real-time или SSE-обновления блока.
- `NS-05` Матчи молодёжной или женской команды.
- `NS-06` Турнирная таблица или статистика сезона.

### Constraints / Assumptions

- `ASM-01` Liverpool FC имеет стабильный team ID = 64 на football-data.org.
- `ASM-02` football-data.org free tier покрывает АПЛ, FA Cup и UCL — достаточно для задачи.
- `CON-01` football-data.org free tier: rate limit 10 req/min; поле `venue` содержит только название стадиона без города и страны — город и страна берутся из статического справочника стадионов.
- `CON-02` Новый env var `FOOTBALL_DATA_API_KEY` обязателен для работы фичи; отсутствие ключа — graceful degrade (блок скрывается).
- `FM-01` API недоступен или возвращает ошибку — показывать данные из кэша, если кэш не истёк; если кэш пуст — скрыть блок без вывода ошибки пользователю.
- `FM-02` В расписании нет предстоящих матчей (конец сезона, пауза) — скрыть блок.
- `FM-03` Матч перенесён — данные обновятся при следующем истечении TTL (≤ 24 ч после переноса).

## How

### Solution

На сервере при рендере главной страницы запрашивается следующий запланированный матч ЛФК через football-data.org API (`/v4/teams/64/matches?status=SCHEDULED&limit=1`). Ответ кэшируется в памяти (или в БД) с TTL = 24 ч. Шаблон главной страницы заполняется данными из кэша или скрывает блок при их отсутствии. Статический справочник стадионов (`venue → city, country`) разрешает недостающие поля города и страны.

### Change Surface

| Surface | Type | Why it changes |
| --- | --- | --- |
| `internal/football/` (новый пакет) | code | HTTP-клиент football-data.org, кэш, справочник стадионов |
| `internal/handler/home.go` | code | Передаёт данные о матче в шаблон главной страницы |
| `web/templates/index.html` (или аналог) | code | Заполняет/скрывает блок «Ближайший матч» |
| `memory-bank/ops/config.md` | doc | Новый env var `FOOTBALL_DATA_API_KEY` |
| `.env.local` (не коммитится) | config | Значение `FOOTBALL_DATA_API_KEY` |

### Flow

1. HTTP-запрос на главную страницу поступает в `HomeHandler`.
2. Handler вызывает `football.NextMatch()` — возвращает данные из кэша, если TTL не истёк.
3. При cache miss — запрос к football-data.org API, результат кладётся в кэш.
4. При ошибке API или пустом ответе — handler передаёт `nil` в шаблон.
5. Шаблон: если данные есть — рендерит блок; если `nil` — скрывает блок.

### Contracts

| Contract ID | Input / Output | Producer / Consumer | Notes |
| --- | --- | --- | --- |
| `CTR-01` | `GET /v4/teams/64/matches?status=SCHEDULED&limit=1` → JSON с полями `utcDate`, `homeTeam`, `awayTeam`, `venue` | football-data.org / `football` пакет | Авторизация через заголовок `X-Auth-Token: $FOOTBALL_DATA_API_KEY` |
| `CTR-02` | Struct `MatchInfo{Opponent, Date, Venue, City, Country, IsHome}` | `football` пакет → `HomeHandler` | Canonical data contract между пакетом и хэндлером |
| `CTR-03` | Env var `FOOTBALL_DATA_API_KEY` | ops / `football` пакет | Обязателен для запуска; отсутствие → блок скрыт |

### Failure Modes

- `FM-01` API недоступен или ошибка ≥ 500 — возврат из кэша (если не истёк), иначе `nil`.
- `FM-02` Нет предстоящих матчей (`[]`) — возврат `nil`, блок скрыт.
- `FM-03` Матч перенесён — новые данные появятся после истечения TTL кэша (≤ 24 ч).

## Verify

### Exit Criteria

- `EC-01` Блок на главной странице отображает имя соперника, дату/время (UTC), стадион, город, страну и метку H/A для ближайшего предстоящего матча.
- `EC-02` При недоступности API или пустом расписании блок не отображается и не показывает ошибку.
- `EC-03` Повторный запрос главной страницы в пределах TTL не делает новый запрос к API (кэш работает).

### Traceability matrix

| Requirement ID | Design refs | Acceptance refs | Checks | Evidence IDs |
| --- | --- | --- | --- | --- |
| `REQ-01` | `CTR-02` | `EC-01`, `SC-01` | `CHK-01` | `EVID-01` |
| `REQ-02` | `CTR-02` | `EC-01`, `SC-01` | `CHK-01` | `EVID-01` |
| `REQ-03` | `CON-01`, `CTR-02` | `EC-01`, `SC-01` | `CHK-01` | `EVID-01` |
| `REQ-04` | `CTR-02` | `EC-01`, `SC-01` | `CHK-01` | `EVID-01` |
| `REQ-05` | `CTR-01`, `CTR-03`, `CON-02` | `EC-01`, `SC-01` | `CHK-02` | `EVID-02` |
| `REQ-06` | `FM-01` | `EC-03`, `SC-03` | `CHK-03` | `EVID-03` |

### Acceptance Scenarios

- `SC-01` Следующий матч в расписании существует. Когда пользователь открывает главную страницу — блок «Ближайший матч» отображает имя соперника, дату/время UTC, стадион, город, страну и метку H/A.
- `SC-02` API недоступен, кэш пуст → блок «Ближайший матч» скрыт, страница рендерится без ошибок.
- `SC-03` Второй запрос главной страницы в пределах TTL → к API запрос не уходит, блок отображает те же данные.

### Checks

| Check ID | Covers | How to check | Expected result | Evidence path |
| --- | --- | --- | --- | --- |
| `CHK-01` | `EC-01`, `SC-01` | Открыть главную страницу при действующем `FOOTBALL_DATA_API_KEY` и наличии предстоящего матча | Блок содержит все 5 полей: соперник, дата/время, стадион, город, страна, H/A | `artifacts/ft-018/verify/chk-01/` |
| `CHK-02` | `REQ-05`, `EC-02`, `SC-02` | Запустить с отсутствующим/невалидным `FOOTBALL_DATA_API_KEY` | Блок «Ближайший матч» не отображается; страница рендерится без паники/500 | `artifacts/ft-018/verify/chk-02/` |
| `CHK-03` | `REQ-06`, `EC-03`, `SC-03` | Запустить сервер с логированием HTTP-запросов в `football` пакете (уровень DEBUG). Открыть главную страницу дважды с интервалом < TTL. В stdout/файле логов найти строки с `api.football-data.org` — должна быть ровно одна. | Ровно одна запись `api.football-data.org` в логах за два запроса к главной | `artifacts/ft-018/verify/chk-03/` |

### Test matrix

| Check ID | Evidence IDs | Evidence path |
| --- | --- | --- |
| `CHK-01` | `EVID-01` | `artifacts/ft-018/verify/chk-01/` |
| `CHK-02` | `EVID-02` | `artifacts/ft-018/verify/chk-02/` |
| `CHK-03` | `EVID-03` | `artifacts/ft-018/verify/chk-03/` |

### Evidence

- `EVID-01` Скриншот главной страницы с заполненным блоком «Ближайший матч».
- `EVID-02` Скриншот главной страницы при отсутствующем API ключе — блок скрыт, страница без ошибок.
- `EVID-03` Фрагмент лога сервера: один запрос к football-data.org за два обращения к главной странице в пределах TTL.

### Eval evidence

- `EVID-04` Eval Draft → Design Ready — accept. 2026-04-26. self-check.

### Evidence contract

| Evidence ID | Artifact | Producer | Path contract | Reused by checks |
| --- | --- | --- | --- | --- |
| `EVID-01` | Скриншот главной страницы | human / verify-runner | `artifacts/ft-018/verify/chk-01/` | `CHK-01` |
| `EVID-02` | Скриншот главной страницы (нет блока) | human / verify-runner | `artifacts/ft-018/verify/chk-02/` | `CHK-02` |
| `EVID-03` | Фрагмент server log | human / verify-runner | `artifacts/ft-018/verify/chk-03/` | `CHK-03` |
