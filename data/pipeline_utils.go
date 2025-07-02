package data

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// ExtractValue retrieves a value of type T from the Pipeline using the provided path.
// If the value is not of type T, it returns an ErrInvalidType error.
// The path can be a simple key or a dot-separated path (e.g., "user.address.city").
//
// Parameters:
//   - c: A pointer to the Pipeline from which to extract the value.
//   - path: The path to the value to be retrieved. Can be a simple key or dot notation (e.g., "user.address.city").
//
// Returns:
//   - value: The value of type T associated with the provided path.
//   - err: An error if the path does not exist or the value is not of type T.
func ExtractValue[T any](c Pipeline, path string) (value T, err error) {
	if !strings.Contains(path, ".") {
		var v any
		v, err = c.Get(path)
		if err != nil {
			return
		}
		return Convert[T](v)
	}
	parts := strings.Split(path, ".")
	rootKey, filter, hasFilter := parseFieldAndFilter(parts[0])
	var current any
	current, err = c.Get(rootKey)
	if err != nil {
		return
	}
	if hasFilter {
		current, err = applyFilter(current, filter)
		if err != nil {
			return
		}
	}
	for i := 1; i < len(parts); i++ {
		field, filter, hasFilter := parseFieldAndFilter(parts[i])
		if len(field) > 0 {
			current, err = navigateToField(current, field)
			if err != nil {
				switch err {
				case ErrFieldNotFound:
					err = fmt.Errorf("%w: field '%s' in path '%s'", ErrFieldNotFound, field, path)
				case ErrInvalidPath:
					err = fmt.Errorf("%w: invalid segment '%s' in path '%s'", ErrInvalidPath, field, path)
				}
				return
			}
			if hasFilter {
				current, err = applyFilter(current, filter)
				if err != nil {
					return
				}
			}
		} else {
			err = ErrInvalidPath
			return
		}
	}
	if current == nil {
		err = ErrInvalidType
		return
	}
	return Convert[T](current)
}

// navigateToField navigates to a field within a value using reflection.
// It handles maps, structs, and other types that can contain nested fields.
func navigateToField(value any, fieldName string) (any, error) {
	if value == nil {
		return nil, ErrFieldNotFound
	}

	v := reflect.ValueOf(value)

	// Handle different types
	switch v.Kind() {
	case reflect.Map:
		// For maps, get the value using the field name as key

		// For string keys, handle directly
		if mv, ok := value.(map[string]any); ok {
			if val, exists := mv[fieldName]; exists {
				return val, nil
			}
			return nil, ErrFieldNotFound
		}

		// For other map types, use reflection
		mapKey := reflect.ValueOf(fieldName)
		mapValue := v.MapIndex(mapKey)
		if !mapValue.IsValid() {
			return nil, ErrFieldNotFound
		}
		return mapValue.Interface(), nil

	case reflect.Struct:
		// For structs, get the field using reflection
		field := v.FieldByName(fieldName)
		if !field.IsValid() {
			return nil, ErrFieldNotFound
		}
		return field.Interface(), nil

	case reflect.Ptr:
		// For pointers, dereference and try again
		if v.IsNil() {
			return nil, ErrFieldNotFound
		}
		return navigateToField(v.Elem().Interface(), fieldName)

	case reflect.Slice, reflect.Array:
		// Try to parse the field name as an index
		index, err := strconv.Atoi(fieldName)
		if err != nil {
			return nil, ErrInvalidPath
		}

		// Check if the index is valid
		if index < 0 || index >= v.Len() {
			return nil, ErrFieldNotFound
		}

		return v.Index(index).Interface(), nil

	default:
		return nil, fmt.Errorf("cannot navigate into value of type %T", value)
	}
}

// Extracts the field name and optional filter from a path segment, e.g. users[0], users[name=="nanda"]
func parseFieldAndFilter(segment string) (fieldName string, filter string, hasFilter bool) {
	if open := strings.Index(segment, "["); open != -1 {
		if close := strings.Index(segment, "]"); close != -1 && close > open {
			fieldName = segment[:open]
			filter = segment[open+1 : close]
			hasFilter = true
			return
		}
	}
	fieldName = segment
	return
}

// Applies a filter to a value (slice, array, or map). Supports index and key==value filters.
func applyFilter(value any, filter string) (any, error) {
	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.Slice, reflect.Array:
		// Index filter: [0]
		if idx, err := strconv.Atoi(filter); err == nil {
			if idx < 0 || idx >= v.Len() {
				return nil, ErrFieldNotFound
			}
			return v.Index(idx).Interface(), nil
		}
		// Key==value filter: [name=="nanda"]
		if strings.Contains(filter, "==") {
			parts := strings.SplitN(filter, "==", 2)
			key := strings.TrimSpace(parts[0])
			val := strings.Trim(strings.TrimSpace(parts[1]), "\"")
			for i := 0; i < v.Len(); i++ {
				item := v.Index(i).Interface()
				itemVal := reflect.ValueOf(item)
				if itemVal.Kind() == reflect.Map {
					if mv, ok := item.(map[string]any); ok {
						if mv[key] == val {
							return item, nil
						}
					}
				} else if itemVal.Kind() == reflect.Struct {
					f := itemVal.FieldByName(key)
					if f.IsValid() && f.Kind() == reflect.String && f.String() == val {
						return item, nil
					}
				}
			}
			return nil, ErrFieldNotFound
		}
	case reflect.Map:
		// Map key filter: [key]
		mapKey := reflect.ValueOf(filter)
		mapValue := v.MapIndex(mapKey)
		if !mapValue.IsValid() {
			return nil, ErrFieldNotFound
		}
		return mapValue.Interface(), nil
	}
	return nil, ErrInvalidPath
}
