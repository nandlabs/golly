// Package cli provides functionality for handling command-line flags.
package cli

// Flag represents a command-line flag.
type Flag struct {
	Name    string   // Name of the flag.
	Aliases []string // Aliases for the flag.
	Usage   string   // Usage information for the flag.
	Default string   // Default value for the flag.
}
