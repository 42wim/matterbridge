package proxy

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"unsafe"

	"github.com/BurntSushi/toml"
)

/*
#include <stdlib.h>
#include "verifproxy.h"

typedef void (*callback_type)(char *);
void goCallback_cgo(char *);

*/
import "C"

type Web3UrlType struct {
	Kind    string `toml:"kind"`
	Web3Url string `toml:"web3Url"`
}
type Config struct {
	Eth2Network      string `toml:"network"`
	TrustedBlockRoot string `toml:"trusted-block-root"`
	// Web3Url          Web3UrlType `toml:"web3-url"`
	Web3Url    string `toml:"web3-url"`
	RpcAddress string `toml:"rpc-address"`
	RpcPort    uint16 `toml:"rpc-port"`
	LogLevel   string `toml:"log-level"`
}

type BeaconBlockHeader struct {
	Slot          uint64 `json:"slot"`
	ProposerIndex uint64 `json:"proposer_index"`
	ParentRoot    string `json:"parent_root"`
	StateRoot     string `json:"state_root"`
}

//export goCallback
func goCallback(json *C.char) {
	goStr := C.GoString(json)
	//C.free(unsafe.Pointer(json))
	fmt.Println("### goCallback " + goStr)
	// var hdr BeaconBlockHeader
	// err := json.NewDecoder([]byte(goStr)).Decode(&hdr)
	// if err != nil {
	// 	fmt.Println("### goCallback json parse error: " + err)
	// }
	// fmt.Println("Unmarshal result: " + hdr)
}
func createTomlFile(cfg *Config) string {
	var buffer bytes.Buffer
	err := toml.NewEncoder(&buffer).Encode(cfg)
	if err != nil {
		return ""
	}
	tomlFileName := "config.toml"
	f, err := os.Create(tomlFileName)
	if err != nil {
		return ""
	}
	defer f.Close()
	f.WriteString(buffer.String())

	return tomlFileName
}

func StartLightClient(ctx context.Context, cfg *Config) {
	fmt.Println("vim-go")
	cb := (C.callback_type)(unsafe.Pointer(C.goCallback_cgo))
	C.setOptimisticHeaderCallback(cb)
	C.setFinalizedHeaderCallback(cb)
	fmt.Println("vim-go 2")

	go func() {
		runtime.LockOSThread()
		// tomlFileName := createTomlFile(cfg)
		// configCStr := C.CString(tomlFileName)
		// C.startLc(configCStr)
		defer runtime.UnlockOSThread()
		jsonBytes, _ := json.Marshal(cfg)
		jsonStr := string(jsonBytes)
		fmt.Println("### jsonStr: ", jsonStr)
		configCStr := C.CString(jsonStr)
		C.startProxyViaJson(configCStr)
		fmt.Println("inside go-func after startLcViaJson")
	}()
	go func() {
		fmt.Println("Before range ctx.Done()")
		for range ctx.Done() {
			fmt.Println("inside go-func ctx.Done()")
			C.quit()
		}
	}()
	fmt.Println("vim-go 3")

}
