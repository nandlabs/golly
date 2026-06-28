package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"

	yaml "github.com/goccy/go-yaml"
)

// Source produces a nested map of config values. Multiple sources can be
// layered with LoadInto; later sources override earlier ones key-by-key.
type Source interface {
	Provide() (map[string]any, error)
}

// FromFile returns a Source that loads the file at path and decodes it
// according to extension: .yaml/.yml → YAML, .json → JSON, .properties →
// Java-style key=value. Unknown extensions are treated as YAML.
func FromFile(path string) Source {
	return &fileSource{path: path}
}

// FromEnv returns a Source that reads every environment variable beginning
// with prefix (or all of them if prefix is empty) and converts the suffix
// into a dot-notation key with lowercased segments split on "_".
//
// FOO_DB__DSN=x → {"db_dsn": "x"} when prefix="FOO_"; pair this with
// `config:"db_dsn"` or its dotted equivalent on the receiving struct.
//
// Double-underscore is preserved as a single underscore to support nested
// keys: FOO_DB__DSN → "db_dsn"; FOO_DB_HOST → "db.host".
func FromEnv(prefix string) Source { return &envSource{prefix: prefix} }

// FromMap returns a Source that yields the given map verbatim. Useful for
// defaults and tests.
func FromMap(m map[string]any) Source { return &mapSource{m: m} }

// LoadInto layers the given sources (later wins on duplicate keys), then
// binds the result into dst (a pointer to a struct) using `config:"path"`
// struct tags. Untagged fields fall back to the lowercased field name.
// Sources are loaded once each in order. Returns an error if any source
// fails or if a tag-named key cannot be coerced to the field's type.
func LoadInto(dst any, sources ...Source) error {
	if dst == nil {
		return errors.New("config: LoadInto dst is nil")
	}
	rv := reflect.ValueOf(dst)
	if rv.Kind() != reflect.Pointer || rv.IsNil() || rv.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("config: LoadInto requires a non-nil pointer to a struct; got %T", dst)
	}

	merged := map[string]any{}
	for i, src := range sources {
		if src == nil {
			continue
		}
		m, err := src.Provide()
		if err != nil {
			return fmt.Errorf("config: source[%d]: %w", i, err)
		}
		mergeMaps(merged, m)
	}

	return bind(rv.Elem(), merged, "")
}

// --- sources ---

type fileSource struct{ path string }

func (s *fileSource) Provide() (map[string]any, error) {
	data, err := os.ReadFile(s.path)
	if err != nil {
		return nil, fmt.Errorf("read %q: %w", s.path, err)
	}
	ext := strings.ToLower(filepath.Ext(s.path))
	out := map[string]any{}
	switch ext {
	case ".json":
		if err := json.Unmarshal(data, &out); err != nil {
			return nil, fmt.Errorf("decode json: %w", err)
		}
	case ".properties":
		return decodeProperties(data)
	default:
		// .yaml / .yml / "" / anything else → YAML
		if err := yaml.Unmarshal(data, &out); err != nil {
			return nil, fmt.Errorf("decode yaml: %w", err)
		}
	}
	return normalizeKeys(out), nil
}

func decodeProperties(data []byte) (map[string]any, error) {
	out := map[string]any{}
	for ln := range strings.Lines(string(data)) {
		ln = strings.TrimSpace(ln)
		if ln == "" || strings.HasPrefix(ln, "#") || strings.HasPrefix(ln, "!") {
			continue
		}
		eq := strings.IndexAny(ln, "=:")
		if eq < 0 {
			continue
		}
		key := strings.TrimSpace(ln[:eq])
		val := strings.TrimSpace(ln[eq+1:])
		setNested(out, splitKey(key), val)
	}
	return out, nil
}

type envSource struct{ prefix string }

func (s *envSource) Provide() (map[string]any, error) {
	out := map[string]any{}
	for _, kv := range os.Environ() {
		eq := strings.IndexByte(kv, '=')
		if eq < 0 {
			continue
		}
		name, val := kv[:eq], kv[eq+1:]
		if s.prefix != "" && !strings.HasPrefix(name, s.prefix) {
			continue
		}
		stripped := strings.TrimPrefix(name, s.prefix)
		// Convert FOO__BAR → "foo_bar", FOO_BAR → "foo.bar"
		key := envNameToKey(stripped)
		setNested(out, splitKey(key), val)
	}
	return out, nil
}

// envNameToKey turns an env-var suffix into a config dot-key. Double
// underscore becomes a literal underscore in the segment (so keys can
// contain underscores), single underscore becomes a path separator.
func envNameToKey(s string) string {
	s = strings.ToLower(s)
	// Use sentinel to protect "__" sequences during the split.
	const sentinel = "\x00"
	s = strings.ReplaceAll(s, "__", sentinel)
	s = strings.ReplaceAll(s, "_", ".")
	s = strings.ReplaceAll(s, sentinel, "_")
	return s
}

type mapSource struct{ m map[string]any }

func (s *mapSource) Provide() (map[string]any, error) {
	return normalizeKeys(deepCopyMap(s.m)), nil
}

// --- merging / binding helpers ---

// mergeMaps deep-merges src into dst. Map values recurse; scalars and slices
// replace.
func mergeMaps(dst, src map[string]any) {
	for k, v := range src {
		if sub, ok := v.(map[string]any); ok {
			if existing, ok2 := dst[k].(map[string]any); ok2 {
				mergeMaps(existing, sub)
				continue
			}
		}
		dst[k] = v
	}
}

// normalizeKeys converts any map[any]any (a quirk of some YAML decoders) into
// map[string]any recursively, leaving all leaf values as-is.
func normalizeKeys(in map[string]any) map[string]any {
	for k, v := range in {
		in[k] = normalizeValue(v)
	}
	return in
}

func normalizeValue(v any) any {
	switch x := v.(type) {
	case map[any]any:
		out := make(map[string]any, len(x))
		for k, val := range x {
			out[fmt.Sprintf("%v", k)] = normalizeValue(val)
		}
		return out
	case map[string]any:
		return normalizeKeys(x)
	case []any:
		for i := range x {
			x[i] = normalizeValue(x[i])
		}
		return x
	}
	return v
}

func deepCopyMap(in map[string]any) map[string]any {
	out := make(map[string]any, len(in))
	for k, v := range in {
		if m, ok := v.(map[string]any); ok {
			out[k] = deepCopyMap(m)
		} else {
			out[k] = v
		}
	}
	return out
}

// splitKey breaks a dotted path into segments. Empty segments are skipped.
func splitKey(k string) []string {
	if k == "" {
		return nil
	}
	parts := strings.Split(k, ".")
	out := parts[:0]
	for _, p := range parts {
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

// setNested sets a value at a dotted path inside m, creating intermediate
// maps as needed.
func setNested(m map[string]any, path []string, val any) {
	if len(path) == 0 {
		return
	}
	for i, p := range path[:len(path)-1] {
		next, ok := m[p].(map[string]any)
		if !ok {
			next = map[string]any{}
			m[p] = next
		}
		_ = i
		m = next
	}
	m[path[len(path)-1]] = val
}

// lookup retrieves a value from a nested map by dotted path. Returns
// (value, true) if found.
func lookup(m map[string]any, path []string) (any, bool) {
	if len(path) == 0 {
		return nil, false
	}
	cur := any(m)
	for _, p := range path {
		mp, ok := cur.(map[string]any)
		if !ok {
			return nil, false
		}
		cur, ok = mp[p]
		if !ok {
			return nil, false
		}
	}
	return cur, true
}

// bind walks a struct value and populates fields from m using config tags
// (or lowercased field name as fallback). prefix is the dotted path of the
// containing struct.
func bind(rv reflect.Value, m map[string]any, prefix string) error {
	rt := rv.Type()
	for i := 0; i < rt.NumField(); i++ {
		sf := rt.Field(i)
		if !sf.IsExported() {
			continue
		}
		tag := sf.Tag.Get("config")
		if tag == "-" {
			continue
		}
		key := tag
		if key == "" {
			key = strings.ToLower(sf.Name)
		}
		fullKey := key
		if prefix != "" {
			fullKey = prefix + "." + key
		}

		// Nested struct: recurse with its sub-map.
		field := rv.Field(i)
		if field.Kind() == reflect.Struct && sf.Type != reflect.TypeOf(time.Time{}) {
			sub, _ := lookup(m, splitKey(key))
			subMap, _ := sub.(map[string]any)
			if subMap == nil {
				subMap = map[string]any{}
			}
			if err := bind(field, subMap, fullKey); err != nil {
				return err
			}
			continue
		}

		val, ok := lookup(m, splitKey(key))
		if !ok {
			continue
		}
		if err := setField(field, val); err != nil {
			return fmt.Errorf("config: field %q: %w", fullKey, err)
		}
	}
	return nil
}

// setField coerces v into the destination field's type and assigns it. The
// coercion handles the common JSON/YAML decode types (string/float64/bool/
// []any/map[string]any) plus string-from-env to typed.
func setField(field reflect.Value, v any) error {
	if v == nil {
		return nil
	}
	if !field.CanSet() {
		return errors.New("field is not settable")
	}

	// Special-case time.Duration before generic numeric handling.
	if field.Type() == reflect.TypeOf(time.Duration(0)) {
		switch x := v.(type) {
		case string:
			d, err := time.ParseDuration(x)
			if err != nil {
				return fmt.Errorf("parse duration: %w", err)
			}
			field.Set(reflect.ValueOf(d))
			return nil
		case int, int64, float64:
			field.SetInt(int64(toFloat64(v)))
			return nil
		}
	}

	switch field.Kind() {
	case reflect.String:
		field.SetString(fmt.Sprintf("%v", v))
	case reflect.Bool:
		switch x := v.(type) {
		case bool:
			field.SetBool(x)
		case string:
			b, err := strconv.ParseBool(x)
			if err != nil {
				return err
			}
			field.SetBool(b)
		default:
			return fmt.Errorf("cannot coerce %T to bool", v)
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n, err := toInt64(v)
		if err != nil {
			return err
		}
		field.SetInt(n)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		n, err := toInt64(v)
		if err != nil {
			return err
		}
		if n < 0 {
			return fmt.Errorf("negative value %d for unsigned field", n)
		}
		field.SetUint(uint64(n))
	case reflect.Float32, reflect.Float64:
		f, err := toFloat64Err(v)
		if err != nil {
			return err
		}
		field.SetFloat(f)
	case reflect.Slice:
		return setSlice(field, v)
	case reflect.Map:
		return setMap(field, v)
	case reflect.Pointer:
		// Allocate and recurse.
		if field.IsNil() {
			field.Set(reflect.New(field.Type().Elem()))
		}
		return setField(field.Elem(), v)
	default:
		// JSON/YAML round-trip catch-all for nested struct-like things.
		raw, err := json.Marshal(v)
		if err != nil {
			return err
		}
		if err := json.Unmarshal(raw, field.Addr().Interface()); err != nil {
			return err
		}
	}
	return nil
}

func setSlice(field reflect.Value, v any) error {
	arr, ok := v.([]any)
	if !ok {
		// Allow comma-separated string from env: "a,b,c".
		if s, ok2 := v.(string); ok2 {
			parts := strings.Split(s, ",")
			arr = make([]any, len(parts))
			for i, p := range parts {
				arr[i] = strings.TrimSpace(p)
			}
		} else {
			return fmt.Errorf("cannot coerce %T to slice", v)
		}
	}
	out := reflect.MakeSlice(field.Type(), len(arr), len(arr))
	for i, item := range arr {
		if err := setField(out.Index(i), item); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	field.Set(out)
	return nil
}

func setMap(field reflect.Value, v any) error {
	src, ok := v.(map[string]any)
	if !ok {
		return fmt.Errorf("cannot coerce %T to map", v)
	}
	out := reflect.MakeMapWithSize(field.Type(), len(src))
	for k, val := range src {
		keyVal := reflect.New(field.Type().Key()).Elem()
		if keyVal.Kind() == reflect.String {
			keyVal.SetString(k)
		} else {
			return fmt.Errorf("non-string map key type %s unsupported", field.Type().Key().Kind())
		}
		valVal := reflect.New(field.Type().Elem()).Elem()
		if err := setField(valVal, val); err != nil {
			return fmt.Errorf("key %q: %w", k, err)
		}
		out.SetMapIndex(keyVal, valVal)
	}
	field.Set(out)
	return nil
}

func toInt64(v any) (int64, error) {
	switch x := v.(type) {
	case int:
		return int64(x), nil
	case int64:
		return x, nil
	case int32:
		return int64(x), nil
	case float64:
		return int64(x), nil
	case float32:
		return int64(x), nil
	case string:
		return strconv.ParseInt(x, 10, 64)
	}
	return 0, fmt.Errorf("cannot coerce %T to int", v)
}

func toFloat64Err(v any) (float64, error) {
	switch x := v.(type) {
	case float64:
		return x, nil
	case float32:
		return float64(x), nil
	case int:
		return float64(x), nil
	case int64:
		return float64(x), nil
	case string:
		return strconv.ParseFloat(x, 64)
	}
	return 0, fmt.Errorf("cannot coerce %T to float", v)
}

func toFloat64(v any) float64 {
	f, _ := toFloat64Err(v)
	return f
}
