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
	"fmt"
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
func New[K comparable, D any](sorts ...Index[K, D]) (o *Omap[K, D], err error) {

	// Create new ordered map object and make maps
	o = new(Omap[K, D])
	o.m = make(dataMap[K, D])
	o.lm = make(listMap)
	o.sm = make(indexMap[K, D])

	// Create mutex to protect ordered map
	o.RWMutex = new(sync.RWMutex)

	// Add default sort index
	o.lm[0] = list.New()
	o.sm[0] = nil

	// Add sort indexes
	for i := range sorts {
		// Skip default sort index TODO: return error
		if sorts[i].Key == 0 {
			err = ErrIncorrectIndexKey
			return
		}
		// Add sort index function and create new list
		o.sm[sorts[i].Key] = sorts[i].Func
		o.lm[sorts[i].Key] = list.New()
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
func (o *Omap[K, D]) Clear() {
	o.Lock()
	defer o.Unlock()

	// Make data map and init index lists
	o.m = make(dataMap[K, D])
	for k := range o.lm {
		o.lm[k].Init()
	}
}

// Len returns the number of elements in the map.
func (o *Omap[K, D]) Len() int {
	o.RLock()
	defer o.RUnlock()
	return len(o.m)
}

// Get gets record from ordered map by key. Returns ok true if found.
func (o *Omap[K, D]) Get(key K) (data D, ok bool) {
	o.RLock()
	defer o.RUnlock()

	// Get list element
	el, ok := o.m[key]
	if !ok {
		return
	}

	// Get records data
	v, _ := el.Value.(recordValue[K, D])
	data = v.Data

	return
}

// GetRecord gets record from ordered map by key. Returns ok true if found.
func (o *Omap[K, D]) GetRecord(key K) (rec *Record[K, D], ok bool) {
	o.RLock()
	defer o.RUnlock()

	// Get record
	rec, ok = o.m[key]
	return
}

// Set adds or updates record in ordered map by key. It adds new record to the
// back of ordered map. If key already exists, it will be updated.
func (o *Omap[K, D]) Set(key K, data D) (err error) {
	o.Lock()
	defer o.Unlock()

	// Check if key already exists and update data and sort lists if exists
	if el, ok := o.m[key]; ok {
		el.Value = data
		o.sortLists()
		return
	}

	// Add new record to back of lists and map
	o.m[key] = o.insertRecord(key, data, back, nil)

	return
}

// SetFirst adds or updates record in ordered map by key. It adds new record to
// the front of ordered map. If key already exists, it will be updated.
func (o *Omap[K, D]) SetFirst(key K, data D) (err error) {
	o.Lock()
	defer o.Unlock()

	// Check if key already exists and update data and sort lists if exists
	if el, ok := o.m[key]; ok {
		el.Value = data
		o.sortLists()
		return
	}

	// Add new record to front of lists and map
	o.m[key] = o.insertRecord(key, data, front, nil)

	return
}

// Del removes record from ordered map by key. Returns ok true and deleted data
// if key exists, and record was successfully removed.
func (o *Omap[K, D]) Del(key K) (data D, ok bool) {
	o.Lock()
	defer o.Unlock()

	// Check if key exists and get data if exists
	rec, ok := o.m[key]
	if !ok {
		return
	}
	data = rec.Data()

	// Remove element from lists
	for k := range o.lm {
		o.lm[k].Remove(rec.element())
	}

	// Remove key from map
	delete(o.m, key)

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
func (o *Omap[K, D]) ForEach(f func(key K, data D), idxKey ...any) {
	o.RLock()
	defer o.RUnlock()

	for rec := o.First(idxKey...); rec != nil; rec = o.Next(rec) {
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
func (o *Omap[K, D]) ForEachRecord(f func(rec *Record[K, D]), idxKey ...any) {
	o.RLock()
	defer o.RUnlock()

	for rec := o.First(idxKey...); rec != nil; rec = o.Next(rec) {
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
func (o *Omap[K, D]) ForEachPair(f func(pair Pair[K, D]), idxKey ...any) {
	o.RLock()
	defer o.RUnlock()

	for rec := o.First(idxKey...); rec != nil; rec = o.Next(rec) {
		f(Pair[K, D]{Key: rec.Key(), Value: rec.Data()})
	}
}

// Pairs returns a slice of key-value pairs in the omap. By default, it iterates
// over default (insertion) index. Use idxKey to iterate over other indexes.
func (o *Omap[K, D]) Pairs(idxKey ...any) []Pair[K, D] {
	o.RLock()
	defer o.RUnlock()

	pairs := make([]Pair[K, D], 0, len(o.m))
	for rec := o.First(idxKey...); rec != nil; rec = o.Next(rec) {
		pairs = append(pairs, Pair[K, D]{Key: rec.Key(), Value: rec.Data()})
	}

	return pairs
}

// First gets first record from ordered map or nil if map is empty or incorrect
// index is passed.
func (o *Omap[K, D]) First(idxKeys ...any) *Record[K, D] {
	o.RLock()
	defer o.RUnlock()

	// Get index list by key
	list, ok := o.getList(idxKeys...)
	if !ok {
		return nil
	}

	return o.elementToRecord(list.Front())
}

// Next gets next record from ordered map or nil if there is last record or input
// record is nil.
func (o *Omap[K, D]) Next(rec *Record[K, D]) *Record[K, D] {
	o.RLock()
	defer o.RUnlock()

	// Return nil if input record is nil
	if rec == nil {
		return nil
	}

	return o.elementToRecord(rec.element().Next())
}

// Prev gets previous record from ordered map or nil if this record is first.
func (o *Omap[K, D]) Prev(rec *Record[K, D]) *Record[K, D] {
	o.RLock()
	defer o.RUnlock()

	return o.elementToRecord(rec.element().Prev())
}

// Last gets last record from ordered map or nil if the list is empty.
func (o *Omap[K, D]) Last(idxKeys ...any) *Record[K, D] {
	o.RLock()
	defer o.RUnlock()

	// Get index list by key
	list, ok := o.getList(idxKeys...)
	if !ok {
		return nil
	}

	return o.elementToRecord(list.Back())
}

// InsertBefore inserts record before element. Returns ErrKeyAllreadySet if key
// already exists.
func (o *Omap[K, D]) InsertBefore(key K, data D, mark *Record[K, D]) (
	err error) {

	o.Lock()
	defer o.Unlock()

	// Check if key already exists
	if _, ok := o.m[key]; ok {
		err = ErrKeyAllreadySet
		return
	}

	// Add new record before selected
	o.m[key] = o.insertRecord(key, data, before, mark)

	return
}

// InsertAfter inserts record after element. Returns ErrKeyAllreadySet if key
// already exists.
func (o *Omap[K, D]) InsertAfter(key K, data D, mark *Record[K, D]) (
	err error) {

	o.Lock()
	defer o.Unlock()

	// Check if key already exists
	if _, ok := o.m[key]; ok {
		err = ErrKeyAllreadySet
		return
	}

	// Add new record before selected
	o.m[key] = o.insertRecord(key, data, after, mark)

	return
}

// MoveToBack moves record to the back of ordered map. It returns ErrRecordNotFound
// if input record is nil.
func (o *Omap[K, D]) MoveToBack(rec *Record[K, D]) (err error) {
	o.Lock()
	defer o.Unlock()

	// Return error if input record is nil
	if rec == nil {
		err = ErrRecordNotFound
		return
	}

	// Move record
	o.lm[0].MoveToBack(rec.element())

	return
}

// MoveToFront moves record to the front of ordered map. It returns ErrRecordNotFound
// if input record is nil.
func (o *Omap[K, D]) MoveToFront(rec *Record[K, D]) (err error) {
	o.Lock()
	defer o.Unlock()

	// Return error if input record is nil
	if rec == nil {
		err = ErrRecordNotFound
		return
	}

	// Move record
	o.lm[0].MoveToFront(rec.element())
	return
}

// MoveBefore moves record rec to the new position before mark record. It returns
// ErrRecordNotFound if input record or mark record is nil.
func (o *Omap[K, D]) MoveBefore(rec, mark *Record[K, D]) (err error) {
	o.Lock()
	defer o.Unlock()

	// Return error if input record or mark record is nil
	if rec == nil || mark == nil {
		err = ErrRecordNotFound
		return
	}

	// Move record
	o.lm[0].MoveBefore(rec.element(), mark.element())

	return
}

// MoveAfter moves record rec to the new position after mark record. It returns
// ErrRecordNotFound if input record or mark record is nil.
func (o *Omap[K, D]) MoveAfter(rec, mark *Record[K, D]) (err error) {
	o.Lock()
	defer o.Unlock()

	// Return error if input record or mark record is nil
	if rec == nil || mark == nil {
		err = ErrRecordNotFound
		return
	}

	// Move record
	o.lm[0].MoveAfter(rec.element(), mark.element())

	return
}

// sortFunc sorts records using sort function. Unsafe (does not lock).
func (o *Omap[K, D]) sortFunc(idxKey any, f func(rec, next *Record[K, D]) int) {

	// Skip if f function not set
	if f == nil {
		return
	}

	// Get index list by key
	l, ok := o.getList(idxKey)
	if !ok {
		return
	}

	// Sort records
	var next *list.Element
	var sorted = make(map[any]any)
	for el := l.Front(); el != nil; el = next {
		next = el.Next()
		if o.sortRecord(idxKey, el, f, sorted) {
			next = l.Front()
		}
	}

}

// sortRecord sorts record using sort function.
func (o *Omap[K, D]) sortRecord(idxKey any, elToMove *list.Element, f func(rec,
	next *Record[K, D]) int, sorted map[any]any) (move bool) {

	// Get index list by key
	list, ok := o.getList(idxKey)
	if !ok {
		return
	}

	// Compare el record with next records using function f and move el if
	// necessary
	for el, elNext := elToMove, elToMove.Next(); ; el, elNext = elNext, elNext.Next() {

		// When the end of the list is reached
		if elNext == nil {
			// If move is set, than move elToMove record
			if move {
				list.MoveAfter(elToMove, el)
				o.printMove(idxKey, false, elToMove, el)
			}
			break
		}

		// Check if records pair already sorted
		k1 := o.elementToRecord(elToMove).Key()
		k2 := o.elementToRecord(elNext).Key()
		if o.checkPair(sorted, k1, k2) {
			if printMode {
				fmt.Printf("   skip %v => %v\n", k1, k2)
			}
			continue
		}

		// Print compare records in printMode
		if printMode {
			fmt.Printf("Compare idx: %v, %v => %v\n", idxKey, k1, k2)
		}

		// Compare elToMove record with elNext record
		if f(o.elementToRecord(elToMove), o.elementToRecord(elNext)) > 0 {
			move = true
			continue
		}

		// If move is set, than move elToMove record
		if move {
			list.MoveBefore(elToMove, elNext)
			o.printMove(idxKey, true, elToMove, elNext)
		}

		break
	}

	return
}

// Directions const values
const (
	back int = iota
	front
	before
	after
)

// insertRecord adds new record to ordered map.
//
//	direction:
//	0 - back,
//	1 - front,
//	2 - insertRecord before
//	3 - insertRecord after
func (o *Omap[K, D]) insertRecord(key K, data D, direction int,
	mark *Record[K, D]) (rec *Record[K, D]) {

	// Create new record and it to basic(insertion) list
	v := recordValue[K, D]{Key: key, Data: data}

	// Add element to basic(insertion) list
	switch direction {
	case 0:
		rec = o.elementToRecord(o.lm[0].PushBack(v))
	case 1:
		rec = o.elementToRecord(o.lm[0].PushFront(v))
	case 2:
		rec = o.elementToRecord(o.lm[0].InsertBefore(v, mark.element()))
	case 3:
		rec = o.elementToRecord(o.lm[0].InsertAfter(v, mark.element()))
	}

	// Add element to back of additional index lists and sort this lists
	var wg sync.WaitGroup
	for k := range o.lm {
		// Skip basic insertion list
		if k == 0 {
			continue
		}

		// Add element to end of list
		o.lm[k].PushFront(v)

		// Sort list
		wg.Add(1)
		go func() {
			o.sortFunc(k, o.sm[k])
			wg.Done()
		}()
	}
	wg.Wait()

	return
}

// sortLists sorts all additional lists.
func (o *Omap[K, D]) sortLists() {
	var wg sync.WaitGroup
	for k := range o.sm {
		// Skip basic insertion list
		if k == 0 {
			continue
		}

		// Sort list
		wg.Add(1)
		go func() {
			o.sortFunc(k, o.sm[k])
			wg.Done()
		}()
	}
	wg.Wait()
}

// getList gets list from ordered map by index key. If index key is not set,
// the function will return default list.
func (o *Omap[K, D]) getList(idxKeys ...any) (list *list.List, ok bool) {
	var idxKey any = 0

	// Use first index key if it is set
	if len(idxKeys) > 0 {
		idxKey = idxKeys[0]
	}

	// Get list by index key
	list, ok = o.lm[idxKey]

	return
}

// checkPair checks if record pair is already sorted.
//
// The function takes pair of records keys prepared to compare and checks if
// this pair is already sorted. If pair is sorted, the function returns true,
// otherwise it returns false.
func (o *Omap[K, D]) checkPair(sorted map[any]any, k1, k2 any) bool {
	// Create array of sorted pair
	sk := []struct{ k1, k2 any }{{k1, k2}, {k2, k1}}

	// Check if pair is already sorted
	for _, k := range sk {
		if _, ok := sorted[k]; ok {
			return true
		}
	}

	// Add pair to sorted map
	sorted[sk[0]] = nil

	return false
}

// Print move records. To enable print move set printMove variable to true.
func (o *Omap[K, D]) printMove(idxKey any, before bool, el, next *list.Element) {

	if !printMode {
		return
	}

	keyFirst := o.elementToRecord(el).Key()
	keyNext := o.elementToRecord(next).Key()

	if before {
		fmt.Printf("Move    idx: %v, MoveBefore: %v => %v\n",
			idxKey, keyFirst, keyNext)
	} else {
		fmt.Printf("Move    idx: %v, MoveAfter: %v => %v\n",
			idxKey, keyFirst, keyNext)
	}
}
