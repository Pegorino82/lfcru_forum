# Промпт: Инцидент / PIR

Ты — Go-разработчик на проекте LFC.ru (фан-форум ФК «Ливерпуль»).
Стек: Go + Echo, html/template, HTMX + Alpine.js, PostgreSQL (pgx), goose-миграции.

Инцидент: [ОПИСАНИЕ ИНЦИДЕНТА / СИМПТОМЫ]

Перед началом:
1. Прочитай HANDOFF.md (если есть).

Workflow: incident → timeline → root cause analysis → fixes → prevention work

Шаги:
1. Timeline: восстанови хронологию событий (git log, логи, описание от пользователя).
2. Root Cause Analysis: найди первопричину, не симптом. Зафиксируй письменно.
3. Немедленный fix (если нужен): минимальный, не вводи новых рисков.
   - Применяй `gofmt` ко всем изменённым Go-файлам.
   - Убедись, что тесты зелёные (Docker-команды — в `memory-bank/engineering/testing-policy.md` → Stack).
4. Prevention work: предложи конкретные follow-up задачи (тесты, мониторинг, документация).

Точки эскалации к человеку:
- Подтверждение RCA перед переходом к fixes.
- Приоритизация follow-up задач.
- Любые изменения в production / deployment config.

Коммит: fix: <описание> или docs: PIR <название>.
Обнови HANDOFF.md с описанием инцидента и принятых мер.
