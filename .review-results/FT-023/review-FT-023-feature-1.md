# Ревью FT-023 Feature Package

**Дата:** 2026-04-28
**Ревьюируемые файлы:**
- `memory-bank/features/FT-023/feature.md`
- `memory-bank/features/FT-023/README.md`
- `memory-bank/adr/ADR-007-wysiwyg-editor-html-storage.md`

**Статус ревью:** Draft stage — 2 нарушения, 2 предупреждения

---

## Нарушения (блокируют переход в Design Ready)

### [BLOCKER-1] FT-023 не зарегистрирован в `features/README.md`

`features/README.md` содержит записи только до FT-021. FT-022 и FT-023 отсутствуют.

По правилу `feature-flow.md` §Package Rules p.4: "README.md создается вместе с feature.md". Регистрация в индексе является частью bootstrap gate-чека.

**Действие:** добавить строку в таблицу `Packages` в `memory-bank/features/README.md`:

```
| [FT-023](FT-023/) | WYSIWYG-редактор статей | `planned` |
```

---

### [BLOCKER-2] Противоречие между ADR-007 и `feature.md` NS-03

`feature.md` NS-03:
> Миграция существующих статей с Markdown на HTML — отдельная задача; вне scope данной фичи.

ADR-007, секция «Риски и mitigation»:
> Существующие Markdown-статьи сломаются | FT-023 scope включает миграцию или fallback-рендер; OQ зафиксирован в implementation-plan.md

`feature.md` — canonical owner scope. ADR-007 противоречит ему. По `feature-flow.md` Boundary Rule 6: "Если меняются scope... сначала обновляется feature.md или ADR, потом downstream".

**Действие:** в ADR-007 в колонке Mitigation для риска "Существующие Markdown-статьи сломаются" заменить текущее описание на:

> FT-023 явно исключает миграцию (NS-03); деградация существующих статей — ожидаемое поведение согласно ASM-03 до выполнения отдельной миграционной задачи.

---

## Предупреждения (нужно устранить до Design Ready gate)

### [WARN-1] OQ-01 использован как ссылка в `feature.md`

`feature.md` CTR-01 содержит: *"Исторические статьи — OQ-01"*.

По `feature-flow.md` Stable Identifiers: `OQ-*` — Plan IDs, они определяются и существуют только в `implementation-plan.md`. В `feature.md` идентификатор `OQ-01` не объявлен и не может быть определён. Ссылка ведёт в никуда.

**Действие:** в `feature.md` CTR-01 заменить ссылку `OQ-01` на prose-описание или отсылку к уже существующему `ASM-03`.

---

### [WARN-2] `architecture.md` не обновлён под HTML-хранение body

ADR-007, секция «Нейтральные / организационные»:
> `memory-bank/domain/architecture.md` необходимо обновить: зафиксировать, что `articles.body` хранит HTML.

Текущий `architecture.md` не содержит записи о формате поля `articles.body`. Поскольку `architecture.md` — canonical input для реализации, обновление нужно выполнить до Design Ready (не откладывать до Plan Ready).

**Действие:** добавить в `memory-bank/domain/architecture.md` в раздел Module Boundaries или отдельным подразделом: формат хранения `articles.body = sanitized HTML` (после перевода ADR-007 в `accepted`).

---

## Соответствие (всё в норме)

| Аспект | Результат |
|---|---|
| Lifecycle stage: `status: draft, delivery_status: planned`, `implementation-plan.md` отсутствует | OK |
| Шаблон: multiple REQ/NS/ASM/CON/CTR/FM/EC/SC/NEG/CHK/EVID → `large.md` | OK |
| Traceability matrix: каждый `REQ-*` привязан к `SC-*`, `CHK-*`, `EVID-*` | OK |
| NEG-01 присутствует (XSS через `<iframe>`) — покрывает CON-01 | OK |
| ADR-dependency handling: CON-02 фиксирует `proposed`-статус ADR-007 как hypothesis | OK |
| ADR-007: зарегистрирован в `adr/README.md`, ≥2 варианта, follow-up, negative consequences | OK |
| Frontend stack: TipTap как vanilla JS — соответствует требованию "без React/Vue" из `frontend.md` | OK |
| Layer stack: bluemonday в handler — соответствует `architecture.md` | OK |
| XSS mitigation: `template.HTML` + bluemonday allowlist | OK |

---

## Итог

| # | Тип | Описание | Файл |
|---|---|---|---|
| 1 | BLOCKER | FT-023 не в `features/README.md` | `memory-bank/features/README.md` |
| 2 | BLOCKER | ADR-007 mitigation противоречит NS-03 | `memory-bank/adr/ADR-007-wysiwyg-editor-html-storage.md` |
| 3 | WARN | `OQ-01` — невалидный идентификатор в `feature.md` | `memory-bank/features/FT-023/feature.md` |
| 4 | WARN | `architecture.md` не отражает HTML-хранение body | `memory-bank/domain/architecture.md` |
