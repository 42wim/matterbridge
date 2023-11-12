// Copyright 2019 Weald Technology Trading
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package ens

import (
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/wealdtech/go-ens/v3/contracts/dnsregistrar"
)

// DNSRegistrar is the structure for the registrar
type DNSRegistrar struct {
	backend      bind.ContractBackend
	domain       string
	Contract     *dnsregistrar.Contract
	ContractAddr common.Address
}

// NewDNSRegistrar obtains the registrar contract for a given domain
func NewDNSRegistrar(backend bind.ContractBackend, domain string) (*DNSRegistrar, error) {
	address, err := RegistrarContractAddress(backend, domain)
	if err != nil {
		return nil, err
	}

	if address == UnknownAddress {
		return nil, fmt.Errorf("no registrar for domain %s", domain)
	}

	contract, err := dnsregistrar.NewContract(address, backend)
	if err != nil {
		return nil, err
	}

	// Ensure this really is a DNS registrar.  To do this confirm that it supports
	// the expected interface.
	supported, err := contract.SupportsInterface(nil, [4]byte{0x1a, 0xa2, 0xe6, 0x41})
	if err != nil {
		return nil, err
	}
	if !supported {
		return nil, fmt.Errorf("purported registrar for domain %s does not support DNS registrar functionality", domain)
	}

	return &DNSRegistrar{
		backend:      backend,
		domain:       domain,
		Contract:     contract,
		ContractAddr: address,
	}, nil
}

// TODO claim
