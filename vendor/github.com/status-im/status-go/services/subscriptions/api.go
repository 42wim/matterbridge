package subscriptions

import (
	"fmt"
	"time"

	"github.com/status-im/status-go/rpc"
)

type API struct {
	rpcPrivateClientFunc func() *rpc.Client
	activeSubscriptions  *Subscriptions
}

func NewPublicAPI(rpcPrivateClientFunc func() *rpc.Client) *API {
	return &API{
		rpcPrivateClientFunc: rpcPrivateClientFunc,
		activeSubscriptions:  NewSubscriptions(100 * time.Millisecond),
	}
}

func (api *API) SubscribeSignal(method string, args []interface{}) (SubscriptionID, error) {
	var (
		filter    filter
		err       error
		namespace = method[:3]
	)

	switch namespace {
	case "shh":
		filter, err = installShhFilter(api.rpcPrivateClientFunc(), method, args)
	case "eth":
		filter, err = installEthFilter(api.rpcPrivateClientFunc(), method, args)
	default:
		err = fmt.Errorf("unexpected namespace: %s", namespace)
	}

	if err != nil {
		return "", fmt.Errorf("[SubscribeSignal] could not subscribe, failed to call %s: %v", method, err)
	}

	return api.activeSubscriptions.Create(namespace, filter)
}

func (api *API) UnsubscribeSignal(id string) error {
	return api.activeSubscriptions.Remove(SubscriptionID(id))
}
