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
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// Name represents an ENS name, for example 'foo.bar.eth'.
type Name struct {
	backend bind.ContractBackend
	// Name is the fully-qualified name of an ENS domain e.g. foo.bar.eth
	Name string
	// Domain is the domain of an ENS domain e.g. bar.eth
	Domain string
	// Label is the name part of an ENS domain e.g. foo
	Label string
	// Contracts
	registry   *Registry
	registrar  *BaseRegistrar
	controller *ETHController
}

// NewName creates an ENS name structure.
// Note that this does not create the name on-chain.
func NewName(backend bind.ContractBackend, name string) (*Name, error) {
	name, err := NormaliseDomain(name)
	if err != nil {
		return nil, err
	}
	domain := Domain(name)
	label, err := DomainPart(name, 1)
	if err != nil {
		return nil, err
	}

	registry, err := NewRegistry(backend)
	if err != nil {
		return nil, err
	}
	registrar, err := NewBaseRegistrar(backend, domain)
	if err != nil {
		return nil, err
	}
	controller, err := NewETHController(backend, domain)
	if err != nil {
		return nil, err
	}

	isValid, err := controller.IsValid(name)
	if err != nil {
		return nil, err
	}
	if !isValid {
		return nil, errors.New("name is not valid according to the rules of the registrar (too short, invalid characters, etc.)")
	}

	return &Name{
		backend:    backend,
		Name:       name,
		Domain:     domain,
		Label:      label,
		registry:   registry,
		registrar:  registrar,
		controller: controller,
	}, nil
}

// IsRegistered returns true if the name is registered in the registrar
func (n *Name) IsRegistered() (bool, error) {
	registrant, err := n.Registrant()
	if err != nil {
		return false, err
	}
	return registrant != UnknownAddress, nil
}

// ExtendRegistration sends a transaction that extends the registration of the name.
func (n *Name) ExtendRegistration(opts *bind.TransactOpts) (*types.Transaction, error) {
	isRegistered, err := n.IsRegistered()
	if err != nil {
		return nil, err
	}
	if !isRegistered {
		return nil, errors.New("name is not registered")
	}

	rentCost, err := n.RentCost()
	if err != nil {
		return nil, err
	}

	if opts.Value == nil || opts.Value.Cmp(rentCost) < 0 {
		return nil, errors.New("not enough funds to extend the registration")
	}

	return n.controller.Renew(opts, n.Name)
}

// RegistrationInterval obtains the minimum interval between commit and reveal
// when registering this name.
func (n *Name) RegistrationInterval() (time.Duration, error) {
	interval, err := n.controller.MinCommitmentInterval()
	if err != nil {
		return time.Duration(0), err
	}
	return time.Duration(interval.Int64()) * time.Second, nil
}

// RegisterStageOne sends a transaction that starts the registration process.
func (n *Name) RegisterStageOne(registrant common.Address, opts *bind.TransactOpts) (*types.Transaction, [32]byte, error) {
	var secret [32]byte
	_, err := rand.Read(secret[:])
	if err != nil {
		return nil, secret, err
	}

	isRegistered, err := n.IsRegistered()
	if err != nil {
		return nil, secret, err
	}
	if isRegistered {
		return nil, secret, errors.New("name is already registered")
	}

	signedTx, err := n.controller.Commit(opts, n.Label, registrant, secret)
	return signedTx, secret, err
}

// RegisterStageTwo sends a transaction that completes the registration process.
// The registrant must be the same as supplied in RegisterStageOne.
// The secret is that returned by RegisterStageOne.
// At least RegistrationInterval() time must have passed since the stage one
// transaction was mined for this to work.
func (n *Name) RegisterStageTwo(registrant common.Address, secret [32]byte, opts *bind.TransactOpts) (*types.Transaction, error) {
	commitTS, err := n.controller.CommitmentTime(n.Label, registrant, secret)
	if err != nil {
		return nil, err
	}
	if commitTS.Cmp(big.NewInt(0)) == 0 {
		return nil, errors.New("stage 2 attempted prior to successful stage 1 transaction")
	}
	commit := time.Unix(commitTS.Int64(), 0)

	minCommitIntervalTS, err := n.controller.MinCommitmentInterval()
	if err != nil {
		return nil, err
	}
	minCommitInterval := time.Duration(minCommitIntervalTS.Int64()) * time.Second
	minStageTwoTime := commit.Add(minCommitInterval)
	if time.Now().Before(minStageTwoTime) {
		return nil, errors.New("too early to send second transaction")
	}

	maxCommitIntervalTS, err := n.controller.MaxCommitmentInterval()
	if err != nil {
		return nil, err
	}
	maxCommitInterval := time.Duration(maxCommitIntervalTS.Int64()) * time.Second
	maxStageTwoTime := commit.Add(maxCommitInterval)
	if time.Now().After(maxStageTwoTime) {
		return nil, errors.New("too late to send second transaction")
	}

	return n.controller.Reveal(opts, n.Label, registrant, secret)
}

// Expires obtain the time at which the registration for this name expires.
func (n *Name) Expires() (time.Time, error) {
	expiryTS, err := n.registrar.Expiry(n.Label)
	if err != nil {
		return time.Unix(0, 0), err
	}

	if expiryTS.Int64() == 0 {
		return time.Unix(0, 0), errors.New("not registered")
	}

	return time.Unix(expiryTS.Int64(), 0), nil
}

// Controller obtains the controller for this name.
// The controller can carry out operations on the name such as setting
// records, but cannot transfer ultimate ownership of the name.
func (n *Name) Controller() (common.Address, error) {
	return n.registry.Owner(n.Name)
}

// SetController sets the controller for this name.
// The controller can carry out operations on the name such as setting
// records, but cannot transfer ultimate ownership of the name.
func (n *Name) SetController(controller common.Address, opts *bind.TransactOpts) (*types.Transaction, error) {
	// Are we the current controller?
	curController, err := n.Controller()
	if err != nil {
		return nil, err
	}
	if curController == opts.From {
		return n.registry.SetOwner(opts, n.Name, controller)
	}

	// Perhaps we are the registrant
	registrant, err := n.Registrant()
	if err != nil {
		return nil, err
	}
	// Are we actually trying to reclaim?
	if registrant == opts.From && opts.From == controller {
		return n.Reclaim(opts)
	}

	return nil, errors.New("not authorised to change the controller")
}

// Reclaim reclaims controller rights by the registrant
func (n *Name) Reclaim(opts *bind.TransactOpts) (*types.Transaction, error) {
	// Ensure the we are the registrant
	registrant, err := n.Registrant()
	if err != nil {
		return nil, err
	}
	if registrant != opts.From {
		return nil, errors.New("not the registrant")
	}
	return n.registrar.Reclaim(opts, n.Name, registrant)
}

// Registrant obtains the registrant for this name.
func (n *Name) Registrant() (common.Address, error) {
	return n.registrar.Owner(n.Label)
}

// Transfer transfers the registration of this name to a new registrant.
func (n *Name) Transfer(registrant common.Address, opts *bind.TransactOpts) (*types.Transaction, error) {
	// Ensure the we are the registrant
	currentRegistrant, err := n.Registrant()
	if err != nil {
		return nil, err
	}
	if currentRegistrant != opts.From {
		return nil, errors.New("not the current registrant")
	}
	return n.registrar.SetOwner(opts, n.Label, registrant)
}

// RentCost returns the cost of rent in Wei-per-second.
func (n *Name) RentCost() (*big.Int, error) {
	return n.controller.RentCost(n.Label)
}

// CreateSubdomain creates a subdomain on the name.
func (n *Name) CreateSubdomain(label string, controller common.Address, opts *bind.TransactOpts) (*types.Transaction, error) {
	// Confirm the subdomain does not already exist
	fqdn := fmt.Sprintf("%s.%s", label, n.Name)
	subdomainController, err := n.registry.Owner(fqdn)
	if err != nil {
		return nil, err
	}
	if subdomainController != UnknownAddress {
		return nil, errors.New("that subdomain already exists")
	}

	return n.registry.SetSubdomainOwner(opts, n.Name, label, controller)
}

// ResolverAddress fetches the address of the resolver contract for the name.
func (n *Name) ResolverAddress() (common.Address, error) {
	return n.registry.ResolverAddress(n.Name)
}

// SetResolverAddress sets the resolver contract address for the name.
func (n *Name) SetResolverAddress(address common.Address, opts *bind.TransactOpts) (*types.Transaction, error) {
	return n.registry.SetResolver(opts, n.Name, address)
}

// Address fetches the address of the name for a given coin type.
// Coin types are defined at https://github.com/satoshilabs/slips/blob/master/slip-0044.md
func (n *Name) Address(coinType uint64) ([]byte, error) {
	resolver, err := NewResolver(n.backend, n.Name)
	if err != nil {
		return nil, err
	}
	return resolver.MultiAddress(coinType)
}

// SetAddress sets the address of the name for a given coin type.
// Coin types are defined at https://github.com/satoshilabs/slips/blob/master/slip-0044.md
func (n *Name) SetAddress(coinType uint64, address []byte, opts *bind.TransactOpts) (*types.Transaction, error) {
	resolver, err := NewResolver(n.backend, n.Name)
	if err != nil {
		return nil, err
	}
	return resolver.SetMultiAddress(opts, coinType, address)
}
