package validator

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
)

/**
Numerical Type Constraints
*/

func min(field field, param string) error {
	return checkMin(field, param, false)
}

func max(field field, param string) error {
	return checkMax(field, param, false)
}

func exclusiveMin(field field, param string) error {
	return checkMin(field, param, true)
}

func exclusiveMax(field field, param string) error {
	return checkMax(field, param, true)
}

func multipleOf(field field, param string) error {
	// TODO : works only for int as of now
	switch field.typ.Kind() {
	case reflect.Int:
		in, _ := field.value.Interface().(int)
		c, err := convertInt(param, 0)
		cInt := int(c)
		if err != nil {
			return err
		}
		valid := in%cInt == 0
		if !valid {
			return fmt.Errorf(ErrMultipleOf, field.name)
		}
	default:
		return fmt.Errorf(ErrInvalidValidationForField, field.name)
	}
	return nil
}

/**
String Type Constraints
*/

func notnull(field field, param string) error {
	switch field.typ.Kind() {
	case reflect.String:
		c, err := convertBool(param)
		if err != nil {
			return fmt.Errorf(ErrBadConstraint, "notnull", param, field.name)
		}
		if c == true {
			in, _ := field.value.Interface().(string)
			if in == "" {
				return fmt.Errorf(ErrNotNull, field.name)
			}
		}
	default:
		return fmt.Errorf(ErrInvalidValidationForField, field.name)
	}
	return nil
}

func minLength(field field, param string) error {
	switch field.typ.Kind() {
	case reflect.String:
		lc, _ := strconv.Atoi(param)
		lv := len(fmt.Sprint(field.value))
		valid := lv > lc
		if !valid {
			return fmt.Errorf(ErrMinLength, field.name)
		}
	default:
		return fmt.Errorf(ErrInvalidValidationForField, field.name)
	}
	return nil
}

func maxLength(field field, param string) error {
	switch field.typ.Kind() {
	case reflect.String:
		lc, _ := strconv.Atoi(param)
		lv := len(fmt.Sprint(field.value))
		valid := lv < lc
		if !valid {
			return fmt.Errorf(ErrMaxLength, field.name)
		}
	default:
		return fmt.Errorf(ErrInvalidValidationForField, field.name)
	}
	return nil
}

func pattern(field field, param string) error {
	switch field.typ.Kind() {
	case reflect.String:
		in, _ := field.value.Interface().(string)
		re, err := regexp.Compile(param)
		if err != nil {
			return fmt.Errorf(ErrBadConstraint, "pattern", param, field.name)
		}
		if !re.MatchString(in) {
			return fmt.Errorf(ErrPattern, field.name)
		}
	default:
		return fmt.Errorf(ErrInvalidValidationForField, field.name)
	}
	return nil
}

func enum(field field, param string) error {
	flag := false
	switch field.value.Kind() {
	case reflect.Int:
		input := field.value.Interface().(int)
		flag = checkIfEnumExists(strconv.Itoa(input), param, ",")
	case reflect.String:
		input := field.value.String()
		flag = checkIfEnumExists(input, param, ",")
	}

	if flag == false {
		return fmt.Errorf(ErrEnums, field.name)
	}
	return nil
}
