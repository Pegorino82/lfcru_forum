---
title: "FT-024: Implementation Plan"
doc_kind: feature
doc_function: derived
purpose: "Execution-план реализации FT-024. Фиксирует discovery context, шаги, риски и test strategy без переопределения canonical feature-фактов."
derived_from:
  - feature.md
status: draft
audience: humans_and_agents
must_not_define:
  - ft_024_scope
  - ft_024_architecture
  - ft_024_acceptance_criteria
  - ft_024_blocker_state
---

# План имплементации FT-024: Профиль пользователя

## Цель текущего плана

Реализовать профиль пользователя: миграция, модель, repo, FuncMap-функции, profile service/handler, шаблоны, обновление существующих шаблонов (header, форум, комментарии). Результат: все SC-01..SC-07 и NEG-01..NEG-06 из `feature.md` проходят в CI.

## Discovery Context / Reference Points

| Path / module | Current role | Why relevant | Reuse / mirror |
| --- | --- | --- | --- |
| `internal/user/model.go` | Модель `User` (ID, Username, Email, …) | Нужно добавить `AvatarURL *string` | Существующие поля — без изменений |
| `internal/user/repo.go` | CRUD по `users`; `GetByEmail`, `GetByID`, `GetByUsernames`, `ListAll` | Нет `GetByUsername` (single); все `Scan` не включают `avatar_url` | Паттерн Scan из `GetByID`; паттерн `ErrNotFound` |
| `internal/tmpl/renderer.go` | Единственный владелец `funcMap`; уже есть `ruDate`, `truncate`, `deref`, `paginate` | Нужно добавить `avatarInitials`, `avatarColor`, `relativeTime` | Паттерн добавления функции в `funcMap` map-литерал |
| `internal/admin/images_handler.go` + `internal/admin/image_service.go` | Эталонный multipart upload: `c.FormFile`, size check, `imgSvc.Save` | Паттерн для `POST /profile/{username}/avatar` | Повторить: формат ошибок, size guard, open/close, service delegation |
| `internal/config/config.go` | `UploadsDir` default `./storage/news` | Путь `./storage/news` специфичен для статей; аватары нужно хранить отдельно — нужен `AvatarsDir` | Паттерн `getEnv(key, default)` |
| `internal/forum/repo.go` | `ListPostsByTopic`, `CreatePost`, `ListPostsAfter` | Нужно добавить `CountByUserID` и `LastPostByUserID` | Паттерн `QueryRow` + `Scan` |
| `internal/comment/repo.go` | `ListByNewsID`, `CreateComment` | Нужно добавить `CountByUserID` и `LastCommentByUserID` | Аналогичный паттерн |
| `internal/middleware/` | `LoadSession` (cookie → user в ctx), `RequireAuth` | `POST /profile/{username}/avatar` требует `RequireAuth` | Применить как middleware group, как в admin routes |
| `cmd/forum/main.go` | DI: config → pool → repos → services → handlers → routes | Точка регистрации `profile` package | Паттерн из `commentSvc := comment.NewService(...)` и регистрации routes |
| `templates/layouts/base.html` `{{define "nav"}}` | Текущий header: `<span class="nav-username">` | Заменить на аватар-кружок + кликабельное имя | Структуру `{{if .User}}` оставить |
| `templates/forum/topic.html` | Рендер постов | Аватар + кликабельное имя автора | Найти место рендера `AuthorUsername` |
| `templates/forum/index.html` / `section.html` | Список разделов / тем | Кликабельное имя последнего автора (если есть) | Найти `.LastPostBy` / `.LastPostUsername` |
| `templates/news/article.html` | Комментарии к статье | Аватар + кликабельное имя автора | `AuthorUsername` в loop |
| `migrations/012_idx_posts_topic_id.sql` | Последняя миграция | Следующий номер: `013` | Формат goose-файла |

## Test Strategy

| Test surface | Canonical refs | Existing coverage | Planned automated coverage | Required local suites | Required CI suites | Manual-only gap | Approval ref |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `GET /profile/{username}` | REQ-02, SC-02, NEG-04 | нет | Go integration test: 200 + данные, 404 для несуществующего | `go test -tags integration -p 1 ./internal/profile/...` | Go Tests job | — | none |
| `GET /profile/{username}/modal` | REQ-01, SC-01, NEG-06 | нет | Playwright: клик → модалка с данными; 5xx fallback через `page.route()` | Playwright E2E | E2E job | — | none |
| `POST /profile/{username}/avatar` | REQ-04, SC-03, SC-04, NEG-01..NEG-05 | нет | Go integration: 200 ok, 422 bad format, 413 too large, 403 ownership; Playwright: upload flow + preview | обе suites | Go Tests + E2E jobs | — | none |
| Header аватар + клик | REQ-03, REQ-06, SC-01 | нет | Playwright: авторизованный пользователь видит аватар в nav, клик → модалка | Playwright E2E | E2E job | — | none |
| Форум посты — аватар + клик | REQ-03, SC-01 | нет | Playwright: страница топика, клик на имя → модалка | Playwright E2E | E2E job | — | none |
| Fallback аватар (инициалы) | REQ-05, SC-06 | нет | Playwright: пользователь без аватара → видны инициалы; цвет стабилен при рефреше | Playwright E2E | E2E job | — | none |
| FuncMap `avatarColor`/`avatarInitials`/`relativeTime` | REQ-05, CON-04 | нет | Go unit тесты детерминизма и граничных случаев | `go test ./internal/tmpl/...` | Go Tests job | — | none |
| Закрытие модалки (ESC, крестик, вне) | SC-05 | нет | Playwright | Playwright E2E | E2E job | — | none |
| Пустые состояния | REQ-07, SC-07 | нет | Playwright: профиль без постов/комментариев | Playwright E2E | E2E job | — | none |

## Open Questions / Ambiguities

| OQ-ID | Question | Why unresolved | Blocks | Default action |
| --- | --- | --- | --- | --- |
| `OQ-01` | `UploadsDir` default — `./storage/news`. Аватары по `CON-03` идут в `$UPLOADS_DIR/avatars/{user_id}.webp`, что даёт `./storage/news/avatars/` — нелогично. Добавить отдельный `AVATARS_DIR` config key или сменить default `UPLOADS_DIR` на `./storage`? | Смена default UPLOADS_DIR ломает существующий путь к news-изображениям (если не обновить volume в docker-compose) | STEP-01 (config) | Добавить отдельный `AvatarsDir string` в config с default `./storage/avatars`; не трогать `UploadsDir` |

## Environment Contract

| Area | Contract | Used by | Failure symptom |
| --- | --- | --- | --- |
| setup | `docker compose -f docker-compose.dev.yml up -d` запущен, postgres healthy | все STEP | connection refused / no such host postgres |
| go test | `docker run --rm --network lfcru_forum_default -v lfcru_gomod:/go/pkg/mod golang:1.23-alpine go test -tags integration -p 1 ./...` (из `ops/development.md`) | CHK-01, CHK-02 | тесты падают с "no such host" — не запускать без `--network` |
| e2e | `docker compose -f docker-compose.e2e.yml up -d` + `npx playwright test` | CHK-03 | 404 или connection refused от app-e2e |
| avatars dir | `./storage/avatars/` должна существовать или создаваться при старте сервиса | STEP-04 | HTTP 500 при первой загрузке аватара |
| CSRF | все POST через `_csrf` hidden field | STEP-05, STEP-06 | 403 от CSRF middleware на upload |

## Preconditions

| PRE-ID | Canonical ref | Required state | Used by | Blocks start |
| --- | --- | --- | --- | --- |
| `PRE-01` | `feature.md` status: active | feature.md переведён в Design Ready | все STEP | yes |
| `PRE-02` | `ADR-005` decision_status: accepted | Filesystem + Docker volume как storage strategy принят | STEP-01, STEP-04 | yes |
| `PRE-03` | `ASM-01` | `idx_users_username` UNIQUE INDEX существует в migrations/001 | STEP-02 (`GetByUsername`) | yes |
| `PRE-04` | `ASM-04` | `RequireAuth` middleware существует в `internal/middleware/` | STEP-05 | yes |
| `PRE-05` | `OQ-01` resolved | `AvatarsDir` добавлен в config | STEP-04 (file save) | yes — решается в STEP-01 |

## Workstreams

| Workstream | Implements | Result | Dependencies |
| --- | --- | --- | --- |
| `WS-1` Backend core | REQ-02, REQ-04, REQ-05, CTR-01..CTR-04 | Migration, model, repo, config, FuncMap, service, handler зарегистрированы | PRE-01..PRE-05 |
| `WS-2` Templates | REQ-01, REQ-03, REQ-06, REQ-07 | Новые шаблоны profile + обновлённые base/forum/news | WS-1 (FuncMap нужен для render) |
| `WS-3` Tests | CHK-01..CHK-03 | Go integration + Playwright тесты зелёные | WS-1, WS-2 |

## Approval Gates

| AG-ID | Trigger | Applies to | Why | Approver |
| --- | --- | --- | --- | --- |
| `AG-01` | Перед merge PR в main | Closure | PR ready for review — человек проверяет CI green + SC-* визуально | пользователь |

## Порядок работ

| Step ID | Actor | Implements | Goal | Touchpoints | Artifact | Verifies | Evidence IDs | Check | Blocked by | Needs approval |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `STEP-01` | agent | CON-03, OQ-01 | Добавить `AvatarsDir` в config; создать миграцию 013 | `internal/config/config.go`, `migrations/013_add_avatar_url_to_users.sql` | config field + SQL файл | — | — | файл создан, goose синтаксис валиден | PRE-01 | none |
| `STEP-02` | agent | REQ-03, REQ-04 | Расширить `user.User` + repo | `internal/user/model.go`, `internal/user/repo.go` | `AvatarURL *string` в модели; `GetByUsername`; обновлённые `Scan` во всех методах | — | — | `go build ./internal/user/...` | STEP-01 | none |
| `STEP-03` | agent | REQ-05, CON-04, ASM-03 | Добавить FuncMap-функции | `internal/tmpl/renderer.go` | `avatarInitials`, `avatarColor`, `relativeTime` в funcMap | CHK-01 (unit) | EVID-01 | `go test ./internal/tmpl/...` | STEP-02 | none |
| `STEP-04` | agent | REQ-02, REQ-04, CTR-01..CTR-04 | Создать profile service | `internal/profile/service.go` (new), `internal/forum/repo.go`, `internal/comment/repo.go` | service с методами `GetProfile`, `SaveAvatar`; новые repo-методы `CountByUserID`, `LastPostByUserID`, `LastCommentByUserID` | CHK-02 (integration) | EVID-01 | `go test -tags integration -p 1 ./internal/profile/...` | STEP-02, STEP-03 | none |
| `STEP-05` | agent | CTR-01..CTR-03, FM-01..FM-06, CON-01, CON-02 | Создать profile handler + зарегистрировать routes | `internal/profile/handler.go` (new), `cmd/forum/main.go` | Handler с 3 методами; routes в Echo | — | — | `go build ./...` | STEP-04 | none |
| `STEP-06` | agent | REQ-01, REQ-02, REQ-07, SC-01..SC-07 | Создать шаблоны профиля | `templates/profile/page.html` (new), `templates/profile/modal.html` (new) | 2 новых шаблона | — | — | `go build ./...` (renderer подхватит) | STEP-05 | none |
| `STEP-07` | agent | REQ-03, REQ-06, SC-01 | Обновить существующие шаблоны | `templates/layouts/base.html`, `templates/forum/topic.html`, `templates/forum/index.html`, `templates/forum/section.html`, `templates/news/article.html` | кликабельные имена/аватары, обновлённый header nav | — | — | визуальная проверка в браузере + E2E | STEP-06 | none |
| `STEP-08` | agent | CHK-01..CHK-03 | Написать Go unit + integration тесты | `internal/tmpl/renderer_test.go`, `internal/profile/service_test.go` | тесты для FuncMap и service | CHK-01, CHK-02 | EVID-01 | `rtk go test` | STEP-04, STEP-05 | none |
| `STEP-09` | agent | CHK-03 | Написать Playwright E2E тесты | `e2e/profile.spec.ts` (new) | тесты SC-01..SC-07, NEG-01..NEG-06 | CHK-03 | EVID-01 | `rtk playwright test` | STEP-07 | none |
| `STEP-10` | agent | EC-01..EC-07 | Финальный verify: push + CI | git push + PR checks | зелёный CI | CHK-01..CHK-03 | EVID-01, EVID-02 | `rtk gh pr checks` | STEP-09 | AG-01 |

## Parallelizable Work

- `PAR-01` STEP-06 (шаблоны профиля) и STEP-08 (Go тесты) могут идти параллельно после STEP-05.
- `PAR-02` STEP-07 (обновление существующих шаблонов) зависит от STEP-06 (нужны `modal.html` и FuncMap-функции в шаблонах).

## Checkpoints

| CP-ID | Refs | Condition | Evidence IDs |
| --- | --- | --- | --- |
| `CP-01` | STEP-01..STEP-05 | `go build ./...` чистый; `GET /profile/{username}` возвращает HTML | — |
| `CP-02` | STEP-06..STEP-07 | Модалка открывается по клику в браузере; header показывает аватар/имя | — |
| `CP-03` | STEP-08..STEP-09 | `rtk go test` + `rtk playwright test` зелёные локально | EVID-01 |
| `CP-04` | STEP-10 | CI: все 3 job-а (Lint, Go Tests, E2E) зелёные | EVID-01, EVID-02 |

## Execution Risks

| ER-ID | Risk | Impact | Mitigation | Trigger |
| --- | --- | --- | --- | --- |
| `ER-01` | `Scan` в `user/repo.go` во всех методах не учитывает новый `avatar_url` → runtime panic | все страницы с пользователем ломаются | STEP-02: обновить все методы AtomIcally, запустить `go test ./internal/user/...` | `go test` panic / nil pointer в runtime |
| `ER-02` | Межмодульные зависимости: `profile/service` напрямую импортирует `forum.Repo` вместо интерфейса | нарушение module boundaries из `architecture.md` | объявить `ForumPostsRepo` и `CommentsRepo` интерфейсы в `profile` пакете | `go vet` / code review |
| `ER-03` | `avatarColor` возвращает разные цвета при изменении палитры → fallback нестабилен | SC-06 / EC-04 падает | Зафиксировать палитру как const; unit-тест на детерминизм | тест в STEP-08 |
| `ER-04` | Playwright E2E: upload тест требует реального файла и запущенного app-e2e контейнера | если контейнер не поднят, тест падает с connection refused | в CI: docker-compose.e2e.yml поднимается в workflow; локально: инструкция в ops/development.md | `npx playwright test` connection refused |

## Stop Conditions / Fallback

| STOP-ID | Refs | Trigger | Immediate action | Safe fallback |
| --- | --- | --- | --- | --- |
| `STOP-01` | ER-01 | `go build ./...` не проходит после STEP-02 | Откатить изменения в `user/model.go` и `user/repo.go`; зафиксировать ошибку в `OQ-01` | `user` package в исходном состоянии |
| `STOP-02` | ER-02 | Code review замечает прямой импорт `forum.Repo` в `profile/service.go` | Ввести интерфейс, перед merge не пушить | `profile/service.go` с интерфейсом |
| `STOP-03` | FM-03 | Запись аватара на FS возвращает ошибку в prod-like тесте | Проверить наличие директории `./storage/avatars/`; добавить `os.MkdirAll` в service.Init | HTTP 500 с пользователю-понятным сообщением |

## Готово для приёмки

- `go build ./...` чистый
- `rtk go test` (unit + integration) зелёные локально
- `rtk playwright test` зелёные локально
- CI: Lint + Go Tests + E2E jobs зелёные на PR
- SC-01..SC-07 визуально проверены в браузере через app-e2e
- PR переведён из draft → ready for review
