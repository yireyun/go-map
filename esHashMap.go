// Copyright 2012-2015 Apcera Inc. All rights reserved.

// HashMap defines a high performance hashmap based on
// fast hashing and fast key comparison. Simple chaining
// is used, relying on the hashing algorithms for good
// distribution
package esMap

import (
	"bytes"
	"errors"
	"unsafe"

	"easy/esUtil/esHash"
)

// HashMap stores Entry items using a given Hash function.
// The Hash function can be overridden.
type HashMap struct {
	Hash func([]byte) uint32
	bkts []*Entry
	msk  uint32
	used uint32
	rsz  bool
}

// BucketSize, must be power of 2
const _BSZ = 8

// Constants for multiples of sizeof(WORD)
const (
	_DSZ  = 2         // 2
	_WSZ  = 4         // 4
	_DWSZ = _WSZ << 1 // 8
)

// DefaultHash to be used unless overridden.
var DefaultHash = esHash.Wukehong

// Stats are reported on HashMaps
type Stats struct {
	NumElements uint32
	NumSlots    uint32
	NumBuckets  uint32
	LongChain   uint32
	AvgChain    float32
}

// NewWithBkts creates a new HashMap using the bkts slice argument.
// bkts must be a power of 2.
func NewHashMapWithBkts(bkts int) (*HashMap, error) {
	if bkts == 0 || (bkts&(bkts-1) != 0) {
		return nil, errors.New("Size of buckets must be power of 2")
	}
	h := HashMap{}
	h.msk = uint32(bkts - 1)
	h.bkts = make([]*Entry, bkts)
	h.Hash = DefaultHash
	h.rsz = true
	return &h, nil
}

// New creates a new HashMap of default size and using the default
// Hashing algorithm.
func NewHashMap() *HashMap {
	h, _ := NewHashMapWithBkts(_BSZ)
	return h
}

// Set will set the key item to data. This will blindly replace any item
// that may have been at key previous.
func (h *HashMap) Set(key []byte, data interface{}) {
	hk := h.Hash(key)
	e := h.bkts[hk&h.msk]
	for e != nil {
		if len(key) == len(e.key) && hk == e.hk && bytes.Equal(key, e.key) {
			// Success, replace data field
			e.data = data
			return
		}
		e = e.next
	}
	// We have a new entry here
	ne := &Entry{hk: hk, key: key, data: data}
	ne.next = h.bkts[hk&h.msk]
	h.bkts[hk&h.msk] = ne
	h.used += 1
	// Check for resizing
	if h.rsz && (h.used > uint32(len(h.bkts))) {
		h.grow()
	}
}

// Get will return the item at key.
func (h *HashMap) Get(key []byte) interface{} {
	hk := h.Hash(key)
	e := h.bkts[hk&h.msk]
	// FIXME: Reorder on GET if chained?
	// We unroll and optimize the comparison of keys.
	for e != nil {
		klen := len(key)
		p1 := uintptr(unsafe.Pointer(&key[0]))
		p2 := uintptr(unsafe.Pointer(&e.key[0]))
		if klen != len(e.key) || hk != e.hk {
			goto next
		}
		if p1 != p2 {
			// We unroll and optimize the key comparison here.
			// Compare _DWSZ at a time
			for ; klen >= _DWSZ; klen -= _DWSZ {
				k1 := *(*uint64)(unsafe.Pointer(p1))
				k2 := *(*uint64)(unsafe.Pointer(p2))
				if k1 != k2 {
					goto next
				}
				p1 += _DWSZ
				p2 += _DWSZ
			}
			// Check by _WSZ if applicable
			if (klen & _WSZ) > 0 {
				k1 := *(*uint32)(unsafe.Pointer(p1))
				k2 := *(*uint32)(unsafe.Pointer(p2))
				if k1 != k2 {
					goto next
				}
				p1 += _WSZ
				p2 += _WSZ
			}
			// Check by _DSZ if applicable
			if (klen & _DSZ) > 0 {
				k1 := *(*uint16)(unsafe.Pointer(p1))
				k2 := *(*uint16)(unsafe.Pointer(p2))
				if k1 != k2 {
					goto next
				}
				p1 += _DSZ
				p2 += _DSZ
			}
			// Check by byte if applicable
			if (klen & 1) > 0 {
				k1 := *(*uint8)(unsafe.Pointer(p1))
				k2 := *(*uint8)(unsafe.Pointer(p2))
				if k1 != k2 {
					goto next
				}
			}
		}
		// Success
		return e.data
	next:
		e = e.next
	}
	return nil
}

// Remove will remove what is associated with key.
func (h *HashMap) Remove(key []byte) {
	hk := h.Hash(key)
	e := &h.bkts[hk&h.msk]
	for *e != nil {
		if len(key) == len((*e).key) && hk == (*e).hk && bytes.Equal(key, (*e).key) {
			// Success
			*e = (*e).next
			h.used -= 1
			// Check for resizing
			lbkts := uint32(len(h.bkts))
			if h.rsz && lbkts > _BSZ && (h.used < lbkts>>2) {
				h.shrink()
			}
			return
		}
		e = &(*e).next
	}
}

// resize is responsible for reallocating the buckets and
// redistributing the hashmap entries.
func (h *HashMap) resize(nsz uint32) {
	nmsk := nsz - 1
	bkts := make([]*Entry, nsz)
	ents := make([]Entry, h.used)
	var ne *Entry
	var i int
	for _, e := range h.bkts {
		for ; e != nil; e = e.next {
			ne, i = &ents[i], i+1
			*ne = *e
			ne.next = bkts[e.hk&nmsk]
			bkts[e.hk&nmsk] = ne
		}
	}
	h.bkts = bkts
	h.msk = nmsk
}

const maxBktSize = (1 << 31) - 1

// grow the HashMap's buckets by 2
func (h *HashMap) grow() {
	// Can't grow beyond maxint for now
	if len(h.bkts) >= maxBktSize {
		return
	}
	h.resize(uint32(len(h.bkts) << 1))
}

// shrink the HashMap's buckets by 2
func (h *HashMap) shrink() {
	if len(h.bkts) <= _BSZ {
		return
	}
	h.resize(uint32(len(h.bkts) >> 1))
}

// Count returns number of elements in the HashMap
func (h *HashMap) Count() uint32 {
	return h.used
}

// AllKeys will return all the keys stored in the HashMap
func (h *HashMap) AllKeys() [][]byte {
	all := make([][]byte, 0, h.used)
	for _, e := range h.bkts {
		for ; e != nil; e = e.next {
			all = append(all, e.key)
		}
	}
	return all
}

// All returns all the Entries in the map
func (h *HashMap) All() []interface{} {
	all := make([]interface{}, 0, h.used)
	for _, e := range h.bkts {
		for ; e != nil; e = e.next {
			all = append(all, e.data)
		}
	}
	return all
}

// Stats will collect general statistics about the HashMap
func (h *HashMap) Stats() *Stats {
	lc, totalc, slots := 0, 0, 0
	for _, e := range h.bkts {
		if e != nil {
			slots += 1
		}
		i := 0
		for ; e != nil; e = e.next {
			i += 1
			if i > lc {
				lc = i
			}
		}
		totalc += i
	}
	l := uint32(len(h.bkts))
	avg := (float32(totalc) / float32(slots))
	return &Stats{
		NumElements: h.used,
		NumBuckets:  l,
		LongChain:   uint32(lc),
		AvgChain:    avg,
		NumSlots:    uint32(slots)}
}

func SilceEqui(b1, b2 []byte) bool {
	i, klen := 0, len(b1)
	if klen != len(b2) {
		return false
	}
	// We unroll and optimize the key comparison here.
	// Compare _DWSZ at a time
	for ; klen >= _DWSZ; klen -= _DWSZ {
		k1 := *(*uint64)(unsafe.Pointer(&b1[i]))
		k2 := *(*uint64)(unsafe.Pointer(&b2[i]))
		if k1 != k2 {
			return false
		}
		i += _DWSZ
	}
	// Check by _WSZ if applicable
	if (klen & _WSZ) > 0 {
		k1 := *(*uint32)(unsafe.Pointer(&b1[i]))
		k2 := *(*uint32)(unsafe.Pointer(&b2[i]))
		if k1 != k2 {
			return false
		}
		i += _WSZ
	}
	// Compare what is left over, byte by byte
	for ; i < len(b1); i++ {
		if b1[i] != b2[i] {
			return false
		}
	}
	return true
}
