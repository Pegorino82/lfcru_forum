---
title: "FT-021: Implementation Plan"
doc_kind: feature
doc_function: derived
purpose: "Execution-план реализации FT-021. Фиксирует discovery context, шаги, риски и test strategy без переопределения canonical feature-фактов."
derived_from:
  - feature.md
status: archived
audience: humans_and_agents
must_not_define:
  - ft_021_scope
  - ft_021_architecture
  - ft_021_acceptance_criteria
  - ft_021_blocker_state
---

# План имплементации

## Цель текущего плана

Переработать визуальный дизайн главной страницы: header, footer, новостная лента с hero-image + excerpt + счётчик комментариев, таблица АПЛ с inline expand (Alpine.js), виджеты матчей и активности форума. Backend: новый тип `HomeNewsItem` в пакете `news`, расширение `LatestPublished`, логика компактного среза standings в handler.

---

## Discovery Context / Reference Points

| Path / module | Current role | Why relevant | Reuse / mirror |
| --- | --- | --- | --- |
| `templates/layouts/base.html` | Глобальный layout: nav, flash, footer, скрипты | Изменение header/footer затронет все страницы — CON-01 | Сохранить структуру skip-link, `<main id="content">`, flash-механизм, подключение HTMX/Alpine/sse.js |
| `templates/home/index.html` | Главная страница: 2-колончатый grid, 5 блоков | Основной шаблон для редизайна | CSS встроен в `<style>` внутри файла — паттерн сохранить |
| `internal/home/handler.go` | `ShowHome`, `HomeData` struct, `FootballSource` interface | Добавить `HomeNewsItem`, `CompactStandings`, `CompactStandingsStart/End` | Паттерн graceful degrade (`footballClient == nil`) и HTMX partial |
| `internal/news/model.go` | `News` struct, `ImageView` struct | `News` уже имеет `Content string` — для excerpt | Создать новый тип `HomeNewsItem` в том же файле рядом с `ImageView` |
| `internal/news/repo.go` | `LatestPublished` — запрашивает только `id, title, published_at` | Нужно добавить CoverImageURL и CommentCount в запрос | Паттерн `rows.Scan` — добавить поля; заменить тип возврата на `[]HomeNewsItem` |
| `internal/football/client.go` | `StandingsEntry` struct с `Position`, `TeamName`, `TeamCrest`, `Points` и др. | Нужен `Position` и `TeamName` для вычисления компактного среза LFC±2 | Struct читается без изменений; логика среза — в handler |
| `migrations/009_create_article_images.sql` | Таблица `article_images (id, article_id, filename, ...)` | JOIN для получения `CoverImageURL` первого изображения | SQL: `LEFT JOIN article_images ai ON ai.article_id = n.id` + `DISTINCT ON` или subquery |
| `migrations/006_create_news_comments.sql` | Таблица `news_comments (id, news_id, ...)` | Subquery для `CommentCount` | SQL: `(SELECT COUNT(*) FROM news_comments WHERE news_id = n.id)` |
| `internal/home/handler_test.go` | Integration tests: `TestHomeHandler_WithData`, `TestHomeHandler_WithStandings` | Нужно обновить под новые поля и compact standings | Паттерн `//go:build integration`, setup с `DATABASE_URL` |
| `homeworks/hw-2/` | Визуальный референс (скриншоты + site.html) | Основной дизайн-референс для шаблонов — **файл не найден** | → см. OQ-01 |
| `static/` | Папка статики (`static/js/sse.js`); `static/img/` не существует | Нужно создать `static/img/` для логотипа LFC | Создать директорию, скачать `lfc_crest.webp` |

---

## Test Strategy

| Test surface | Canonical refs | Existing coverage | Planned automated coverage | Required local suites | Required CI | Manual-only gap / justification | Approval ref |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `internal/news/repo.go` — `LatestPublishedForHome` | `REQ-03`, `CHK-03` | `LatestPublished` покрыт в `repo_test.go` | Добавить/обновить тест: статья с изображением → `CoverImageURL` заполнен; статья без → пустая строка; `CommentCount` равен реальному COUNT | `docker compose run --rm test` (integration) | нет CI | — | none |
| `internal/home/handler.go` — compact standings logic | `REQ-04`, `SC-03`, `NEG-01`, `NEG-02`, `CHK-04` | `TestHomeHandler_WithStandings` | Добавить unit-тест функции `compactStandingsSlice(standings, lfcName)` для позиций: 1-е место, 2-е место, середина таблицы, 19-е, 20-е | `docker compose run --rm test` | нет CI | — | none |
| `templates/home/index.html` + `base.html` — визуальный рендер | `REQ-01`, `REQ-02`, `REQ-04`–`REQ-07`, `SC-01`–`SC-04`, `CHK-01`, `CHK-02`, `CHK-04` | Нет Playwright-тестов | Playwright: screenshot главной, assertions на `.news-hero`/placeholder, expand таблицы, header, footer; smoke-проверка /news, /forum, /login, /register | `docker compose up -d && npx playwright test` | нет CI | Ручная сверка с макетом hw-2 (если OQ-01 снят) | AG-01 |

---

## Open Questions / Ambiguities

| OQ ID | Question | Why unresolved | Blocks | Default action / escalation owner |
| --- | --- | --- | --- | --- |
| `OQ-01` | `homeworks/hw-2/` (скриншоты + site.html) не найден в репозитории | Папка отсутствует при grounding: есть только hw-0 и hw-1 | WS-3 (шаблоны) — без визуального референса невозможно точно реализовать новый дизайн | Блокировать реализацию шаблонов до получения макета; эскалация к владельцу задачи |
| `OQ-02` | Источник excerpt для новостей: обрезать `Content` (HTML) или использовать отдельное `excerpt`-поле? | `Content` содержит HTML — обрезка может сломать теги; отдельного поля в схеме нет | `STEP-02`, `STEP-03` — влияет на SQL-запрос и шаблон | По умолчанию: брать `SUBSTRING(content, 1, 200)` + sanitize или `LEFT(content, 200)`; если нужно чистое поле — эскалация к владельцу (потребует миграцию) |

---

## Environment Contract

| Area | Contract | Used by | Failure symptom |
| --- | --- | --- | --- |
| setup | `docker compose up -d` (postgres + app); `goose up` для миграций | STEP-01–04, CHK-03 | Ошибки подключения к БД при тестах |
| test | `docker compose run --rm test` — integration tests с `//go:build integration` | CHK-03, CHK-04 | Test build fails без флага `-tags integration` |
| go build | `docker run --rm -v $(pwd):/app -w /app golang:1.23-alpine go build ./...` | STEP-01–04 | Go не установлен на хосте — все go-команды через Docker |
| playwright | `docker compose up -d && npx playwright test` (Playwright установлен или через npx) | CHK-01, CHK-02, CHK-04 | App не запущен или порт недоступен |
| static assets | `static/img/lfc_crest.webp` доступен по HTTP через `/static/img/lfc_crest.webp` | STEP-05, CHK-01 | 404 на логотип в шаблоне |

---

## Preconditions

| PRE ID | Canonical ref | Required state | Used by steps | Blocks start |
| --- | --- | --- | --- | --- |
| `PRE-01` | `ASM-02` | FT-020 смержен в main; `Standings []football.StandingsEntry` доступен в HomeData | STEP-04 | no (данные уже есть) |
| `PRE-02` | `CON-02` | `static/img/` создана, `lfc_crest.webp` скачан локально | STEP-06 | yes для STEP-06 |
| `PRE-03` | `OQ-01` | `homeworks/hw-2/` получен от владельца задачи | STEP-06 (шаблоны) | **yes** — блокирует визуальную реализацию |

---

## Workstreams

| Workstream | Implements | Result | Owner | Dependencies |
| --- | --- | --- | --- | --- |
| `WS-1` Backend | `REQ-03` | Новый тип `HomeNewsItem`, обновлённый `LatestPublishedForHome`, compact standings в handler | agent | PRE-01 |
| `WS-2` Static assets | `REQ-01`, `CON-02` | `static/img/lfc_crest.webp` скачан и отдаётся по `/static/img/lfc_crest.webp` | agent | — |
| `WS-3` Templates | `REQ-01`, `REQ-02`, `REQ-04`–`REQ-07` | Обновлены `base.html` и `home/index.html` | agent | PRE-03 (OQ-01), WS-1 |
| `WS-4` Tests | `CHK-01`–`CHK-04` | Integration + Playwright тесты зелёные | agent | WS-1, WS-3 |

---

## Approval Gates

| AG ID | Trigger | Applies to | Why approval required | Approver / evidence |
| --- | --- | --- | --- | --- |
| `AG-01` | WS-3 завершён — шаблоны обновлены | STEP-06, STEP-07 | Визуальная проверка с макетом hw-2 не автоматизируется полностью; human должен сверить скриншот Playwright с оригиналом | Владелец задачи; evidence: Playwright screenshot в `EVID-01` |
| `AG-02` | OQ-01 требует получения hw-2 от владельца | WS-3 целиком | Без визуального референса невозможно корректно реализовать дизайн | Владелец задачи; evidence: папка `homeworks/hw-2/` в репозитории |

---

## Порядок работ

| Step ID | Actor | Implements | Goal | Touchpoints | Artifact | Verifies | Evidence IDs | Check command / procedure | Blocked by | Needs approval | Escalate if |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `STEP-01` | agent | `REQ-03` | Создать `HomeNewsItem` struct в `internal/news/model.go` с полями `ID, Title, CoverImageURL, CommentCount, PublishedAt` и функцию/метод `ExcerptText() string` (обрезка Content до 200 символов) | `internal/news/model.go` | Новый тип `HomeNewsItem` | `CHK-03` | `EVID-03` | `docker run ... go build ./internal/news/...` — компилируется | — | none | Если Content содержит сложный HTML — эскалация к OQ-02 |
| `STEP-02` | agent | `REQ-03` | Добавить метод `LatestPublishedForHome(ctx, limit)` в `internal/news/repo.go`: SQL с LEFT JOIN article_images (первое изображение, DISTINCT ON article_id) и subquery COUNT(news_comments) | `internal/news/repo.go` | Новый метод, расширенный SQL | `CHK-03` | `EVID-03` | `docker run ... go build ./internal/news/...` | STEP-01 | none | Если JOIN выдаёт N×M строк — пересмотреть на subquery |
| `STEP-03` | agent | `REQ-03` | Обновить `HomeData` в `internal/home/handler.go`: заменить `News []news.News` на `News []news.HomeNewsItem`, вызвать `LatestPublishedForHome` | `internal/home/handler.go` | Обновлённый `HomeData`, вызов нового метода | `CHK-03` | `EVID-03` | `docker run ... go build ./...` | STEP-02 | none | Если компиляция ломает downstream (news/article.html, admin) — проверить, что `news.News` не затронут |
| `STEP-04` | agent | `REQ-04` | Добавить в handler логику компактного среза standings: найти позицию LFC по `TeamName`, вычислить `[start, end)` индексы (2+LFC+2 с граничными случаями), добавить `CompactStandingsStart/End int` в `HomeData` | `internal/home/handler.go` | `compactStandingsRange()` хелпер + поля в `HomeData` | `CHK-04` | `EVID-04` | Unit-test `compactStandingsRange` (dockerized): позиции 1, 2, 10, 19, 20 | STEP-03 | none | Если LFC не найден в standings — пустой compact (graceful degrade) |
| `STEP-05` | agent | `REQ-01`, `CON-02` | Создать `static/img/`, скачать `https://www.liverpoolfc.com/liverpoolfc_crest.webp` в `static/img/lfc_crest.webp` | `static/img/lfc_crest.webp` | Файл логотипа | `CHK-01` | `EVID-01` | `curl -I /static/img/lfc_crest.webp` возвращает 200 (после запуска app) | PRE-02 | none | Если URL недоступен — скачать альтернативный источник или использовать SVG-заглушку |
| `STEP-06` | agent | `REQ-01`–`REQ-02`, `REQ-05`–`REQ-07` | Обновить `templates/layouts/base.html`: новый header (логотип `lfc_crest.webp`, nav «Новости/Форум» слева, кнопки auth справа), новый footer (3 колонки + соцсети) | `templates/layouts/base.html` | Обновлённый base.html | `CHK-01`, `SC-01`, `SC-04` | `EVID-01` | Playwright: screenshot `/`, `/news`, `/login` — header и footer рендерятся без JS-ошибок | PRE-03 (AG-02), STEP-05 | AG-01 | Если /login или /forum ломаются визуально — откат base.html |
| `STEP-07` | agent | `REQ-02`, `REQ-04`, `REQ-05`, `REQ-06` | Обновить `templates/home/index.html`: новостная лента (hero-image + placeholder FM-01, excerpt, дата, счётчик), таблица АПЛ (compact x-show + full expand, Alpine x-data), виджеты матчей (новый layout), активность форума (тег-бейдж) | `templates/home/index.html` | Обновлённый index.html | `CHK-01`, `CHK-02`, `CHK-04`, `SC-02`, `SC-03` | `EVID-01`, `EVID-02`, `EVID-04` | Playwright: expand table, `.news-hero` vs placeholder, форум-бейдж | STEP-03, STEP-04, STEP-06 | AG-01 | Alpine не реагирует на клик → проверить порядок загрузки Alpine (см. memory feedback) |
| `STEP-08` | agent | `CHK-03` | Обновить `internal/home/handler_test.go` и `internal/news/repo_test.go` под новые типы и поля | `internal/home/handler_test.go`, `internal/news/repo_test.go` | Зелёные integration tests | `CHK-03` | `EVID-03` | `docker compose run --rm test -run TestHome` | STEP-03, STEP-04 | none | Если тест не находит `HomeNewsItem` — проверить импорты |
| `STEP-09` | agent | `CHK-01`, `CHK-02`, `CHK-04` | Написать Playwright-тесты: главная (screenshot, assertions), news-hero/placeholder, expand таблицы, smoke /news /forum /login /register | Playwright test file | Playwright test suite | `CHK-01`, `CHK-02`, `CHK-04`, `EC-01`–`EC-04` | `EVID-01`, `EVID-02`, `EVID-04` | `npx playwright test --reporter=list` | STEP-07 | none | Тест падает → это блокер closure gate (autonomy-boundaries) |

---

## Parallelizable Work

- `PAR-01` WS-2 (STEP-05, скачать логотип) может идти параллельно с WS-1 (STEP-01–04) — нет общего write-surface.
- `PAR-02` STEP-06 и STEP-07 нельзя распараллеливать: оба пишут в разные шаблоны, но STEP-07 зависит от STEP-06 (base.html).
- `PAR-03` STEP-08 и STEP-09 можно начинать параллельно после STEP-07, если тест-среда уже поднята.

---

## Checkpoints

| CP ID | Refs | Condition | Evidence IDs |
| --- | --- | --- | --- |
| `CP-01` | STEP-01–03 | Backend компилируется; `LatestPublishedForHome` возвращает `HomeNewsItem` с `CoverImageURL` и `CommentCount` | `EVID-03` |
| `CP-02` | STEP-04, STEP-08 | `compactStandingsRange` unit-тест зелёный для всех граничных случаев; handler integration tests зелёные | `EVID-03`, `EVID-04` |
| `CP-03` | STEP-06–07, AG-01 | Главная рендерится, expand таблицы работает, base.html не ломает /login и /forum; human визуально подтвердил соответствие макету | `EVID-01` |
| `CP-04` | STEP-09 | Все Playwright-тесты зелёные; evidence artifacts зафиксированы | `EVID-01`, `EVID-02`, `EVID-04` |

---

## Execution Risks

| ER ID | Risk | Impact | Mitigation | Trigger |
| --- | --- | --- | --- | --- |
| `ER-01` | `homeworks/hw-2/` не получен до начала WS-3 | WS-3 заблокирован полностью; визуальные требования не определены | AG-02: остановиться и запросить hw-2 у владельца | OQ-01 не снят до старта STEP-06 |
| `ER-02` | `base.html` — глобальный: ошибка header/footer ломает `/login`, `/register`, `/forum` | Регрессия на всех страницах | STEP-06 включает Playwright smoke-check всех затронутых страниц до коммита | Playwright assertions fail на любой из затронутых страниц |
| `ER-03` | Alpine.js `x-show` не реагирует на клик после HTMX swap | Expand таблицы не работает | Сбрасывать Alpine-state в `htmx:before-swap` (зафиксированный feedback в memory) | JS-ошибка в консоли или Playwright assertion fail на expand |
| `ER-04` | `LEFT JOIN article_images` возвращает дубли (N статей × M изображений) | `HomeNewsItem` с дублированными записями | Использовать `DISTINCT ON (n.id)` или subquery | Тест показывает > 5 новостей при limit=5 |
| `ER-05` | Content (HTML) при обрезке до 200 символов оставляет незакрытые теги | Сломанная вёрстка в excerpt | Стриппинг HTML-тегов в Go перед обрезкой или использование `strings.TrimSpace(strings.ReplaceAll(content, html_tags, ""))` | OQ-02 эскалирован |

---

## Stop Conditions / Fallback

| STOP ID | Related refs | Trigger | Immediate action | Safe fallback state |
| --- | --- | --- | --- | --- |
| `STOP-01` | `OQ-01`, `AG-02` | `homeworks/hw-2/` не получен к старту STEP-06 | Остановить WS-3, зафиксировать блокер в PR, уведомить владельца | WS-1 и WS-2 завершены, WS-3 ожидает макета |
| `STOP-02` | `ER-02`, `CON-01` | Playwright smoke на `/login` или `/forum` падает после STEP-06 | Откатить `base.html` до предыдущего коммита, проанализировать причину | base.html в исходном состоянии; feature PR не содержит сломанный шаблон |
| `STOP-03` | `ER-03` | Alpine expand не работает после 2 итераций исправлений | Эскалировать к human: использовать HTMX вместо Alpine или другой механизм | Таблица рендерится в полном виде без compact/expand (degraded mode) |

---

## Готово для приемки

- CP-01 пройден: backend компилируется, новые поля заполнены
- CP-02 пройден: unit + integration тесты зелёные
- CP-03 пройден: визуальное подтверждение по макету (AG-01)
- CP-04 пройден: Playwright-тесты зелёные, evidence artifacts зафиксированы
- `feature.md` → `delivery_status: in_progress`
- Все условия из gate Plan Ready → Execution в `feature-flow.md` выполнены
- `EVID-06: Eval DR→PR — accept. 2026-04-27. evaluator agent.`
