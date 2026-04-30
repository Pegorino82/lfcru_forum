# Memory Bank Review — Консистентность и непротиворечивость

**Дата**: 2026-04-11
**Ревьювер**: Claude Sonnet 4.6
**Ветка**: memory-bank

---

## Объём проверки

Проверены все 37 файлов `memory-bank/` + `CLAUDE.md`:
- `dna/` — 6 файлов (governance-слой)
- `domain/` — 4 файла (product context)
- `engineering/` — 5 файлов (техническая политика)
- `flows/` — 11 файлов (lifecycle + шаблоны)
- `ops/` — 6 файлов (operations)
- `features/`, `prd/`, `adr/`, `use-cases/` — index-only README (4 файла)

---

## Итог

**Критических противоречий не обнаружено.** Документация консистентна и готова к использованию.

---

## Детальные результаты

### 1. Dependency tree

✅ Линейный и ацикличный. Все `derived_from` указывают upstream:

```
principles.md
  ├── governance.md
  │   ├── frontmatter.md
  │   ├── lifecycle.md
  │   ├── cross-references.md
  │   └── (регламентирует domain/, engineering/, flows/, ops/)
  ├── domain/problem.md (canonical)
  ├── domain/architecture.md
  ├── domain/frontend.md
  ├── engineering/autonomy-boundaries.md
  ├── engineering/coding-style.md
  ├── engineering/git-workflow.md
  ├── engineering/testing-policy.md (derived from feature-flow.md)
  ├── flows/feature-flow.md
  ├── flows/workflows.md
  └── ops/* (all from dna/governance.md)
```

Единственный потенциальный цикл (`testing-policy.md` ↔ `feature-flow.md`) разрешён корректно: `testing-policy.md` — downstream, `feature-flow.md` не ссылается обратно.

---

### 2. Frontmatter schema

✅ Все поля из `dna/frontmatter.md` заполнены корректно.

| Поле | Требование | Результат |
|---|---|---|
| `status` | обязательно везде | присутствует |
| `derived_from` | для active non-root | корректно заполнено |
| `delivery_status` | для feature | в шаблонах (draft) |
| `decision_status` | для ADR | в шаблонах (proposed) |
| `doc_kind` | для governance-docs | соответствует папке |
| `doc_function` | для governance-docs | корректно (canonical/index/template) |

---

### 3. CLAUDE.md ↔ memory-bank

✅ Полная синхронизация. Все 13 ссылок из таблицы навигации в CLAUDE.md ведут на существующие файлы.

| CLAUDE.md | memory-bank | Статус |
|---|---|---|
| Продукт, роли, workflows | `domain/problem.md` | ✅ |
| Архитектура, слои, модули | `domain/architecture.md` | ✅ |
| Шаблоны, HTMX, Alpine.js | `domain/frontend.md` | ✅ |
| Конвенции кода | `engineering/coding-style.md` | ✅ |
| Тест-политика | `engineering/testing-policy.md` | ✅ |
| Автономия, эскалация | `engineering/autonomy-boundaries.md` | ✅ |
| Запуск, тесты, Docker | `ops/development.md` | ✅ |
| Env vars, конфигурация | `ops/config.md` | ✅ |
| Production, staging | `ops/stages.md` | ✅ |
| Релиз, деплой | `ops/release.md` | ✅ |
| Маршрутизация задач | `flows/workflows.md` | ✅ |
| Feature lifecycle | `flows/feature-flow.md` | ✅ |
| Реализованные фичи | `features/README.md` | ✅ |

---

### 4. Stable identifiers ↔ шаблоны

✅ Все ID-серии из `flows/feature-flow.md` корректно покрыты шаблонами.

- `short.md` использует: `REQ-*`, `NS-*`, `SC-*`, `CON-*`, `EC-*`, `CHK-*`, `EVID-*`
- `large.md` расширяет: + `ASM-*`, `DEC-*`, `CTR-*`, `FM-*`, `RB-*`, `NEG-*`, `RJ-*`, `INV-*`
- `implementation-plan.md` использует: `PRE-*`, `STEP-*`, `CHK-*`, `EVID-*`, `OQ-*`, `AG-*`

---

### 5. Lifecycle gates

✅ Gates в `flows/feature-flow.md` последовательны и проверяемы.

| Gate | Предикаты | Статус |
|---|---|---|
| Bootstrap | README + выбор шаблона | ✅ |
| Draft → Design Ready | `feature.md: active`, ≥1 `REQ-*`, `NS-*`, `SC-*`, traceability | ✅ |
| Design Ready → Plan Ready | grounding, `implementation-plan.md: active`, ≥1 `PRE-*`, `STEP-*` | ✅ |
| Execution → Done | все `CHK-*`/`EVID-*` заполнены, тесты зелёные, simplify review | ✅ |
| → Cancelled | `delivery_status: cancelled`, plan отсутствует или archived | ✅ |

---

### 6. Module contracts: architecture ↔ config

✅ `domain/architecture.md` и `ops/config.md` синхронизированы: оба указывают `internal/config/config.go` как canonical owner env-переменных.

---

### 7. Testing ownership: testing-policy ↔ feature-flow

✅ Оба документа синхронны в части: кто владеет test cases (`feature.md`) и кто — стратегией (`implementation-plan.md`). `feature-flow.md` явно помечает свою секцию как выжимку из `testing-policy.md` с напоминанием синхронизировать при изменении.

---

### 8. `must_not_define` контракты

✅ Все шаблоны корректно задекларировали, что они **не определяют**:

- `feature/short.md`, `feature/large.md` → `must_not_define: implementation_sequence`
- `implementation-plan.md` → `must_not_define: ft_xxx_scope, ft_xxx_architecture, ft_xxx_acceptance_criteria, ft_xxx_blocker_state`
- `testing-policy.md` → `must_not_define: feature_acceptance_criteria, feature_scope`
- `prd/PRD-XXX.md` → `must_not_define: implementation_sequence, architecture_decision, feature_level_verify_contract`

---

### 9. Orphaned files

✅ Не обнаружены. Все файлы зарегистрированы в parent README.

---

### 10. Empty catalogs

✅ `adr/`, `features/`, `prd/`, `use-cases/` содержат только index README — нормальное состояние для шаблона без инстанцированных документов. Принцип 9 в `dna/principles.md` корректно фиксирует intent создавать ADR, не требуя их наличия.

---

## Замечания (не дефекты)

| # | Описание | Файл | Серьёзность |
|---|---|---|---|
| 1 | `testing-policy.md` и `feature-flow.md` дублируют секции `Sufficient Coverage` и `Test Ownership Summary` — это explicit выжимка, задокументированная явно | `engineering/testing-policy.md`, `flows/feature-flow.md` | ℹ️ info |
| 2 | `ops/config.md` не содержит явного напоминания о `internal/config/config.go` как canonical source кода — оно описано только в `domain/architecture.md` | `ops/config.md` | ℹ️ info |
| 3 | `dna/principles.md` Принцип 9 требует ADR для каждого архитектурного решения, но `adr/` пуст — принятые решения (Echo, pgx, goose) нигде не задокументированы в ADR-формате | `adr/README.md`, `dna/principles.md:9` | ⚠️ minor |

---

## Рекомендации

**По замечанию #3** (наиболее значимое): рассмотреть создание ретроспективных ADR для ключевых архитектурных решений — выбор Echo, pgx, goose, HTMX, SSE. Это унифицирует документацию с задекларированным принципом.

**По замечаниям #1 и #2**: не требуют действий, так как являются сознательными дизайн-решениями.

---

## Вывод

Memory-bank **консистентен и непротиворечив**. Dependency tree ацикличен, frontmatter-контракты соблюдены, lifecycle gates проверяемы, CLAUDE.md полностью согласован со структурой. Единственное содержательное замечание — отсутствие ADR при наличии принципа их обязательного создания.
