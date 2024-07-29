package assert

import (
	"fmt"
	"testing"
)

func TestEqual(t *testing.T) {
	// Test when the expected and actual values are equal
	if !Equal(t, 10, 10) {
		t.Errorf("Equal failed: expected values to be equal, but they are not")
	}

}

func TestNotEqual(t *testing.T) {
	// Test when the expected and actual values are not equal
	if !NotEqual(t, 10, 20) {
		t.Errorf("NotEqual failed: expected values to be not equal, but they are")
	}

}
func TestTrue(t *testing.T) {
	// Test when the condition is true
	if !True(t, true) {
		t.Errorf("True failed: expected condition to be true, but it is false")
	}
}
func TestFalse(t *testing.T) {
	// Test when the condition is false
	if !False(t, true) {
		t.Errorf("False failed: expected condition to be false, but it is true")
	}
}

func TestNil(t *testing.T) {
	// Test when the value is nil
	if !Nil(t, nil) {
		t.Errorf("Nil failed: expected value to be nil, but it is not")
	}
}

func TestNotNil(t *testing.T) {
	// Test when the value is not nil
	if !NotNil(t, "not nil") {
		t.Errorf("NotNil failed: expected value to be not nil, but it is nil")
	}
}

func TestError(t *testing.T) {
	// Test when the error is not nil
	Error(t, fmt.Errorf("some error"))
}

func TestNoError(t *testing.T) {
	// Test when the error is nil
	if !NoError(t, nil) {
		t.Errorf("NoError failed: expected no error, but got an error")
	}
}

func TestMapContains(t *testing.T) {
	// Test when the map contains the key-value pair
	m := map[string]interface{}{
		"key": "value",
	}
	if !MapContains(t, m, "key", "value") {
		t.Errorf("MapContains failed: expected key-value pair to be present in the map, but it is missing")
	}
}

func TestMapMissing(t *testing.T) {
	// Test when the map does not contain the key-value pair
	m := map[string]interface{}{
		"key": "value",
	}
	if !MapMissing(t, m, "key1", "value") {
		t.Errorf("MapMissing failed: expected key-value pair to be missing from the map, but it is present")
	}
}

func TestHasKey(t *testing.T) {
	// Test when the key is a member of the map
	m := map[string]interface{}{
		"key": "value",
	}
	if !HasKey(t, m, "key") {
		t.Errorf("HasKey failed: expected key to be present in the map, but it is missing")
	}
}

func TestHasValue(t *testing.T) {
	// Test when the value is a member of the map
	m := map[string]interface{}{
		"key": "value",
	}
	if !HasValue(t, m, "value") {
		t.Errorf("HasValue failed: expected value to be present in the map, but it is missing")
	}
}

func TestListHas(t *testing.T) {
	// Test when the list contains the value
	list := []interface{}{"value1", "value2", "value3"}
	if !ListHas(t, "value2", list...) {
		t.Errorf("ListHas failed: expected value to be present in the list, but it is missing")
	}
}

func TestListMissing(t *testing.T) {
	// Test when the list does not contain the value
	list := []interface{}{"value1", "value2", "value3"}
	if !ListMissing(t, "value4", list...) {
		t.Errorf("ListMissing failed: expected value to be missing from the list, but it is present")
	}
}

func TestEmpty(t *testing.T) {
	// Test when the array is empty
	arr := []interface{}{}
	if !Empty(t, arr) {
		t.Errorf("Empty failed: expected array to be empty, but it is not")
	}
}
