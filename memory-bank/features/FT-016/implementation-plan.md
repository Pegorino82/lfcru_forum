---
title: "FT-016: Implementation Plan"
doc_kind: feature
doc_function: derived
purpose: "Execution-план реализации FT-016 (SSE real-time посты форума). Фиксирует discovery context, шаги, риски и test strategy без переопределения canonical feature-фактов."
derived_from:
  - feature.md
status: archived
audience: humans_and_agents
must_not_define:
  - ft016_scope
  - ft016_architecture
  - ft016_acceptance_criteria
  - ft016_blocker_state
---

# План имплементации

## Цель текущего плана

Реализовать in-process SSE broadcast для форума: новые посты доставляются другим участникам топика без перезагрузки страницы. Закрыть все CHK-* из feature.md.

---

## Discovery Context / Reference Points

| Path / module | Current role | Why relevant | Reuse / mirror |
| --- | --- | --- | --- |
| `internal/forum/handler.go` | CRUD-хэндлеры форума | `CreatePost` — точка вставки `hub.Broadcast`; структура `Handler` расширяется полем `hub *Hub` | Паттерн `c.Render(status, "tpl#block", data)`, HTMX-detection через `HX-Request` |
| `internal/forum/service.go` | Бизнес-логика; не меняется | Hub не проходит через service (CON из feature.md) | Вызов `h.svc.GetTopicWithPosts()` возвращает `[]PostView` — из него извлекается новый пост по `postID` для рендера фрагмента |
| `internal/forum/model.go` | Типы `Post`, `PostView`, `Topic` | `PostView` (с `AuthorUsername`) — тип данных для шаблона `post.html` | Передавать `PostView` в шаблон, не `Post` |
| `internal/tmpl/renderer.go` | Echo renderer + `RenderPartial` | `RenderPartial(w, pageKey, blockName, data)` — рендер partial в `bytes.Buffer` для hub.Broadcast | Менять тип поля `renderer` в `Handler` с `echo.Renderer` → `*tmpl.Renderer` чтобы вызвать `RenderPartial` напрямую |
| `cmd/forum/main.go:112` | Инициализация `forumHandler` | Точка wire: создать `hub`, передать в конструктор, зарегистрировать маршрут SSE, добавить `e.Static("/static", "static")` | Паттерн инициализации рядом с `forumSvc` |
| `templates/forum/topic.html:249-299` | Рендер списка постов (блок `posts-list`) | Инлайн-рендер постов переезжает в `partials/post.html`; контейнер `#posts-list` получает SSE-атрибуты | Существующий CSS и структура поста (`div.post`, reply-блок) |
| `migrations/011_*.sql` | Последняя миграция | Следующий номер — `012` | Паттерн `CREATE INDEX CONCURRENTLY` |
| `internal/forum/handler_test.go` | Integration тесты (`//go:build integration`) | SSE тесты идут сюда; использует `httptest`, real DB, `sync.Once` | `testDB()`, `insertUser()` из `test_helpers.go` |
| `internal/forum/service_test.go` | Unit тесты с `mockRepo` | `hub_test.go` — unit, в пакете `forum` | `mockRepo` pattern с func-полями |
| `deploy/nginx/` | Reverse-proxy конфиг | Нужен `proxy_buffering off` для SSE location (CON из feature.md) | Существующий блок `location` |

---

## Test Strategy

| Test surface | Canonical refs | Existing coverage | Planned automated coverage | Required local suites | CI suites | Manual-only gap / justification | Approval ref |
| --- | --- | --- | --- | --- | --- | --- | --- |
| Hub unit: subscribe/broadcast/unsubscribe/ctx cancel | `REQ-01`, `REQ-02`, `CHK-03` | нет | `internal/forum/hub_test.go` (unit, пакет `forum`) | `docker run ... go test ./internal/forum/` | нет CI | — | none |
| SSE endpoint + live delivery | `REQ-01`, `REQ-02`, `CHK-01`, `CHK-04`, `CHK-05`, `CHK-06` | нет | `internal/forum/handler_test.go` (integration) | `docker run ... go test -tags integration ./internal/forum/` | нет CI | — | none |
| Catch-up с `Last-Event-ID` | `REQ-03`, `CHK-02` | нет | `internal/forum/handler_test.go` (integration) | то же | нет CI | — | none |
| Шаблонный тест атрибутов topic.html | `CHK-07` | нет | `templates/forum/topic_test.go` | `docker run ... go test ./templates/forum/` | нет CI | — | none |
| Ручная приёмка SC-01, SC-02 | `SC-01`, `SC-02` | — | не автоматизировано (требует два браузера) | — | — | Нет headless-фреймворка; playwright отсутствует в стеке | `AG-01` |

---

## Open Questions / Ambiguities

| OQ ID | Question | Why unresolved | Blocks | Default action |
| --- | --- | --- | --- | --- |
| `OQ-01` | Источник `sse.js` (htmx-ext-sse): vendor в `static/js/` или CDN? | feature.md указывает `/static/js/sse.js` — значит локальный файл, но версия не зафиксирована | `STEP-06` | Скачать актуальный релиз `sse.js` из репозитория `bigskysoftware/htmx-ext-sse` и положить в `static/js/sse.js` |
| `OQ-02` | nginx-конфиг: расположение файла и имя location-блока | `deploy/nginx/` не исследован детально | `STEP-10` | Найти существующий `location /forum` или добавить отдельный location для `/forum/topics/*/events` с `proxy_buffering off` |

---

## Environment Contract

| Area | Contract | Used by | Failure symptom |
| --- | --- | --- | --- |
| setup | `DATABASE_URL` — env var с реальной БД; goose миграции прогоняются в `setupPool` | все integration тесты | `t.Skip("DATABASE_URL not set")` |
| build | `docker run --rm -v $(pwd):/app -w /app golang:1.23-alpine go build ./...` | все шаги компиляции | build error |
| test (unit) | `docker run ... go test ./internal/forum/` | `STEP-02` | тест не компилируется |
| test (integration) | `docker run ... go test -tags integration ./internal/forum/` + `DATABASE_URL` | `STEP-09` | тест skip или fail |
| migration | `docker compose run --rm app goose up` или через тест `setupPool` | `STEP-08` | индекс не создан, catch-up медленнее |

---

## Preconditions

| PRE ID | Canonical ref | Required state | Used by steps | Blocks start |
| --- | --- | --- | --- | --- |
| `PRE-01` | `feature.md status: active` | Выполнено | все | no |
| `PRE-02` | `CON-07` / `architecture.md` | ADR in-process hub зафиксирован в architecture.md | все | no (уже выполнено) |
| `PRE-03` | `ASM-01` | Single Go-process deployment | все | no |
| `PRE-04` | `CON-01` | No external SSE library — только stdlib | `STEP-01` | yes |

---

## Workstreams

| Workstream | Implements | Result | Owner | Dependencies |
| --- | --- | --- | --- | --- |
| `WS-1` | `REQ-01`, `REQ-02`, `CON-04`, `CON-05`, `FM-02`, `FM-03` | `hub.go` + `hub_test.go` | agent | — |
| `WS-2` | `CTR-03`, `CON-02` | `templates/forum/partials/post.html` + обновлённый `topic.html` | agent | — |
| `WS-3` | `REQ-01`, `REQ-02`, `REQ-03`, `CTR-01`, `CTR-02`, `FM-01`, `FM-04` | Обновлённый `handler.go` (`StreamEvents`, `CreatePost` + broadcast) | agent | WS-1, WS-2 |
| `WS-4` | — | Обновлённый `main.go` (wire hub, маршрут SSE, static route) | agent | WS-3 |
| `WS-5` | `CTR-02` | `migrations/012_idx_posts_topic_id.sql` | agent | — |
| `WS-6` | `ASM-02` | `static/js/sse.js` (htmx-ext-sse) | agent | OQ-01 |
| `WS-7` | `CHK-01`..`CHK-07` | Integration + unit тесты зелёные | agent | WS-3, WS-4, WS-5, WS-6 |
| `WS-8` | `CON из feature.md (nginx)` | nginx конфиг с `proxy_buffering off` | agent | OQ-02 |

---

## Approval Gates

| AG ID | Trigger | Applies to | Why approval required | Approver / evidence |
| --- | --- | --- | --- | --- |
| `AG-01` | Финальная приёмка SC-01, SC-02 | `WS-7`, closure gate | Требует двух браузеров / пользователей; не покрывается automated тестами | Человек (автор задачи) подтверждает: «SC-01 и SC-02 проверены вручную» |

---

## Порядок работ

| Step ID | Actor | Implements | Goal | Touchpoints | Artifact | Verifies | Evidence IDs | Check command / procedure | Blocked by | Needs approval | Escalate if |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `STEP-01` | agent | `CON-04`, `CON-05`, `FM-02`, `FM-03` | Создать `internal/forum/hub.go`: `Hub` со `sync.RWMutex`, `map[int64][]subscriber`, `Subscribe(ctx, topicID, userID) (<-chan string, error)`, `Broadcast(topicID, authorUserID int64, fragment string)`, `unsubscribe` при `ctx.Done()`. **Broadcast** запускает запись в каждый канал через горутину с `defer recover()` — защита от паники при записи в закрытый канал (CON-04) | `internal/forum/hub.go` (new) | `hub.go` компилируется | — | — | `docker run ... go build ./internal/forum/` | PRE-04 | none | panic в broadcast или утечка горутин |
| `STEP-02` | agent | `CHK-03`, `CHK-04` | Создать `internal/forum/hub_test.go`: unit-тесты Subscribe/Broadcast/ctx-cancel/topic-isolation/limit | `internal/forum/hub_test.go` (new) | тесты зелёные | `CHK-03` | `EVID-01` | `docker run ... go test ./internal/forum/ -run TestHub` | STEP-01 | none | тест падает из-за race |
| `STEP-03` | agent | `CTR-03`, `CON-02` | Создать `templates/forum/partials/post.html`: `{{define "post"}}` с корневым тегом `<article class="post" id="post-{{.ID}}">`, тот же HTML что сейчас инлайн в topic.html (reply-блок, контент, автор) | `templates/forum/partials/post.html` (new) | partial шаблон | — | — | визуальная проверка рендера `ShowTopic` | — | none | partial не подхватывается renderer |
| `STEP-04` | agent | `ASM-02`, `CTR-01`, `CHK-07` | Обновить `topic.html`: (a) заменить инлайн-рендер постов на `{{template "post" .}}` через `{{range .Posts}}`; (b) добавить на `<div id="posts-list">` атрибуты `hx-ext="sse"`, `sse-connect="/forum/topics/{{.Topic.ID}}/events"`, `sse-swap="post-added"`, `hx-swap="beforeend"`; (c) добавить `<script src="/static/js/sse.js">` | `templates/forum/topic.html` | обновлённый topic.html | `CHK-07` | `EVID-03` | визуальная проверка + шаблонный тест атрибутов | STEP-03 | none | сломался initial render постов |
| `STEP-05` | agent | `REQ-01`, `REQ-02`, `CTR-01`, `CTR-02`, `FM-01`, `FM-03`, `FM-04` | Обновить `handler.go`: (a) поле `renderer echo.Renderer` → `renderer *tmpl.Renderer`; (b) добавить поле `hub *Hub`; (c) обновить `NewHandler`; (d) добавить метод `StreamEvents`: проверка топика → `hub.Subscribe` → headers SSE → catch-up из БД (LIMIT 50, ORDER BY id ASC) → если кол-во catch-up постов == 50, отправить фрейм `event: catch-up-overflow\ndata:\n\n` (CTR-02) → loop `select {case msg: write frame; case ctx.Done: return}`; (e) в `CreatePost` success: найти новый `PostView` по `postID` в результате `GetTopicWithPosts` → рендер в `bytes.Buffer` через `h.renderer.RenderPartial` → `strings.NewReplacer("\n"," ","\r"," ")` → `h.hub.Broadcast` | `internal/forum/handler.go` | обновлённый handler.go | `CHK-01`, `CHK-02`, `CHK-05`, `CHK-06` | `EVID-01`, `EVID-02`, `EVID-03` | `docker run ... go build ./...` | STEP-01, STEP-03 | none | deadlock в broadcast или goroutine leak |
| `STEP-06` | agent | `ASM-02` | Скачать `sse.js` из `bigskysoftware/htmx-ext-sse` (latest release) → положить в `static/js/sse.js` | `static/js/sse.js` (new) | файл в репозитории | — | — | `curl -f http://localhost:PORT/static/js/sse.js` (после STEP-07) | OQ-01 | none | 404 на /static/js/sse.js |
| `STEP-07` | agent | — | Обновить `main.go`: (a) `forumHub := forum.NewHub()`; (b) `forumHandler := forum.NewHandler(forumSvc, renderer, forumHub)`; (c) `e.GET("/forum/topics/:id/events", forumHandler.StreamEvents)`; (d) `e.Static("/static", "static")` | `cmd/forum/main.go` | обновлённый main.go | — | — | `docker run ... go build ./cmd/forum/` | STEP-05 | none | import cycle или маршрут не регистрируется |
| `STEP-08` | agent | `CTR-02` | Создать `migrations/012_idx_posts_topic_id.sql` с `CREATE INDEX CONCURRENTLY idx_posts_topic_id_id ON forum_posts(topic_id, id);` (goose: `-- +goose Up` / `-- +goose Down`) | `migrations/012_idx_posts_topic_id.sql` (new) | миграция применяется | — | — | `goose status` или через `setupPool` в тестах | — | none | `CONCURRENTLY` не поддерживается в транзакции goose — использовать `-- +goose NO TRANSACTION` |
| `STEP-09` | agent | `CHK-01`..`CHK-07` | (a) Добавить integration тесты в `handler_test.go`: `TestStreamEvents_LiveDelivery` (CHK-01), `TestStreamEvents_CatchUp` (CHK-02), `TestStreamEvents_CatchUpOverflow` (CTR-02 overflow: создать 51 пост, подключиться с `Last-Event-ID=0`, проверить что получен `event: catch-up-overflow` после первых 50), `TestStreamEvents_TopicIsolation` (CHK-04), `TestStreamEvents_Unauthenticated` (CHK-05), `TestStreamEvents_SubscriberLimit` (CHK-06). (b) Создать `templates/forum/topic_test.go` (пакет `templates_test`) с шаблонным тестом атрибутов topic.html (CHK-07) — рендер `topic.html` через `html/template`, проверка наличия `hx-ext="sse"`, `sse-connect`, `sse-swap="post-added"`, `hx-swap="beforeend"` в HTML | `internal/forum/handler_test.go`, `templates/forum/topic_test.go` (new) | тесты зелёные | `CHK-01`..`CHK-07` | `EVID-01`, `EVID-02`, `EVID-03` | `docker run ... go test -tags integration ./internal/forum/` && `docker run ... go test ./templates/forum/` | STEP-05, STEP-07, STEP-08 | none | тест flaky по таймингу — увеличить таймаут чтения SSE-фрейма |
| `STEP-10` | agent | nginx CON | Добавить в nginx конфиг location для `/forum/topics/` с `proxy_buffering off; proxy_read_timeout 3600s;` | `deploy/nginx/*.conf` | обновлённый nginx конфиг | — | — | визуальная проверка или `nginx -t` | OQ-02 | none | SSE-фреймы не доходят до браузера через proxy |

---

## Parallelizable Work

- `PAR-01` STEP-01/02 (WS-1), STEP-03 (WS-2), STEP-08 (WS-5), STEP-06 (WS-6) — независимы, можно делать параллельно.
- `PAR-02` STEP-05 (handler.go) зависит от STEP-01 (hub) и STEP-03 (partial). Нельзя начинать до завершения обоих.
- `PAR-03` STEP-07 (main.go) зависит от STEP-05. После — STEP-09 (тесты) зависит от всего.

---

## Checkpoints

| CP ID | Refs | Condition | Evidence IDs |
| --- | --- | --- | --- |
| `CP-01` | STEP-01, STEP-02 | Hub компилируется; unit-тесты hub зелёные | `EVID-01` |
| `CP-02` | STEP-04, STEP-05, STEP-07 | `go build ./...` зелёный; `ShowTopic` рендерит посты корректно (визуально) | — |
| `CP-03` | STEP-09 | Все integration тесты зелёные | `EVID-01`, `EVID-02`, `EVID-03` |

---

## Execution Risks

| Risk ID | Risk | Impact | Mitigation | Trigger |
| --- | --- | --- | --- | --- |
| `ER-01` | `CREATE INDEX CONCURRENTLY` не работает внутри goose-транзакции | миграция падает | Добавить `-- +goose NO TRANSACTION` в файл миграции | goose error при `Up` |
| `ER-02` | Goroutine leak в `StreamEvents` при быстром disconnect клиента | память растёт | `ctx.Done()` должен быть в select-ветке, `defer hub.unsubscribe(...)` | тест CHK-03 / `EC-03` |
| `ER-03` | Шаблонный ключ `templates/forum/partials/post.html` не совпадает с ожидаемым в `RenderPartial` | runtime error при broadcast | Проверить, что renderer подхватывает файл при старте; pageKey = `"templates/forum/partials/post.html"`, blockName = `"post"` | panic или error при первом Broadcast |
| `ER-04` | `sse.js` version mismatch с htmx версией в topic.html | SSE не работает в браузере | Проверить совместимость версий htmx и htmx-ext-sse | sse-connect не устанавливается |

---

## Stop Conditions / Fallback

| Stop ID | Related refs | Trigger | Immediate action | Safe fallback state |
| --- | --- | --- | --- | --- |
| `STOP-01` | `ER-02`, `FM-03` | Горутина StreamEvents не завершается при отмене контекста (тест CHK-03 красный) | Остановить STEP-09; исправить hub.Unsubscribe логику в STEP-01 | Без SSE-эндпоинта; topic.html без SSE-атрибутов |
| `STOP-02` | `ER-03` | `RenderPartial` возвращает ошибку при первом Broadcast | Логировать и пропускать broadcast (пост сохранён); расследовать template key | Forum работает без SSE (деградация, не падение) |

---

## Готово для приёмки

- Все CHK-01..CHK-07 зелёные (automated).
- `EC-03` подтверждён через CHK-03 (unit-тест).
- SC-01, SC-02 подтверждены человеком вручную (`AG-01`).
- `feature.md` → `delivery_status: in_progress` (при старте реализации) → `done` (после приёмки).
- `implementation-plan.md` → `status: archived`.
