---
title: "FT-025: Implementation Plan"
doc_kind: feature
doc_function: derived
purpose: "Execution-план реализации FT-025 — отображение аватарок в header, постах форума и комментариях к новостям."
derived_from:
  - feature.md
status: final
audience: humans_and_agents
must_not_define:
  - ft_025_scope
  - ft_025_architecture
  - ft_025_acceptance_criteria
  - ft_025_blocker_state
---

> **Lifecycle note (E-1):** `status: draft` корректен во время ревью. Переход в `status: active` выполняется при подтверждении Plan Ready человеком (до первого коммита с кодом) — см. STEP-00.

# План имплементации

## Цель текущего плана

Добавить отображение аватарок пользователей в трёх поверхностях (header, посты форума, комментарии к новостям) с кликом по аватарке → quick-view модалка профиля.

## Discovery Context / Reference Points

| Path / module | Current role | Why relevant | Reuse / mirror |
| --- | --- | --- | --- |
| `internal/tmpl/renderer.go` | FuncMap с `avatarColor`, `avatarInitials`, `avatarPalette` | ASM-02 подтверждён: функции уже зарегистрированы, изменений не требуется | Использовать `{{avatarColor .AuthorUsername}}` / `{{avatarInitials .AuthorUsername}}` по аналогии с `templates/profile/page.html` |
| `internal/forum/model.go` | `PostView` — проекция поста для шаблонов и SSE | Нет поля `AuthorAvatarURL *string` — нужно добавить | Паттерн: `AuthorUsername string` рядом |
| `internal/forum/repo.go:110` | `ListPostsByTopic` — LEFT JOIN users уже есть | Только добавить `u.avatar_url` в SELECT и Scan | Добавлять последним в SELECT; Scan в том же порядке (ER-01) |
| `internal/forum/repo.go:285` | `ListPostsAfter` — LEFT JOIN users уже есть | SSE catch-up: тот же паттерн | Идентично ListPostsByTopic |
| `internal/forum/handler.go:428` | `renderPostFragment` использует `PostView` для SSE | SSE получит `AuthorAvatarURL` автоматически после изменения модели и репо | — |
| `internal/comment/model.go` | `CommentView` — проекция комментария | Нет поля `AuthorAvatarURL *string` | Паттерн: `AuthorUsername string` рядом |
| `internal/comment/repo.go:19` | `ListByNewsID` — **INNER JOIN** users (не LEFT JOIN как в forum) | Разница от forum/repo: comment/repo использует INNER JOIN; поведение идентично для существующих строк — OQ-01 | Добавлять `u.avatar_url` последним в SELECT; Scan в том же порядке |
| `templates/layouts/base.html` | `{{define "nav"}}` — header навигация | `.User.AvatarURL *string` уже доступен (user.User целиком) | Паттерн аватар-кружка из `templates/profile/modal.html` |
| `templates/forum/partials/post.html` | `{{define "post"}}` — рендер поста и SSE-фрагмент | Нужно аватар-кружок + HTMX-триггер модалки | Паттерн `.profile-link` из `templates/news/article.html:127` |
| `templates/news/article.html` | `{{define "comments-list"}}` — список комментариев | Нужно аватар-кружок + HTMX-триггер модалки | Паттерн `.profile-link` уже есть в этом же файле |
| `e2e/profile.spec.ts` | Существующие E2E тесты профиля | Паттерн: `data-testid`, `storageState`, `login()`, фиксированные ID через `OVERRIDING SYSTEM VALUE` | Зеркалить структуру; seed/teardown в `global-setup.ts`/`global-teardown.ts` |

## Test Strategy

| Test surface | Canonical refs | Existing coverage | Planned automated coverage | Required local suites | Required CI suites | Manual-only gap | Approval ref |
| --- | --- | --- | --- | --- | --- | --- | --- |
| Go unit (forum/comment model+repo) | REQ-02, REQ-03, CON-02 | Существующие unit-тесты forum и comment | Существующие тесты должны проходить (новые unit не требуются — compile-check достаточен; scan-extension покрыто integration) | `docker run --rm -v "$(pwd)":/app -w /app -v lfcru_gomod:/root/go/pkg/mod golang:1.23-alpine go test ./...` | Go Tests (CI): unit + integration `go test -tags integration -p 1 ./internal/...` | — | none |
| Go integration (forum/comment repo) | REQ-02, REQ-03, CON-02 | Существующие integration-тесты forum и comment | Существующие integration-тесты должны проходить после Scan-extension (scan order change → panic → blocker) | — *(не запускаются локально; только CI)* | Go Tests (CI): `go test -tags integration -p 1 ./internal/...` | — | none |
| Playwright: avatar in nav | REQ-01, SC-01, CHK-01 | Нет | `e2e/profile/avatar-display.spec.ts`: `[data-testid="nav-avatar"]` visible + click → modal | `npx playwright test` | E2E (CI) | — | none |
| Playwright: avatar in forum post | REQ-02, SC-02, CHK-02 | Нет | `e2e/forum/avatar-display.spec.ts`: `[data-testid="post-avatar"]` visible + click → modal; seed: topic_id=9999 (e2e fixture из global-setup.ts) | `npx playwright test` | E2E (CI) | — | none |
| Playwright: avatar in news comment | REQ-03, SC-03, CHK-03 | Нет | `e2e/news/avatar-display.spec.ts`: `[data-testid="comment-avatar"]` visible + click → modal; seed: news_id из e2e fixture | `npx playwright test` | E2E (CI) | — | none |
| Playwright: initials fallback | REQ-04, SC-04, CHK-04 | `e2e/profile.spec.ts` SC-06 | `e2e/profile/avatar-display.spec.ts`: initials visible in nav/forum/comments, `page.on('console')` 0 errors | `npx playwright test` | E2E (CI) | — | none |

### E2E Test Data Contract

- **Форум (CHK-02):** использовать существующий e2e-топик `topic_id=9999` из `global-setup.ts` (паттерн из `e2e/profile.spec.ts:TOPIC_URL`). Автор поста — `e2e_user` (фиксированный ID из seed).
- **Новости/комментарии (CHK-03):** использовать существующий e2e-article или добавить seed в `global-setup.ts` с фиксированным `news_id` через `INSERT ... OVERRIDING SYSTEM VALUE`. Teardown — в `global-teardown.ts`.
- **Nav (CHK-01, CHK-04):** авторизованный пользователь `e2e_user` — session через `login()` helper.

## Open Questions / Ambiguities

| Open Question ID | Question | Why unresolved | Blocks | Default action |
| --- | --- | --- | --- | --- |
| `OQ-01` | `comment/repo.go` использует INNER JOIN, `forum/repo.go` — LEFT JOIN. Нужно ли менять comment на LEFT JOIN? | Поведение идентично для существующих данных (все комментарии имеют автора). INNER JOIN корректен: если автор удалён — комментарий не отображается (аналогично username). | STEP-04 | Оставить INNER JOIN; добавить `u.avatar_url` как nullable через `*string`. Если integration-тест падает — эскалация. |

## Environment Contract

| Area | Contract | Used by | Failure symptom |
| --- | --- | --- | --- |
| Go unit-тесты | `docker run --rm -v "$(pwd)":/app -w /app -v lfcru_gomod:/root/go/pkg/mod golang:1.23-alpine go test ./...` | STEP-08 | Compilation error или test failure |
| Go integration-тесты | Запускаются **только в CI** — не запускать локально (testing-policy.md) | CI: Go Tests job | CI Go Tests fail |
| E2E тесты (prerequisite) | `docker compose -f docker-compose.dev.yml up -d` затем `docker compose -f docker-compose.e2e.yml up -d` | STEP-09 | Playwright timeout (app недоступен) |
| E2E тесты (run) | `npx playwright test` | STEP-09 | Assertion failure или screenshot |
| Go на хосте | Не используется (PCON-04) | все Go-шаги | Если кто-то запускает `go` напрямую — нарушение |

## Preconditions

| Precondition ID | Canonical ref | Required state | Used by steps | Blocks start |
| --- | --- | --- | --- | --- |
| `PRE-01` | ASM-02 | `avatarColor`/`avatarInitials` зарегистрированы в funcMap — подтверждено grounding | STEP-05, STEP-06, STEP-07 | yes |
| `PRE-02` | CON-02 | `PostView` и `CommentView` не содержат `AuthorAvatarURL` до STEP-01/STEP-03 | STEP-02, STEP-04 | yes |
| `PRE-03` | feature.md status: active | Design Ready получен; Plan Ready подтверждён человеком | STEP-00 | yes |

## Workstreams

| Workstream | Implements | Result | Dependencies |
| --- | --- | --- | --- |
| `WS-1` | REQ-02, REQ-03 | Go-модели и SQL обновлены (forum + comment) | PRE-02 |
| `WS-2` | REQ-01, REQ-02, REQ-03, REQ-04 | Шаблоны обновлены (base, post, article) | WS-1 (post.html и article.html зависят от новых полей) |
| `WS-3` | CHK-01..04, EVID-01..04 | E2E спеки созданы | WS-2 |

## Approval Gates

Нет рискованных/необратимых действий. Все изменения — локальные файлы. Merge PR требует human review (pipeline).

## Порядок работ

| Step ID | Actor | Implements | Goal | Touchpoints | Artifact | Verifies | Evidence IDs | Check | Blocked by | Needs approval |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `STEP-00` | agent | — | Перевести `implementation-plan.md → status: active` и `feature.md → delivery_status: in_progress` | `implementation-plan.md`, `feature.md` | Оба файла обновлены | — | — | `git diff` | PRE-03 | none |
| `STEP-01` | agent | REQ-02 | Добавить `AuthorAvatarURL *string` в `PostView` | `internal/forum/model.go` | Обновлённый `PostView` | — | — | Компилируется | PRE-02 | none |
| `STEP-02` | agent | REQ-02 | Добавить `u.avatar_url` последним в SELECT и Scan для `ListPostsByTopic` и `ListPostsAfter` | `internal/forum/repo.go` | Обновлённые SQL-запросы | CHK-02 | EVID-02 | compile | STEP-01 | none |
| `STEP-03` | agent | REQ-03 | Добавить `AuthorAvatarURL *string` в `CommentView` | `internal/comment/model.go` | Обновлённый `CommentView` | — | — | Компилируется | PRE-02 | none |
| `STEP-04` | agent | REQ-03 | Добавить `u.avatar_url` последним в SELECT и Scan для `ListByNewsID` (INNER JOIN остаётся — OQ-01) | `internal/comment/repo.go` | Обновлённый SQL | CHK-03 | EVID-03 | compile | STEP-03 | none |
| `STEP-05` | agent | REQ-01, REQ-04 | Добавить аватар-кружок в nav header; клик = HTMX-триггер модалки; `data-testid="nav-avatar"` | `templates/layouts/base.html` | Обновлённый `{{define "nav"}}` | CHK-01, CHK-04 | EVID-01, EVID-04 | визуальная проверка | PRE-01 | none |
| `STEP-06` | agent | REQ-02, REQ-04 | Добавить аватар-кружок слева от имени автора поста; клик = HTMX-триггер; `data-testid="post-avatar"` | `templates/forum/partials/post.html` | Обновлённый `{{define "post"}}` | CHK-02, CHK-04 | EVID-02, EVID-04 | визуальная проверка | STEP-01, PRE-01 | none |
| `STEP-07` | agent | REQ-03, REQ-04 | Добавить аватар-кружок слева от имени автора комментария; клик = HTMX-триггер; `data-testid="comment-avatar"` | `templates/news/article.html` | Обновлённый `{{define "comments-list"}}` | CHK-03, CHK-04 | EVID-03, EVID-04 | визуальная проверка | STEP-03, PRE-01 | none |
| `STEP-08` | agent | CON-01 | Прогнать unit-тесты | docker + golang:1.23-alpine | Вывод `go test ./...` | — | — | `docker run --rm -v "$(pwd)":/app -w /app -v lfcru_gomod:/root/go/pkg/mod golang:1.23-alpine go test ./...` | STEP-01..07 | none |
| `STEP-09` | agent | CHK-01..04 | Создать E2E спеки с seeding согласно E2E Test Data Contract | `e2e/profile/`, `e2e/forum/`, `e2e/news/` | Три новых спека | CHK-01..04 | EVID-01..04 | `npx playwright test` (после поднятия стеков) | STEP-08 | none |
| `STEP-10` | agent | — | Simplify review: убедиться что изменения минимальны, нет лишних абстракций | все изменённые файлы | — | — | — | code review checklist | STEP-09 | none |
| `STEP-11` | agent | — | Коммит и push | git | Коммит в feat/FT-025-avatar-display | — | — | `git status` | STEP-10 | none |

## Parallelizable Work

- `PAR-01` STEP-01 и STEP-03 можно выполнять параллельно (разные пакеты).
- `PAR-02` STEP-02 и STEP-04 зависят от STEP-01/STEP-03 соответственно, но независимы друг от друга.
- `PAR-03` STEP-05 не зависит от WS-1 (header использует `.User.AvatarURL` из user.User напрямую).

## Checkpoints

| Checkpoint ID | Refs | Condition | Evidence IDs |
| --- | --- | --- | --- |
| `CP-01` | STEP-01..04 | Go-модели и SQL обновлены, `go test ./...` зелёный | — |
| `CP-02` | STEP-05..07 | Все три шаблона обновлены | — |
| `CP-03` | STEP-08..09 | Unit-тесты pass локально, E2E спеки созданы | EVID-01..04 |
| `CP-04` | STEP-10 | Simplify review пройден | — |

## Execution Risks

| Risk ID | Risk | Impact | Mitigation | Trigger |
| --- | --- | --- | --- | --- |
| `ER-01` | Порядок Scan в repo не совпадёт с порядком SELECT после добавления `u.avatar_url` | Runtime panic или неправильные данные | Добавлять `u.avatar_url` последним в SELECT; Scan в том же порядке | `go test ./...` compile error или integration-тест падает |
| `ER-02` | `post.html` используется для SSE-фрагмента — изменения шаблона могут повлиять на SSE-доставку | SSE перестаёт работать | Проверить что `renderPostFragment` продолжает работать; `e2e/forum/reply.spec.ts` должен остаться зелёным | `e2e/forum/reply.spec.ts` падает после STEP-06 |

## Stop Conditions / Fallback

| Stop ID | Refs | Trigger | Immediate action | Safe fallback |
| --- | --- | --- | --- | --- |
| `STOP-01` | ER-01 | `go test ./...` падает после Scan-изменений | Откатить Scan до рабочего порядка, пересмотреть SELECT | Состояние до STEP-02/STEP-04 |
| `STOP-02` | ER-02 | `e2e/forum/reply.spec.ts` падает после STEP-06 | Откатить изменение шаблона поста | Состояние до STEP-06 |

## Готово для приемки

- `go test ./...` (unit) — зелёный локально
- CI (Lint, Go Tests unit+integration, E2E) — зелёный
- Все SC-* из `feature.md` закрыты
- Simplify review (STEP-10) пройден
