package assert

import (
	"reflect"
	"testing"
)

// Equal compares the expected and actual values and logs an error if they are not equal
func Equal(t *testing.T, expected, actual any) {
	//if expected is nil and actual is not nil
	if expected == nil && actual != nil {
		t.Errorf("Expected: %v, Actual: %v", expected, actual)
	} else if expected != nil && actual == nil {
		t.Errorf("Expected: %v, Actual: %v", expected, actual)

	} else if expected == nil && actual == nil {
		//if both are nil, then they are equal
		return
		//if types of expected and actual are different

	} else if !reflect.DeepEqual(expected, actual) {
		t.Errorf("Expected: %v, Actual: %v", expected, actual)
	}

}

// NotEqual compares the expected and actual values and logs an error if they are equal
func NotEqual(t *testing.T, expected, actual any) {
	//if expected is nil and actual is not nil
	if expected == nil && actual != nil {
		t.Errorf("Expected: %v, Actual: %v", expected, actual)
	} else if expected != nil && actual == nil {
		t.Errorf("Expected: %v, Actual: %v", expected, actual)

	} else if expected == nil && actual == nil {
		t.Errorf("Expected: %v, Actual: %v", expected, actual)
	} else if reflect.DeepEqual(expected, actual) {
		t.Errorf("Expected: %v, Actual: %v", expected, actual)
	}
}

// True logs an error if the condition is false
func True(t *testing.T, condition bool) {
	if !condition {
		t.Errorf("Expected: true, Actual: false")
	}
}

// False logs an error if the condition is true
func False(t *testing.T, condition bool) {
	if condition {
		t.Errorf("Expected: false, Actual: true")
	}
}

// Nil logs an error if the value is not nil
func Nil(t *testing.T, value any) {
	if value != nil {
		t.Errorf("Expected: nil, Actual: %v", value)
	}
}

// NotNil logs an error if the value is nil
func NotNil(t *testing.T, value any) {
	if value == nil {
		t.Errorf("Expected: not nil, Actual: nil")
	}
}

// Error logs an error if the error is nil
func Error(t *testing.T, err error) {
	if err == nil {
		t.Errorf("Expected: error, Actual: nil")
	}
}

// NoError logs an error if the error is not nil
func NoError(t *testing.T, err error) {
	if err != nil {
		t.Errorf("Expected: no error, Actual: %v", err)
	}
}

// MapContains logs an error if the map does not contain the key-value pair
func MapContains(t *testing.T, m map[string]any, key string, value any) {
	if v, ok := m[key]; !ok || v != value {
		t.Errorf("Expected: %v to be in %v", value, m)
	}
}

// MapMissing logs an error if the map contains the key-value pair
func MapMissing(t *testing.T, m map[string]any, key string, value any) {
	if v, ok := m[key]; ok && v == value {
		t.Errorf("Expected: %v not to be in %v", value, m)
	}
}

// HasKey logs an error if the key is not a member of the map
func HasKey(t *testing.T, m map[string]any, key string) {
	if _, ok := m[key]; !ok {
		t.Errorf("Expected: %v to be a key in %v", key, m)
	}
}

// HasValue logs an error if the value is not a member of the map
func HasValue(t *testing.T, m map[string]any, value any) {
	for _, v := range m {
		if v == value {
			return
		}
	}
	t.Errorf("Expected: %v to be a value in %v", value, m)
}

// ListHas logs an error if the list does not contain the value
func ListHas(t *testing.T, value any, list ...any) {
	for _, v := range list {
		if reflect.DeepEqual(v, value) {
			return
		}
	}
	t.Errorf("Expected: %v to be in %v", value, list)
}

// ListMissing logs an error if the list contains the value
func ListMissing(t *testing.T, value any, list ...any) {
	for _, v := range list {
		if reflect.DeepEqual(v, value) {
			t.Errorf("Expected: %v not to be in %v", value, list)
		}
	}
}
