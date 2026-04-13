---
title: "FT-006: News — список статей"
doc_kind: feature
doc_function: canonical
purpose: "Страница со списком опубликованных новостей с offset-пагинацией и сортировкой по дате убыв."
derived_from:
  - ../../domain/problem.md
status: active
delivery_status: done
audience: humans_and_agents
must_not_define:
  - implementation_sequence
---

# FT-006: News — список статей

## What

### Problem

Болельщики не могут просматривать список новостей на сайте: существует только маршрут `/news/:id` (страница статьи), но нет страницы-каталога. Чтобы попасть на нужную статью, нужно знать её ID или перейти с главной страницы (где показываются только последние 3 новости).

### Outcome

| Metric ID | Metric | Baseline | Target | Measurement method |
| --- | --- | --- | --- | --- |
| `MET-01` | Доступность списка новостей | Отсутствует маршрут `/news` | `GET /news` → 200, список опубликованных статей | Интеграционный тест handler |
| `MET-02` | Пагинация работает | — | При `?page=N` отображается корректная страница | Интеграционный тест handler |

### Scope

- `REQ-01` `GET /news` — страница со списком опубликованных новостей, отсортированных по `published_at DESC`.
- `REQ-02` Offset-пагинация с размером страницы 20 (хардкод). Параметр `?page=N` (1-based, default=1).
- `REQ-03` Каждая новость в списке показывает: заголовок (ссылка на `/news/:id`), дату публикации.
- `REQ-04` HTMX-совместимость: при `HX-Request: true` — partial (`content`-блок), иначе — полный документ.

### Non-Scope

- `NS-01` Создание, редактирование, удаление новостей — вне scope.
- `NS-02` Полнотекстовый поиск по новостям — вне scope.
- `NS-03` Фильтрация по тегам, рубрикам — вне scope.
- `NS-04` Cursor-based пагинация — вне scope (достаточно offset для текущего масштаба).
- `NS-05` Настройка page size через конфиг или UI — вне scope.

### Constraints / Assumptions

- `ASM-01` Таблица `news` уже существует (migration 004). `GetPublishedByID` и `LatestPublished` в `Repo` уже реализованы.
- `ASM-02` Роут `/news/:id` существует и не изменяется этой фичей.
- `CON-01` Go на хосте не установлен — все команды выполняются через Docker.
- `CON-02` Шаблоны используют `html/template`, inline CSS допускается (как в существующих шаблонах).
- `DEC-01` Тип пагинации: **offset-based** выбран как достаточный для текущего масштаба (фан-сайт, не миллионы записей). Cursor-based откладывается на будущее.

## How

### Solution

Добавить в `internal/news/repo.go` метод `ListPublished(ctx, limit, offset int) ([]News, int64, error)`, возвращающий страницу новостей и общее количество. Добавить в `Handler` метод `ShowList`, зарегистрировать роут `GET /news`, создать шаблон `templates/news/list.html` по образцу существующих шаблонов.

Пагинация: 20 записей на страницу, `OFFSET = (page-1)*20`, `LIMIT 20`. Общее количество нужно для отображения номеров страниц.

### Change Surface

| Surface | Type | Why it changes |
| --- | --- | --- |
| `internal/news/repo.go` | code | Добавить `ListPublished` с `LIMIT/OFFSET` и `COUNT(*)` |
| `internal/news/handler.go` | code | Добавить `ShowList`, `ListData` struct, регистрацию роута `GET /news` |
| `templates/news/list.html` | code | Новый шаблон списка новостей |
| `cmd/forum/main.go` | code | Регистрация роута `GET /news` через `handler.RegisterRoutes` |
| `internal/news/repo_test.go` | code | Тесты для `ListPublished` |
| `internal/news/handler_test.go` | code | Интеграционные тесты для `ShowList` |

### Flow

1. Пользователь открывает `GET /news?page=2`.
2. Handler парсит `page` (default 1, min 1).
3. Repo: `SELECT id, title, published_at FROM news WHERE is_published=true ORDER BY published_at DESC LIMIT 20 OFFSET 20`.
4. Repo: `SELECT COUNT(*) FROM news WHERE is_published=true` (один запрос в транзакции или отдельно).
5. Handler вычисляет `totalPages = ceil(total/20)`, собирает `ListData`.
6. Рендер: полный документ или partial при `HX-Request`.

### Contracts

| Contract ID | Input / Output | Producer / Consumer | Notes |
| --- | --- | --- | --- |
| `CTR-01` | `GET /news?page=N` → HTML | Handler / Browser | page ≤ 1 → redirect или страница 1; page > totalPages → пустой список или последняя страница |
| `CTR-02` | `ListPublished(ctx, limit, offset)` → `([]News, int64, error)` | Repo / Handler | int64 = общее количество опубликованных; пустой слайс при отсутствии данных |

### Failure Modes

- `FM-01` Невалидный `?page` (не число, 0, отрицательный) → treat as page=1.
- `FM-02` `?page` > totalPages → показать пустой список с сообщением (не 404).
- `FM-03` DB error при загрузке → 500 с user-friendly сообщением, `slog.Error`.

## Verify

### Exit Criteria

- `EC-01` `GET /news` возвращает 200 со списком опубликованных новостей, отсортированных по дате убыв.
- `EC-02` `GET /news?page=2` возвращает вторую страницу (если есть записи).
- `EC-03` Черновики не появляются в списке.
- `EC-04` Пагинация не отображается при кол-ве записей ≤ 20.
- `EC-05` Автоматические тесты зелёные.

### Traceability matrix

| Requirement ID | Design refs | Acceptance refs | Checks | Evidence IDs |
| --- | --- | --- | --- | --- |
| `REQ-01` | `ASM-01`, `CTR-01`, `FM-03` | `EC-01`, `SC-01` | `CHK-01` | `EVID-01` |
| `REQ-02` | `DEC-01`, `CON-01`, `CTR-02`, `FM-01`, `FM-02` | `EC-02`, `EC-04`, `SC-02`, `SC-03` | `CHK-01` | `EVID-01` |
| `REQ-03` | `CTR-01` | `EC-01`, `SC-01` | `CHK-01` | `EVID-01` |
| `REQ-04` | `ASM-02`, `CON-02` | `SC-04` | `CHK-01` | `EVID-01` |

### Acceptance Scenarios

- `SC-01` Открыть `GET /news` — отображается список опубликованных новостей: заголовок-ссылка, дата; новые сверху.
- `SC-02` При >20 опубликованных новостях `?page=2` показывает вторую страницу, `?page=1` — первую.
- `SC-03` `?page=0`, `?page=-1`, `?page=abc` → страница 1 (нет 4xx).
- `SC-04` HTMX-запрос (`HX-Request: true`) → partial HTML без `<html>/<body>`.
- `SC-05` Черновик (`is_published=false`) не виден в списке.

### Checks

| Check ID | Covers | How to check | Expected result | Evidence path |
| --- | --- | --- | --- | --- |
| `CHK-01` | `EC-01`..`EC-05`, `SC-01`..`SC-05` | `docker compose -f docker-compose.dev.yml run --rm app go test -tags integration -p 1 ./internal/news/...` | Все тесты зелёные | stdout теста |

### Test matrix

| Check ID | Evidence IDs | Evidence path |
| --- | --- | --- |
| `CHK-01` | `EVID-01` | stdout `go test` |

### Evidence

- `EVID-01` Вывод `go test` с `ok internal/news` и без FAIL.

### Evidence contract

| Evidence ID | Artifact | Producer | Path contract | Reused by checks |
| --- | --- | --- | --- | --- |
| `EVID-01` | stdout go test | docker test run | stdout | `CHK-01` |
