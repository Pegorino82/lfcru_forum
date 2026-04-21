---
title: "FT-016: Real-time обновление постов форума (SSE)"
doc_kind: feature
doc_function: canonical
purpose: "Реализация real-time доставки новых постов форума через Server-Sent Events: пользователь видит посты других участников без перезагрузки страницы."
derived_from:
  - ../../domain/problem.md
  - ../../domain/architecture.md
  - ../../domain/frontend.md
status: active
delivery_status: planned
audience: humans_and_agents
must_not_define:
  - implementation_sequence
---

# FT-016: Real-time обновление постов форума (SSE)

## What

### Problem

Пользователи, одновременно просматривающие один топик форума, не получают новые посты без ручной перезагрузки страницы. SSE указан в архитектуре как механизм real-time, но инфраструктура для форума не реализована.

### Scope

- `REQ-01` Когда пользователь А публикует пост в топике, пользователь Б (открытый тот же топик) видит новый пост без перезагрузки страницы.
- `REQ-02` При открытии SSE-соединения пользователь подписывается на конкретный топик по его ID.
- `REQ-03` После разрыва и восстановления SSE-соединения пользователь не теряет посты, появившиеся во время разрыва (используется `Last-Event-ID`).

### Non-Scope

- `NS-01` WebSocket — используем SSE согласно архитектурному решению.
- `NS-02` Real-time для комментариев к новостям — отдельная задача.
- `NS-03` Push-уведомления в браузере (Notification API) — вне MVP.
- `NS-04` Масштабирование hub между несколькими инстансами (multi-pod) — вне текущего scope.
- `NS-05` Отображение числа онлайн-пользователей в топике.
- `NS-06` Real-time обновление/удаление уже опубликованных постов — вне MVP (события доставляются только при публикации).

### Constraints / Assumptions

- `ASM-01` Приложение запускается как один Go-процесс — in-process broadcast hub достаточен.
- `ASM-02` HTMX SSE extension (`htmx-ext-sse`) используется на клиенте для прослушивания событий и вставки постов.
- `CON-01` SSE реализуется через stdlib (`http.Flusher`) без внешних зависимостей.
- `CON-02` Каждый SSE-ответ содержит HTML-фрагмент нового поста (прямой append через `hx-ext="sse"`) — не JSON.
- `CON-03` SSE-эндпоинт `GET /forum/topics/:id/events` доступен всем пользователям (включая неаутентифицированных) — форумные топики публичны, real-time события публичны в той же мере.
- `CON-04` Hub потокобезопасен: все операции с internal map защищены `sync.RWMutex`; write lock при subscribe/unsubscribe, read lock при broadcast. Broadcast-горутина защищена `defer recover()` от паники при записи в закрытый канал.
- `CON-05` Максимум 200 SSE-подписчиков на один топик (подписчик = одно активное SSE-соединение, т.е. одна открытая вкладка браузера). При превышении лимита `StreamEvents` возвращает `503 Service Unavailable` до установки SSE-соединения.
- `CON-06` Сессия пользователя проверяется только при установке SSE-соединения. Инвалидация сессии (logout, бан) во время активного соединения не закрывает его автоматически — **известное ограничение MVP**.
- `CON-07` **ADR**: Для MVP (ASM-01, single Go-process) используется in-process broadcast hub вместо PostgreSQL `LISTEN/NOTIFY`, описанного в `architecture.md`. `architecture.md` должен быть обновлён с этим решением до начала реализации.

## How

### Solution

Добавить in-process broadcast hub (`forum/hub.go`): map `topicID → []chan string`. `CreatePost` после записи в БД рендерит HTML-фрагмент через `templates/forum/partials/post.html`, заменяет `\n`/`\r` пробелами и вызывает `hub.Broadcast(topicID, authorUserID, fragment)` — канал(ы) автора пропускаются, чтобы избежать дублирования поста в его DOM. `GET /forum/topics/:id/events` — SSE endpoint, который подписывает клиента на канал топика и стримит события. Доступен без аутентификации (CON-03).

На клиенте контейнер постов размечается: `hx-ext="sse" sse-connect="/forum/topics/{{.ID}}/events" sse-swap="post-added" hx-swap="beforeend"`. Расширение htmx-ext-sse подключается через `<script src="/static/js/sse.js">`.

### Change Surface

| Surface | Type | Why it changes |
| --- | --- | --- |
| `internal/forum/hub.go` | code (new) | In-process broadcast hub: subscribe/unsubscribe/broadcast по topicID. Внутренняя структура подписчика: `struct{ userID int64; ch chan string }`. Анонимные пользователи подписываются с `userID = 0`; значение 0 никогда не совпадает с реальным authorUserID, поэтому анонимы всегда получают все события — дополнительная логика не нужна. |
| `internal/forum/handler.go` | code | `CreatePost`: вызвать hub.Broadcast после успешной записи; добавить метод `StreamEvents` |
| `internal/forum/service.go` | code | Не изменяется: hub не проходит через service-слой |
| `internal/forum/handler.go` (constructor) | code | `forumHandler` получает `hub *Hub` как поле; конструктор принимает hub явно (dependency injection) |
| `cmd/forum/main.go` | code | Создать hub, передать в forumHandler, зарегистрировать маршрут `GET /forum/topics/:id/events` |
| `templates/forum/topic.html` | code | Добавить `hx-ext="sse"`, `sse-connect="/forum/topics/{{.ID}}/events"`, `sse-swap="post-added"`, `hx-swap="beforeend"` на контейнер постов; подключить htmx-ext-sse |
| `templates/forum/partials/post.html` | code (new) | Partial-шаблон одного поста (`<article class="post" id="post-{{.ID}}">`) — используется и для initial page render, и для SSE-фрагментов |
| `deploy/nginx/` (location `/forum/topics/*/events`) | config | Добавить `proxy_buffering off;` — без этого nginx буферизует SSE-поток. Заголовок `X-Accel-Buffering: no` выставляет приложение (CTR-01); дублировать через `proxy_set_header` не нужно. |
| `migrations/XXXX_idx_posts_topic_id.sql` | migration (new) | `CREATE INDEX CONCURRENTLY idx_posts_topic_id_id ON posts(topic_id, id)` для catch-up запроса |

### Flow

1. Пользователь Б открывает `/forum/topics/:id` — браузер устанавливает SSE-соединение на `/forum/topics/:id/events`.
2. Hub регистрирует канал для Б в map под ключом `topicID`.
3. Пользователь А публикует пост: `CreatePost` записывает в БД, рендерит HTML-фрагмент через `post.html`, заменяет `\n`/`\r` пробелами, вызывает `hub.Broadcast(topicID, authorUserID, fragment)`.
4. Hub отправляет фрагмент во все каналы для `topicID`, пропуская канал(ы) с `userID == authorUserID`.
5. SSE endpoint пишет `data: <html-fragment>\n\n` в ответ Б.
6. HTMX SSE extension принимает событие, вставляет фрагмент в контейнер постов.
7. Пользователь Б видит новый пост без перезагрузки.

### Contracts

| Contract ID | Input / Output | Producer / Consumer | Notes |
| --- | --- | --- | --- |
| `CTR-01` | SSE event с полями `id`, `event`, `data`; HTTP-заголовки ответа | `StreamEvents` / HTMX SSE ext | Поля фрейма: `id: <post_db_id>`, `event: post-added`, `data: <html>` (HTML-фрагмент одного поста в одну строку — символы `\n` и `\r` заменены пробелами перед отправкой). Обязательные HTTP-заголовки ответа: `Content-Type: text/event-stream`, `Cache-Control: no-cache`, `Connection: keep-alive`, `X-Accel-Buffering: no`. |
| `CTR-02` | `Last-Event-ID` header | клиент / `StreamEvents` | Database post ID (integer) последнего полученного поста. Валидация: если значение не парсится как int64 > 0 — заголовок игнорируется (catch-up не выполняется, только live). Запрос: `WHERE topic_id = :id AND id > last_event_id ORDER BY id ASC LIMIT 50`. При наличии более 50 пропущенных постов клиент получает первые 50 от точки отрыва (в хронологическом порядке); после них сервер отправляет `event: catch-up-overflow` как сигнал для UI-подсказки "Обновите страницу для полной истории". |
| `CTR-03` | HTML-фрагмент поста | `handler.go` / `templates/forum/partials/post.html` | Корневой тег `<article class="post" id="post-{{.ID}}">`. Один и тот же шаблон используется для catch-up и live-событий. Шаблон Go `html/template` — автоматическое HTML-экранирование пользовательского контента (не использовать `template.HTML`). |

### Failure Modes

- `FM-01` SSE-соединение обрывается (сеть, таймаут): пользователь видит существующие посты, не получает новые. После переподключения HTMX SSE extension пересоединяется автоматически с `Last-Event-ID`.
- `FM-02` Hub переполнен (медленный клиент): использовать буферизованный канал размером ≥ 16 сообщений (достаточен для типичного всплеска активности) с `select { case ch <- msg: default: }` — дроп допустим только при исчерпании буфера; медленный клиент теряет события, не блокирует других. Восстановление пропущенных событий через `Last-Event-ID` возможно только при переподключении, но не при дропе в рамках живого соединения.
- `FM-03` Горутина `StreamEvents` утекает при закрытии соединения: hub должен unsubscribe при `ctx.Done()`, удалить канал из slice для `topicID`; если slice становится пустым — удалить ключ `topicID` из map во избежание накопления мёртвых записей.
- `FM-04` Ошибки до установки SSE-соединения возвращаются как стандартные HTTP-ответы (не SSE-фреймы): топик не существует → `404 Not Found`; превышен лимит подписчиков (CON-05) → `503 Service Unavailable`; внутренняя ошибка → `500 Internal Server Error`. `Content-Type: text/plain`.
- `FM-05` Глобальный лимит SSE-соединений по всем топикам — вне текущего scope (NS-04, single Go-process). Защита от перегрузки при масштабировании возложена на будущий multi-pod механизм.

## Verify

### Exit Criteria

- `EC-01` Пользователь Б получает новый пост в открытом топике без перезагрузки ≤ 2 секунды после публикации пользователем А.
- `EC-02` При разрыве и восстановлении SSE-соединения посты, опубликованные во время разрыва, доставляются через `Last-Event-ID`.
- `EC-03` После закрытия соединения горутина `StreamEvents` завершается, hub не хранит канал клиента (нет утечки).

### Traceability matrix

| Requirement ID | Design refs | Acceptance refs | Checks | Evidence IDs |
| --- | --- | --- | --- | --- |
| `REQ-01` | `ASM-01`, `ASM-02`, `CON-01`, `CON-02`, `CTR-01`, `FM-01`, `FM-02` | `EC-01`, `SC-01` | `CHK-01` | `EVID-01` |
| `REQ-02` | `CTR-01`, `CTR-02` | `EC-01`, `SC-01` | `CHK-01`, `CHK-04` | `EVID-01` |
| `REQ-03` | `CTR-02` | `EC-02`, `SC-02` | `CHK-02` | `EVID-02` |

### Acceptance Scenarios

- `SC-01` Пользователи А и Б открывают один топик. А публикует пост. Б видит новый пост в своём браузере без перезагрузки, не более чем через 2 секунды.
- `SC-02` Б имеет открытое SSE-соединение. Сетевое соединение прерывается на 5 секунд (симуляция сетевого сбоя — браузер обнаруживает disconnect и разрывает EventSource). За это время А публикует пост. После автоматического восстановления HTMX SSE extension отправляет `Last-Event-ID`, Б получает пропущенный пост.

### Negative / Edge Cases

- `NEG-01` Клиент разрывает SSE-соединение (закрывает вкладку): hub удаляет канал, горутина завершается без паники.
- `NEG-02` В топике нет активных SSE-клиентов: `hub.Broadcast` не создаёт лишних аллокаций.
- `NEG-03` Неаутентифицированный пользователь открывает `/forum/topics/:id/events` — SSE-соединение устанавливается (200), события доставляются в обычном режиме (CON-03).
- `NEG-04` Превышен лимит подписчиков на топик (CON-05) — новый запрос получает `503`, существующие соединения не затрагиваются.

### Checks

| Check ID | Covers | How to check | Expected result |
| --- | --- | --- | --- |
| `CHK-01` | `EC-01`, `SC-01` | Интеграционный тест: открыть SSE HTTP-эндпоинт через `httptest`, вызвать CreatePost, прочитать SSE-фрейм из response body | HTTP-ответ: `Content-Type: text/event-stream`; фрейм содержит `event: post-added`, `id: <post_id>`, `data: <html-fragment>` |
| `CHK-02` | `EC-02`, `SC-02` | Интеграционный тест: создать посты, затем подключиться с `Last-Event-ID` меньшим ID последнего поста; проверить catch-up из БД | Пропущенные посты доставлены в порядке возрастания ID до начала live-стриминга |
| `CHK-03` | `EC-03`, `NEG-01` | Unit-тест hub: Subscribe → Cancel context → проверить что канал удалён и ключ topicID удалён из map если подписчиков не осталось | Hub не содержит канал и не содержит пустой записи topicID после ctx.Done() |
| `CHK-04` | `REQ-02` | Интеграционный тест: клиент А подписан на топик 1, клиент Б — на топик 2; пост публикуется в топик 2 | Клиент А не получает событие в течение 200 мс; клиент Б получает |
| `CHK-05` | `NEG-03`, `CON-03` | Интеграционный тест: GET `/forum/topics/:id/events` без сессионной куки | HTTP 200, `Content-Type: text/event-stream` — соединение установлено |
| `CHK-06` | `NEG-04`, `CON-05` | Интеграционный тест: открыть 201 соединений на один топик | 201-й запрос получает `503 Service Unavailable` |
| `CHK-07` | `CTR-03` | Шаблонный тест: render `topic.html` и проверить атрибуты в HTML | Присутствуют `hx-ext="sse"`, `sse-connect`, `sse-swap="post-added"`, `hx-swap="beforeend"` |

### Test matrix

| Check ID | Evidence IDs | Evidence path |
| --- | --- | --- |
| `CHK-01` | `EVID-01` | `internal/forum/hub_test.go`, `internal/forum/handler_test.go` |
| `CHK-02` | `EVID-02` | `internal/forum/handler_test.go` |
| `CHK-03` | `EVID-01` | `internal/forum/hub_test.go` |
| `CHK-04` | `EVID-01` | `internal/forum/handler_test.go` |
| `CHK-05` | `EVID-03` | `internal/forum/handler_test.go` |
| `CHK-06` | `EVID-03` | `internal/forum/handler_test.go` |
| `CHK-07` | `EVID-03` | `templates/forum/topic_test.go` |

### Evidence

- `EVID-01` Юнит-тесты hub + интеграционные тесты StreamEvents — зелёные.
- `EVID-02` Интеграционный тест reconnect с `Last-Event-ID` — зелёный.
- `EVID-03` Интеграционные тесты доступа без сессии, лимита подписчиков и шаблонный тест атрибутов — зелёные.

### Evidence contract

| Evidence ID | Artifact | Producer | Path contract | Reused by checks |
| --- | --- | --- | --- | --- |
| `EVID-01` | Вывод `go test ./internal/forum/...` и `go test -tags integration ./internal/forum/...` | CI / локальный запуск | `internal/forum/hub_test.go`, `internal/forum/handler_test.go` | `CHK-01`, `CHK-03`, `CHK-04` |
| `EVID-02` | Вывод `go test -tags integration ./internal/forum/...` | CI / локальный запуск | `internal/forum/handler_test.go` | `CHK-02` |
| `EVID-03` | Вывод `go test -tags integration ./internal/forum/...` и `go test ./templates/forum/...` | CI / локальный запуск | `internal/forum/handler_test.go`, `templates/forum/topic_test.go` | `CHK-05`, `CHK-06`, `CHK-07` |
