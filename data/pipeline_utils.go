package data

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// ExtractValue retrieves a value of type T from the Pipeline using the provided path.
// If the value is not of type T, it returns an ErrInvalidType error.
// The path can be a simple key or a dot-separated path (e.g., "user.address.city").
//
// Parameters:
//   - c: A pointer to the Pipeline from which to extract the value.
//   - path: The path to the value to be retrieved. Can be a simple key or dot notation (e.g., "user.address.city").
//     The path also supports filters for arrays and maps using square brackets, e.g.:
//   - users[0]                // Index access
//   - users[name=="nanda"]   // String equality filter
//   - users[age>25]           // Numeric comparison filter
//   - users[age>=18 && city=="blr"] // Logical AND
//   - users[city=="blr" || city=="nyc"] // Logical OR
//   - users[(age>18 && city=="blr") || (age<10)] // Grouping with parentheses
//   - users[city!="nyc"]      // Not equal
//     Supported filter operators: ==, !=, >, <, >=, <=, &&, ||, parentheses for grouping.
//
// Returns:
//   - value: The value of type T associated with the provided path.
//   - err: An error if the path does not exist or the value is not of type T.
func ExtractValue[T any](c Pipeline, path string) (value T, err error) {
	parts := extractPath(path)
	rootKey, filter, hasFilter := parseFieldAndFilter(parts[0])
	var current any
	current, err = c.Get(rootKey)
	if err != nil {
		return
	}
	if hasFilter {
		current, err = applyFilter(current, filter)
		if err != nil {
			return
		}
	}
	for i := 1; i < len(parts); i++ {
		field, filter, hasFilter := parseFieldAndFilter(parts[i])
		if len(field) > 0 {
			current, err = navigateToField(current, field)
			if err != nil {
				switch err {
				case ErrFieldNotFound:
					err = fmt.Errorf("%w: field '%s' in path '%s'", ErrFieldNotFound, field, path)
				case ErrInvalidPath:
					err = fmt.Errorf("%w: invalid segment '%s' in path '%s'", ErrInvalidPath, field, path)
				}
				return
			}
			if hasFilter {
				current, err = applyFilter(current, filter)
				if err != nil {
					return
				}
			}
		} else {
			err = ErrInvalidPath
			return
		}
	}
	if current == nil {
		err = ErrInvalidType
		return
	}
	return Convert[T](current)
}

// navigateToField navigates to a field within a value using reflection.
// It handles maps, structs, and other types that can contain nested fields.
func navigateToField(value any, fieldName string) (any, error) {
	if value == nil {
		return nil, ErrFieldNotFound
	}

	v := reflect.ValueOf(value)

	// Handle different types
	switch v.Kind() {
	case reflect.Map:
		// For maps, get the value using the field name as key

		// For string keys, handle directly
		if mv, ok := value.(map[string]any); ok {
			if val, exists := mv[fieldName]; exists {
				return val, nil
			}
			return nil, ErrFieldNotFound
		}

		// For other map types, use reflection
		mapKey := reflect.ValueOf(fieldName)
		mapValue := v.MapIndex(mapKey)
		if !mapValue.IsValid() {
			return nil, ErrFieldNotFound
		}
		return mapValue.Interface(), nil

	case reflect.Struct:
		// For structs, get the field using reflection
		field := v.FieldByName(fieldName)
		if !field.IsValid() {
			return nil, ErrFieldNotFound
		}
		return field.Interface(), nil

	case reflect.Ptr:
		// For pointers, dereference and try again
		if v.IsNil() {
			return nil, ErrFieldNotFound
		}
		return navigateToField(v.Elem().Interface(), fieldName)

	case reflect.Slice, reflect.Array:
		// Try to parse the field name as an index
		index, err := strconv.Atoi(fieldName)
		if err != nil {
			return nil, ErrInvalidPath
		}

		// Check if the index is valid
		if index < 0 || index >= v.Len() {
			return nil, ErrFieldNotFound
		}

		return v.Index(index).Interface(), nil

	default:
		return nil, fmt.Errorf("cannot navigate into value of type %T", value)
	}
}

// Extracts the field name and optional filter from a path segment, e.g. users[0], users[name=="nanda"], users[address.zip>20000 && phones[type=="home"].number=="123"]
func parseFieldAndFilter(segment string) (fieldName string, filter string, hasFilter bool) {
	open := strings.Index(segment, "[")
	if open == -1 {
		fieldName = segment
		return
	}
	// Find the matching closing bracket, handling nested brackets and quoted strings
	bracketLevel := 0
	inQuotes := false
	for i := open; i < len(segment); i++ {
		c := segment[i]
		if c == '"' {
			inQuotes = !inQuotes
		}
		if !inQuotes {
			if c == '[' {
				bracketLevel++
			} else if c == ']' {
				bracketLevel--
				if bracketLevel == 0 {
					fieldName = segment[:open]
					filter = segment[open+1 : i]
					hasFilter = true
					return
				}
			}
		}
	}
	// If we get here, no matching closing bracket was found
	fieldName = segment
	return
}

// PathSegment represents a segment of a path, possibly with a filter.
type PathSegment struct {
	Field  string
	Filter string // empty if no filter
}

// parseComplexPath parses a path like phones[type=="home"].number into segments.
func parseComplexPath(s string) []PathSegment {
	var segments []PathSegment
	var sb strings.Builder
	bracketLevel := 0
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == '[' {
			bracketLevel++
		}
		if c == ']' {
			if bracketLevel > 0 {
				bracketLevel--
			}
		}
		if c == '.' && bracketLevel == 0 {
			segments = append(segments, parseFieldAndFilterToSegment(sb.String()))
			sb.Reset()
		} else {
			sb.WriteByte(c)
		}
	}
	if sb.Len() > 0 {
		segments = append(segments, parseFieldAndFilterToSegment(sb.String()))
	}
	return segments
}

func parseFieldAndFilterToSegment(segment string) PathSegment {
	field, filter, hasFilter := parseFieldAndFilter(segment)
	if hasFilter {
		return PathSegment{field, filter}
	}
	return PathSegment{segment, ""}
}

// resolveComplexPath navigates through nested maps/structs/arrays for a path with filters.
func resolveComplexPath(item any, path []PathSegment) (any, error) {
	current := item
	var err error
	for _, seg := range path {
		current, err = navigateToField(current, seg.Field)
		if err != nil {
			return nil, err
		}
		if seg.Filter != "" {
			current, err = applyFilter(current, seg.Filter)
			if err != nil {
				return nil, err
			}
		}
	}
	return current, nil
}

// FilterExpr represents a parsed filter expression.
type FilterExpr interface {
	Eval(item any) (bool, error)
}

// ComparisonExpr supports ==, !=, >, <, >=, <=
type ComparisonExpr struct {
	Path  []PathSegment // supports dot-separated path with filters
	Op    string
	Value any
}

func (c *ComparisonExpr) Eval(item any) (bool, error) {
	fieldVal, err := resolveComplexPath(item, c.Path)
	if err != nil {
		return false, nil
	}
	v := reflect.ValueOf(fieldVal)
	if v.Kind() == reflect.Slice || v.Kind() == reflect.Array {
		for i := 0; i < v.Len(); i++ {
			val := v.Index(i).Interface()
			ok, _ := compareValue(val, c.Op, c.Value)
			if ok {
				return true, nil
			}
		}
		return false, nil
	}
	// If the value is itself a slice/array (e.g., from a map value), check each element
	if v.Kind() == reflect.Map {
		iter := v.MapRange()
		for iter.Next() {
			val := iter.Value().Interface()
			ok, _ := compareValue(val, c.Op, c.Value)
			if ok {
				return true, nil
			}
		}
		return false, nil
	}
	return compareValue(fieldVal, c.Op, c.Value)
}

// compareValue performs the comparison logic for a single value
func compareValue(fieldVal any, op string, cmpVal any) (bool, error) {
	// Try numeric comparison
	fv, fvOk := toFloat(fieldVal)
	var cv float64
	var cvOk bool
	switch v := cmpVal.(type) {
	case float64:
		cv, cvOk = v, true
	case string:
		cv, cvOk = toFloat(v)
	}
	if fvOk && cvOk {
		switch op {
		case ">":
			return fv > cv, nil
		case ">=":
			return fv >= cv, nil
		case "<":
			return fv < cv, nil
		case "<=":
			return fv <= cv, nil
		case "==":
			return fv == cv, nil
		case "!=":
			return fv != cv, nil
		}
	}
	// Fallback to string comparison
	fs := toString(fieldVal)
	cs := toString(cmpVal)
	switch op {
	case "==":
		return fs == cs, nil
	case "!=":
		return fs != cs, nil
	case ">":
		return fs > cs, nil
	case ">=":
		return fs >= cs, nil
	case "<":
		return fs < cs, nil
	case "<=":
		return fs <= cs, nil
	}
	return false, nil
}

// LogicalExpr supports AND, OR, grouping

type LogicalExpr struct {
	Op    string // "&&" or "||"
	Left  FilterExpr
	Right FilterExpr
}

func (l *LogicalExpr) Eval(item any) (bool, error) {
	lv, err := l.Left.Eval(item)
	if err != nil {
		return false, err
	}
	if l.Op == "&&" {
		if !lv {
			return false, nil
		}
		rv, err := l.Right.Eval(item)
		return rv, err
	} else if l.Op == "||" {
		if lv {
			return true, nil
		}
		rv, err := l.Right.Eval(item)
		return rv, err
	}
	return false, nil
}

// ParenExpr for grouping

type ParenExpr struct {
	Inner FilterExpr
}

func (p *ParenExpr) Eval(item any) (bool, error) {
	return p.Inner.Eval(item)
}

// ExistsExpr checks if a path resolves to a non-nil value
// Used for expressions like users[address.city=="blr"]
type ExistsExpr struct {
	Path []PathSegment
}

func (e *ExistsExpr) Eval(item any) (bool, error) {
	val, err := resolveComplexPath(item, e.Path)
	if err != nil {
		return false, nil
	}
	if val == nil {
		return false, nil
	}
	v := reflect.ValueOf(val)
	if v.Kind() == reflect.Slice || v.Kind() == reflect.Array {
		return v.Len() > 0, nil
	}
	return true, nil
}

// Helper: convert to float64 if possible
func toFloat(v any) (float64, bool) {
	switch val := v.(type) {
	case int:
		return float64(val), true
	case int8:
		return float64(val), true
	case int16:
		return float64(val), true
	case int32:
		return float64(val), true
	case int64:
		return float64(val), true
	case uint:
		return float64(val), true
	case uint8:
		return float64(val), true
	case uint16:
		return float64(val), true
	case uint32:
		return float64(val), true
	case uint64:
		return float64(val), true
	case float32:
		return float64(val), true
	case float64:
		return val, true
	case string:
		f, err := strconv.ParseFloat(val, 64)
		if err == nil {
			return f, true
		}
	}
	return 0, false
}

// Helper: convert to string
func toString(v any) string {
	switch val := v.(type) {
	case string:
		return val
	case fmt.Stringer:
		return val.String()
	default:
		return fmt.Sprintf("%v", v)
	}
}

// Tokenizer for filter expressions
type token struct {
	type_ string
	val   string
}

func tokenizeFilter(s string) []token {
	tokens := []token{}
	i := 0
	for i < len(s) {
		switch {
		case s[i] == ' ':
			i++
		case s[i] == '.':
			tokens = append(tokens, token{"DOT", "."})
			i++
		case s[i] == '[':
			tokens = append(tokens, token{"LBRACK", "["})
			i++
		case s[i] == ']':
			tokens = append(tokens, token{"RBRACK", "]"})
			i++
		case i+1 < len(s) && s[i:i+2] == "&&":
			tokens = append(tokens, token{"AND", "&&"})
			i += 2
		case i+1 < len(s) && s[i:i+2] == "||":
			tokens = append(tokens, token{"OR", "||"})
			i += 2
		case i+1 < len(s) && (s[i:i+2] == ">=" || s[i:i+2] == "<=" || s[i:i+2] == "==" || s[i:i+2] == "!="):
			tokens = append(tokens, token{"OP", s[i : i+2]})
			i += 2
		case s[i] == '>' || s[i] == '<':
			tokens = append(tokens, token{"OP", string(s[i])})
			i++
		case s[i] == '"':
			j := i + 1
			for j < len(s) && s[j] != '"' {
				j++
			}
			if j < len(s) {
				tokens = append(tokens, token{"STRING", s[i+1 : j]})
				i = j + 1
			} else {
				tokens = append(tokens, token{"STRING", s[i+1:]})
				i = len(s)
			}
		case isAlpha(s[i]):
			j := i
			for j < len(s) && (isAlphaNum(s[j]) || s[j] == '_') {
				j++
			}
			tokens = append(tokens, token{"IDENT", s[i:j]})
			i = j
		case isDigit(s[i]):
			j := i
			for j < len(s) && (isDigit(s[j]) || s[j] == '.') {
				j++
			}
			tokens = append(tokens, token{"NUMBER", s[i:j]})
			i = j
		default:
			i++
		}
	}
	return tokens
}

func isAlpha(b byte) bool { return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') }
func isAlphaNum(b byte) bool {
	return isAlpha(b) || isDigit(b)
}
func isDigit(b byte) bool { return b >= '0' && b <= '9' }

// Parser for filter expressions
func parseFilterExpr(tokens []token) (FilterExpr, int) {
	return parseOrExpr(tokens, 0)
}

func parseOrExpr(tokens []token, pos int) (FilterExpr, int) {
	left, pos := parseAndExpr(tokens, pos)
	for pos < len(tokens) && tokens[pos].type_ == "OR" {
		pos++
		right, newPos := parseAndExpr(tokens, pos)
		left = &LogicalExpr{"||", left, right}
		pos = newPos
	}
	return left, pos
}

func parseAndExpr(tokens []token, pos int) (FilterExpr, int) {
	left, pos := parsePrimaryExpr(tokens, pos)
	for pos < len(tokens) && tokens[pos].type_ == "AND" {
		pos++
		right, newPos := parsePrimaryExpr(tokens, pos)
		left = &LogicalExpr{"&&", left, right}
		pos = newPos
	}
	return left, pos
}

// Helper to collect path tokens up to the next top-level OP/AND/OR, handling nested brackets and parentheses
func collectPathTokens(tokens []token, pos int) ([]token, int) {
	var pathToks []token
	bracketLevel := 0
	parenLevel := 0
	for pos < len(tokens) {
		t := tokens[pos]
		switch t.type_ {
		case "LBRACK":
			bracketLevel++
		case "RBRACK":
			if bracketLevel > 0 {
				bracketLevel--
			}
		case "LPAREN":
			parenLevel++
		case "RPAREN":
			if parenLevel > 0 {
				parenLevel--
			}
		}
		// Only break for OP/AND/OR at top level (not inside any bracket or paren)
		if (t.type_ == "OP" || t.type_ == "AND" || t.type_ == "OR") && bracketLevel == 0 && parenLevel == 0 {
			break
		}
		pathToks = append(pathToks, t)
		pos++
	}
	return pathToks, pos
}

func parsePrimaryExpr(tokens []token, pos int) (FilterExpr, int) {
	start := pos
	// Handle parenthesis at the start
	if tokens[start].type_ == "LPAREN" {
		inner, newPos := parseOrExpr(tokens, start+1)
		if newPos < len(tokens) && tokens[newPos].type_ == "RPAREN" {
			return &ParenExpr{inner}, newPos + 1
		}
		return &ParenExpr{inner}, newPos
	}
	// Collect path tokens robustly
	pathToks, newPos := collectPathTokens(tokens, pos)
	pos = newPos
	if len(pathToks) > 0 && pos < len(tokens) && tokens[pos].type_ == "OP" {
		pathStr := tokensToString(pathToks)
		path := parseComplexPath(pathStr)
		op := tokens[pos].val
		if pos+1 >= len(tokens) {
			return nil, pos + 1
		}
		valTok := tokens[pos+1]
		var val any
		if valTok.type_ == "NUMBER" {
			f, _ := strconv.ParseFloat(valTok.val, 64)
			val = f
		} else {
			val = valTok.val
		}
		return &ComparisonExpr{path, op, val}, pos + 2
	}
	// If we collected path tokens but no OP, treat as ExistsExpr
	if len(pathToks) > 0 {
		pathStr := tokensToString(pathToks)
		path := parseComplexPath(pathStr)
		return &ExistsExpr{path}, pos
	}
	return nil, pos + 1
}

func tokensToString(tokens []token) string {
	var sb strings.Builder
	for _, t := range tokens {
		sb.WriteString(t.val)
	}
	return sb.String()
}

// Applies a filter to a value (slice, array, or map). Supports index and key==value filters, and complex expressions.
func applyFilter(value any, filter string) (any, error) {
	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.Slice, reflect.Array:
		// Index filter: [0]
		if idx, err := strconv.Atoi(filter); err == nil {
			if idx < 0 || idx >= v.Len() {
				return nil, ErrFieldNotFound
			}
			return v.Index(idx).Interface(), nil
		}
		// Complex filter: parse and evaluate
		tokens := tokenizeFilter(filter)
		expr, _ := parseFilterExpr(tokens)
		for i := 0; i < v.Len(); i++ {
			item := v.Index(i).Interface()
			if expr != nil {
				ok, err := expr.Eval(item)
				if err == nil && ok {
					return item, nil
				}
			}
		}
		return nil, ErrFieldNotFound
	case reflect.Map:
		// Map key filter: [key]
		mapKey := reflect.ValueOf(filter)
		mapValue := v.MapIndex(mapKey)
		if !mapValue.IsValid() {
			return nil, ErrFieldNotFound
		}
		return mapValue.Interface(), nil
	}
	return nil, ErrInvalidPath
}

// extractPath splits a path into segments, respecting brackets so that dots inside brackets are not split.
func extractPath(path string) []string {
	var parts []string
	var sb strings.Builder
	bracketLevel := 0
	for i := 0; i < len(path); i++ {
		c := path[i]
		if c == '[' {
			bracketLevel++
		}
		if c == ']' {
			if bracketLevel > 0 {
				bracketLevel--
			}
		}
		if c == '.' && bracketLevel == 0 {
			parts = append(parts, sb.String())
			sb.Reset()
		} else {
			sb.WriteByte(c)
		}
	}
	if sb.Len() > 0 {
		parts = append(parts, sb.String())
	}
	return parts
}

// EvaluateCondition evaluates a filter-like condition string against the pipeline root.
// Returns true if the condition is satisfied, false otherwise.
func EvaluateCondition(p Pipeline, condition string) bool {
	tokens := tokenizeFilter(condition)
	expr, _ := parseFilterExpr(tokens)
	if expr == nil {
		return false
	}
	root, err := p.Get("") // Get the root object; adjust if your Pipeline root access differs
	if err != nil {
		return false
	}
	ok, err := expr.Eval(root)
	if err != nil {
		return false
	}
	return ok
}

// SetValue sets a value at the specified path in the Pipeline, supporting filters and nested paths.
// For a nested path, the final item must be a Pipeline object, and the value is set using its Set method.
func SetValue(c Pipeline, path string, value any) error {
	parts := extractPath(path)
	if len(parts) == 0 {
		return ErrInvalidPath
	}
	rootKey, filter, hasFilter := parseFieldAndFilter(parts[0])
	var current any
	var err error
	current, err = c.Get(rootKey)
	if err != nil && !hasFilter && len(parts) == 1 {
		current = c
	} else if err != nil {
		return err
	}
	if hasFilter {
		current, err = applyFilter(current, filter)
		if err != nil {
			return err
		}
	}
	// Traverse up to the second last part
	for i := 1; i < len(parts)-1; i++ {
		field, filter, hasFilter := parseFieldAndFilter(parts[i])
		if len(field) > 0 {
			var next any
			next, err = navigateToField(current, field)
			if err != nil {
				switch err {
				case ErrFieldNotFound:
					return fmt.Errorf("%w: field '%s' in path '%s'", ErrFieldNotFound, field, path)
				case ErrInvalidPath:
					return fmt.Errorf("%w: invalid segment '%s' in path '%s'", ErrInvalidPath, field, path)
				}
				return err
			}
			if hasFilter {
				next, err = applyFilter(next, filter)
				if err != nil {
					return err
				}
			}
			current = next
		} else {
			return ErrInvalidPath
		}
	}
	// Now set the value on the final segment
	lastField, lastFilter, lastHasFilter := parseFieldAndFilter(parts[len(parts)-1])
	if lastHasFilter {
		// If the last segment has a filter, resolve to the filtered item and set value if it's a Pipeline
		final, err := navigateToField(current, lastField)
		if err != nil {
			return err
		}
		final, err = applyFilter(final, lastFilter)
		if err != nil {
			return err
		}
		if p, ok := final.(Pipeline); ok {
			return p.Set("", value) // Set at root of filtered Pipeline
		}
		return fmt.Errorf("final item is not a Pipeline, cannot set value")
	} else {
		// If the last segment is a field, set it if current is a Pipeline
		if p, ok := current.(Pipeline); ok {
			// Try to set the value, even if the field does not exist yet
			err := p.Set(lastField, value)
			if err != nil {
				// If the error is ErrFieldNotFound, create the field
				if err == ErrFieldNotFound {
					// Try to create the field by setting it
					return p.Set(lastField, value)
				}
				return err
			}
			return nil
		}
		return fmt.Errorf("final item is not a Pipeline, cannot set value")
	}
}
