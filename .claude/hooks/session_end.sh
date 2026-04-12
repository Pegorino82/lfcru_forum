#!/usr/bin/env bash
set -euo pipefail

LOG_DIR="${CLAUDE_LOG_DIR:-$HOME/.claude/logs}"

if [[ ! -f "$LOG_DIR/.current_session" ]]; then
  exit 0
fi

SESSION_FILE=$(cat "$LOG_DIR/.current_session")
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

cat <<EOF >> "$SESSION_FILE"

## 📊 Session Summary

**End Time (UTC):** $TIMESTAMP

---
EOF

rm -f "$LOG_DIR/.current_session"
