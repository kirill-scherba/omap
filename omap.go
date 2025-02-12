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
)

var (
	ErrMapIsEmpty     = errors.New("map is empty")
	ErrRecordNotFound = errors.New("record not found")
	ErrKeyAllreadySet = errors.New("key allready exists")
)

type Omap[K comparable, D any] struct {
	m             Map[K]    // Map by key
	l             list.List // List of map elements in order of insert and sort
	*sync.RWMutex           // Mutex to protect ordered map operations
}
type Map[K comparable] map[K]Record
type Record *list.Element

// recordValue is a struct that contains key and value of map and list element.
type recordValue[K comparable, D any] struct {
	Key  K
	Data D
}

// New creates a new ordered map object with key of type T.
func New[K comparable, D any]() *Omap[K, D] {
	o := new(Omap[K, D])

	o.RWMutex = new(sync.RWMutex)
	o.m = make(Map[K])
	o.l.Init()

	return o
}

// Len returns the number of elements in the map.
func (o *Omap[K, D]) Len() int {
	o.RLock()
	defer o.RUnlock()
	return len(o.m)
}

// Get gets record from ordered map by key. Returns ok true if found.
func (o *Omap[K, D]) Get(key K, lock ...bool) (data D, ok bool) {
	if len(lock) == 0 || lock[0] {
		o.RLock()
		defer o.RUnlock()
	}

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

// Set adds or updates record in ordered map by key. It adds new record to the
// back of ordered map. If key allready exists, it will be updated.
func (o *Omap[K, D]) Set(key K, data D) (err error) {
	o.Lock()
	defer o.Unlock()

	// Check if key allready exists and update data if exists
	el, ok := o.m[key]
	if ok {
		el.Value = data
		return
	}

	// Add new record
	o.m[key] = o.l.PushBack(recordValue[K, D]{Key: key, Data: data})
	return
}

// SetFirst adds or updates record in ordered map by key. It adds new record to
// the front of ordered map. If key allready exists, it will be updated.
func (o *Omap[K, D]) SetFirst(key K, data D) (err error) {
	o.Lock()
	defer o.Unlock()

	// Check if key allready exists and update data if exists
	el, ok := o.m[key]
	if ok {
		el.Value = data
		return
	}

	// Add new record
	o.m[key] = o.l.PushFront(recordValue[K, D]{Key: key, Data: data})
	return
}

// Del removes record from ordered map by key. Returns ok true and deleted data
// if key exists, and record was successfully removed.
func (o *Omap[K, D]) Del(key K) (data D, ok bool) {
	o.Lock()
	defer o.Unlock()

	// Check if key exists
	el, ok := o.m[key]
	if !ok {
		return
	}

	// Remove record
	data, ok = o.l.Remove(el).(D)
	delete(o.m, key)

	return
}

// Clear removes all records from ordered map.
func (o *Omap[K, D]) Clear() {
	o.Lock()
	defer o.Unlock()

	o.l.Init()
	o.m = make(Map[K])
}

// First gets first record from ordered map or nil if map is empty.
func (o *Omap[K, D]) First() Record {
	o.RLock()
	defer o.RUnlock()

	return o.l.Front()
}

// Next gets next record from ordered map or nil if there is last record or input
// record is nil.
func (o *Omap[K, D]) Next(rec Record) Record {
	o.RLock()
	defer o.RUnlock()

	// Return nil if input record is nil
	if rec == nil {
		return nil
	}

	return (*list.Element)(rec).Next()
}

// Prev gets previous record from ordered map or nil if this record is first.
func (o *Omap[K, D]) Prev(rec Record) Record {
	o.RLock()
	defer o.RUnlock()

	return (*list.Element)(rec).Prev()
}

// Last gets last record from ordered map.
func (o *Omap[K, D]) Last() Record {
	o.RLock()
	defer o.RUnlock()

	return o.l.Back()
}

// InsertBefore inserts record before element. Returns ErrKeyAllreadySet if key
// allready exists.
func (o *Omap[K, D]) InsertBefore(key K, data D, rec Record) (err error) {
	o.Lock()
	defer o.Unlock()

	// Check if key allready exists
	if _, ok := o.m[key]; ok {
		err = ErrKeyAllreadySet
		return
	}

	o.m[key] = o.l.InsertBefore(recordValue[K, D]{Key: key, Data: data}, rec)
	return
}

// InsertAfter inserts record after element. Returns ErrKeyAllreadySet if key
// allready exists.
func (o *Omap[K, D]) InsertAfter(key K, data D, rec Record) (err error) {
	o.Lock()
	defer o.Unlock()

	// Check if key allready exists
	if _, ok := o.m[key]; ok {
		err = ErrKeyAllreadySet
		return
	}

	o.m[key] = o.l.InsertAfter(recordValue[K, D]{Key: key, Data: data}, rec)
	return
}

// MoveToBack moves record to the back of ordered map. It returns ErrRecordNotFound
// if input record is nil.
func (o *Omap[K, D]) MoveToBack(rec Record) (err error) {
	o.Lock()
	defer o.Unlock()

	// Return error if input record is nil
	if rec == nil {
		err = ErrRecordNotFound
		return
	}

	// Move record
	o.l.MoveToBack(rec)
	return
}

// MoveToFront moves record to the front of ordered map. It returns ErrRecordNotFound
// if input record is nil.
func (o *Omap[K, D]) MoveToFront(rec Record) (err error) {
	o.Lock()
	defer o.Unlock()

	// Return error if input record is nil
	if rec == nil {
		err = ErrRecordNotFound
		return
	}

	// Move record
	o.l.MoveToFront(rec)
	return
}

// MoveBefore moves record rec to the new position before mark record. It returns
// ErrRecordNotFound if input record or mark record is nil.
func (o *Omap[K, D]) MoveBefore(rec, mark Record) (err error) {
	o.Lock()
	defer o.Unlock()

	// Return error if input record or mark record is nil
	if rec == nil || mark == nil {
		err = ErrRecordNotFound
		return
	}

	// Move record
	o.l.MoveBefore(rec, mark)
	return
}

// MoveAfter moves record rec to the new position after mark record. It returns
// ErrRecordNotFound if input record or mark record is nil.
func (o *Omap[K, D]) MoveAfter(rec, mark Record) (err error) {
	o.Lock()
	defer o.Unlock()

	// Return error if input record or mark record is nil
	if rec == nil || mark == nil {
		err = ErrRecordNotFound
		return
	}

	// Move record
	o.l.MoveAfter(rec, mark)
	return
}

// RecordValue gets record value from element. It returns ErrRecordNotFound if
// input record is nil.
func (o *Omap[K, D]) RecordValue(rec Record) (key K, data D, err error) {

	// Return error if input record is nil
	if rec == nil {
		err = ErrRecordNotFound
		return
	}

	// Get record value
	if v, ok := rec.Value.(recordValue[K, D]); ok {
		key = v.Key
		data = v.Data
	}
	return
}

// SortFunc sorts records using sort function.
func (o *Omap[K, D]) SortFunc(f func(rec Record, next Record) int) {
	o.Lock()
	defer o.Unlock()

	// Get first records
	var next *list.Element
	for el := o.l.Front(); el != nil; el = next {

		// Get next record
		next = el.Next()

		// Compare el record with next records using function f and move el if
		// neccessary
		o.sortRecord(el, f)
	}
}

// sortRecord sorts record using sort function.
func (o *Omap[K, D]) sortRecord(elToMove *list.Element, f func(rec Record,
	next Record) int) {

	// Compare el record with next records using function f and move el if
	// neccessary
	move := false

	for el, elNext := elToMove, elToMove.Next(); ; el, elNext = elNext, elNext.Next() {

		// When the end of the list is reached
		if elNext == nil {
			// If move is set, than move elToMove record
			if move {
				o.l.MoveAfter(elToMove, el)
				o.printMove(false, elToMove, el)
			}
			break
		}

		// Compare elToMove record with elNext record
		if f(elToMove, elNext) > 0 {
			move = true
			continue
		}

		// If move is set, than move elToMove record
		if move {
			o.l.MoveBefore(elToMove, elNext)
			o.printMove(true, elToMove, elNext)
		}

		break
	}
}

// Print move records
func (o *Omap[K, D]) printMove(before bool, elmove, next *list.Element) {
	key, _, _ := o.RecordValue(elmove)
	keyNext, _, _ := o.RecordValue(next)
	if before {
		fmt.Printf("MoveBefore: %v => %v\n", key, keyNext)
	} else {
		fmt.Printf("MoveAfter: %v => %v\n", key, keyNext)
	}
}
