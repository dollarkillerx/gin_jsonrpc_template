package api

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/google/feitian/pkg/common/resp"
)

// RpcMethod defines the interface for a JSON-RPC method
// Name: method name; Execute: business logic; RequireAuth: whether it needs auth
// For the base scaffold we do not implement auth; it always returns false.

type RpcMethod interface {
	Name() string
	Execute(ctx context.Context, params json.RawMessage) (interface{}, error)
	RequireAuth() bool
}

type RpcHandler struct {
	methods map[string]RpcMethod
	mu      sync.RWMutex
}

func NewRpcHandler() *RpcHandler {
	return &RpcHandler{methods: make(map[string]RpcMethod)}
}

func (h *RpcHandler) RegisterMethod(method RpcMethod) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.methods[method.Name()] = method
}

func (h *RpcHandler) getMethod(name string) (RpcMethod, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	m, ok := h.methods[name]
	return m, ok
}

func (h *RpcHandler) HandleRpcRequest(ctx *gin.Context) {
	var request resp.RpcRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		resp.ErrorReturn(ctx, request.Id, fmt.Errorf("invalid request: %v", err))
		return
	}

	if request.JsonRPC != "2.0" {
		resp.ErrorReturn(ctx, request.Id, fmt.Errorf("unsupported jsonrpc version: %s", request.JsonRPC))
		return
	}

	method, exists := h.getMethod(request.Method)
	if !exists {
		resp.ErrorReturn(ctx, request.Id, fmt.Errorf("method not found: %s", request.Method))
		return
	}

	// For base scaffold we skip auth check. If method.RequireAuth() returns true, still allow.
	result, err := method.Execute(ctx, request.Params)
	if err != nil {
		resp.ErrorReturn(ctx, request.Id, err)
		return
	}

	resp.SimpleReturn(ctx, request.Id, result)
}
