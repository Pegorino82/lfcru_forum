---
title: "FT-008: Admin — управление статьями"
doc_kind: feature
doc_function: canonical
purpose: "CRUD статей в admin-панели, превью, workflow черновик → ревью → публикация. Depends on FT-007, ADR-006, UC-001."
derived_from:
  - ../../domain/problem.md
  - ../../use-cases/UC-001-article-publishing.md
  - ../../adr/ADR-006-article-status-machine.md
status: active
delivery_status: planned
audience: humans_and_agents
must_not_define:
  - implementation_sequence
---

# FT-008: Admin — управление статьями

## What

### Problem

Admin/Moderator не может создавать, редактировать и публиковать статьи через UI. Нет workflow для согласования черновиков между участниками команды перед публикацией.

### Outcome

| Metric ID | Metric | Baseline | Target | Measurement method |
| --- | --- | --- | --- | --- |
| `MET-01` | Возможность создать статью | Нет UI | Admin/Mod создаёт статью через форму | Интеграционный тест |
| `MET-02` | Статусный workflow | Только is_published | draft/in_review/published работают | Интеграционный тест |
| `MET-03` | Превью до публикации | Нет | GET /admin/articles/:id/preview → 200, рендер как публичная страница | Интеграционный тест |

### Scope

- `REQ-01` Список черновиков и статей с фильтром по статусу (`GET /admin/articles`).
- `REQ-02` Форма создания новой статьи: заголовок, текст (plain textarea или Markdown — OQ-01). Сохраняется как `draft`.
- `REQ-03` Форма редактирования (`GET /admin/articles/:id/edit`, `POST /admin/articles/:id`).
- `REQ-04` Превью статьи до публикации (`GET /admin/articles/:id/preview`) — рендер через публичный шаблон.
- `REQ-05` Смена статуса: черновик → опубликовать, черновик → отправить на ревью, in_review → опубликовать, опубликовано → снять с публикации.
- `REQ-06` Схема БД: добавить `status news_status`, `reviewer_id UUID REFERENCES users(id)`. Убрать `is_published` (миграция данных). Зависит от ADR-006.
- `REQ-07` Обратная совместимость: публичный маршрут `/news` и `/news/:id` (FT-006) корректно работают после миграции (`WHERE status = 'published'`).

### Non-Scope

- `NS-01` WYSIWYG-редактор — OQ-01; пока plain textarea.
- `NS-02` Удаление статей через UI — добавить позже.
- `NS-03` История изменений статьи.
- `NS-04` Загрузка изображений — в FT-009.
- `NS-05` Комментарии к статьям.

### Constraints / Assumptions

- `ASM-01` FT-007 реализован: middleware `RequireAdminOrMod` и группа `/admin` существуют.
- `ASM-02` ADR-006 принят (`decision_status: accepted`) до начала реализации.
- `ASM-03` Таблица `news` существует (migration 004); `is_published` → `status` через новую миграцию.
- `CON-01` Go на хосте не установлен — все команды через Docker.
- `CON-02` Миграция схемы — необратима; требует AG-* approval.
- `DEC-01` ADR-006: `status` как PostgreSQL enum `news_status`. До перевода ADR-006 в `accepted` — блокер для реализации.

## How

### Solution

Добавить goose-миграцию: enum `news_status`, новая колонка `status`, data migration, удаление `is_published`. Добавить в `internal/news/` методы repo и service для admin-операций. Зарегистрировать admin-handlers в группе `/admin/articles`.

### Change Surface

| Surface | Type | Why it changes |
| --- | --- | --- |
| `migrations/` | data | Новая goose-миграция: enum + status колонка + data migration |
| `internal/news/repo.go` | code | Методы: CreateDraft, UpdateArticle, ChangeStatus, ListByStatus, GetByIDAdmin |
| `internal/news/service.go` | code | Бизнес-логика смены статусов, валидация reviewer |
| `internal/admin/articles_handler.go` | code | HTTP handlers: список, форма, сохранение, превью, смена статуса |
| `templates/admin/articles/` | code | list.html, edit.html, preview wrapper |
| `internal/news/repo_test.go` | code | Тесты новых методов |
| `internal/admin/articles_handler_test.go` | code | Интеграционные тесты |

### Flow

1. Admin открывает `GET /admin/articles` — видит список всех статей с фильтром по статусу.
2. Нажимает «Новая статья» → `GET /admin/articles/new` → форма.
3. Заполняет заголовок + текст → `POST /admin/articles` → сохраняется как `draft`.
4. Открывает превью `GET /admin/articles/:id/preview` — видит статью как читатель.
5. Нажимает «Отправить на ревью» → `POST /admin/articles/:id/status` body: `status=in_review, reviewer_id=XYZ`.
6. Reviewer видит статью в своём списке, открывает, нажимает «Опубликовать» → `status=published`, устанавливается `published_at`.

### Contracts

| Contract ID | Input / Output | Producer / Consumer | Notes |
| --- | --- | --- | --- |
| `CTR-01` | `POST /admin/articles` → redirect `/admin/articles/:id/edit` | Handler / Browser | Создаёт статью со статусом `draft` |
| `CTR-02` | `POST /admin/articles/:id/status` body: `status=published` → 200/redirect | Handler / Browser | Устанавливает `published_at = now()` |
| `CTR-03` | `GET /news?page=N` → только `status = 'published'` | Repo / Handler | Обратная совместимость после миграции |

### Failure Modes

- `FM-01` Попытка перевести статью в `published` без заголовка или текста → 400, форма с ошибкой.
- `FM-02` DB error при смене статуса → 500, slog.Error.
- `FM-03` Попытка перейти в недопустимый статус (e.g., `published → in_review`) → 400.

### ADR Dependencies

| ADR | Current `decision_status` | Used for | Execution rule |
| --- | --- | --- | --- |
| [ADR-006](../../adr/ADR-006-article-status-machine.md) | `proposed` | Схема БД: enum `news_status`, колонка `status` | Не реализовывать до перевода ADR-006 в `accepted` |

## Verify

### Exit Criteria

- `EC-01` Создание черновика через форму → статус `draft`, статья не видна в `/news`.
- `EC-02` Превью черновика → рендер идентичен публичной странице статьи.
- `EC-03` Публикация → статус `published`, `published_at` заполнена, статья видна в `/news`.
- `EC-04` FT-006 (`GET /news`) продолжает работать после миграции схемы.
- `EC-05` Смена статуса в недопустимом направлении → 400.
- `EC-06` Автоматические тесты зелёные.

### Traceability matrix

| Requirement ID | Design refs | Acceptance refs | Checks | Evidence IDs |
| --- | --- | --- | --- | --- |
| `REQ-01` | `ASM-01`, `CTR-01` | `EC-01`, `SC-01` | `CHK-01` | `EVID-01` |
| `REQ-02` | `DEC-01`, `CON-02` | `EC-01`, `SC-02` | `CHK-01` | `EVID-01` |
| `REQ-03` | `CTR-01` | `SC-03` | `CHK-01` | `EVID-01` |
| `REQ-04` | `ASM-01` | `EC-02`, `SC-04` | `CHK-01` | `EVID-01` |
| `REQ-05` | `CTR-02`, `FM-03` | `EC-03`, `EC-05`, `SC-05`, `SC-06` | `CHK-01` | `EVID-01` |
| `REQ-06` | `DEC-01`, `CON-02` | `EC-01`, `EC-03` | `CHK-01`, `CHK-02` | `EVID-01`, `EVID-02` |
| `REQ-07` | `CTR-03`, `ASM-03` | `EC-04`, `SC-07` | `CHK-01` | `EVID-01` |

### Acceptance Scenarios

- `SC-01` Admin открывает `/admin/articles` — видит список статей с колонкой статуса.
- `SC-02` Admin создаёт статью через форму → статья появляется в списке со статусом `draft`.
- `SC-03` Admin редактирует черновик → изменения сохранены.
- `SC-04` Admin открывает превью черновика → видит статью в публичном оформлении.
- `SC-05` Admin нажимает «Отправить на ревью» → статус меняется на `in_review`.
- `SC-06` Reviewer нажимает «Опубликовать» → статус `published`, статья видна в `/news`.
- `SC-07` После миграции `GET /news` возвращает 200 с теми же статьями, что были `is_published=true` до миграции.

### Checks

| Check ID | Covers | How to check | Expected result | Evidence path |
| --- | --- | --- | --- | --- |
| `CHK-01` | `EC-01`..`EC-06`, `SC-01`..`SC-07` | `docker compose -f docker-compose.dev.yml run --rm app go test -tags integration -p 1 ./internal/news/... ./internal/admin/...` | Все тесты зелёные | stdout теста |
| `CHK-02` | `EC-04` (миграция) | `docker compose -f docker-compose.dev.yml run --rm app go test -tags integration -p 1 ./internal/news/...` после миграции | `ok internal/news` без FAIL | stdout теста |

### Test matrix

| Check ID | Evidence IDs | Evidence path |
| --- | --- | --- |
| `CHK-01` | `EVID-01` | stdout `go test` |
| `CHK-02` | `EVID-02` | stdout `go test` (post-migration) |

### Evidence

- `EVID-01` Вывод `go test` с `ok internal/news`, `ok internal/admin` и без FAIL.
- `EVID-02` Вывод `go test internal/news` после выполнения миграции — без FAIL.

### Evidence contract

| Evidence ID | Artifact | Producer | Path contract | Reused by checks |
| --- | --- | --- | --- | --- |
| `EVID-01` | stdout go test | docker test run | stdout | `CHK-01` |
| `EVID-02` | stdout go test (post-migration) | docker test run | stdout | `CHK-02` |

## Open Questions

- `OQ-01` Формат текста статьи: plain textarea (отображать как `<pre>` или параграфы) или Markdown? Влияет на выбор рендерера.
- `OQ-02` Нужна ли возможность Admin публиковать сразу из формы создания (без отдельной кнопки «Опубликовать»)?
