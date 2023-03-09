package api

import (
	"bytes"
	"encoding/json"

	"github.com/vmihailenco/msgpack/v5"
)

// ExecuteWithArgs a universal method for calling a sequence of other methods
// while saving and filtering interim results.
//
// The Args map variable allows you to retrieve the parameters passed during
// the request and avoids code formatting.
//
//	return Args.code; // return parameter "code"
//	return Args.v; // return parameter "v"
//
// https://vk.com/dev/execute
func (vk *VK) ExecuteWithArgs(code string, params Params, obj interface{}) error {
	token := vk.getToken()

	reqParams := Params{
		"code":         code,
		"access_token": token,
		"v":            vk.Version,
	}

	resp, err := vk.Handler("execute", params, reqParams)
	if err != nil {
		return err
	}

	var decoderErr error

	if vk.msgpack {
		dec := msgpack.NewDecoder(bytes.NewReader(resp.Response))
		dec.SetCustomStructTag("json")

		decoderErr = dec.Decode(&obj)
	} else {
		decoderErr = json.Unmarshal(resp.Response, &obj)
	}

	if decoderErr != nil {
		return decoderErr
	}

	if resp.ExecuteErrors != nil {
		return &resp.ExecuteErrors
	}

	return err
}

// Execute a universal method for calling a sequence of other methods while
// saving and filtering interim results.
//
// https://vk.com/dev/execute
func (vk *VK) Execute(code string, obj interface{}) error {
	return vk.ExecuteWithArgs(code, Params{}, obj)
}

func fmtBool(value bool) string {
	if value {
		return "1"
	}

	return "0"
}
