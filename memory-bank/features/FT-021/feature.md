---
title: "FT-021: Редизайн главной страницы"
doc_kind: feature
doc_function: canonical
purpose: "Переработка визуального дизайна главной страницы по макету homeworks/hw-2/. Presentation layer + backend-расширение HomeData.News (CoverImageURL, CommentCount)."
derived_from:
  - ../../domain/problem.md
status: active
delivery_status: planned
audience: humans_and_agents
must_not_define:
  - implementation_sequence
---

# FT-021: Редизайн главной страницы

## What

### Problem

Главная страница сайта использует минимальный дизайн, не отражающий визуальную идентичность клуба и не предоставляющий пользователю достаточного контекста (нет hero-изображений новостей, нет структурированной правой колонки, нет выраженного footer). Необходима переработка по готовому макету `homeworks/hw-2/`.

### Outcome

| Metric ID | Metric | Baseline | Target | Measurement method |
| --- | --- | --- | --- | --- |
| `MET-01` | Главная страница соответствует макету hw-2 визуально | Не соответствует | Все разделы реализованы по макету | Playwright screenshot + ручная сверка |
| `MET-02` | Все затронутые страницы рендерятся без регрессий | N/A | /news, /forum, /login, /register, /news/{id} — без визуальных поломок | Playwright assertions |

### Scope

- `REQ-01` Header: логотип LFC (локальный файл `static/img/lfc_crest.webp`), ссылки «Новости / Форум» слева, кнопки «Войти / Регистрация» справа.
- `REQ-02` Новостная лента (левая колонка): каждая статья — hero-изображение + заголовок + excerpt + дата + счётчик комментариев.
- `REQ-03` Backend: добавить `CoverImageURL` (JOIN с `article_images`) и `CommentCount` в `HomeData.News`.
- `REQ-04` Таблица АПЛ (правая колонка): компактный срез — 2 выше LFC + LFC + 2 ниже LFC (с граничными случаями). При клике — inline expand полной таблицы через Alpine.js toggle. Иконки клубов сохранить.
- `REQ-05` Виджеты матчей (правая колонка): шаблон и CSS без реальных данных — «Последний матч» (крупный счёт) и «Следующий матч» (формат «Everton VS Liverpool» с выделением Liverpool).
- `REQ-06` Активность форума (правая колонка): стилизованный список — тег-бейдж категории, автор, ответы, время.
- `REQ-07` Footer: 3 колонки («О сообществе» / «Разделы» / «Мы в соцсетях»), иконки соцсетей с `href="#"`.

### Non-Scope

- `NS-01` Football API, кэш, `football.Client` — не трогать.
- `NS-02` SSE / real-time форума (`hub.go`, `sse.js`) — не трогать.
- `NS-03` Auth, сессии, CSRF, middleware — не трогать.
- `NS-04` CSS комментариев в `templates/news/article.html` — не трогать.
- `NS-05` `templates/admin/layouts/layout.html` — отдельный layout, не затронут.
- `NS-06` Реальные данные для виджетов матчей — шаблон/CSS only, данные — отдельная фича.
- `NS-07` Переход на отдельную страницу таблицы АПЛ — inline expand достаточен.

### Constraints / Assumptions

- `ASM-01` Макет `homeworks/hw-2/` (скриншоты + `site.html`) является референсом дизайна.
- `ASM-02` Данные таблицы АПЛ уже реализованы в FT-020 и доступны в шаблоне.
- `CON-01` `templates/layouts/base.html` — глобальный layout: изменения header/footer затронут ВСЕ страницы сайта. Регрессионная проверка обязательна.
- `CON-02` Логотип LFC должен отдаваться локально (hotlinking ненадёжен) — скачать в `static/img/lfc_crest.webp`.

### Do Not Touch

- `NT-01` Иконки клубов в таблице (`.standings-team img`) — сохранить без изменений.

## How

### Solution

Переработать `templates/layouts/base.html` (header + footer) и `templates/home/index.html` (все виджеты правой колонки и новостная лента). Backend-слой расширяется минимально: добавить JOIN с `article_images` и COUNT из `comments` в запрос для главной страницы. Inline expand таблицы АПЛ реализуется через Alpine.js `x-show` + `x-data` без дополнительных HTMX-запросов.

### Change Surface

| Surface | Type | Why it changes |
| --- | --- | --- |
| `templates/layouts/base.html` | template | Header + footer redesign (REQ-01, REQ-07) |
| `templates/home/index.html` | template | Новостная лента + правая колонка (REQ-02, REQ-04, REQ-05, REQ-06) |
| `internal/handler/home.go` | code | Расширение HomeData.News: CoverImageURL + CommentCount (REQ-03) |
| `internal/store/home_store.go` | code | JOIN с article_images, COUNT комментариев (REQ-03) |
| `static/img/lfc_crest.webp` | static asset | Логотип LFC, отдаётся локально (REQ-01, CON-02) |
| `static/css/main.css` | css | Новые стили для всех виджетов и layout |

### Flow

1. Пользователь открывает `/` (главную страницу).
2. Handler вызывает `HomeStore.GetHomeData()` — получает статьи с `CoverImageURL` и `CommentCount`, таблицу АПЛ, активность форума.
3. Шаблон рендерит: header → левая колонка (новости с hero-image) → правая колонка (матчи-заглушки + таблица АПЛ compact + форум-активность) → footer.
4. Alpine.js управляет состоянием expand таблицы: `x-data="{ expanded: false }"` на контейнере, `x-show` на строках вне top-5.

### Contracts

| Contract ID | Input / Output | Producer / Consumer | Notes |
| --- | --- | --- | --- |
| `CTR-01` | `HomeData.News[].CoverImageURL string` | `HomeStore` → шаблон | Пустая строка если нет изображения — шаблон показывает placeholder |
| `CTR-02` | `HomeData.News[].CommentCount int` | `HomeStore` → шаблон | 0 если нет комментариев |

### Failure Modes

- `FM-01` Статья без hero-изображения: `CoverImageURL == ""` — шаблон показывает CSS-placeholder (div с фоновым цветом клуба), блок изображения не скрывается.
- `FM-02` Изменение `base.html` ломает страницы за пределами главной (/login, /register, /news/{id}, /forum) — проверяется регрессионным Playwright-сценарием.

## Verify

### Exit Criteria

- `EC-01` Главная страница визуально соответствует макету `homeworks/hw-2/`.
- `EC-02` Статьи с изображением показывают hero; статьи без изображения показывают CSS-placeholder.
- `EC-03` Таблица АПЛ: компактный срез (5 команд) отображается по умолчанию; клик разворачивает полную таблицу inline.
- `EC-04` Страницы /news, /forum, /login, /register, /news/{id} рендерятся без визуальных регрессий.

### Traceability matrix

| Requirement ID | Design refs | Acceptance refs | Checks | Evidence IDs |
| --- | --- | --- | --- | --- |
| `REQ-01` | `CON-01`, `CON-02` | `EC-01`, `SC-01` | `CHK-01` | `EVID-01` |
| `REQ-02` | `CTR-01`, `FM-01` | `EC-01`, `EC-02`, `SC-02` | `CHK-02` | `EVID-02` |
| `REQ-03` | `CTR-01`, `CTR-02` | `EC-02`, `SC-02` | `CHK-03` | `EVID-03` |
| `REQ-04` | `ASM-02`, `NT-01` | `EC-03`, `SC-03` | `CHK-04` | `EVID-04` |
| `REQ-05` | `NS-06` | `EC-01`, `SC-01` | `CHK-01` | `EVID-01` |
| `REQ-06` | — | `EC-01`, `SC-01` | `CHK-01` | `EVID-01` |
| `REQ-07` | `CON-01` | `EC-01`, `SC-01` | `CHK-01` | `EVID-01` |

### Acceptance Scenarios

- `SC-01` Пользователь открывает `/` — видит header с логотипом LFC, навигацией и кнопками; footer с 3 колонками; виджеты матчей и активности форума в правой колонке.
- `SC-02` Новостная лента: статья с изображением показывает hero-img; статья без изображения показывает CSS-placeholder. Отображаются excerpt, дата, счётчик комментариев.
- `SC-03` Таблица АПЛ: по умолчанию видны 5 команд (2+LFC+2); клик на таблицу разворачивает полную таблицу inline; повторный клик сворачивает.
- `SC-04` Открытие /news, /forum, /login, /register — header и footer выглядят корректно, layout не сломан.

### Negative Cases

- `NEG-01` LFC на 1-м месте в таблице: compact-срез показывает LFC + 4 команды ниже.
- `NEG-02` LFC на 2-м месте: compact-срез показывает 1 команду выше + LFC + 3 команды ниже.
- `NEG-03` Статья без title или excerpt — страница не падает, поля рендерятся как пустые строки.

### Checks

| Check ID | Covers | How to check | Expected result | Evidence path |
| --- | --- | --- | --- | --- |
| `CHK-01` | `EC-01`, `SC-01`, `SC-04` | Playwright screenshot + ручная сверка с макетом | Главная и все затронутые страницы соответствуют макету без JS-ошибок | `artifacts/ft-021/verify/chk-01/` |
| `CHK-02` | `EC-02`, `SC-02`, `NEG-03` | Playwright: открыть главную, проверить `.news-hero` для статей с и без изображения | С изображением — `<img>`, без — CSS-placeholder div | `artifacts/ft-021/verify/chk-02/` |
| `CHK-03` | `REQ-03` | `docker compose exec db psql` + Go unit-test HomeStore | CoverImageURL и CommentCount заполнены корректно | `artifacts/ft-021/verify/chk-03/` |
| `CHK-04` | `EC-03`, `SC-03`, `NEG-01`, `NEG-02` | Playwright: клик на таблицу АПЛ — проверить expand/collapse; граничные случаи с mock-данными | Expand разворачивает все строки, иконки клубов целы, NT-01 соблюдён | `artifacts/ft-021/verify/chk-04/` |

### Test matrix

| Check ID | Evidence IDs | Evidence path |
| --- | --- | --- |
| `CHK-01` | `EVID-01` | `artifacts/ft-021/verify/chk-01/` |
| `CHK-02` | `EVID-02` | `artifacts/ft-021/verify/chk-02/` |
| `CHK-03` | `EVID-03` | `artifacts/ft-021/verify/chk-03/` |
| `CHK-04` | `EVID-04` | `artifacts/ft-021/verify/chk-04/` |

### Evidence

- `EVID-01` Playwright screenshot главной + затронутых страниц, вывод Playwright-теста (pass).
- `EVID-02` Playwright-скриншоты новостной ленты: статья с hero-img и статья с placeholder.
- `EVID-03` Вывод Go-теста `TestHomeStoreGetHomeData` с CoverImageURL и CommentCount.
- `EVID-04` Playwright-скриншоты таблицы АПЛ: compact и expanded; граничные случаи.
- `EVID-05` Eval Draft→Design Ready — accept. 2026-04-27. self-check.

### Evidence contract

| Evidence ID | Artifact | Producer | Path contract | Reused by checks |
| --- | --- | --- | --- | --- |
| `EVID-01` | Playwright screenshot + test output | Playwright agent | `artifacts/ft-021/verify/chk-01/` | `CHK-01` |
| `EVID-02` | Playwright screenshot (hero + placeholder) | Playwright agent | `artifacts/ft-021/verify/chk-02/` | `CHK-02` |
| `EVID-03` | Go test output | docker run golang:1.23-alpine | `artifacts/ft-021/verify/chk-03/` | `CHK-03` |
| `EVID-04` | Playwright screenshot (compact + expanded + edge cases) | Playwright agent | `artifacts/ft-021/verify/chk-04/` | `CHK-04` |
