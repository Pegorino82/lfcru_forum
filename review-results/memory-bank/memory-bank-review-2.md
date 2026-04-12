# Memory Bank Review 2 — Consistency & Contradictions

Date: 2026-04-11

---

## CRITICAL

Критических блокеров не обнаружено.

---

## HIGH

### 1. `go test ./...` без Docker-контекста в testing-policy

**Файл:** `engineering/testing-policy.md:107`

Строка гласит: "Перед handoff агент прогоняет unit-тесты (`go test ./...`) и integration-тесты затронутых пакетов". Голое `go test ./...` некорректно — на хосте нет Go-тулчейна (`development.md:14`: "Всё работает в Docker. Go на хосте не нужен."). При этом сам раздел Stack в `testing-policy.md:32` правильно приводит полные `docker run`-команды.

**Риск:** Агент попытается выполнить `go test ./...` напрямую — команда не найдена.

**Нужно:** Строку 107 привести в соответствие — сослаться на Stack-команды или написать полный `docker run`.

### 2. Модули `forum` и `content` отмечены `*(планируется)*` — форум уже реализован

**Файлы:** `domain/architecture.md:23–24` vs git-история

Таблица модулей содержит:
```
| `forum` *(планируется)* | …
| `content` *(планируется)* | …
```

Коммиты в истории (`fix: replies displayed after parent post`, `fix: forum reply form sends GET instead of POST`, `feat(005-iter2): HTTP Layer`) подтверждают, что форум реализован. `*(планируется)*` — устаревший статус.

**Нужно:** Убрать пометку у `forum` или уточнить, что именно ещё не реализовано (например, только `content`).

> Это продолжение проблемы #3 из `memory-bank-review.md` — не было исправлено.

---

## MEDIUM — Дублирование / SSoT нарушен

| # | Проблема | Файлы-участники |
|---|---|---|
| 3 | Команды запуска тестов продублированы дословно | `engineering/testing-policy.md:28–44` + `ops/development.md:51–74` |
| 4 | Правило `-p 1` (integration race condition) повторено в трёх местах | `domain/architecture.md:55`, `engineering/testing-policy.md:43`, `ops/development.md:74` |
| 5 | Sentinel errors паттерн объявлен в двух canonical документах | `domain/architecture.md:59` + `engineering/coding-style.md:17` |

**По #3:** `ops/development.md` — логичный canonical owner команд запуска. `testing-policy.md` должен ссылаться на него, а не дублировать.

**По #4:** `architecture.md` — лишний. Достаточно в `testing-policy.md` и `development.md`.

**По #5:** `coding-style.md` — owner синтаксиса `var ErrXxx = errors.New(...)`. `architecture.md` должен описывать только ownership (объявлять в `domain/errors.go`) и паттерн использования (`errors.Is` в handler), но не синтаксис объявления.

---

## LOW — Косметика

| # | Проблема | Файл |
|---|---|---|
| 6 | `CLAUDE.md` не упоминает `git-workflow.md` в навигационной таблице | `CLAUDE.md` |
| 7 | `dna/` документы не имеют поля `title` в frontmatter | `dna/principles.md`, `dna/governance.md`, `dna/lifecycle.md`, `dna/cross-references.md` |

**По #6:** Файл `memory-bank/engineering/git-workflow.md` существует и активен, но не включён в таблицу "Что нужно / Куда идти" в `CLAUDE.md`. Агент может не найти его при первичной ориентации.

**По #7:** Все domain/engineering/ops/flows-документы имеют `title`. Governance-документы в `dna/` — нет. Согласно frontmatter-schema это не обязательное поле, но снижает навигационную однородность.

---

## Итого

| Severity | Кол-во |
|---|---|
| CRITICAL | 0 |
| HIGH | 2 |
| MEDIUM | 3 |
| LOW | 2 |

**Главный вывод:** Критических нарушений нет. Основные проблемы — устаревший статус модуля `forum` в `architecture.md` и риск ввести агента в заблуждение командой `go test ./...` без Docker-контекста. SSoT нарушен в нескольких местах дублированием команд и паттернов.
