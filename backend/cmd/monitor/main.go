package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"openclaw-monitor/internal/api"
	"openclaw-monitor/internal/collector"
	"openclaw-monitor/internal/config"
	"openclaw-monitor/internal/gateway"
	"openclaw-monitor/internal/store"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config load failed: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	snapStore := store.New()
	snapStore.Update(func(s *store.Snapshot) {
		s.AuthError = cfg.AuthError
	})

	gw := gateway.NewClient(cfg.GatewayURL, cfg.GatewayToken, cfg.GatewayPassword)
	go gw.Run(ctx)

	hub := api.NewHub()

	poller := collector.Poller{
		Client:   gw,
		Store:    snapStore,
		StateDir: cfg.StateDir,
	}
	poller.Start(ctx)

	go func() {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()
		for {
			snapStore.Update(func(s *store.Snapshot) {
				s.GatewayConnected = gw.IsConnected()
			})
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
			}
		}
	}()

	go func() {
		for evt := range gw.Events() {
			hub.Broadcast(map[string]any{
				"type":    "gateway_event",
				"payload": evt,
			})
		}
	}()

	server := api.NewServer(snapStore, cfg.LogDir, hub)
	httpServer := &http.Server{
		Addr:    cfg.ListenAddr,
		Handler: server.Routes(),
	}

	go func() {
		log.Printf("openclaw-monitor backend listening on %s", cfg.ListenAddr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("http server failed: %v", err)
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	_ = httpServer.Shutdown(shutdownCtx)
	gw.Close()
}
