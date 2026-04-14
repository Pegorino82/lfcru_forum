---
title: "FT-009: Admin — загрузка изображений для статей"
doc_kind: feature
doc_function: canonical
purpose: "Multipart-загрузка изображений, нормализация размера, конвертация в WebP, хранение на файловой системе. Depends on FT-007, ADR-005."
derived_from:
  - ../../domain/problem.md
  - ../../adr/ADR-005-image-storage.md
status: active
delivery_status: planned
audience: humans_and_agents
must_not_define:
  - implementation_sequence
---

# FT-009: Admin — загрузка изображений для статей

## What

### Problem

Статьи не могут содержать изображения. Без них контент беднее. Admin/Moderator нужна возможность прикреплять изображения к черновику перед публикацией.

### Outcome

| Metric ID | Metric | Baseline | Target | Measurement method |
| --- | --- | --- | --- | --- |
| `MET-01` | Загрузка изображения | Невозможна | `POST /admin/articles/:id/images` → 200, файл сохранён | Интеграционный тест |
| `MET-02` | Нормализация размера | Нет | Изображение ≤ max_width px, формат WebP | Unit-тест transform |

### Scope

- `REQ-01` `POST /admin/articles/:id/images` — принимает multipart/form-data с файлом изображения. Сохраняет нормализованный WebP на файловой системе.
- `REQ-02` Нормализация: resize до максимальной ширины (OQ-01, default 1200px) с сохранением пропорций; конвертация в WebP.
- `REQ-03` Запись о загруженном изображении в БД: таблица `article_images` (`id`, `article_id`, `filename`, `original_filename`, `created_at`).
- `REQ-04` `DELETE /admin/articles/:id/images/:image_id` — удаляет запись из БД и файл с диска.
- `REQ-05` Список изображений статьи в форме редактирования (`GET /admin/articles/:id/edit` включает существующие изображения).
- `REQ-06` Отображение изображений в публичной статье: CSS `aspect-ratio: 16/9; object-fit: cover; width: 100%`.

### Non-Scope

- `NS-01` CDN, object storage, cloud upload — ADR-005 принят, только файловая система.
- `NS-02` On-the-fly изменение размера через URL-параметры (как в imgproxy).
- `NS-03` srcset / responsive images — достаточно одного размера (CSS object-fit).
- `NS-04` Drag-and-drop загрузка — стандартный `<input type="file">`.
- `NS-05` Видео или другие медиа-форматы.
- `NS-06` Ограничение количества изображений на статью.

### Constraints / Assumptions

- `ASM-01` ADR-005 принят: `UPLOADS_DIR` env-переменная, Docker volume.
- `ASM-02` FT-007 реализован: `RequireAdminOrMod` middleware.
- `ASM-03` Таблица `article_images` создаётся новой goose-миграцией.
- `CON-01` Go на хосте не установлен — все команды через Docker.
- `CON-02` Максимальный размер загружаемого файла — OQ-01 (default предлагается 10 MB до нормализации).
- `CON-03` Поддерживаемые форматы на входе: JPEG, PNG, WebP. GIF и SVG — вне scope.

## How

### Solution

Добавить таблицу `article_images`. При загрузке: декодировать изображение, resize до max_width (Go `image` stdlib + `golang.org/x/image` или `github.com/disintegration/imaging`), encode в WebP (OQ-02 — нужна библиотека), сохранить в `$UPLOADS_DIR/{article_id}/{uuid}.webp`, записать в БД. nginx раздаёт `/uploads/*` как статику.

### Change Surface

| Surface | Type | Why it changes |
| --- | --- | --- |
| `migrations/` | data | Новая миграция: таблица `article_images` |
| `internal/admin/images_handler.go` | code | Upload, Delete endpoints |
| `internal/admin/image_service.go` | code | Resize + WebP encode + save logic |
| `internal/admin/images_repo.go` | code | CRUD для `article_images` |
| `templates/admin/articles/edit.html` | code | Секция загрузки изображений |
| `templates/news/article.html` | code | Отображение изображений в публичной статье |
| `internal/config/config.go` | code | `UploadsDir string` |
| `docker-compose.dev.yml` | config | Volume mount для UploadsDir |
| `docker-compose.prod.yml` | config | Volume mount для UploadsDir |

### Flow

1. Admin открывает форму редактирования статьи.
2. Выбирает файл → `POST /admin/articles/:id/images` (multipart).
3. Handler: validates type/size → decode → resize → encode WebP → save to disk → INSERT article_images.
4. Handler возвращает partial HTML с превью загруженного изображения (HTMX swap).
5. Изображения отображаются на странице статьи в порядке загрузки.

### Contracts

| Contract ID | Input / Output | Producer / Consumer | Notes |
| --- | --- | --- | --- |
| `CTR-01` | `POST /admin/articles/:id/images` multipart → 200 + partial HTML | Handler / HTMX | Возвращает `<img>` с путём к новому файлу |
| `CTR-02` | `DELETE /admin/articles/:id/images/:image_id` → 200 | Handler / HTMX | Удаляет файл + запись |
| `CTR-03` | `GET /uploads/{article_id}/{uuid}.webp` → файл | nginx / Browser | nginx раздаёт статику, Go не участвует |

### Failure Modes

- `FM-01` Файл превышает лимит размера → 400 с сообщением.
- `FM-02` Неподдерживаемый формат → 400.
- `FM-03` Ошибка записи на диск → 500, файл не сохранён, запись в БД не создаётся.
- `FM-04` Файл удалён с диска вручную, но запись в БД есть → broken img в публичной статье. Mitigation: обработать 404 от nginx gracefully в шаблоне.

### ADR Dependencies

| ADR | Current `decision_status` | Used for | Execution rule |
| --- | --- | --- | --- |
| [ADR-005](../../adr/ADR-005-image-storage.md) | `accepted` | Хранение файлов, UPLOADS_DIR, Docker volume | Принято — можно реализовывать |

## Verify

### Exit Criteria

- `EC-01` Загрузка JPEG → сохранён `.webp` файл в `$UPLOADS_DIR/{article_id}/`, запись в `article_images`.
- `EC-02` Загруженное изображение ≤ max_width по ширине, пропорции сохранены.
- `EC-03` Удаление изображения → файл удалён с диска, запись удалена из БД.
- `EC-04` Попытка загрузить файл >10MB → 400.
- `EC-05` Попытка загрузить GIF → 400.
- `EC-06` Изображения отображаются в публичной статье с CSS `aspect-ratio: 16/9`.
- `EC-07` Автоматические тесты зелёные.

### Traceability matrix

| Requirement ID | Design refs | Acceptance refs | Checks | Evidence IDs |
| --- | --- | --- | --- | --- |
| `REQ-01` | `ASM-01`, `ASM-02`, `CTR-01`, `FM-01`, `FM-02`, `FM-03` | `EC-01`, `SC-01` | `CHK-01` | `EVID-01` |
| `REQ-02` | `CON-02`, `CON-03` | `EC-02`, `SC-02` | `CHK-01`, `CHK-02` | `EVID-01`, `EVID-02` |
| `REQ-03` | `ASM-03`, `CTR-01` | `EC-01` | `CHK-01` | `EVID-01` |
| `REQ-04` | `CTR-02`, `FM-04` | `EC-03`, `SC-03` | `CHK-01` | `EVID-01` |
| `REQ-05` | `CTR-01` | `SC-04` | `CHK-01` | `EVID-01` |
| `REQ-06` | `CON-03` | `EC-06`, `SC-05` | `CHK-01` | `EVID-01` |

### Acceptance Scenarios

- `SC-01` Admin загружает JPEG через форму редактирования → изображение появляется в списке и в превью.
- `SC-02` Загруженное изображение имеет ширину ≤ max_width, формат WebP.
- `SC-03` Admin удаляет изображение → оно пропадает из формы.
- `SC-04` Форма редактирования статьи показывает все ранее загруженные изображения.
- `SC-05` Публичная страница статьи отображает изображения с правильными пропорциями на мобильном и десктопе.

### Checks

| Check ID | Covers | How to check | Expected result | Evidence path |
| --- | --- | --- | --- | --- |
| `CHK-01` | `EC-01`..`EC-07`, `SC-01`..`SC-05` | `docker compose -f docker-compose.dev.yml run --rm app go test -tags integration -p 1 ./internal/admin/...` | Все тесты зелёные | stdout теста |
| `CHK-02` | `EC-02` (resize unit) | `docker compose -f docker-compose.dev.yml run --rm app go test ./internal/admin/... -run TestImageResize` | Тест нормализации пройден | stdout теста |

### Test matrix

| Check ID | Evidence IDs | Evidence path |
| --- | --- | --- |
| `CHK-01` | `EVID-01` | stdout `go test` |
| `CHK-02` | `EVID-02` | stdout `go test -run TestImageResize` |

### Evidence

- `EVID-01` Вывод `go test` с `ok internal/admin` и без FAIL.
- `EVID-02` Вывод unit-теста resize с `PASS`.

### Evidence contract

| Evidence ID | Artifact | Producer | Path contract | Reused by checks |
| --- | --- | --- | --- | --- |
| `EVID-01` | stdout go test | docker test run | stdout | `CHK-01` |
| `EVID-02` | stdout go test (unit) | docker test run | stdout | `CHK-02` |

## Open Questions

- `OQ-01` Максимальная ширина нормализации: 1200px? 800px? Зависит от ширины контент-колонки в дизайне.
- `OQ-02` Go-библиотека для WebP encode: `golang.org/x/image/webp` только декодирует; encode требует `github.com/chai2010/webp` или CGO (`libwebp`). Выбор влияет на Dockerfile. Альтернатива — хранить JPEG/PNG без конвертации.
