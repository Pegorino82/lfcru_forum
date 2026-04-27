Изучи карточку в [trello](https://trello.com/c/v5zs9C1z) через Trello API.

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

**Шаг 1 — Trello (ПЕРВЫМ, до всего остального):**
```
PUT https://api.trello.com/1/cards/{shortLink}?key={TRELLO_API_KEY}&token={TRELLO_TOKEN}&idList=69e908732098656229043150
```

**Шаг 2 — Ветка и worktree (из корня основного репозитория):**
```bash
git worktree add ../lfcru_forum-FT-XXX -b feat/FT-XXX-slug
```

**Шаг 3 — Draft PR (сразу после worktree, до первого коммита):**
```bash
gh pr create --repo Pegorino82/lfcru_forum --draft \
  --title "[WIP][FT-XXX] Краткое описание" \
  --body "Closes #issue — feat/FT-XXX-slug"
```

**Шаг 4 — Вся дальнейшая работа исключительно внутри `../lfcru_forum-FT-XXX`.**
Прямая работа в основной директории ЗАПРЕЩЕНА.
Все создание файлов, коммиты и push — только из worktree-папки.

Создай Bootstrap Feature Package (внутри worktree):
1. Создай `memory-bank/features/FT-XXX/README.md`
2. Создай `memory-bank/features/FT-XXX/feature.md` в статусе draft
3. Доведи feature.md до Design Ready (status: active, ≥1 REQ-*, NS-*, SC-*, CHK-*, EVID-*)
4. Создай `memory-bank/features/FT-XXX/implementation-plan.md` → Plan Ready

После завершения разработки:
- Обнови `feature.md` → `delivery_status: done`
- Переведи PR из draft в ready for review
- Запроси подтверждение перед перемещением карточки в DONE
