import { useEffect, useMemo, useState } from "react";
import { useQuery } from "@tanstack/react-query";

const API_BASE = import.meta.env.VITE_API_BASE || "";

const apiUrl = (path: string) => {
  if (!API_BASE) {
    return path;
  }
  return API_BASE.replace(/\/$/, "") + path;
};

const wsUrl = () => {
  if (API_BASE) {
    return API_BASE.replace(/^http/, "ws").replace(/\/$/, "") + "/ws";
  }
  const { protocol, host } = window.location;
  const scheme = protocol === "https:" ? "wss" : "ws";
  return `${scheme}://${host}/ws`;
};

type Envelope = {
  source: string;
  updatedAt: number;
  error?: string;
  data?: unknown;
};

type Meta = {
  gatewayConnected: boolean;
  authError?: string;
};

const tabs = ["Overview", "Agent Events", "Sessions", "Logs", "Usage"] as const;

type GatewayEvent = {
  event?: string;
  payload?: unknown;
};

export default function App() {
  const [activeTab, setActiveTab] = useState<(typeof tabs)[number]>("Overview");
  const [agentEvents, setAgentEvents] = useState<GatewayEvent[]>([]);
  const [rawEvents, setRawEvents] = useState<GatewayEvent[]>([]);
  const [logCursor, setLogCursor] = useState<number>(0);
  const [logLines, setLogLines] = useState<string[]>([]);

  const meta = useQuery({
    queryKey: ["meta"],
    queryFn: () => fetch(apiUrl("/api/meta")).then((r) => r.json() as Promise<Meta>),
    refetchInterval: 4000,
  });

  const health = useQuery({
    queryKey: ["health"],
    queryFn: () => fetch(apiUrl("/api/health")).then((r) => r.json() as Promise<Envelope>),
    refetchInterval: 5000,
  });

  const status = useQuery({
    queryKey: ["status"],
    queryFn: () => fetch(apiUrl("/api/status")).then((r) => r.json() as Promise<Envelope>),
    refetchInterval: 8000,
  });

  const sessions = useQuery({
    queryKey: ["sessions"],
    queryFn: () => fetch(apiUrl("/api/sessions")).then((r) => r.json() as Promise<Envelope>),
    refetchInterval: 10000,
  });

  const usage = useQuery({
    queryKey: ["usage"],
    queryFn: () => fetch(apiUrl("/api/usage")).then((r) => r.json() as Promise<Envelope>),
    refetchInterval: 20000,
  });

  const fetchLogs = async () => {
    const query = new URLSearchParams();
    if (logCursor) {
      query.set("cursor", String(logCursor));
    }
    const res = await fetch(apiUrl(`/api/logs?${query.toString()}`));
    const data = await res.json();
    if (data?.lines) {
      setLogCursor(data.cursor || 0);
      setLogLines(data.lines);
    }
  };

  useEffect(() => {
    fetchLogs();
  }, []);

  useEffect(() => {
    const ws = new WebSocket(wsUrl());
    ws.onmessage = (event) => {
      try {
        const parsed = JSON.parse(event.data);
        if (parsed?.type === "gateway_event") {
          const payload = parsed.payload as { event?: string; payload?: unknown };
          const evt: GatewayEvent = {
            event: payload?.event,
            payload: payload?.payload,
          };
          setRawEvents((prev) => [evt, ...prev].slice(0, 200));
          if (payload?.event === "agent") {
            setAgentEvents((prev) => [evt, ...prev].slice(0, 200));
          }
        }
      } catch {
        // ignore
      }
    };
    return () => ws.close();
  }, []);

  const overviewMeta = useMemo(() => {
    const connected = meta.data?.gatewayConnected;
    return {
      connected: connected ? "Online" : "Offline",
      authError: meta.data?.authError,
    };
  }, [meta.data]);

  return (
    <div className="app">
      <header className="topbar">
        <div>
          <div className="title">OpenClaw Monitor</div>
          <div className="subtitle">Local observability for your gateway</div>
        </div>
        <div className={`status-pill ${overviewMeta.connected === "Online" ? "ok" : "warn"}`}>
          {overviewMeta.connected}
        </div>
      </header>

      <nav className="tabs">
        {tabs.map((tab) => (
          <button
            key={tab}
            className={tab === activeTab ? "tab active" : "tab"}
            onClick={() => setActiveTab(tab)}
          >
            {tab}
          </button>
        ))}
      </nav>

      <main className="content">
        {activeTab === "Overview" && (
          <section className="grid">
            <div className="card">
              <h3>Gateway</h3>
              <p>{overviewMeta.connected}</p>
              {overviewMeta.authError && <p className="error">{overviewMeta.authError}</p>}
            </div>
            <div className="card">
              <h3>Health</h3>
              <pre>{prettyJson(health.data)}</pre>
            </div>
            <div className="card">
              <h3>Status</h3>
              <pre>{prettyJson(status.data)}</pre>
            </div>
          </section>
        )}

        {activeTab === "Agent Events" && (
          <section className="list">
            {agentEvents.length === 0 && <div className="empty">No agent events yet.</div>}
            {agentEvents.map((evt, idx) => (
              <div className="list-item" key={idx}>
                <div className="list-title">agent event</div>
                <pre>{prettyJson(evt.payload)}</pre>
              </div>
            ))}
          </section>
        )}

        {activeTab === "Sessions" && (
          <section className="card">
            <h3>Sessions Snapshot</h3>
            <pre>{prettyJson(sessions.data)}</pre>
          </section>
        )}

        {activeTab === "Logs" && (
          <section className="card">
            <div className="card-row">
              <h3>Logs Tail</h3>
              <button className="ghost" onClick={fetchLogs}>
                Refresh
              </button>
            </div>
            <pre className="log-block">{logLines.join("\n") || "(no logs)"}</pre>
          </section>
        )}

        {activeTab === "Usage" && (
          <section className="card">
            <h3>Usage Summary</h3>
            <pre>{prettyJson(usage.data)}</pre>
          </section>
        )}

        {activeTab !== "Agent Events" && activeTab !== "Overview" && activeTab !== "Sessions" && activeTab !== "Logs" && activeTab !== "Usage" && (
          <section className="card">Coming soon.</section>
        )}
      </main>

      <aside className="side-panel">
        <div className="side-title">Live Events</div>
        <div className="side-list">
          {rawEvents.length === 0 && <div className="empty">No events yet.</div>}
          {rawEvents.slice(0, 12).map((evt, idx) => (
            <div className="side-item" key={idx}>
              <div className="tag">{evt.event || "event"}</div>
              <pre>{truncate(prettyJson(evt.payload), 180)}</pre>
            </div>
          ))}
        </div>
      </aside>
    </div>
  );
}

function prettyJson(value: unknown) {
  if (!value) return "(empty)";
  try {
    return JSON.stringify(value, null, 2);
  } catch {
    return String(value);
  }
}

function truncate(value: string, max: number) {
  if (value.length <= max) return value;
  return value.slice(0, max) + "…";
}
