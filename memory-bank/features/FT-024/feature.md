---
title: "FT-024: Профиль пользователя"
doc_kind: feature
doc_function: canonical
purpose: "Профиль пользователя: quick-view модалка по клику на аватар/имя, полная страница профиля /profile/{username}, загрузка аватара, fallback-инициалы."
derived_from:
  - ../../domain/problem.md
  - ../../adr/ADR-005-image-storage.md
status: active
delivery_status: planned
audience: humans_and_agents
must_not_define:
  - implementation_sequence
---

# FT-024: Профиль пользователя

## What

### Problem

Пользователи не имеют публичной идентичности на сайте: нет аватаров, нет страницы профиля, нельзя узнать кто есть кто. Это снижает вовлечённость и ощущение сообщества.

### Outcome

| Metric ID | Metric | Baseline | Target | Measurement method |
| --- | --- | --- | --- | --- |
| `MET-01` | Профиль открываем по клику на имя/аватар | 0% (функция отсутствует) | 100% поверхностей покрыты кликабельным именем | Playwright E2E |
| `MET-02` | Аватары загружены хотя бы у 1 пользователя после релиза | 0 | ≥1 | `SELECT COUNT(*) FROM users WHERE avatar_url IS NOT NULL` |

### Scope

- `REQ-01` Quick-view модалка: открывается по клику на аватар или имя пользователя в любом месте сайта (форум, комментарии, header). Содержит: аватар, имя, дату регистрации, кол-во постов + последний пост (ссылка + превью ~40 символов + relative time), кол-во комментариев + последний комментарий (ссылка + превью + relative time). Закрывается: крестик, клик вне модалки, ESC.
- `REQ-02` Полная страница профиля по маршруту `/profile/{username}`: содержит те же данные, что и модалка. Публичная (доступна гостям).
- `REQ-03` Аватар пользователя: отображается в header (рядом с именем), в постах форума (слева от имени), в комментариях к новостям, в модалке и на странице профиля.
- `REQ-04` Загрузка аватара: доступна только авторизованному владельцу профиля. Поддерживаемые форматы: JPEG, PNG, WebP. Макс. размер: 5 МБ. После выбора файла — превью. Сохранение — по явному подтверждению.
- `REQ-05` Fallback аватара при отсутствии загруженного: инициалы пользователя (первая буква имени) на цветном фоне. Цвет детерминирован по хешу username (не меняется от рендера к рендеру).
- `REQ-06` Header: имя пользователя заменяется на `[аватар-кружок] + имя`, кликабельно → открывает модалку своего профиля.
- `REQ-07` Пустые состояния: «Пользователь ещё не писал на форуме» / «Пользователь ещё не оставлял комментарии» при отсутствии активности.

### Non-Scope

- `NS-01` Редактирование имени пользователя — не в этой фиче.
- `NS-02` Список всех постов/комментариев пользователя с пагинацией на странице профиля — только последний пост и последний комментарий.
- `NS-03` Смена пароля, email, других данных профиля — не в этой фиче.
- `NS-04` Удаление аватара пользователем — не в этой фиче. После загрузки аватара вернуться к fallback-инициалам через UI нельзя.
- `NS-05` Аватары в admin-панели (список пользователей) — не в этой фиче.
- `NS-06` CDN, S3 или иное облачное хранилище — файловая система по ADR-005.

### Constraints / Assumptions

- `CON-01` CSRF-токен обязателен для `POST /profile/avatar` (PCON-02).
- `CON-02` Только владелец профиля может изменить аватар; проверка — по `session.UserID == profile.UserID`.
- `CON-03` Аватары хранятся на файловой системе по ADR-005: `$UPLOADS_DIR/avatars/{user_id}.webp`. При повторной загрузке файл перезаписывается. Явного delete-эндпоинта нет (NS-04) — перезапись является единственным способом сменить аватар.
- `CON-04` Относительное время (relative time) вычисляется на сервере при рендере шаблона (Go-функция в `FuncMap`), не на клиенте.
- `ASM-01` Username уникален (case-insensitive UNIQUE INDEX существует в миграции 001).
- `ASM-02` Модалка реализована через Alpine.js (`x-data`, `x-show`) + HTMX для загрузки данных профиля по клику.
- `ASM-03` Цвет fallback-аватара детерминирован: вычисляется на сервере хешированием username → индекс в фиксированной палитре. Алгоритм не меняется после релиза.
- `ASM-04` Существующая session middleware применяется к маршрутам `/profile/*`, требующим авторизации (`POST /profile/avatar`), и перенаправляет неавторизованных → `/login`. В handler достаточно ownership-проверки (CON-02).

## How

### Solution

Добавить поле `avatar_url` в таблицу `users` (миграция), расширить `user.User` моделью, реализовать Service-слой профиля (`internal/profile/service.go`) с domain-логикой (валидация формата/размера, конвертация в WebP, ownership-проверка, запись FS, обновление `avatar_url`), Handler-слой (`internal/profile/handler.go`) для парсинга запросов и вызова Service, добавить HTMX-эндпоинт `GET /profile/{username}/modal` для quick-view, обновить шаблоны форума и комментариев (кликабельные имена), обновить header. Fallback-аватар рендерится на сервере через Go FuncMap (SVG + детерминированный цвет) — это позволяет получить одинаковый вывод в полных страницах и HTMX-фрагментах без зависимости от JS; CSS-only вариант отклонён из-за отсутствия нативного детерминированного цвета без скриптов.

### Change Surface

| Surface | Type | Why it changes |
| --- | --- | --- |
| `migrations/013_add_avatar_url_to_users.sql` | data | Добавить колонку `avatar_url TEXT` в `users` |
| `internal/user/model.go` | code | Добавить `AvatarURL *string` в `User` |
| `internal/user/repo.go` | code | Добавить `GetByUsername`, обновить `Scan` для `avatar_url` |
| `internal/user/repo.go` | code | Добавить `GetByUsername`; обновить `Scan` для `avatar_url` |
| `internal/profile/service.go` (new) | code | Domain-логика: валидация формата/размера, конвертация в WebP, ownership-проверка (CON-02), запись FS, обновление `users.avatar_url`; агрегация профильных данных через ForumRepo и CommentRepo интерфейсы |
| `internal/tmpl/renderer.go` | code | Добавить FuncMap-функции: `avatarColor`/`avatarInitials` (REQ-05, ASM-03) и `relativeTime` (CON-04) |
| `internal/profile/handler.go` (new) | code | Парсит запрос, вызывает Service; маршруты `GET /profile/{username}`, `GET /profile/{username}/modal`, `POST /profile/{username}/avatar` |
| `internal/config/config.go` | code | `UploadsDir` уже есть (ADR-005, FT-009) — проверить |
| `templates/profile/page.html` (new) | code | Полная страница профиля |
| `templates/profile/modal.html` (new) | code | HTMX-фрагмент модалки |
| `templates/layouts/base.html` | code | Header: аватар + имя → кликабельно |
| `templates/forum/topic.html` | code | Аватар + кликабельное имя в постах |
| `templates/forum/index.html` / `templates/forum/section.html` | code | Кликабельное имя автора последнего поста (если отображается) |
| `templates/news/article.html` | code | Аватар + кликабельное имя в комментариях |
| `cmd/forum/main.go` | code | Регистрация profile routes |
| inline стили и скрипты в шаблонах | code | Стили модалки, аватар-кружка; inline `<script>`: FileReader-превью аватара до submit (отдельного `static/css/` нет) |

### Flow

**Quick view:**
1. Пользователь кликает на имя/аватар → HTMX `hx-get="/profile/{username}/modal"` → сервер рендерит `modal.html` фрагмент с данными.
2. Alpine.js показывает overlay-модалку (`x-show`). Фон затемняется. Скролл страницы заблокирован.
3. Модалка содержит: аватар, имя, дату регистрации, статистику постов/комментариев. Кнопка «Открыть профиль» → `<a href="/profile/{username}">`.
4. Закрытие: крестик / клик вне / ESC → `x-show = false`.

**Полная страница профиля:**
1. `GET /profile/{username}` → хэндлер получает пользователя по username, считает посты и комментарии, берёт последний пост и комментарий.
2. Рендерит `templates/profile/page.html`.

**Загрузка аватара:**
1. Владелец на своей странице профиля/модалке видит кнопку «Изменить аватар».
2. `<input type="file">` → JS-превью через `FileReader`.
3. Пользователь подтверждает → `POST /profile/{username}/avatar` (multipart, с CSRF).
4. Сервер: валидация формата и размера, конвертация в WebP, запись в `$UPLOADS_DIR/avatars/{user_id}.webp`, обновление `users.avatar_url`.
5. Ответ: обновлённый фрагмент аватара (HTMX OOB swap) или редирект.

### Contracts

| Contract ID | Input / Output | Producer / Consumer | Notes |
| --- | --- | --- | --- |
| `CTR-01` | `GET /profile/{username}` → HTML страница | handler / browser | Публичный; 404 если пользователь не найден |
| `CTR-02` | `GET /profile/{username}/modal` → HTML фрагмент | handler / HTMX | Публичный; рендерит modal.html partial |
| `CTR-03` | `POST /profile/{username}/avatar` multipart `avatar` field | browser / handler | Требует авторизации + CSRF; handler проверяет `session.UserID == target_user.ID` (CON-02) → 403 если чужой; возвращает обновлённый аватар-фрагмент |
| `CTR-04` | `users.avatar_url TEXT NULL` | migration / repo | NULL = нет аватара; значение = относительный путь `/uploads/avatars/{user_id}.webp` |

### Failure Modes

- `FM-01` Загружен файл неподдерживаемого формата → HTTP 422, сообщение пользователю в UI.
- `FM-02` Файл превышает 5 МБ → HTTP 413, сообщение пользователю в UI.
- `FM-03` Ошибка записи на файловую систему → HTTP 500, сообщение «Не удалось сохранить аватар».
- `FM-04` Пользователь с таким username не найден → HTTP 404 на `/profile/{username}`.
- `FM-05` Попытка изменить аватар чужого профиля → HTTP 403.
- `FM-06` Модалка открыта, но HTMX-запрос `GET /profile/{username}/modal` вернул 5xx → показать fallback-сообщение об ошибке внутри модалки (не закрывать).

### ADR Dependencies

| ADR | Current `decision_status` | Used for | Execution rule |
| --- | --- | --- | --- |
| [../../adr/ADR-005-image-storage.md](../../adr/ADR-005-image-storage.md) | `accepted` | Хранение аватаров: filesystem + Docker volume, путь `$UPLOADS_DIR/avatars/{user_id}.webp` | Canonical input — следовать без альтернатив |

## Verify

### Exit Criteria

- `EC-01` Клик на имя/аватар пользователя открывает quick-view модалку с корректными данными.
- `EC-02` Страница `/profile/{username}` доступна гостям и содержит все указанные данные.
- `EC-03` Владелец профиля может загрузить аватар; аватар отображается везде (header, посты, модалка, профиль).
- `EC-04` Fallback-аватар (инициалы) отображается корректно при отсутствии загруженного; цвет стабилен.
- `EC-05` Закрытие модалки работает всеми тремя способами (крестик, клик вне, ESC).
- `EC-06` Ошибки загрузки аватара (формат, размер) корректно отображаются пользователю.
- `EC-07` Попытка изменить аватар чужого профиля возвращает 403.

### Traceability matrix

| Requirement ID | Design refs | Acceptance refs | Checks | Evidence IDs |
| --- | --- | --- | --- | --- |
| `REQ-01` | `ASM-02`, `CTR-02`, `FM-06` | `EC-01`, `EC-05`, `SC-01`, `SC-05`, `NEG-06` | `CHK-01` | `EVID-01` |
| `REQ-02` | `CTR-01`, `FM-04` | `EC-02`, `SC-02`, `NEG-04` | `CHK-01` | `EVID-01` |
| `REQ-03` | `CTR-04` | `EC-03`, `SC-03` | `CHK-01` | `EVID-01` |
| `REQ-04` | `CON-01`, `CON-02`, `CON-03`, `CTR-03`, `FM-01`, `FM-02`, `FM-03` | `EC-03`, `EC-06`, `EC-07`, `SC-03`, `SC-04`, `NEG-01`, `NEG-02`, `NEG-03` | `CHK-01`, `CHK-02` | `EVID-01`, `EVID-02` |
| `REQ-05` | `ASM-03` | `EC-04`, `SC-06` | `CHK-01` | `EVID-01` |
| `REQ-06` | `ASM-02`, `CTR-02` | `EC-01`, `SC-01` | `CHK-01` | `EVID-01` |
| `REQ-07` | — | `SC-07` | `CHK-01` | `EVID-01` |

### Acceptance Scenarios

- `SC-01` Авторизованный пользователь кликает на своё имя/аватар в header → открывается модалка со своими данными (поведение идентично клику везде на сайте) → кнопка «Открыть профиль» ведёт на `/profile/{username}`.
- `SC-02` Гость открывает `/profile/someuser` напрямую → страница рендерится без кнопки «Изменить аватар».
- `SC-03` Авторизованный пользователь на своём профиле загружает PNG 2 МБ → превью показано → подтверждает → аватар обновлён, отображается в header и на странице профиля.
- `SC-04` Пользователь пытается загрузить GIF → получает сообщение об ошибке формата.
- `SC-05` Модалка закрывается по ESC, по крестику и по клику вне.
- `SC-06` У пользователя нет аватара → отображаются инициалы на цветном фоне; при рефреше страницы цвет тот же.
- `SC-07` Нет постов/комментариев → отображаются пустые состояния с соответствующим текстом.

### Negative / Edge Cases

- `NEG-01` Загрузка файла > 5 МБ → ответ 413, UI показывает ошибку.
- `NEG-02` Загрузка файла неподдерживаемого типа (например, PDF) → ответ 422, UI показывает ошибку.
- `NEG-03` `POST /profile/avatar` без авторизации → редирект на `/login`.
- `NEG-04` `GET /profile/nonexistent` → 404 страница.
- `NEG-05` `POST /profile/{other_username}/avatar` авторизованным пользователем → 403.
- `NEG-06` HTMX-запрос `GET /profile/{username}/modal` возвращает 5xx → модалка показывает fallback-сообщение об ошибке.

### Checks

| Check ID | Covers | How to check | Expected result | Evidence path |
| --- | --- | --- | --- | --- |
| `CHK-01` | `EC-01`..`EC-05`, `EC-07`, все SC-*, `NEG-04`, `NEG-05`, `NEG-06` | `rtk go test ./...` + Playwright E2E | Все тесты зелёные | CI run / локальный вывод |
| `CHK-02` | `EC-06`, `NEG-01`, `NEG-02` | Playwright: загрузка невалидных файлов | Ошибки отображаются в UI | CI run / скриншоты |

### Test matrix

| Check ID | Evidence IDs | Evidence path |
| --- | --- | --- |
| `CHK-01` | `EVID-01` | CI: Go Tests + E2E jobs |
| `CHK-02` | `EVID-02` | CI: E2E job / Playwright screenshots |

### Evidence

- `EVID-01` Зелёный CI run: Go Tests + E2E (Playwright) — все сценарии SC-01..SC-07.
- `EVID-02` Playwright-тесты для NEG-01, NEG-02: файл превышает лимит и неверный формат → ошибка в UI.
- `EVID-03` Eval Draft→DR (feature.md review) — accept, итерация 3 (self-check на trivial path fix после 2 evaluator итераций). 2026-05-03. evaluator agent
- `EVID-04` Brief loop — accept, итерация 2. 2026-05-03. improve-loop.sh / evaluator agent
- `EVID-05` Spec loop — accept, итерация 3 (self-check после 2 evaluator итераций). 2026-05-03. improve-loop.sh / evaluator agent

### Evidence contract

| Evidence ID | Artifact | Producer | Path contract | Reused by checks |
| --- | --- | --- | --- | --- |
| `EVID-01` | CI run URL (Go Tests + E2E green) | CI / Playwright | GitHub Actions run на PR | `CHK-01` |
| `EVID-02` | Playwright test results для upload-ошибок | Playwright | GitHub Actions E2E job | `CHK-02` |
