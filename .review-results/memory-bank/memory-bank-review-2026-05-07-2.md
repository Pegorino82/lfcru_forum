# Memory-Bank Review — Консистентность и однозначность

| Meta | |
|---|---|
| Дата | 2026-05-07 |
| Модель | claude-opus-4-6 |
| Время ревью | 36m 27s |
| Охват | workflows.md, feature-flow.md, feature-execution-loop.md, eval.md, trello.md, autonomy-boundaries.md, git-workflow.md, coding-style.md, testing-policy.md, bugfix-execution-loop.md, .promts/create-feature-from-trello-card.md, domain/architecture.md, domain/problem.md, ops/development.md |
| Фокус | противоречия, неоднозначности, битые ссылки, дублирование с расхождением |

---

## Находки

### 1. [противоречие]
**Файлы:** `.promts/create-feature-from-trello-card.md` vs `git-workflow.md`, `workflows.md`, `bugfix-execution-loop.md`
**Суть:** Bug fix создаётся на ветке `feat/FT-XXX-slug` вместо `fix/FT-XXX-slug`
**Цитата A (create-feature-from-trello-card.md, Шаг 1, строка 50):**
> `git worktree add ../lfcru_forum-FT-XXX -b feat/FT-XXX-slug`

Эта команда выполняется ДО routing по label карточки (строка 70). Таким образом для bug fix тоже создаётся ветка с префиксом `feat/`.

**Цитата B (git-workflow.md, строка 34):**
> | Баг-фикс | `fix/FT-XXX-slug` |

**Цитата C (workflows.md, строка 85):**
> баг-фикс ведётся в git worktree на ветке `fix/FT-XXX-slug`

**Цитата D (bugfix-execution-loop.md, строка 72):**
> `git branch -d fix/FT-XXX-slug`

**Вопрос к владельцу:** Промпт должен разделять создание worktree на два пути (feat/ vs fix/) в зависимости от label, или routing по label нужно вынести до создания worktree?

---

### 2. [противоречие]
**Файлы:** `feature-flow.md` — frontmatter vs body
**Суть:** Frontmatter запрещает определять `trello_planning_timing`, но body определяет его inline
**Цитата A (feature-flow.md, frontmatter, строки 19):**
> `must_not_define: trello_planning_timing`

**Цитата B (feature-flow.md, строка 81):**
> если задача зафиксирована в task tracker → карточка переведена в статус PLANNING **немедленно при получении задачи** — до чтения файлов, до любых вопросов и до обсуждения (canonical: `trello.md`)

**Вопрос к владельцу:** Тег `(canonical: trello.md)` подразумевает ссылку, а не определение, но формулировка с конкретным таймингом выглядит как inline-определение. Оставить ссылку без деталей или убрать из `must_not_define`?

---

### 3. [дублирование с расхождением]
**Файлы:** `workflows.md` vs `testing-policy.md` vs `feature-execution-loop.md` vs `create-feature-from-trello-card.md`
**Суть:** Три разных указателя на источник Docker-команды для запуска тестов, плюс inline-копия команды
**Цитата A (workflows.md, строка 135):**
> Запусти unit-тесты (Docker-командой из [testing-policy.md](../engineering/testing-policy.md))

**Цитата B (testing-policy.md, строка 32):**
> **Актуальные команды запуска** (единственный источник) — [`ops/development.md`](../ops/development.md) § «Go-тесты». Использовать дословно.

**Цитата C (feature-execution-loop.md, строки 141-146):**
> ```docker run --rm -v "$(pwd)":/app ... golang:1.23-alpine go test ./...```

**Цитата D (create-feature-from-trello-card.md, строка 197):**
> Запусти unit-тесты локально командой из `memory-bank/ops/development.md` § «Go-тесты»

Цепочка ссылок: `workflows.md` → `testing-policy.md` → `ops/development.md` (canonical). Но `feature-execution-loop.md` имеет inline-копию команды. Если команда в `ops/development.md` изменится, inline-копия не обновится автоматически.

**Вопрос к владельцу:** Убрать inline-команду из `feature-execution-loop.md` и заменить ссылкой на `ops/development.md`? Или обновить `workflows.md` чтобы он ссылался напрямую на `ops/development.md`?

---

### 4. [неоднозначность]
**Файлы:** `bugfix-execution-loop.md` vs `workflows.md`
**Суть:** Session Protocol (unit-тесты, simplify review) и bugfix execution loop (стадии tests, verification) пересекаются — неясно, когда именно тесты запускаются
**Цитата A (bugfix-execution-loop.md, строка 58):**
> **Session Protocol** (`workflows.md` § «В конце сеанса») выполняется **дополнительно** к стадии `fix` — оба протокола действуют, не заменяют друг друга.

**Цитата B (workflows.md, Session Protocol, строка 135-136):**
> 1. Запусти unit-тесты ... убедись, что зелёные.
> 2. Simplify review ...

Bugfix loop имеет отдельную стадию `tests` (после `fix`) и `closure`. Session Protocol тоже включает unit-тесты и simplify review. Если Session Protocol применяется «дополнительно к стадии fix», агент может запустить тесты дважды: один раз по Session Protocol на стадии fix, и ещё раз на стадии `tests`.

**Вопрос к владельцу:** Session Protocol для bugfix loop следует применять только на стадии `closure` (а не `fix`)? Или тесты из Session Protocol заменяют стадию `tests` bugfix loop?

---

### 5. [неоднозначность]
**Файлы:** `bugfix-execution-loop.md` vs `testing-policy.md`
**Суть:** Bugfix execution loop не упоминает Playwright/E2E для багфиксов с UI-изменениями
**Цитата A (bugfix-execution-loop.md, стадия verification, строка 55):**
> Вручную проверить что баг не воспроизводится, заполнить EVID-1

**Цитата B (testing-policy.md, строки 107-108):**
> Browser-специфика и HTMX/Alpine.js-взаимодействия — **покрываются Playwright**, не являются основанием для manual-only.
> UI-изменения обязаны пройти Playwright-верификацию

Bugfix loop не содержит стадии E2E smoke (в отличие от feature execution loop, этап 8). Для UI-багфиксов `testing-policy.md` требует Playwright, но `bugfix-execution-loop.md` этого не предусматривает.

**Вопрос к владельцу:** Добавить Playwright-проверку в bugfix loop для UI-багов? Или считать что CI (включающий E2E) покрывает это требование?

---

### 6. [дублирование с расхождением]
**Файлы:** `eval.md` — секция Gate Checklists vs секция Evaluator Agent Protocol
**Суть:** Gate DR→PR в чеклисте не упоминает исключение для малых планов (≤ 3 STEP-*), хотя Protocol и feature-flow.md его определяют
**Цитата A (eval.md, Gate Checklists DR→PR, строка 132):**
> Форма: **evaluator agent** для `large.md`; self-check для `short.md`.

**Цитата B (eval.md, Evaluator Agent Protocol, строка 69):**
> `large.md`, gate DR → PR для малых планов (≤ 3 STEP-*) → self-check допустим.

**Цитата C (feature-flow.md, строка 126):**
> Для `large.md` с планом ≤ 3 STEP-* — self-check допустим.

**Вопрос к владельцу:** Добавить примечание в Gate Checklists DR→PR секцию eval.md об исключении для ≤ 3 STEP-*?

---

### 7. [неоднозначность]
**Файлы:** `workflows.md`
**Суть:** Термин «облегчённый FT-пакет» для багфикса не содержит ссылки на определение его состава
**Цитата A (workflows.md, строки 87-88):**
> **FT-пакет:** облегчённый. Создаётся без `feature.md` и `implementation-plan.md`.

Описание говорит что НЕ входит, но не говорит что входит (README.md? run-state?). Нет ссылки на `bugfix-execution-loop.md`, где состав определён полностью.

**Цитата B (workflows.md, строка 91):**
> Execution loop (state-pack, stages, resume protocol): [`bugfix-execution-loop.md`](bugfix-execution-loop.md).

Ссылка на execution loop есть, но она стоит отдельно от описания FT-пакета.

**Вопрос к владельцу:** Добавить ссылку на `bugfix-execution-loop.md` § «FT-пакет баг-фикса» прямо в описание облегчённого пакета?

---

## Сводная таблица

| # | Тип | Файлы | Серьёзность |
|---|-----|-------|-------------|
| 1 | противоречие | `create-feature-from-trello-card.md` vs `git-workflow.md`, `workflows.md`, `bugfix-execution-loop.md` | high |
| 2 | противоречие | `feature-flow.md` frontmatter vs body | med |
| 3 | дублирование | `workflows.md` vs `testing-policy.md` vs `feature-execution-loop.md` vs `create-feature-from-trello-card.md` | med |
| 4 | неоднозначность | `bugfix-execution-loop.md` vs `workflows.md` | med |
| 5 | неоднозначность | `bugfix-execution-loop.md` vs `testing-policy.md` | med |
| 6 | дублирование | `eval.md` Gate Checklists vs Protocol | low |
| 7 | неоднозначность | `workflows.md` — «облегчённый FT-пакет» | low |
