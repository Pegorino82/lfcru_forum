---
title: "FT-023: Implementation Plan"
doc_kind: feature
doc_function: derived
purpose: "Execution-план реализации FT-023. Фиксирует discovery context, шаги, риски и test strategy без переопределения canonical feature-фактов."
derived_from:
  - feature.md
status: draft
audience: humans_and_agents
must_not_define:
  - ft_023_scope
  - ft_023_architecture
  - ft_023_acceptance_criteria
---

# План имплементации FT-023

## Цель текущего плана

Заменить `<textarea>` в форме редактирования статьи на TipTap WYSIWYG-редактор (vanilla JS, ESM CDN). Добавить bluemonday-санитизацию в handler сохранения. Переключить рендеринг статьи с `RenderMarkdown` на `template.HTML` напрямую.

---

## Discovery Context / Reference Points

| Path / module | Current role | Why relevant | Reuse / mirror |
| --- | --- | --- | --- |
| `internal/admin/articles_handler.go` | `Create()`, `Update()` — читают `c.FormValue("content")` → `news.News{Content: content}` → repo | Основная точка добавления bluemonday-санитизации | Паттерн `c.FormValue` → struct → repo; sanitize после `FormValue`, до передачи в repo |
| `internal/news/model.go` | `News.Content string` — поле тела статьи | **OQ-01**: feature.md называет `articles.body`; реально — таблица `news`, колонка `content`, Go-поле `News.Content` | Использовать реальное имя `Content` во всех шагах |
| `internal/news/repo.go` | `CreateDraft`, `UpdateArticle` — SQL с колонкой `content` | Изменений не требует — хранит строку as-is | Читать без изменений |
| `internal/news/handler.go` | `ShowArticle` — `ContentHTML: RenderMarkdown(article.Content)` | Нужно заменить на `ContentHTML: template.HTML(article.Content)` | `ArticleData.ContentHTML` уже `template.HTML` — только поменять источник |
| `internal/news/markdown.go` | `RenderMarkdown(content string) template.HTML` | Используется в `ShowArticle` и admin `Preview`; после FT-023 убирается из обоих | Функция остаётся — может понадобиться для будущей миграции |
| `internal/admin/articles_handler.go` (`Preview`) | `ContentHTML: news.RenderMarkdown(article.Content)` | Аналогичная замена — preview должен рендерить HTML как есть | Паттерн тот же, что и `ShowArticle` |
| `internal/admin/articles_handler.go` (`ChangeStatus`) | Проверяет `article.Content == ""` перед публикацией | **OQ-02**: после TipTap пустой редактор отдаёт `<p></p>` — строка непустая, проверка пройдёт. Реального бага нет; пустая статья публикуется с `<p></p>`. Нет блокера. | Документировать как known behavior; не менять проверку |
| `templates/admin/articles/edit.html` | `<textarea id="content" name="content" rows="20">` + inline `<style>` | Основной шаблон для замены: textarea → TipTap-контейнер + тулбар; HTMX image upload уже есть | Сохранить паттерн inline `<style>`, HTMX image upload без изменений; добавить hidden input `name="content"` |
| `templates/news/article.html` | `{{.ContentHTML}}` — рендерит `template.HTML` | Изменений не требует — уже ожидает `template.HTML` | Читать без изменений |
| `static/js/sse.js` | Единственный JS-файл в `static/js/` | Паттерн подключения JS-файла | Создать `static/js/editor.js` рядом |
| `internal/admin/articles_handler_test.go` | Integration tests (`//go:build integration`), `newArticlesServer`, паттерн POST через httptest | Использовать для CHK-03 (XSS-санитизация) | Паттерн: `newArticlesServer` + `http.NewRequest` + `form.Values` |
| `e2e/global-setup.ts` | Создаёт `e2e_user` с ролью `user` | **OQ-03**: для E2E тестов редактора нужен admin-пользователь; текущий setup не создаёт его | Добавить `e2e_admin` с ролью `admin` в global-setup |
| `e2e/forum/reply.spec.ts` | Playwright E2E паттерн: `page.goto`, `page.fill`, `page.click`, `expect(page.locator(...))` | Референс для написания CHK-01/CHK-02 тестов | Зеркалировать паттерн |
| `go.mod` | `github.com/yuin/goldmark v1.8.2` — единственная зависимость для markdown | bluemonday ещё не добавлен | Добавить `github.com/microcosm-cc/bluemonday` |

---

## Test Strategy

| Test surface | Canonical refs | Existing coverage | Planned automated coverage | Required local suites | Required CI | Manual-only gap |
| --- | --- | --- | --- | --- | --- | --- |
| `articles_handler.go` — sanitize в `Create`/`Update` | `REQ-04`, `CHK-03`, `SC-03`, `NEG-01` | POST/update покрыты в `articles_handler_test.go` | Добавить тест: POST с XSS payload в `content` → GET → проверить отсутствие `<script>` и `<iframe>` в body ответа | Unit (`go test ./...`) | Go Tests CI (integration) | — |
| `static/js/editor.js` + `edit.html` — TipTap инициализация | `REQ-01`, `REQ-02`, `REQ-03`, `CHK-01`, `CHK-02` | Нет | Playwright E2E: открыть `/admin/articles/{id}/edit`, кликнуть по кнопкам тулбара, проверить DOM-элементы в редакторе и в view | — | E2E CI (Playwright) | AG-01: ручная проверка TipTap в браузере (localhost:8081) перед первым пушем |
| `news/handler.go` — рендеринг HTML | `REQ-04`, `CTR-01` | `handler_test.go` покрывает роутинг | Добавить тест: статья с HTML body → GET `/news/{id}` → `ContentHTML` содержит ожидаемый HTML | Unit (`go test ./...`) | Go Tests CI (integration) | — |

---

## Open Questions / Ambiguities

| OQ ID | Question | Why unresolved | Blocks | Default action |
| --- | --- | --- | --- | --- |
| `OQ-01` | `feature.md` называет `articles.body`; реально — таблица `news`, колонка `content`, Go-поле `News.Content` | Терминологическое расхождение выявлено при grounding | Именование в коде и тестах | Использовать `Content` / `news.content` во всём плане; feature.md не менять — расхождение косметическое |
| `OQ-02` | Пустой TipTap отдаёт `<p></p>` — `ChangeStatus` проверка `article.Content == ""` не срабатывает | Выявлено при grounding; не является блокером | STEP-02 | Known behavior; пустая статья с `<p></p>` технически публикуется. Не менять проверку без явного запроса |
| `OQ-03` | E2E global-setup создаёт только `user`-роль; редактор доступен только `admin` | Выявлено при grounding | STEP-07, STEP-08 | Добавить `e2e_admin` в `global-setup.ts` в рамках STEP-07 |
| `OQ-04` | TipTap ESM CDN доступность: приложение работает в Docker; CDN-скрипты загружаются браузером (не сервером) — конфликта нет | Гипотетическая проблема; CDN = client-side | — | ESM CDN загружается браузером из сети; Docker-контейнер не блокирует. Проверить на STEP-05 при первом запуске |
| `OQ-05` | `architecture.md` не обновлён: `news.content` хранит HTML вместо Markdown — ожидает перевода ADR-007 в `accepted` | ADR-007 ещё `proposed` | — | Обновить `architecture.md` после перевода ADR-007 в `accepted`; оформить как follow-up в PR |
| `OQ-06` | Sanitize в Handler — намеренное отклонение от layer stack (`architecture.md`: domain-логика → Service) | `internal/admin` не имеет Service-слоя; articles_handler.go вызывает repo напрямую; введение Service-слоя вне scope FT-023 | STEP-02 | Принято как обоснованное отклонение: sanitize в handler, пока Service-слой не введён; зафиксировано здесь для трассируемости |
| `OQ-07` | `feature.md` описывает рендеринг как `{{ .Body | safeHTML }}` (шаблонная функция), план реализует через `ContentHTML: template.HTML(article.Content)` в Go-хендлере | Существующий паттерн кодовой базы использует `ContentHTML template.HTML` в struct — добавление `safeHTML` template func потребует изменения engine и всех шаблонов | STEP-03 | Реализовать через `template.HTML(article.Content)` в handler (соответствует текущему паттерну); feature.md не менять — это реализационная деталь, не scope |

---

## Environment Contract

| Area | Contract | Used by | Failure symptom |
| --- | --- | --- | --- |
| Go build | `docker run --rm -v $(pwd):/app -w /app -v lfcru_gomod:/root/go/pkg/mod golang:1.23-alpine go build ./...` | Все STEP с Go | Go не установлен на хосте |
| Go deps | `docker run ... golang:1.23-alpine go mod download` после изменения go.mod | PRE-01 | Ошибки импорта bluemonday |
| Integration tests | `docker run --rm -v "$(pwd)":/app -w /app -v lfcru_gomod:/root/go/pkg/mod --network lfcru_forum_default -e DATABASE_URL="postgres://postgres:postgres@postgres:5432/lfcru_test?sslmode=disable" golang:1.23-alpine go test -tags integration -p 1 ./internal/...` (требует `docker compose -f docker-compose.dev.yml up -d`) | CHK-03 (только CI) | `no such host: postgres` — контейнер без `--network` |
| E2E | `docker compose -f docker-compose.e2e.yml up -d` + `npx playwright test` | CHK-01, CHK-02 | App недоступен на localhost:8081 |
| Static assets | `static/js/editor.js` раздаётся по `/static/js/editor.js` через Echo static middleware | STEP-05 | 404 на скрипт редактора |

---

## Preconditions

| PRE ID | Canonical ref | Required state | Used by steps | Blocks start |
| --- | --- | --- | --- | --- |
| `PRE-01` | `ASM-02`, `DEC-01` | `github.com/microcosm-cc/bluemonday` добавлен в go.mod и go.sum | STEP-01 | yes |
| `PRE-02` | `ASM-01` | Upload endpoint `/admin/articles/:id/images` работает (проверить HTMX upload в edit.html) | STEP-06 | no (уже реализовано в FT-009) |

---

## Workstreams

| Workstream | Implements | Result | Dependencies |
| --- | --- | --- | --- |
| `WS-1` Backend | `REQ-04`, `CTR-01`, `CTR-02` | bluemonday в handler, рендеринг через `template.HTML` | PRE-01 |
| `WS-2` Frontend | `REQ-01`, `REQ-02`, `REQ-03` | TipTap в edit.html, editor.js | WS-1 (hidden input pattern) |
| `WS-3` Tests | `CHK-01`, `CHK-02`, `CHK-03` | Integration + E2E тесты зелёные | WS-1, WS-2 |

---

## Порядок работ

| Step ID | Implements | Goal | Touchpoints | Blocked by | Escalate if |
| --- | --- | --- | --- | --- | --- |
| `STEP-01` | `PRE-01` | Добавить bluemonday в go.mod: `docker run ... go get github.com/microcosm-cc/bluemonday` | `go.mod`, `go.sum` | — | go get падает → проверить сеть, использовать GOPROXY |
| `STEP-02` | `REQ-04`, `CTR-02` | В `articles_handler.go` создать `sanitizeArticleBody(s string) string` с bluemonday allowlist (p, h1-h3, strong, em, s, a href, img src alt, figure, figcaption, div style text-align, br); применить в `Create()` и `Update()` после `c.FormValue("content")` | `internal/admin/articles_handler.go` | STEP-01 | bluemonday strippt нужные теги → пересмотреть allowlist |
| `STEP-03` | `REQ-04`, `CTR-01` | В `news/handler.go` `ShowArticle`: заменить `ContentHTML: RenderMarkdown(article.Content)` на `ContentHTML: template.HTML(article.Content)` | `internal/news/handler.go` | STEP-02 | Рендеринг ломается → sanitization не применена на сохранении (STEP-02 не выполнен) |
| `STEP-04` | `REQ-04` | В `admin/articles_handler.go` `Preview`: аналогичная замена `news.RenderMarkdown` → `template.HTML` | `internal/admin/articles_handler.go` | STEP-02 | — |
| `STEP-05` | `REQ-01`, `REQ-02`, `REQ-03` | Заменить `<textarea>` в `edit.html` на: TipTap-контейнер `<div id="editor">`, `<input type="hidden" name="content" id="content-input">`, тулбар с кнопками (bold, italic, strike, h1/h2/h3, link, align-left/center/right, image-upload); подключить `static/js/editor.js` через `<script type="module">` | `templates/admin/articles/edit.html` | STEP-02 | TipTap не инициализируется → проверить ESM CDN URL и консоль браузера |
| `STEP-06` | `REQ-01`, `REQ-02`, `REQ-03` | Создать `static/js/editor.js`: импорт TipTap extensions через ESM CDN (jsDelivr), инициализация Editor с extensions (StarterKit, TextAlign, Link, Image), sync `editor.getHTML()` → `#content-input` перед submit формы, обработчик кнопок тулбара, обработчик image-upload (FormData POST на HTMX endpoint → URL → `editor.commands.setImage`) | `static/js/editor.js` (new) | STEP-05 | HTMX upload и TipTap insert не согласованы → проверить порядок событий (ASM-01) |
| `STEP-07` | `CHK-03`, `SC-03`, `NEG-01` | Добавить `TestArticlesHandler_XSSSanitization` в `articles_handler_test.go`: create article → update с `<script>alert(1)</script><iframe>` в content → GET admin view → assert отсутствие `<script>` и `<iframe>` в ответе. Добавить `e2e_admin` (role=admin) в `e2e/global-setup.ts` | `internal/admin/articles_handler_test.go`, `e2e/global-setup.ts` | STEP-02 | Тест красный → sanitization не работает, проверить allowlist |
| `STEP-08` | `CHK-01`, `SC-01`, `MET-01` | Написать Playwright E2E тест: войти как admin, открыть `/admin/articles/{id}/edit`, кликнуть bold/italic/strike/h2/link/align-center → сохранить → открыть `/news/{id}` → assert `<strong>`, `<em>`, `<s>`, `<h2>`, `<a>`, `style="text-align:center"` в DOM | `e2e/` (new spec file) | STEP-05, STEP-06, STEP-07 | Playwright не может взаимодействовать с TipTap → проверить селекторы ProseMirror `.ProseMirror` |
| `STEP-09` | `CHK-02`, `SC-02` | Написать Playwright E2E тест: войти как admin, открыть редактор → кликнуть image-upload → загрузить тестовый файл → ввести caption → сохранить → открыть view → assert `<figure>`, `<img>`, `<figcaption>` в DOM | `e2e/` (продолжение spec файла) | STEP-08 | Upload не вставляет image в TipTap → проверить STEP-06 upload handler |

---

## Parallelizable Work

- `PAR-01` STEP-03 и STEP-04 можно выполнять параллельно — разные файлы, нет shared write-surface.
- `PAR-02` STEP-05 и STEP-06 выполняются последовательно: шаблон подключает скрипт, скрипт пишется под шаблон.
- `PAR-03` STEP-08 и STEP-09 — один spec-файл, последовательно.

---

## Approval Gates

| AG ID | Trigger | Why approval required | Approver / procedure |
| --- | --- | --- | --- |
| `AG-01` | STEP-06 завершён (editor.js создан, шаблон обновлён) | TipTap-взаимодействия (contenteditable, тулбар, upload) не покрываются автоматически до выхода E2E — нужно убедиться, что редактор инициализируется и тулбар кликабелен | Human: открыть `localhost:8081/admin/articles/{id}/edit`, кликнуть bold/italic/image, убедиться в отсутствии JS-ошибок в консоли; дать "ок" перед STEP-08 |

---

## Checkpoints

| CP ID | Refs | Condition | Evidence |
| --- | --- | --- | --- |
| `CP-01` | STEP-01–04 | `docker run ... go build ./...` — зелёный; bluemonday применяется, рендеринг через `template.HTML` | `EVID-03` (build log) |
| `CP-02` | STEP-05–06, AG-01 | TipTap инициализируется в браузере, тулбар работает, submit формы отправляет HTML в `content`; AG-01 подтверждён человеком | AG-01 approval ref |
| `CP-03` | STEP-07 | Integration тест CHK-03 зелёный: XSS payload удалён | `EVID-03` |
| `CP-04` | STEP-08–09 | E2E тесты CHK-01/CHK-02 зелёные; артефакты в `artifacts/ft-023/verify/` | `EVID-01`, `EVID-02` |

---

## Execution Risks

| ER ID | Risk | Impact | Mitigation |
| --- | --- | --- | --- |
| `ER-01` | TipTap ESM CDN недоступен в браузере (CSP или сеть) | TipTap не инициализируется | Проверить консоль на STEP-05; fallback — скачать bundle локально в `static/js/tiptap/` |
| `ER-02` | bluemonday allowlist обрезает нужные теги | Форматирование теряется при сохранении | Проверить allowlist на тестовой статье на STEP-07; расширять только явно |
| `ER-03` | TipTap selectors в Playwright (`contenteditable`, `.ProseMirror`) нестандартны | E2E тест не может вводить текст | Использовать `page.locator('.ProseMirror').click()` + `page.keyboard.type()` |
| `ER-04` | HTMX image upload response возвращает HTML-фрагмент (`image_item.html`), а не JSON с URL | TipTap не может получить URL из ответа | На STEP-06: парсить URL из возвращённого img-тега или добавить `data-url` атрибут в `image_item.html` |

---

## Stop Conditions

| STOP ID | Trigger | Immediate action |
| --- | --- | --- |
| `STOP-01` | `ER-01` не устранён за 2 итерации | Скачать TipTap bundle локально; эскалировать если нет npm |
| `STOP-02` | E2E тест CHK-01 или CHK-02 красный после 3 итераций | Остановить, зафиксировать блокер, эскалировать к human |

---

## Готово для приёмки

- CP-01: backend компилируется, bluemonday применяется
- CP-02: TipTap работает в браузере
- CP-03: integration тест CHK-03 зелёный
- CP-04: E2E тесты CHK-01/CHK-02 зелёные, artifacts зафиксированы
- `feature.md` → `delivery_status: in_progress` (перед началом execution)
