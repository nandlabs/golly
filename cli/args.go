// Package cli provides functionality for handling command-line arguments.
package cli

import (
	"strings"
)

// Args is an interface that defines methods for retrieving command-line arguments.
type Args interface {
	Get(n int) string
	First() string
	FetchArgs() *ArgsCli
}

// args is a type that represents command-line arguments.
type args []string

// ArgsCli is a struct that holds the parsed command-line arguments.
type ArgsCli struct {
	inputCommands []string
	inputFlags    []string
}

// Get returns the nth command-line argument.
func (a *args) Get(n int) string {
	if len(*a) > n {
		return (*a)[n]
	}
	return ""
}

// First returns the first command-line argument.
func (a *args) First() string {
	return a.Get(0)
}

// FetchArgs parses the command-line arguments and returns an ArgsCli object.
func (a *args) FetchArgs() *ArgsCli {
	var outputCommands []string
	var outputFlags []string
	var tail []string
	if len(*a) >= 2 {
		tail = (*a)[1:]
	}
	for _, item := range tail {
		if isFlag(item) {
			trimmedItem := strings.TrimPrefix(strings.TrimPrefix(item, "-"), "--")
			outputFlags = append(outputFlags, trimmedItem)
		} else {
			outputCommands = append(outputCommands, item)
		}
	}
	return &ArgsCli{
		inputCommands: outputCommands,
		inputFlags:    outputFlags,
	}
}

// isFlag checks if the given string is a command-line flag.
func isFlag(item string) bool {
	if strings.HasPrefix(item, "-") || strings.HasPrefix(item, "--") {
		return true
	}
	return false
}

// checkForHelp checks if the "-help" or "-h" flag is present in the command-line arguments.
func (a *args) checkForHelp() (isPresent bool) {
	programArgs := (*a)[1:]
	if len(programArgs) > 0 {
		lastItem := programArgs[len(programArgs)-1]
		if lastItem == "-help" || lastItem == "-h" {
			isPresent = true
		}
	}

	return
}
