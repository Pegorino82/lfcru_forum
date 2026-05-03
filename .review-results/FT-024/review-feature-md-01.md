---
gate: "Draft → Design Ready"
artifact: "memory-bank/features/FT-024/feature.md"
date: 2026-05-03
outcome: revise
---

# Review: FT-024 feature.md — Gate Draft → Design Ready

**Date:** 2026-05-03
**Gate:** Draft → Design Ready
**Artifact:** `memory-bank/features/FT-024/feature.md`
**Outcome:** revise

---

## Результаты проверки

### A. Идентификаторы и структура

- **A-1** OK — Plan IDs (`OQ-*`, `PRE-*`, `STEP-*`, `WS-*`, `AG-*`, `PAR-*`, `CP-*`, `ER-*`, `STOP-*`) в feature.md отсутствуют.
- **A-2** OK — Присутствуют: `REQ-01..07`, `NS-01..06`, `SC-01..07`, `CHK-01..02`, `EVID-01..02`. Все серии полны и без пропусков.
- **A-3** OK — Каждый объявленный идентификатор определён явно в своей секции.
- **A-4** OK — `large.md` обоснован: фича содержит `ASM-*`, `CTR-*`, `FM-*`, несколько `SC-*`, несколько `CHK-*`/`EVID-*`, ADR-dependent design rules. Все условия для `short.md` нарушены — выбор `large.md` корректен.
- **A-5** OK — Запись `FT-024 | Профиль пользователя (quick view + страница + аватар) | planned` присутствует в `memory-bank/features/README.md`. `delivery_status: planned` совпадает с frontmatter `feature.md`.

### B. Traceability

- **B-1** OK — Каждый `REQ-01..07` прослеживается к ≥1 `SC-*` через traceability matrix (REQ-01→SC-01,SC-05; REQ-02→SC-02; REQ-03→SC-03; REQ-04→SC-03,SC-04; REQ-05→SC-06; REQ-06→SC-01; REQ-07→SC-07).
- **B-2** OK — Все `SC-*` покрыты через `CHK-01` (обобщённо по цепочке REQ→CHK). Структура трассировки через REQ-уровень, а не SC-уровень, нестандартна, но функционально достаточна.
- **B-3** OK — `CHK-01`→`EVID-01`, `CHK-02`→`EVID-02`. Каждый `CHK-*` имеет соответствующий `EVID-*`.
- **B-4** OK — Критичные failure modes присутствуют (`FM-01..FM-06`); `NEG-01..NEG-05` явно определены.

### C. Непротиворечивость

- **C-1** **MEDIUM** — В traceability matrix для `REQ-05` указан `design ref: CON-04`. Однако `CON-04` фиксирует: «Относительное время вычисляется на сервере при рендере шаблона (Go-функция в FuncMap)». `REQ-05` — это fallback-аватар (инициалы на цветном фоне с детерминированным цветом по хешу username). Семантической связи `REQ-05 → CON-04` нет. Правильная design ref для REQ-05 отсутствует или должна ссылаться на иной constraint (детерминированный алгоритм хеширования имени пользователя для цвета), который не зафиксирован явно.

  **Цитата:** строка `| REQ-05 | CON-04 | EC-04, SC-06 | CHK-01 | EVID-01 |` в Traceability matrix.
  **Норма:** Traceability contract (feature-flow.md § «Traceability Contract»): design refs должны указывать на реальные constraints/assumptions, которые формируют данный requirement.
  **Исправление:** Либо удалить `CON-04` из design refs `REQ-05` (заменить на `—`), либо добавить отдельный `CON-*` или `ASM-*` о детерминированном алгоритме хеширования для цвета аватара и сослаться на него.

- **C-2** OK — Нет циклических зависимостей `derived_from`. `feature.md` → `domain/problem.md` и `adr/ADR-005-image-storage.md`. ADR-005 → `features/FT-009/feature.md`. Цикла нет.
- **C-3** OK — ADR-005 имеет `decision_status: accepted`; в feature.md секция `ADR Dependencies` трактует его как «Canonical input — следовать без альтернатив». Соответствует требованию.
- **C-4** OK — `NS-06` (нет CDN/S3) не противоречит `REQ-04` (хранение на файловой системе по ADR-005). `ASM-01` (уникальность username) не противоречит `CON-02` (проверка авторизации по UserID). Внутренних противоречий не обнаружено.

### D. Соответствие архитектуре

- **D-1** **MEDIUM** — В Change Surface указан путь `static/css/ или inline`. Директория `static/css/` не существует в репозитории (существуют только `static/js/` и `static/img/`). Путь к несуществующей директории вводит в заблуждение.

  **Цитата:** строка `| static/css/ или inline | code | Стили модалки, аватар-кружка |` в Change Surface.
  **Норма:** D-1 требует, чтобы пути в Change Surface реально существовали в репозитории.
  **Исправление:** Заменить на `static/js/` или `templates/profile/page.html` (inline стили) — либо явно зафиксировать, что CSS будет только inline в шаблонах, убрав путь `static/css/`.

- **D-1** **LOW** — В Change Surface упомянуты `templates/forum/sections.html` / `topics.html`. Реальные файлы в репозитории: `templates/forum/index.html` (список разделов) и `templates/forum/section.html` (список тем и последний пост раздела). Имена файлов не совпадают с реальными.

  **Цитата:** строка `| templates/forum/sections.html / topics.html | code | Кликабельное имя автора последнего поста |` в Change Surface.
  **Норма:** D-1 требует соответствия реальным путям в репозитории.
  **Исправление:** Заменить на `templates/forum/index.html` / `templates/forum/section.html`.

- **D-2** OK — Новые шаблоны `templates/profile/page.html` и `templates/profile/modal.html` следуют паттерну `templates/<domain>/` из `frontend.md`. Обновляемые шаблоны (`templates/layouts/base.html`, `templates/forum/topic.html`, `templates/news/article.html`) — существуют и корректно поименованы (с учётом замечания D-1 LOW выше).
- **D-3** OK — `static/css/ или inline` — при условии исправления D-1 MEDIUM. `static/js/` существует.
- **D-4** OK — Логика получения профиля (посты, комментарии, relative time) отнесена к `internal/user/service.go`, handler только получает данные и рендерит шаблон. Соответствует архитектурному паттерну Handler→Service→Repo.

### E. Тестовая политика

- **E-1** OK — Нет UI-изменений, помеченных manual-only. Все acceptance scenarios покрыты `CHK-01` (Playwright E2E).
- **E-2** OK — HTMX/Alpine.js взаимодействия (модалка, загрузка аватара, закрытие по ESC) покрыты `CHK-01` (Playwright E2E). `testing-policy.md` §«Когда Manual-Only Допустим» явно указывает: «Browser-специфика и HTMX/Alpine.js-взаимодействия — покрываются Playwright».
- **E-3** OK — `EVID-01` producer: CI/Playwright (Go Tests + E2E). `EVID-02` producer: Playwright. Соответствует методам CHK-01 и CHK-02.
- **E-4** N/A — Manual-only gaps отсутствуют; данный пункт не применяется.

### F. Системные ограничения

- **F-1** OK — `CON-01` явно фиксирует: «CSRF-токен обязателен для `POST /profile/avatar` (PCON-02)». `CTR-03` фиксирует: «Требует авторизации + CSRF». PCON-02 соблюдён.
- **F-2** OK — Нет молчаливых допущений по security-инфраструктуре. `CON-02` явно фиксирует проверку `session.UserID == profile.UserID`. `NEG-03` покрывает неавторизованную попытку. `NEG-05` покрывает попытку изменить чужой профиль.

### G. Glossary и Use Case

- **G-1** **LOW** — Термины «quick-view» и «relative time» активно используются в feature.md (в purpose, REQ-01, CON-04, Flow), но отсутствуют в `memory-bank/domain/glossary.md`. По feature-flow.md §12: «Если фича вводит новые доменные или архитектурные термины, соответствующие записи в glossary.md должны быть добавлены или обновлены к моменту Design Ready».

  **Цитата:** `purpose: "...quick-view модалка..."`, `CON-04: "Относительное время (relative time) вычисляется на сервере..."`.
  **Норма:** feature-flow.md §12 + glossary.md как canonical для term definitions.
  **Исправление:** Добавить в `memory-bank/domain/glossary.md` определения: «quick-view» (модалка с краткой информацией о пользователе, открываемая по клику без перехода на страницу профиля) и «relative time» (человекочитаемое относительное время, например «2 часа назад», вычисляемое на сервере через Go FuncMap).

- **G-2** OK — FT-024 создаёт новый project-level сценарий (просмотр профиля пользователя), однако по feature-flow.md §11 соответствующий `UC-*` должен быть создан/обновлён «до closure» (Done gate), не до Design Ready. Требование не является блокером для данного gate.

---

## Итог

**Outcome: revise**

### Замечания (требуют исправления перед переходом в Design Ready)

1. **[MEDIUM — C-1]** Traceability matrix, строка `REQ-05`: неверный design ref `CON-04` (relative time). `REQ-05` — fallback-аватар, CON-04 к нему не относится. Исправить: убрать `CON-04` из design refs REQ-05 (заменить на `—`) или добавить явный `ASM-*`/`CON-*` о детерминированном хеш-алгоритме цвета и сослаться на него.

2. **[MEDIUM — D-1]** Change Surface, строка `static/css/ или inline`: директория `static/css/` не существует. Исправить: заменить на явную формулировку «inline стили в шаблонах» без упоминания несуществующего пути `static/css/`.

3. **[LOW — D-1]** Change Surface, строка `templates/forum/sections.html / topics.html`: реальные файлы называются `templates/forum/index.html` и `templates/forum/section.html`. Исправить имена файлов.

4. **[LOW — G-1]** Термины «quick-view» и «relative time» используются в feature.md, но не определены в `memory-bank/domain/glossary.md`. Добавить определения согласно feature-flow.md §12.
