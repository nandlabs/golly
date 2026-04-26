package data

// Error is a struct that represents an error
type Error struct {
	//Code is the error code
	Code string `json:"code" yaml:"code"`
	//Message is the error message
	Message string `json:"message" yaml:"message"`
	//Details is the error details
	// This is an optional field
	Details string `json:"details,omitempty" yaml:"details,omitempty"`
}
