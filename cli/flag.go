// Package cli provides functionality for handling command-line flags.

package cli

import (
	"flag"
	"strings"
)

// mappedFlags is a map that stores the flags and their corresponding values.
var mappedFlags = make(map[string]interface{})

// flagMap is a map that stores the flags and their corresponding Flag objects.
var flagMap = make(map[string]*Flag)

// Flag represents a command-line flag.
type Flag struct {
	Name    string      // Name of the flag.
	Usage   string      // Usage description of the flag.
	Aliases []string    // Aliases for the flag.
	Default interface{} // Default value of the flag.
	Value   interface{} // Current value of the flag.
}

// HelpFlag is a built-in flag that represents the help flag.
var HelpFlag = &Flag{
	Name:    "help",
	Usage:   "show help",
	Aliases: []string{"-h", "--help"},
	Default: "",
}

// hasFlag checks if a flag exists in a list of flags.
func hasFlag(flags []*Flag, flag *Flag) bool {
	for _, exist := range flags {
		if flag == exist {
			return true
		}
	}
	return false
}

// setFlags sets the flags based on the commandFlags and inputFlags.
func setFlags(commandFlags []*Flag, inputFlags []string) {
	parsedFlags := parseFlags(commandFlags, inputFlags)
	for _, f := range parsedFlags {
		if f.Name == "help" {
			f.AddHelpFlag()
		} else {
			f.AddFlagToSet()
		}
	}
}

// AddFlagToSet adds the flag to the flag set.
func (f *Flag) AddFlagToSet() {
	flag.String(f.Name, f.Value.(string), f.Usage)
}

// AddHelpFlag adds the help flag to the flag set.
func (f *Flag) AddHelpFlag() {
	flag.Bool(f.Name, true, f.Usage)
}

// parseFlags parses the inputFlags and returns the corresponding Flag objects.
func parseFlags(commandFlags []*Flag, inputFlags []string) []*Flag {
	createFlagMap(commandFlags)
	var result []*Flag
	for _, item := range inputFlags {
		itemArr := strings.Split(item, "=")
		if len(itemArr) > 1 {
			key := itemArr[0]
			val := itemArr[1]
			mappedFlag := flagMap[key]
			result = append(result, &Flag{
				Name:    mappedFlag.Name,
				Usage:   mappedFlag.Usage,
				Aliases: nil,
				Default: mappedFlag.Default,
				Value:   val,
			})
		}
	}
	return result
}

// createFlagMap creates a map of aliases to flags.
func createFlagMap(commandFlags []*Flag) {
	for _, item := range commandFlags {
		for _, alias := range item.Aliases {
			flagMap[alias] = &Flag{
				Name:    item.Name,
				Usage:   item.Usage,
				Aliases: nil,
				Default: item.Default,
				Value:   nil,
			}
		}
	}
}
