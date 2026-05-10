# Memory-Bank Review — Консистентность и однозначность

| Meta | |
|---|---|
| Дата | 2026-05-07 |
| Модель | claude-opus-4-6 |
| Охват | ядровые flow-документы, engineering-конвенции, промпт `create-feature-from-trello-card.md` |
| Фокус | противоречия, неоднозначности, битые ссылки, дублирование с расхождением |

---

## Найденные проблемы

---

### 1. [противоречие] improve-loop.sh vs Agent tool для запуска improve-loops

**Файлы:** `flows/feature-execution-loop.md` vs `flows/brief-improve-loop.md`, `flows/spec-improve-loop.md`, `.promts/create-feature-from-trello-card.md`

**Суть:** `feature-execution-loop.md` явно запрещает `improve-loop.sh` в сессии агента и требует Agent tool. Три других документа предписывают запуск через скрипт.

**Цитата A (feature-execution-loop.md:67):**
> **Агент:** запускать через **Agent tool** с инстанцированным промптом из `memory-bank/flows/templates/prompts/brief-loop.md`. **Не использовать** `improve-loop.sh` — скрипт вызывает `claude --print` изнутри сессии, что приводит к зависанию.

**Цитата B (create-feature-from-trello-card.md:165-168):**
> ```bash
> ./scripts/improve-loop.sh \
>   memory-bank/flows/templates/prompts/brief-loop.md \
>   memory-bank/features/FT-XXX/feature.md
> ```

**Цитата C (brief-improve-loop.md:71-74, Runner Contract):**
> ```bash
> ./scripts/improve-loop.sh \
>   memory-bank/flows/templates/prompts/brief-loop.md \
>   memory-bank/features/FT-XXX/feature.md
> ```

**Решение владельца:** Agent tool — canonical-метод для сессии агента. Скрипт `improve-loop.sh` — только для ручного/CI запуска человеком. Добавить комментарий в скрипт об этом. Обновить Runner Contract в `brief-improve-loop.md`, `spec-improve-loop.md` и промпт.

---

### 2. [противоречие] Eval DR gate для large.md — нужен ли evaluator agent помимо loops?

**Файлы:** `flows/eval.md` vs `flows/feature-flow.md` vs `flows/feature-execution-loop.md`

**Суть:** `eval.md` говорит, что для DR gate достаточно brief-loop + spec-loop (evaluator agent не нужен). `feature-flow.md` требует дополнительный evaluator agent (`review-feature-md.md`) для `large.md`. `feature-execution-loop.md` ссылается на `eval.md` для DR evaluator, но `eval.md` его не определяет.

**Цитата A (eval.md:117-121):**
> ### Draft → Design Ready
> Форма: **brief-loop + spec-loop (evaluator agents)**. Self-check не нужен — все критерии этого gate покрыты loops: REQ-\*/NS-\* → brief-loop, SC-\*/CHK-\*/EVID-\* → spec-loop.

**Цитата B (feature-flow.md:93-101):**
> **Feature.md review (для `large.md`):** перед показом человеку запусти evaluator agent через **Agent tool**:
> 1. Инстанциируй шаблон `memory-bank/flows/templates/prompts/review-feature-md.md` — замени `{{FT_ID}}`, `{{FEATURE_PATH}}`, `{{DATE}}`
> ...
> Для `short.md` — brief-loop + spec-loop (evaluator agents) достаточны

**Цитата C (feature-execution-loop.md:83-84):**
> Для `large.md`: запустить evaluator agent (DR gate из `eval.md`) → если accept, evaluator создаёт `evals/DR-eval.md`

**Решение владельца:** для `large.md` DR gate — loops + evaluator agent (`review-feature-md.md`). Обоснование: loops проверяют секции изолированно, evaluator читает весь документ и ловит кросс-секционные разрывы (How реализует What? Verify покрывает все REQ-*?). Для `short.md` — loops достаточны. Обновить `eval.md`: добавить `large.md` DR gate в Evaluator Agent Protocol и в Gate Чеклисты.

---

### 3. [противоречие] `gh repo view` в worktree — использовать или нет?

**Файлы:** `.promts/create-feature-from-trello-card.md` vs `engineering/git-workflow.md`

**Суть:** промпт предписывает `gh repo view` как шаг верификации перед `gh pr create`. Git workflow явно запрещает это.

**Цитата A (create-feature-from-trello-card.md:60):**
> ```bash
> gh repo view  # убедиться что контекст — Pegorino82/lfcru_forum
> ```

**Цитата B (git-workflow.md:89):**
> `gh repo view` в worktree-директории ненадёжен — может вернуть чужой репозиторий. Не использовать как проверку перед `gh pr create`. Флаг `--repo Pegorino82/lfcru_forum` достаточен.

**Решение владельца:** убрать `gh repo view` из промпта. `git-workflow.md` — canonical; флаг `--repo` достаточен.

---

### 4. [дублирование с расхождением] evals/ directory отсутствует в Bootstrap промпта

**Файлы:** `flows/feature-flow.md` vs `.promts/create-feature-from-trello-card.md`

**Суть:** `feature-flow.md` Bootstrap gate требует создание папки `evals/` с `strategy.md` и `summary.md`. Промпт Bootstrap (шаги 1-4 для feature) не упоминает `evals/`.

**Цитата A (feature-flow.md:90):**
> - [ ] создана папка `evals/` с `strategy.md` и `summary.md` по шаблонам из `templates/feature/evals/` (внутри worktree); `strategy.md` заполнен формами для каждого gate согласно типу фичи

**Цитата B (create-feature-from-trello-card.md:153-162):** шаги Bootstrap для feature перечисляют `README.md`, `feature.md`, `run-state/` — но не `evals/`.

**Решение владельца:** добавить создание `evals/` в Bootstrap промпта — явно, без разночтений.

---

### 5. [противоречие] Тайминг PLANNING — «немедленно при получении» vs «до файловых операций»

**Файлы:** `flows/trello.md` vs `flows/feature-flow.md`

**Суть:** `trello.md` и промпт: PLANNING выставляется **немедленно** при получении задачи, **до чтения файлов и до обсуждения**. `feature-flow.md`: PLANNING — «до любых файловых операций» (в контексте Bootstrap), что допускает чтение файлов и обсуждение перед перемещением.

**Цитата A (trello.md:89):**
> **HARD STOP — PLANNING:** карточка переводится в PLANNING **немедленно при получении задачи** — до чтения файлов, до любых вопросов и до обсуждения.

**Цитата B (feature-flow.md:76):**
> - [ ] если задача зафиксирована в task tracker → карточка переведена в статус "обсуждается" (PLANNING) до любых файловых операций

**Решение владельца:** canonical тайминг — из `trello.md`: немедленно при получении задачи (сигнал-замок на доске). Подтянуть формулировку в `feature-flow.md` до того же уровня строгости.

---

### 6. [дублирование с расхождением] docs-коммит протокол не отражён в промпте и execution loop

**Файлы:** `flows/workflows.md` vs `.promts/create-feature-from-trello-card.md` vs `flows/feature-execution-loop.md`

**Суть:** `workflows.md` Session Protocol (шаг 4) требует отдельный `docs:`-коммит с хешем после fix/feat-коммита. Ни промпт, ни `feature-execution-loop.md` не упоминают этот шаг.

**Цитата A (workflows.md:139):**
> 4. Получи хеш коммита (`git log --oneline -1`) и впиши его в документацию: `HANDOFF.md` и FT-пакет README (если есть). Сделай отдельный `docs:`-коммит.

**Цитата B:** `create-feature-from-trello-card.md` bug-fix stages (analysis→fix→tests→verification→closure) и feature execution — нет упоминания docs-коммита.

**Решение владельца:** docs-коммит протокол обязателен. Добавить системно в промпт и execution loop — хеши должны записываться по протоколу, а не ситуативно.

---

### 7. [неоднозначность] run-state/ для фичей — не описан в workflows.md

**Файлы:** `flows/workflows.md` vs `flows/feature-execution-loop.md` vs `.promts/create-feature-from-trello-card.md`

**Суть:** `workflows.md` описывает `run-state/` только для баг-фикса (строки 82-83). `feature-execution-loop.md` и промпт используют `run-state/` и для фичей. `workflows.md` не ссылается на `feature-execution-loop.md` и не упоминает state-pack для feature workflow.

**Цитата A (workflows.md:82-83):**
> - `run-state/FT-XXX/active-context.md` — текущая стадия и чекпоинты (тип: bugfix)
> - `run-state/FT-XXX/stage-log.md` — 5 стадий: analysis → fix → tests → verification → closure

**Цитата B (feature-execution-loop.md:27-33):** State-pack таблица описывает `run-state/` для фичей — `active-context.md`, `stage-log.md`, `HANDOFF.md`.

**Решение владельца:** вынести баг-фикс execution в отдельный документ (аналог `feature-execution-loop.md` для багов). `workflows.md` остаётся чистым router'ом для всех типов.

---

### 8. [неоднозначность] Нарушение canonical_for границ между документами

**Файлы:** все flow-документы

**Суть:** четыре документа описывают пересекающиеся аспекты одного flow: `workflows.md` (routing + session protocol), `feature-flow.md` (gates и predicates), `feature-execution-loop.md` (пошаговый execution), `create-feature-from-trello-card.md` (промпт). Между canonical-документами есть противоречия (#1, #2, #5, #6, #7, #10, #11).

**Решение владельца:** проблема не в отсутствии иерархии, а в нарушении существующих `canonical_for` границ. Механизм `canonical_for` уже определяет, какой документ владеет какой темой. Рекомендация: при исправлении каждого противоречия свериться с `canonical_for` и убедиться, что тема определена ровно в одном canonical-документе. Промпт — derived-документ, при конфликте canonical побеждает.

---

### 9. [неоднозначность] «минимальный коммит» в bug-fix flow

**Файлы:** `.promts/create-feature-from-trello-card.md`

**Суть:** стадия `fix` требует «Сделать минимальный коммит». Термин не определён.

**Цитата (create-feature-from-trello-card.md:144):**
> | `fix` | Сделать минимальный коммит, заполнить `## Fix` в README |

**Решение владельца:** убрать «минимальный», оставить просто «коммит».

---

### 10. [противоречие] feature-execution-loop.md диаграмма vs текст — improve-loop.sh

**Файлы:** `flows/feature-execution-loop.md` (внутри одного файла)

**Суть:** Mermaid-диаграмма (строки 39-40) показывает `improve-loop.sh` в узлах, а текст (строка 67) явно запрещает его использование.

**Цитата A (диаграмма, строка 39):**
> B[1. Brief Improve Loop\nimprove-loop.sh brief-loop.md]

**Цитата B (текст, строка 67):**
> **Не использовать** `improve-loop.sh` — скрипт вызывает `claude --print` изнутри сессии, что приводит к зависанию.

**Решение владельца:** косметика — при исправлении #1 заодно обновить диаграмму на `Agent tool`.

---

### 11. [неоднозначность] Session Protocol vs execution loop closure — кто выполняет?

**Файлы:** `flows/workflows.md` vs `flows/feature-execution-loop.md`

**Суть:** `workflows.md` Session Protocol «применяется ко всем типам workflow». Но `feature-execution-loop.md` определяет собственный closure (шаг 11) и не ссылается на Session Protocol. Протоколы дополняют друг друга: Session Protocol — тесты, HANDOFF, docs-коммит; execution loop closure — evals, PR, run-state.

**Цитата A (workflows.md:134):**
> Применяется ко всем типам workflow

**Цитата B:** `feature-execution-loop.md` шаг 11 (Closure) не ссылается на Session Protocol.

**Решение владельца:** оба протокола действуют, не заменяют друг друга. Добавить в `feature-execution-loop.md` явную пометку: «Session Protocol из `workflows.md` выполняется тоже».

---

### 12. [битая ссылка] feature-execution-loop.md → eval.md для DR gate evaluator (large.md)

**Файлы:** `flows/feature-execution-loop.md` → `flows/eval.md`

**Суть:** `feature-execution-loop.md` шаг 3 ссылается на `eval.md` для DR evaluator agent (`large.md`), но `eval.md` не определяет evaluator agent для DR gate — говорит, что loops достаточны.

**Цитата A (feature-execution-loop.md:84):**
> Для `large.md`: запустить evaluator agent (DR gate из `eval.md`)

**Цитата B (eval.md:117-121):**
> ### Draft → Design Ready
> Форма: **brief-loop + spec-loop (evaluator agents)**. Self-check не нужен

**Решение владельца:** связана с #2. `eval.md` должен определить DR evaluator для `large.md` — добавить в Evaluator Agent Protocol и Gate Чеклисты.

---

## Сводная таблица

| # | Тип | Файлы | Серьёзность | Решение |
|---|-----|-------|-------------|---------|
| 1 | противоречие | `feature-execution-loop.md` vs `brief-improve-loop.md`, `spec-improve-loop.md`, промпт | **high** | Agent tool для агента, скрипт для человека/CI |
| 2 | противоречие | `eval.md` vs `feature-flow.md` vs `feature-execution-loop.md` (DR gate large.md) | **high** | loops + evaluator для large.md; обновить eval.md |
| 3 | противоречие | промпт vs `git-workflow.md` (`gh repo view`) | **high** | убрать `gh repo view` из промпта |
| 4 | дублирование | `feature-flow.md` vs промпт (evals/ в Bootstrap) | **high** | добавить evals/ в промпт |
| 5 | противоречие | `trello.md` vs `feature-flow.md` (тайминг PLANNING) | **med** | canonical — trello.md (немедленно); подтянуть feature-flow.md |
| 6 | дублирование | `workflows.md` vs промпт, `feature-execution-loop.md` (docs-коммит) | **med** | docs-коммит обязателен; добавить в промпт и execution loop |
| 7 | неоднозначность | `workflows.md` — run-state/ только для bugfix | **med** | вынести bugfix execution в отдельный документ; workflows.md — чистый router |
| 8 | неоднозначность | нарушение canonical_for границ между документами | **med** | проблема в нарушении границ, не в иерархии; при fix сверяться с canonical_for |
| 9 | неоднозначность | «минимальный коммит» не определён | **low** | убрать «минимальный», оставить «коммит» |
| 10 | противоречие | `feature-execution-loop.md` — диаграмма vs текст | **low** | обновить диаграмму вместе с #1 |
| 11 | неоднозначность | Session Protocol vs execution loop closure | **med** | оба действуют; добавить явную ссылку в execution loop |
| 12 | битая ссылка | `feature-execution-loop.md` → `eval.md` (DR evaluator) | **high** | обновить eval.md вместе с #2 |

**Итого:** 5 high, 5 med, 2 low. Все пункты имеют решение владельца.
