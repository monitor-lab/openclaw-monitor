package collector

import (
	"context"
	"encoding/json"
	"time"

	"openclaw-monitor/internal/gateway"
	"openclaw-monitor/internal/store"
)

type Poller struct {
	Client   *gateway.Client
	Store    *store.Store
	StateDir string
}

func (p *Poller) Start(ctx context.Context) {
	go p.poll(ctx, "health", nil, 5*time.Second, func(env store.Envelope) { p.Store.Update(func(s *store.Snapshot) { s.Health = env }) })
	go p.poll(ctx, "status", nil, 10*time.Second, func(env store.Envelope) { p.Store.Update(func(s *store.Snapshot) { s.Status = env }) })
	go p.poll(ctx, "usage.cost", map[string]any{"days": 30}, 30*time.Second, func(env store.Envelope) { p.Store.Update(func(s *store.Snapshot) { s.Usage = env }) })
	go p.poll(ctx, "sessions.list", map[string]any{}, 10*time.Second, func(env store.Envelope) { p.Store.Update(func(s *store.Snapshot) { s.Sessions = env }) })
	go p.pollLocalSessions(ctx, 30*time.Second)
}

func (p *Poller) poll(ctx context.Context, method string, params any, interval time.Duration, apply func(store.Envelope)) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		apply(p.call(method, params))
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func (p *Poller) call(method string, params any) store.Envelope {
	ctx, cancel := context.WithTimeout(context.Background(), 6*time.Second)
	defer cancel()
	payload, err := p.Client.Call(ctx, method, params)
	if err != nil {
		return store.NewEnvelope("gateway", nil, err)
	}
	return store.NewEnvelope("gateway", payload, nil)
}

func (p *Poller) pollLocalSessions(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		local, err := ReadLocalSessions(p.StateDir)
		if err == nil {
			data, _ := json.Marshal(local)
			p.Store.Update(func(s *store.Snapshot) {
				if s.Sessions.Source != "gateway" || s.Sessions.Error != "" {
					s.Sessions = store.NewEnvelope("local", data, nil)
				}
			})
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}
