#!/usr/bin/env bash
# Stop hook: сохраняет transcript_path для обработки в session_end.sh
set -euo pipefail

LOG_DIR="${CLAUDE_LOG_DIR:-$HOME/.claude/logs}"
payload="$(cat)"

TRANSCRIPT_PATH=$(echo "$payload" | jq -r '.transcript_path // ""' 2>/dev/null || true)
if [[ -n "$TRANSCRIPT_PATH" ]]; then
  echo "$TRANSCRIPT_PATH" > "$LOG_DIR/.transcript_path"
fi
