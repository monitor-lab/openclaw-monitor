# OpenClaw Monitor

Local monitoring dashboard for OpenClaw. This repo contains:

- `backend/` Go service that reads gateway + local files/logs (read-only).
- `frontend/` React UI that visualizes health, sessions, logs, and agent events.

## Quickstart

### Backend

```bash
cd backend
go mod tidy
go run ./cmd/monitor
```

Scripts:

```bash
# foreground
./scripts/dev.sh

# background
./scripts/start.sh
./scripts/stop.sh
./scripts/restart.sh
```

Environment variables (optional):

- `OPENCLAW_MONITOR_GATEWAY_URL` (default `ws://127.0.0.1:18789`)
- `OPENCLAW_MONITOR_GATEWAY_TOKEN` (fallback `OPENCLAW_GATEWAY_TOKEN`)
- `OPENCLAW_MONITOR_GATEWAY_PASSWORD` (fallback `OPENCLAW_GATEWAY_PASSWORD`)
- `OPENCLAW_MONITOR_STATE_DIR` (default `~/.openclaw`)
- `OPENCLAW_MONITOR_LOG_DIR` (default `/tmp/openclaw`)
- `OPENCLAW_MONITOR_LISTEN` (default `127.0.0.1:8077`)

If no gateway token/password is found in env, the backend reads `~/.openclaw/openclaw.json` for `gateway.auth.*`.

### Frontend

```bash
cd frontend
npm install
npm run dev
```

Scripts:

```bash
# foreground
./scripts/dev.sh

# background
./scripts/start.sh
./scripts/stop.sh
./scripts/restart.sh
```

The frontend proxies `/api` and `/ws` to the backend during development.

## Notes
- V1 is read-only. No OpenClaw config writes or control actions.
- If the gateway is down, the backend falls back to local sessions/logs.
