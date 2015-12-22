// Copyright 2012-2013 Apcera Inc. All rights reserved.

package esMap

import (
	"errors"
	"math/rand"
	"time"
)

// We use init to setup the random number generator
func init() {
	rand.Seed(time.Now().UnixNano())
}

type HashCache struct {
	HashMap
}

// New creates a new HashMap of default size and using the default
// Hashing algorithm.
func NewHashCache() *HashCache {
	h, _ := NewHashCacheWithBkts(make([]*Entry, _BSZ))
	return h
}

// NewWithBkts creates a new HashMap using the bkts slice argument.
// len(bkts) must be a power of 2.
func NewHashCacheWithBkts(bkts []*Entry) (*HashCache, error) {
	l := len(bkts)
	if l == 0 || (l&(l-1) != 0) {
		return nil, errors.New("Size of buckets must be power of 2")
	}
	h := HashCache{}
	h.msk = uint32(l - 1)
	h.bkts = bkts
	h.Hash = DefaultHash
	h.rsz = true
	return &h, nil
}

// RemoveRandom can be used for a random policy eviction.
// This is stochastic but very fast and does not impede
// performance like LRU, LFU or even ARC based implementations.
func (h *HashCache) RemoveRandom() {
	if h.used == 0 {
		return
	}
	index := (rand.Int()) & int(h.msk)
	// Walk forward til we find an entry
	for i := index; i < len(h.bkts); i++ {
		e := &h.bkts[i]
		if *e != nil {
			*e = (*e).next
			h.used -= 1
			return
		}
	}
	// If we are here we hit end and did not remove anything,
	// use the index and walk backwards.
	for i := index; i >= 0; i-- {
		e := &h.bkts[i]
		if *e != nil {
			*e = (*e).next
			h.used -= 1
			return
		}
	}
	panic("Should not reach here..")
}
