package data

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

var ErrInvalidType = errors.New("invalid type")
var ErrKeyNotFound = errors.New("key not found")
var ErrInvalidPath = errors.New("invalid path")
var ErrFieldNotFound = errors.New("field not found")

// Pipeline is a struct that represents a data processing pipeline.
// It contains a map to store key-value pairs, where keys are strings and values can be

const (
	InstanceIdKey = "__instanceId__"
	ErrorKey      = "__error__"
)

type Pipeline interface {
	// Id returns the unique identifier of the pipeline instance.
	Id() string
	// Get retrieves the value associated with the given key from the pipeline.
	// If the key is not found, it returns an ErrKeyNotFound error.
	Get(key string) (value any, err error)
	// Has checks if the given key exists in the pipeline's data.
	Has(key string) bool
	// Keys returns a slice of all keys present in the pipeline's data.
	Keys() []string
	// Set assigns the given value to the specified key in the pipeline's data.
	// If the key already exists, its value will be updated.
	Set(key string, value any) error
	// Delete removes the entry with the specified key from the pipeline's data.
	// If the key does not exist, it does nothing and returns nil.
	Delete(key string) error
	// Map returns a copy of the pipeline's data as a map[string]any.
	// This map contains all key-value pairs stored in the pipeline.
	Map() map[string]any
	// GetError retrieves the error value from the pipeline instance.
	// It uses the ExtractValue function to obtain the value associated with the ErrorKey.
	GetError() (errMsg string)
	// SetError assigns the provided error message to the pipeline instance.
	// It uses the Set function to assign the error message to the ErrorKey.
	SetError(errMsg string)
	// MergeFrom combines the key-value pairs from the provided map into the pipeline's data.
	// If a key already exists in the pipeline, its value will be updated with the new
	// value.
	MergeFrom(data map[string]any) error
	// Merge combines the key-value pairs from the provided pipeline into the current pipeline.
	// If a key already exists in the current pipeline, its value will be updated with the
	// new value.
	Merge(pipeline Pipeline) error
	// Clone creates a deep copy of the current pipeline instance.
	// It returns a new pipeline instance with a duplicated map containing
	// the same key-value pairs as the original pipeline.
	Clone() Pipeline
}

// MapPipeline represents a pipeline that processes data stored in a map.
// The data is stored as key-value pairs where the key is a string and the value can be of any type.
type MapPipeline struct {
	data map[string]any
}

// NewPipeline creates a new instance of a Pipeline with the given ID.
// It initializes the pipeline's data map and sets the InstanceIdKey to the provided ID.
//
// Parameters:
//   - id: A string representing the unique identifier for the pipeline instance.
//
// Returns:
//   - pipeline: A Pipeline instance with the specified ID.
func NewPipeline(id string) (pipeline Pipeline) {
	pipeline = &MapPipeline{
		data: make(map[string]any),
	}
	pipeline.Set(InstanceIdKey, id)

	return
}

// NewPipelineFrom creates a new instance of a Pipeline with the given ID and initial values.
// It initializes the pipeline's data map and sets the provided values.
// Additionally, it sets the InstanceIdKey to the provided ID.
//
// Parameters:
//   - values: A map containing initial key-value pairs to be set in the pipeline.
//
// Returns:
//
//	A Pipeline instance with the specified ID and initial values.
func NewPipelineFrom(values map[string]any) (pipeline Pipeline) {
	pipeline = &MapPipeline{

		data: make(map[string]any),
	}
	for k, v := range values {
		pipeline.Set(k, v)
	}
	return
}

// Id returns the instance ID of the Pipeline.
// It extracts the ID value from the Pipeline using the InstanceIdKey.
func (p *MapPipeline) Id() (id string) {
	id, _ = ExtractValue[string](p, InstanceIdKey)
	return
}

// // StepId retrieves the step identifier from the Pipeline instance.
// // It uses the ExtractValue function to obtain the value associated with the StepIdKey.
// // Returns the step identifier as a string.
// func (p *Pipeline) StepId() (stepId string) {
// 	stepId, _ = ExtractValue[string](p, StepIdKey)
// 	return
// }

// Get retrieves the value associated with the given key from the Pipeline.
// If the key is found, the value is returned along with a nil error.
// If the key is not found, an ErrKeyNotFound error is returned.
//
// Parameters:
//   - key: The key to look up in the Pipeline.
//
// Returns:
//   - value: The value associated with the key, if found.
//   - err: An error indicating whether the key was found or not.
func (p *MapPipeline) Get(key string) (value any, err error) {

	if v, ok := p.data[key]; ok {
		value = v
	} else {
		err = ErrKeyNotFound
	}
	return
}

// Has checks if the given key exists in the Pipeline's data map.
// It returns true if the key is present, otherwise false.
//
// Parameters:
//
//	key - the key to be checked in the data map.
//
// Returns:
//
//	bool - true if the key exists, false otherwise.
func (p *MapPipeline) Has(key string) bool {
	_, ok := p.data[key]
	return ok
}

// Keys returns a slice of all the keys present in the Pipeline's data.
// It iterates over the map and collects each key into a slice, which is then returned.
func (p *MapPipeline) Keys() []string {

	keys := make([]string, 0, len(p.data))
	for k := range p.data {
		keys = append(keys, k)
	}
	return keys
}

// Set assigns the given value to the specified key in the Pipeline's data map.
// If the key already exists, its value will be updated.
//
// Parameters:
//
//	key: The key to which the value should be assigned.
//	value: The value to be assigned to the specified key.
//
// Returns:
//
//	An error if the operation fails, otherwise nil.
func (p *MapPipeline) Set(key string, value any) error {

	p.data[key] = value
	return nil
}

// Delete removes the entry with the specified key from the Pipeline's data.
// If the key does not exist, the function does nothing and returns nil.
//
// Parameters:
//
//	key - The key of the entry to be deleted.
//
// Returns:
//
//	An error if the deletion fails, otherwise nil.
func (p *MapPipeline) Delete(key string) error {
	delete(p.data, key)
	return nil
}

// Map creates and returns a new map with the same key-value pairs as the
// Pipeline's internal data. The returned map has keys of type string and
// values of type any.
func (p *MapPipeline) Map() map[string]any {
	data := make(map[string]any, len(p.data))
	for k, v := range p.data {
		data[k] = v
	}
	return data
}

// GetError retrieves the error value from the Pipeline instance.
// It uses the ExtractValue function to obtain the value associated with the ErrorKey.
// Returns the error message as a string.
func (p *MapPipeline) GetError() (errMsg string) {
	errMsg, _ = ExtractValue[string](p, ErrorKey)
	return
}

// SetError assigns the provided error message to the Pipeline instance.
// It uses the Set function to assign the error message to the ErrorKey.
//
// Parameters:
//   - errMsg: A string representing the error message to be assigned.
func (p *MapPipeline) SetError(errMsg string) {
	p.Set(ErrorKey, errMsg)
}

// MergeFrom combines the key-value pairs from the provided map into the Pipeline's data.
// If a key already exists in the Pipeline, its value will be updated with the new value.
//
// Parameters:
//   - data: A map containing key-value pairs to be merged into the Pipeline.
//
// Returns:
//
//	An error if the merge operation fails, otherwise nil.
func (p *MapPipeline) MergeFrom(data map[string]any) error {
	for k, v := range data {
		p.data[k] = v
	}
	return nil
}

// Merge combines the key-value pairs from the provided Pipeline into the current Pipeline.
// If a key already exists in the current Pipeline, its value will be updated with the new value.
// Parameters:
//   - pipeline: A Pipeline instance containing key-value pairs to be merged into the current Pipeline.
//
// Returns:
//
//	An error if the merge operation fails, otherwise nil.
func (p *MapPipeline) Merge(pipeline Pipeline) error {
	// for k, v := range pipeline.data {
	// 	p.data[k] = v
	// }
	for _, key := range pipeline.Keys() {
		value, err := pipeline.Get(key)
		if err != nil {
			p.Set(key, value)
		}
	}

	return nil
}

// Clone creates a deep copy of the current Pipeline instance.
// It returns a new Pipeline instance with a duplicated map containing
// the same key-value pairs as the original Pipeline.
func (p *MapPipeline) Clone() Pipeline {
	data := make(map[string]any, len(p.data))
	for k, v := range p.data {
		data[k] = v
	}
	return &MapPipeline{
		data: data,
	}
}

// evaluateCondition evaluates a condition string using variables from the context.
// It supports basic comparison and logical operators: ==, !=, <, >, <=, >=, &&, ||.
func EvaluateCondition(p Pipeline, condition string) (bool, error) {
	// Tokenize the input condition string into smaller parts (e.g., variables, operators).
	tokens := tokenize(condition)

	// Convert the infix tokens (e.g., "a == b && c") into postfix notation (Reverse Polish Notation).
	postfix, err := infixToPostfix(tokens)
	if err != nil {
		return false, err
	}

	// Evaluate the postfix expression using the workflow context.
	return evaluatePostfix(postfix, p)
}

// tokenize splits the condition string into tokens for parsing.
func tokenize(condition string) []string {
	var tokens []string
	var currentToken strings.Builder

	// Check if a character is part of an operator
	isOperator := func(r rune) bool {
		return strings.ContainsRune("!=<>&|()", r)
	}

	// Iterate through each character in the condition
	for _, ch := range condition {
		switch {
		case ch == ' ': // Skip spaces
			if currentToken.Len() > 0 {
				tokens = append(tokens, currentToken.String())
				currentToken.Reset()
			}
		case isOperator(ch): // If the character is an operator, split the token
			if currentToken.Len() > 0 {
				tokens = append(tokens, currentToken.String())
				currentToken.Reset()
			}
			tokens = append(tokens, string(ch))
		default: // Add characters to the current token
			currentToken.WriteRune(ch)
		}
	}
	// Append the last token, if any
	if currentToken.Len() > 0 {
		tokens = append(tokens, currentToken.String())
	}
	return tokens
}

// precedence determines the precedence of operators.
func precedence(op string) int {
	switch op {
	case "||": // Logical OR has the lowest precedence
		return 1
	case "&&": // Logical AND has higher precedence than OR
		return 2
	case "==", "!=", "<", ">", "<=", ">=": // Comparison operators
		return 3
	case "(": // Parentheses have no precedence themselves
		return 0
	default: // Unknown operators
		return -1
	}
}

// infixToPostfix converts an infix expression (e.g., "a == b && c") to postfix (Reverse Polish Notation).
func infixToPostfix(tokens []string) ([]string, error) {
	var postfix []string
	var stack []string

	// Process each token
	for _, token := range tokens {
		switch token {
		case "&&", "||", "==", "!=", "<", ">", "<=", ">=": // If the token is an operator
			// Pop higher or equal precedence operators from the stack to postfix
			for len(stack) > 0 && precedence(stack[len(stack)-1]) >= precedence(token) {
				postfix = append(postfix, stack[len(stack)-1])
				stack = stack[:len(stack)-1]
			}
			stack = append(stack, token) // Push the current operator onto the stack
		case "(": // Push opening parentheses onto the stack
			stack = append(stack, token)
		case ")": // Process closing parentheses
			// Pop from the stack until an opening parenthesis is encountered
			for len(stack) > 0 && stack[len(stack)-1] != "(" {
				postfix = append(postfix, stack[len(stack)-1])
				stack = stack[:len(stack)-1]
			}
			// Check for mismatched parentheses
			if len(stack) == 0 {
				return nil, errors.New("mismatched parentheses")
			}
			stack = stack[:len(stack)-1] // Pop the opening parenthesis
		default: // Otherwise, it's an operand (variable or literal)
			postfix = append(postfix, token)
		}
	}

	// Pop any remaining operators from the stack
	for len(stack) > 0 {
		if stack[len(stack)-1] == "(" {
			return nil, errors.New("mismatched parentheses")
		}
		postfix = append(postfix, stack[len(stack)-1])
		stack = stack[:len(stack)-1]
	}

	return postfix, nil
}

// evaluatePostfix evaluates a postfix expression using the context.
func evaluatePostfix(postfix []string, pipeline Pipeline) (bool, error) {
	var stack []any

	// Helper function to compare two operands with an operator
	compare := func(a, b any, op string) (bool, error) {
		switch op {
		case "==": // Equality
			return a == b, nil
		case "!=": // Inequality
			return a != b, nil
		case "<": // Less than
			return a.(float64) < b.(float64), nil
		case ">": // Greater than
			return a.(float64) > b.(float64), nil
		case "<=": // Less than or equal
			return a.(float64) <= b.(float64), nil
		case ">=": // Greater than or equal
			return a.(float64) >= b.(float64), nil
		default:
			return false, fmt.Errorf("unknown operator: %s", op)
		}
	}

	// Process each token in the postfix expression
	for _, token := range postfix {
		switch token {
		case "&&", "||": // Logical operators
			if len(stack) < 2 {
				return false, errors.New("invalid logical expression")
			}
			// Pop two operands
			b := stack[len(stack)-1].(bool)
			a := stack[len(stack)-2].(bool)
			stack = stack[:len(stack)-2]

			// Perform the logical operation
			if token == "&&" {
				stack = append(stack, a && b)
			} else {
				stack = append(stack, a || b)
			}
		case "==", "!=", "<", ">", "<=", ">=": // Comparison operators
			if len(stack) < 2 {
				return false, errors.New("invalid comparison expression")
			}
			// Pop two operands
			b := stack[len(stack)-1]
			a := stack[len(stack)-2]
			stack = stack[:len(stack)-2]

			// Perform the comparison
			result, err := compare(a, b, token)
			if err != nil {
				return false, err
			}
			stack = append(stack, result)
		default: // Operands (variables or literals)
			value, err := parseToken(token, pipeline)
			if err != nil {
				return false, err
			}
			stack = append(stack, value)
		}
	}

	// There should be exactly one result left on the stack
	if len(stack) != 1 {
		return false, errors.New("invalid postfix expression")
	}
	return stack[0].(bool), nil
}

// parseToken converts a token into a value using the context.
func parseToken(token string, pipeline Pipeline) (any, error) {
	// If the token is a string literal (surrounded by quotes)
	if strings.HasPrefix(token, `"`) && strings.HasSuffix(token, `"`) {
		return token[1 : len(token)-1], nil
	}

	// If the token is a number, parse it
	if value, err := strconv.ParseFloat(token, 64); err == nil {
		return value, nil
	}

	// Otherwise, assume it's a variable in the context
	if pipeline.Has(token) {
		return pipeline.Get(token)
	}

	return nil, fmt.Errorf("unknown variable or value: %s", token)
}

// ExtractValue retrieves a value of type T from the Pipeline using the provided path.
// If the value is not of type T, it returns an ErrInvalidType error.
// The path can be a simple key or a dot-separated path (e.g., "user.address.city").
//
// Parameters:
//   - c: A pointer to the Pipeline from which to extract the value.
//   - path: The path to the value to be retrieved. Can be a simple key or dot notation (e.g., "user.address.city").
//
// Returns:
//   - value: The value of type T associated with the provided path.
//   - err: An error if the path does not exist or the value is not of type T.
func ExtractValue[T any](c Pipeline, path string) (value T, err error) {
	// If the path doesn't contain dots, use the simple key lookup
	if !strings.Contains(path, ".") {
		var v any
		v, err = c.Get(path)
		if err != nil {
			return
		}

		return Convert[T](v)
	}

	// For dot notation paths, split the path and navigate through the structure
	parts := strings.Split(path, ".")
	rootKey := parts[0]

	var current any
	current, err = c.Get(rootKey)
	if err != nil {
		return
	}

	// Navigate through the nested structure
	for i := 1; i < len(parts); i++ {
		if len(parts[i]) > 0 {
			current, err = navigateToField(current, parts[i])
		} else {
			err = ErrInvalidPath
		}
		if err != nil {
			// Wrap the error with more context
			switch err {
			case ErrFieldNotFound:
				err = fmt.Errorf("%w: field '%s' in path '%s'", ErrFieldNotFound, parts[i], path)
			case ErrInvalidPath:
				err = fmt.Errorf("%w: invalid segment '%s' in path '%s'", ErrInvalidPath, parts[i], path)
			}
			return
		}
	}

	// Try to convert the final value to type T
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

// Wrap initializes a new Pipeline with the provided ID and values.
// It sets the InstanceIdKey in the pipeline with the given ID.
//
// Parameters:
//   - id: A string representing the unique identifier for the pipeline instance.
//   - values: A map containing key-value pairs to initialize the pipeline data.
//
// Returns:
//   - pipeline: A Pipeline interface initialized with the provided values and ID.
func Wrap(id string, values map[string]any) (pipeline *MapPipeline) {
	pipeline = &MapPipeline{
		data: values,
	}
	pipeline.Set(InstanceIdKey, id)
	return
}
