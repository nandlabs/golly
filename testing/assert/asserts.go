package assert

import (
	"testing"

	"oss.nandlabs.io/golly/assertion"
)

// Equal compares the expected and actual values and logs an error if they are not equal
func Equal(t *testing.T, expected, actual any) bool {
	val := assertion.Equal(expected, actual)
	if !val {
		t.Errorf("Expected: %v, Actual: %v", expected, actual)
	}
	return val

}

// NotEqual compares the expected and actual values and logs an error if they are equal
func NotEqual(t *testing.T, expected, actual any) bool {
	val := assertion.NotEqual(expected, actual)
	if !val {
		t.Errorf("Expected: %v, Actual: %v", expected, actual)
	}
	return val
}

// True logs an error if the condition is false
func True(t *testing.T, condition bool) bool {
	if !condition {
		t.Errorf("Expected: true, Actual: false")
	}
	return true

}

// False logs an error if the condition is true
func False(t *testing.T, condition bool) bool {
	if !condition {
		t.Errorf("Expected: false, Actual: true")
	}
	return true

}

// Nil logs an error if the value is not nil
func Nil(t *testing.T, value any) bool {
	if value != nil {
		t.Errorf("Expected: nil, Actual: %v", value)
	}
	return true

}

// NotNil logs an error if the value is nil
func NotNil(t *testing.T, value any) bool {
	if value == nil {
		t.Errorf("Expected: not nil, Actual: nil")
	}
	return true
}

// Error logs an error if the error is nil
func Error(t *testing.T, err error) {
	if err == nil {
		t.Errorf("Expected: error, Actual: nil")
	}
}

// NoError logs an error if the error is not nil
func NoError(t *testing.T, err error) bool {
	if err != nil {
		t.Errorf("Expected: no error, Actual: %v", err)
	}
	return true
}

// MapContains logs an error if the map does not contain the key-value pair
func MapContains(t *testing.T, m map[string]any, key string, value any) bool {
	val := assertion.MapContains(m, key, value)
	if !val {
		t.Errorf("Expected: %v to be in %v", value, m)
	}
	return val
}

// MapMissing logs an error if the map contains the key-value pair
func MapMissing(t *testing.T, m map[string]any, key string, value any) bool {
	val := assertion.MapMissing(m, key, value)
	if !val {
		t.Errorf("Expected: %v not to be in %v", value, m)
	}
	return val
}

// HasKey logs an error if the key is not a member of the map
func HasKey(t *testing.T, m map[string]any, key string) bool {
	if _, ok := m[key]; !ok {
		t.Errorf("Expected: %v to be a key in %v", key, m)
	}
	return true
}

// HasValue logs an error if the value is not a member of the map
func HasValue(t *testing.T, m map[string]any, value any) bool {
	val := assertion.HasValue(m, value)
	if !val {
		t.Errorf("Expected: %v to be in %v", value, m)
	}
	return val
}

// ListHas logs an error if the list does not contain the value
func ListHas[S ~[]E, E any](t *testing.T, value any, list S) bool {
	val := assertion.ListHas(value, list)
	if !val {
		t.Errorf("Expected: %v to be in %v", value, list)
	}
	return val
}

// ListMissing logs an error if the list contains the value
func ListMissing[S ~[]E, E any](t *testing.T, value any, list S) bool {
	val := assertion.ListMissing(value, list)
	if !val {
		t.Errorf("Expected: %v not to be in %v", value, list)
	}
	return val
}

// Empty checks if an array is empty
func Empty(t *testing.T, obj any) bool {
	val := assertion.Empty(obj)
	if !val {
		t.Errorf("Expected: empty, Actual: not empty")
	}
	return val
}

// NotEmpty checks if an array is not empty
func NotEmpty(t *testing.T, obj any) bool {
	val := assertion.NotEmpty(obj)
	if !val {
		t.Errorf("Expected: not empty, Actual: empty")
	}
	return val
}

// Len checks if the length of the array is equal to the expected length
func Len(t *testing.T, obj any, length int) bool {
	val := assertion.Len(obj, length)
	if !val {
		t.Errorf("Expected: %v not found", length)
	}
	return val
}
