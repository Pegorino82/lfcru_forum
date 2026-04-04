# brief
## 1
- @memory-bank/features/002/brief.md проверь на  полноту и однозначность.
Критерии:
1. Проблема конкретна и измерима (не «улучшить», а конкретная метрика или боль)
2. Назван стейкхолдер или пользователь, для которого решаем
3. Понятен контекст: откуда задача, почему важна сейчас
4. Brief НЕ содержит решения — только проблему и желаемый результат
5. Нет двусмысленных формулировок («быстро», «удобно», «при необходимости»)

Для каждого найденного замечания укажи:
- Что именно не так (цитата из документа)
- Почему это проблема
- Как исправить (конкретное предложение)

Если замечаний нет — напиши «0 замечаний, Brief готов к работе».

Усли замечания есть - сохрани результат ревью в @homeworks/hw-1/review_result.md.
Если файл существует - перепиши, чтобы не копить результаты предыдущих итераций.
Укажи также шаги исправления.

## 2
Прочитай @homeworks/hw-1/review_result.md и внеси соответствующие измения в @memory-bank/features/001/brief.md.
После внесения изменений в @memory-bank/features/001/brief.md очисти @homeworks/hw-1/review_result.md.

# rspec
## 1
- напиши спеку на основании @memory-bank/features/002/brief.md и сохрани в @memory-bank/features/002/rspec.md
документы должны соответсовать подходу sdd

- /spec-reviewer:spec-review -s memory-bank/features/002/rspec.md

## 2
в @memory-bank/features/001/rspec.md в ## 12 есть отложенные улучшения, внеси их в spec

# plan
- сделай Implementation Plan на основе @memory-bank/features/002/rspec.md. результат сохрани в @memory-bank/features/002/implementation_plan.md

- актуализируй @memory-bank/features/001/plan.md согласно @memory-bank/features/001/rspec.md

- @memory-bank/features/002/plan.md.
Проверь этот план на совместимость с текущей кодовой базой. Какие файлы затронешь? Есть ли конфликты? Осуществимо ли это?

# implementation
- начинай реализацию согласно @memory-bank/features/002/plan.md.

# CLOUDE.md
Посмотри @memory-bank/features/001/rspec.md @memory-bank/features/002/rspec.md
Надо добавить в @CLOUDE.md секцию с архитектурой проекта.