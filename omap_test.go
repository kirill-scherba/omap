package omap

import "testing"

func TestOmap(t *testing.T) {
	t.Log("test")

	// Enable print move
	printMove = true

	m := New[int, int](CompareRecordByKey[int, int])

	t.Log("\nset records:")

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

	t.Log("\nlist records get with Get function:")

	// Print all records
	for i := 1; i <= 20; i++ {
		d, ok := m.Get(int(i))
		t.Log(i, ok, d)
	}

	t.Log()

	// Get first record from ordered map
	rec := m.First()
	if rec == nil {
		t.Fatal("first record is nil")
	}
	t.Log("first key:", rec.Key(), "data:", rec.Data())

	// Get next record from ordered map
	rec = m.Next(rec)
	if rec == nil {
		t.Fatal("next record is nil")
	}
	t.Log("next key :", rec.Key(), "data:", rec.Data())

	// Get previous record from ordered map
	rec = m.Prev(rec)
	if rec == nil {
		t.Fatal("prev record is nil")
	}
	t.Log("prev key :", rec.Key(), "data:", rec.Data())

	// Get last record from ordered map
	rec = m.Last()
	if rec == nil {
		t.Fatal("last record is nil")
	}
	t.Log("last key :", rec.Key(), "data:", rec.Data())

	// Move last record to the front of ordered map
	err := m.MoveToFront(rec)
	first := m.First()
	if first == nil {
		t.Fatal("first record is nil")
	}
	t.Log("first key:", first.Key(), "data:", first.Data(), "err:", err)

	// Get next record from ordered map
	rec = m.Next(first)
	if rec == nil {
		t.Fatal("next record is nil")
	}
	t.Log("next key :", rec.Key(), "data:", rec.Data(), "err:")

	// Move berfore
	err = m.MoveBefore(rec, first)
	first = m.First()
	if first == nil {
		t.Fatal("first record is nil")
	}
	t.Log("first key:", first.Key(), "data:", first.Data(), "err:", err)

	t.Log()

	t.Log("\nlist after move records:")

	// Print all records
	for rec := m.First(); rec != nil; rec = m.Next(rec) {
		t.Log(rec.Key(), rec.Data())
	}

	t.Log("\nlist sorted by function:")

	// Sort records using sort function
	m.sortFunc(0, func(rec1, rec2 *Record[int, int]) int {
		return rec1.Key() - rec2.Key()
	})

	// Print all records by default(insertion) order
	for rec := m.First(); rec != nil; rec = m.Next(rec) {
		t.Log(rec.Key(), rec.Data())
	}

	t.Log("\nlist sorted by keys index:")

	// Sort records using sort function
	// m.SortFunc(0, CompareRecordByKey[int, int])

	// Print all records by key order
	for rec := m.First(1); rec != nil; rec = m.Next(rec) {
		t.Log(rec.Key(), rec.Data())
	}
}
