package omap

import "testing"

func TestOmap(t *testing.T) {
	t.Log("test")

	m := New[int, int](CompareRecordByKey[int, int])
	m.Set(1, 2)
	m.Set(2, 3)
	m.Set(3, 4)
	m.Set(4, 5)
	m.Set(5, 6)
	m.Set(6, 7)
	m.Set(7, 8)
	m.Set(8, 9)
	m.Set(9, 10)
	m.Set(10, 11)
	m.Set(11, 12)
	m.Set(12, 13)
	m.Set(13, 14)
	m.Set(14, 15)
	m.Set(15, 16)
	m.Set(16, 17)
	m.Set(17, 18)
	m.Set(18, 19)
	m.Set(19, 20)
	m.Set(20, 21)

	// Print all records
	for i := 1; i <= 20; i++ {
		d, ok := m.Get(int(i))
		t.Log(i, ok, d)
	}

	t.Log()

	// Get first record from ordered map
	rec := m.First()
	key, d, err := RecordValue[int, int](rec)
	t.Log("first key:", key, "data:", d, "err:", err)

	// Get next record from ordered map
	rec = m.Next(rec)
	key, d, err = RecordValue[int, int](rec)
	t.Log("next key :", key, "data:", d, "err:", err)

	// Get previous record from ordered map
	rec = m.Prev(rec)
	key, d, err = RecordValue[int, int](rec)
	t.Log("prev key :", key, "data:", d, "err:", err)

	// Get last record from ordered map
	rec = m.Last()
	key, d, err = RecordValue[int, int](rec)
	t.Log("last key :", key, "data:", d, "err:", err)

	// Move last record to the front of ordered map
	m.MoveToFront(rec)
	first := m.First()
	key, d, err = RecordValue[int, int](first)
	t.Log("first key:", key, "data:", d, "err:", err)

	// Get next record from ordered map
	rec = m.Next(first)
	key, d, err = RecordValue[int, int](rec)
	t.Log("next key :", key, "data:", d, "err:", err)

	// Move berfore
	m.MoveBefore(rec, first)
	first = m.First()
	key, d, err = RecordValue[int, int](first)
	t.Log("first key:", key, "data:", d, "err:", err)

	t.Log()

	// Print all records
	for rec := m.First(); rec != nil; rec = m.Next(rec) {
		key, d, err := RecordValue[int, int](rec)
		t.Log(key, d, err)
	}

	// Sort records using sort function
	m.sortFunc(0, func(rec1, rec2 Record) int {
		v1, _, _ := RecordValue[int, int](rec1)
		v2, _, _ := RecordValue[int, int](rec2)
		return v2 - v1
	})

	// Print all records by default(insertion) order
	for rec := m.First(); rec != nil; rec = m.Next(rec) {
		key, d, err := RecordValue[int, int](rec)
		t.Log(key, d, err)
	}

	t.Log("\nlist by keys:")

	// Sort records using sort function
	// m.SortFunc(0, CompareRecordByKey[int, int])

	// Print all records by key order
	for rec := m.First(1); rec != nil; rec = m.Next(rec) {
		key, d, err := RecordValue[int, int](rec)
		t.Log(key, d, err)
	}
}
