---
title: "FT-016: Real-time обновление постов форума (SSE)"
doc_kind: feature
doc_function: canonical
purpose: "Реализация real-time доставки новых постов форума через Server-Sent Events: пользователь видит посты других участников без перезагрузки страницы."
derived_from:
  - ../../domain/problem.md
  - ../../domain/architecture.md
  - ../../domain/frontend.md
status: draft
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

### Constraints / Assumptions

- `ASM-01` Приложение запускается как один Go-процесс — in-process broadcast hub достаточен.
- `ASM-02` HTMX SSE extension (`htmx-ext-sse`) используется на клиенте для прослушивания событий и вставки постов.
- `CON-01` SSE реализуется через stdlib (`http.Flusher`) без внешних зависимостей.
- `CON-02` Каждый SSE-ответ содержит HTML-фрагмент нового поста (HTMX out-of-band или прямой append) — не JSON.
- `DEC-01` Формат SSE-события и HTMX-интеграция требуют выбора между `hx-swap-oob` и `hx-ext="sse"` — решение не принято, блокирует `How`.

## How

### Solution

Добавить in-process broadcast hub (`forum/hub.go`): map `topicID → []chan string`. `CreatePost` после записи в БД пушит HTML-фрагмент поста в hub. `GET /forum/topics/:id/events` — SSE endpoint, который подписывает клиента на канал топика и стримит события.

На клиенте: подключить HTMX SSE extension, добавить `hx-ext="sse" sse-connect="/forum/topics/:id/events"` на контейнер постов, настроить `sse-swap` для append новых постов.

### Change Surface

| Surface | Type | Why it changes |
| --- | --- | --- |
| `internal/forum/hub.go` | code (new) | In-process broadcast hub: subscribe/unsubscribe/broadcast по topicID |
| `internal/forum/handler.go` | code | `CreatePost`: вызвать hub.Broadcast после успешной записи; добавить метод `StreamEvents` |
| `internal/forum/service.go` | code | Передать hub в handler или сделать hub частью handler |
| `cmd/forum/main.go` | code | Создать hub, передать в forumHandler, зарегистрировать маршрут `GET /forum/topics/:id/events` |
| `templates/forum/topic.html` | code | Добавить `hx-ext="sse"`, `sse-connect`, `sse-swap` на контейнер постов |

### Flow

1. Пользователь Б открывает `/forum/topics/:id` — браузер устанавливает SSE-соединение на `/forum/topics/:id/events`.
2. Hub регистрирует канал для Б в map под ключом `topicID`.
3. Пользователь А публикует пост: `CreatePost` записывает в БД, рендерит HTML-фрагмент поста, вызывает `hub.Broadcast(topicID, fragment)`.
4. Hub отправляет фрагмент во все каналы для `topicID`.
5. SSE endpoint пишет `data: <html-fragment>\n\n` в ответ Б.
6. HTMX SSE extension принимает событие, вставляет фрагмент в контейнер постов.
7. Пользователь Б видит новый пост без перезагрузки.

### Contracts

| Contract ID | Input / Output | Producer / Consumer | Notes |
| --- | --- | --- | --- |
| `CTR-01` | SSE event: `data: <html>\n\n` | `StreamEvents` / HTMX SSE ext | HTML-фрагмент одного поста; `event: post-added` для HTMX `sse-swap` |
| `CTR-02` | `Last-Event-ID` header | клиент / `StreamEvents` | ID последнего полученного поста; handler догоняет посты из БД |

### Failure Modes

- `FM-01` SSE-соединение обрывается (сеть, таймаут): пользователь видит существующие посты, не получает новые. После переподключения HTMX SSE extension пересоединяется автоматически с `Last-Event-ID`.
- `FM-02` Hub переполнен (медленный клиент): использовать небуферизованный или буферизованный канал с `select { case ch <- msg: default: }` — медленный клиент теряет события, не блокирует других.
- `FM-03` Горутина `StreamEvents` утекает при закрытии соединения: hub должен unsubscribe при `ctx.Done()`.

## Verify

### Exit Criteria

- `EC-01` Пользователь Б получает новый пост в открытом топике без перезагрузки ≤ 2 секунды после публикации пользователем А.
- `EC-02` При разрыве и восстановлении SSE-соединения посты, опубликованные во время разрыва, доставляются через `Last-Event-ID`.
- `EC-03` После закрытия соединения горутина `StreamEvents` завершается, hub не хранит канал клиента (нет утечки).

### Traceability matrix

| Requirement ID | Design refs | Acceptance refs | Checks | Evidence IDs |
| --- | --- | --- | --- | --- |
| `REQ-01` | `ASM-01`, `ASM-02`, `CON-01`, `CON-02`, `CTR-01`, `FM-01`, `FM-02` | `EC-01`, `SC-01` | `CHK-01` | `EVID-01` |
| `REQ-02` | `CTR-01`, `CTR-02` | `EC-01`, `SC-01` | `CHK-01` | `EVID-01` |
| `REQ-03` | `CTR-02` | `EC-02`, `SC-02` | `CHK-02` | `EVID-02` |

### Acceptance Scenarios

- `SC-01` Пользователи А и Б открывают один топик. А публикует пост. Б видит новый пост в своём браузере без перезагрузки, не более чем через 2 секунды.
- `SC-02` Б имеет открытое SSE-соединение. Соединение обрывается на 5 секунд. За это время А публикует пост. После восстановления соединения Б получает пропущенный пост через `Last-Event-ID`.

### Negative / Edge Cases

- `NEG-01` Клиент разрывает SSE-соединение (закрывает вкладку): hub удаляет канал, горутина завершается без паники.
- `NEG-02` В топике нет активных SSE-клиентов: `hub.Broadcast` не создаёт лишних аллокаций.

### Checks

| Check ID | Covers | How to check | Expected result |
| --- | --- | --- | --- |
| `CHK-01` | `EC-01`, `SC-01` | Интеграционный тест: открыть SSE-соединение, вызвать CreatePost, прочитать событие из канала | Событие получено, содержит HTML-фрагмент нового поста |
| `CHK-02` | `EC-02`, `SC-02` | Интеграционный тест: имитировать reconnect с `Last-Event-ID`, проверить dogfooding из БД | Пропущенные посты доставлены |
| `CHK-03` | `EC-03`, `NEG-01` | Unit-тест hub: Subscribe → Cancel context → проверить что канал удалён | Hub не содержит канал после ctx.Done() |

### Test matrix

| Check ID | Evidence IDs | Evidence path |
| --- | --- | --- |
| `CHK-01` | `EVID-01` | `internal/forum/hub_test.go`, `internal/forum/handler_test.go` |
| `CHK-02` | `EVID-02` | `internal/forum/handler_test.go` |
| `CHK-03` | `EVID-01` | `internal/forum/hub_test.go` |

### Evidence

- `EVID-01` Юнит-тесты hub + интеграционные тесты StreamEvents — зелёные.
- `EVID-02` Интеграционный тест reconnect с `Last-Event-ID` — зелёный.

### Evidence contract

| Evidence ID | Artifact | Producer | Path contract | Reused by checks |
| --- | --- | --- | --- | --- |
| `EVID-01` | Вывод `go test ./internal/forum/...` и `go test -tags integration ./internal/forum/...` | CI / локальный запуск | `internal/forum/hub_test.go`, `internal/forum/handler_test.go` | `CHK-01`, `CHK-03` |
| `EVID-02` | Вывод `go test -tags integration ./internal/forum/...` | CI / локальный запуск | `internal/forum/handler_test.go` | `CHK-02` |
