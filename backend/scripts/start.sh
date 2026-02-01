#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BACKEND_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
PID_FILE="$SCRIPT_DIR/monitor.pid"
LOG_FILE="$SCRIPT_DIR/monitor.log"

if [[ -f "$PID_FILE" ]]; then
  PID="$(cat "$PID_FILE")"
  if [[ -n "$PID" ]] && kill -0 "$PID" 2>/dev/null; then
    echo "backend already running (pid $PID)"
    exit 0
  fi
fi

cd "$BACKEND_DIR"
nohup go run ./cmd/monitor > "$LOG_FILE" 2>&1 &
PID=$!

echo "$PID" > "$PID_FILE"
echo "backend started (pid $PID)"
