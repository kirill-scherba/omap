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
