// Package main demonstrates the CLI framework for building command-line applications.
package main

import (
	"fmt"
	"os"

	"oss.nandlabs.io/golly/cli"
)

func main() {
	// Create a new CLI application
	app := cli.NewCLI()
	app.AddVersion("1.0.0")

	// Add a "greet" command
	greetCmd := cli.NewCommand("greet", "Greet a user", "1.0.0", func(ctx *cli.Context) error {
		name, ok := ctx.GetFlag("name")
		if !ok {
			name = "World"
		}
		fmt.Printf("Hello, %s!\n", name)
		return nil
	})
	greetCmd.Flags = []*cli.Flag{
		{Name: "name", Usage: "Name to greet", Default: "World"},
	}

	// Add a "version" command
	versionCmd := cli.NewCommand("version", "Show version", "1.0.0", func(ctx *cli.Context) error {
		fmt.Println("golly-cli v1.0.0")
		return nil
	})

	// Add a "math" command with subcommands
	mathCmd := cli.NewCommand("math", "Math operations", "1.0.0", nil)
	addCmd := cli.NewCommand("add", "Add two numbers", "1.0.0", func(ctx *cli.Context) error {
		a, _ := ctx.GetFlag("a")
		b, _ := ctx.GetFlag("b")
		fmt.Printf("%s + %s\n", a, b)
		return nil
	})
	addCmd.Flags = []*cli.Flag{
		{Name: "a", Usage: "First number"},
		{Name: "b", Usage: "Second number"},
	}
	mathCmd.AddSubCommand(addCmd)

	app.AddCommand(greetCmd)
	app.AddCommand(versionCmd)
	app.AddCommand(mathCmd)

	// Execute — reads os.Args automatically
	if err := app.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
