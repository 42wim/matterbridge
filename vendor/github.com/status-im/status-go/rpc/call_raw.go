package rpc

import (
	"context"
	"encoding/json"

	gethrpc "github.com/ethereum/go-ethereum/rpc"
)

const (
	jsonrpcVersion        = "2.0"
	errInvalidMessageCode = -32700 // from go-ethereum/rpc/errors.go
)

// for JSON-RPC responses obtained via CallRaw(), we have no way
// to know ID field from actual response. web3.js (primary and
// only user of CallRaw()) will validate response by checking
// ID field for being a number:
// https://github.com/ethereum/web3.js/blob/develop/lib/web3/jsonrpc.js#L66
// thus, we will use zero ID as a workaround of this limitation
var defaultMsgID = json.RawMessage(`0`)

// CallRaw performs a JSON-RPC call with already crafted JSON-RPC body. It
// returns string in JSON format with response (successul or error).
func (c *Client) CallRaw(body string) string {
	ctx := context.Background()
	return c.callRawContext(ctx, json.RawMessage(body))
}

// jsonrpcMessage represents JSON-RPC message
type jsonrpcMessage struct {
	Version string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
}

type jsonrpcRequest struct {
	jsonrpcMessage
	ChainID uint64          `json:"chainId"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type jsonrpcSuccessfulResponse struct {
	jsonrpcMessage
	Result json.RawMessage `json:"result"`
}

type jsonrpcErrorResponse struct {
	jsonrpcMessage
	Error jsonError `json:"error"`
}

// jsonError represents Error message for JSON-RPC responses.
type jsonError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// callRawContext performs a JSON-RPC call with already crafted JSON-RPC body and
// given context. It returns string in JSON format with response (successful or error).
//
// TODO(divan): this function exists for compatibility and uses default
// go-ethereum's RPC client under the hood. It adds some unnecessary overhead
// by first marshalling JSON string into object to use with normal Call,
// which is then umarshalled back to the same JSON. The same goes with response.
// This is waste of CPU and memory and should be avoided if possible,
// either by changing exported API (provide only Call, not CallRaw) or
// refactoring go-ethereum's client to allow using raw JSON directly.
func (c *Client) callRawContext(ctx context.Context, body json.RawMessage) string {
	if isBatch(body) {
		return c.callBatchMethods(ctx, body)
	}

	return c.callSingleMethod(ctx, body)
}

// callBatchMethods handles batched JSON-RPC requests, calling each of
// individual requests one by one and constructing proper batched response.
//
// See http://www.jsonrpc.org/specification#batch for details.
//
// We can't use gethtrpc.BatchCall here, because each call should go through
// our routing logic and router to corresponding destination.
func (c *Client) callBatchMethods(ctx context.Context, msgs json.RawMessage) string {
	var requests []json.RawMessage

	err := json.Unmarshal(msgs, &requests)
	if err != nil {
		return newErrorResponse(errInvalidMessageCode, err, defaultMsgID)
	}

	// run all methods sequentially, this seems to be main
	// objective to use batched requests.
	// See: https://github.com/ethereum/wiki/wiki/JavaScript-API#batch-requests
	responses := make([]json.RawMessage, len(requests))
	for i := range requests {
		resp := c.callSingleMethod(ctx, requests[i])
		responses[i] = json.RawMessage(resp)
	}

	data, err := json.Marshal(responses)
	if err != nil {
		c.log.Error("Failed to marshal batch responses:", "error", err)
		return newErrorResponse(errInvalidMessageCode, err, defaultMsgID)
	}

	return string(data)
}

// callSingleMethod executes single JSON-RPC message and constructs proper response.
func (c *Client) callSingleMethod(ctx context.Context, msg json.RawMessage) string {
	// unmarshal JSON body into json-rpc request
	chainID, method, params, id, err := methodAndParamsFromBody(msg)
	if err != nil {
		return newErrorResponse(errInvalidMessageCode, err, id)
	}

	if chainID == 0 {
		chainID = c.UpstreamChainID
	}

	// route and execute
	var result json.RawMessage
	err = c.CallContext(ctx, &result, chainID, method, params...)

	// as we have to return original JSON, we have to
	// analyze returned error and reconstruct original
	// JSON error response.
	if err != nil && err != gethrpc.ErrNoResult {
		if er, ok := err.(gethrpc.Error); ok {
			return newErrorResponse(er.ErrorCode(), err, id)
		}

		return newErrorResponse(errInvalidMessageCode, err, id)
	}

	// finally, marshal answer
	return newSuccessResponse(result, id)
}

// methodAndParamsFromBody extracts Method and Params of
// JSON-RPC body into values ready to use with ethereum-go's
// RPC client Call() function. A lot of empty interface usage is
// due to the underlying code design :/
func methodAndParamsFromBody(body json.RawMessage) (uint64, string, []interface{}, json.RawMessage, error) {
	msg, err := unmarshalMessage(body)
	if err != nil {
		return 0, "", nil, nil, err
	}
	params := []interface{}{}
	if msg.Params != nil {
		err = json.Unmarshal(msg.Params, &params)
		if err != nil {
			return 0, "", nil, nil, err
		}
	}

	return msg.ChainID, msg.Method, params, msg.ID, nil
}

// unmarshalMessage tries to unmarshal JSON-RPC message.
func unmarshalMessage(body json.RawMessage) (*jsonrpcRequest, error) {
	var msg jsonrpcRequest
	err := json.Unmarshal(body, &msg)
	return &msg, err
}

func newSuccessResponse(result json.RawMessage, id json.RawMessage) string {
	if id == nil {
		id = defaultMsgID
	}

	msg := &jsonrpcSuccessfulResponse{
		jsonrpcMessage: jsonrpcMessage{
			ID:      id,
			Version: jsonrpcVersion,
		},
		Result: result,
	}
	data, err := json.Marshal(msg)
	if err != nil {
		return newErrorResponse(errInvalidMessageCode, err, id)
	}

	return string(data)
}

func newErrorResponse(code int, err error, id json.RawMessage) string {
	if id == nil {
		id = defaultMsgID
	}

	errMsg := &jsonrpcErrorResponse{
		jsonrpcMessage: jsonrpcMessage{
			ID:      id,
			Version: jsonrpcVersion,
		},
		Error: jsonError{
			Code:    code,
			Message: err.Error(),
		},
	}

	data, _ := json.Marshal(errMsg)
	return string(data)
}

// isBatch returns true when the first non-whitespace characters is '['
// code from go-ethereum's rpc client (rpc/client.go)
func isBatch(msg json.RawMessage) bool {
	for _, c := range msg {
		// skip insignificant whitespace (http://www.ietf.org/rfc/rfc4627.txt)
		if c == 0x20 || c == 0x09 || c == 0x0a || c == 0x0d {
			continue
		}
		return c == '['
	}
	return false
}
