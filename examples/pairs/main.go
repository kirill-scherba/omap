package main

import (
	"fmt"
	"log"

	"github.com/kirill-scherba/omap"
)

func main() {
	// Create a new ordered map
	m, err := omap.New[string, string]()
	if err != nil {
		log.Fatal(err)
	}

	// Insert some key-value pairs
	m.Set("key1", "value1")
	m.Set("key2", "value2")
	m.Set("key3", "value3")

	// Iterate over the omap in order
	for _, pair := range m.Pairs() {
		fmt.Printf("%s: %s\n", pair.Key, pair.Value)
	}
}
