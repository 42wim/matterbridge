// Copyright 2017-2019 Weald Technology Trading
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
	"errors"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/wealdtech/go-ens/v3/contracts/reverseregistrar"
)

// ReverseRegistrar is the structure for the reverse registrar
type ReverseRegistrar struct {
	Contract     *reverseregistrar.Contract
	ContractAddr common.Address
}

// NewReverseRegistrar obtains the reverse registrar
func NewReverseRegistrar(backend bind.ContractBackend) (*ReverseRegistrar, error) {
	registry, err := NewRegistry(backend)
	if err != nil {
		return nil, err
	}

	// Obtain the registry address from the registrar
	address, err := registry.Owner("addr.reverse")
	if err != nil {
		return nil, err
	}
	if address == UnknownAddress {
		return nil, errors.New("no registrar for that network")
	}
	return NewReverseRegistrarAt(backend, address)
}

// NewReverseRegistrarAt obtains the reverse registrar at a given address
func NewReverseRegistrarAt(backend bind.ContractBackend, address common.Address) (*ReverseRegistrar, error) {
	contract, err := reverseregistrar.NewContract(address, backend)
	if err != nil {
		return nil, err
	}
	return &ReverseRegistrar{
		Contract:     contract,
		ContractAddr: address,
	}, nil
}

// SetName sets the name
func (r *ReverseRegistrar) SetName(opts *bind.TransactOpts, name string) (tx *types.Transaction, err error) {
	return r.Contract.SetName(opts, name)
}

// DefaultResolverAddress obtains the default resolver address
func (r *ReverseRegistrar) DefaultResolverAddress() (common.Address, error) {
	return r.Contract.DefaultResolver(nil)
}
