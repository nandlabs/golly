// TestGenerateSchema tests the GenerateSchema function with various input types.
// It verifies that the function correctly generates JSON schemas for different Go types:
// - primitive types (string, integer, float, boolean)
// - composite types (slices, maps, structs)
// - pointers to structs
// - structs with custom JSON tags
// - error handling for unsupported types (channels)
//
// Each test case validates that the generated schema matches the expected schema structure,
// including correct type identification, items definitions for arrays, additional items for maps,
// and property mappings for struct fields with proper JSON tag handling.
package data

import (
	"reflect"
	"testing"
)

func TestGenerateSchema(t *testing.T) {
	tests := []struct {
		name        string
		input       interface{}
		expected    *Schema
		expectError bool
	}{
		{
			name:  "string type",
			input: "",
			expected: &Schema{
				Type: "string",
			},
		},
		{
			name:  "integer type",
			input: 0,
			expected: &Schema{
				Type: "integer",
			},
		},
		{
			name:  "float type",
			input: 0.0,
			expected: &Schema{
				Type: "number",
			},
		},
		{
			name:  "boolean type",
			input: false,
			expected: &Schema{
				Type: "boolean",
			},
		},
		{
			name:  "slice type",
			input: []string{},
			expected: &Schema{
				Type: "array",
				Items: &Schema{
					Type: "string",
				},
			},
		},
		{
			name:  "map type",
			input: map[string]int{},
			expected: &Schema{
				Type:       "object",
				Properties: map[string]*Schema{},
				AdditionalItems: &Schema{
					Type: "integer",
				},
			},
		},
		{
			name: "struct type",
			input: struct {
				Name string `json:"name"`
				Age  int    `json:"age"`
			}{},
			expected: &Schema{
				Type: "object",
				Properties: map[string]*Schema{
					"name": {
						Type: "string",
					},
					"age": {
						Type: "integer",
					},
				},
			},
		},
		{
			name: "struct with pointer",
			input: &struct {
				Name string `json:"name"`
			}{},
			expected: &Schema{
				Type: "object",
				Properties: map[string]*Schema{
					"name": {
						Type: "string",
					},
				},
			},
		},
		{
			name: "struct with custom json tag",
			input: struct {
				UserName string `json:"user_name"`
				IsActive bool   `json:"is_active,omitempty"`
			}{},
			expected: &Schema{
				Type: "object",
				Properties: map[string]*Schema{
					"user_name": {
						Type: "string",
					},
					"is_active": {
						Type: "boolean",
					},
				},
			},
		},
		{
			name:        "unsupported type",
			input:       make(chan int),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema, err := GenerateSchema(reflect.TypeOf(tt.input))

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if schema == nil {
				t.Errorf("schema is nil")
				return
			}

			if schema.Type != tt.expected.Type {
				t.Errorf("expected type %s but got %s", tt.expected.Type, schema.Type)
			}

			if tt.expected.Items != nil {
				if schema.Items == nil {
					t.Errorf("expected Items but got nil")
				} else if schema.Items.Type != tt.expected.Items.Type {
					t.Errorf("expected Items.Type %s but got %s", tt.expected.Items.Type, schema.Items.Type)
				}
			}

			if tt.expected.AdditionalItems != nil {
				if schema.AdditionalItems == nil {
					t.Errorf("expected AdditionalItems but got nil")
				} else if schema.AdditionalItems.Type != tt.expected.AdditionalItems.Type {
					t.Errorf("expected AdditionalItems.Type %s but got %s", tt.expected.AdditionalItems.Type, schema.AdditionalItems.Type)
				}
			}

			if tt.expected.Properties != nil {
				if schema.Properties == nil {
					t.Errorf("expected Properties but got nil")
				} else if len(schema.Properties) != len(tt.expected.Properties) {
					t.Errorf("expected Properties length %d but got %d", len(tt.expected.Properties), len(schema.Properties))
				} else {
					for k, v := range tt.expected.Properties {
						prop, ok := schema.Properties[k]
						if !ok {
							t.Errorf("expected property %s but not found", k)
						} else if prop.Type != v.Type {
							t.Errorf("expected property %s type %s but got %s", k, v.Type, prop.Type)
						}
					}
				}
			}
		})
	}
}
