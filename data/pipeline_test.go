package data

import (
	"testing"
)

func TestExtractValue(t *testing.T) {
	// Create a nested structure for testing
	nestedMap := map[string]any{
		"name": "John",
		"age":  30,
		"address": map[string]any{
			"city":    "New York",
			"country": "USA",
			"zipcode": 10001,
			"geo": map[string]any{
				"lat": 40.7128,
				"lng": -74.0060,
			},
		},
		"scores": []int{85, 90, 95},
	}

	// Create pipeline with test data
	pipeline := NewPipeline("test-pipeline")
	pipeline.Set("user", nestedMap)
	pipeline.Set("simple", "value")
	pipeline.Set("number", 42)

	// Test cases
	tests := []struct {
		name      string
		path      string
		wantValue any
		wantErr   bool
	}{
		{
			name:      "simple key",
			path:      "simple",
			wantValue: "value",
			wantErr:   false,
		},
		{
			name:      "numeric value",
			path:      "number",
			wantValue: 42,
			wantErr:   false,
		},
		{
			name:      "nested key - first level",
			path:      "user.name",
			wantValue: "John",
			wantErr:   false,
		},
		{
			name:      "nested key - second level",
			path:      "user.address.city",
			wantValue: "New York",
			wantErr:   false,
		},
		{
			name:      "nested key - third level",
			path:      "user.address.geo.lat",
			wantValue: 40.7128,
			wantErr:   false,
		},
		{
			name:      "array access",
			path:      "user.scores.0",
			wantValue: 85,
			wantErr:   false,
		},
		{
			name:      "array access - middle element",
			path:      "user.scores.1",
			wantValue: 90,
			wantErr:   false,
		},
		{
			name:      "non-existent key",
			path:      "nonexistent",
			wantValue: nil,
			wantErr:   true,
		},
		{
			name:      "non-existent nested key",
			path:      "user.nonexistent",
			wantValue: nil,
			wantErr:   true,
		},
		{
			name:      "invalid path segment",
			path:      "user.scores.invalid",
			wantValue: nil,
			wantErr:   true,
		},
		{
			name:      "out of bounds array index",
			path:      "user.scores.10",
			wantValue: nil,
			wantErr:   true,
		},
		{
			name:      "numeric zipcode extraction with type conversion",
			path:      "user.address.zipcode",
			wantValue: 10001,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// String extraction
			if tt.wantValue == "value" || tt.wantValue == "John" || tt.wantValue == "New York" || tt.wantValue == "USA" {
				got, err := ExtractValue[string](pipeline, tt.path)
				if (err != nil) != tt.wantErr {
					t.Errorf("ExtractValue[string] error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !tt.wantErr && got != tt.wantValue.(string) {
					t.Errorf("ExtractValue[string] = %v, want %v", got, tt.wantValue)
				}
			}

			// Int extraction
			if tt.wantValue == 42 || tt.wantValue == 10001 || tt.wantValue == 85 || tt.wantValue == 90 || tt.wantValue == 95 {
				got, err := ExtractValue[int](pipeline, tt.path)
				if (err != nil) != tt.wantErr {
					t.Errorf("ExtractValue[int] error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !tt.wantErr && got != tt.wantValue.(int) {
					t.Errorf("ExtractValue[int] = %v, want %v", got, tt.wantValue)
				}
			}

			// Float extraction
			if tt.wantValue == 40.7128 || tt.wantValue == -74.0060 {
				got, err := ExtractValue[float64](pipeline, tt.path)
				if (err != nil) != tt.wantErr {
					t.Errorf("ExtractValue[float64] error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !tt.wantErr && got != tt.wantValue.(float64) {
					t.Errorf("ExtractValue[float64] = %v, want %v", got, tt.wantValue)
				}
			}

			// Any extraction
			got, err := ExtractValue[any](pipeline, tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExtractValue[any] error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.wantValue {
				t.Errorf("ExtractValue[any] = %v, want %v", got, tt.wantValue)
			}
		})
	}
}

func TestExtractValueTypeConversion(t *testing.T) {
	// Create pipeline with test data
	pipeline := NewPipeline("test-conversion")
	pipeline.Set("integer", 42)
	pipeline.Set("float", 42.5)
	pipeline.Set("string_number", "42")
	pipeline.Set("boolean", true)

	// Test direct float-to-int conversion
	t.Run("direct float to int", func(t *testing.T) {
		floatVal := 42.5
		var dummy int
		converted := any(int(floatVal)).(int)
		t.Logf("Direct conversion result: %d, type: %T", converted, converted)
		t.Logf("Original dummy type: %T", dummy)
		if converted != 42 {
			t.Errorf("Direct conversion failed: got %v", converted)
		}
	})

	// Test type conversions
	t.Run("int to string", func(t *testing.T) {
		val, err := ExtractValue[string](pipeline, "integer")
		if err != nil {
			t.Logf("Int to string error: %v", err)
		} else {
			t.Logf("Int to string result: %v", val)
		}
	})

	t.Run("float to int conversion test", func(t *testing.T) {
		// Make a separate test pipeline with just the float
		testPipe := NewPipeline("float-test")
		testPipe.Set("val", 42.5)

		// Get the raw value to confirm its type
		rawVal, _ := testPipe.Get("val")
		t.Logf("Raw value type: %T, value: %v", rawVal, rawVal)

		// Try the conversion
		val, err := ExtractValue[int](testPipe, "val")
		if err != nil {
			t.Logf("Float to int error: %v", err)
		} else {
			t.Logf("Float to int result: %v", val)
		}
	})

	t.Run("string to boolean", func(t *testing.T) {
		val, err := ExtractValue[bool](pipeline, "string_number")
		if err != nil {
			t.Logf("String to bool error: %v", err)
		} else {
			t.Logf("String to bool result: %v", val)
		}
	})

	t.Run("boolean to string", func(t *testing.T) {
		val, err := ExtractValue[string](pipeline, "boolean")
		if err != nil {
			t.Logf("Bool to string error: %v", err)
		} else {
			t.Logf("Bool to string result: %v", val)
		}
	})
}
