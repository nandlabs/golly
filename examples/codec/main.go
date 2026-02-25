// Package main demonstrates the codec package for encoding/decoding data.
package main

import (
	"fmt"
	"log"

	"oss.nandlabs.io/golly/codec"
)

type Person struct {
	Name string `json:"name" xml:"name" yaml:"name"`
	Age  int    `json:"age" xml:"age" yaml:"age"`
	City string `json:"city" xml:"city" yaml:"city"`
}

func main() {
	person := Person{Name: "Alice", Age: 30, City: "Portland"}

	// JSON encoding
	jsonCodec := codec.JsonCodec()
	jsonStr, err := jsonCodec.EncodeToString(person)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("JSON:", jsonStr)

	// Decode from JSON string
	var decoded Person
	err = jsonCodec.DecodeString(jsonStr, &decoded)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Decoded from JSON: %+v\n", decoded)

	// YAML encoding
	yamlCodec := codec.YamlCodec()
	yamlStr, err := yamlCodec.EncodeToString(person)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("YAML:\n" + yamlStr)

	// XML encoding
	xmlCodec := codec.XmlCodec()
	xmlStr, err := xmlCodec.EncodeToString(person)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("XML:", xmlStr)

	// Get codec by content type
	c, err := codec.GetDefault("application/json")
	if err != nil {
		log.Fatal(err)
	}
	s, _ := c.EncodeToString(map[string]string{"key": "value"})
	fmt.Println("Dynamic codec:", s)
}
