#!/usr/bin/env bash
# improve-loop.sh — runner для brief/spec improve loop
#
# Использование:
#   ./scripts/improve-loop.sh <prompt-file> <artifact-path>
#
# Примеры:
#   ./scripts/improve-loop.sh \
#     memory-bank/flows/templates/prompts/brief-loop.md \
#     memory-bank/features/FT-023/feature.md
#
#   ./scripts/improve-loop.sh \
#     memory-bank/flows/templates/prompts/spec-loop.md \
#     memory-bank/features/FT-023/feature.md

set -euo pipefail

PROMPT_FILE="${1:-}"
ARTIFACT_PATH="${2:-}"

if [[ -z "$PROMPT_FILE" || -z "$ARTIFACT_PATH" ]]; then
  echo "Использование: $0 <prompt-file> <artifact-path>" >&2
  exit 1
fi

if [[ ! -f "$PROMPT_FILE" ]]; then
  echo "Ошибка: prompt-файл не найден: $PROMPT_FILE" >&2
  exit 1
fi

if [[ ! -f "$ARTIFACT_PATH" ]]; then
  echo "Ошибка: артефакт не найден: $ARTIFACT_PATH" >&2
  exit 1
fi

# Определяем FT_ID из пути к артефакту (например: memory-bank/features/FT-023/feature.md → FT-023)
FT_ID=$(echo "$ARTIFACT_PATH" | grep -oE 'FT-[0-9]+' | head -1)
if [[ -z "$FT_ID" ]]; then
  FT_ID="unknown"
fi

DATE=$(date +%Y-%m-%d)

# Подставляем переменные в промпт
RESOLVED_PROMPT=$(sed \
  -e "s|{{ARTIFACT_PATH}}|$ARTIFACT_PATH|g" \
  -e "s|{{FT_ID}}|$FT_ID|g" \
  -e "s|{{DATE}}|$DATE|g" \
  "$PROMPT_FILE")

echo "=== improve-loop.sh ==="
echo "Prompt:   $PROMPT_FILE"
echo "Artifact: $ARTIFACT_PATH"
echo "FT_ID:    $FT_ID"
echo "Date:     $DATE"
echo "========================"
echo ""

# Запускаем claude в non-interactive режиме
echo "$RESOLVED_PROMPT" | claude --print
