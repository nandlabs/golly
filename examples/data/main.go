// Package main demonstrates the data package with Pipeline and Schema generation.
package main

import (
	"fmt"
	"log"
	"reflect"

	"oss.nandlabs.io/golly/codec"
	"oss.nandlabs.io/golly/data"
)

type User struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Age   int    `json:"age"`
}

func main() {
	// --- Pipeline: key-value data container ---
	pipeline := data.NewPipeline("example-pipeline")
	fmt.Println("Pipeline ID:", pipeline.Id())

	// Set values
	_ = pipeline.Set("name", "Alice")
	_ = pipeline.Set("age", 30)
	_ = pipeline.Set("tags", []string{"admin", "user"})

	// Get values
	name, err := pipeline.Get("name")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Name:", name)

	// Check existence
	fmt.Println("Has 'age':", pipeline.Has("age"))
	fmt.Println("Has 'address':", pipeline.Has("address"))

	// List keys
	fmt.Println("Keys:", pipeline.Keys())

	// Extract typed value
	age, err := data.ExtractValue[int](pipeline, "age")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Age (typed):", age)

	// --- Schema generation from struct ---
	schema, err := data.GenerateSchema(reflect.TypeOf(User{}))
	if err != nil {
		log.Fatal(err)
	}

	c := codec.JsonCodec()
	s, err := c.EncodeToString(schema)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("\nGenerated Schema:")
	fmt.Println(s)
}
