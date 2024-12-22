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
	if !ListHas(value, list) {
		t.Errorf("ListHas() = false, want true")
	}

	// Test case 2
	value = 4
	list = []interface{}{1, 2, 3}
	if ListHas(value, list) {
		t.Errorf("ListHas() = true, want false")
	}
}

func TestListMissing(t *testing.T) {
	// Test case 1
	value := 1
	list := []interface{}{1, 2, 3}
	if ListMissing(value, list) {
		t.Errorf("ListMissing() = true, want false")
	}

	// Test case 2
	value = 4
	list = []interface{}{1, 2, 3}
	if !ListMissing(value, list) {
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
func TestLen(t *testing.T) {
	// Test case 1
	obj1 := []int{1, 2, 3}
	expected1 := 3
	if !Len(obj1, expected1) {
		t.Errorf("Len() = false, want true")
	}

	// Test case 2
	obj2 := "hello"
	expected2 := 5

	if !Len(obj2, expected2) {
		t.Errorf("Len() = true, want false")
	}

	// Test case 3
	obj3 := map[string]int{
		"key1": 1,
		"key2": 2,
	}
	expected3 := 2
	if !Len(obj3, expected3) {
		t.Errorf("Len() = false, want true")
	}

	// Test case 4
	obj4 := []string{}
	expected4 := 0
	if !Len(obj4, expected4) {
		t.Errorf("Len() = false, want true")
	}

	// Test case 5
	obj5 := make(chan int, 10)
	expected5 := 10
	if Len(obj5, expected5) {
		t.Errorf("Len() = true, want false")
	}
}
func TestNotEmpty(t *testing.T) {
	// Test case 1
	obj1 := ""
	if NotEmpty(obj1) {
		t.Errorf("NotEmpty() = true, want false")
	}

	// Test case 2
	obj2 := []int{1, 2, 3}
	if !NotEmpty(obj2) {
		t.Errorf("NotEmpty() = true, want false")
	}

	// Test case 3
	obj3 := map[string]interface{}{
		"key": "value",
	}
	if !NotEmpty(obj3) {
		t.Errorf("NotEmpty() = true, want false")
	}

	// Test case 4
	obj4 := struct {
		Field string
	}{
		Field: "value",
	}
	if !NotEmpty(obj4) {
		t.Errorf("NotEmpty() = true, want false")
	}

	// Test case 5
	obj5 := 10
	if !NotEmpty(obj5) {
		t.Errorf("NotEmpty() = true, want false")
	}

	// Test case 6
	obj6 := []int{}
	if NotEmpty(obj6) {
		t.Errorf("NotEmpty() = false, want true")
	}

	// Test case 7
	obj7 := map[string]interface{}{}
	if NotEmpty(obj7) {
		t.Errorf("NotEmpty() = false, want true")
	}

	// Test case 8
	obj8 := struct{}{}
	if NotEmpty(obj8) {
		t.Errorf("NotEmpty() = false, want true")
	}

	// Test case 9
	var obj9 interface{} = nil
	if NotEmpty(obj9) {
		t.Errorf("NotEmpty() = false, want true")
	}
}
func TestElementsMatch(t *testing.T) {
	// Test case 1
	list1 := []int{1, 2, 3}
	expectedElements1 := []int{1, 2, 3}
	if !ElementsMatch(list1, expectedElements1...) {
		t.Errorf("Elements() = false, want true")
	}

	// Test case 2
	list2 := []string{"a", "b", "c"}
	expectedElements2 := []string{"a", "b", "c"}
	if !ElementsMatch(list2, expectedElements2...) {
		t.Errorf("Elements() = false, want true")
	}

	// Test case 3
	list3 := []float64{1.1, 2.2, 3.3}
	expectedElements3 := []float64{1.1, 2.2, 3.3}
	if !ElementsMatch(list3, expectedElements3...) {
		t.Errorf("Elements() = false, want true")
	}

	// Test case 4
	list4 := []int{1, 2, 3}
	expectedElements4 := []int{3, 2, 1}
	if ElementsMatch(list4, expectedElements4...) {
		t.Errorf("Elements() = true, want false")
	}

	// Test case 5
	list5 := []string{"a", "b", "c"}
	expectedElements5 := []string{"a", "b"}
	if ElementsMatch(list5, expectedElements5...) {
		t.Errorf("Elements() = true, want false")
	}

	// Test case 6
	list6 := []interface{}{1, "two", 3.0}
	expectedElements6 := []interface{}{1, "two", 3.0}
	if !ElementsMatch(list6, expectedElements6...) {
		t.Errorf("Elements() = false, want true")
	}

	// Test case 7
	list7 := []interface{}{1, "two", 3.0}
	expectedElements7 := []interface{}{1, "two", 4.0}
	if ElementsMatch(list7, expectedElements7...) {
		t.Errorf("Elements() = true, want false")
	}
}
