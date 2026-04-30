# Ревью FT-023 Feature Package

**Объект ревью:** `memory-bank/features/FT-023/` (feature.md, README.md) + `memory-bank/adr/ADR-007-wysiwyg-editor-html-storage.md`
**Эталон:** `memory-bank/flows/feature-flow.md`, `memory-bank/domain/glossary.md`, `memory-bank/use-cases/UC-001-article-publishing.md`
**Дата:** 2026-04-28

---

## Соответствует стандартам

- **Package Rule 2–3** — `feature.md` + `README.md` присутствуют, `implementation-plan.md` отсутствует (корректно при `status: draft`)
- **Обязательные секции** — `What`, `How`, `Verify` есть, структура полная
- **Минимальный набор идентификаторов** — `REQ-*`, `NS-*`, `ASM-*`, `CON-*`, `CTR-*`, `FM-*`, `SC-*`, `NEG-*`, `CHK-*`, `EVID-*` — все присутствуют
- **Traceability matrix** — каждый `REQ-*` прослеживается к `SC-*` / `CHK-*` / `EVID-*`
- **ADR dependency** — ADR-007 правильно указан с текущим `decision_status: proposed`, CON-02 фиксирует правило выполнения
- **Выбор шаблона** — `large.md` оправдан (несколько `ASM-*`, `CTR-*`, `FM-*`, ADR-зависимость)

---

## Критические нарушения

### 1. OQ-* идентификаторы в `feature.md` — нарушение identifier taxonomy

В `feature.md` используются ссылки на `OQ-01` в трёх местах:
- `NS-03`: `...обрабатываются отдельным OQ в implementation-plan.md`
- Flow step 1: `...если Markdown — OQ-01`
- `FM-03`: `...см. OQ-01 в implementation-plan.md`

`OQ-*` — Plan ID, определённый только для `implementation-plan.md` (`feature-flow.md` строки 193, 213). `feature.md` не должен содержать OQ-ссылки. Незакрытая неопределённость выражается через prose или `ASM-*` / `CON-*`, а сам `OQ-01` появится только в плане.

**Исправление:** убрать упоминания `OQ-01` из `feature.md`, неопределённость по Markdown-телам оформить как `ASM-*` или `CON-*`.

### 2. Циклическая зависимость ADR-007 <-> feature.md — нарушение governance

`ADR-007` имеет `derived_from: ../features/FT-023/feature.md`, при этом `feature.md` имеет `derived_from: ../../adr/ADR-007-wysiwyg-editor-html-storage.md`. Циклические зависимости явно запрещены (`glossary.md`: "Authority течёт upstream → downstream; циклические зависимости запрещены"). ADR должен быть upstream-владельцем решения, `feature.md` — downstream-потребителем.

**Исправление:** убрать `../features/FT-023/feature.md` из `derived_from` ADR-007.

---

## Нарушения, блокирующие переход Draft -> Design Ready

### 3. Glossary не обновлён (`feature-flow.md` Package Rule 12)

> "Если фича вводит новые доменные или архитектурные термины... записи в `glossary.md` должны быть добавлены до Design Ready."

FT-023 вводит термины, отсутствующие в `glossary.md`:

| Термин | Где используется |
|---|---|
| `WYSIWYG` | feature.md, ADR-007 |
| `TipTap` | feature.md, ADR-007 |
| `bluemonday` | feature.md, ADR-007 |
| `allowlist` | feature.md CON-01, CTR-02 |
| `safeHTML` / `template.HTML` | feature.md Flow, ADR-007 |

**Исправление:** добавить перечисленные термины в `memory-bank/domain/glossary.md` до перевода `feature.md` в `status: active`.

---

## Средние проблемы

### 4. `ADR-007` имеет `status: draft`

ADR-007 создан и активно используется обоими документами. `status: draft` предполагает, что документ ещё не готов. Если он заморожен намеренно до подтверждения `decision_status: proposed -> accepted`, это стоит явно прояснить: либо перевести в `status: active`, либо добавить комментарий в ADR.

### 5. `UC-001` не обновлён для FT-023

`UC-001` (article publishing) в traceability ссылается только на `FT-008`, `FT-009`. FT-023 материально меняет сценарий редактирования статьи (замена Markdown -> WYSIWYG). Согласно `feature-flow.md` Rule 11, UC-001 должен быть обновлён до closure. Не блокирует Design Ready, но необходимо учесть в плане.

### 6. `architecture.md` не обновлён

ADR-007 явно указывает в "Follow-up" и "Consequences -> Neutral": `memory-bank/domain/architecture.md` должен быть обновлён (зафиксировать, что `articles.body` хранит HTML). При текущем `decision_status: proposed` это можно отложить, но нужно оформить как `OQ-*` в будущем `implementation-plan.md`.

---

## Резюме

| Приоритет | Что сделать |
|---|---|
| Критично — сейчас | Убрать `OQ-01` из `feature.md` -> заменить на prose / `ASM-*` / `CON-*` |
| Критично — сейчас | Исправить circular `derived_from` в `ADR-007` (убрать `feature.md` из его `derived_from`) |
| До Design Ready | Добавить в `glossary.md`: WYSIWYG, TipTap, bluemonday, allowlist, safeHTML |
| До Closure | Обновить `UC-001` traceability — добавить FT-023 |
| Уточнить | Статус `ADR-007`: обоснованно `draft` или перевести в `active` |
