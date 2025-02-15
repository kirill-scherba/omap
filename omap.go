// Copyright 2025 Kirill Scherba <kirill@scherba.ru>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Omap is Golang package for working with thread safe ordered maps. The ordered
// map contains the golang map, list and mutex to execute Ordered Map functions.
//
// The Ordered Map is a map that remembers the order of items. The map can be
// iterated over to retrieve the items in the order they were added.
package omap

import (
	"container/list"
	"errors"
	"iter"
	"sync"

	"golang.org/x/exp/constraints"
)

var (
	ErrMapIsEmpty        = errors.New("map is empty")
	ErrRecordNotFound    = errors.New("record not found")
	ErrKeyAllreadySet    = errors.New("key already exists")
	ErrIncorrectIndexKey = errors.New("incorrect index key name")
)

// Print mode is variable to enable print debug messages.
var printMode = false

// Omap is a concurrent safe multi index ordered map.
type Omap[K comparable, D any] struct {

	// Map by key
	m dataMap[K, D]

	// Map of index list elements in order of insert and any custom.
	// There is one default sort list key:
	//   0 - by insertion
	lm listMap

	// Sort functions map
	sm indexMap[K, D]

	// Indexes module
	Idx *Indexes[K, D]

	// Mutex to protect ordered map operations
	*sync.RWMutex
}
type indexMap[K comparable, D any] map[any]SortIndexFunc[K, D]
type dataMap[K comparable, D any] map[K]*Record[K, D]
type listMap map[any]*list.List

// Index is a sort index definition struct.
type Index[K comparable, D any] struct {
	Key  any
	Func SortIndexFunc[K, D]
}
type SortIndexFunc[K comparable, D any] func(rec, next *Record[K, D]) int

// Pair represents a key-value pair in the ordered map.
type Pair[K comparable, D any] struct {
	Key   K
	Value D
}

// New creates a new ordered map object with key of type T and data of type D.
func New[K comparable, D any](sorts ...Index[K, D]) (m *Omap[K, D], err error) {

	// Create new ordered map object and make maps
	m = new(Omap[K, D])
	m.m = make(dataMap[K, D])
	m.lm = make(listMap)
	m.sm = make(indexMap[K, D])

	m.Idx = (*Indexes[K, D])(m)

	// Create mutex to protect ordered map
	m.RWMutex = new(sync.RWMutex)

	// Add default sort index
	m.lm[0] = list.New()
	m.sm[0] = nil

	// Add sort indexes
	for i := range sorts {
		// Skip default sort index TODO: return error
		if sorts[i].Key == 0 {
			err = ErrIncorrectIndexKey
			return
		}
		// Add sort index function and create new list
		m.sm[sorts[i].Key] = sorts[i].Func
		m.lm[sorts[i].Key] = list.New()
	}

	return
}

// CompareByKey compares two records by their keys.
//
// This function returns a negative value if rec1 key is less than rec2 key,
// zero if the keys are equal, and a positive value if rec1 key is greater
// than rec2 key.
func CompareByKey[K constraints.Ordered, D any](r1, r2 *Record[K, D]) int {
	key1 := r1.Key()
	key2 := r2.Key()

	switch {
	case key1 > key2:
		return 1

	case key1 < key2:
		return -1

	default:
		return 0
	}
}

// Clear removes all records from ordered map.
func (m *Omap[K, D]) Clear() {
	m.Lock()
	defer m.Unlock()

	// Make data map and init index lists
	m.m = make(dataMap[K, D])
	for k := range m.lm {
		m.lm[k].Init()
	}
}

// Len returns the number of elements in the map.
func (m *Omap[K, D]) Len() int {
	m.RLock()
	defer m.RUnlock()
	return len(m.m)
}

// Get gets record from ordered map by key. Returns ok true if found.
func (m *Omap[K, D]) Get(key K) (data D, ok bool) {
	m.RLock()
	defer m.RUnlock()

	// Get list element
	el, ok := m.m[key]
	if !ok {
		return
	}

	// Get records data
	v, _ := el.Value.(recordValue[K, D])
	data = v.Data

	return
}

// GetRecord gets record from ordered map by key. Returns ok true if found.
func (m *Omap[K, D]) GetRecord(key K) (rec *Record[K, D], ok bool) {
	m.RLock()
	defer m.RUnlock()

	// Get record
	rec, ok = m.m[key]
	return
}

// Set adds or updates record in ordered map by key. It adds new record to the
// back of ordered map. If key already exists, it will be updated.
func (m *Omap[K, D]) Set(key K, data D) (err error) {
	m.Lock()
	defer m.Unlock()

	// Check if key already exists and update data and sort lists if exists
	if el, ok := m.m[key]; ok {
		el.Value = data
		m.Idx.sort()
		return
	}

	// Add new record to back of lists and map
	m.m[key] = m.Idx.insert(key, data, back, nil)

	return
}

// SetFirst adds or updates record in ordered map by key. It adds new record to
// the front of ordered map. If key already exists, it will be updated.
func (m *Omap[K, D]) SetFirst(key K, data D) (err error) {
	m.Lock()
	defer m.Unlock()

	// Check if key already exists and update data and sort lists if exists
	if el, ok := m.m[key]; ok {
		el.Value = data
		m.Idx.sort()
		return
	}

	// Add new record to front of lists and map
	m.m[key] = m.Idx.insert(key, data, front, nil)

	return
}

// Del removes record from ordered map by key. Returns ok true and deleted data
// if key exists, and record was successfully removed.
func (m *Omap[K, D]) Del(key K) (data D, ok bool) {
	m.Lock()
	defer m.Unlock()

	// Check if key exists and get data if exists
	rec, ok := m.m[key]
	if !ok {
		return
	}
	data = rec.Data()

	// Remove element from lists
	for k := range m.lm {
		m.lm[k].Remove(rec.element())
	}

	// Remove key from map
	delete(m.m, key)

	return
}

// ForEach calls function f for each key and value present in the map.
//
// By default, it iterates over default (insertion) index. Use idxKey to iterate
// over other indexes.
//
// Function f is called for each key and value present in the map. The order of
// iteration is determined by the index. If the index is not specified, the
// default (insertion) index is used.
func (m *Omap[K, D]) ForEach(f func(key K, data D), idxKey ...any) {
	m.RLock()
	defer m.RUnlock()

	for rec := m.Idx.First(idxKey...); rec != nil; rec = m.Idx.Next(rec) {
		f(rec.Key(), rec.Data())
	}
}

// ForEachRecord calls function f for each record present in the map.
//
// By default, it iterates over default (insertion) index. Use idxKey to iterate
// over other indexes.
//
// It allows to handle records directly, which could be useful for example to
// call methods on the record, or to get the index of the record in the list.
func (m *Omap[K, D]) ForEachRecord(f func(rec *Record[K, D]), idxKey ...any) {
	m.RLock()
	defer m.RUnlock()

	for rec := m.Idx.First(idxKey...); rec != nil; rec = m.Idx.Next(rec) {
		f(rec)
	}
}

// ForEachPair calls function f for each key-value pair present in the map.
//
// By default, it iterates over default (insertion) index. Use idxKey to iterate
// over other indexes.
//
// It allows to handle key-value pairs directly, which could be useful for
// example to call methods on the pair, or to get the index of the pair in the
// list.
func (m *Omap[K, D]) ForEachPair(f func(pair Pair[K, D]), idxKey ...any) {
	m.RLock()
	defer m.RUnlock()

	for rec := m.Idx.First(idxKey...); rec != nil; rec = m.Idx.Next(rec) {
		f(Pair[K, D]{Key: rec.Key(), Value: rec.Data()})
	}
}

// Pairs returns a slice of key-value pairs in the omap. By default, it iterates
// over default (insertion) index. Use idxKey to iterate over other indexes.
func (m *Omap[K, D]) Pairs(idxKey ...any) []Pair[K, D] {
	m.RLock()
	defer m.RUnlock()

	pairs := make([]Pair[K, D], 0, len(m.m))
	for rec := m.Idx.First(idxKey...); rec != nil; rec = m.Idx.Next(rec) {
		pairs = append(pairs, Pair[K, D]{Key: rec.Key(), Value: rec.Data()})
	}

	return pairs
}

// Records returns an iterator over the omap records. By default, it iterates
// over default (insertion) index. Use idxKey to iterate over other indexes.
//
// The iteration stops when the function passed to the iterator returns false.
//
// This function is safe for concurrent read access. But all write Omap methods
// will lock during iteration.
func (m *Omap[K, D]) Records(idxKey ...any) iter.Seq2[K, D] {
	return func(yield func(K, D) bool) {
		m.RLock()
		defer m.RUnlock()

		for rec := m.Idx.First(idxKey...); rec != nil; rec = m.Idx.Next(rec) {
			if !yield(rec.Key(), rec.Data()) {
				return
			}
		}
	}
}

// Refresh refreshes the index lists.
//
// The indexes automatically sorts when a new record is added or updated with
// the Set or SetFirst methods.
//
// If you directly update the map data (D type) use this method to refresh the
// index lists.
func (m *Omap[K, D]) Refresh() {
	m.Lock()
	defer m.Unlock()

	m.Idx.sort()
}
