package subscriptions

import (
	"fmt"

	"github.com/status-im/status-go/rpc"
)

type whisperFilter struct {
	id        string
	rpcClient *rpc.Client
}

func installShhFilter(rpcClient *rpc.Client, method string, args []interface{}) (*whisperFilter, error) {

	if err := validateShhMethod(method); err != nil {
		return nil, err
	}

	var result string

	err := rpcClient.Call(&result, rpcClient.UpstreamChainID, method, args...)

	if err != nil {
		return nil, err
	}

	filter := &whisperFilter{
		id:        result,
		rpcClient: rpcClient,
	}

	return filter, nil
}

func (wf *whisperFilter) getChanges() ([]interface{}, error) {
	var result []interface{}

	err := wf.rpcClient.Call(&result, wf.rpcClient.UpstreamChainID, "shh_getFilterMessages", wf.getID())

	return result, err
}

func (wf *whisperFilter) getID() string {
	return wf.id
}

func (wf *whisperFilter) uninstall() error {
	return wf.rpcClient.Call(nil, wf.rpcClient.UpstreamChainID, "shh_deleteMessageFilter", wf.getID())
}

func validateShhMethod(method string) error {
	if method != "shh_newMessageFilter" {
		return fmt.Errorf("unexpected filter method: %s", method)
	}
	return nil
}
