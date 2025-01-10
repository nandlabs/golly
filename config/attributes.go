package config

// Interface Attributes is the interface that represents the attributes of a message
type Attributes interface {
	// Set adds a new attribute to the message
	Set(k string, v any)
	// Get returns the value of the attribute
	Get(k string) any
	// GetAsString returns the value of the attribute as a string
	GetAsString(k string) string
	// GetAsInt returns the value of the attribute as an int
	GetAsInt(k string) int
	// GetAsFloat returns the value of the attribute as a float
	GetAsFloat(k string) float64
	// GetAsBool returns the value of the attribute as a bool
	GetAsBool(k string) bool
	// GetAsBytes returns the value of the attribute as a byte array
	GetAsBytes(k string) []byte
	// GetAsArray returns the value of the attribute as an array
	GetAsArray(k string) []any
	// GetAsMap returns the value of the attribute as a map
	GetAsMap(k string) map[string]any
	// Remove removes the attribute from the message
	Remove(k string)
	// Keys returns the keys of the attributes
	Keys() []string
	// AsMap returns the attributes as a map
	AsMap() map[string]any
	// ThreadSafe makes the attributes thread safe
	ThreadSafe(bool)
	// IsThreadSafe returns true if the attributes are thread safe
	IsThreadSafe() bool
	// Merge merges this attribute with  another attributes
	Merge(m Attributes)
}
