Изучи карточку в [trello](https://trello.com/c/Oa76ZjDv) через Trello API.

⛔ НЕМЕДЛЕННО — до чтения файлов и до любого обсуждения — переведи карточку TODO → PLANNING:
```
PUT https://api.trello.com/1/cards/{shortLink}?key={TRELLO_API_KEY}&token={TRELLO_TOKEN}&idList=69f06f1b601a68bf46282cdf
```

Перед началом прочитай:
- `memory-bank/flows/trello.md` — lifecycle и правила синхронизации
- `memory-bank/flows/feature-flow.md` — gate-чеклисты
- `memory-bank/engineering/git-workflow.md` — worktree и PR
- `memory-bank/ops/trello-board.md` — стабильные List ID

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

**Routing по label карточки:**

- Label `bug fix` → **облегчённый FT-пакет**: только `README.md` (без feature.md и implementation-plan.md). README содержит: описание бага, условия воспроизведения, корневую причину, ссылку на коммит, regression-тест. Дальнейшие шаги Bootstrap Feature Package не выполняются.
- Label `feature` (или отсутствует) → **полный feature package**: выполняй шаги ниже.

Создай Bootstrap Feature Package (внутри worktree):
1. Создай `memory-bank/features/FT-XXX/README.md`
2. Выбери шаблон feature.md по критериям из `memory-bank/flows/feature-flow.md` § «Выбор шаблона»: `short.md` если фичу можно описать минимальным набором (1 SC-*, 1 CHK-*, 1 EVID-*, без ASM-*/DEC-*/CTR-*/FM-*, без контрактных изменений); иначе `large.md`. Зафикси выбор явно.
3. Создай `memory-bank/features/FT-XXX/feature.md` по выбранному шаблону в статусе draft
4. Доведи feature.md до Design Ready (status: active, ≥1 REQ-*, NS-*, SC-*, CHK-*, EVID-*; каждый REQ-* прослеживается к SC-*)
5. Создай `memory-bank/features/FT-XXX/implementation-plan.md` → Plan Ready:
   - выполни grounding: пройдись по relevant paths, existing patterns, dependencies
   - зафикси discovery context: relevant paths, local reference patterns, unresolved questions (OQ-*), test surfaces, execution environment
   - план содержит ≥1 PRE-*, ≥1 STEP-*, ≥1 CHK-*, ≥1 EVID-*

После завершения разработки (Execution → Done gate):
- Запусти unit-тесты локально командой из `memory-bank/ops/development.md` § «Go-тесты» — должны быть зелёными
- Убедись что CI зелёный: `rtk gh pr checks` — все jobs (Lint, Go Tests, E2E) должны пройти
- Переведи PR из draft в ready for review
- Дождись merge (⛔ HARD STOP — не закрывать артефакты до merge)

После merge PR:
- `feature.md` → `delivery_status: done`
- `implementation-plan.md` → `status: archived`
- Удали worktree: `git worktree remove ../lfcru_forum-FT-XXX && git branch -d feat/FT-XXX-slug`
- Запроси подтверждение перед перемещением карточки в DONE
