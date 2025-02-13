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
	ErrMapIsEmpty     = errors.New("map is empty")
	ErrRecordNotFound = errors.New("record not found")
	ErrKeyAllreadySet = errors.New("key allready exists")
)

type Omap[K constraints.Ordered /* comparable */, D any] struct {

	// Map by key
	m Map[K]

	// List of map elements in order of insert and sort.
	// There is to default sort list:
	//   0 - by insertion
	//   1 - by key
	l []*list.List

	// Sort functions
	s []func(rec Record, next Record) int

	// Mutex to protect ordered map operations
	*sync.RWMutex
}
type Map[K comparable] map[K]Record
type Record *list.Element

// recordValue is a struct that contains key and value of map and list element.
type recordValue[K comparable, D any] struct {
	Key  K
	Data D
}

// New creates a new ordered map object with key of type T.
func New[K constraints.Ordered /* comparable */, D any](sorts ...func(rec Record, next Record) int) *Omap[K, D] {
	o := new(Omap[K, D])

	o.RWMutex = new(sync.RWMutex)
	o.m = make(Map[K])

	o.s = append(o.s, nil)
	o.l = append(o.l, list.New())

	for i := range sorts {
		o.s = append(o.s, sorts[i])
		o.l = append(o.l, list.New())
	}

	return o
}

// RecordValue gets record value from element. It returns ErrRecordNotFound if
// input record is nil.
func RecordValue[K constraints.Ordered /* comparable */, D any](rec Record) (key K, data D, err error) {

	// Return error if input record is nil
	if rec == nil {
		err = ErrRecordNotFound
		return
	}

	// Get record value
	if v, ok := rec.Value.(recordValue[K, D]); ok {
		key = v.Key
		data = v.Data
	} else {
		err = fmt.Errorf("not recordValue %T", rec.Value)
	}
	return
}

func CompareRecordByKey[K constraints.Ordered /* comparable */, D any](rec1, rec2 Record) int {
	// v1, _, _ := RecordValue[K, D](rec1)
	// v2, _, _ := RecordValue[K, D](rec2)

	// fmt.Println("!!!")

	r, _ := rec1.Value.(recordValue[K, D])
	v1, _ := r.Key, r.Data

	r, _ = rec2.Value.(recordValue[K, D])
	v2, _ := r.Key, r.Data

	switch {
	case v1 > v2:
		return 1

	case v1 < v2:
		return -1

	default:
		return 0
	}
}

// Clear removes all records from ordered map.
func (o *Omap[K, D]) Clear() {
	o.Lock()
	defer o.Unlock()

	o.m = make(Map[K])
	for i := range o.l {
		o.l[i].Init()
	}
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
	if el, ok := o.m[key]; ok {
		el.Value = data
		// TODO: sort additional lists
		return
	}

	// Add new record to back of lists and map
	o.m[key] = o.insertRecord(key, data, 0, nil)

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

	// Add new record to front of lists and map
	o.m[key] = o.insertRecord(key, data, 1, nil)

	return
}

// Del removes record from ordered map by key. Returns ok true and deleted data
// if key exists, and record was successfully removed.
func (o *Omap[K, D]) Del(key K) (data D, ok bool) {
	o.Lock()
	defer o.Unlock()

	// Check if key exists and get data if exists
	el, ok := o.m[key]
	if !ok {
		return
	}
	data = el.Value.(D)

	// Remove element from lists
	for i := range o.l {
		o.l[i].Remove(el)
	}

	// Remove key from map
	delete(o.m, key)

	return
}

// First gets first record from ordered map or nil if map is empty.
func (o *Omap[K, D]) First(idxs ...int) Record {
	o.RLock()
	defer o.RUnlock()

	var idx int
	if len(idxs) > 0 {
		idx = idxs[0]
	}

	if idx >= len(o.l) {
		return nil
	}

	return o.l[idx].Front()
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
func (o *Omap[K, D]) Last(idxs ...int) Record {
	o.RLock()
	defer o.RUnlock()

	var idx int
	if len(idxs) > 0 {
		idx = idxs[0]
	}

	if idx >= len(o.l) {
		return nil
	}

	return o.l[idx].Back()
}

// InsertBefore inserts record before element. Returns ErrKeyAllreadySet if key
// allready exists.
func (o *Omap[K, D]) InsertBefore(idx int, key K, data D, rec Record) (err error) {
	o.Lock()
	defer o.Unlock()

	// Check if key allready exists
	if _, ok := o.m[key]; ok {
		err = ErrKeyAllreadySet
		return
	}

	// Add new record before selected
	o.m[key] = o.insertRecord(key, data, 2, rec)

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

	// o.m[key] = o.l.InsertAfter(recordValue[K, D]{Key: key, Data: data}, rec)
	// Add new record before selected
	o.m[key] = o.insertRecord(key, data, 3, rec)

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
	o.l[0].MoveToBack(rec)

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
	o.l[0].MoveToFront(rec)
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
	o.l[0].MoveBefore(rec, mark)

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
	o.l[0].MoveAfter(rec, mark)

	return
}

// sortFunc sorts records using sort function.
func (o *Omap[K, D]) sortFunc(idx int, f func(rec Record, next Record) int) {
	// o.Lock()
	// defer o.Unlock()

	// Skip if f function not set
	if f == nil {
		return
	}

	var next *list.Element
	for el := o.l[idx].Front(); el != nil; el = next {
		next = el.Next()
		o.sortRecord(idx, el, f)
	}
}

// sortRecord sorts record using sort function.
func (o *Omap[K, D]) sortRecord(idx int, elToMove *list.Element, f func(rec Record,
	next Record) int) {

	// Compare el record with next records using function f and move el if
	// neccessary
	move := false

	for el, elNext := elToMove, elToMove.Next(); ; el, elNext = elNext, elNext.Next() {

		// When the end of the list is reached
		if elNext == nil {
			// If move is set, than move elToMove record
			if move {
				o.l[idx].MoveAfter(elToMove, el)
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
			o.l[idx].MoveBefore(elToMove, elNext)
			o.printMove(true, elToMove, elNext)
		}

		break
	}
}

// insertRecord adds new record to ordered map.
//
//	direction:
//	0 - back,
//	1 - front,
//	2 - insertRecord before
//	3 - insertRecord after
func (o *Omap[K, D]) insertRecord(key K, data D, direction int, mark Record) (el *list.Element) {
	// Create new record and it to basic(insertion) list
	rec := recordValue[K, D]{Key: key, Data: data}

	// Add element to basic(insertion) list
	switch direction {
	case 0:
		el = o.l[0].PushBack(rec)
	case 1:
		el = o.l[0].PushFront(rec)
	case 2:
		el = o.l[0].InsertBefore(rec, mark)
	case 3:
		el = o.l[0].InsertAfter(rec, mark)
	}

	// Add element to back of additional lists and sort this lists
	for i := range o.l {
		if i == 0 {
			continue
		}
		o.l[i].PushFront(rec)
		o.sortFunc(i, o.s[i])
	}

	return
}

// Print move records
func (o *Omap[K, D]) printMove(before bool, elmove, next *list.Element) {
	key, _, _ := RecordValue[int, int](elmove)
	keyNext, _, _ := RecordValue[int, int](next)
	if before {
		fmt.Printf("MoveBefore: %v => %v\n", key, keyNext)
	} else {
		fmt.Printf("MoveAfter: %v => %v\n", key, keyNext)
	}
}
