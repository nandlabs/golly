package config

import (
	"strings"
	"testing"
)

func TestProperties_Get(t *testing.T) {
	p := NewProperties()
	p.Put("key1", "value1")
	p.Put("key2", "value2")

	// Test case 1: Key exists
	expected1 := "value1"
	actual1 := p.Get("key1", "")
	if actual1 != expected1 {
		t.Errorf("Expected %s, but got %s", expected1, actual1)
	}

	// Test case 2: Key does not exist, return default value
	expected2 := "default"
	actual2 := p.Get("key3", "default")
	if actual2 != expected2 {
		t.Errorf("Expected %s, but got %s", expected2, actual2)
	}
}

func TestProperties_Put(t *testing.T) {
	p := NewProperties()

	// Test case 1: Put new key-value pair
	expected1 := "value1"
	actual1 := p.Put("key1", "value1")
	if actual1 != "" {
		t.Errorf("Expected empty string, but got %s", actual1)
	}
	if p.Get("key1", "") != expected1 {
		t.Errorf("Expected %s, but got %s", expected1, p.Get("key1", ""))
	}

	// Test case 2: Put existing key-value pair
	expected2 := "value1"
	actual2 := p.Put("key1", "value2")

	if actual2 != expected2 {
		t.Errorf("Expected %s, but got %s", expected2, actual2)
	}
	if p.Get("key1", "") != "value2" {
		t.Errorf("Expected %s, but got %s", expected2, p.Get("key1", ""))
	}
}

func TestProperties_ReadFrom(t *testing.T) {
	p := NewProperties()

	// Test case 1: Read properties from reader
	input := strings.NewReader("key1=value1\nkey2=value2\n")
	err := p.Load(input)
	if err != nil {
		t.Errorf("Error reading properties: %s", err.Error())
	}

	expected1 := "value1"
	actual1 := p.Get("key1", "")
	if actual1 != expected1 {
		t.Errorf("Expected %s, but got %s", expected1, actual1)
	}

	expected2 := "value2"
	actual2 := p.Get("key2", "")
	if actual2 != expected2 {
		t.Errorf("Expected %s, but got %s", expected2, actual2)
	}
}
