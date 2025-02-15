// Copyright 2025 Kirill Scherba <kirill@scherba.ru>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Record of ordered map definition.

package omap

import "container/list"

// Record is a struct that contains list element and methods.
type Record[K comparable, D any] list.Element

// recordValue is a struct that contains key and value of ordered map. It is
// used to store key and data in list element.
type recordValue[K comparable, D any] struct {
	Key  K
	Data D
}

// Key returns record key.
func (r *Record[K, D]) Key() (key K) {
	if v, ok := r.Value.(recordValue[K, D]); ok {
		key = v.Key
	}
	return
}

// Data returns record data.
func (r *Record[K, D]) Data() (data D) {
	if v, ok := r.Value.(recordValue[K, D]); ok {
		data = v.Data
	}
	return
}

// element returns list element from record.
func (r *Record[K, D]) element() *list.Element {
	return (*list.Element)(r)
}

// elementToRecord converts list element to Record.
func (o *Omap[K, D]) elementToRecord(el *list.Element) *Record[K, D] {
	return (*Record[K, D])(el)
}
