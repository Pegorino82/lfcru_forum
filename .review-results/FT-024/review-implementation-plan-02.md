---
gate: "Design Ready → Plan Ready"
artifact: "memory-bank/features/FT-024/implementation-plan.md"
date: "2026-05-03"
outcome: revise
---

# Результат ревью: implementation-plan.md (FT-024) — итерация 2

**Gate:** Design Ready → Plan Ready
**Artifact:** `memory-bank/features/FT-024/implementation-plan.md`
**Date:** 2026-05-03
**Outcome:** revise

---

## Статус исправлений из review-01

Все блокеры и высокие замечания из итерации 1 устранены:

- Environment Contract: команда unit-тестов скопирована дословно из `ops/development.md` (`go test ./...`) ✓
- Test Strategy: integration тесты в `Required local suites` → `—` ✓
- E2E prerequisite: оба шага (dev-stack + e2e-stack) явно указаны в правильном порядке ✓
- Discovery Context / PRE-04: путь к middleware исправлен на `internal/auth/middleware.go` ✓
- STEP-00 добавлен как HARD STOP для lifecycle-переходов ✓
- Test assertions конкретизированы (struct fields, HTTP body) ✓
- E2E seeding описан явно ✓
- STOP-04 добавлен для ER-03 ✓
- STEP-10 (Simplify Review) добавлен явно ✓

---

## Новые замечания

### 1. MEDIUM — D-1: Команда unit-тестов в Test Strategy не совпадает с каноничной

**Цитата из плана** (секция `Test Strategy`, строка `FuncMap avatarColor/avatarInitials/relativeTime`, колонка `Required local suites`):
```
docker run --rm -v "$(pwd)":/app -w /app -v lfcru_gomod:/root/go/pkg/mod golang:1.23-alpine go test ./internal/tmpl/...
```

**Норма:** `memory-bank/ops/development.md` § «Running Tests» — unit-команда:
```bash
docker run --rm \
  -v "$(pwd)":/app -w /app \
  -v lfcru_gomod:/root/go/pkg/mod \
  golang:1.23-alpine \
  go test ./...
```

**Норма:** `memory-bank/engineering/testing-policy.md` § «Stack → Go-тесты»:
> «Актуальные команды запуска (единственный источник) — `ops/development.md` § «Go-тесты». Использовать дословно.»
> «⛔ Не изобретать `docker run` вручную.»

**Нарушение:** в Test Strategy указана команда `go test ./internal/tmpl/...` — это пакетная вариация, которой нет в `ops/development.md`. Единственный каноничный вариант unit-команды — `go test ./...`. Отклонение нарушает требование дословного копирования (D-1).

**Требуемое исправление:** в колонке `Required local suites` для строки FuncMap заменить команду на каноничную: `docker run --rm -v "$(pwd)":/app -w /app -v lfcru_gomod:/root/go/pkg/mod golang:1.23-alpine go test ./...`.

---

### 2. MEDIUM — G-1: ER-04 не имеет соответствующего STOP-*

**Цитата из плана** (секция `Execution Risks`, строка `ER-04`):
```
ER-04 | Playwright E2E: upload тест требует реального файла и запущенного app-e2e контейнера |
если контейнер не поднят, тест падает с connection refused |
в CI: docker-compose.e2e.yml поднимается в workflow; локально: инструкция в ops/development.md |
npx playwright test connection refused
```

**Имеющиеся STOP-*:** STOP-01 (ER-01), STOP-02 (ER-02), STOP-03 (FM-03), STOP-04 (ER-03). ER-04 не покрыт.

**Норма:** `memory-bank/engineering/autonomy-boundaries.md` § «Правило эскалации»:
> «Если замечания не уменьшаются после 2–3 итераций — проблема upstream, а не в коде. Остановись и предложи вернуться на предыдущий этап.»

Требование проверки G-1: «Каждый ER-* имеет соответствующий STOP-* или явный escalation threshold ("N итераций → остановись и эскалируй"). Без этого агент не знает, когда прекратить retry.»

**Требуемое исправление:** добавить `STOP-05` для ER-04: trigger — `npx playwright test` падает с connection refused при запуске E2E; немедленное действие — убедиться, что оба стека подняты (`docker compose -f docker-compose.dev.yml up -d` + `docker compose -f docker-compose.e2e.yml up -d`); если после рестарта стеков ошибка сохраняется — остановиться и эскалировать человеку.

---

## Итог

**Outcome: revise**

Средние (2 пункта):
1. Test Strategy: `Required local suites` для FuncMap строки содержит пакетную вариацию команды (`./internal/tmpl/...`) вместо каноничной (`./...`) из `ops/development.md` — нарушение D-1.
2. ER-04 не имеет соответствующего STOP-* — нарушение G-1.

Блокеров и высоких замечаний нет.
