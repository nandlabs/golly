// Package main demonstrates the UUID generation utilities.
package main

import (
	"fmt"
	"log"

	"oss.nandlabs.io/golly/uuid"
)

func main() {
	// Generate a V1 UUID (time-based)
	v1, err := uuid.V1()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("UUID V1:", v1.String())

	// Generate a V4 UUID (random)
	v4, err := uuid.V4()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("UUID V4:", v4.String())

	// Generate a V3 UUID (namespace + name, MD5)
	v3, err := uuid.V3("6ba7b810-9dad-11d1-80b4-00c04fd430c8", "example.com")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("UUID V3:", v3.String())

	// Parse a UUID string
	parsed, err := uuid.ParseUUID(v4.String())
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Parsed UUID:", parsed.String())
}
