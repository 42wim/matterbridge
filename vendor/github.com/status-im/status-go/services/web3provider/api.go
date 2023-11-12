package web3provider

import (
	"encoding/json"
	"errors"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/log"
	signercore "github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/status-im/status-go/account"
	"github.com/status-im/status-go/eth-node/types"
	"github.com/status-im/status-go/services/typeddata"
	"github.com/status-im/status-go/transactions"
)

const Web3SendAsyncReadOnly = "web3-send-async-read-only"
const RequestAPI = "api-request"

const Web3SendAsyncCallback = "web3-send-async-callback"
const ResponseAPI = "api-response"
const Web3ResponseError = "web3-response-error"

const PermissionWeb3 = "web3"
const PermissionContactCode = "contact-code"
const PermissionUnknown = "unknown"

const ethCoinbase = "eth_coinbase"

var ErrorInvalidAPIRequest = errors.New("invalid API request")
var ErrorUnknownPermission = errors.New("unknown permission")

var authMethods = []string{
	"eth_accounts",
	"eth_coinbase",
	"eth_sendTransaction",
	"eth_sign",
	"keycard_signTypedData",
	"eth_signTypedData",
	"eth_signTypedData_v3",
	"personal_sign",
}

var signMethods = []string{
	"eth_sign",
	"personal_sign",
	"eth_signTypedData",
	"eth_signTypedData_v3",
	"eth_signTypedData_v4",
}

var accMethods = []string{
	"eth_accounts",
	"eth_coinbase",
}

func NewAPI(s *Service) *API {
	return &API{
		s: s,
	}
}

// API is class with methods available over RPC.
type API struct {
	s *Service
}

type ETHPayload struct {
	ID       interface{}   `json:"id,omitempty"`
	JSONRPC  string        `json:"jsonrpc"`
	From     string        `json:"from"`
	Method   string        `json:"method"`
	Params   []interface{} `json:"params"`
	Password string        `json:"password,omitempty"`
	ChainID  uint64        `json:"chainId,omitempty"`
}

type JSONRPCResponse struct {
	ID      interface{} `json:"id,omitempty"`
	JSONRPC string      `json:"jsonrpc"`
	Result  interface{} `json:"result"`
}
type Web3SendAsyncReadOnlyRequest struct {
	Title     string      `json:"title,omitempty"`
	MessageID interface{} `json:"messageId"`
	Payload   ETHPayload  `json:"payload"`
	Hostname  string      `json:"hostname"`
	Address   string      `json:"address,omitempty"`
}

type Web3SendAsyncReadOnlyError struct {
	Code    uint   `json:"code"`
	Message string `json:"message,omitempty"`
}

type Web3SendAsyncReadOnlyResponse struct {
	ProviderResponse

	MessageID interface{} `json:"messageId"`
	Error     interface{} `json:"error,omitempty"`
	Result    interface{} `json:"result,omitempty"`
}

type APIRequest struct {
	MessageID  interface{} `json:"messageId,omitempty"`
	Address    string      `json:"address,omitempty"`
	Hostname   string      `json:"hostname"`
	Permission string      `json:"permission"`
}

type APIResponse struct {
	ProviderResponse

	MessageID  interface{} `json:"messageId,omitempty"`
	Permission string      `json:"permission"`
	Data       interface{} `json:"data,omitempty"`
	IsAllowed  bool        `json:"isAllowed"`
}

type ProviderResponse struct {
	ResponseType string `json:"type"`
}

func (api *API) ProcessRequest(requestType string, payload json.RawMessage) (interface{}, error) {
	switch requestType {
	case RequestAPI:
		var request APIRequest
		if err := json.Unmarshal([]byte(payload), &request); err != nil {
			return nil, err
		}
		return api.ProcessAPIRequest(request)
	case Web3SendAsyncReadOnly:
		var request Web3SendAsyncReadOnlyRequest
		if err := json.Unmarshal(payload, &request); err != nil {
			return nil, err
		}
		return api.ProcessWeb3ReadOnlyRequest(request)
	default:
		return nil, errors.New("invalid request type")
	}
}

func contains(item string, elems []string) bool {
	for _, x := range elems {
		if x == item {
			return true
		}
	}
	return false
}

// web3Call returns a response from a read-only eth RPC method
func (api *API) web3Call(request Web3SendAsyncReadOnlyRequest) (*Web3SendAsyncReadOnlyResponse, error) {
	var rpcResult interface{}
	var errMsg interface{}

	if request.Payload.Method == "personal_ecRecover" {
		data, err := hexutil.Decode(request.Payload.Params[0].(string))
		if err != nil {
			return nil, err
		}
		sig, err := hexutil.Decode(request.Payload.Params[1].(string))
		if err != nil {
			return nil, err
		}

		addr, err := api.EcRecover(data, sig)
		if err != nil {
			return nil, err
		}
		rpcResult = JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      request.Payload.ID,
			Result:  addr.String(),
		}
	} else {
		ethPayload, err := json.Marshal(request.Payload)
		if err != nil {
			return nil, err
		}

		response := api.s.rpcClient.CallRaw(string(ethPayload))
		if response == "" {
			errMsg = Web3ResponseError
		}
		rpcResult = json.RawMessage(response)
	}

	return &Web3SendAsyncReadOnlyResponse{
		ProviderResponse: ProviderResponse{
			ResponseType: Web3SendAsyncCallback,
		},
		MessageID: request.MessageID,
		Error:     errMsg,
		Result:    rpcResult,
	}, nil
}

func (api *API) web3NoPermission(request Web3SendAsyncReadOnlyRequest) (*Web3SendAsyncReadOnlyResponse, error) {
	return &Web3SendAsyncReadOnlyResponse{
		ProviderResponse: ProviderResponse{
			ResponseType: Web3SendAsyncCallback,
		},
		MessageID: request.MessageID,
		Error: Web3SendAsyncReadOnlyError{
			Code:    4100,
			Message: "The requested method and/or account has not been authorized by the user.",
		},
	}, nil
}

func (api *API) web3AccResponse(request Web3SendAsyncReadOnlyRequest) (*Web3SendAsyncReadOnlyResponse, error) {
	dappsAddress, err := api.s.accountsDB.GetDappsAddress()
	if err != nil {
		return nil, err
	}

	var result interface{}
	if request.Payload.Method == ethCoinbase {
		result = dappsAddress
	} else {
		result = []types.Address{dappsAddress}
	}

	return &Web3SendAsyncReadOnlyResponse{
		ProviderResponse: ProviderResponse{
			ResponseType: Web3SendAsyncCallback,
		},
		MessageID: request.MessageID,
		Result: JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      request.Payload.ID,
			Result:  result,
		},
	}, nil
}

func (api *API) getVerifiedWalletAccount(address, password string) (*account.SelectedExtKey, error) {
	exists, err := api.s.accountsDB.AddressExists(types.HexToAddress(address))
	if err != nil {
		log.Error("failed to query db for a given address", "address", address, "error", err)
		return nil, err
	}

	if !exists {
		log.Error("failed to get a selected account", "err", transactions.ErrInvalidTxSender)
		return nil, transactions.ErrAccountDoesntExist
	}

	key, err := api.s.accountsManager.VerifyAccountPassword(api.s.config.KeyStoreDir, address, password)
	if err != nil {
		log.Error("failed to verify account", "account", address, "error", err)
		return nil, err
	}

	return &account.SelectedExtKey{
		Address:    key.Address,
		AccountKey: key,
	}, nil
}

func (api *API) web3SignatureResponse(request Web3SendAsyncReadOnlyRequest) (*Web3SendAsyncReadOnlyResponse, error) {
	var err error
	var signature types.HexBytes
	if request.Payload.Method == "eth_signTypedData" || request.Payload.Method == "eth_signTypedData_v3" {
		raw := json.RawMessage(request.Payload.Params[1].(string))
		var data typeddata.TypedData
		err = json.Unmarshal(raw, &data)
		if err == nil {
			signature, err = api.signTypedData(data, request.Payload.From, request.Payload.Password)
		}
	} else if request.Payload.Method == "eth_signTypedData_v4" {
		signature, err = api.signTypedDataV4(request.Payload.Params[1].(signercore.TypedData), request.Payload.From, request.Payload.Password)
	} else {
		signature, err = api.signMessage(request.Payload.Params[0], request.Payload.From, request.Payload.Password)
	}

	if err != nil {
		log.Error("could not sign message", "err", err)
		return &Web3SendAsyncReadOnlyResponse{
			ProviderResponse: ProviderResponse{
				ResponseType: Web3SendAsyncCallback,
			},
			MessageID: request.MessageID,
			Error: Web3SendAsyncReadOnlyError{
				Code:    4100,
				Message: err.Error(),
			},
		}, nil
	}

	return &Web3SendAsyncReadOnlyResponse{
		ProviderResponse: ProviderResponse{
			ResponseType: Web3SendAsyncCallback,
		},
		MessageID: request.MessageID,
		Result: JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      request.Payload.ID,
			Result:  signature,
		},
	}, nil
}

func (api *API) ProcessWeb3ReadOnlyRequest(request Web3SendAsyncReadOnlyRequest) (*Web3SendAsyncReadOnlyResponse, error) {
	hasPermission, err := api.s.permissionsDB.HasPermission(request.Hostname, request.Address, PermissionWeb3)
	if err != nil {
		return nil, err
	}

	if contains(request.Payload.Method, authMethods) && !hasPermission {
		return api.web3NoPermission(request)
	}

	if contains(request.Payload.Method, accMethods) {
		return api.web3AccResponse(request)
	} else if contains(request.Payload.Method, signMethods) {
		return api.web3SignatureResponse(request)
	} else if request.Payload.Method == "eth_sendTransaction" {
		jsonString, err := json.Marshal(request.Payload.Params[0])
		if err != nil {
			return nil, err
		}

		var trxArgs transactions.SendTxArgs
		if err := json.Unmarshal(jsonString, &trxArgs); err != nil {
			return nil, err
		}

		hash, err := api.sendTransaction(request.Payload.ChainID, trxArgs, request.Payload.Password, Web3SendAsyncReadOnly)
		if err != nil {
			log.Error("could not send transaction message", "err", err)
			return &Web3SendAsyncReadOnlyResponse{
				ProviderResponse: ProviderResponse{
					ResponseType: Web3SendAsyncCallback,
				},
				MessageID: request.MessageID,
				Error:     Web3ResponseError,
			}, nil
		}

		return &Web3SendAsyncReadOnlyResponse{
			ProviderResponse: ProviderResponse{
				ResponseType: Web3SendAsyncCallback,
			},
			MessageID: request.MessageID,
			Result: JSONRPCResponse{
				JSONRPC: "2.0",
				ID:      request.Payload.ID,
				Result:  hash,
			},
		}, nil
	} else {
		return api.web3Call(request)
	}
}

func (api *API) ProcessAPIRequest(request APIRequest) (*APIResponse, error) {
	if request.Permission == "" {
		return nil, ErrorInvalidAPIRequest
	}
	hasPermission, err := api.s.permissionsDB.HasPermission(request.Hostname, request.Address, request.Permission)
	if err != nil {
		return nil, err
	}

	if !hasPermission {
		// Not allowed
		return &APIResponse{
			ProviderResponse: ProviderResponse{
				ResponseType: ResponseAPI,
			},
			Permission: request.Permission,
			MessageID:  request.MessageID,
			IsAllowed:  false,
		}, nil
	}
	var data interface{}
	switch request.Permission {
	case PermissionWeb3:
		dappsAddress, err := api.s.accountsDB.GetDappsAddress()
		if err != nil {
			return nil, err
		}
		response := make([]interface{}, 1)
		response[0] = dappsAddress
		data = response
	case PermissionContactCode:
		pubKey, err := api.s.accountsDB.GetPublicKey()
		if err != nil {
			return nil, err
		}
		data = pubKey
	default:
		return nil, ErrorUnknownPermission
	}
	return &APIResponse{
		ProviderResponse: ProviderResponse{
			ResponseType: ResponseAPI,
		},
		Permission: request.Permission,
		MessageID:  request.MessageID,
		Data:       data,
		IsAllowed:  true,
	}, nil
}
