Прочитал memory-bank/flows/trello.md, feature-flow.md, git-workflow.md, ops/trello-board.md.

Обсудил карточку. Требования ясны. Номер фичи: FT-XXX.

Черновик подготовлен мысленно — в файловой системе ничего не создаётся до завершения gate.

---

## ⛔ HARD STOP — выполняю gate перед первым коммитом

**Шаг 1 — Trello (ПЕРВЫМ, до всего остального):**

```
PUT https://api.trello.com/1/cards/Yh5wlot4?key=KEY&token=TOKEN&idList=69e908732098656229043150
```

Карточка перемещена в IN PROGRESS.

**Шаг 2 — Ветка и worktree:**

```bash
git worktree add ../lfcru_forum-FT-XXX -b feat/FT-XXX-impl
```

**Шаг 3 — Draft PR (до первого коммита):**

```bash
gh pr create --repo Pegorino82/lfcru_forum --draft \
  --title "[WIP][FT-XXX] Реализация" \
  --body "Closes #issue"
```

**Шаг 4 — Вся дальнейшая работа исключительно внутри ../lfcru_forum-FT-XXX**

Создаю feature package:

- ../lfcru_forum-FT-XXX/memory-bank/features/FT-XXX/README.md
- ../lfcru_forum-FT-XXX/memory-bank/features/FT-XXX/feature.md (status: active)
- ../lfcru_forum-FT-XXX/memory-bank/features/FT-XXX/implementation-plan.md

```bash
git add .
git commit -m "feat(FT-XXX): add feature package"
```
