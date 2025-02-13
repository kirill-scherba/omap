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

var printMove = false

type Omap[K comparable, D any] struct {

	// Map by key
	m Map[K, D]

	// List of map elements in order of insert and sort.
	// There is to default sort list:
	//   0 - by insertion
	//   1 - by key
	l []*list.List

	// Sort functions
	s []func(rec, next *Record[K, D]) int

	// Mutex to protect ordered map operations
	*sync.RWMutex
}
type Map[K comparable, D any] map[K]*Record[K, D]

// New creates a new ordered map object with key of type T.
func New[K comparable, D any](sorts ...func(rec, next *Record[K, D]) int) *Omap[K, D] {
	o := new(Omap[K, D])

	o.RWMutex = new(sync.RWMutex)
	o.m = make(Map[K, D])

	o.s = append(o.s, nil)
	o.l = append(o.l, list.New())

	for i := range sorts {
		o.s = append(o.s, sorts[i])
		o.l = append(o.l, list.New())
	}

	return o
}

// CompareRecordByKey compares two records by their keys.
// The function returns a negative value if rec1 key is less than rec2 key,
// zero if the keys are equal, and a positive value if rec1 key is greater
// than rec2 key.
func CompareRecordByKey[K constraints.Ordered, D any](rec1, rec2 *Record[K, D]) int {
	key1 := rec1.Key()
	key2 := rec2.Key()

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

	o.m = make(Map[K, D])
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

	// Check if key allready exists and update data and sort lists if exists
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
// the front of ordered map. If key allready exists, it will be updated.
func (o *Omap[K, D]) SetFirst(key K, data D) (err error) {
	o.Lock()
	defer o.Unlock()

	// Check if key allready exists and update data and sort lists if exists
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
	for i := range o.l {
		o.l[i].Remove(rec.element())
	}

	// Remove key from map
	delete(o.m, key)

	return
}

// First gets first record from ordered map or nil if map is empty.
func (o *Omap[K, D]) First(idxs ...int) *Record[K, D] {
	o.RLock()
	defer o.RUnlock()

	var idx int
	if len(idxs) > 0 {
		idx = idxs[0]
	}

	if idx >= len(o.l) {
		return nil
	}

	return o.elementToRecord(o.l[idx].Front())
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

// Last gets last record from ordered map.
func (o *Omap[K, D]) Last(idxs ...int) *Record[K, D] {
	o.RLock()
	defer o.RUnlock()

	var idx int
	if len(idxs) > 0 {
		idx = idxs[0]
	}

	if idx >= len(o.l) {
		return nil
	}

	return o.elementToRecord(o.l[idx].Back())
}

// InsertBefore inserts record before element. Returns ErrKeyAllreadySet if key
// allready exists.
func (o *Omap[K, D]) InsertBefore(idx int, key K, data D, mark *Record[K, D]) (
	err error) {

	o.Lock()
	defer o.Unlock()

	// Check if key allready exists
	if _, ok := o.m[key]; ok {
		err = ErrKeyAllreadySet
		return
	}

	// Add new record before selected
	o.m[key] = o.insertRecord(key, data, before, mark)

	return
}

// InsertAfter inserts record after element. Returns ErrKeyAllreadySet if key
// allready exists.
func (o *Omap[K, D]) InsertAfter(key K, data D, mark *Record[K, D]) (err error) {
	o.Lock()
	defer o.Unlock()

	// Check if key allready exists
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
	o.l[0].MoveToBack(rec.element())

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
	o.l[0].MoveToFront(rec.element())
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
	o.l[0].MoveBefore(rec.element(), mark.element())

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
	o.l[0].MoveAfter(rec.element(), mark.element())

	return
}

// sortFunc sorts records using sort function. Unsafe (does not lock).
func (o *Omap[K, D]) sortFunc(idx int, f func(rec, next *Record[K, D]) int) {

	// Skip if f function not set
	if f == nil {
		return
	}

	// Sort records
	var next *list.Element
	for el := o.l[idx].Front(); el != nil; el = next {
		next = el.Next()
		o.sortRecord(idx, el, f)
	}
}

// sortRecord sorts record using sort function.
func (o *Omap[K, D]) sortRecord(idx int, elToMove *list.Element, f func(rec,
	next *Record[K, D]) int) {

	// Compare el record with next records using function f and move el if
	// neccessary
	move := false

	for el, elNext := elToMove, elToMove.Next(); ; el, elNext = elNext, elNext.Next() {

		// When the end of the list is reached
		if elNext == nil {
			// If move is set, than move elToMove record
			if move {
				o.l[idx].MoveAfter(elToMove, el)
				o.printMove(idx, false, elToMove, el)
			}
			break
		}

		// Compare elToMove record with elNext record
		if f(o.elementToRecord(elToMove), o.elementToRecord(elNext)) > 0 {
			move = true
			continue
		}

		// If move is set, than move elToMove record
		if move {
			o.l[idx].MoveBefore(elToMove, elNext)
			o.printMove(idx, true, elToMove, elNext)
		}

		break
	}
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
		rec = o.elementToRecord(o.l[0].PushBack(v))
	case 1:
		rec = o.elementToRecord(o.l[0].PushFront(v))
	case 2:
		rec = o.elementToRecord(o.l[0].InsertBefore(v, mark.element()))
	case 3:
		rec = o.elementToRecord(o.l[0].InsertAfter(v, mark.element()))
	}

	// Add element to back of additional lists and sort this lists
	for i := range o.l {
		if i == 0 {
			continue
		}
		o.l[i].PushBack(v)
		o.sortFunc(i, o.s[i])
	}

	return
}

// sortLists sorts all additional lists.
func (o *Omap[K, D]) sortLists() {
	for i := range o.s {
		// Skip basic insertion list
		if i == 0 {
			continue
		}
		// Sort additional list
		o.sortFunc(i, o.s[i])
	}
}

// Print move records. To enable print move set printMove variable to true.
func (o *Omap[K, D]) printMove(idx int, before bool, el, next *list.Element) {

	if !printMove {
		return
	}

	key := o.elementToRecord(el).Key()
	keyNext := o.elementToRecord(next).Key()

	if before {
		fmt.Printf("idx: %v, MoveBefore: %v => %v\n", idx, key, keyNext)
	} else {
		fmt.Printf("idx: %v, MoveAfter: %v => %v\n", idx, key, keyNext)
	}
}
