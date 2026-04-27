---
title: "FT-020: Таблица чемпионата АПЛ на главной странице"
doc_kind: feature
doc_function: canonical
purpose: "Добавить блок с актуальной таблицей АПЛ на главную страницу: 20 команд, позиция, И/ЗМ/ПМ/РМ/О, цветовые зоны (топ-5 / 6-е / вылет), иконки команд при наличии. Данные — football-data.org, кэш 24ч (1ч в выходные) со сбросом после матча LFC. Перегруппировка колонок лейаута."
derived_from:
  - ../../domain/problem.md
status: active
delivery_status: in_progress
audience: humans_and_agents
trello: https://trello.com/c/v5zs9C1z
must_not_define:
  - implementation_sequence
---

# FT-020: Таблица чемпионата АПЛ на главной странице

## What

### Problem

На главной странице lfc.ru болельщик видит ближайший и последний матч, но не знает текущее положение ЛФК в таблице АПЛ. Чтобы узнать позицию, нужно переходить на другой сайт. Добавление таблицы замыкает ключевой информационный контекст о команде в рамках одной страницы.

### Outcome

| Metric ID | Metric | Baseline | Target | Measurement method |
| --- | --- | --- | --- | --- |
| `MET-01` | Таблица АПЛ отображается на главной странице | Блок отсутствует | Таблица 20 команд с актуальными данными отображается на главной | Ручная проверка в браузере |

### Scope

- `REQ-01` Отображать таблицу чемпионата АПЛ текущего сезона: все 20 команд.
- `REQ-02` Колонки таблицы: позиция (№), название команды (Команда), сыгранных игр (И), забито (ЗМ), пропущено (ПМ), разница мячей (РМ), очки (О).
- `REQ-03` Цветовое выделение строк: места 1–5 — голубой, место 6 — фиолетовый, места 18–20 — красный.
- `REQ-04` Если football-data.org возвращает поле `crest` (URL иконки) в объекте team — отображать иконку перед названием в колонке Команда.
- `REQ-05` Блок «Таблица» располагается в правой колонке главной страницы, после блоков «Последний матч» (FT-019) и «Ближайший матч» (FT-018).
- `REQ-06` Блок «Последнее на форуме» перемещается из full-width-строки в левую колонку, под блок «Последние новости».
- `REQ-07` Данные получать через football-data.org API: `GET /v4/competitions/PL/standings` (текущий сезон определяется сервисом автоматически).
- `REQ-08` Кэш TTL: понедельник–пятница = 24ч; суббота–воскресенье = 1ч. Дополнительный сброс кэша при обнаружении нового завершённого матча LFC (через изменение `LastMatch` в `football.Client`).

### Non-Scope

- `NS-01` Таблицы прошлых сезонов.
- `NS-02` Real-time / SSE обновления таблицы.
- `NS-03` Форма команд (серия последних матчей).
- `NS-04` Статистика конкретного тура (бомбардиры, голы по минутам).
- `NS-05` Переход на детальную страницу команды по клику.
- `NS-06` Proxy или локальный кэш иконок команд на сервере.
- `NS-07` Таблица других соревнований (Лига чемпионов, Кубок FA и т.д.).

### Constraints / Assumptions

- `ASM-01` football-data.org free tier возвращает standings для PL текущего сезона через `/v4/competitions/PL/standings`; структура ответа: `standings[0].table[]` (TOTAL standings).
- `ASM-02` Поле `crest` присутствует в объекте `team` ответа standings (подтверждено для PL free tier).
- `ASM-03` Иконки команд загружаются браузером напрямую с CDN football-data.org — дополнительной инфраструктуры не требуется.
- `CON-01` football-data.org free tier: rate limit 10 req/min; один запрос на standings возвращает всю таблицу.
- `CON-02` `FOOTBALL_DATA_API_KEY` уже задан и используется пакетом `football` (FT-018/019); отсутствие ключа → graceful degrade (блок скрывается).
- `CON-03` CSS-классы для цветовых зон должны работать корректно при изменении позиции команды в течение сезона — позиция берётся из поля `position` ответа API, не жёстко задаётся в шаблоне.

## How

### Solution

При рендере главной страницы handler вызывает новый метод `football.Client.Standings(ctx)`, который запрашивает `/v4/competitions/PL/standings` и кэширует результат с TTL, зависящим от дня недели (24ч / 1ч). Кэш сбрасывается принудительно при обнаружении нового завершённого матча LFC — это происходит внутри `Client`, когда `LastMatch()` возвращает match ID, отличный от последнего кэшированного. Шаблон главной страницы получает поле `Standings []football.StandingsEntry` и рендерит таблицу в правой колонке. Лейаут перегруппировывается: `home-forum` переезжает в левую колонку под `home-news`, правая колонка получает три секции (последний матч, ближайший матч, таблица).

### Change Surface

| Surface | Type | Why it changes |
| --- | --- | --- |
| `internal/football/client.go` | code | Новый тип `StandingsEntry`; метод `Standings(ctx)` с TTL-логикой по дням недели и событийным сбросом кэша |
| `internal/football/client_test.go` | code | Unit-тесты для `Standings`: happy path, cache hit, weekday/weekend TTL, cache invalidation при новом LastMatch |
| `internal/home/handler.go` | code | `FootballSource` интерфейс: добавить `Standings(ctx)`; `HomeData`: добавить `Standings []football.StandingsEntry`; `ShowHome`: вызов `Standings()` |
| `internal/home/handler_test.go` | code | Новые integration test cases с `Standings` |
| `templates/home/index.html` | code | Новая секция таблицы в правой колонке; перегруппировка grid: `home-forum` → левая колонка под `home-news`; CSS для цветовых зон |

### Flow

1. HTTP-запрос на главную поступает в `HomeHandler`.
2. Handler параллельно вызывает `NextMatch(ctx)`, `LastMatch(ctx)`, `Standings(ctx)`.
3. `Standings(ctx)`: проверяет кэш → при miss запрашивает API → кэширует с TTL (weekday/weekend).
4. При cache miss или при обнаружении нового `lastMatchID` в `LastMatch()` — кэш standings сбрасывается.
5. `HomeData.Standings` передаётся в шаблон.
6. Шаблон: если `Standings` не пуст → рендерит таблицу в правой колонке после `home-match`; иначе скрывает блок.
7. `home-forum` рендерится в левой колонке под `home-news`.

### Contracts

| Contract ID | Input / Output | Producer / Consumer | Notes |
| --- | --- | --- | --- |
| `CTR-01` | `GET /v4/competitions/PL/standings` → JSON `{standings: [{stage, type, table: [{position, team{name,crest}, playedGames, goalsFor, goalsAgainst, goalDifference, points}]}]}` | football-data.org / `football.Client` | `standings[0]` = TOTAL; авторизация `X-Auth-Token` |
| `CTR-02` | `StandingsEntry{Position int, TeamName string, TeamCrest string, PlayedGames, GoalsFor, GoalsAgainst, GoalDifference, Points int}` | `football.Client` → `HomeHandler` → шаблон | `TeamCrest` — URL; пустая строка если поле отсутствует |
| `CTR-03` | TTL = 24ч (пн–пт) или 1ч (сб–вс) по `time.Now().Weekday()`; принудительный сброс при смене `lastMatchID` | `football.Client` internal | Сброс standings cache — side effect `LastMatch()` при новом match ID |

### Failure Modes

- `FM-01` API недоступен или ошибка ≥ 500 — вернуть кэш если не истёк; если кэш пуст — `Standings()` возвращает nil, блок скрывается без ошибки.
- `FM-02` API возвращает пустой массив `table` — `Standings()` возвращает nil, блок скрывается.
- `FM-03` Иконка команды (`crest`) не загружается в браузере — атрибут `alt` с названием команды; отображение таблицы не нарушается.
- `FM-04` `FOOTBALL_DATA_API_KEY` отсутствует — `Standings()` возвращает nil (поведение совместимо с FM-01).

## Verify

### Exit Criteria

- `EC-01` Блок таблицы АПЛ отображается на главной странице в правой колонке, после «Ближайшего матча»; содержит 20 строк с корректными данными по колонкам №, Команда, И, ЗМ, ПМ, РМ, О.
- `EC-02` Строки 1–5 выделены голубым, строка 6 — фиолетовым, строки 18–20 — красным.
- `EC-03` Если `crest` URL присутствует — иконка отображается перед названием команды.
- `EC-04` При недоступности API или пустом кэше — блок таблицы не отображается, страница рендерится без ошибок.
- `EC-05` Блок «Последнее на форуме» отображается в левой колонке под «Последними новостями».

### Traceability matrix

| Requirement ID | Design refs | Acceptance refs | Checks | Evidence IDs |
| --- | --- | --- | --- | --- |
| `REQ-01` | `CTR-01`, `CTR-02`, `ASM-01` | `EC-01`, `SC-01` | `CHK-01` | `EVID-01` |
| `REQ-02` | `CTR-02` | `EC-01`, `SC-01` | `CHK-01` | `EVID-01` |
| `REQ-03` | `CON-03` | `EC-02`, `SC-01` | `CHK-01` | `EVID-01` |
| `REQ-04` | `ASM-02`, `CTR-02` | `EC-03`, `SC-01` | `CHK-01` | `EVID-01` |
| `REQ-05` | | `EC-01`, `SC-01` | `CHK-01` | `EVID-01` |
| `REQ-06` | | `EC-05`, `SC-01` | `CHK-01` | `EVID-01` |
| `REQ-07` | `CTR-01`, `CON-01`, `CON-02` | `EC-04`, `SC-02` | `CHK-02` | `EVID-02` |
| `REQ-08` | `CTR-03`, `FM-01` | `SC-03`, `SC-04` | `CHK-03`, `CHK-04` | `EVID-03`, `EVID-04` |

### Acceptance Scenarios

- `SC-01` API доступен, таблица получена. Когда пользователь открывает главную страницу — в правой колонке отображается таблица 20 команд; колонки №, Команда (с иконкой), И, ЗМ, ПМ, РМ, О заполнены; строки 1–5 голубые, строка 6 фиолетовая, строки 18–20 красные. «Последнее на форуме» — в левой колонке под новостями.
- `SC-02` API недоступен, кэш пуст. Главная страница рендерится без ошибок; блок таблицы не отображается; остальные блоки работают нормально.
- `SC-03` Второй запрос главной страницы в пределах TTL (например, в будний день). К API новый запрос не уходит; блок показывает те же данные.
- `SC-04` Новый завершённый матч LFC обнаружен (`LastMatch` вернул новый match ID). При следующем запросе standings кэш сброшен и данные обновлены из API.

### Checks

| Check ID | Covers | How to check | Expected result | Evidence path |
| --- | --- | --- | --- | --- |
| `CHK-01` | `EC-01`, `EC-02`, `EC-03`, `EC-05`, `SC-01` | Открыть главную страницу при действующем `FOOTBALL_DATA_API_KEY` | Таблица 20 команд в правой колонке; цветовые зоны корректны; иконки отображаются; форум — в левой колонке | `artifacts/ft-020/verify/chk-01/` |
| `CHK-02` | `EC-04`, `SC-02` | Запустить с невалидным/отсутствующим `FOOTBALL_DATA_API_KEY` | Блок таблицы не отображается; страница рендерится без паники/500 | `artifacts/ft-020/verify/chk-02/` |
| `CHK-03` | `REQ-08`, `SC-03` | Unit-тест `TestClient_Standings_CacheHit`: два вызова `Standings()` в пределах TTL — `callCount == 1` | Тест зелёный | `artifacts/ft-020/verify/chk-03/` |
| `CHK-04` | `REQ-08`, `SC-04` | Unit-тест `TestClient_Standings_InvalidatedOnNewLastMatch`: смена `lastMatchID` сбрасывает standings кэш | Тест зелёный | `artifacts/ft-020/verify/chk-04/` |

### Test matrix

| Check ID | Evidence IDs | Evidence path |
| --- | --- | --- |
| `CHK-01` | `EVID-01` | `artifacts/ft-020/verify/chk-01/` |
| `CHK-02` | `EVID-02` | `artifacts/ft-020/verify/chk-02/` |
| `CHK-03` | `EVID-03` | `artifacts/ft-020/verify/chk-03/` |
| `CHK-04` | `EVID-04` | `artifacts/ft-020/verify/chk-04/` |

### Evidence

- `EVID-01` Скриншот главной страницы — таблица АПЛ в правой колонке, цветовые зоны, иконки, форум в левой колонке.
- `EVID-02` Скриншот главной страницы — блок таблицы отсутствует при невалидном API key; страница без ошибок.
- `EVID-03` Результат unit-теста `TestClient_Standings_CacheHit` — зелёный.
- `EVID-04` Результат unit-теста `TestClient_Standings_InvalidatedOnNewLastMatch` — зелёный.

### Evidence contract

| Evidence ID | Artifact | Producer | Path contract | Reused by checks |
| --- | --- | --- | --- | --- |
| `EVID-01` | Скриншот главной страницы | human / verify-runner | `artifacts/ft-020/verify/chk-01/` | `CHK-01` |
| `EVID-02` | Скриншот главной страницы (нет блока) | human / verify-runner | `artifacts/ft-020/verify/chk-02/` | `CHK-02` |
| `EVID-03` | Вывод go test | verify-runner | `artifacts/ft-020/verify/chk-03/` | `CHK-03` |
| `EVID-04` | Вывод go test | verify-runner | `artifacts/ft-020/verify/chk-04/` | `CHK-04` |
