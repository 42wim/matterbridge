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
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/wealdtech/go-ens/v3/contracts/dnssecoracle"
)

// DNSSECOracle is the structure for the DNSSEC oracle
type DNSSECOracle struct {
	backend      bind.ContractBackend
	domain       string
	Contract     *dnssecoracle.Contract
	ContractAddr common.Address
}

// NewDNSSECOracle obtains the DNSSEC oracle contract for a given domain
func NewDNSSECOracle(backend bind.ContractBackend, domain string) (*DNSSECOracle, error) {
	registrar, err := NewDNSRegistrar(backend, domain)
	if err != nil {
		return nil, err
	}

	address, err := registrar.Contract.Oracle(nil)
	if err != nil {
		return nil, err
	}

	contract, err := dnssecoracle.NewContract(address, backend)
	if err != nil {
		return nil, err
	}

	return &DNSSECOracle{
		backend:      backend,
		domain:       domain,
		Contract:     contract,
		ContractAddr: address,
	}, nil
}
