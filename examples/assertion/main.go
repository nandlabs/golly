// Package main demonstrates the assertion utilities.
package main

import (
	"fmt"

	"oss.nandlabs.io/golly/assertion"
)

func main() {
	// Equality checks
	fmt.Println("Equal(1, 1):", assertion.Equal(1, 1))       // true
	fmt.Println("NotEqual(1, 2):", assertion.NotEqual(1, 2)) // true

	// Empty / not-empty checks
	fmt.Println("Empty(\"\"):", assertion.Empty(""))           // true
	fmt.Println("NotEmpty(\"hi\"):", assertion.NotEmpty("hi")) // true

	// Length check
	fmt.Println("Len([1,2,3], 3):", assertion.Len([]int{1, 2, 3}, 3)) // true

	// Map checks
	m := map[string]any{"name": "golly", "version": "1.0"}
	fmt.Println("MapContains:", assertion.MapContains(m, "name", "golly")) // true
	fmt.Println("HasValue:", assertion.HasValue(m, "1.0"))                 // true

	// List checks
	list := []string{"a", "b", "c"}
	fmt.Println("ListHas(\"b\"):", assertion.ListHas("b", list))         // true
	fmt.Println("ListMissing(\"z\"):", assertion.ListMissing("z", list)) // true

	// Elements match (order-independent)
	fmt.Println("ElementsMatch:", assertion.ElementsMatch([]int{3, 1, 2}, 1, 2, 3)) // true
}
