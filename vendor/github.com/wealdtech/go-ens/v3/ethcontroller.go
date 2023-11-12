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
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/wealdtech/go-ens/v3/contracts/ethcontroller"
)

// ETHController is the structure for the .eth controller contract
type ETHController struct {
	backend      bind.ContractBackend
	Contract     *ethcontroller.Contract
	ContractAddr common.Address
	domain       string
}

// NewETHController creates a new controller for a given domain
func NewETHController(backend bind.ContractBackend, domain string) (*ETHController, error) {
	registry, err := NewRegistry(backend)
	if err != nil {
		return nil, err
	}
	resolver, err := registry.Resolver(domain)
	if err != nil {
		return nil, err
	}

	// Obtain the controller from the resolver
	controllerAddress, err := resolver.InterfaceImplementer([4]byte{0x01, 0x8f, 0xac, 0x06})
	if err != nil {
		return nil, err
	}

	return NewETHControllerAt(backend, domain, controllerAddress)
}

// NewETHControllerAt creates a .eth controller at a given address
func NewETHControllerAt(backend bind.ContractBackend, domain string, address common.Address) (*ETHController, error) {
	contract, err := ethcontroller.NewContract(address, backend)
	if err != nil {
		return nil, err
	}
	return &ETHController{
		backend:      backend,
		Contract:     contract,
		ContractAddr: address,
		domain:       domain,
	}, nil
}

// IsValid returns true if the domain is considered valid by the controller.
func (c *ETHController) IsValid(domain string) (bool, error) {
	name, err := UnqualifiedName(domain, c.domain)
	if err != nil {
		return false, fmt.Errorf("invalid name %s", domain)
	}
	return c.Contract.Valid(nil, name)
}

// IsAvailable returns true if the domain is available for registration.
func (c *ETHController) IsAvailable(domain string) (bool, error) {
	name, err := UnqualifiedName(domain, c.domain)
	if err != nil {
		return false, fmt.Errorf("invalid name %s", domain)
	}
	return c.Contract.Available(nil, name)
}

// MinRegistrationDuration returns the minimum duration for which a name can be registered
func (c *ETHController) MinRegistrationDuration() (time.Duration, error) {
	tmp, err := c.Contract.MINREGISTRATIONDURATION(nil)
	if err != nil {
		return 0 * time.Second, err
	}

	return time.Duration(tmp.Int64()) * time.Second, nil
}

// RentCost returns the cost of rent in wei-per-second.
func (c *ETHController) RentCost(domain string) (*big.Int, error) {
	name, err := UnqualifiedName(domain, c.domain)
	if err != nil {
		return nil, fmt.Errorf("invalid name %s", domain)
	}
	return c.Contract.RentPrice(nil, name, big.NewInt(1))
}

// MinCommitmentInterval returns the minimum time that has to pass between a commit and reveal
func (c *ETHController) MinCommitmentInterval() (*big.Int, error) {
	return c.Contract.MinCommitmentAge(nil)
}

// MaxCommitmentInterval returns the maximum time that has to pass between a commit and reveal
func (c *ETHController) MaxCommitmentInterval() (*big.Int, error) {
	return c.Contract.MaxCommitmentAge(nil)
}

// CommitmentHash returns the commitment hash for a label/owner/secret tuple
func (c *ETHController) CommitmentHash(domain string, owner common.Address, secret [32]byte) (common.Hash, error) {
	name, err := UnqualifiedName(domain, c.domain)
	if err != nil {
		return common.BytesToHash([]byte{}), fmt.Errorf("invalid name %s", domain)
	}

	commitment, err := c.Contract.MakeCommitment(nil, name, owner, secret)
	if err != nil {
		return common.BytesToHash([]byte{}), err
	}
	hash := common.BytesToHash(commitment[:])
	return hash, err
}

// CommitmentTime states the time at which a commitment was registered on the blockchain.
func (c *ETHController) CommitmentTime(domain string, owner common.Address, secret [32]byte) (*big.Int, error) {
	hash, err := c.CommitmentHash(domain, owner, secret)
	if err != nil {
		return nil, err
	}

	return c.Contract.Commitments(nil, hash)
}

// Commit sends a commitment to register a domain.
func (c *ETHController) Commit(opts *bind.TransactOpts, domain string, owner common.Address, secret [32]byte) (*types.Transaction, error) {
	name, err := UnqualifiedName(domain, c.domain)
	if err != nil {
		return nil, fmt.Errorf("invalid name %s", domain)
	}

	commitment, err := c.Contract.MakeCommitment(nil, name, owner, secret)
	if err != nil {
		return nil, errors.New("failed to create commitment")
	}

	if opts.Value != nil && opts.Value.Cmp(big.NewInt(0)) != 0 {
		return nil, errors.New("commitment should have 0 value")
	}

	return c.Contract.Commit(opts, commitment)
}

// Reveal reveals a commitment to register a domain.
func (c *ETHController) Reveal(opts *bind.TransactOpts, domain string, owner common.Address, secret [32]byte) (*types.Transaction, error) {
	name, err := UnqualifiedName(domain, c.domain)
	if err != nil {
		return nil, fmt.Errorf("invalid name %s", domain)
	}

	if opts == nil {
		return nil, errors.New("transaction options required")
	}
	if opts.Value == nil {
		return nil, errors.New("no ether supplied with transaction")
	}

	commitTS, err := c.CommitmentTime(name, owner, secret)
	if err != nil {
		return nil, err
	}
	if commitTS.Cmp(big.NewInt(0)) == 0 {
		return nil, errors.New("no commitment present")
	}
	commit := time.Unix(commitTS.Int64(), 0)

	minCommitIntervalTS, err := c.MinCommitmentInterval()
	if err != nil {
		return nil, err
	}
	minCommitInterval := time.Duration(minCommitIntervalTS.Int64()) * time.Second
	minRevealTime := commit.Add(minCommitInterval)
	if time.Now().Before(minRevealTime) {
		return nil, errors.New("commitment too young to reveal")
	}

	maxCommitIntervalTS, err := c.MaxCommitmentInterval()
	if err != nil {
		return nil, err
	}
	maxCommitInterval := time.Duration(maxCommitIntervalTS.Int64()) * time.Second
	maxRevealTime := commit.Add(maxCommitInterval)
	if time.Now().After(maxRevealTime) {
		return nil, errors.New("commitment too old to reveal")
	}

	// Calculate the duration given the rent cost and the value
	costPerSecond, err := c.RentCost(domain)
	if err != nil {
		return nil, errors.New("failed to obtain rent cost")
	}
	duration := new(big.Int).Div(opts.Value, costPerSecond)

	// Ensure duration is greater than minimum duration
	minDuration, err := c.MinRegistrationDuration()
	if err != nil {
		return nil, err
	}
	if big.NewInt(int64(minDuration.Seconds())).Cmp(duration) >= 0 {
		return nil, fmt.Errorf("not enough funds to cover minimum duration of %v", minDuration)
	}

	return c.Contract.Register(opts, name, owner, duration, secret)
}

// Renew renews a registered domain.
func (c *ETHController) Renew(opts *bind.TransactOpts, domain string) (*types.Transaction, error) {
	name, err := UnqualifiedName(domain, c.domain)
	if err != nil {
		return nil, fmt.Errorf("invalid name %s", domain)
	}

	// See if we're registered at all - fetch the owner to find out
	registry, err := NewRegistry(c.backend)
	if err != nil {
		return nil, err
	}
	owner, err := registry.Owner(domain)
	if err != nil {
		return nil, err
	}
	if owner == UnknownAddress {
		return nil, fmt.Errorf("%s not registered", domain)
	}

	// Calculate the duration given the rent cost and the value
	costPerSecond, err := c.RentCost(domain)
	if err != nil {
		return nil, errors.New("failed to obtain rent cost")
	}
	duration := new(big.Int).Div(opts.Value, costPerSecond)

	return c.Contract.Renew(opts, name, duration)
}
