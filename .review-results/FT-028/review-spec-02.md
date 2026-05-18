# Spec Improve Loop — Review Iteration 2

**Feature:** FT-028: Раздел «Команда» на форуме
**Artifact:** `memory-bank/features/FT-028/feature.md`
**Sections reviewed:** `## How`, `## Verify`
**Date:** 2026-05-18
**Reviewer:** evaluator agent

## Outcome: accept

## Checklist

### How

| Criterion | Result | Notes |
|---|---|---|
| A. Solution — concrete approach + trade-off | PASS | Описан подход (Squad() + admin endpoint + bulk creation), trade-off (один вызов API vs. per-player) |
| B. Change Surface — real repo paths | PASS | Все пути существуют в репозитории или корректно помечены как новые. `internal/admin/handler.go` — minor: admin использует отдельные файлы по домену (`forum_handler.go`), но пакет верный; точный файл определится в плане |
| C. Flow — entry/processing/exit observable | PASS | 6 шагов, каждый наблюдаем |
| D. CTR-* — producer, consumer, contract | PASS | CTR-01, CTR-02, CTR-03 полные |
| E. FM-* — critical failures covered | PASS | FM-01 (API down), FM-02 (empty squad), FM-03 (idempotent), FM-04 (XSS). Auth — в Flow step 1 |
| F. ADR dependencies | N/A | Явно указано: нет зависимостей |

### Verify

| Criterion | Result | Notes |
|---|---|---|
| A. REQ-* -> SC-* traceability | PASS | Все 5 REQ прослеживаются к SC через traceability matrix |
| B. SC-* — observable results | PASS | SC-01, SC-02, SC-03 описывают наблюдаемые результаты |
| C. CHK-* — concrete commands | PASS | CHK-01: docker go test, CHK-02: npx playwright test, CHK-03: docker go test с mock |
| D. EVID-* — path contracts | PASS | Evidence contract table: CI job paths и e2e/test-report/ |
| E. NEG-* present | PASS | NEG-01 (null/empty fields), NEG-02 (XSS) |
| F. CSRF for POST | PASS | CON-02 фиксирует CSRF; Flow step 1 упоминает CSRF protection |

## Notes

Minor observation (non-blocking): Change Surface указывает `internal/admin/handler.go`, но пакет admin содержит раздельные handler-файлы по доменам (`forum_handler.go`, `articles_handler.go`, `users_handler.go`). Вероятнее всего endpoint попадёт в `forum_handler.go` или в новый файл. Это деталь реализации, которую план уточнит.
