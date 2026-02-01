#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
FRONTEND_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
PID_FILE="$SCRIPT_DIR/frontend.pid"
LOG_FILE="$SCRIPT_DIR/frontend.log"

if [[ -f "$PID_FILE" ]]; then
  PID="$(cat "$PID_FILE")"
  if [[ -n "$PID" ]] && kill -0 "$PID" 2>/dev/null; then
    echo "frontend already running (pid $PID)"
    exit 0
  fi
fi

cd "$FRONTEND_DIR"
nohup npm run dev -- --host 127.0.0.1 --port 5173 > "$LOG_FILE" 2>&1 &
PID=$!

echo "$PID" > "$PID_FILE"
echo "frontend started (pid $PID)"
