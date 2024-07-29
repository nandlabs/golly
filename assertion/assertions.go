package assertion

import (
	"reflect"
	"testing"
)

// Equal compares the expected and actual values and logs an error if they are not equal
func Equal(expected, actual any) bool {
	//if expected is nil and actual is not nil
	if expected == nil && actual != nil {
		return false
	} else if expected != nil && actual == nil {
		return false
	} else if expected == nil && actual == nil {
		//if both are nil, then they are equal
		return true
	}
	return reflect.DeepEqual(expected, actual)

}

// NotEqual compares the expected and actual values and logs an error if they are equal
func NotEqual(expected, actual any) bool {

	return !Equal(expected, actual)

}

// MapContains logs an error if the map does not contain the key-value pair
func MapContains(m map[string]any, key string, value any) bool {
	if v, ok := m[key]; !ok && v != value {
		return false
	}
	return true
}

// MapMissing logs an error if the map contains the key-value pair
func MapMissing(m map[string]any, key string, value any) bool {

	return !MapContains(m, key, value)
}

// HasValue logs an error if the value is not a member of the map
func HasValue(m map[string]any, value any) bool {
	for _, v := range m {
		if v == value {
			return true
		}
	}
	return false
}

// ListHas logs an error if the list does not contain the value
func ListHas(value any, list ...any) bool {
	for _, v := range list {
		if reflect.DeepEqual(v, value) {
			return true
		}
	}
	return false
}

// ListMissing logs an error if the list contains the value
func ListMissing(t *testing.T, value any, list ...any) bool {
	for _, v := range list {
		if reflect.DeepEqual(v, value) {
			return false
		}
	}
	return true
}

// Empty checks if an object  is empty
func Empty(obj any) bool {
	if obj == nil {
		return true
	}
	val := reflect.ValueOf(obj)
	switch val.Kind() {
	case reflect.Ptr:
		// if pointer is nil return true
		if val.IsNil() {
			return true
		}
		// if pointer is not nil get the refefenced value
		referencedValue := val.Elem().Interface()
		return Empty(referencedValue)
	// if the object is a slice, array, map, string or channel check if it is empty
	case reflect.Array, reflect.Chan, reflect.Map, reflect.Slice, reflect.String:
		return val.Len() == 0
	// if the object is a struct check if all its fields are empty
	case reflect.Struct:
		for i := 0; i < val.NumField(); i++ {
			if !Empty(val.Field(i).Interface()) {
				return false
			}
		}
		return true
	case reflect.Interface:
		return val.IsNil()
	case reflect.Bool:
		return !val.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return val.Int() == 0
	case reflect.Complex128, reflect.Complex64:
		return val.Complex() == 0
	case reflect.Func, reflect.UnsafePointer:
		return false // we cannot check if a function or unsafe pointer is empty
	default:
		// check if the object is a zero value
		return reflect.DeepEqual(obj, reflect.Zero(val.Type()).Interface())
	}

}

// NotEmpty checks if an object is not empty
func NotEmpty(obj any) bool {
	return !Empty(obj)
}

// Len checks if the length of an object is equal to the expected length
func Len(obj any, expected int) bool {
	val := reflect.ValueOf(obj)
	switch val.Kind() {
	case reflect.Array, reflect.Chan, reflect.Map, reflect.Slice, reflect.String:
		return val.Len() == expected
	default:
		return false
	}
}
