package data

import (
	"testing"
)

func TestConvert(t *testing.T) {
	// Test cases table
	tests := []struct {
		name     string
		input    any
		wantType string
		want     any
		wantErr  bool
	}{
		// String conversions
		{"int to string", 42, "string", "42", false},
		{"float to string", 42.5, "string", "42.5", false},
		{"bool to string true", true, "string", "true", false},
		{"bool to string false", false, "string", "false", false},

		// Int conversions
		{"string to int valid", "42", "int", 42, false},
		{"string to int invalid", "not-a-number", "int", 0, true},
		{"float to int", 42.7, "int", 42, false}, // Truncates
		{"bool to int true", true, "int", 1, false},
		{"bool to int false", false, "int", 0, false},

		// Float conversions
		{"string to float valid", "42.5", "float64", 42.5, false},
		{"string to float invalid", "not-a-number", "float64", 0.0, true},
		{"int to float", 42, "float64", 42.0, false},
		{"bool to float true", true, "float64", 1.0, false},
		{"bool to float false", false, "float64", 0.0, false},

		// Bool conversions
		{"string to bool true", "true", "bool", true, false},
		{"string to bool false", "false", "bool", false, false},
		{"string to bool 1", "1", "bool", true, false},
		{"string to bool 0", "0", "bool", false, false},
		{"int to bool non-zero", 42, "bool", true, false},
		{"int to bool zero", 0, "bool", false, false},
		{"float to bool non-zero", 42.5, "bool", true, false},
		{"float to bool zero", 0.0, "bool", false, false},

		// Edge cases
		{"nil to string", nil, "string", "", false},
		{"nil to int", nil, "int", 0, false},
		{"nil to bool", nil, "bool", false, false},

		// Direct type assertions
		{"string identity", "test", "string", "test", false},
		{"int identity", 42, "int", 42, false},
		{"bool identity", true, "bool", true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch tt.wantType {
			case "string":
				got, err := Convert[string](tt.input)
				if (err != nil) != tt.wantErr {
					t.Errorf("Convert() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !tt.wantErr && got != tt.want {
					t.Errorf("Convert() got = %v, want %v", got, tt.want)
				}
			case "int":
				got, err := Convert[int](tt.input)
				if (err != nil) != tt.wantErr {
					t.Errorf("Convert() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !tt.wantErr && got != tt.want {
					t.Errorf("Convert() got = %v, want %v", got, tt.want)
				}
			case "float64":
				got, err := Convert[float64](tt.input)
				if (err != nil) != tt.wantErr {
					t.Errorf("Convert() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !tt.wantErr && got != tt.want {
					t.Errorf("Convert() got = %v, want %v", got, tt.want)
				}
			case "bool":
				got, err := Convert[bool](tt.input)
				if (err != nil) != tt.wantErr {
					t.Errorf("Convert() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !tt.wantErr && got != tt.want {
					t.Errorf("Convert() got = %v, want %v", got, tt.want)
				}
			}
		})
	}

	// Additional tests for uint conversions
	t.Run("string to uint", func(t *testing.T) {
		got, err := Convert[uint](stringValue("42"))
		if err != nil {
			t.Errorf("Convert[uint] error = %v", err)
		} else if got != 42 {
			t.Errorf("Convert[uint] got = %v, want %v", got, 42)
		}
	})

	t.Run("negative int to uint should fail", func(t *testing.T) {
		_, err := Convert[uint](-42)
		if err == nil {
			t.Errorf("Convert[uint] should fail with negative int")
		}
	})

	// Test custom struct conversion - this should fail
	type Person struct {
		Name string
		Age  int
	}

	t.Run("struct conversion should fail", func(t *testing.T) {
		p := Person{Name: "John", Age: 30}
		_, err := Convert[map[string]interface{}](p)
		if err == nil {
			t.Errorf("Convert[map] should fail with struct input")
		}
	})
}

// Helper function to ensure types are consistent in tests
func stringValue(s string) any {
	return s
}
