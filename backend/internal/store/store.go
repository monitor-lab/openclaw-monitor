package store

import (
	"encoding/json"
	"sync"
	"time"
)

type Envelope struct {
	Source    string          `json:"source"`
	UpdatedAt int64           `json:"updatedAt"`
	Error     string          `json:"error,omitempty"`
	Data      json.RawMessage `json:"data,omitempty"`
}

type Snapshot struct {
	GatewayConnected bool     `json:"gatewayConnected"`
	AuthError        string   `json:"authError,omitempty"`
	Health           Envelope `json:"health"`
	Status           Envelope `json:"status"`
	Sessions         Envelope `json:"sessions"`
	Usage            Envelope `json:"usage"`
}

type Store struct {
	mu   sync.RWMutex
	snap Snapshot
}

func New() *Store {
	return &Store{snap: Snapshot{}}
}

func (s *Store) Update(fn func(*Snapshot)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	fn(&s.snap)
}

func (s *Store) Snapshot() Snapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.snap
}

func NewEnvelope(source string, data []byte, err error) Envelope {
	out := Envelope{Source: source, UpdatedAt: time.Now().UnixMilli()}
	if err != nil {
		out.Error = err.Error()
		return out
	}
	if data != nil {
		out.Data = data
	}
	return out
}
