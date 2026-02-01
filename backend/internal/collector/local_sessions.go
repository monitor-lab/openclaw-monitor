package collector

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
)

type LocalSessionStore struct {
	AgentID string                 `json:"agentId"`
	Path    string                 `json:"path"`
	Store   map[string]any         `json:"store"`
}

type LocalSessionsSnapshot struct {
	Agents []LocalSessionStore `json:"agents"`
}

func ReadLocalSessions(stateDir string) (LocalSessionsSnapshot, error) {
	agentsDir := filepath.Join(stateDir, "agents")
	entries, err := os.ReadDir(agentsDir)
	if err != nil {
		return LocalSessionsSnapshot{}, err
	}
	var out []LocalSessionStore
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		agentID := entry.Name()
		path := filepath.Join(agentsDir, agentID, "sessions", "sessions.json")
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		var store map[string]any
		if err := json.Unmarshal(data, &store); err != nil {
			continue
		}
		out = append(out, LocalSessionStore{AgentID: agentID, Path: path, Store: store})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].AgentID < out[j].AgentID })
	return LocalSessionsSnapshot{Agents: out}, nil
}
