package validator

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

/**
constraints conversion from string
*/

func convertInt(param string, bit int) (int64, error) {
	i, err := strconv.ParseInt(param, 0, bit)
	if err != nil {
		return 0, ErrConversionFailed
	}
	return i, nil
}

func convertUint(param string, bit int) (uint64, error) {
	i, err := strconv.ParseUint(param, 0, bit)
	if err != nil {
		return 0, ErrConversionFailed
	}
	return i, nil
}

func convertFloat(param string, bit int) (float64, error) {
	i, err := strconv.ParseFloat(param, bit)
	if err != nil {
		return 0, ErrConversionFailed
	}
	return i, nil
}

func convertBool(param string) (bool, error) {
	i, err := strconv.ParseBool(param)
	if err != nil {
		return false, ErrConversionFailed
	}
	return i, nil
}

func checkMin(field field, param string, isExclusive bool) error {
	val := field.value
	valid := true
	switch field.typ.Kind() {
	case reflect.Int:
		c, err := convertInt(param, 0)
		if err != nil {
			return fmt.Errorf(ErrBadConstraint, "checkMin", param, field.name)
		}
		cInt := int(c)
		in, _ := val.Interface().(int)
		if isExclusive {
			valid = in >= cInt
		} else {
			valid = in > cInt
		}
	case reflect.Int8:
		c, err := convertInt(param, 8)
		if err != nil {
			return fmt.Errorf(ErrBadConstraint, "checkMin", param, field.name)
		}
		cInt := int8(c)
		in, _ := val.Interface().(int8)
		if isExclusive {
			valid = in >= cInt
		} else {
			valid = in > cInt
		}
	case reflect.Int16:
		c, err := convertInt(param, 16)
		if err != nil {
			return fmt.Errorf(ErrBadConstraint, "checkMin", param, field.name)
		}
		cInt := int16(c)
		in, _ := val.Interface().(int16)
		if isExclusive {
			valid = in >= cInt
		} else {
			valid = in > cInt
		}
	case reflect.Int32:
		c, err := convertInt(param, 32)
		if err != nil {
			return fmt.Errorf(ErrBadConstraint, "checkMin", param, field.name)
		}
		cInt := int32(c)
		in, _ := val.Interface().(int32)
		if isExclusive {
			valid = in >= cInt
		} else {
			valid = in > cInt
		}
	case reflect.Int64:
		c, err := convertInt(param, 64)
		if err != nil {
			return fmt.Errorf(ErrBadConstraint, "checkMin", param, field.name)
		}
		in, _ := val.Interface().(int64)
		if isExclusive {
			valid = in >= c
		} else {
			valid = in > c
		}
	case reflect.Uint:
		c, err := convertUint(param, 0)
		if err != nil {
			return fmt.Errorf(ErrBadConstraint, "checkMin", param, field.name)
		}
		cUint := uint(c)
		in, _ := val.Interface().(uint)
		if isExclusive {
			valid = in >= cUint
		} else {
			valid = in > cUint
		}
	case reflect.Uint8:
		c, err := convertUint(param, 8)
		if err != nil {
			return fmt.Errorf(ErrBadConstraint, "checkMin", param, field.name)
		}
		cUint := uint8(c)
		in, _ := val.Interface().(uint8)
		if isExclusive {
			valid = in >= cUint
		} else {
			valid = in > cUint
		}
	case reflect.Uint16:
		c, err := convertUint(param, 16)
		if err != nil {
			return fmt.Errorf(ErrBadConstraint, "checkMin", param, field.name)
		}
		cUint := uint16(c)
		in, _ := val.Interface().(uint16)
		if isExclusive {
			valid = in >= cUint
		} else {
			valid = in > cUint
		}
	case reflect.Uint32:
		c, err := convertUint(param, 32)
		if err != nil {
			return fmt.Errorf(ErrBadConstraint, "checkMin", param, field.name)
		}
		cUint := uint32(c)
		in, _ := val.Interface().(uint32)
		if isExclusive {
			valid = in >= cUint
		} else {
			valid = in > cUint
		}
	case reflect.Uint64:
		c, err := convertUint(param, 64)
		if err != nil {
			return fmt.Errorf(ErrBadConstraint, "checkMin", param, field.name)
		}
		in, _ := val.Interface().(uint64)
		if isExclusive {
			valid = in >= c
		} else {
			valid = in > c
		}
	case reflect.Uintptr:
		/*c, err := convertUint(param)
		if err != nil {
			return err
		}
		valid = input.Uint() < c*/
		valid = true
	case reflect.Float32:
		c, err := convertFloat(param, 32)
		if err != nil {
			return fmt.Errorf(ErrBadConstraint, "checkMax", param, field.name)
		}
		cFloat := float32(c)
		in, _ := val.Interface().(float32)
		if isExclusive {
			valid = in >= cFloat
		} else {
			valid = in > cFloat
		}
	case reflect.Float64:
		c, err := convertFloat(param, 64)
		if err != nil {
			return fmt.Errorf(ErrBadConstraint, "checkMax", param, field.name)
		}
		cFloat := c
		in, _ := val.Interface().(float64)
		if isExclusive {
			valid = in >= cFloat
		} else {
			valid = in > cFloat
		}
	default:
		return fmt.Errorf(ErrInvalidValidationForField, field.name)
	}
	if !valid {
		if isExclusive {
			return fmt.Errorf(ErrExclusiveMin, field.name)
		} else {
			return fmt.Errorf(ErrMin, field.name)
		}
	}
	return nil
}

func checkMax(field field, param string, isExclusive bool) error {
	valid := true
	val := field.value
	switch field.typ.Kind() {
	case reflect.Int:
		c, err := convertInt(param, 0)
		if err != nil {
			return fmt.Errorf(ErrBadConstraint, "checkMax", param, field.name)
		}
		cInt := int(c)
		in, _ := val.Interface().(int)
		if isExclusive {
			valid = in <= cInt
		} else {
			valid = in < cInt
		}
	case reflect.Int8:
		c, err := convertInt(param, 8)
		if err != nil {
			return fmt.Errorf(ErrBadConstraint, "checkMax", param, field.name)
		}
		cInt := int8(c)
		in, _ := val.Interface().(int8)
		if isExclusive {
			valid = in <= cInt
		} else {
			valid = in < cInt
		}
	case reflect.Int16:
		c, err := convertInt(param, 16)
		if err != nil {
			return fmt.Errorf(ErrBadConstraint, "checkMax", param, field.name)
		}
		cInt := int16(c)
		in, _ := val.Interface().(int16)
		if isExclusive {
			valid = in <= cInt
		} else {
			valid = in < cInt
		}
	case reflect.Int32:
		c, err := convertInt(param, 32)
		if err != nil {
			return fmt.Errorf(ErrBadConstraint, "checkMax", param, field.name)
		}
		cInt := int32(c)
		in, _ := val.Interface().(int32)
		if isExclusive {
			valid = in <= cInt
		} else {
			valid = in < cInt
		}
	case reflect.Int64:
		c, err := convertInt(param, 64)
		if err != nil {
			return fmt.Errorf(ErrBadConstraint, "checkMax", param, field.name)
		}
		in, _ := val.Interface().(int64)
		if isExclusive {
			valid = in <= c
		} else {
			valid = in < c
		}
	case reflect.Uint:
		c, err := convertUint(param, 0)
		if err != nil {
			return fmt.Errorf(ErrBadConstraint, "checkMax", param, field.name)
		}
		cUint := uint(c)
		in, _ := val.Interface().(uint)
		if isExclusive {
			valid = in <= cUint
		} else {
			valid = in < cUint
		}
	case reflect.Uint8:
		c, err := convertUint(param, 8)
		if err != nil {
			return fmt.Errorf(ErrBadConstraint, "checkMax", param, field.name)
		}
		cUint := uint8(c)
		in, _ := val.Interface().(uint8)
		if isExclusive {
			valid = in <= cUint
		} else {
			valid = in < cUint
		}
	case reflect.Uint16:
		c, err := convertUint(param, 16)
		if err != nil {
			return fmt.Errorf(ErrBadConstraint, "checkMax", param, field.name)
		}
		cUint := uint16(c)
		in, _ := val.Interface().(uint16)
		if isExclusive {
			valid = in <= cUint
		} else {
			valid = in < cUint
		}
	case reflect.Uint32:
		c, err := convertUint(param, 32)
		if err != nil {
			return fmt.Errorf(ErrBadConstraint, "checkMax", param, field.name)
		}
		cUint := uint32(c)
		in, _ := val.Interface().(uint32)
		if isExclusive {
			valid = in <= cUint
		} else {
			valid = in < cUint
		}
	case reflect.Uint64:
		c, err := convertUint(param, 64)
		if err != nil {
			return fmt.Errorf(ErrBadConstraint, "checkMax", param, field.name)
		}
		in, _ := val.Interface().(uint64)
		if isExclusive {
			valid = in <= c
		} else {
			valid = in < c
		}
	case reflect.Uintptr:
		/*c, err := convertUint(param)
		if err != nil {
			return err
		}
		valid = input.Uint() < c*/
		valid = true
	case reflect.Float32:
		c, err := convertFloat(param, 32)
		if err != nil {
			return fmt.Errorf(ErrBadConstraint, "checkMax", param, field.name)
		}
		cFloat := float32(c)
		in, _ := val.Interface().(float32)
		if isExclusive {
			valid = in <= cFloat
		} else {
			valid = in < cFloat
		}
	case reflect.Float64:
		c, err := convertFloat(param, 64)
		if err != nil {
			return fmt.Errorf(ErrBadConstraint, "checkMax", param, field.name)
		}
		cFloat := c
		in, _ := val.Interface().(float64)
		if isExclusive {
			valid = in <= cFloat
		} else {
			valid = in < cFloat
		}
	default:
		return fmt.Errorf(ErrInvalidValidationForField, field.name)
	}
	if !valid {
		if isExclusive {
			return fmt.Errorf(ErrExclusiveMax, field.name)
		} else {
			return fmt.Errorf(ErrMax, field.name)
		}
	}
	return nil
}

func checkIfEnumExists(val string, param string, separator string) bool {
	params := strings.Split(param, separator)
	for _, en := range params {
		if val == en {
			return true
		}
	}
	return false
}
