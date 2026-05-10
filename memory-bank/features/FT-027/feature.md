---
title: "FT-027: Единый дизайн разделов сайта"
doc_kind: feature
doc_function: canonical
purpose: "Привести визуальный стиль страниц Новости, Форум (разделы) и Форум (темы раздела) к единому дизайну, эталоном которого служит главная страница."
derived_from:
  - ../../domain/problem.md
  - ../../domain/frontend.md
status: active
delivery_status: in_progress
audience: humans_and_agents
must_not_define:
  - implementation_sequence
---

# FT-027: Единый дизайн разделов сайта

## What

### Problem

Страницы сайта визуально несогласованны. Главная (`/`) после FT-021 имеет современный дизайн: карточки новостей с изображениями, виджеты, адаптивная сетка. Остальные страницы выглядят бедно:

- **Новости** (`/news`) — плоский список «заголовок + дата» без изображений, анонсов и метаданных. Контент визуально уже, чем на главной.
- **Форум — разделы** (`/forum`) — одна маленькая карточка по центру страницы с большим пустым пространством. Выглядит инородно.
- **Форум — темы раздела** (`/forum/sections/X`) — стилистически не согласован с остальным сайтом, контент визуально уже.

### Outcome

| Metric ID | Metric | Baseline | Target | Measurement method |
| --- | --- | --- | --- | --- |
| `MET-01` | Визуальная согласованность разделов | 1 из 4 страниц в едином стиле | 4 из 4 страниц в едином стиле | Визуальная проверка + Playwright screenshots |

### Scope

- `REQ-01` Страница новостей (`/news`) отображает новости в формате карточек: изображение (или placeholder), заголовок, анонс, дата, количество комментариев — аналогично карточкам на главной.
- `REQ-02` Карточки новостей располагаются в адаптивной сетке: 1 колонка на мобильных, 2 на планшетах, 3 на десктопе.
- `REQ-03` Пагинация на странице новостей сохраняется и визуально соответствует стилю сайта.
- `REQ-04` Форум — список разделов (`/forum`) отображает разделы в формате карточек на полную ширину контейнера.
- `REQ-05` Форум — темы раздела (`/forum/sections/X`) стилистически приведён к единому виду: ширина контента, карточки тем, типографика соответствуют остальным страницам.

### Non-Scope

- `NS-01` Тема форума (`/forum/topics/X`) — не меняется, к ней нет замечаний.
- `NS-02` Header и footer — не меняются, они уже единообразны.
- `NS-03` Главная страница (`/`) — эталон, не меняется.
- `NS-04` Содержимое карточки раздела форума (описание, статистика) — минимальное: название + количество тем. Расширение — отдельная задача.
- `NS-05` Мобильный responsive для header/footer/topic — не в scope.

### Constraints / Assumptions

- `ASM-01` Go-структура `News` в handler `/news` уже содержит поля для изображения и анонса (аналогично `HomeNewsItem` на главной), либо их можно получить через существующий repo-слой.
- `ASM-02` Контейнер `.container` (960px max-width) из `base.html` — общий для всех страниц. Менять его ширину не нужно.
- `CON-01` Все стили пишутся inline в шаблонах (нет отдельных CSS-файлов) — как в остальных шаблонах проекта.
- `CON-02` CSRF-токен обязателен для POST-запросов (PCON-02), но пагинация использует GET — не затрагивается.
- `DEC-01` Breakpoints для адаптивной сетки новостей: мобильные (<640px) — 1 колонка, планшеты (640–1023px) — 2 колонки, десктоп (≥1024px) — 3 колонки.
- `DEC-02` Page size для карточной пагинации новостей: 9 (кратно 3 колонкам, визуально ровная сетка).
- `DEC-03` Адаптивная сетка новостей реализуется через CSS Grid.

## How

### Solution

Переработать шаблоны трёх страниц, взяв за визуальный эталон стиль главной: карточки с border-radius, box-shadow, единая типографика и цветовая схема (#c8102e акцент). Для новостей — CSS Grid с адаптивными breakpoints. Handler новостей адаптировать под карточный формат (передавать image/excerpt, скорректировать page size). Форум — расширить карточки разделов на полную ширину, унифицировать стиль тем.

**Trade-off:** альтернатива — shared CSS-файл (`static/css/sections.css`) вместо inline стилей в шаблонах. Отклонено: проект использует inline стили во всех существующих шаблонах (CON-01), вынос в отдельный файл нарушит единообразие и потребует подключения нового ресурса в `base.html`.

### Change Surface

| Surface | Type | Why it changes |
| --- | --- | --- |
| `templates/news/list.html` | code | Полная переработка: плоский список → карточная сетка с изображениями |
| `internal/news/handler.go` | code | Передача дополнительных полей (image, excerpt, comment count), изменение page size |
| `templates/forum/index.html` | code | Переработка: маленькая карточка по центру → full-width карточки |
| `templates/forum/section.html` | code | Стилистическая унификация: карточки тем, типографика |
| `internal/news/repo.go` | code | Возможно: расширение SQL-запроса для получения image/excerpt в списке |

### Flow

1. Пользователь открывает `/news`, `/forum`, `/forum/sections/X`.
2. Handler запрашивает данные из сервиса/репозитория.
3. Шаблон рендерит данные в карточном формате с адаптивной сеткой.
4. Браузер применяет CSS Grid breakpoints для responsive layout.

### Contracts

| Contract ID | Input / Output | Producer / Consumer | Notes |
| --- | --- | --- | --- |
| `CTR-01` | `ListData.Items` — расширенный набор полей (image, excerpt, comment count) | `news.Handler` / `templates/news/list.html` | Должен быть совместим с существующей пагинацией |

### Failure Modes

- `FM-01` У новости нет изображения → показывать placeholder (как на главной — `news-placeholder` с иконкой).
- `FM-02` Пустой список разделов форума → показывать сообщение «Разделов пока нет».
- `FM-03` Невалидный параметр пагинации (page=0, page=-1, page=NaN) → handler нормализует к page=1 (существующее поведение сохраняется).

### ADR Dependencies

Нет ADR-зависимостей.

## Verify

### Exit Criteria

- `EC-01` Страницы `/news`, `/forum`, `/forum/sections/X` визуально соответствуют стилю главной.
- `EC-02` Карточки новостей отображаются в адаптивной сетке (1/2/3 колонки) при разных viewport.
- `EC-03` Пагинация на `/news` работает корректно с новым page size.
- `EC-04` Существующая функциональность (ссылки, навигация, breadcrumbs) не сломана.

### Traceability matrix

| Requirement ID | Design refs | Acceptance refs | Checks | Evidence IDs |
| --- | --- | --- | --- | --- |
| `REQ-01` | `ASM-01`, `CON-01`, `CTR-01`, `FM-01` | `EC-01`, `SC-01` | `CHK-01`, `CHK-04` | `EVID-01`, `EVID-04` |
| `REQ-02` | `DEC-01`, `CON-01` | `EC-02`, `SC-02` | `CHK-02` | `EVID-02` |
| `REQ-03` | `DEC-02`, `CON-02`, `FM-03` | `EC-03`, `SC-03`, `NEG-01`, `NEG-02` | `CHK-03`, `CHK-05` | `EVID-03`, `EVID-05` |
| `REQ-04` | `CON-01`, `FM-02` | `EC-01`, `SC-04` | `CHK-01` | `EVID-01` |
| `REQ-05` | `CON-01` | `EC-01`, `SC-05` | `CHK-01` | `EVID-01` |

### Acceptance Scenarios

- `SC-01` Открыть `/news` — все новости отображаются карточками с изображением (или placeholder), заголовком, анонсом, датой и количеством комментариев.
- `SC-02` Изменить viewport: 400px → 1 колонка карточек, 800px → 2 колонки, 1280px → 3 колонки.
- `SC-03` На `/news` перейти на 2-ю страницу — карточки загружаются, пагинация показывает корректное состояние.
- `SC-04` Открыть `/forum` — разделы отображаются карточками на полную ширину контейнера с названием и количеством тем.
- `SC-05` Открыть `/forum/sections/X` — темы отображаются в стиле, согласованном с остальными страницами.

### Negative Scenarios

- `NEG-01` Открыть `/news?page=0` и `/news?page=-1` — страница отображает первую страницу новостей без ошибок.
- `NEG-02` Открыть `/news?page=999` (несуществующая страница) — страница отображается без ошибок (пустой список или редирект на последнюю страницу).

### Checks

| Check ID | Covers | How to check | Expected result | Evidence path |
| --- | --- | --- | --- | --- |
| `CHK-01` | `EC-01`, `SC-01`, `SC-04`, `SC-05` | Playwright: скриншоты `/news`, `/forum`, `/forum/sections/X` при 1280px viewport | Карточный layout, единый стиль | `artifacts/ft-027/verify/chk-01/` |
| `CHK-02` | `EC-02`, `SC-02` | Playwright: скриншоты `/news` при 400px, 800px, 1280px viewport | 1, 2, 3 колонки соответственно | `artifacts/ft-027/verify/chk-02/` |
| `CHK-03` | `EC-03`, `SC-03` | Playwright: перейти на страницу 2 в `/news`, проверить наличие карточек и пагинации | Карточки загружаются, навигация корректна | `artifacts/ft-027/verify/chk-03/` |
| `CHK-04` | `EC-04` | Playwright: проверить отсутствие JS-ошибок в консоли на всех изменённых страницах | Нет ошибок в консоли | `artifacts/ft-027/verify/chk-04/` |
| `CHK-05` | `NEG-01`, `NEG-02` | Playwright: открыть `/news?page=0`, `/news?page=-1`, `/news?page=999` — проверить HTTP 200 и отсутствие ошибок | Страницы отображаются без ошибок | `artifacts/ft-027/verify/chk-05/` |

### Test matrix

| Check ID | Evidence IDs | Evidence path |
| --- | --- | --- |
| `CHK-01` | `EVID-01` | `artifacts/ft-027/verify/chk-01/` |
| `CHK-02` | `EVID-02` | `artifacts/ft-027/verify/chk-02/` |
| `CHK-03` | `EVID-03` | `artifacts/ft-027/verify/chk-03/` |
| `CHK-04` | `EVID-04` | `artifacts/ft-027/verify/chk-04/` |
| `CHK-05` | `EVID-05` | `artifacts/ft-027/verify/chk-05/` |

### Evidence

- `EVID-01` Скриншоты всех изменённых страниц при десктопном viewport (1280px).
- `EVID-02` Скриншоты `/news` при 3 viewport breakpoints (400px, 800px, 1280px).
- `EVID-03` Скриншот `/news?page=2` с карточками и пагинацией.
- `EVID-04` Лог Playwright: отсутствие JS-ошибок в консоли.
- `EVID-05` Скриншоты/логи Playwright для невалидных параметров пагинации.
- `EVID-06` Brief loop — accept. 2026-05-07. evaluator agent
- `EVID-07` Spec loop — accept. 2026-05-07. evaluator agent

### Evidence contract

| Evidence ID | Artifact | Producer | Path contract | Reused by checks |
| --- | --- | --- | --- | --- |
| `EVID-01` | Screenshots (PNG) | Playwright | `artifacts/ft-027/verify/chk-01/` | `CHK-01` |
| `EVID-02` | Screenshots (PNG) | Playwright | `artifacts/ft-027/verify/chk-02/` | `CHK-02` |
| `EVID-03` | Screenshot (PNG) | Playwright | `artifacts/ft-027/verify/chk-03/` | `CHK-03` |
| `EVID-04` | Console log (TXT) | Playwright | `artifacts/ft-027/verify/chk-04/` | `CHK-04` |
| `EVID-05` | Screenshots/logs (PNG/TXT) | Playwright | `artifacts/ft-027/verify/chk-05/` | `CHK-05` |
