---
title: "FT-025: Отображение аватарок пользователей"
doc_kind: feature
doc_function: canonical
purpose: "Отображение аватарок пользователей в трёх UI-поверхностях: header навигации, посты форума, комментарии к новостям. Клик по аватарке открывает quick-view модалку профиля."
derived_from:
  - ../../domain/problem.md
  - ../../domain/frontend.md
status: active
delivery_status: in_progress
audience: humans_and_agents
must_not_define:
  - implementation_sequence
---

# FT-025: Отображение аватарок пользователей

## What

### Problem

В FT-024 реализованы загрузка и хранение аватарок пользователей, а также quick-view модалка профиля. Однако сами аватарки нигде не отображаются: ни в header, ни в постах форума, ни в комментариях к новостям. Пользователь загружает аватарку, но никогда её не видит в интерфейсе. Кроме того, клик по имени автора в этих поверхностях уже открывает модалку, но кликнуть по аватарке нельзя.

### Outcome

| Metric ID | Metric | Baseline | Target | Measurement method |
| --- | --- | --- | --- | --- |
| `MET-01` | Аватарка отображается в header | не отображается | отображается | ручная проверка + Playwright |
| `MET-02` | Аватарка отображается в постах форума | не отображается | отображается | ручная проверка + Playwright |
| `MET-03` | Аватарка отображается в комментариях к новостям | не отображается | отображается | ручная проверка + Playwright |

### Scope

- `REQ-01` В header: слева от username авторизованного пользователя отображается аватарка (кружок с фото или инициалами). Клик по аватарке открывает quick-view модалку профиля.
- `REQ-02` В постах форума: слева от имени автора поста отображается аватарка. Клик по аватарке открывает quick-view модалку профиля.
- `REQ-03` В комментариях к новостям: слева от имени автора комментария отображается аватарка. Клик по аватарке открывает quick-view модалку профиля.
- `REQ-04` Если у пользователя нет загруженной аватарки — отображается кружок с инициалами на фоне, вычисленном через hash от username (те же функции `avatarColor`/`avatarInitials`, что и на странице профиля).

### Non-Scope

- `NS-01` Загрузка и хранение аватарок — реализованы в FT-024, не меняются.
- `NS-02` Страница профиля и модалка профиля — реализованы в FT-024, не меняются.
- `NS-03` Аватарки для гостей (неавторизованных) — гости не могут писать посты или комментарии, случай не существует.
- `NS-04` Аватарки в любых других поверхностях (страница форума-раздела, список тем, admin-панель) — вне scope.
- `NS-05` Изменение размеров хранимых аватарок или формата хранения — вне scope.

### Constraints / Assumptions

- `ASM-01` `user.User.AvatarURL *string` уже присутствует в модели; в header `.User` передаётся целиком — дополнительных изменений структуры не требуется.
- `ASM-02` FuncMap функции `avatarColor` и `avatarInitials` уже зарегистрированы в `tmpl/renderer.go` и доступны во всех шаблонах.
- `CON-01` PCON-01: SQL только через параметризованные запросы — JOIN для получения `avatar_url` должен использовать только параметризованные запросы.
- `CON-02` `PostView` и `CommentView` не содержат `AvatarURL` — необходимо расширить модели и обновить SQL.

## How

### Solution

Для header: чисто шаблонное изменение `base.html` — `.User.AvatarURL` уже доступен в контексте. Добавляем аватарку слева от username, оборачиваем её в тот же HTMX-триггер модалки.

Для форума и комментариев: добавляем поле `AuthorAvatarURL *string` в `PostView` и `CommentView`, расширяем SQL-запросы LEFT JOIN с `users`, обновляем шаблоны `post.html` и `article.html`.

**Trade-off:** JOIN в SQL vs lazy HTMX-fetch аватарки по клику. Выбран JOIN: аватарка нужна при первом рендере страницы (без дополнительных запросов), JOIN не добавляет N+1, а `users.avatar_url` уже индексирован через PRIMARY KEY. HTMX-fetch потребовал бы отдельного эндпоинта и дополнительного round-trip на каждый пост/комментарий.

`DEC-01` — Решение принято: JOIN при листинге постов/комментариев. Альтернатива (lazy fetch) отклонена из-за N+1 и сложности шаблона.

### Change Surface

| Surface | Type | Why it changes |
| --- | --- | --- |
| `templates/layouts/base.html` | template | Добавить аватарку в nav для REQ-01 |
| `templates/forum/partials/post.html` | template | Добавить аватарку автора поста для REQ-02 |
| `templates/news/article.html` | template | Добавить аватарку автора комментария для REQ-03 |
| `internal/forum/model.go` | code | Добавить `AuthorAvatarURL *string` в `PostView` |
| `internal/forum/repo.go` | code | Добавить `u.avatar_url` в SELECT и Scan для `ListPostsByTopic` и `ListPostsAfter` |
| `internal/comment/model.go` | code | Добавить `AuthorAvatarURL *string` в `CommentView` |
| `internal/comment/repo.go` | code | Добавить `u.avatar_url` в SELECT и Scan |
| `internal/tmpl/renderer.go` | code | Проверить при grounding: `avatarColor`/`avatarInitials` уже зарегистрированы (ASM-02) — изменений не требуется |

### Flow

1. Авторизованный пользователь открывает любую страницу с header / форум-тему / статью с комментариями.
2. Backend рендерит страницу: для форума и комментариев — JOIN с `users` для получения `avatar_url`.
3. Шаблон рендерит аватарку (фото или инициалы) слева от имени автора.
4. Пользователь кликает на аватарку → HTMX-запрос `/profile/{username}/modal` → появляется quick-view модалка профиля.

### Failure Modes

- `FM-01` `avatar_url` NULL в БД → шаблон отображает инициалы через `avatarColor`/`avatarInitials`; фото не нужно.
- `FM-02` Файл аватарки удалён с диска, но URL есть в БД → сломанная картинка. Выходит за scope — проблема доставки статики, не данного feature.

## Verify

### Exit Criteria

- `EC-01` В header авторизованного пользователя отображается аватарка (фото или инициалы) слева от username; клик открывает модалку.
- `EC-02` В постах форума отображается аватарка автора слева от имени; клик открывает модалку.
- `EC-03` В комментариях к новостям отображается аватарка автора слева от имени; клик открывает модалку.
- `EC-04` Если аватарки нет — кружок с инициалами на hash-цвете; страница без JS-ошибок в консоли.

### Traceability matrix

| Requirement ID | Design refs | Acceptance refs | Checks | Evidence IDs |
| --- | --- | --- | --- | --- |
| `REQ-01` | `ASM-01`, `ASM-02` | `EC-01`, `SC-01` | `CHK-01` | `EVID-01` |
| `REQ-02` | `ASM-02`, `CON-02` | `EC-02`, `SC-02` | `CHK-02` | `EVID-02` |
| `REQ-03` | `ASM-02`, `CON-02` | `EC-03`, `SC-03` | `CHK-03` | `EVID-03` |
| `REQ-04` | `ASM-02`, `FM-01` | `EC-04`, `SC-04` | `CHK-04` | `EVID-04` |

### Acceptance Scenarios

- `SC-01` Авторизованный пользователь с аватаркой: в header слева от имени отображается фото-аватарка; клик по ней открывает модалку профиля.
- `SC-02` Авторизованный пользователь просматривает тему форума: у каждого поста слева от имени автора — аватарка (фото или инициалы); клик по аватарке открывает модалку профиля автора.
- `SC-03` Авторизованный пользователь просматривает статью с комментариями: у каждого комментария слева от имени — аватарка; клик открывает модалку.
- `SC-04` Пользователь без загруженной аватарки: во всех трёх поверхностях отображается кружок с инициалами на hash-цвете; JS-ошибок в консоли нет.

### Checks

| Check ID | Covers | How to check | Expected result | Evidence path |
| --- | --- | --- | --- | --- |
| `CHK-01` | `EC-01`, `SC-01` | Playwright `e2e/profile/avatar-display.spec.ts`: открыть любую страницу под auth-пользователем, assert `[data-testid="nav-avatar"]` visible, click → modal visible | Аватарка видна, клик открывает модалку | `e2e/test-report/index.html` |
| `CHK-02` | `EC-02`, `SC-02` | Playwright `e2e/forum/avatar-display.spec.ts`: открыть тему форума, assert `[data-testid="post-avatar"]` visible у каждого поста, click → modal | Аватарка видна у каждого поста | `e2e/test-report/index.html` |
| `CHK-03` | `EC-03`, `SC-03` | Playwright `e2e/news/avatar-display.spec.ts`: открыть статью с комментариями, assert `[data-testid="comment-avatar"]` visible, click → modal | Аватарка видна у каждого комментария | `e2e/test-report/index.html` |
| `CHK-04` | `EC-04`, `SC-04` | Playwright `e2e/profile/avatar-display.spec.ts`: пользователь без аватарки — assert `[data-testid*="initials"]` visible, `page.on('console')` 0 errors | Инициалы видны, 0 JS-ошибок | `e2e/test-report/index.html` |

### Test matrix

| Check ID | Evidence IDs | Evidence path |
| --- | --- | --- |
| `CHK-01` | `EVID-01` | `e2e/test-report/index.html` |
| `CHK-02` | `EVID-02` | `e2e/test-report/index.html` |
| `CHK-03` | `EVID-03` | `e2e/test-report/index.html` |
| `CHK-04` | `EVID-04` | `e2e/test-report/index.html` |

### Evidence

- `EVID-01` Playwright HTML-отчёт (`e2e/test-report/index.html`): спек `e2e/profile/avatar-display.spec.ts`, тест `nav avatar visible and opens modal` — pass.
- `EVID-02` Playwright HTML-отчёт (`e2e/test-report/index.html`): спек `e2e/forum/avatar-display.spec.ts`, тест `forum post avatar visible and opens modal` — pass.
- `EVID-03` Playwright HTML-отчёт (`e2e/test-report/index.html`): спек `e2e/news/avatar-display.spec.ts`, тест `comment avatar visible and opens modal` — pass.
- `EVID-04` Playwright HTML-отчёт (`e2e/test-report/index.html`): спек `e2e/profile/avatar-display.spec.ts`, тест `initials fallback, no console errors` — pass.

### Eval Evidence

- `EVID-BR-01`: Brief loop — accept. 2026-05-04. improve-loop.sh / evaluator agent → [review-brief-01.md](../../.review-results/FT-025/review-brief-01.md)
- `EVID-SP-01`: Spec loop — accept. 2026-05-04. improve-loop.sh / evaluator agent → [review-spec-02.md](../../.review-results/FT-025/review-spec-02.md)
- `EVID-PR-01`: Eval DR→PR — accept. 2026-05-05. evaluator agent → [review-implementation-plan-02.md](../../.review-results/FT-025/review-implementation-plan-02.md)

### Evidence contract

| Evidence ID | Artifact | Producer | Path contract | Reused by checks |
| --- | --- | --- | --- | --- |
| `EVID-01` | Playwright HTML report | CI / local npx playwright test | `e2e/test-report/index.html` | `CHK-01` |
| `EVID-02` | Playwright HTML report | CI / local npx playwright test | `e2e/test-report/index.html` | `CHK-02` |
| `EVID-03` | Playwright HTML report | CI / local npx playwright test | `e2e/test-report/index.html` | `CHK-03` |
| `EVID-04` | Playwright HTML report | CI / local npx playwright test | `e2e/test-report/index.html` | `CHK-04` |
