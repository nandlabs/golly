package data

import (
	"errors"
	"reflect"
	"strings"
)

// ErrUnsupportedType is returned when the type is not supported
var ErrUnsupportedType = errors.New("unsupported type")

// GenerateSchema generates schema of the given type
func GenerateSchema(t reflect.Type) (schema *Schema, err error) {
	switch t.Kind() {
	case reflect.Struct:
		schema = &Schema{
			Type:       "object",
			Properties: make(map[string]*Schema),
		}
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			if field.PkgPath != "" {
				continue
			}
			prop, err := GenerateSchema(field.Type)
			if err != nil {
				return nil, err
			}
			fieldName := field.Name
			if tag, ok := field.Tag.Lookup("json"); ok {

				// check and Split the tag by comma and take the first part as the field name
				parts := strings.Split(tag, ",")
				if parts[0] != "" {
					fieldName = parts[0]
				} else {
					fieldName = tag
				}

			}
			schema.Properties[fieldName] = prop
		}
	case reflect.Slice:
		schema = &Schema{
			Type: "array",
		}
		elemSchema, err := GenerateSchema(t.Elem())
		if err != nil {
			return nil, err
		}
		schema.Items = elemSchema
	case reflect.Map:
		schema = &Schema{
			Type:       "object",
			Properties: make(map[string]*Schema),
		}
		elemSchema, err := GenerateSchema(t.Elem())
		if err != nil {
			return nil, err
		}
		schema.AdditionalItems = elemSchema

	case reflect.String:
		schema = &Schema{
			Type: "string",
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		schema = &Schema{
			Type: "integer",
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		schema = &Schema{
			Type: "integer",
		}
	case reflect.Float32, reflect.Float64:
		schema = &Schema{
			Type: "number",
		}
	case reflect.Bool:
		schema = &Schema{
			Type: "boolean",
		}

	default:
		err = ErrUnsupportedType
	}

	return
}
