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
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/wealdtech/go-ens/v3/contracts/auctionregistrar"
)

// AuctionRegistrar is the structure for the auction registrar contract
type AuctionRegistrar struct {
	backend      bind.ContractBackend
	domain       string
	Contract     *auctionregistrar.Contract
	ContractAddr common.Address
}

// AuctionEntry is an auction entry
type AuctionEntry struct {
	State        string
	Deed         common.Address
	Registration time.Time
	Value        *big.Int
	HighestBid   *big.Int
}

// NewAuctionRegistrar creates a new auction registrar for a given domain
func NewAuctionRegistrar(backend bind.ContractBackend, domain string) (*AuctionRegistrar, error) {
	address, err := RegistrarContractAddress(backend, domain)
	if err != nil {
		return nil, err
	}

	return NewAuctionRegistrarAt(backend, domain, address)
}

// NewAuctionRegistrarAt creates an auction registrar for a given domain at a given address
func NewAuctionRegistrarAt(backend bind.ContractBackend, domain string, address common.Address) (*AuctionRegistrar, error) {
	contract, err := auctionregistrar.NewContract(address, backend)
	if err != nil {
		return nil, err
	}
	return &AuctionRegistrar{
		backend:      backend,
		domain:       domain,
		Contract:     contract,
		ContractAddr: address,
	}, nil
}

// State returns the state of a nam
func (r *AuctionRegistrar) State(name string) (string, error) {
	entry, err := r.Entry(name)
	if err != nil {
		return "", err
	}
	if entry == nil {
		return "", fmt.Errorf("no entry for %s", name)
	}
	return entry.State, nil
}

// Entry obtains a registrar entry for a name
func (r *AuctionRegistrar) Entry(domain string) (*AuctionEntry, error) {
	name, err := UnqualifiedName(domain, r.domain)
	if err != nil {
		return nil, fmt.Errorf("invalid name %s", domain)
	}

	labelHash, err := LabelHash(name)
	if err != nil {
		return nil, err
	}
	status, deedAddress, registration, value, highestBid, err := r.Contract.Entries(nil, labelHash)
	if err != nil {
		return nil, err
	}

	entry := &AuctionEntry{
		Deed:       deedAddress,
		Value:      value,
		HighestBid: highestBid,
	}
	entry.Registration = time.Unix(registration.Int64(), 0)
	switch status {
	case 0:
		entry.State = "Available"
	case 1:
		entry.State = "Bidding"
	case 2:
		// Might be won or owned
		registry, err := NewRegistry(r.backend)
		if err != nil {
			return nil, err
		}

		owner, err := registry.Owner(domain)
		if err != nil {
			return nil, err
		}

		if owner == UnknownAddress {
			entry.State = "Won"
		} else {
			entry.State = "Owned"
		}
	case 3:
		entry.State = "Forbidden"
	case 4:
		entry.State = "Revealing"
	case 5:
		entry.State = "Unavailable"
	default:
		entry.State = "Unknown"
	}

	return entry, nil
}

// Migrate migrates a domain to the permanent registrar
func (r *AuctionRegistrar) Migrate(opts *bind.TransactOpts, domain string) (*types.Transaction, error) {
	name, err := UnqualifiedName(domain, r.domain)
	if err != nil {
		return nil, fmt.Errorf("invalid name %s", domain)
	}

	labelHash, err := LabelHash(name)
	if err != nil {
		return nil, err
	}
	return r.Contract.TransferRegistrars(opts, labelHash)
}

// Release releases a domain
func (r *AuctionRegistrar) Release(opts *bind.TransactOpts, domain string) (*types.Transaction, error) {
	name, err := UnqualifiedName(domain, r.domain)
	if err != nil {
		return nil, fmt.Errorf("invalid name %s", domain)
	}

	labelHash, err := LabelHash(name)
	if err != nil {
		return nil, err
	}
	return r.Contract.ReleaseDeed(opts, labelHash)
}

// Owner obtains the owner of the deed that represents the name.
func (r *AuctionRegistrar) Owner(domain string) (common.Address, error) {
	name, err := UnqualifiedName(domain, r.domain)
	if err != nil {
		return UnknownAddress, err
	}

	entry, err := r.Entry(name)
	if err != nil {
		return UnknownAddress, err
	}

	deed, err := NewDeedAt(r.backend, entry.Deed)
	if err != nil {
		return UnknownAddress, err
	}
	return deed.Owner()
}

// SetOwner sets the owner of the deed that represents the name.
func (r *AuctionRegistrar) SetOwner(opts *bind.TransactOpts, domain string, address common.Address) (*types.Transaction, error) {
	name, err := UnqualifiedName(domain, r.domain)
	if err != nil {
		return nil, fmt.Errorf("invalid name %s", domain)
	}

	labelHash, err := LabelHash(name)
	if err != nil {
		return nil, err
	}
	return r.Contract.Transfer(opts, labelHash, address)
}

// ShaBid calculates the hash for a bid.
func (r *AuctionRegistrar) ShaBid(hash [32]byte, address common.Address, value *big.Int, salt [32]byte) ([32]byte, error) {
	return r.Contract.ShaBid(nil, hash, address, value, salt)
}
