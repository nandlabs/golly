// Package cli provides functionality for handling command-line flags.

package cli

type Flag struct {
	Name    string
	Aliases []string
	Usage   string
	Default string
}
