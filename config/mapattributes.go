package config

import (
	"strconv"
	"sync"
)

// MapAttributes is a simple implementation of the Attributes interface that stores the attributes in a map
// It can be made thread safe by setting the threadSafe flag to true
// Default is not thread safe

type MapAttributes struct {
	attrs      map[string]any
	mutex      sync.RWMutex
	threadSafe bool
}

// Set adds a new attribute to the message
// if the attribute already exists it will be replaced
// if the threadsSafe flag is set to true, the method will lock the map
// if attrs map is empty in the struct, it will be created
func (m *MapAttributes) Set(k string, v any) {
	if m.threadSafe {
		m.mutex.Lock()
		defer m.mutex.Unlock()
	}
	if m.attrs == nil {
		m.attrs = make(map[string]any)
	}
	m.attrs[k] = v
}

// Get returns the value of the attribute
func (m *MapAttributes) Get(k string) any {
	if m.threadSafe {
		m.mutex.RLock()
		defer m.mutex.RUnlock()
	}
	if m.attrs == nil {
		return nil
	}
	return m.attrs[k]

}

// GetAsString returns the value of the attribute as a string
// if first checks if the attribute is a string, if not it tries to convert it to a string
// if it fails it returns an empty string
func (m *MapAttributes) GetAsString(k string) string {

	v := m.Get(k)
	if v != nil {
		return v.(string)
	}
	return ""

}

// GetAsInt returns the value of the attribute as an int
// if first checks if the attribute is an int, if not it tries to convert it to an int
// if it fails it returns 0
func (m *MapAttributes) GetAsInt(k string) int {
	v := m.Get(k)
	if v != nil {
		switch t := v.(type) {
		case int:
			return t
		case float32:
			return int(t)
		case float64:
			return int(t)
		case string:
			i, err := strconv.Atoi(t)
			if err == nil {
				return i
			}
		case bool:
			if t {
				return 1
			} else {
				return 0
			}

		}

	}
	return 0
}

// GetAsFloat returns the value of the attribute as a float
// if first checks if the attribute is a float, if not it tries to convert it to a float
// if it fails it returns 0
func (m *MapAttributes) GetAsFloat(k string) float64 {
	if v, ok := m.attrs[k].(float64); ok {
		return v
	}
	return 0
}

// GetAsBool returns the value of the attribute as a bool
// if first checks if the attribute is a bool, if not it tries to convert it to a bool
// if it fails it returns false
func (m *MapAttributes) GetAsBool(k string) bool {
	v := m.Get(k)

	if v != nil {
		switch t := v.(type) {
		case int:
			return t != 0
		case float32:
			return t != 0.0
		case float64:
			return t != 0.0
		case string:
			b, err := strconv.ParseBool(t)
			if err == nil {
				return b
			}
		case bool:
			return t

		}

	}
	return false
}

// GetAsBytes returns the value of the attribute as a byte array
// if first checks if the attribute is a byte array, if not it tries to convert it to a byte array
// if it fails it returns nil
func (m *MapAttributes) GetAsBytes(k string) []byte {
	v := m.Get(k)
	if v != nil {
		return v.([]byte)
	}

	return nil
}

// GetAsArray returns the value of the attribute as an array
// if first checks if the attribute is an array, if not it tries to convert it to an array
// if it fails it returns nil
func (m *MapAttributes) GetAsArray(k string) []any {
	v := m.Get(k)
	if v != nil {
		return v.([]any)
	}

	return nil
}

// GetAsMap returns the value of the attribute as a map
// if first checks if the attribute is a map, if not it tries to convert it to a map
// if it fails it returns nil
func (m *MapAttributes) GetAsMap(k string) map[string]any {
	v := m.Get(k)
	if v != nil {
		return v.(map[string]any)
	}
	return nil
}

// Remove removes the attribute from the message
func (m *MapAttributes) Remove(k string) {
	if m.threadSafe {
		m.mutex.Lock()
		defer m.mutex.Unlock()
	}
	if m.attrs != nil {
		delete(m.attrs, k)
	}
}

// Keys returns the keys of the attributes
func (m *MapAttributes) Keys() []string {
	if m.threadSafe {
		m.mutex.RLock()
		defer m.mutex.RUnlock()
	}
	if m.attrs == nil {
		//return an empty array
		return []string{}

	}
	keys := make([]string, 0, len(m.attrs))
	for k := range m.attrs {
		keys = append(keys, k)
	}
	return keys
}

// AsMap returns the attributes as a map.
// This is an expensive function as it creates a new map and copies all the attributes to it
func (m *MapAttributes) AsMap() map[string]any {
	if m.threadSafe {
		m.mutex.RLock()
		defer m.mutex.RUnlock()
	}
	if m.attrs == nil {
		//return an empty map
		return make(map[string]any)
	}
	newMap := make(map[string]any)
	for k, v := range m.attrs {
		newMap[k] = v
	}
	return newMap

}

// ThreadSafe makes the attributes thread safe
func (m *MapAttributes) ThreadSafe(ts bool) {
	m.threadSafe = ts
}

// IsThreadSafe returns true if the attributes are thread safe
func (m *MapAttributes) IsThreadSafe() bool {
	return m.threadSafe
}

// Merge merges this attributes with another attributes
func (m *MapAttributes) Merge(other Attributes) {
	if other == nil {
		return
	}
	if m.threadSafe {
		m.mutex.Lock()
		defer m.mutex.Unlock()
	}

	for _, k := range other.Keys() {
		m.Set(k, other.Get(k))
	}

}

// NewMapAttributes creates a new MapAttributes
// By default the attributes are  thread safe
// To make them thread safe call the MakeThreadSafe() method
func NewMapAttributes() *MapAttributes {
	return &MapAttributes{
		attrs:      make(map[string]any),
		threadSafe: false,
		mutex:      sync.RWMutex{},
	}
}
