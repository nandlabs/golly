package validator

import (
	"errors"
	"strings"
	"testing"

	"oss.nandlabs.io/golly/data"
)

func ptr[T any](v T) *T { return &v }

// ---- compile ----

func TestCompile_NilSchema(t *testing.T) {
	if _, err := CompileSchema(nil); err == nil {
		t.Fatal("expected error compiling nil schema")
	}
}

func TestCompile_InvalidPattern(t *testing.T) {
	_, err := CompileSchema(&data.Schema{
		Type:    data.SchemaTypeString,
		Pattern: ptr("([unbalanced"),
	})
	if err == nil {
		t.Fatal("expected compile error for invalid regex")
	}
}

// ---- type checking ----

func TestType_Mismatch(t *testing.T) {
	v, _ := CompileSchema(&data.Schema{Type: data.SchemaTypeString})
	err := v.Validate(42)
	if err == nil {
		t.Fatal("expected type error")
	}
	if !strings.Contains(err.Error(), "type") {
		t.Errorf("expected error to mention type, got %v", err)
	}
}

func TestType_Integer_RejectsFractional(t *testing.T) {
	v, _ := CompileSchema(&data.Schema{Type: data.SchemaTypeInteger})
	if err := v.Validate(1.5); err == nil {
		t.Fatal("expected fractional float to fail integer type")
	}
	if err := v.Validate(2.0); err != nil {
		t.Errorf("2.0 should be valid integer; got %v", err)
	}
}

func TestType_Nullable(t *testing.T) {
	v, _ := CompileSchema(&data.Schema{Type: data.SchemaTypeString, Nullable: true})
	if err := v.Validate(nil); err != nil {
		t.Errorf("nullable schema should accept nil; got %v", err)
	}
	v2, _ := CompileSchema(&data.Schema{Type: data.SchemaTypeString})
	if err := v2.Validate(nil); err == nil {
		t.Error("non-nullable schema should reject nil")
	}
}

// ---- string ----

func TestString_LengthAndPattern(t *testing.T) {
	v, _ := CompileSchema(&data.Schema{
		Type:      data.SchemaTypeString,
		MinLength: ptr(3),
		MaxLength: ptr(5),
		Pattern:   ptr("^[a-z]+$"),
	})
	if err := v.Validate("ab"); err == nil {
		t.Error("expected minLength failure for 'ab'")
	}
	if err := v.Validate("abcdef"); err == nil {
		t.Error("expected maxLength failure for 'abcdef'")
	}
	if err := v.Validate("Abc"); err == nil {
		t.Error("expected pattern failure for 'Abc'")
	}
	if err := v.Validate("abc"); err != nil {
		t.Errorf("'abc' should pass; got %v", err)
	}
}

func TestString_Formats(t *testing.T) {
	cases := []struct {
		format, ok, bad string
	}{
		{"email", "alice@example.com", "not-an-email"},
		{"uuid", "00000000-0000-0000-0000-000000000000", "not-a-uuid"},
		{"date-time", "2024-01-02T03:04:05Z", "2024/01/02"},
		{"date", "2024-01-02", "not-a-date"},
		{"uri", "https://example.com/x", "no scheme"},
	}
	for _, c := range cases {
		t.Run(c.format, func(t *testing.T) {
			v, _ := CompileSchema(&data.Schema{Type: data.SchemaTypeString, Format: ptr(c.format)})
			if err := v.Validate(c.ok); err != nil {
				t.Errorf("expected %q (%s) to pass; got %v", c.ok, c.format, err)
			}
			if err := v.Validate(c.bad); err == nil {
				t.Errorf("expected %q (%s) to fail", c.bad, c.format)
			}
		})
	}
}

// ---- number ----

func TestNumber_BoundsAndMultiple(t *testing.T) {
	v, _ := CompileSchema(&data.Schema{
		Type:             data.SchemaTypeNumber,
		Minimum:          ptr(0.0),
		ExclusiveMaximum: ptr(10.0),
		MultipleOf:       ptr(2.5),
	})
	cases := []struct {
		in any
		ok bool
	}{
		{-0.1, false}, // < min
		{0.0, true},   // == min OK
		{2.5, true},
		{7.5, true},
		{10.0, false}, // == exclusiveMax fails
		{2.6, false},  // not multipleOf
	}
	for _, c := range cases {
		err := v.Validate(c.in)
		if (err == nil) != c.ok {
			t.Errorf("Validate(%v): err=%v, expected ok=%v", c.in, err, c.ok)
		}
	}
}

// ---- enum ----

func TestEnum(t *testing.T) {
	v, _ := CompileSchema(&data.Schema{Enum: []any{"red", "green", 7}})
	if err := v.Validate("red"); err != nil {
		t.Errorf("'red' should be in enum; got %v", err)
	}
	if err := v.Validate(7); err != nil {
		t.Errorf("int 7 should be in enum; got %v", err)
	}
	if err := v.Validate(7.0); err != nil {
		t.Errorf("float 7.0 should equal int 7 (numeric normalisation); got %v", err)
	}
	if err := v.Validate("blue"); err == nil {
		t.Error("'blue' should not be in enum")
	}
}

// ---- array ----

func TestArray_ItemsAndConstraints(t *testing.T) {
	v, _ := CompileSchema(&data.Schema{
		Type:        data.SchemaTypeArray,
		Items:       &data.Schema{Type: data.SchemaTypeString},
		MinItems:    ptr(1),
		MaxItems:    ptr(3),
		UniqueItems: true,
	})
	if err := v.Validate([]any{}); err == nil {
		t.Error("empty array should fail minItems")
	}
	if err := v.Validate([]any{"a", "b", "c", "d"}); err == nil {
		t.Error("4 items should fail maxItems")
	}
	if err := v.Validate([]any{"a", "a"}); err == nil {
		t.Error("duplicate items should fail uniqueItems")
	}
	if err := v.Validate([]any{"a", 1}); err == nil {
		t.Error("non-string item should fail items.type")
	}
	if err := v.Validate([]any{"a", "b"}); err != nil {
		t.Errorf("['a','b'] should pass; got %v", err)
	}
}

// ---- object ----

func TestObject_RequiredAndProperties(t *testing.T) {
	v, _ := CompileSchema(&data.Schema{
		Type: data.SchemaTypeObject,
		Properties: map[string]*data.Schema{
			"name": {Type: data.SchemaTypeString},
			"age":  {Type: data.SchemaTypeInteger, Minimum: ptr(0.0)},
		},
		Required: []string{"name"},
	})

	if err := v.Validate(map[string]any{"age": 30}); err == nil {
		t.Error("missing required 'name' should fail")
	}
	if err := v.Validate(map[string]any{"name": "alice", "age": -1}); err == nil {
		t.Error("negative age should fail minimum on nested property")
	}
	if err := v.Validate(map[string]any{"name": "alice"}); err != nil {
		t.Errorf("only required 'name' should pass; got %v", err)
	}
}

func TestObject_ErrorPathIsJSONPointer(t *testing.T) {
	v, _ := CompileSchema(&data.Schema{
		Type: data.SchemaTypeObject,
		Properties: map[string]*data.Schema{
			"user": {
				Type: data.SchemaTypeObject,
				Properties: map[string]*data.Schema{
					"email": {Type: data.SchemaTypeString, Format: ptr("email")},
				},
			},
		},
	})
	err := v.Validate(map[string]any{
		"user": map[string]any{"email": "bad"},
	})
	if err == nil {
		t.Fatal("expected error for bad email")
	}
	if !strings.Contains(err.Error(), "/user/email") {
		t.Errorf("expected /user/email in error path; got %v", err)
	}
}

// ---- composition ----

func TestAnyOf(t *testing.T) {
	v, _ := CompileSchema(&data.Schema{
		AnyOf: []*data.Schema{
			{Type: data.SchemaTypeString},
			{Type: data.SchemaTypeInteger},
		},
	})
	for _, ok := range []any{"x", 5} {
		if err := v.Validate(ok); err != nil {
			t.Errorf("%v should match anyOf; got %v", ok, err)
		}
	}
	if err := v.Validate(true); err == nil {
		t.Error("bool should fail anyOf{string,integer}")
	}
}

func TestOneOf_ExactlyOneMatches(t *testing.T) {
	v, _ := CompileSchema(&data.Schema{
		OneOf: []*data.Schema{
			{Type: data.SchemaTypeInteger, Maximum: ptr(10.0)}, // matches 5
			{Type: data.SchemaTypeInteger, Minimum: ptr(0.0)},  // also matches 5 → not exactly one
		},
	})
	if err := v.Validate(5); err == nil {
		t.Error("5 matches both subschemas; oneOf should fail")
	}
}

func TestAllOf(t *testing.T) {
	v, _ := CompileSchema(&data.Schema{
		AllOf: []*data.Schema{
			{Type: data.SchemaTypeString, MinLength: ptr(3)},
			{Type: data.SchemaTypeString, Pattern: ptr("^[a-z]+$")},
		},
	})
	if err := v.Validate("ab"); err == nil {
		t.Error("'ab' fails minLength via allOf")
	}
	if err := v.Validate("Abc"); err == nil {
		t.Error("'Abc' fails pattern via allOf")
	}
	if err := v.Validate("abc"); err != nil {
		t.Errorf("'abc' should pass allOf; got %v", err)
	}
}

func TestNot(t *testing.T) {
	v, _ := CompileSchema(&data.Schema{
		Not: &data.Schema{Type: data.SchemaTypeString},
	})
	if err := v.Validate("x"); err == nil {
		t.Error("string should fail Not(string)")
	}
	if err := v.Validate(42); err != nil {
		t.Errorf("non-string should pass Not(string); got %v", err)
	}
}

// ---- multi-error ----

func TestMultiError_JoinsAndExposesLeaves(t *testing.T) {
	v, _ := CompileSchema(&data.Schema{
		Type: data.SchemaTypeObject,
		Properties: map[string]*data.Schema{
			"a": {Type: data.SchemaTypeString, MinLength: ptr(5)},
			"b": {Type: data.SchemaTypeInteger, Minimum: ptr(10.0)},
		},
		Required: []string{"a", "b"},
	})
	err := v.Validate(map[string]any{"a": "x", "b": 1})
	if err == nil {
		t.Fatal("expected multi-error")
	}
	// Unwrap the joined errors.
	type unwrapMany interface{ Unwrap() []error }
	uw, ok := err.(unwrapMany)
	if !ok {
		t.Fatalf("expected errors.Join multi-error; got %T", err)
	}
	leaves := uw.Unwrap()
	if len(leaves) < 2 {
		t.Errorf("expected ≥2 leaf errors; got %d", len(leaves))
	}
	for _, e := range leaves {
		var se *SchemaError
		if !errors.As(e, &se) {
			t.Errorf("leaf %v is not *SchemaError", e)
		}
	}
}

// ---- JSON pointer escaping ----

func TestEscapePointer(t *testing.T) {
	cases := []struct{ in, want string }{
		{"a", "a"},
		{"a/b", "a~1b"},
		{"a~b", "a~0b"},
		{"a~b/c", "a~0b~1c"},
	}
	for _, c := range cases {
		if got := escapePointer(c.in); got != c.want {
			t.Errorf("escapePointer(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}
