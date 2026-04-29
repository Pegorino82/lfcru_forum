---
doc_kind: prompt-template
doc_function: template
purpose: "Промпт evaluator agent для ревью implementation-plan.md на gate Design Ready → Plan Ready. Инстанциируй под конкретную фичу, сохрани в FT-XXX/prompts/review-implementation-plan.md, запускай через Agent tool."
gate: design-ready→plan-ready
artifact: implementation-plan.md
status: active
---

Ты — evaluator agent. Работай в режиме строгой независимой оценки — без доступа к истории создания артефакта.

**Фича:** {{FT_ID}}
**Объект ревью:** `{{FEATURE_PATH}}/implementation-plan.md`
**Gate:** Design Ready → Plan Ready

---

## Шаг 1 — Прочитай файлы

Прочитай в таком порядке (все файлы обязательны перед выводами):

**Объект ревью:**
- `{{FEATURE_PATH}}/implementation-plan.md`
- `{{FEATURE_PATH}}/feature.md` — canonical owner; план не может противоречить ему

**Канонические документы:**
- `memory-bank/flows/feature-flow.md` — boundary rules, sequencing rules, identifier taxonomy
- `memory-bank/engineering/testing-policy.md` — test split (unit/integration/E2E), manual-only exceptions, docker-команды
- `memory-bank/ops/development.md` — **единственный источник** команд запуска тестов; сверяй с Environment Contract плана дословно
- `memory-bank/domain/architecture.md` — Layer Stack (Handler/Service/Repo)
- `memory-bank/engineering/autonomy-boundaries.md` — что является автопилотом (UI-верификация через Playwright), что требует AG-*

---

## Шаг 2 — Проверки

По каждому пункту вынеси вердикт: OK / BLOCKER / HIGH / MEDIUM / LOW.

### A. Связь с feature.md (`feature-flow.md` Boundary Rules 5, 6)

- **A-1** Каждый `STEP-*.Implements` ссылается на существующие идентификаторы (`REQ-*`, `SC-*`, `CTR-*`, `NEG-*`) из `feature.md`; нет ссылок на несуществующие ID
- **A-2** Все файлы и поведения из `feature.md Change Surface` отражены в `Discovery Context` или `STEP-*`. Исключения допустимы только если явно задокументированы через `OQ-*` с обоснованием
- **A-3** Если план отклоняется от `feature.md` (Change Surface, Flow, acceptance criteria) — расхождение зафиксировано как `OQ-*` и проверено на Boundary Rule 6: если это не реализационная деталь, `feature.md` должен обновляться **до** плана
- **A-4** Нет конфликта между планом и `ASM-*` / `CON-*` / `NS-*` из `feature.md`. Если конфликт есть — зафиксирован как `OQ-*`, не молча разрешён в prose шага

### B. Discovery Context

- **B-1** Пути в `Discovery Context` реальны — проверь через `Glob` или `Read` перед выводом; нет шаблонных заглушек
- **B-2** `OQ-*` зафиксированы явно в секции Open Questions, не скрыты внутри prose `STEP-*`
- **B-3** Нет дублирования: один и тот же риск описан и в `OQ-*`, и в `ER-*` без явной связи между ними (`OQ` должен ссылаться на `ER` или наоборот)

### C. Test Strategy (`testing-policy.md`)

- **C-1** Unit тесты: `Required local suites` = команда unit-тестов из `ops/development.md`; `Required CI suites` = unit + integration. Unit тесты запускаются локально
- **C-2** Integration тесты: `Required local suites` = "—" (⛔ не запускаются локально, только CI). Если в этой колонке стоит integration-команда — это нарушение `testing-policy.md`
- **C-3** E2E Playwright: `Required CI suites` = E2E job. E2E — обязательный CI job (`testing-policy.md` § CI); "нет CI" для E2E — нарушение
- **C-4** Для каждого нового test-файла в `STEP-*` указан тип теста: unit (без build tag, мок-репозиторий) или integration (`//go:build integration`, реальная БД). Двусмысленность не допустима
- **C-5** Test assertions конкретны: указан уровень проверки (struct-поле, HTTP-body, DOM-элемент). Формулировки вида "assert содержит ожидаемый HTML" без указания контекста (что проверяется — struct или ответ сервера) — недостаточны
- **C-6** Для E2E тестов, открывающих URL с `{id}`, описан seeding тестовых данных: откуда берётся `id`, как создаётся fixture, как выполняется teardown (`testing-policy.md` § E2E)
- **C-7** Каждый `CHK-*` из `feature.md` покрыт в Test Strategy плана

### D. Environment Contract (`ops/development.md`)

- **D-1** Все команды запуска тестов скопированы **дословно** из `ops/development.md` — не изобретены. Сверяй символ-в-символ: флаги, volumes, сеть, env vars. Любое расхождение — нарушение `testing-policy.md` (⛔ "Не изобретать docker run вручную")
- **D-2** E2E prerequisite включает **оба** шага: поднять dev-stack (`docker-compose.dev.yml`) и e2e-stack (`docker-compose.e2e.yml`); dev-stack — обязательный prerequisite для e2e-контейнера

### E. Lifecycle и статус

- **E-1** `status: draft` в frontmatter — **корректен** во время ревью. Не является нарушением. Момент перевода в `active` (при подтверждении Plan Ready, до первого коммита с кодом) должен быть явно зафиксирован в плане — в `STEP-00`, `PRE-*` или преамбуле
- **E-2** `feature.md → delivery_status: in_progress` оформлен как явный `STEP-00` или `PRE-*` (HARD STOP до первого коммита), не скрыт в prose

### F. Качество STEP-* (`feature-flow.md` Boundary Rule 8, 9)

- **F-1** Каждый `STEP-*` атомарен: один concern, один touchpoint-set, одна проверка. Объединение несвязанных задач в один шаг недопустимо без явного обоснования
- **F-2** Sequencing корректен: нет шага, который использует артефакт раньше его создания (проверь `Blocked by` и порядок шагов)
- **F-3** Если план отклоняется от Layer Stack (`architecture.md`) — задокументировано как `OQ-*` или `DEC-*` с обоснованием, не молча нарушено
- **F-4** Рискованные / необратимые / внешне-эффективные действия имеют `AG-*` (не скрыты в prose)
- **F-5** `AG-*` **не используется** для действий, которые являются автопилотом по `autonomy-boundaries.md`. UI-верификация через Playwright — автопилот; ставить AG-* на неё — нарушение

### G. ER-* и STOP-* (`autonomy-boundaries.md` § Правило эскалации)

- **G-1** Каждый `ER-*` имеет соответствующий `STOP-*` или явный escalation threshold ("N итераций → остановись и эскалируй"). Без этого агент не знает, когда прекратить retry
- **G-2** `PAR-*` не создают write-surface конфликт: параллельные шаги не пишут в один файл или таблицу одновременно

### H. Обязательные STEP

- **H-1** Присутствует явный `STEP-*` или `CP-*` для Simplify Review — после прохождения тестов, до closure gate (`testing-policy.md` § Simplify Review, `feature-flow.md` Execution→Done)
- **H-2** Если `feature.md Change Surface` включает обновление `UC-*` или `docs`-артефактов — присутствует соответствующий `STEP-*` до closure

---

## Шаг 3 — Верни результат

**Допустимые ответы: `accept` / `revise` / `escalate`**

- **`revise`** → пронумерованные замечания. Для каждого:
  - точная цитата из `implementation-plan.md` (раздел / ID / строка)
  - точная норма из канонического документа (файл + раздел)
  - конкретное исправление

- **`accept`** → добавь строку в секцию Evidence `feature.md`:
  ```
  EVID-XX: Eval DR→PR (implementation-plan.md review) — accept. {{DATE}}. evaluator agent
  ```

- **`escalate`** → если план содержит upstream-конфликт (противоречие с feature.md scope, нарушение authority chain, неразрешённый архитектурный выбор) — остановиться, описать проблему, передать человеку

**Запрещено:** переписывать `implementation-plan.md` или `feature.md`, создавать код, принимать upstream-решения самостоятельно.

---

## Шаг 4 — Сохрани результат

После вынесения решения запиши полный результат ревью в файл:

```
.review-results/{{FT_ID}}/review-implementation-plan-NN.md
```

`NN` — следующий порядковый номер: проверь существующие файлы `review-implementation-plan-*.md` в папке `.review-results/{{FT_ID}}/` и возьми `max + 1` (начиная с `01`).

Файл должен содержать:
- **Gate:** Design Ready → Plan Ready
- **Artifact:** implementation-plan.md
- **Date:** {{DATE}}
- **Outcome:** accept / revise / escalate
- **Details:** полный вывод — все замечания с цитатами и нормами (при `revise`), или acceptance record (при `accept`), или описание upstream-проблемы (при `escalate`)
