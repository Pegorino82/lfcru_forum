---
title: "FT-024: Design Ready Eval"
doc_kind: feature
doc_function: gate-eval
ft_id: FT-024
gate: "Draft→Design Ready"
status: closed
date: 2026-05-03
audience: humans_and_agents
---

# FT-024: Design Ready Eval

## Checklist

### A. Идентификаторы и структура
- [x] A-1: Нет Plan IDs в feature.md — OK
- [x] A-2: REQ-*, NS-*, SC-*, CHK-*, EVID-* присутствуют — OK
- [x] A-3: Все идентификаторы определены явно — OK
- [x] A-4: large.md выбран обоснованно (множество REQ-*, SC-*, FM-*, контракты) — OK
- [x] A-5: Запись FT-024 в features/README.md, delivery_status: planned — OK

### B. Traceability
- [x] B-1: Каждый REQ-* → ≥1 SC-* в traceability matrix — OK
- [x] B-2: Каждый SC-* → CHK-01 — OK
- [x] B-3: CHK-01→EVID-01, CHK-02→EVID-02 — OK
- [x] B-4: NEG-01..NEG-05 присутствуют для критичных FM-* — OK

### C. Непротиворечивость
- [x] C-1: Нет противоречий с ADR-005 — OK
- [x] C-2: Нет циклических derived_from — OK
- [x] C-3: ADR-005 accepted → используется как canonical input — OK
- [x] C-4: NS-* и REQ-* не конфликтуют; ASM-* и CON-* согласованы — OK

### D. Соответствие архитектуре
- [x] D-1: Все пути в Change Surface существуют — OK (после исправлений)
- [x] D-2: Шаблоны templates/<domain>/ — OK
- [x] D-3: static/css/ удалён, используются inline стили — OK
- [x] D-4: Бизнес-логика в Service-слое — OK

### E. Тестовая политика
- [x] E-1: UI-изменения → Playwright (CHK-01, CHK-02) — OK
- [x] E-2: HTMX/Alpine.js покрыты Playwright — OK
- [x] E-3: EVID-* producer — CI/Playwright — OK
- [x] E-4: Нет manual-only gaps — OK

### F. Системные ограничения
- [x] F-1: CSRF зафиксирован в CON-01 для POST /profile/avatar — OK
- [x] F-2: Авторизация через middleware зафиксирована явно (CON-02) — OK

### G. Glossary и Use Case
- [x] G-1: quick-view и relative time добавлены в glossary.md — OK
- [x] G-2: Фича не меняет существующий UC — UC не требуется

## Iterations

| # | Date | Outcome | Findings |
|---|---|---|---|
| 1 | 2026-05-03 | revise | 4 замечания: C-1 (CON-04 в REQ-05), D-1 (static/css/ несуществующий), D-1 (неверные имена templates/forum), G-1 (quick-view, relative time в glossary) |
| 2 | 2026-05-03 | revise | 1 замечание HIGH: D-1 (templates/news/show.html → article.html) |
| 3 | 2026-05-03 | accept | Self-check на trivial path fix (лимит 2 evaluator итерации исчерпан) |

## Decision

**Outcome:** accept
**Date:** 2026-05-03
**EVID:** EVID-03 (в feature.md)
