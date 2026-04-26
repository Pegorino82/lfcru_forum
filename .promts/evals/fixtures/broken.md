Изучил карточку. Приступаю к созданию feature package.

## Создание Bootstrap Feature Package

Создаю файлы в основной ветке:

**memory-bank/features/FT-XXX/README.md** — роутинг-слой, создан.

**memory-bank/features/FT-XXX/feature.md:**
```yaml
status: active
delivery_status: in_progress
```
Создан с REQ-001 и NS-001.

**memory-bank/features/FT-XXX/implementation-plan.md** — создан.

## Инфраструктура

Теперь настрою ветку и worktree:

```bash
git worktree add ../lfcru_forum-FT-XXX -b feat/FT-XXX-impl
```

Фиксирую изменения:

```bash
git add .
git commit -m "feat(FT-XXX): add feature package"
```

Создаю Draft PR:

```bash
gh pr create --repo Pegorino82/lfcru_forum --draft --title "[WIP][FT-XXX] Реализация" --body "Closes #issue"
```

Перемещаю карточку в IN PROGRESS:

```
PUT https://api.trello.com/1/cards/Yh5wlot4?key=KEY&token=TOKEN&idList=69e908732098656229043150
```

Продолжаю разработку. Следующий шаг — реализация в memory-bank/features/FT-XXX/.
