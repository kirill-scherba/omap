package main

import (
	"fmt"
	"log"

	"github.com/kirill-scherba/omap"
)

func main() {
	fmt.Println("Ordered map records example")

	// Create a new ordered map
	m, err := omap.New(omap.Index[string, string]{
		Key:  "Key",
		Func: omap.CompareByKey[string, string],
	},
	)
	if err != nil {
		log.Fatal(err)
	}

	// Insert some key-value pairs
	m.Set("key3", "value3")
	m.Set("key1", "value1")
	m.Set("key2", "value2")

	// Iterate over the omap in default (insertion) order
	fmt.Println("\nIterate over the omap in default (insertion) order:")
	for key, value := range m.Records() {
		fmt.Printf("%s: %s\n", key, value)
	}

	fmt.Println("\nIterate over the omap in key order:")
	// Iterate over the omap in key order
	for key, value := range m.Records("Key") {
		fmt.Printf("%s: %s\n", key, value)
	}
}
