package data

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
)

var (
	// ErrUnsupportedConversion is returned when a conversion between types is not supported
	ErrUnsupportedConversion = errors.New("unsupported type conversion")
)

// Convert converts the input of any type to the specified output type T.
// It returns the converted value and an error if the conversion fails.
func Convert[T any](in any) (out T, err error) {
	// If input is nil, return zero value of T
	if in == nil {
		return out, nil
	}

	// Get the type of the output parameter
	outType := reflect.TypeOf((*T)(nil)).Elem()

	// If input is already of the required output type, return it directly
	if reflect.TypeOf(in) == outType {
		return in.(T), nil
	}

	// Handle type assertion first
	if val, ok := in.(T); ok {
		return val, nil
	}

	// Get reflect values
	inValue := reflect.ValueOf(in)
	outValue := reflect.New(outType).Elem()

	// Handle conversion based on destination type
	switch outType.Kind() {
	case reflect.String:
		// Convert to string
		switch inValue.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			outValue.SetString(strconv.FormatInt(inValue.Int(), 10))
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			outValue.SetString(strconv.FormatUint(inValue.Uint(), 10))
		case reflect.Float32, reflect.Float64:
			outValue.SetString(strconv.FormatFloat(inValue.Float(), 'f', -1, 64))
		case reflect.Bool:
			outValue.SetString(strconv.FormatBool(inValue.Bool()))
		default:
			return out, fmt.Errorf("%w: cannot convert %T to string", ErrUnsupportedConversion, in)
		}

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		// Convert to int
		switch inValue.Kind() {
		case reflect.String:
			i, err := strconv.ParseInt(inValue.String(), 10, 64)
			if err != nil {
				return out, fmt.Errorf("failed to convert string to int: %w", err)
			}
			if outType.Kind() == reflect.Int64 {
				outValue.SetInt(i)
			} else {
				// Check if the int64 value fits into the target int type
				outValue.SetInt(i)
			}
		case reflect.Float32, reflect.Float64:
			outValue.SetInt(int64(inValue.Float()))
		case reflect.Bool:
			if inValue.Bool() {
				outValue.SetInt(1)
			} else {
				outValue.SetInt(0)
			}
		default:
			return out, fmt.Errorf("%w: cannot convert %T to int", ErrUnsupportedConversion, in)
		}

	case reflect.Float32, reflect.Float64:
		// Convert to float
		switch inValue.Kind() {
		case reflect.String:
			f, err := strconv.ParseFloat(inValue.String(), 64)
			if err != nil {
				return out, fmt.Errorf("failed to convert string to float: %w", err)
			}
			outValue.SetFloat(f)
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			outValue.SetFloat(float64(inValue.Int()))
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			outValue.SetFloat(float64(inValue.Uint()))
		case reflect.Bool:
			if inValue.Bool() {
				outValue.SetFloat(1.0)
			} else {
				outValue.SetFloat(0.0)
			}
		default:
			return out, fmt.Errorf("%w: cannot convert %T to float", ErrUnsupportedConversion, in)
		}

	case reflect.Bool:
		// Convert to bool
		switch inValue.Kind() {
		case reflect.String:
			b, err := strconv.ParseBool(inValue.String())
			if err != nil {
				// Try numeric string conversion
				if i, err := strconv.ParseInt(inValue.String(), 10, 64); err == nil {
					outValue.SetBool(i != 0)
				} else {
					return out, fmt.Errorf("failed to convert string to bool: %w", err)
				}
			} else {
				outValue.SetBool(b)
			}
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			outValue.SetBool(inValue.Int() != 0)
		case reflect.Float32, reflect.Float64:
			outValue.SetBool(inValue.Float() != 0)
		default:
			return out, fmt.Errorf("%w: cannot convert %T to bool", ErrUnsupportedConversion, in)
		}

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		// Convert to uint
		switch inValue.Kind() {
		case reflect.String:
			u, err := strconv.ParseUint(inValue.String(), 10, 64)
			if err != nil {
				return out, fmt.Errorf("failed to convert string to uint: %w", err)
			}
			outValue.SetUint(u)
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if inValue.Int() < 0 {
				return out, fmt.Errorf("cannot convert negative int to uint")
			}
			outValue.SetUint(uint64(inValue.Int()))
		case reflect.Float32, reflect.Float64:
			if inValue.Float() < 0 {
				return out, fmt.Errorf("cannot convert negative float to uint")
			}
			outValue.SetUint(uint64(inValue.Float()))
		case reflect.Bool:
			if inValue.Bool() {
				outValue.SetUint(1)
			} else {
				outValue.SetUint(0)
			}
		default:
			return out, fmt.Errorf("%w: cannot convert %T to uint", ErrUnsupportedConversion, in)
		}

	case reflect.Slice:
		// Handle slice conversions if needed
		return out, fmt.Errorf("%w: slice conversion not supported yet", ErrUnsupportedConversion)

	case reflect.Map:
		// Handle map conversions if needed
		return out, fmt.Errorf("%w: map conversion not supported yet", ErrUnsupportedConversion)

	default:
		return out, fmt.Errorf("%w: cannot convert %T to %s", ErrUnsupportedConversion, in, outType.Kind())
	}

	// Get the converted value as interface{} and then cast to T
	if outValue.CanInterface() {
		result := outValue.Interface()
		if res, ok := result.(T); ok {
			return res, nil
		}
	}

	return out, fmt.Errorf("%w: conversion resulted in incompatible type", ErrUnsupportedConversion)
}
