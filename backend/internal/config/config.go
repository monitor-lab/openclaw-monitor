package config

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/titanous/json5"
)

const (
	defaultGatewayURL = "ws://127.0.0.1:18789"
	defaultLogDir     = "/tmp/openclaw"
	defaultListenAddr = "127.0.0.1:8077"
)

type PollIntervals struct {
	Health   time.Duration
	Status   time.Duration
	Sessions time.Duration
	Usage    time.Duration
}

type Config struct {
	GatewayURL      string
	GatewayToken    string
	GatewayPassword string
	StateDir        string
	LogDir          string
	ListenAddr      string
	AuthError       string
	Poll            PollIntervals
}

func Load() (*Config, error) {
	stateDir := firstNonEmpty(
		os.Getenv("OPENCLAW_MONITOR_STATE_DIR"),
		os.Getenv("OPENCLAW_STATE_DIR"),
		defaultStateDir(),
	)
	cfg := &Config{
		GatewayURL:   firstNonEmpty(os.Getenv("OPENCLAW_MONITOR_GATEWAY_URL"), defaultGatewayURL),
		StateDir:     stateDir,
		LogDir:       firstNonEmpty(os.Getenv("OPENCLAW_MONITOR_LOG_DIR"), defaultLogDir),
		ListenAddr:   firstNonEmpty(os.Getenv("OPENCLAW_MONITOR_LISTEN"), defaultListenAddr),
		GatewayToken: firstNonEmpty(os.Getenv("OPENCLAW_MONITOR_GATEWAY_TOKEN"), os.Getenv("OPENCLAW_GATEWAY_TOKEN")),
		GatewayPassword: firstNonEmpty(
			os.Getenv("OPENCLAW_MONITOR_GATEWAY_PASSWORD"),
			os.Getenv("OPENCLAW_GATEWAY_PASSWORD"),
		),
		Poll: PollIntervals{
			Health:   5 * time.Second,
			Status:   10 * time.Second,
			Sessions: 10 * time.Second,
			Usage:    30 * time.Second,
		},
	}

	if cfg.GatewayToken == "" && cfg.GatewayPassword == "" {
		token, password, err := readGatewayAuth(filepath.Join(cfg.StateDir, "openclaw.json"))
		if err != nil {
			cfg.AuthError = err.Error()
		} else {
			cfg.GatewayToken = token
			cfg.GatewayPassword = password
		}
	}

	if cfg.GatewayToken == "" && cfg.GatewayPassword == "" {
		cfg.AuthError = "gateway auth token/password missing (set env OPENCLAW_MONITOR_GATEWAY_TOKEN or OPENCLAW_GATEWAY_TOKEN)"
	}

	return cfg, nil
}

func defaultStateDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".openclaw")
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		trimmed := strings.TrimSpace(v)
		if trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func readGatewayAuth(path string) (string, string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", "", err
	}
	var raw map[string]any
	if err := json5.Unmarshal(data, &raw); err != nil {
		return "", "", err
	}
	gw, ok := raw["gateway"].(map[string]any)
	if !ok {
		return "", "", errors.New("gateway config not found")
	}
	auth, _ := gw["auth"].(map[string]any)
	if auth == nil {
		return "", "", errors.New("gateway.auth not found")
	}
	token := toString(auth["token"])
	password := toString(auth["password"])
	if token == "" && password == "" {
		return "", "", errors.New("gateway.auth token/password empty")
	}
	return token, password, nil
}

func toString(value any) string {
	if value == nil {
		return ""
	}
	if s, ok := value.(string); ok {
		return strings.TrimSpace(s)
	}
	return ""
}
