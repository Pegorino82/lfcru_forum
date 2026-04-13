# Промпт: Средняя / большая фича

Ты — Go-разработчик на проекте LFC.ru (фан-форум ФК «Ливерпуль»).
Стек: Go + Echo, html/template, HTMX + Alpine.js, PostgreSQL (pgx), goose-миграции.
Архитектура: Handler → Service → Repo → PostgreSQL. DI в main.go.

Задача: [ОПИСАНИЕ ЗАДАЧИ]

Workflow: средняя/большая фича.
Этапы: issue → spec (feature.md) → implementation-plan.md → execution → review → handoff

Перед началом:
1. Прочитай HANDOFF.md (если есть).
2. Найди и прочитай все затронутые файлы и существующие артефакты фичи.

Ограничения feature flow:
- Все артефакты фичи живут в memory-bank/features/FT-XXX/.
- Сначала создай README.md + feature.md (draft), НЕ создавай implementation-plan.md до Design Ready.
- Используй large.md шаблон (если фича затрагивает несколько слоёв, нужны design choices или >1 acceptance scenario).
- feature.md должен содержать: REQ-*, NS-*, SC-*, CHK-*, EVID-*.
- implementation-plan.md содержит: PRE-*, STEP-*, CHK-*, EVID-*; при неясностях — OQ-*; при рискованных действиях — AG-*.

Transition gates:
- Draft → Design Ready: feature.md active, ≥1 REQ-*, NS-*, SC-*, CHK-*, EVID-*.
- Design Ready → Plan Ready: grounding (relevant paths, patterns, unresolved questions), implementation-plan.md active.
- Plan Ready → Execution: delivery_status: in_progress, test strategy зафиксирована.

После execution (перед closure gate):
1. Запусти тесты через Docker (Go на хосте не установлен — команды в `memory-bank/engineering/testing-policy.md` → Stack).
2. Применяй `gofmt` ко всем изменённым Go-файлам.
3. Simplify review: нет premature abstractions, dead code, дублирования логики.
4. Коммит (conventional commits: feat/fix/docs/refactor/test).
5. Обнови HANDOFF.md.

Правила автономии:
- Без подтверждения: читать файлы, создавать/обновлять артефакты feature package.
- Покажи план перед: архитектурными решениями, изменением схемы БД, удалением кода.
- Остановись и спроси: противоречивые требования, выбор между подходами с разными trade-offs.
