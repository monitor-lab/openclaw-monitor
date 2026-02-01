package gateway

import "encoding/json"

const ProtocolVersion = 3

// Request/response frames

type RequestFrame struct {
	Type   string `json:"type"`
	ID     string `json:"id"`
	Method string `json:"method"`
	Params any    `json:"params,omitempty"`
}

type ResponseFrame struct {
	Type    string          `json:"type"`
	ID      string          `json:"id"`
	OK      bool            `json:"ok"`
	Payload json.RawMessage `json:"payload,omitempty"`
	Error   *ErrorShape     `json:"error,omitempty"`
}

type ErrorShape struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

type EventFrame struct {
	Type        string          `json:"type"`
	Event       string          `json:"event"`
	Payload     json.RawMessage `json:"payload,omitempty"`
	Seq         *int64          `json:"seq,omitempty"`
	StateVersion *StateVersion  `json:"stateVersion,omitempty"`
}

type StateVersion struct {
	Presence *int64 `json:"presence,omitempty"`
	Health   *int64 `json:"health,omitempty"`
}

type HelloOk struct {
	Type     string `json:"type"`
	Protocol int    `json:"protocol"`
	Server   any    `json:"server"`
	Features any    `json:"features"`
	Snapshot any    `json:"snapshot"`
	Auth     any    `json:"auth,omitempty"`
	Policy   any    `json:"policy"`
}

type ConnectParams struct {
	MinProtocol int           `json:"minProtocol"`
	MaxProtocol int           `json:"maxProtocol"`
	Client      ConnectClient `json:"client"`
	Role        string        `json:"role,omitempty"`
	Scopes      []string      `json:"scopes,omitempty"`
	Auth        *ConnectAuth  `json:"auth,omitempty"`
}

type ConnectClient struct {
	ID          string `json:"id"`
	DisplayName string `json:"displayName,omitempty"`
	Version     string `json:"version"`
	Platform    string `json:"platform"`
	Mode        string `json:"mode"`
	InstanceID  string `json:"instanceId,omitempty"`
}

type ConnectAuth struct {
	Token    string `json:"token,omitempty"`
	Password string `json:"password,omitempty"`
}
