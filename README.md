# Ordered map

Omap is Golang package for working with thread safe ordered maps. The ordered
map contains the golang map, list and mutex to execute Ordered Map functions.

The Ordered Map is a map that remembers the order of items. The map can be
iterated over to retrieve the items in the order they were added.

[![GoDoc](https://godoc.org/github.com/kirill-scherba/omap?status.svg)](https://godoc.org/github.com/kirill-scherba/omap/)
[![Go Report Card](https://goreportcard.com/badge/github.com/kirill-scherba/omap)](https://goreportcard.com/report/github.com/kirill-scherba/omap)

## Basic usage example

```go
package main

import (
    "fmt"
    "log"

    "github.com/kirill-scherba/omap"
)

func main() {
    fmt.Println("Ordered map basic example")

    // Struct to store in ordered map
    type Person struct {
        Name string
        Age  int
    }

    // Create new ordered map
    o, err := omap.New[string, *Person]()
    if err != nil {
        log.Fatal(err)
    }

    // Add records to ordered map
    o.Set("John", &Person{Name: "John", Age: 30})
    o.Set("Jane", &Person{Name: "Jane", Age: 25})
    o.Set("Bob", &Person{Name: "Bob", Age: 40})
    o.Set("Alice", &Person{Name: "Alice", Age: 35})

    // Print all records
    fmt.Print("\nList records ordered by default (insertion):\n\n")
    o.ForEach(func(key string, data *Person) {
        fmt.Println(key, data)
    })
}
```

Execute this example on Go playground: [https://go.dev/play/p/T1VMf1J5n4_H](https://go.dev/play/p/T1VMf1J5n4_H)

Or run it on your local machine: `go run ./examples/basic/`

## Multiple indexes example

By default, the ordered map is created with one default (insertion) index.

You can create additional indexes by adding there sort functions and names to
the ordered map creation function. All indexes are recalculated on each
operation with map.

```go
package main

import (
    "fmt"
    "log"
    "strings"

    "github.com/kirill-scherba/omap"
)

// Struct to store in ordered map
type Person struct {
    Name string
    Age  int
}

func main() {
    fmt.Println("Ordered map multiple indexes example")

    // Create new ordered map with indexes by Name and Age
    o, err := omap.New(
        omap.Index[string, *Person]{Key: "Name", Func: CompareByName},
        omap.Index[string, *Person]{Key: "Key", Func: omap.CompareByKey[string, *Person]},
        omap.Index[string, *Person]{Key: "AgeAsc", Func: CompareByAgeAsc},
        omap.Index[string, *Person]{Key: "AgeDesc", Func: CompareByAgeDesc},
    )
    if err != nil {
        log.Fatal(err)
    }

    // Add records to ordered map
    o.Set("John", &Person{Name: "John", Age: 30})
    o.Set("Jane", &Person{Name: "Jane", Age: 25})
    o.Set("Bob", &Person{Name: "Bob", Age: 40})
    o.Set("Alice", &Person{Name: "Alice", Age: 35})

    // Print all records
    fmt.Print("\nList records ordered by default (insertion):\n\n")
    o.ForEach(func(key string, data *Person) {
        fmt.Println(key, data)
    })

    // Print all records by Name index
    fmt.Print("\nList records ordered by Name index:\n\n")
    o.ForEach(func(key string, data *Person) {
        fmt.Println(key, data)
    }, "Name")

    // Print all records by Age ascending index
    fmt.Print("\nList records ordered by Age index:\n\n")
    o.ForEach(func(key string, data *Person) {
        fmt.Println(key, data)
    }, "AgeAsc")

    // Print all records by Age descending index
    fmt.Print("\nList records ordered by Age index:\n\n")
    o.ForEach(func(key string, data *Person) {
        fmt.Println(key, data)
    }, "AgeDesc")
}

func CompareByName(r1, r2 *omap.Record[string, *Person]) int {
    return strings.Compare(
        strings.ToLower(r1.Data().Name), strings.ToLower(r2.Data().Name),
    )
}

func CompareByAgeAsc(r1, r2 *omap.Record[string, *Person]) int {
    return r1.Data().Age - r2.Data().Age
}

func CompareByAgeDesc(r1, r2 *omap.Record[string, *Person]) int {
    return r2.Data().Age - r1.Data().Age
}

```

Execute this example on Go playground: [https://go.dev/play/p/ppUx3Afn7ky](https://go.dev/play/p/ppUx3Afn7ky)

Or run it on your local machine: `go run ./examples/multiple_index/`

## Licence

[BSD](LICENSE)
