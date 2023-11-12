// Copyright 2017 Weald Technology Trading
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
	"math/big"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/wealdtech/go-ens/v3/contracts/auctionregistrar"
	"github.com/wealdtech/go-ens/v3/contracts/registry"
	"github.com/wealdtech/go-ens/v3/util"
)

// Registry is the structure for the registry contract
type Registry struct {
	backend      bind.ContractBackend
	Contract     *registry.Contract
	ContractAddr common.Address
}

// NewRegistry obtains the ENS registry
func NewRegistry(backend bind.ContractBackend) (*Registry, error) {
	address, err := RegistryContractAddress(backend)
	if err != nil {
		return nil, err
	}
	return NewRegistryAt(backend, address)
}

// NewRegistryAt obtains the ENS registry at a given address
func NewRegistryAt(backend bind.ContractBackend, address common.Address) (*Registry, error) {
	contract, err := registry.NewContract(address, backend)
	if err != nil {
		return nil, err
	}
	return &Registry{
		backend:      backend,
		Contract:     contract,
		ContractAddr: address,
	}, nil
}

// Owner returns the address of the owner of a name
func (r *Registry) Owner(name string) (common.Address, error) {
	nameHash, err := NameHash(name)
	if err != nil {
		return UnknownAddress, err
	}
	return r.Contract.Owner(nil, nameHash)
}

// ResolverAddress returns the address of the resolver for a name
func (r *Registry) ResolverAddress(name string) (common.Address, error) {
	nameHash, err := NameHash(name)
	if err != nil {
		return UnknownAddress, err
	}
	return r.Contract.Resolver(nil, nameHash)
}

// SetResolver sets the resolver for a name
func (r *Registry) SetResolver(opts *bind.TransactOpts, name string, address common.Address) (*types.Transaction, error) {
	nameHash, err := NameHash(name)
	if err != nil {
		return nil, err
	}
	return r.Contract.SetResolver(opts, nameHash, address)
}

// Resolver returns the resolver for a name
func (r *Registry) Resolver(name string) (*Resolver, error) {
	address, err := r.ResolverAddress(name)
	if err != nil {
		return nil, err
	}
	return NewResolverAt(r.backend, name, address)
}

// SetOwner sets the ownership of a domain
func (r *Registry) SetOwner(opts *bind.TransactOpts, name string, address common.Address) (*types.Transaction, error) {
	nameHash, err := NameHash(name)
	if err != nil {
		return nil, err
	}
	return r.Contract.SetOwner(opts, nameHash, address)
}

// SetSubdomainOwner sets the ownership of a subdomain, potentially creating it in the process
func (r *Registry) SetSubdomainOwner(opts *bind.TransactOpts, name string, subname string, address common.Address) (*types.Transaction, error) {
	nameHash, err := NameHash(name)
	if err != nil {
		return nil, err
	}
	labelHash, err := LabelHash(subname)
	if err != nil {
		return nil, err
	}
	return r.Contract.SetSubnodeOwner(opts, nameHash, labelHash, address)
}

// RegistryContractAddress obtains the address of the registry contract for a chain.
// This is (currently) the same for all chains.
func RegistryContractAddress(backend bind.ContractBackend) (common.Address, error) {
	//	chainID := big.NewInt(0)
	//	if reflect.TypeOf(backend).String() == "*ethclient.Client" {
	//		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	//		defer cancel()
	//		var err error
	//		chainID, err = backend.(*ethclient.Client).NetworkID(ctx)
	//		if err != nil {
	//			return UnknownAddress, err
	//		}
	//	}

	// Instantiate the registry contract.  The same for all chains.
	return common.HexToAddress("00000000000C2E074eC69A0dFb2997BA6C7d2e1e"), nil
}

// RegistryContractFromRegistrar obtains the registry contract given an
// existing registrar contract
func RegistryContractFromRegistrar(backend bind.ContractBackend, registrar *auctionregistrar.Contract) (*registry.Contract, error) {
	if registrar == nil {
		return nil, errors.New("no registrar contract")
	}
	registryAddress, err := registrar.Ens(nil)
	if err != nil {
		return nil, err
	}
	return registry.NewContract(registryAddress, backend)
}

// SetResolver sets the resolver for a name
func SetResolver(session *registry.ContractSession, name string, resolverAddr *common.Address) (*types.Transaction, error) {
	nameHash, err := NameHash(name)
	if err != nil {
		return nil, err
	}
	return session.SetResolver(nameHash, *resolverAddr)
}

// SetSubdomainOwner sets the owner for a subdomain of a name
func SetSubdomainOwner(session *registry.ContractSession, name string, subdomain string, ownerAddr *common.Address) (*types.Transaction, error) {
	nameHash, err := NameHash(name)
	if err != nil {
		return nil, err
	}
	labelHash, err := LabelHash(subdomain)
	if err != nil {
		return nil, err
	}
	return session.SetSubnodeOwner(nameHash, labelHash, *ownerAddr)
}

// CreateRegistrySession creates a session suitable for multiple calls
func CreateRegistrySession(chainID *big.Int, wallet *accounts.Wallet, account *accounts.Account, passphrase string, contract *registry.Contract, gasPrice *big.Int) *registry.ContractSession {
	// Create a signer
	signer := util.AccountSigner(chainID, wallet, account, passphrase)

	// Return our session
	session := &registry.ContractSession{
		Contract: contract,
		CallOpts: bind.CallOpts{
			Pending: true,
		},
		TransactOpts: bind.TransactOpts{
			From:     account.Address,
			Signer:   signer,
			GasPrice: gasPrice,
		},
	}

	return session
}
