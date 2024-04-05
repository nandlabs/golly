// Package cli provides a command-line interface framework for building command-line applications in Go.
// This file contains the implementation of the help command and related functions.

package cli

import (
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
	"text/template"
)

// HelpFlags defines the flags used to display help.
var HelpFlags = [2]string{"--help", "-h"}

// PrintHelp is a function that prints the help information using the provided template and data.
var PrintHelp helpPrinter = printHelp

// PrintCustomHelp is a function that prints custom help information using the provided template, data, and custom functions.
var PrintCustomHelp helpCustomPrinter = printCustomHelp

// helpPrinter is a function type that defines the signature for printing help information.
type helpPrinter func(w io.Writer, template string, data interface{})

// helpCustomPrinter is a function type that defines the signature for printing custom help information.
type helpCustomPrinter func(w io.Writer, template string, data interface{}, customFunc map[string]interface{})

// helpCommand is the help command implementation.
var helpCommand = &Command{
	Name:      "help",
	Aliases:   []string{"-h", "--help"},
	Usage:     "Shows a list of commands or help for one command",
	ArgsUsage: "[command]",
	Action: func(conTxt *Context) error {
		args := conTxt.Args()
		argsPresent := args.First() != ""

		if conTxt.Command.Name == "help" || conTxt.Command.Name == "h" {
			conTxt = conTxt.parentContext
		}
		if argsPresent {
			return ShowCommandHelp(conTxt)
		}

		if conTxt.parentContext.App == nil {
			_ = ShowAppHelp(conTxt)
			return nil
		}

		return nil
	},
}

// ShowCommandHelp displays help information for a specific command.
func ShowCommandHelp(conTxt *Context) error {
	helpTemplate := CommandHelpTemplate
	PrintHelp(conTxt.App.writer(), helpTemplate, conTxt.Command)
	return nil
}

// ShowAppHelp displays help information for the entire application.
func ShowAppHelp(conTxt *Context) error {
	tpl := AppHelpTemplate
	printHelp(conTxt.App.writer(), tpl, conTxt.App)
	return nil
}

// printHelp is a helper function that prints the help information using the provided template and data.
func printHelp(out io.Writer, template string, data interface{}) {
	PrintCustomHelp(out, template, data, nil)
}

// printCustomHelp is a helper function that prints custom help information using the provided template, data, and custom functions.
func printCustomHelp(out io.Writer, templ string, data interface{}, customFuncs map[string]interface{}) {
	const maxLineLength = 1000

	funcMap := template.FuncMap{
		"join":           strings.Join,
		"subtract":       subtract,
		"indent":         indent,
		"nindent":        nindent,
		"trim":           strings.TrimSpace,
		"wrap":           func(input string, offset int) string { return wrap(input, offset, maxLineLength) },
		"offset":         offset,
		"offsetCommands": offsetCommands,
	}
	w := tabwriter.NewWriter(out, 1, 8, 2, ' ', 0)
	t := template.Must(template.New("help").Funcs(funcMap).Parse(templ))
	t.New("helpNameTemplate").Parse(helpNameTemplate)
	t.New("usageTemplate").Parse(usageTemplate)
	t.New("descriptionTemplate").Parse(descriptionTemplate)
	t.New("visibleCommandTemplate").Parse(visibleCommandTemplate)

	err := t.Execute(w, data)
	if err != nil {
		fmt.Println(err)
	}
	_ = w.Flush()
}

// subtract is a helper function that subtracts two integers.
func subtract(a, b int) int {
	return a - b
}

// indent is a helper function that indents a string with the specified number of spaces.
func indent(spaces int, v string) string {
	pad := strings.Repeat(" ", spaces)
	return pad + strings.Replace(v, "\n", "\n"+pad, -1)
}

// nindent is a helper function that indents a string with the specified number of spaces and adds a newline character.
func nindent(spaces int, v string) string {
	return "\n" + indent(spaces, v)
}

// wrap is a helper function that wraps a string to a specified line width with the specified offset and padding.
func wrap(input string, offset int, wrapAt int) string {
	var ss []string

	lines := strings.Split(input, "\n")
	padding := strings.Repeat(" ", offset)

	for i, line := range lines {
		if line == "" {
			ss = append(ss, line)
		} else {
			wrapped := wrapLine(line, offset, wrapAt, padding)
			if i == 0 {
				ss = append(ss, wrapped)
			} else {
				ss = append(ss, padding+wrapped)
			}
		}
	}
	return strings.Join(ss, "\n")
}

// wrapLine is a helper function that wraps a single line of text to a specified line width with the specified offset and padding.
func wrapLine(input string, offset int, wrapAt int, padding string) string {
	if wrapAt <= offset || len(input) <= wrapAt-offset {
		return input
	}

	lineWidth := wrapAt - offset
	words := strings.Fields(input)
	if len(words) == 0 {
		return input
	}

	wrapped := words[0]
	spaceLeft := lineWidth - len(wrapped)
	for _, word := range words[1:] {
		if len(word)+1 > spaceLeft {
			wrapped += "\n" + padding + word
			spaceLeft = lineWidth - len(word)
		} else {
			wrapped += " " + word
			spaceLeft -= 1 + len(word)
		}
	}
	return wrapped
}

// offset is a helper function that calculates the offset of a string with the specified fixed value.
func offset(input string, fixed int) int {
	return len(input) + fixed
}

// offsetCommands is a helper function that calculates the offset of commands with the specified fixed value.
func offsetCommands(cmds []*Command, fixed int) int {
	var max int = 0
	for _, cmd := range cmds {
		s := strings.Join(cmd.Names(), ", ")
		if len(s) > max {
			max = len(s)
		}
	}
	return max + fixed
}
