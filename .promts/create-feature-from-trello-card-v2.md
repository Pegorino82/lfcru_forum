Изучи карточку в [trello](https://trello.com/c/vWjzMaXc) через Trello API.

Перед началом прочитай:
- `memory-bank/flows/trello.md` — lifecycle и правила синхронизации
- `memory-bank/flows/feature-flow.md` — gate-чеклисты
- `memory-bank/engineering/git-workflow.md` — worktree и PR
- `memory-bank/ops/trello-board.md` — стабильные List ID

Давай обсудим карточку. Если описание неполное — задай уточняющие вопросы **перед** созданием Draft.

После получения ответов создай Bootstrap Feature Package:
1. Определи номер фичи (следующий FT-XXX из `memory-bank/features/`)
2. Создай `memory-bank/features/FT-XXX/README.md`
3. Создай `memory-bank/features/FT-XXX/feature.md` в статусе draft
4. Доведи feature.md до Design Ready (status: active, ≥1 REQ-*, NS-*, SC-*, CHK-*, EVID-*)
5. Создай `memory-bank/features/FT-XXX/implementation-plan.md` → Plan Ready

⛔ HARD STOP — ПЕРЕД ПЕРВЫМ КОММИТОМ С КОДОМ. Выполни в точном порядке:

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
Прямая работа в основной директории после создания worktree ЗАПРЕЩЕНА.
Все создание файлов, коммиты и push — только из worktree-папки.

После завершения разработки:
- Обнови `feature.md` → `delivery_status: done`
- Переведи PR из draft в ready for review
- Запроси подтверждение перед перемещением карточки в DONE
