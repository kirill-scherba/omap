# Ordered map

Omap is Golang package for working with thread safe ordered maps. The ordered
map contains the golang map, list and mutex to execute Ordered Map functions.

The Ordered Map is a map that remembers the order of items. The map can be
iterated over to retrieve the items in the order they were added.

[![GoDoc](https://godoc.org/github.com/kirill-scherba/omap?status.svg)](https://godoc.org/github.com/kirill-scherba/omap/)
[![Go Report Card](https://goreportcard.com/badge/github.com/kirill-scherba/omap)](https://goreportcard.com/report/github.com/kirill-scherba/omap)

## Introduction to the omap Go Package

The omap Go package is a lightweight and efficient library for working with
ordered maps in Go. An ordered map is a data structure that combines the
benefits of a map and a list, allowing you to store key-value pairs in a
specific order.

## What is omap?

Omap is a Go package that provides an implementation of an ordered map. It is
designed to be fast, efficient, and easy to use. Omap is particularly useful
when you need to store data in a specific order, such as when working with
configuration files, caching, or data processing pipelines.

## Key Features of omap

- Ordered: omap preserves the order in which key-value pairs are inserted,
allowing you to iterate over the map in a specific order.

- Fast lookups: omap uses a hash table to store key-value pairs, making lookups
fast and efficient.

- Efficient insertion and deletion: omap uses a linked list to store the order of
key-value pairs, making insertion and deletion operations efficient.
Using omap

## Using omap

To use omap, you can install it using the following command:

```bash
go get github.com/kirill-scherba/omap
```

Here is an example of how to use omap:

```go
package main

import (
    "fmt"
    "github.com/fatih/omap"
)

func main() {
    // Create a new omap
    m := omap.New()

    // Insert some key-value pairs
    m.Set("key1", "value1")
    m.Set("key2", "value2")
    m.Set("key3", "value3")

    // Iterate over the omap in order
    for _, pair := range m.Pairs() {
        fmt.Printf("%s: %s\n", pair.Key, pair.Value)
    }
}
```

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
