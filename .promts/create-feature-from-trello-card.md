Изучи карточку в [trello](https://trello.com/c/Oa76ZjDv) через Trello API — автопилот, подтверждение не требуется.

⛔ НЕМЕДЛЕННО — до чтения файлов и до любого обсуждения — переведи карточку TODO → PLANNING — без запроса подтверждения (автопилот по `autonomy-boundaries.md`):
```
PUT https://api.trello.com/1/cards/{shortLink}?key={TRELLO_API_KEY}&token={TRELLO_TOKEN}&idList=69f06f1b601a68bf46282cdf
```

Перед началом прочитай:
- `memory-bank/flows/trello.md` — lifecycle и правила синхронизации
- `memory-bank/flows/feature-flow.md` — gate-чеклисты, identifier taxonomy (OQ-* только в плане)
- `memory-bank/flows/feature-execution-loop.md` — большой цикл выполнения фичи со state-pack
- `memory-bank/engineering/git-workflow.md` — worktree и PR
- `memory-bank/ops/trello-board.md` — стабильные List ID
- `memory-bank/domain/architecture.md` — Layer Stack (Service/Handler/Repo), канонические пути файлов
- `memory-bank/domain/frontend.md` — структура `templates/<domain>/` и `static/`
- `memory-bank/domain/problem.md` — системные ограничения PCON-* (CSRF, auth и др.)
- `memory-bank/engineering/testing-policy.md` — правила CHK: UI-изменения → Playwright обязателен
- `memory-bank/domain/glossary.md` — термины; проверь перед Design Ready, что новые термины добавлены

Давай обсудим карточку. Если описание неполное — задай уточняющие вопросы **перед** созданием Draft.

После получения ответов — до определения номера фичи — выполни оценку масштаба:

**Оценка PRD** (читай критерии в `memory-bank/prd/README.md`):
Предложи создать PRD, если выполняется хотя бы одно условие:
- карточка описывает инициативу, которая явно или вероятно распадётся на ≥ 2 feature packages;
- до проектирования реализации нужно зафиксировать users, goals, product scope и success metrics;
- есть риск смешать продуктовые требования с architecture/design detail внутри одного feature.md.

Если PRD нужен — опиши причину одним предложением и **запроси подтверждение** перед созданием. PRD создаётся в основном репозитории в `memory-bank/prd/` по шаблону `memory-bank/flows/templates/prd/PRD-XXX.md` до старта feature package.

**Оценка ADR** (читай критерии в `memory-bank/adr/README.md`):
Предложи создать ADR, если карточка подразумевает архитектурное или инженерное решение с реальными альтернативами:
- выбор технологии, библиотеки или внешнего сервиса;
- выбор паттерна хранения, API-контракта, протокола или стратегии;
- решение, которое станет upstream-фактом для нескольких downstream-документов.

Если ADR нужен — опиши суть решения и альтернативы одним предложением и **запроси подтверждение**. ADR создаётся в `memory-bank/adr/` по шаблону `memory-bank/flows/templates/adr/ADR-XXX.md` со статусом `proposed`; перевод в `accepted` — после подтверждения человеком.

Если ни PRD, ни ADR не нужны — явно сообщи об этом и продолжай.

---

После оценки масштаба определи номер фичи (следующий FT-XXX из `memory-bank/features/`).

⛔ HARD STOP — ПЕРЕД СОЗДАНИЕМ ЛЮБЫХ ФАЙЛОВ. Выполни в точном порядке **без запроса подтверждения** — все шаги ниже являются автопилотом по `autonomy-boundaries.md`:

**Шаг 1 — Ветка и worktree (из корня основного репозитория):**
```bash
git worktree add ../lfcru_forum-FT-XXX -b feat/FT-XXX-slug
```

**Шаг 2 — Trello: PLANNING → IN PROGRESS (сразу после создания worktree):**
```
PUT https://api.trello.com/1/cards/{shortLink}?key={TRELLO_API_KEY}&token={TRELLO_TOKEN}&idList=69e908732098656229043150
```

**Шаг 3 — Draft PR (сразу после worktree, до первого коммита):**
```bash
gh repo view  # убедиться что контекст — Pegorino82/lfcru_forum
gh pr create --repo Pegorino82/lfcru_forum --draft \
  --title "[WIP][FT-XXX] Краткое описание" \
  --body "Closes #issue — feat/FT-XXX-slug"
```

**Шаг 4 — Вся дальнейшая работа исключительно внутри `../lfcru_forum-FT-XXX`.**
Прямая работа в основной директории ЗАПРЕЩЕНА.
Все создание файлов, коммиты и push — только из worktree-папки.
Чтение любых файлов из `../lfcru_forum-FT-XXX` не требует подтверждения — автопилот по `autonomy-boundaries.md`.

**Routing по label карточки:**

- Label `bug fix` → **облегчённый FT-пакет**: только `README.md` (без feature.md и implementation-plan.md). README содержит: описание бага, условия воспроизведения, корневую причину, ссылку на коммит, regression-тест. Дальнейшие шаги Bootstrap Feature Package не выполняются.
- Label `feature` (или отсутствует) → **полный feature package**: выполняй шаги ниже.

Создай Bootstrap Feature Package (внутри worktree):
1. Создай `memory-bank/features/FT-XXX/README.md`
1b. ⛔ Обнови глобальный индекс `memory-bank/features/README.md` — добавь строку FT-XXX в таблицу Packages
    (статус `planned`, название из карточки). Этот файл обновляется из feature-ветки (worktree).
1c. Инициализируй state-pack — скопируй шаблоны в worktree, заменив `FT-XXX` на реальный ID:
    - `run-state/FT-XXX/active-context.md`
    - `run-state/FT-XXX/stage-log.md`
2. Выбери шаблон feature.md по критериям из `memory-bank/flows/feature-flow.md` § «Выбор шаблона»: `short.md` если фичу можно описать минимальным набором (1 SC-*, 1 CHK-*, 1 EVID-*, без ASM-*/DEC-*/CTR-*/FM-*, без контрактных изменений); иначе `large.md`. Зафикси выбор явно.
3. Создай `memory-bank/features/FT-XXX/feature.md` по выбранному шаблону в статусе draft
4. Наполни feature.md до gate-ready (≥1 REQ-*, NS-*, SC-*, CHK-*, EVID-*; каждый REQ-* прослеживается к SC-*)
5. Запусти **Brief Improve Loop** (автопилот, подтверждение не требуется):
   ```bash
   ./scripts/improve-loop.sh \
     memory-bank/flows/templates/prompts/brief-loop.md \
     memory-bank/features/FT-XXX/feature.md
   ```
   - Итерации до `accept` (max 2, затем escalate к человеку)
   - Обнови `run-state/FT-XXX/stage-log.md` → строка `brief-loop`: done/escalated
6. Запусти **Spec Improve Loop** (автопилот, подтверждение не требуется):
   ```bash
   ./scripts/improve-loop.sh \
     memory-bank/flows/templates/prompts/spec-loop.md \
     memory-bank/features/FT-XXX/feature.md
   ```
   - Итерации до `accept` (max 2, затем escalate к человеку)
   - Обнови `run-state/FT-XXX/stage-log.md` → строка `spec-loop`: done/escalated
7. ⛔ HARD STOP — покажи `feature.md` человеку и дождись явного "ок" перед переводом в Design Ready
   - Обнови `run-state/FT-XXX/active-context.md` → stage: awaiting-dr-approval
8. После подтверждения: `feature.md` → `status: active` (Design Ready)
   - Обнови `run-state/FT-XXX/stage-log.md` → строка `dr-approval`: done
9. Создай `memory-bank/features/FT-XXX/implementation-plan.md`:
   - выполни grounding: пройдись по relevant paths, existing patterns, dependencies
   - зафикси discovery context: relevant paths, local reference patterns, unresolved questions (OQ-*), test surfaces, execution environment
   - план содержит ≥1 PRE-*, ≥1 STEP-*, ≥1 CHK-*, ≥1 EVID-*
   - Обнови `run-state/FT-XXX/stage-log.md` → строка `plan`: done
10. ⛔ HARD STOP — покажи `implementation-plan.md` человеку и дождись явного "ок" перед переводом в Plan Ready
    - Обнови `run-state/FT-XXX/active-context.md` → stage: awaiting-pr-approval
11. После подтверждения: `implementation-plan.md` → `status: active` (Plan Ready)
    - Обнови `run-state/FT-XXX/stage-log.md` → строка `pr-approval`: done

После завершения разработки (Execution → Done gate):
- Выполни STEP-* из `implementation-plan.md` по порядку
- После каждого CP-* обновляй `run-state/FT-XXX/active-context.md` → completed steps
- Зафикси все изменения: `git add . && git commit -m "feat(FT-XXX): <краткое описание>"`
- Запуш ветку: `git push`
- Обнови `run-state/FT-XXX/stage-log.md` → строка `impl`: done
- Запусти unit-тесты локально командой из `memory-bank/ops/development.md` § «Go-тесты» — должны быть зелёными
- Обнови `run-state/FT-XXX/stage-log.md` → строка `unit-tests`: pass/fail
- Убедись что CI зелёный: `rtk gh pr checks` — все jobs (Lint, Go Tests, E2E) должны пройти. ⛔ Запускай ТОЛЬКО после `git push` — иначе CI проверяет устаревший код
- Только если ALL jobs green — переведи PR из draft в ready for review: `gh pr ready`
- Обнови `run-state/FT-XXX/stage-log.md` → строка `closure`: done
- Дождись merge (⛔ HARD STOP — не закрывать артефакты до merge)

После merge PR:
- `feature.md` → `delivery_status: done`
- `implementation-plan.md` → `status: archived`
- Удали worktree: `git worktree remove ../lfcru_forum-FT-XXX && git branch -d feat/FT-XXX-slug`
- Запроси подтверждение перед перемещением карточки в DONE

---

**Resume Protocol** — при возобновлении прерванной работы:
1. Прочитай `HANDOFF.md` → найди FT_ID и текущий stage
2. Прочитай `run-state/FT-XXX/active-context.md` → восстанови контекст
3. Прочитай `run-state/FT-XXX/stage-log.md` → определи следующий незавершённый этап
4. Продолжи с первого `pending` этапа
