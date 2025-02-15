package omap

import (
	"strings"
	"testing"
)

func TestOmap(t *testing.T) {
	t.Log("TestOmap")

	// Enable print move
	printMode = true

	m, err := New(Index[int, int]{Key: "key", Func: CompareByKey[int, int]})
	if err != nil {
		t.Fatal(err)
	}

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
	err = m.MoveToFront(rec)
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
		return rec2.Key() - rec1.Key()
	})

	// Print all records by default(insertion) order
	for rec := m.First(); rec != nil; rec = m.Next(rec) {
		t.Log(rec.Key(), rec.Data())
	}

	t.Log("\nlist sorted by key index:")

	// Sort records using sort function
	// m.SortFunc(0, CompareRecordByKey[int, int])

	// Print all records by key order
	for rec := m.First("key"); rec != nil; rec = m.Next(rec) {
		t.Log(rec.Key(), rec.Data())
	}
}

func TestBasicExample(t *testing.T) {
	t.Log("TestBasicExample")

	// Struct to store in ordered map
	type Person struct {
		Name string
		Age  int
	}

	// Create new ordered map
	o, err := New[string, Person]()
	if err != nil {
		t.Fatal(err)
	}

	// Add records to ordered map
	o.Set("John", Person{Name: "John", Age: 30})
	o.Set("Jane", Person{Name: "Jane", Age: 25})
	o.Set("Bob", Person{Name: "Bob", Age: 40})
	o.Set("Alice", Person{Name: "Alice", Age: 35})

	// Print all records
	t.Log("\nlist records ordered by default (insertion):")
	for rec := o.First(); rec != nil; rec = o.Next(rec) {
		t.Log(rec.Key(), rec.Data())
	}
}

func TestForEach(t *testing.T) {
	t.Log("TestBasicExample")

	// Struct to store in ordered map
	type Person struct {
		Name string
		Age  int
	}

	// Create new ordered map
	o, err := New[string, *Person]()
	if err != nil {
		t.Fatal(err)
	}

	// Add records to ordered map
	o.Set("John", &Person{Name: "John", Age: 30})
	o.Set("Jane", &Person{Name: "Jane", Age: 25})
	o.Set("Bob", &Person{Name: "Bob", Age: 40})
	o.Set("Alice", &Person{Name: "Alice", Age: 35})

	// Print all records
	t.Log("\nlist records ordered by default (insertion):")
	o.ForEach(func(key string, data *Person) {
		t.Log(key, data)
	})
}

// Struct to store in ordered map
type Person struct {
	Name string
	Age  int
}

func TestForEachIndex(t *testing.T) {
	t.Log("TestBasicExample")

	printMode = true

	// Create new ordered map with indexes by Name and Age
	o, err := New(
		Index[string, *Person]{Key: "Name", Func: CompareRecordsByName},
		Index[string, *Person]{Key: "Key", Func: CompareByKey[string, *Person]},
		Index[string, *Person]{Key: "AgeAsc", Func: CompareRecordsByAgeAsc},
		Index[string, *Person]{Key: "AgeDesc", Func: CompareRecordsByAgeDesc},
	)
	if err != nil {
		t.Fatal(err)
	}

	// Add records to ordered map
	o.Set("John", &Person{Name: "John", Age: 30})
	o.Set("Jane", &Person{Name: "Jane", Age: 25})
	o.Set("Bob", &Person{Name: "Bob", Age: 40})
	o.Set("Alice", &Person{Name: "Alice", Age: 35})

	// Print all records ordered by default (insertion)
	t.Log("\nlist records ordered by default (insertion):")
	o.ForEach(func(key string, data *Person) {
		t.Log(key, data)
	})

	// Print all records ordered by Name index
	t.Log("\nlist records ordered by Name index:")
	o.ForEach(func(key string, data *Person) {
		t.Log(key, data)
	}, "Name")

	// Print all records by Age index
	t.Log("\nlist records ordered by Age ascending index:")
	o.ForEach(func(key string, data *Person) {
		t.Log(key, data)
	}, "AgeAsc")

	// Print all records by Age index
	t.Log("\nlist records ordered by Age descending index:")
	o.ForEach(func(key string, data *Person) {
		t.Log(key, data)
	}, "AgeDesc")

}

func CompareRecordsByName(r1, r2 *Record[string, *Person]) int {
	return strings.Compare(
		strings.ToLower(r1.Data().Name), strings.ToLower(r2.Data().Name),
	)
}

func CompareRecordsByAgeAsc(r1, r2 *Record[string, *Person]) int {
	return r1.Data().Age - r2.Data().Age
}

func CompareRecordsByAgeDesc(r1, r2 *Record[string, *Person]) int {
	return r2.Data().Age - r1.Data().Age
}
