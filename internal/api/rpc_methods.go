package api

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// PingMethod: simple health method (no auth)

type PingMethod struct{}

func (m *PingMethod) Name() string { return "ping" }

func (m *PingMethod) Execute(ctx context.Context, params json.RawMessage) (interface{}, error) {
	return map[string]any{
		"pong":    true,
		"time":    time.Now().Unix(),
		"message": "pong",
	}, nil
}

func (m *PingMethod) RequireAuth() bool { return false }

// EchoMethod: echoes input params (no auth)

type EchoMethod struct{}

func (m *EchoMethod) Name() string { return "echo" }

func (m *EchoMethod) Execute(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var input map[string]any
	if len(params) > 0 {
		if err := json.Unmarshal(params, &input); err != nil {
			return nil, fmt.Errorf("invalid params: %v", err)
		}
	}
	return map[string]any{
		"echo": input,
		"time": time.Now().Unix(),
	}, nil
}

func (m *EchoMethod) RequireAuth() bool { return false }
