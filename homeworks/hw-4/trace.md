# Трасса запуска: Feature Execution Loop — FT-023

**Фича:** FT-023 WYSIWYG-редактор статей
**Период:** 2026-04-28 — 2026-04-29
**Финальный статус:** `blocked` (implementation pending; артефакты готовы, ждут старта Execution)

> Трасса ретроспективна: прогон выполнялся до формализации feature-execution-loop.md,
> но точно соответствует его этапам. Материал — `.review-results/FT-023/`.

---

## Этапы прогона

### Этап 1–2: Brief + Spec Improve Loop (2026-04-28)

На момент прогона малые циклы не были формализованы как отдельные процессы.
Evaluator agent запускался вручную через Agent tool — функционально эквивалентно `improve-loop.sh`.

**Итерация 1 — `revise`**
Файл: `.review-results/FT-023/review-FT-023-feature.md`

Критические замечания:
- OQ-* идентификаторы в `feature.md` — нарушение identifier taxonomy (OQ-* принадлежат только плану)
- Sanity-check: `DEC-01` (ADR-007 proposed) не зафиксирован как hypothesis

**Итерация 2 — `revise`**
Файл: `.review-results/FT-023/review-FT-023-feature-1.md` (2026-04-28)

Критические замечания:
- Неверные пути в Change Surface: `web/templates/article/edit.html` → реальный путь `templates/admin/articles/edit.html`

**Итерация 3 — `revise`**
Файл: `.review-results/FT-023/review-FT-023-feature-2.md`

Критические замечания:
- Санитизация (bluemonday) отнесена к Handler-слою вместо Service-слоя (нарушение architecture.md Layer Stack)
- CSRF (PCON-02) не зафиксирован через ASM-* — молчаливое допущение

**⛔ STOP между сессиями.** После 3 итераций — пауза. Следующая сессия возобновила с исправленным feature.md.

**Итерация 4 — `accept`**
Файл: `.review-results/FT-023/review-FT-023-feature-3.md`

Все блокеры устранены. EVID-* добавлен в feature.md.

---

### Этап 3: HITL — Design Ready gate (2026-04-28)

Feature.md показан человеку. Получено подтверждение перехода в Design Ready.
`feature.md → status: active`, `delivery_status: in_progress`.

---

### Этап 4: Implementation Plan + eval DR→PR (2026-04-28 — 2026-04-29)

**Итерация 1 — `revise`**
Файл: `.review-results/FT-023/review-FT-023-plan.md` (2026-04-28)

Критические замечания:
- Environment Contract: команда integration-тестов без `--network lfcru_forum_default` и `-e DATABASE_URL`
- E2E prerequisite неполный: не указан `docker-compose.dev.yml` как обязательный dev-стек перед e2e

**Итерация 2 — `revise`**
Файл: `.review-results/FT-023/review-FT-023-plan-1.md` (2026-04-29)

Критические замечания:
- `status: draft` в плане при Plan Ready gate — mixed-up гейты (Plan Ready vs Execution)
- Sequencing: STEP-00 смешивает действия двух разных gates

**⛔ STOP между сессиями.** Замечания зафиксированы в `.review-results/`, следующая сессия возобновила с исправленным планом.

**Итерация 3 — `revise`**
Файл: `.review-results/FT-023/review-FT-023-plan-2.md` (2026-04-29)

Замечания уровня HIGH:
- `status: draft` — gate ambiguity сохранилась частично
- AG-* на UI-verification: нарушение autonomy-boundaries.md (Playwright — автопилот, не требует AG-*)

**Итерация 4 — `revise` → финальная правка**
Файл: `.review-results/FT-023/review-FT-023-plan-3.md` (2026-04-29)

**Итерация 5 — `accept`**
Файл: `.review-results/FT-023/review-FT-023-plan-4.md` (2026-04-29)

Все блокеры устранены. EVID-* добавлен в feature.md.

---

### Этап 5: HITL — Plan Ready gate

Ожидает явного подтверждения человека. **Текущий статус прогона: `blocked` (awaiting-human).**

---

## Stage Log (фактический)

| Stage        | Status   | Outcome      | Date       | Ref |
|-------------|----------|--------------|------------|-----|
| brief-loop  | done     | accept (4 iter) | 2026-04-28 | `.review-results/FT-023/review-FT-023-feature-3.md` |
| spec-loop   | done     | accept (4 iter) | 2026-04-28 | `.review-results/FT-023/review-FT-023-feature-3.md` |
| dr-approval | done     | approved     | 2026-04-28 | — |
| plan        | done     | accept (5 iter) | 2026-04-29 | `.review-results/FT-023/review-FT-023-plan-4.md` |
| pr-approval | blocked  | awaiting     | —          | — |
| impl        | pending  | —            | —          | — |
| unit-tests  | pending  | —            | —          | — |
| e2e-smoke   | pending  | —            | —          | — |
| verification| pending  | —            | —          | — |
| closure     | pending  | —            | —          | — |

---

## Stop / Resume

| Событие | Причина | Возобновление |
|---|---|---|
| Stop после итерации 3 brief/spec loop | 3+ revise, требовалась правка feature.md | Следующая сессия прочитала HANDOFF.md → продолжила с исправленным feature.md |
| Stop после итерации 1 plan loop | замечания зафиксированы, требовалась правка плана | Следующая сессия (2026-04-29) возобновила с review-FT-023-plan-1.md как контекстом |
| Stop на pr-approval | HITL gate — ждём подтверждения человека | — (текущее состояние) |

---

## Наблюдения

1. **Малые циклы работают.** 4 итерации для feature.md и 5 для плана — норма для large.md с ADR-зависимостью. Каждая итерация убирала конкретный класс ошибок.

2. **Наиболее частые ошибки в brief/spec:** неверные пути в Change Surface, нарушение Layer Stack (Handler vs Service), молчаливые допущения по CSRF.

3. **Наиболее частые ошибки в плане:** неполный Environment Contract (docker-команды), смешение gate-предикатов, AG-* на автопилотных действиях.

4. **State между сессиями работал:** HANDOFF.md + `.review-results/` позволяли следующей сессии точно знать, с какого ревью возобновить.
