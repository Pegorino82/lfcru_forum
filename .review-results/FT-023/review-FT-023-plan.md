# Ревью implementation-plan.md (FT-023) vs memory-bank

**Документ:** `memory-bank/features/FT-023/implementation-plan.md`
**Ревью против:** `testing-policy.md`, `autonomy-boundaries.md`, `ops/development.md`, `domain/architecture.md`, `flows/feature-flow.md`, `features/FT-023/feature.md`

---

## Критические расхождения

### 1. Environment Contract — команда integration-тестов неполная (строка 73)

**В плане:**
```
docker run ... golang:1.23-alpine go test -tags integration -p 1 ./internal/admin/...
```

**Проблема:** Не хватает `--network lfcru_forum_default`, `-e DATABASE_URL=...`, `-v lfcru_gomod:/root/go/pkg/mod`.

**Источник:** `testing-policy.md`:
> "⛔ Не изобретать `docker run` вручную. Причина: без `--network lfcru_forum_default` hostname `postgres` не резолвится — тесты падают с `no such host`."

Команду нужно брать дословно из `ops/development.md` § «Go-тесты».

**Блокирует:** CP-03 — тесты упадут с `no such host` при локальном запуске.

---

### 2. Test Strategy — "Required local suites: Integration" (строка 49)

**В плане:** для sanitize-поверхности указано `Integration (-tags integration)` как required local suite.

**Источник:** `testing-policy.md`:
> "Перед handoff агент прогоняет **только unit-тесты**. Integration-тесты запускаются только в CI."

Это прямое противоречие: integration-тесты не должны числиться обязательными локально.

**Блокирует:** Plan Ready gate.

---

### 3. Test Strategy — "Required CI: нет CI" для E2E-поверхности (строка 50)

**В плане:** для Playwright E2E-поверхности указано `нет CI`.

**Источник:** `testing-policy.md`:
> "Все три job-а обязательны для merge (branch protection на `main`)."

E2E Playwright — один из трёх обязательных CI job-ов. Указывать `нет CI` — неверно.

**Блокирует:** Plan Ready gate.

---

## Умеренные расхождения

### 4. `status: draft` в frontmatter (строка 8)

`feature-flow.md` требует `implementation-plan.md → status: active` как предусловие Plan Ready gate. Документ не переведён в `active`.

**Блокирует:** Plan Ready gate.

---

### 5. CP-02 — manual-only gap без AG-*

**В плане:** CP-02 содержит *"Ручная проверка localhost:8080"* — это manual-only gap.

**Источник:** `feature-flow.md`:
> "каждый manual-only gap имеет причину, ручную процедуру и `AG-*` с approval ref."

В плане `AG-*` не определён. Нарушает Plan Ready → Execution gate.

---

### 6. Sanitization в Handler без обоснования архитектурного отклонения

**В плане:** STEP-02 добавляет `sanitizeArticleBody()` прямо в `articles_handler.go`.

**Источник:** `architecture.md`:
> "Handler — парсит запрос, вызывает Service, рендерит шаблон. Service — содержит всю domain-логику: валидацию…"

Санитизация — domain-логика, она должна жить в Service-слое. Если в `news`-пакете нет Service-слоя, это нужно явно зафиксировать как обоснованное отклонение (`DEC-*` или `CON-*`), а не молча нарушить layer stack.

---

### 7. Расхождение подхода рендеринга с feature.md

**В feature.md** (строки 70, 81): рендеринг описан как `{{ .Body | safeHTML }}` в шаблоне.

**В implementation-plan** (STEP-03): реализует через `ContentHTML: template.HTML(article.Content)` в Go-хендлере; шаблон уже использует `{{.ContentHTML}}` как `template.HTML`.

Это два разных подхода к одной задаче. Расхождение не задокументировано как `OQ-*`, хотя план де-факто выбирает реализацию, отличную от описанной в canonical `feature.md`.

**Источник:** `feature-flow.md`:
> "Если меняются architecture или acceptance criteria — сначала обновляется `feature.md`, потом downstream-план."

---

## Незначительные замечания

### 8. Путь Change Surface в feature.md: `internal/handler/article.go`

`feature.md` (строка 68) указывает несуществующий путь. Реальные файлы — `internal/admin/articles_handler.go` и `internal/news/handler.go`. План правильно их grounding-ует в Discovery Context, однако расхождение не зафиксировано явно — ни как `OQ-*`, ни как примечание к Discovery.

---

## Итоговая таблица

| # | Тип | Описание | Блокирует |
|---|---|---|---|
| 1 | Критическое | Неверная Docker-команда для integration-тестов | CP-03 упадёт с `no such host` |
| 2 | Критическое | Integration тесты заявлены как required local — нарушение testing-policy | Plan Ready gate |
| 3 | Критическое | E2E отмечено `нет CI` — неверно, E2E обязателен в CI | Plan Ready gate |
| 4 | Умеренное | `status: draft`, не переведён в `active` | Plan Ready gate |
| 5 | Умеренное | CP-02 manual gap без AG-* и approval ref | Plan Ready → Execution gate |
| 6 | Умеренное | Sanitization в Handler без обоснования отклонения от layer stack | Архитектурная целостность |
| 7 | Умеренное | Подход рендеринга расходится с feature.md, не задокументировано как OQ-* | Трассируемость |
| 8 | Незначительное | Путь `internal/handler/article.go` в feature.md не существует | — |
