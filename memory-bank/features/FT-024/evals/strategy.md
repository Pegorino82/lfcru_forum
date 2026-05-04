---
title: "FT-024: Eval Strategy"
doc_kind: feature
doc_function: eval-strategy
ft_id: FT-024
status: active
audience: humans_and_agents
---

# FT-024 Eval Strategy

## Gates & Forms

| Gate | Форма | Evaluator |
|---|---|---|
| Draft → Design Ready | evaluator agent | Agent tool |
| Design Ready → Plan Ready | evaluator agent | Agent tool |
| Plan Ready → Execution | human approval | пользователь |
| Execution → Done | hybrid: CI + evaluator agent + human AG-* | Agent tool + AG-* |

> Фича `large.md` → evaluator agent обязателен на DR→PR и Execution→Done.

## Risk Areas

- **Файловая система**: загрузка аватара затрагивает файловую систему (ADR-005) — важно проверить корректное хранение, удаление и путь к файлу.
- **CSRF**: `POST /profile/avatar` требует CSRF-токен (PCON-02) — пропуск ведёт к уязвимости.
- **Авторизация**: только владелец профиля может изменить аватар — неправильная проверка открывает несанкционированную запись.
- **Fallback аватара**: детерминированный цвет по хешу имени — риск нестабильного результата при изменении алгоритма.
- **Клик по имени/аватару везде**: изменение шаблонов форума и комментариев — широкая поверхность Playwright-тестов.
- **Multipart upload**: парсинг файла на стороне Go — риски при некорректном Content-Type или превышении лимита.
