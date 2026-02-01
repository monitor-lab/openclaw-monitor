package gateway

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
)

type Client struct {
	url      string
	token    string
	password string

	connMu  sync.Mutex
	conn    *websocket.Conn
	pending map[string]chan ResponseFrame
	events  chan EventFrame
	helloCh chan struct{}
	disconnectCh chan struct{}

	connected atomic.Bool
	closed    atomic.Bool
}

func NewClient(url, token, password string) *Client {
	return &Client{
		url:      url,
		token:    token,
		password: password,
		pending:  map[string]chan ResponseFrame{},
		events:   make(chan EventFrame, 256),
		helloCh:  make(chan struct{}),
	}
}

func (c *Client) Events() <-chan EventFrame {
	return c.events
}

func (c *Client) IsConnected() bool {
	return c.connected.Load()
}

func (c *Client) Close() {
	if c.closed.Swap(true) {
		return
	}
	c.connMu.Lock()
	defer c.connMu.Unlock()
	if c.conn != nil {
		_ = c.conn.Close()
	}
}

func (c *Client) Run(ctx context.Context) {
	backoff := time.Second
	for {
		if c.closed.Load() {
			return
		}
		err := c.connectOnce(ctx)
		if err == nil {
			backoff = time.Second
			// Wait until connection drops or context cancels.
			select {
			case <-ctx.Done():
				c.Close()
				return
			case <-c.disconnectCh:
				// reconnect on disconnect
			}
		} else {
			select {
			case <-ctx.Done():
				return
			case <-time.After(backoff):
				backoff = minDuration(backoff*2, 30*time.Second)
			}
		}
	}
}

func (c *Client) Call(ctx context.Context, method string, params any) (json.RawMessage, error) {
	if !c.connected.Load() {
		return nil, errors.New("gateway not connected")
	}
	id := newID()
	respCh := make(chan ResponseFrame, 1)
	c.connMu.Lock()
	if c.conn == nil {
		c.connMu.Unlock()
		return nil, errors.New("gateway connection missing")
	}
	c.pending[id] = respCh
	req := RequestFrame{
		Type:   "req",
		ID:     id,
		Method: method,
		Params: params,
	}
	if err := c.conn.WriteJSON(req); err != nil {
		delete(c.pending, id)
		c.connMu.Unlock()
		return nil, err
	}
	c.connMu.Unlock()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case resp := <-respCh:
		if !resp.OK {
			if resp.Error != nil {
				return nil, fmt.Errorf("gateway error: %s", resp.Error.Message)
			}
			return nil, errors.New("gateway error")
		}
		return resp.Payload, nil
	}
}

func (c *Client) connectOnce(ctx context.Context) error {
	c.connMu.Lock()
	if c.conn != nil {
		_ = c.conn.Close()
	}
	c.pending = map[string]chan ResponseFrame{}
	c.helloCh = make(chan struct{})
	c.disconnectCh = make(chan struct{})
	c.connMu.Unlock()

	dialer := websocket.Dialer{Proxy: http.ProxyFromEnvironment}
	conn, _, err := dialer.DialContext(ctx, c.url, nil)
	if err != nil {
		return err
	}

	c.connMu.Lock()
	c.conn = conn
	c.connMu.Unlock()

	if err := c.sendConnect(); err != nil {
		_ = conn.Close()
		return err
	}

	go c.readLoop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-c.helloCh:
		c.connected.Store(true)
		return nil
	case <-time.After(8 * time.Second):
		return errors.New("gateway connect timeout")
	}
}

func (c *Client) sendConnect() error {
	client := ConnectClient{
		ID:          "gateway-client",
		DisplayName: "openclaw-monitor",
		Version:     "0.1.0",
		Platform:    fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
		Mode:        "backend",
		InstanceID:  newID(),
	}
	params := ConnectParams{
		MinProtocol: ProtocolVersion,
		MaxProtocol: ProtocolVersion,
		Client:      client,
		Role:        "operator",
		Scopes:      []string{},
	}
	if c.token != "" || c.password != "" {
		params.Auth = &ConnectAuth{Token: c.token, Password: c.password}
	}
	req := RequestFrame{
		Type:   "req",
		ID:     newID(),
		Method: "connect",
		Params: params,
	}

	c.connMu.Lock()
	defer c.connMu.Unlock()
	if c.conn == nil {
		return errors.New("gateway socket missing")
	}
	return c.conn.WriteJSON(req)
}

func (c *Client) readLoop() {
	for {
		c.connMu.Lock()
		conn := c.conn
		c.connMu.Unlock()
		if conn == nil {
			select {
			case <-c.disconnectCh:
			default:
				close(c.disconnectCh)
			}
			return
		}

		_, data, err := conn.ReadMessage()
		if err != nil {
			c.connected.Store(false)
			c.connMu.Lock()
			if c.conn != nil {
				_ = c.conn.Close()
				c.conn = nil
			}
			for id, ch := range c.pending {
				close(ch)
				delete(c.pending, id)
			}
			c.connMu.Unlock()
			select {
			case <-c.disconnectCh:
			default:
				close(c.disconnectCh)
			}
			return
		}

		var envelope struct {
			Type string `json:"type"`
		}
		if err := json.Unmarshal(data, &envelope); err != nil {
			continue
		}

		switch envelope.Type {
		case "hello-ok":
			select {
			case <-c.helloCh:
				// already closed
			default:
				close(c.helloCh)
			}
		case "res":
			var resp ResponseFrame
			if err := json.Unmarshal(data, &resp); err != nil {
				continue
			}
			c.connMu.Lock()
			ch := c.pending[resp.ID]
			delete(c.pending, resp.ID)
			c.connMu.Unlock()
			if ch != nil {
				ch <- resp
				close(ch)
			}
		case "event":
			var evt EventFrame
			if err := json.Unmarshal(data, &evt); err != nil {
				continue
			}
			select {
			case c.events <- evt:
			default:
				// drop if slow
			}
		}
	}
}

func newID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

func minDuration(a, b time.Duration) time.Duration {
	if a < b {
		return a
	}
	return b
}
