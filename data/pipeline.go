package data

import (
	"errors"
)

var ErrInvalidType = errors.New("invalid type")
var ErrKeyNotFound = errors.New("key not found")
var ErrInvalidPath = errors.New("invalid path")
var ErrFieldNotFound = errors.New("field not found")

// Pipeline is a struct that represents a data processing pipeline.
// It contains a map to store key-value pairs, where keys are strings and values can be

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
	GetError() (err error)
	// HasError checks if there is an error associated with the pipeline instance.
	// It returns true if an error exists, otherwise false.
	HasError() bool
	// SetError assigns the provided error message to the pipeline instance.
	// It uses the Set function to assign the error message to the ErrorKey.
	SetError(err error)
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
	id   string
	err  error          // Instance ID of the pipeline
	data map[string]any // The data map holds the key-value pairs for the pipeline.
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
		id:   id,
		err:  nil,
		data: make(map[string]any),
	}

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

	return p.id
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
func (p *MapPipeline) GetError() (err error) {

	return p.err
}

// HasError checks if there is an error associated with the Pipeline instance.
// It returns true if an error exists, otherwise false.
func (p *MapPipeline) HasError() bool {
	return p.err != nil
}

// SetError assigns the provided error message to the Pipeline instance.
// It uses the Set function to assign the error message to the ErrorKey.
//
// Parameters:
//   - errMsg: A string representing the error message to be assigned.
func (p *MapPipeline) SetError(err error) {
	p.err = err
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
	for _, key := range pipeline.Keys() {
		value, err := pipeline.Get(key)
		if err == nil {
			p.Set(key, value)
		}
	}
	return nil
}

// Clone creates a deep copy of the current Pipeline instance.
// It returns a new Pipeline instance with a duplicated map containing
// the same key-value pairs as the original Pipeline.
func (p *MapPipeline) Clone() Pipeline {
	// Create a new MapPipeline instance
	clone := &MapPipeline{
		id:   p.id,
		err:  p.err,
		data: make(map[string]any, len(p.data)),
	}

	// Copy the data from the original pipeline to the clone
	for k, v := range p.data {
		clone.data[k] = v
	}

	return clone
}
