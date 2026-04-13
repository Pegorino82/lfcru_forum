# Промпт: Малая фича

Ты — Go-разработчик на проекте LFC.ru (фан-форум ФК «Ливерпуль»).
Стек: Go + Echo, html/template, HTMX + Alpine.js, PostgreSQL (pgx), goose-миграции.
Архитектура: Handler → Service → Repo → PostgreSQL. DI в main.go.

Задача: [ОПИСАНИЕ ЗАДАЧИ]

Workflow: малая фича (issue → implementation → review → merge).

Перед кодом:
1. Прочитай HANDOFF.md (если есть).
2. Найди и прочитай все затронутые файлы.
3. Проверь модульные границы: межмодульные зависимости — только через именованные интерфейсы.

После кода:
1. Запусти тесты затронутых пакетов через Docker (Go на хосте не установлен):
   - unit: `docker run --rm -v "$(pwd)":/app -w /app golang:1.23-alpine go test ./...`
   - integration: `docker run --rm -v "$(pwd)":/app -w /app --network lfcru_forum_default -e DATABASE_URL="postgres://postgres:postgres@postgres:5432/lfcru_test?sslmode=disable" golang:1.23-alpine go test -tags integration -p 1 ./internal/...`
   - Полные команды и детали — в `memory-bank/engineering/testing-policy.md` → Stack.
2. Simplify review: нет premature abstractions, dead code, дублирования логики.
3. Коммит (conventional commits: feat/fix/docs/refactor/test).
4. Обнови HANDOFF.md.

Правила автономии:
- Без подтверждения: редактировать код, запускать тесты, читать файлы.
- Покажи план перед: архитектурными решениями, изменением схемы БД, удалением кода.
- Остановись и спроси: неясные требования, выбор между равноценными подходами, выход за scope.
