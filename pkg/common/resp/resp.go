package resp

import (
	"encoding/json"

	"github.com/gin-gonic/gin"
)

// RpcRequest defines a JSON-RPC 2.0 request structure
// Only fields used by our minimal server are included
// https://www.jsonrpc.org/specification

type RpcRequest struct {
	JsonRPC string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params"`
	Id      string          `json:"id"`
}

type RpcError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type RpcResponse struct {
	JsonRPC string          `json:"jsonrpc"`
	Id      string          `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *RpcError       `json:"error,omitempty"`
}

func SimpleReturn(ctx *gin.Context, id string, data interface{}) {
	Return(ctx, 200, id, data, nil)
}

func ErrorReturn(ctx *gin.Context, id string, err error) {
	Return(ctx, 200, id, nil, err)
}

func Return(ctx *gin.Context, code int, id string, data interface{}, err error) {
	response := RpcResponse{
		JsonRPC: "2.0",
		Id:      id,
	}

	if err != nil {
		response.Error = &RpcError{Code: -32000, Message: err.Error()}
	} else {
		jsonData, _ := json.Marshal(data)
		response.Result = jsonData
	}

	ctx.JSON(code, response)
}
