package validator

import (
	"encoding/json"
	"testing"

	"oss.nandlabs.io/golly/data"
)

// FuzzCompileSchema asserts CompileSchema never panics when given arbitrary
// JSON-decoded into a data.Schema. Malformed input must error rather than
// crash.
func FuzzCompileSchema(f *testing.F) {
	seeds := []string{
		`{}`,
		`{"type":"string"}`,
		`{"type":"object","properties":{"name":{"type":"string"}},"required":["name"]}`,
		`{"type":"object","properties":{"a":{"type":"integer","minimum":0,"maximum":10}}}`,
		`{"type":"array","items":{"type":"string"},"minItems":1}`,
		`{"type":"string","pattern":"^[a-z]+$"}`,
		`{"type":"string","pattern":"["}`, // invalid regex
		`{"type":"object","properties":{"x":{"$ref":"#/definitions/x"}}}`,
		`{"enum":[1,2,3]}`,
		`{"oneOf":[{"type":"string"},{"type":"integer"}]}`,
		`{"anyOf":[{"type":"string"}]}`,
		`{"allOf":[{"type":"object"}]}`,
		`{"type":"string","minLength":-1}`,
		`{"type":"object","properties":{}}`,
		`{"type":"object","minProperties":0,"maxProperties":1000}`,
	}
	for _, s := range seeds {
		f.Add([]byte(s))
	}
	f.Fuzz(func(t *testing.T, raw []byte) {
		var s data.Schema
		if err := json.Unmarshal(raw, &s); err != nil {
			return // not a valid schema doc — fine
		}
		_, _ = CompileSchema(&s) // must not panic; error is fine
	})
}

// FuzzValidate asserts Validate never panics on arbitrary JSON-decoded data
// against a fixed, moderately complex schema.
func FuzzValidate(f *testing.F) {
	schema := &data.Schema{
		Type: "object",
		Properties: map[string]*data.Schema{
			"name": {Type: "string", MinLength: ptrInt(1), MaxLength: ptrInt(64)},
			"age":  {Type: "integer", Minimum: ptrFloat(0), Maximum: ptrFloat(150)},
			"tags": {Type: "array", Items: &data.Schema{Type: "string"}, MaxItems: ptrInt(10)},
			"role": {Enum: []any{"admin", "user", "guest"}},
		},
		Required: []string{"name"},
	}
	v, err := CompileSchema(schema)
	if err != nil {
		f.Fatalf("seed schema failed to compile: %v", err)
	}

	seeds := []string{
		`{"name":"Alice"}`,
		`{"name":"Bob","age":30,"role":"admin"}`,
		`{"name":""}`,
		`{"age":"not-a-number"}`,
		`null`,
		`[]`,
		`"plain string"`,
		`{"name":"X","tags":["a","b","c"]}`,
		`{"name":"X","tags":1}`,
		`{"name":"X","role":"bad-role"}`,
	}
	for _, s := range seeds {
		f.Add([]byte(s))
	}
	f.Fuzz(func(t *testing.T, raw []byte) {
		var value any
		if err := json.Unmarshal(raw, &value); err != nil {
			return // not valid JSON — fine
		}
		_ = v.Validate(value) // success or error are both fine; must not panic
	})
}

func ptrInt(i int) *int           { return &i }
func ptrFloat(f float64) *float64 { return &f }
