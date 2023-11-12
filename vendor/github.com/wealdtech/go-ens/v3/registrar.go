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
)

// RegistrarContractAddress obtains the registrar contract address for a given domain
func RegistrarContractAddress(backend bind.ContractBackend, domain string) (common.Address, error) {
	// Obtain a registry contract
	registry, err := NewRegistry(backend)
	if err != nil {
		return UnknownAddress, err
	}

	// Obtain the registrar address from the registry
	address, err := registry.Owner(domain)
	if address == UnknownAddress {
		err = fmt.Errorf("no registrar for %s", domain)
	}

	return address, err
}
