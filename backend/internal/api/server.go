package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"openclaw-monitor/internal/collector"
	"openclaw-monitor/internal/store"
)

type Server struct {
	store  *store.Store
	logDir string
	hub    *Hub
}

func NewServer(store *store.Store, logDir string, hub *Hub) *Server {
	return &Server{store: store, logDir: logDir, hub: hub}
}

func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/health", s.handleHealth)
	mux.HandleFunc("/api/status", s.handleStatus)
	mux.HandleFunc("/api/sessions", s.handleSessions)
	mux.HandleFunc("/api/usage", s.handleUsage)
	mux.HandleFunc("/api/logs", s.handleLogs)
	mux.HandleFunc("/api/meta", s.handleMeta)
	mux.HandleFunc("/ws", s.handleWS)
	return withCORS(mux)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, s.store.Snapshot().Health)
}

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, s.store.Snapshot().Status)
}

func (s *Server) handleSessions(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, s.store.Snapshot().Sessions)
}

func (s *Server) handleUsage(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, s.store.Snapshot().Usage)
}

func (s *Server) handleMeta(w http.ResponseWriter, r *http.Request) {
	snap := s.store.Snapshot()
	writeJSON(w, map[string]any{
		"gatewayConnected": snap.GatewayConnected,
		"authError":        snap.AuthError,
	})
}

func (s *Server) handleLogs(w http.ResponseWriter, r *http.Request) {
	cursor := parseInt64(r.URL.Query().Get("cursor"))
	limit := parseInt(r.URL.Query().Get("limit"))
	maxBytes := parseInt(r.URL.Query().Get("maxBytes"))
	res, err := collector.TailLocalLogs(s.logDir, cursor, limit, maxBytes)
	if err != nil {
		writeJSON(w, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, res)
}

func (s *Server) handleWS(w http.ResponseWriter, r *http.Request) {
	initial := map[string]any{
		"type":    "snapshot",
		"payload": s.store.Snapshot(),
	}
	s.hub.HandleWS(w, r, initial)
}

func writeJSON(w http.ResponseWriter, value any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(value)
}

func parseInt(value string) int {
	if value == "" {
		return 0
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0
	}
	return parsed
}

func parseInt64(value string) int64 {
	if value == "" {
		return 0
	}
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0
	}
	return parsed
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
