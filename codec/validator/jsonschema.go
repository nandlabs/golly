// Package validator's JSON-Schema validator validates arbitrary decoded data
// (maps, slices, scalars — typically the result of json.Unmarshal into any)
// against a *data.Schema. It is the companion to data.GenerateSchema, which
// produces schemas from struct types.
//
// Usage:
//
//	schema, _ := data.GenerateSchema(reflect.TypeOf(MyType{}))
//	v, err := validator.CompileSchema(schema)
//	if err != nil { return err }
//	if err := v.Validate(decodedJSON); err != nil { return err }
//
// The validator is intentionally scoped to the subset of JSON Schema needed
// for tool-call / API payload validation (Draft 2020-12 subset):
//   - type (object, array, string, number, integer, boolean, null)
//   - object: properties, required, minProperties, maxProperties
//   - array:  items, minItems, maxItems, uniqueItems
//   - string: minLength, maxLength, pattern, format (email/uri/uuid/date-time)
//   - number/integer: minimum, maximum, exclusiveMinimum, exclusiveMaximum, multipleOf
//   - enum, const (via single-element enum)
//   - composition: allOf, anyOf, oneOf, not
//   - nullability via Schema.Nullable
//
// Stdlib only — no external JSON-Schema libraries.
package validator

import (
	"errors"
	"fmt"
	"math"
	"net/mail"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"time"

	"oss.nandlabs.io/golly/data"
)

// SchemaValidator is a compiled schema ready to validate data. Reuse it
// across many Validate calls — compilation pre-builds regexes and walks the
// schema tree once.
type SchemaValidator struct {
	root *compiledSchema
}

// SchemaError is one violation produced during Validate. Path is a JSON
// Pointer (RFC 6901-style) locating the failing value, Message is human
// readable, Rule is the schema keyword that failed (e.g. "required",
// "minimum", "type"). Errors implement error; multiple errors are joined
// via errors.Join.
type SchemaError struct {
	Path    string
	Message string
	Rule    string
}

// Error renders the violation in "path: rule — message" form.
func (e *SchemaError) Error() string {
	if e.Path == "" {
		return fmt.Sprintf("%s: %s", e.Rule, e.Message)
	}
	return fmt.Sprintf("%s: %s — %s", e.Path, e.Rule, e.Message)
}

// CompileSchema turns a data.Schema tree into a reusable SchemaValidator.
// It pre-compiles regular expressions and returns an error if any are
// invalid (so callers see schema bugs at compile time, not validate time).
func CompileSchema(s *data.Schema) (*SchemaValidator, error) {
	if s == nil {
		return nil, errors.New("validator: cannot compile a nil schema")
	}
	root, err := compile(s)
	if err != nil {
		return nil, err
	}
	return &SchemaValidator{root: root}, nil
}

// Validate checks v against the compiled schema. Returns nil on success, or
// a multi-error built with errors.Join when one or more constraints fail.
// Each leaf error is a *SchemaError with a JSON-Pointer Path.
func (v *SchemaValidator) Validate(value any) error {
	ctx := &validateCtx{}
	v.root.validate(value, "", ctx)
	if len(ctx.errs) == 0 {
		return nil
	}
	wrapped := make([]error, len(ctx.errs))
	for i, e := range ctx.errs {
		wrapped[i] = e
	}
	return errors.Join(wrapped...)
}

// --- internal compiled form ---

type compiledSchema struct {
	src *data.Schema

	pattern *regexp.Regexp

	properties map[string]*compiledSchema
	required   map[string]struct{}

	items           *compiledSchema
	additionalItems *compiledSchema

	allOf []*compiledSchema
	anyOf []*compiledSchema
	oneOf []*compiledSchema
	not   *compiledSchema
}

type validateCtx struct {
	errs []*SchemaError
}

func (c *validateCtx) push(path, rule, msg string) {
	c.errs = append(c.errs, &SchemaError{Path: path, Rule: rule, Message: msg})
}

func compile(s *data.Schema) (*compiledSchema, error) {
	if s == nil {
		return nil, nil
	}
	cs := &compiledSchema{src: s}

	if s.Pattern != nil && *s.Pattern != "" {
		re, err := regexp.Compile(*s.Pattern)
		if err != nil {
			return nil, fmt.Errorf("validator: invalid regex %q: %w", *s.Pattern, err)
		}
		cs.pattern = re
	}

	if len(s.Properties) > 0 {
		cs.properties = make(map[string]*compiledSchema, len(s.Properties))
		for k, sub := range s.Properties {
			c, err := compile(sub)
			if err != nil {
				return nil, fmt.Errorf("validator: properties[%q]: %w", k, err)
			}
			cs.properties[k] = c
		}
	}
	if len(s.Required) > 0 {
		cs.required = make(map[string]struct{}, len(s.Required))
		for _, r := range s.Required {
			cs.required[r] = struct{}{}
		}
	}

	var err error
	if cs.items, err = compile(s.Items); err != nil {
		return nil, fmt.Errorf("validator: items: %w", err)
	}
	if cs.additionalItems, err = compile(s.AdditionalItems); err != nil {
		return nil, fmt.Errorf("validator: additionalItems: %w", err)
	}
	if cs.not, err = compile(s.Not); err != nil {
		return nil, fmt.Errorf("validator: not: %w", err)
	}
	if cs.allOf, err = compileList(s.AllOf, "allOf"); err != nil {
		return nil, err
	}
	if cs.anyOf, err = compileList(s.AnyOf, "anyOf"); err != nil {
		return nil, err
	}
	if cs.oneOf, err = compileList(s.OneOf, "oneOf"); err != nil {
		return nil, err
	}
	return cs, nil
}

func compileList(list []*data.Schema, name string) ([]*compiledSchema, error) {
	if len(list) == 0 {
		return nil, nil
	}
	out := make([]*compiledSchema, len(list))
	for i, sub := range list {
		c, err := compile(sub)
		if err != nil {
			return nil, fmt.Errorf("validator: %s[%d]: %w", name, i, err)
		}
		out[i] = c
	}
	return out, nil
}

// --- validation ---

func (cs *compiledSchema) validate(v any, path string, ctx *validateCtx) {
	if cs == nil {
		return
	}

	// Nullability
	if v == nil {
		if cs.src.Nullable || cs.src.Type == "" || cs.src.Type == "null" {
			return
		}
		ctx.push(path, "type", "value is null but type is not nullable")
		return
	}

	// Type check + per-type rules
	if cs.src.Type != "" {
		if !matchesType(cs.src.Type, v) {
			ctx.push(path, "type", fmt.Sprintf("value of type %s does not match required type %s", goKind(v), cs.src.Type))
			return // skip further constraint checks for the wrong type
		}
		switch cs.src.Type {
		case data.SchemaTypeString:
			cs.validateString(v.(string), path, ctx)
		case data.SchemaTypeNumber, data.SchemaTypeInteger:
			cs.validateNumber(toFloat(v), path, ctx)
		case data.SchemaTypeArray:
			cs.validateArray(v, path, ctx)
		case data.SchemaTypeObject:
			cs.validateObject(v, path, ctx)
		}
	}

	// Enum (works for any type via deep equality).
	if len(cs.src.Enum) > 0 && !inEnum(v, cs.src.Enum) {
		ctx.push(path, "enum", "value is not one of the allowed enum values")
	}

	// Composition keywords.
	cs.validateComposition(v, path, ctx)
}

func (cs *compiledSchema) validateString(s, path string, ctx *validateCtx) {
	if cs.src.MinLength != nil && len(s) < *cs.src.MinLength {
		ctx.push(path, "minLength", fmt.Sprintf("length %d is less than minimum %d", len(s), *cs.src.MinLength))
	}
	if cs.src.MaxLength != nil && len(s) > *cs.src.MaxLength {
		ctx.push(path, "maxLength", fmt.Sprintf("length %d exceeds maximum %d", len(s), *cs.src.MaxLength))
	}
	if cs.pattern != nil && !cs.pattern.MatchString(s) {
		ctx.push(path, "pattern", fmt.Sprintf("value does not match pattern %q", *cs.src.Pattern))
	}
	if cs.src.Format != nil {
		if err := checkFormat(*cs.src.Format, s); err != nil {
			ctx.push(path, "format", err.Error())
		}
	}
}

func (cs *compiledSchema) validateNumber(n float64, path string, ctx *validateCtx) {
	if cs.src.Minimum != nil && n < *cs.src.Minimum {
		ctx.push(path, "minimum", fmt.Sprintf("value %v is less than minimum %v", n, *cs.src.Minimum))
	}
	if cs.src.ExclusiveMinimum != nil && n <= *cs.src.ExclusiveMinimum {
		ctx.push(path, "exclusiveMinimum", fmt.Sprintf("value %v is not strictly greater than %v", n, *cs.src.ExclusiveMinimum))
	}
	if cs.src.Maximum != nil && n > *cs.src.Maximum {
		ctx.push(path, "maximum", fmt.Sprintf("value %v exceeds maximum %v", n, *cs.src.Maximum))
	}
	if cs.src.ExclusiveMaximum != nil && n >= *cs.src.ExclusiveMaximum {
		ctx.push(path, "exclusiveMaximum", fmt.Sprintf("value %v is not strictly less than %v", n, *cs.src.ExclusiveMaximum))
	}
	if cs.src.MultipleOf != nil && *cs.src.MultipleOf != 0 {
		ratio := n / *cs.src.MultipleOf
		if math.Abs(ratio-math.Round(ratio)) > 1e-9 {
			ctx.push(path, "multipleOf", fmt.Sprintf("value %v is not a multiple of %v", n, *cs.src.MultipleOf))
		}
	}
}

func (cs *compiledSchema) validateArray(v any, path string, ctx *validateCtx) {
	arr, ok := toSlice(v)
	if !ok {
		return
	}
	if cs.src.MinItems != nil && len(arr) < *cs.src.MinItems {
		ctx.push(path, "minItems", fmt.Sprintf("array has %d items, minimum %d", len(arr), *cs.src.MinItems))
	}
	if cs.src.MaxItems != nil && len(arr) > *cs.src.MaxItems {
		ctx.push(path, "maxItems", fmt.Sprintf("array has %d items, maximum %d", len(arr), *cs.src.MaxItems))
	}
	if cs.src.UniqueItems {
		if dup := firstDuplicate(arr); dup != "" {
			ctx.push(path, "uniqueItems", "array contains duplicate items at "+dup)
		}
	}
	for i, item := range arr {
		itemPath := fmt.Sprintf("%s/%d", path, i)
		if cs.items != nil {
			cs.items.validate(item, itemPath, ctx)
		}
	}
}

func (cs *compiledSchema) validateObject(v any, path string, ctx *validateCtx) {
	m, ok := toMap(v)
	if !ok {
		return
	}
	if cs.src.MinProperties != nil && len(m) < *cs.src.MinProperties {
		ctx.push(path, "minProperties", fmt.Sprintf("object has %d properties, minimum %d", len(m), *cs.src.MinProperties))
	}
	if cs.src.MaxProperties != nil && len(m) > *cs.src.MaxProperties {
		ctx.push(path, "maxProperties", fmt.Sprintf("object has %d properties, maximum %d", len(m), *cs.src.MaxProperties))
	}
	for req := range cs.required {
		if _, present := m[req]; !present {
			ctx.push(path, "required", "missing required property: "+req)
		}
	}
	// Iterate in a stable order for deterministic error output.
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		if sub, ok := cs.properties[k]; ok {
			sub.validate(m[k], path+"/"+escapePointer(k), ctx)
		}
	}
}

func (cs *compiledSchema) validateComposition(v any, path string, ctx *validateCtx) {
	if len(cs.allOf) > 0 {
		for i, sub := range cs.allOf {
			sub.validate(v, fmt.Sprintf("%s/allOf/%d", path, i), ctx)
		}
	}
	if len(cs.anyOf) > 0 {
		matched := false
		for _, sub := range cs.anyOf {
			if subValid(sub, v) {
				matched = true
				break
			}
		}
		if !matched {
			ctx.push(path, "anyOf", "value does not match any subschema")
		}
	}
	if len(cs.oneOf) > 0 {
		matches := 0
		for _, sub := range cs.oneOf {
			if subValid(sub, v) {
				matches++
			}
		}
		if matches != 1 {
			ctx.push(path, "oneOf", fmt.Sprintf("value matches %d subschemas; expected exactly 1", matches))
		}
	}
	if cs.not != nil && subValid(cs.not, v) {
		ctx.push(path, "not", "value matches forbidden subschema")
	}
}

func subValid(cs *compiledSchema, v any) bool {
	if cs == nil {
		return true
	}
	tmp := &validateCtx{}
	cs.validate(v, "", tmp)
	return len(tmp.errs) == 0
}

// --- helpers ---

func matchesType(t string, v any) bool {
	switch t {
	case data.SchemaTypeString:
		_, ok := v.(string)
		return ok
	case data.SchemaTypeBool:
		_, ok := v.(bool)
		return ok
	case data.SchemaTypeNumber:
		return isNumber(v)
	case data.SchemaTypeInteger:
		return isInteger(v)
	case data.SchemaTypeArray:
		_, ok := toSlice(v)
		return ok
	case data.SchemaTypeObject:
		_, ok := toMap(v)
		return ok
	case "null":
		return v == nil
	}
	return true
}

func isNumber(v any) bool {
	switch v.(type) {
	case int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64,
		float32, float64:
		return true
	}
	return false
}

func isInteger(v any) bool {
	switch x := v.(type) {
	case int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64:
		return true
	case float32:
		return float64(x) == math.Trunc(float64(x))
	case float64:
		return x == math.Trunc(x)
	}
	return false
}

func toFloat(v any) float64 {
	switch x := v.(type) {
	case int:
		return float64(x)
	case int8:
		return float64(x)
	case int16:
		return float64(x)
	case int32:
		return float64(x)
	case int64:
		return float64(x)
	case uint:
		return float64(x)
	case uint8:
		return float64(x)
	case uint16:
		return float64(x)
	case uint32:
		return float64(x)
	case uint64:
		return float64(x)
	case float32:
		return float64(x)
	case float64:
		return x
	}
	return 0
}

func toSlice(v any) ([]any, bool) {
	if s, ok := v.([]any); ok {
		return s, true
	}
	return nil, false
}

func toMap(v any) (map[string]any, bool) {
	if m, ok := v.(map[string]any); ok {
		return m, true
	}
	return nil, false
}

func goKind(v any) string {
	switch v.(type) {
	case nil:
		return "null"
	case string:
		return "string"
	case bool:
		return "boolean"
	case []any:
		return "array"
	case map[string]any:
		return "object"
	}
	if isInteger(v) {
		return "integer"
	}
	if isNumber(v) {
		return "number"
	}
	return "unknown"
}

func inEnum(v any, allowed []any) bool {
	for _, a := range allowed {
		if deepEqual(a, v) {
			return true
		}
	}
	return false
}

// deepEqual compares two decoded-JSON values for equality. We can't use
// reflect.DeepEqual directly because json numbers all decode to float64
// while the enum may carry ints — normalise via toFloat for numerics.
func deepEqual(a, b any) bool {
	if a == nil || b == nil {
		return a == b
	}
	if isNumber(a) && isNumber(b) {
		return toFloat(a) == toFloat(b)
	}
	switch ax := a.(type) {
	case []any:
		bx, ok := b.([]any)
		if !ok || len(ax) != len(bx) {
			return false
		}
		for i := range ax {
			if !deepEqual(ax[i], bx[i]) {
				return false
			}
		}
		return true
	case map[string]any:
		bx, ok := b.(map[string]any)
		if !ok || len(ax) != len(bx) {
			return false
		}
		for k, av := range ax {
			bv, present := bx[k]
			if !present || !deepEqual(av, bv) {
				return false
			}
		}
		return true
	}
	return a == b
}

func firstDuplicate(arr []any) string {
	for i := 0; i < len(arr); i++ {
		for j := i + 1; j < len(arr); j++ {
			if deepEqual(arr[i], arr[j]) {
				return fmt.Sprintf("indices %d and %d", i, j)
			}
		}
	}
	return ""
}

// escapePointer escapes a JSON Pointer reference token per RFC 6901.
func escapePointer(s string) string {
	s = strings.ReplaceAll(s, "~", "~0")
	s = strings.ReplaceAll(s, "/", "~1")
	return s
}

// checkFormat validates a small but commonly useful subset of JSON Schema formats.
// Unknown formats are accepted (per the spec, format is advisory unless the
// validator chooses to enforce it).
func checkFormat(format, s string) error {
	switch format {
	case "email":
		if _, err := mail.ParseAddress(s); err != nil {
			return fmt.Errorf("value is not a valid email address")
		}
	case "uri", "uri-reference":
		if _, err := url.Parse(s); err != nil {
			return fmt.Errorf("value is not a valid URI")
		}
		if format == "uri" {
			u, _ := url.Parse(s)
			if u == nil || u.Scheme == "" {
				return fmt.Errorf("value is not an absolute URI")
			}
		}
	case "uuid":
		if !uuidRegex.MatchString(s) {
			return fmt.Errorf("value is not a valid UUID")
		}
	case "date-time":
		if _, err := time.Parse(time.RFC3339, s); err != nil {
			return fmt.Errorf("value is not a valid RFC3339 date-time")
		}
	case "date":
		if _, err := time.Parse("2006-01-02", s); err != nil {
			return fmt.Errorf("value is not a valid date")
		}
	case "time":
		if _, err := time.Parse("15:04:05Z07:00", s); err != nil {
			return fmt.Errorf("value is not a valid time")
		}
	}
	return nil
}

var uuidRegex = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
