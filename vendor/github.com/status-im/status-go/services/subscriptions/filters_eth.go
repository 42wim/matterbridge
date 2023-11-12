package subscriptions

import (
	"fmt"

	"github.com/status-im/status-go/rpc"
)

type ethFilter struct {
	id        string
	rpcClient *rpc.Client
}

func installEthFilter(rpcClient *rpc.Client, method string, args []interface{}) (*ethFilter, error) {

	if err := validateEthMethod(method); err != nil {
		return nil, err
	}

	var result string

	err := rpcClient.Call(&result, rpcClient.UpstreamChainID, method, args...)

	if err != nil {
		return nil, err
	}

	filter := &ethFilter{
		id:        result,
		rpcClient: rpcClient,
	}

	return filter, nil

}

func (ef *ethFilter) getID() string {
	return ef.id
}

func (ef *ethFilter) getChanges() ([]interface{}, error) {
	var result []interface{}

	err := ef.rpcClient.Call(&result, ef.rpcClient.UpstreamChainID, "eth_getFilterChanges", ef.getID())

	return result, err
}

func (ef *ethFilter) uninstall() error {
	return ef.rpcClient.Call(nil, ef.rpcClient.UpstreamChainID, "eth_uninstallFilter", ef.getID())
}

func validateEthMethod(method string) error {
	for _, allowedMethod := range []string{
		"eth_newFilter",
		"eth_newBlockFilter",
		"eth_newPendingTransactionFilter",
	} {
		if method == allowedMethod {
			return nil
		}
	}

	return fmt.Errorf("unexpected filter method: %s", method)
}
