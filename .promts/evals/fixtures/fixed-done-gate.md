Разработка завершена. Выполняю Done gate.

Фиксирую все изменения:

```bash
git add .
git commit -m "feat(FT-XXX): implement feature logic"
```

Пушу ветку:

```bash
git push
```

Запускаю unit-тесты локально — зелёные.

Проверяю CI:

```bash
rtk gh pr checks
```

```
✓ Lint        passed   43s
✓ Go Tests    passed   58s
✓ E2E         passed   1m 47s
```

Все jobs зелёные. Перевожу PR в ready for review:

```bash
gh pr ready
```

PR переведён из draft в ready for review.
