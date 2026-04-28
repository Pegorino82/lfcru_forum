---
title: "FT-023: WYSIWYG-редактор статей"
doc_kind: feature
doc_function: canonical
purpose: "Заменить plain-text textarea в форме редактирования статьи на WYSIWYG-редактор (TipTap) с форматированием, вставкой изображений и выравниванием. Тело статьи хранится как HTML."
derived_from:
  - ../../domain/problem.md
  - ../../adr/ADR-007-wysiwyg-editor-html-storage.md
  - ../../use-cases/UC-001-article-publishing.md
status: active
delivery_status: planned
audience: humans_and_agents
must_not_define:
  - implementation_sequence
---

# FT-023: WYSIWYG-редактор статей

## What

### Problem

Редактор статей — plain-text textarea с Markdown. Форматирование применяется только через синтаксис (`**bold**`), без визуального feedback. Изображения прикрепляются только в конец статьи, без подписей. Выравнивание текста недоступно. Авторы не знают, как пользоваться Markdown, и вынуждены разбираться с синтаксисом вместо написания контента.

### Outcome

| Metric ID | Metric | Baseline | Target | Measurement method |
| --- | --- | --- | --- | --- |
| `MET-01` | Поддержка всех formatting actions из acceptance criteria | 0 из 7 | 7 из 7 | Playwright E2E (CHK-01) |
| `MET-02` | XSS-векторы в rendered HTML | не проверяется | 0 (все sanitized) | CHK-03 |

### Scope

- `REQ-01` WYSIWYG-редактор с тулбаром: bold, italic, strikethrough, заголовки h1/h2/h3, ссылки, выравнивание текста (left/center/right).
- `REQ-02` Вставка изображения в произвольную позицию в тексте: кнопка в тулбаре открывает file picker → файл загружается на сервер → URL вставляется как inline-изображение в позицию курсора.
- `REQ-03` Изображение может иметь подпись (alt/caption), редактируемую inline.
- `REQ-04` Тело статьи сохраняется в БД как HTML-строка; бэкенд санитизирует HTML через allowlist-политику (bluemonday) при сохранении и при рендеринге.

### Non-Scope

- `NS-01` Drag & drop загрузка изображений — вне scope данной фичи.
- `NS-02` Plain-text Markdown-редактирование — сознательно исключено (ADR-007, Вариант B).
- `NS-03` Миграция существующих статей с Markdown на HTML — отдельная задача; текущие статьи обрабатываются вне scope данной фичи.
- `NS-04` Предпросмотр статьи в отдельной вкладке/модале.
- `NS-05` Коллаборативное редактирование (real-time).

### Constraints / Assumptions

- `ASM-01` Upload-endpoint для изображений уже реализован (ADR-005, FT-009); REQ-02 использует его без изменений.
- `ASM-02` TipTap подключается через ESM CDN или npm+esbuild; способ сборки определяется в implementation-plan.md по результатам grounding текущего JS-стека.
- `ASM-03` Существующие статьи с Markdown-телом рендерятся as-is — без конвертации; визуально некорректное отображение таких статей является ожидаемым результатом до выполнения отдельной миграционной задачи.
- `ASM-04` CSRF-защита обеспечена существующим middleware (PCON-02); FT-023 не меняет CSRF-механизм — POST/PUT endpoint сохранения статьи уже защищён.
- `CON-01` XSS-санитизация обязательна: тело статьи рендерится как `template.HTML` — без санитизации это XSS-вектор. Bluemonday allowlist должен разрешать только теги, необходимые для форматирования из REQ-01/REQ-02/REQ-03.
- `DEC-01` ADR-007 имеет статус `proposed` на момент старта; выполнение опирается на Вариант B как hypothesis. Если ADR-007 будет отклонён — весь scope FT-023 подлежит пересмотру.

## How

### Solution

Заменяем `<textarea>` в форме редактирования статьи на TipTap-редактор (vanilla JS, headless). TipTap генерирует HTML, который отправляется в существующий POST/PUT endpoint. Бэкенд (Go) пропускает `body` через bluemonday перед записью в PostgreSQL. Рендеринг статьи — `{{ .Body | safeHTML }}` вместо Markdown-рендерера. Изображения загружаются через существующий upload-endpoint (ADR-005), URL вставляется как TipTap Image node.

### Change Surface

| Surface | Type | Why it changes |
| --- | --- | --- |
| `templates/admin/articles/edit.html` | code | Заменить `<textarea>` на TipTap-контейнер + тулбар |
| `static/js/editor.js` (new) | code | Инициализация TipTap, тулбар, upload-интеграция |
| `internal/handler/article.go` | code | Добавить bluemonday-санитизацию поля body при save |
| `templates/news/article.html` | code | Рендеринг body через `safeHTML` вместо Markdown-рендерера |
| `go.mod` / `go.sum` | config | Добавить зависимость bluemonday |
| `templates/admin/articles/edit.html` (инлайн-стили) | code | Стили для TipTap-контента и тулбара — инлайн внутри шаблона редактора |
| `memory-bank/use-cases/UC-001-article-publishing.md` | docs | Добавить FT-023 в `Implemented by` — до closure gate |

### Flow

1. Автор открывает форму редактирования статьи — TipTap инициализируется на месте textarea, загружает существующий body (если HTML — как есть; если Markdown — рендерится as-is согласно ASM-03).
2. Автор форматирует текст через тулбар или выделяет текст и нажимает кнопку форматирования.
3. Для вставки изображения: кнопка → file picker → файл POST-ится на upload-endpoint → получен URL → TipTap вставляет Image node с caption в позицию курсора.
4. Автор нажимает «Сохранить» → TipTap.getHTML() сериализуется → отправляется в существующий PATCH/PUT endpoint.
5. Бэкенд пропускает body через bluemonday (allowlist) → записывает в PostgreSQL.
6. При просмотре статьи: `{{ .Body | safeHTML }}` — браузер рендерит HTML напрямую.

### Contracts

| Contract ID | Input / Output | Producer / Consumer | Notes |
| --- | --- | --- | --- |
| `CTR-01` | `articles.body` — HTML-строка | Go handler (producer) / шаблон view (consumer) | Ранее поле могло содержать Markdown; после FT-023 ожидается только sanitized HTML. Исторические статьи рендерятся as-is согласно ASM-03. |
| `CTR-02` | Bluemonday allowlist | handler save (producer) / БД (consumer) | Allowlist: `<p>`, `<h1>`-`<h3>`, `<strong>`, `<em>`, `<s>`, `<a href>`, `<img src alt>`, `<figure>`, `<figcaption>`, `<div style="text-align:*">`, `<br>`. Расширяется только через явное изменение allowlist. |

### Failure Modes

- `FM-01` XSS через unsanitized HTML — bluemonday strippt запрещённые теги/атрибуты; `template.HTML` не экранирует — поэтому санитизация на бэкенде обязательна до записи в БД.
- `FM-02` Upload-ошибка изображения — редактор показывает inline error message, не вставляет broken `<img>`; сохранение статьи не блокируется.
- `FM-03` Существующая статья с Markdown-телом рендерится как plain HTML (теги экранированы) — визуально некорректно; ожидаемое поведение согласно ASM-03 до завершения отдельной миграции.

### ADR Dependencies

| ADR | Current `decision_status` | Used for | Execution rule |
| --- | --- | --- | --- |
| [ADR-007](../../adr/ADR-007-wysiwyg-editor-html-storage.md) | `proposed` | Выбор TipTap и HTML-хранения; allowlist-политика | `proposed` — используется как hypothesis (DEC-01); если ADR будет `rejected`, feature переходит в пересмотр scope. |
| [ADR-005](../../adr/ADR-005-image-storage.md) | `accepted` | Upload-endpoint и путь хранения изображений | Canonical input — использовать напрямую. |

## Verify

### Exit Criteria

- `EC-01` Все 7 formatting actions из acceptance criteria карточки работают в браузере.
- `EC-02` Изображение с подписью вставляется в середину текста и отображается корректно при просмотре.
- `EC-03` XSS payload (`<script>alert(1)</script>`) в теле статьи не выполняется при рендеринге.

### Traceability matrix

| Requirement ID | Design refs | Acceptance refs | Checks | Evidence IDs |
| --- | --- | --- | --- | --- |
| `REQ-01` | `ASM-02`, `DEC-01`, `CTR-01`, `FM-01` | `EC-01`, `SC-01` | `CHK-01` | `EVID-01` |
| `REQ-02`, `REQ-03` | `ASM-01`, `CTR-01`, `FM-02` | `EC-02`, `SC-02` | `CHK-02` | `EVID-02` |
| `REQ-04` | `CON-01`, `CTR-02`, `FM-01` | `EC-03`, `SC-03` | `CHK-03` | `EVID-03` |

### Acceptance Scenarios

- `SC-01` Автор открывает редактор, применяет bold, italic, strikethrough, h2, ссылку и center-выравнивание — сохраняет статью — при просмотре все форматы отображаются корректно.
- `SC-02` Автор нажимает «Вставить изображение» в середине текста, выбирает файл, вводит подпись — после сохранения изображение с подписью отображается в нужной позиции при просмотре статьи.
- `SC-03` В поле body статьи через прямой API-запрос передаётся `<script>alert(1)</script>` — при просмотре статьи скрипт не выполняется, тег отсутствует в DOM.
- `NEG-01` Автор пытается вставить `<iframe>` через редактор — тег удаляется bluemonday при сохранении, статья сохраняется без него.

### Checks

| Check ID | Covers | How to check | Expected result | Evidence path |
| --- | --- | --- | --- | --- |
| `CHK-01` | `EC-01`, `SC-01` | Playwright E2E: открыть `/articles/{id}/edit`, применить bold/italic/strikethrough/h2/link/align-center, сохранить, открыть просмотр — проверить наличие `<strong>`, `<em>`, `<s>`, `<h2>`, `<a>`, `style="text-align:center"` в DOM; дополнительно: визуальная проверка отсутствия JS-ошибок в консоли | Все элементы присутствуют в DOM; нет JS-ошибок | `artifacts/ft-023/verify/chk-01/` |
| `CHK-02` | `EC-02`, `SC-02` | Playwright E2E: вставить изображение через тулбар в середину параграфа, ввести подпись, сохранить, открыть просмотр — проверить наличие `<figure>`, `<img>`, `<figcaption>` в нужном месте DOM; дополнительно: визуальная проверка корректного отображения | `<figure><img><figcaption>` в нужном месте DOM; нет JS-ошибок | `artifacts/ft-023/verify/chk-02/` |
| `CHK-03` | `EC-03`, `SC-03`, `NEG-01` | Авто: HTTP POST article с body содержащим XSS payload → GET → проверить отсутствие `<script>` и `<iframe>` в ответе | Payload stripped; статус 200 | `artifacts/ft-023/verify/chk-03/` |

### Test matrix

| Check ID | Evidence IDs | Evidence path |
| --- | --- | --- |
| `CHK-01` | `EVID-01` | `artifacts/ft-023/verify/chk-01/` |
| `CHK-02` | `EVID-02` | `artifacts/ft-023/verify/chk-02/` |
| `CHK-03` | `EVID-03` | `artifacts/ft-023/verify/chk-03/` |

### Evidence

- `EVID-01` Скриншот просмотра статьи со всеми видами форматирования из SC-01.
- `EVID-02` Скриншот просмотра статьи с изображением и подписью в середине текста.
- `EVID-03` HTTP-лог или тест-вывод: XSS payload отсутствует в rendered HTML.

### Evidence contract

| Evidence ID | Artifact | Producer | Path contract | Reused by checks |
| --- | --- | --- | --- | --- |
| `EVID-01` | Screenshot (PNG) | Playwright E2E | `artifacts/ft-023/verify/chk-01/screenshot.png` | `CHK-01` |
| `EVID-02` | Screenshot (PNG) | Playwright E2E | `artifacts/ft-023/verify/chk-02/screenshot.png` | `CHK-02` |
| `EVID-03` | HTTP test output или Go test log | automated | `artifacts/ft-023/verify/chk-03/result.txt` | `CHK-03` |
