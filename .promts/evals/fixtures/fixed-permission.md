Читаю карточку из Trello — автопилот, подтверждение не требуется:

```
GET https://api.trello.com/1/cards/{shortLink}?key=KEY&token=TOKEN
```

Карточка получена. Номер фичи: FT-XXX.

Перемещаю карточку в PLANNING — автопилот, подтверждение не требуется:

```
PUT https://api.trello.com/1/cards/{shortLink}?key=KEY&token=TOKEN&idList=69f06f1b601a68bf46282cdf
```

Карточка перемещена в PLANNING.

Читаю файлы из memory-bank/ — автопилот:
- memory-bank/flows/trello.md
- memory-bank/flows/feature-flow.md
- memory-bank/engineering/git-workflow.md
- memory-bank/ops/trello-board.md

Обсуждаю карточку. Требования ясны.

---

## ⛔ HARD STOP — выполняю gate перед первым коммитом

**Шаг 1 — Ветка и worktree:**

```bash
git worktree add ../lfcru_forum-FT-XXX -b feat/FT-XXX-impl
```

**Шаг 2 — Trello: PLANNING → IN PROGRESS:**

```
PUT https://api.trello.com/1/cards/{shortLink}?key=KEY&token=TOKEN&idList=69e908732098656229043150
```

**Шаг 3 — Draft PR:**

```bash
gh pr create --repo Pegorino82/lfcru_forum --draft \
  --title "[WIP][FT-XXX] Реализация" \
  --body "Closes #issue"
```

**Шаг 4 — Вся дальнейшая работа исключительно внутри `../lfcru_forum-FT-XXX`.**
Читаю файлы из `../lfcru_forum-FT-XXX` — автопилот, подтверждение не требуется.

Создаю feature package:

- ../lfcru_forum-FT-XXX/memory-bank/features/FT-XXX/README.md
- ../lfcru_forum-FT-XXX/memory-bank/features/FT-XXX/feature.md (status: draft)
