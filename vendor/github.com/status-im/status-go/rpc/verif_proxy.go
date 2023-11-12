//go:build nimbus_light_client
// +build nimbus_light_client

package rpc

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/log"

	gethrpc "github.com/ethereum/go-ethereum/rpc"

	proxy "github.com/siphiuel/lc-proxy-wrapper"

	"github.com/status-im/status-go/params"
)

type VerifProxy struct {
	config *proxy.Config
	client *gethrpc.Client
	log    log.Logger
}

func init() {
	verifProxyInitFn = func(c *Client) {
		ctx := context.Background()
		var testConfig = proxy.Config{
			Eth2Network:      "mainnet",
			TrustedBlockRoot: "0xc5182cdb750fe088138b0d475683cda26a96befc24de16fb17bcf49d9cadf2f7",
			Web3Url:          c.upstreamURL,
			RpcAddress:       "127.0.0.1",
			RpcPort:          8545,
			LogLevel:         "INFO",
		}
		proxy.StartLightClient(ctx, &testConfig)
		verifProxy, err := newVerifProxy(&testConfig, c.log)
		if err != nil {
			c.RegisterHandler(
				params.BalanceMethodName,
				func(ctx context.Context, v uint64, params ...interface{}) (interface{}, error) {
					addr := params[0].(common.Address)
					return verifProxy.GetBalance(ctx, addr)
				},
			)
		}
	}
}

func newVerifProxy(cfg *proxy.Config, log log.Logger) (*VerifProxy, error) {
	endpoint := "http://" + cfg.RpcAddress + ":" + fmt.Sprint(cfg.RpcPort)
	client, err := gethrpc.DialHTTP(endpoint)
	if err != nil {
		log.Error("Error when creating VerifProxy client", err)
		return nil, err
	}
	proxy := &VerifProxy{cfg, client, log}
	return proxy, nil
}

func (p *VerifProxy) GetBalance(ctx context.Context, address common.Address) (interface{}, error) {
	var result hexutil.Big
	err := p.client.CallContext(ctx, &result, "eth_getBalance", address, "latest")
	if err != nil {
		p.log.Error("Error when invoking GetBalance", err)
		return nil, err
	}
	return result, nil

}
