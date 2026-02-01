# OpenClaw Monitor Design (2026-02-01)

## Goals
- Provide a read-only monitoring dashboard for a local OpenClaw instance.
- Aggregate data from three sources: gateway WebSocket/HTTP RPC, local state files, and local log files.
- Present real-time updates in the UI with a simple, resilient data pipeline.

## Non-Goals (V1)
- No writes to OpenClaw config or runtime.
- No agent/session control (restart/abort/etc.).
- No external storage or cloud deployment.

## Data Sources
1. **Gateway WebSocket** (ws://127.0.0.1:18789)
   - RPC methods: `health`, `status`, `usage.cost`, `usage.status`, `sessions.list`, `logs.tail`.
   - Events: `agent`, `chat`, `heartbeat`, `presence`, `health` (broadcasted).
2. **Local State Files** (`~/.openclaw`)
   - `agents/*/sessions/sessions.json`
   - `agents/*/sessions/*.jsonl` transcripts
   - `subagents/runs.json`
   - `cron/runs/*.jsonl`
3. **Local Logs** (`/tmp/openclaw/openclaw-YYYY-MM-DD.log`)
   - Tail and parse structured JSON lines.

## Authentication
- Token resolution order:
  1. Env: `OPENCLAW_MONITOR_GATEWAY_TOKEN` or `OPENCLAW_GATEWAY_TOKEN`
  2. Config: `~/.openclaw/openclaw.json` (`gateway.auth.token`)
- If no token is available, surface a clear error to the UI.

## Backend (Go)
- Long-lived gateway WS connection with handshake and reconnect.
- RPC wrapper to issue requests and await responses by id.
- Event normalization (`Event{type, ts, source, payload}`) and in-memory cache.
- REST endpoints (read-only):
  - `/api/health`, `/api/status`, `/api/usage`, `/api/sessions`, `/api/logs`
- WebSocket to frontend for streaming deltas: `agent-events`, `logs`, `health`.

## Frontend (React)
- React + Vite + TypeScript
- Data: TanStack Query for polling, WebSocket for streams.
- Pages:
  - Overview (health/status/summary)
  - Agent Runs (timeline of lifecycle/tool events)
  - Sessions (list + preview)
  - Logs (tail + filters)
  - Usage (daily cost/tokens)

## Error Handling
- Gateway down → fallback to local files/logs, mark `source=degraded`.
- Parse errors → isolate and continue.

## Security
- Read-only by design.
- No OpenClaw config or control actions in V1.

## Future Extensions
- Controlled operations with explicit enable flag and role-based permissions.
- Persistent metrics store for long-term history.
