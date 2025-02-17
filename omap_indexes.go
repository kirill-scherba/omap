// Copyright 2025 Kirill Scherba <kirill@scherba.ru>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Indexes of ordered map definition.

package omap

import (
	"container/list"
	"fmt"
	"sync"
)

// Indexes provides methods to handle index lists.
type Indexes[K comparable, D any] Omap[K, D]

// First gets first record from ordered map or nil if map is empty or incorrect
// index is passed.
func (in *Indexes[K, D]) First(idxKeys ...any) *Record[K, D] {
	in.RLock()
	defer in.RUnlock()

	// Get index list by key
	list, ok := in.getList(idxKeys...)
	if !ok {
		return nil
	}

	return in.elementToRecord(list.Front())
}

// Next gets next record from ordered map or nil if there is last record or input
// record is nil.
func (in *Indexes[K, D]) Next(rec *Record[K, D]) *Record[K, D] {
	in.RLock()
	defer in.RUnlock()

	// Return nil if input record is nil
	if rec == nil {
		return nil
	}

	return in.elementToRecord(rec.element().Next())
}

// Prev gets previous record from ordered map or nil if this record is first.
func (in *Indexes[K, D]) Prev(rec *Record[K, D]) *Record[K, D] {
	in.RLock()
	defer in.RUnlock()

	return in.elementToRecord(rec.element().Prev())
}

// Last gets last record from ordered map or nil if the list is empty.
func (in *Indexes[K, D]) Last(idxKeys ...any) *Record[K, D] {
	in.RLock()
	defer in.RUnlock()

	// Get index list by key
	list, ok := in.getList(idxKeys...)
	if !ok {
		return nil
	}

	return in.elementToRecord(list.Back())
}

// InsertBefore inserts record before element. Returns ErrKeyAllreadySet if key
// already exists.
func (in *Indexes[K, D]) InsertBefore(key K, data D, mark *Record[K, D]) (
	err error) {

	in.Lock()
	defer in.Unlock()

	// Check if key already exists
	if _, ok := in.m[key]; ok {
		err = ErrKeyAllreadySet
		return
	}

	// Add new record before selected
	in.m[key] = in.insert(key, data, before, mark)

	return
}

// InsertAfter inserts record after element. Returns ErrKeyAllreadySet if key
// already exists.
func (in *Indexes[K, D]) InsertAfter(key K, data D, mark *Record[K, D]) (
	err error) {

	in.Lock()
	defer in.Unlock()

	// Check if key already exists
	if _, ok := in.m[key]; ok {
		err = ErrKeyAllreadySet
		return
	}

	// Add new record before selected
	in.m[key] = in.insert(key, data, after, mark)

	return
}

// MoveToBack moves record to the back of ordered map. It returns ErrRecordNotFound
// if input record is nil.
func (in *Indexes[K, D]) MoveToBack(rec *Record[K, D]) (err error) {
	in.Lock()
	defer in.Unlock()

	// Return error if input record is nil
	if rec == nil {
		err = ErrRecordNotFound
		return
	}

	// Move record
	in.lm[0].MoveToBack(rec.element())

	return
}

// MoveToFront moves record to the front of ordered map. It returns ErrRecordNotFound
// if input record is nil.
func (in *Indexes[K, D]) MoveToFront(rec *Record[K, D]) (err error) {
	in.Lock()
	defer in.Unlock()

	// Return error if input record is nil
	if rec == nil {
		err = ErrRecordNotFound
		return
	}

	// Move record
	in.lm[0].MoveToFront(rec.element())
	return
}

// MoveBefore moves record rec to the new position before mark record. It returns
// ErrRecordNotFound if input record or mark record is nil.
func (in *Indexes[K, D]) MoveBefore(rec, mark *Record[K, D]) (err error) {
	in.Lock()
	defer in.Unlock()

	// Return error if input record or mark record is nil
	if rec == nil || mark == nil {
		err = ErrRecordNotFound
		return
	}

	// Move record
	in.lm[0].MoveBefore(rec.element(), mark.element())

	return
}

// MoveAfter moves record rec to the new position after mark record. It returns
// ErrRecordNotFound if input record or mark record is nil.
func (in *Indexes[K, D]) MoveAfter(rec, mark *Record[K, D]) (err error) {
	in.Lock()
	defer in.Unlock()

	// Return error if input record or mark record is nil
	if rec == nil || mark == nil {
		err = ErrRecordNotFound
		return
	}

	// Move record
	in.lm[0].MoveAfter(rec.element(), mark.element())

	return
}

// sortFunc sorts records in list by index key using sort function.
// Unsafe (does not lock).
func (in *Indexes[K, D]) sortFunc(idxKey any, f func(rec, next *Record[K, D]) int) {

	// Skip if f function not set
	if f == nil {
		return
	}

	// Get index list by key
	l, ok := in.getList(idxKey)
	if !ok {
		return
	}

	// Sort records in list
	var next *list.Element
	for el := l.Front(); el != nil; el = next {
		next = el.Next()
		in.sortRecord(idxKey, el, f)
	}
}

// sortRecord sorts record using sort function.
func (in *Indexes[K, D]) sortRecord(idxKey any, el *list.Element, f func(rec,
	next *Record[K, D]) int) (move bool) {

	// Get index list by key
	list, ok := in.getList(idxKey)
	if !ok {
		return
	}

	// Compare el record with next records using function f and move el if
	// necessary
	for next := el.Next(); next != nil; next = next.Next() {
		rec1, rec2 := in.elementToRecord(el), in.elementToRecord(next)

		// Print in print mode
		if printMode {
			fmt.Printf("Compare idx: %v, %v => %v\n",
				idxKey, rec1.Key(), rec2.Key())
		}

		// Compare elToMove record with next record
		if f(rec1, rec2) > 0 {
			move = true
			continue
		}

		// If move is set, than move elToMove record
		if move {
			list.MoveBefore(el, next)
			in.printMove(idxKey, true, el, next)
			return
		}
	}

	// If move is set, than move elToMove to the back
	if move {
		list.MoveToBack(el)
		in.printMove(idxKey, true, el, el.Prev())
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

// insert adds new record to ordered map.
//
//	direction:
//	0 - back,
//	1 - front,
//	2 - insert before
//	3 - insert after
func (in *Indexes[K, D]) insert(key K, data D, direction int,
	mark *Record[K, D]) (rec *Record[K, D]) {

	// Create new record and it to basic(insertion) list
	v := &recordValue[K, D]{Key: key, Data: data}

	// Add element to basic(insertion) list
	switch direction {
	case 0:
		rec = in.elementToRecord(in.lm[0].PushBack(v))
	case 1:
		rec = in.elementToRecord(in.lm[0].PushFront(v))
	case 2:
		rec = in.elementToRecord(in.lm[0].InsertBefore(v, mark.element()))
	case 3:
		rec = in.elementToRecord(in.lm[0].InsertAfter(v, mark.element()))
	}

	// Add element to back of additional index lists and sort this lists
	var wg sync.WaitGroup
	for k := range in.lm {
		// Skip basic insertion list
		if k == 0 {
			continue
		}

		// Add element to the top of list
		in.lm[k].PushFront(v)

		// Sort list
		wg.Add(1)
		go func() {
			in.sortFunc(k, in.sm[k])
			wg.Done()
		}()
	}
	wg.Wait()

	return
}

// sort sorts all additional index lists.
func (in *Indexes[K, D]) sort() {
	var wg sync.WaitGroup
	for k := range in.sm {
		// Skip basic insertion list
		if k == 0 {
			continue
		}

		// Sort list
		wg.Add(1)
		go func() {
			in.sortFunc(k, in.sm[k])
			wg.Done()
		}()
	}
	wg.Wait()
}

// getList gets list from ordered map by index key. If index key is not set,
// the function will return default list.
func (in *Indexes[K, D]) getList(idxKeys ...any) (list *list.List, ok bool) {
	var idxKey any = 0

	// Use first index key if it is set
	if len(idxKeys) > 0 {
		idxKey = idxKeys[0]
	}

	// Get list by index key
	list, ok = in.lm[idxKey]

	return
}

// checkPair checks if record pair is already sorted.
//
// The function takes pair of records keys prepared to compare and checks if
// this pair is already sorted. If pair is sorted, the function returns true,
// otherwise it returns false.
func (in *Indexes[K, D]) checkPair(sorted map[any]any, k1, k2 any) bool {
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
func (in *Indexes[K, D]) printMove(idxKey any, before bool, el, next *list.Element) {

	if !printMode {
		return
	}

	keyFirst := in.elementToRecord(el).Key()
	keyNext := in.elementToRecord(next).Key()

	if before {
		fmt.Printf("Move    idx: %v, MoveBefore: %v => %v\n",
			idxKey, keyFirst, keyNext)
	} else {
		fmt.Printf("Move    idx: %v, MoveAfter: %v => %v\n",
			idxKey, keyFirst, keyNext)
	}
}
