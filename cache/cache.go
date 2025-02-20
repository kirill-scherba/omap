// Copyright 2025 Kirill Scherba <kirill@scherba.ru>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package cache provides an inmemory cache implementation based on ordered map.
//
// The cache is a generic type that can store any type of data. The cache is
// implemented with an ordered map, which is a thread-safe map that remembers
// the order of items. The cache is limited to the size specified when creating
// a new cache object.
//
// The cache provides the following methods:
//   - Set: adds a new item to the cache. If the item already exists, the old
//     item is replaced with the new one.
//   - Get: returns the item associated with the given key.
//   - Del: deletes the item associated with the given key.
//   - Len: returns the number of items in the cache.
package cache

import (
	"github.com/kirill-scherba/omap"
)

// Cache is a struct that contains an ordered map to store T objects. The
// ordered map is implemented with omap, which is a thread-safe ordered map.
// The size of the cache is limited to the value of the size field.
type Cache[T any] struct {
	// Omap is an ordered map to store T objects.
	m *omap.Omap[string, T]
	// size is the maximum number of elements in the cache.
	size int
}

// New creates new cache object.
//
// Parameters:
//   - size: the maximum number of elements in the cache.
//
// Returns:
//   - c: the new cache object.
//   - err: an error if the operation fails.
func New[T any](size int) (c *Cache[T], err error) {
	// Create new omap object
	m, err := omap.New[string, T]()
	if err != nil {
		return
	}

	// Create new Cache object
	c = &Cache[T]{m, size}
	return
}

// Add data to cache by key.
//
// Parameters:
//   - key: the key to add record to cache.
//   - data: the data to add to cache.
//
// Returns:
//   - err: an error if the operation fails.
func (c *Cache[T]) Set(key string, data T) (err error) {

	// Add new record to top of index list
	err = c.m.SetFirst(key, data)
	if err != nil {
		return
	}

	// Check cache size and remove last record if size is exceeded
	if c.m.Len() > c.size {
		// Remove last record from the cache
		c.m.DelLast()
	}

	return
}

// Get record from cache by key.
//
// Parameters:
//   - key: the key to get record from cache.
//
// Returns:
//   - data: the data from cache if the operation is successful.
//   - ok: true if the operation is successful.
func (c *Cache[T]) Get(key string) (data T, ok bool) {

	// Get players saves from cache
	rec, ok := c.m.GetRecord(key)
	if !ok {
		return
	}
	data = rec.Data()

	// Move players saves up in basic index lists
	c.m.Idx.MoveUp(rec)

	return
}

// Del removes record from cache by key.
//
// Parameters:
//   - key: the key to remove record from cache.
//
// Returns:
//   - data: the data from cache if the operation is successful.
//   - ok: true if the operation is successful.
func (c *Cache[T]) Del(key string) (data T, ok bool) {
	return c.m.Del(key)
}

// Len returns the number of items in the cache.
//
// Returns:
//   - len: the number of items in the cache.
func (c *Cache[T]) Len() int {
	return c.m.Len()
}
