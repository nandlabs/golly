package assertion

import (
	"testing"
)

func TestEqual(t *testing.T) {
	// Test case 1
	expected := 10
	actual := 10
	if !Equal(expected, actual) {
		t.Errorf("Equal() = false, want true")
	}

	// Test case 2
	expected1 := "hello"
	actual1 := "world"
	if Equal(expected1, actual1) {
		t.Errorf("Equal() = true, want false")
	}

	// Test case 3
	var expected2 interface{}
	actual2 := "value"
	if Equal(expected2, actual2) {
		t.Errorf("Equal() = true, want false")
	}

	// Test case 4
	expected3 := []int{1, 2, 3}
	actual3 := []int{1, 2, 3}
	if !Equal(expected3, actual3) {
		t.Errorf("Equal() = false, want true")
	}
}

func TestNotEqual(t *testing.T) {
	// Test case 1
	expected := 10
	actual := 10
	if NotEqual(expected, actual) {
		t.Errorf("NotEqual() = true, want false")
	}

	// Test case 2
	expected1 := "hello"
	actual1 := "world"
	if !NotEqual(expected1, actual1) {
		t.Errorf("NotEqual() = false, want true")
	}

	// Test case 3
	var expected3 interface{} = nil
	actual3 := "value"
	if !NotEqual(expected3, actual3) {
		t.Errorf("NotEqual() = false, want true")
	}

	// Test case 4
	expected4 := []int{1, 2, 3}
	actual4 := []int{1, 2, 3}
	if NotEqual(expected4, actual4) {
		t.Errorf("NotEqual() = true, want false")
	}
}

func TestMapContains(t *testing.T) {
	// Test case 1
	m := map[string]interface{}{
		"key1": "value1",
		"key2": "value2",
	}
	key := "key1"
	value := "value1"
	if !MapContains(m, key, value) {
		t.Errorf("MapContains() = false, want true")
	}

	// Test case 2
	m = map[string]interface{}{
		"key1": "value1",
		"key2": "value2",
	}
	key = "key3"
	value = "value3"
	if MapContains(m, key, value) {
		t.Errorf("MapContains() = true, want false")
	}
}

func TestMapMissing(t *testing.T) {
	// Test case 1
	m := map[string]interface{}{
		"key1": "value1",
		"key2": "value2",
	}
	key := "key1"
	value := "value1"
	if MapMissing(m, key, value) {
		t.Errorf("MapMissing() = true, want false")
	}

	// Test case 2
	m = map[string]interface{}{
		"key1": "value1",
		"key2": "value2",
	}
	key = "key3"
	value = "value3"
	if !MapMissing(m, key, value) {
		t.Errorf("MapMissing() = false, want true")
	}
}

func TestHasValue(t *testing.T) {
	// Test case 1
	m := map[string]interface{}{
		"key1": "value1",
		"key2": "value2",
	}
	value := "value1"
	if !HasValue(m, value) {
		t.Errorf("HasValue() = false, want true")
	}

	// Test case 2
	m = map[string]interface{}{
		"key1": "value1",
		"key2": "value2",
	}
	value = "value3"
	if HasValue(m, value) {
		t.Errorf("HasValue() = true, want false")
	}
}

func TestListHas(t *testing.T) {
	// Test case 1
	value := 1
	list := []interface{}{1, 2, 3}
	if !ListHas(value, list...) {
		t.Errorf("ListHas() = false, want true")
	}

	// Test case 2
	value = 4
	list = []interface{}{1, 2, 3}
	if ListHas(value, list...) {
		t.Errorf("ListHas() = true, want false")
	}
}

func TestListMissing(t *testing.T) {
	// Test case 1
	value := 1
	list := []interface{}{1, 2, 3}
	if ListMissing(t, value, list...) {
		t.Errorf("ListMissing() = true, want false")
	}

	// Test case 2
	value = 4
	list = []interface{}{1, 2, 3}
	if !ListMissing(t, value, list...) {
		t.Errorf("ListMissing() = false, want true")
	}
}

func TestEmpty(t *testing.T) {
	// Test case 1
	obj := ""
	if !Empty(obj) {
		t.Errorf("Empty() = false, want true")
	}

	// Test case 2
	obj2 := []int{}
	if !Empty(obj2) {
		t.Errorf("Empty() = false, want true")
	}

	// Test case 3
	obj3 := map[string]interface{}{}
	if !Empty(obj3) {
		t.Errorf("Empty() = false, want true")
	}

	// Test case 4
	obj4 := struct{}{}
	if !Empty(obj4) {
		t.Errorf("Empty() = false, want true")
	}

	// Test case 5
	var obj5 interface{} = nil
	if !Empty(obj5) {
		t.Errorf("Empty() = false, want true")
	}

	// Test case 6
	obj6 := "hello"
	if Empty(obj6) {
		t.Errorf("Empty() = true, want false")
	}

	// Test case 7
	obj7 := []int{1, 2, 3}
	if Empty(obj7) {
		t.Errorf("Empty() = true, want false")
	}

	// Test case 8
	obj8 := map[string]interface{}{
		"key": "value",
	}
	if Empty(obj8) {
		t.Errorf("Empty() = true, want false")
	}

	// Test case 9
	obj9 := struct {
		Field string
	}{
		Field: "value",
	}
	if Empty(obj9) {
		t.Errorf("Empty() = true, want false")
	}

	// Test case 10
	obj10 := 10
	if Empty(obj10) {
		t.Errorf("Empty() = true, want false")
	}
}
