// Copyright (c) 2021 Tulir Asokan
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package appservice

import "sync"

type TransactionIDCache struct {
	array    []string
	arrayPtr int
	hash     map[string]struct{}
	lock     sync.RWMutex
}

func NewTransactionIDCache(size int) *TransactionIDCache {
	return &TransactionIDCache{
		array: make([]string, size),
		hash:  make(map[string]struct{}),
	}
}

func (txnIDC *TransactionIDCache) IsProcessed(txnID string) bool {
	txnIDC.lock.RLock()
	_, exists := txnIDC.hash[txnID]
	txnIDC.lock.RUnlock()
	return exists
}

func (txnIDC *TransactionIDCache) MarkProcessed(txnID string) {
	txnIDC.lock.Lock()
	txnIDC.hash[txnID] = struct{}{}
	if txnIDC.array[txnIDC.arrayPtr] != "" {
		for i := 0; i < len(txnIDC.array)/8; i++ {
			delete(txnIDC.hash, txnIDC.array[txnIDC.arrayPtr+i])
			txnIDC.array[txnIDC.arrayPtr+i] = ""
		}
	}
	txnIDC.array[txnIDC.arrayPtr] = txnID
	txnIDC.lock.Unlock()
}
