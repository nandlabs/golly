package main

import (
	"fmt"
	"reflect"

	"oss.nandlabs.io/golly/codec"
	"oss.nandlabs.io/golly/data"
)

type Person struct {
	Name    string   `json:"name"`
	Age     int      `json:"age"`
	Address *Address `json:"address"`
}

type Address struct {
	Street  string `json:"street"`
	City    string `json:"city"`
	ZIP     string `json:"zip"`
	Country string `json:"country"`
}

func main() {
	// This will generate the schema for the Person struct
	schema, err := data.GenerateSchema(reflect.TypeOf(Person{}))
	if err != nil {
		panic(err)
	}
	// Print the generated schema
	c := codec.JsonCodec()
	s, err := c.EncodeToString(schema)
	if err != nil {
		panic(err)
	}
	fmt.Println(s)

}
