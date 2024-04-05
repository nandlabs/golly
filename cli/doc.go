// Package cli provides a command-line interface (CLI) framework for Go applications.
//
// This package offers a set of utilities and abstractions to build command-line interfaces
// with ease. It includes features such as command parsing, flag handling, and subcommand support.
//
// Usage:
// To use this package, import it in your Go code:
//
//	import "github.com/nandlabs/golly/cli"
//
// Example:
// Here's a simple example that demonstrates how to use the `cli` package:
//
//	package main
//
//	import (
//	    "fmt"
//	    "github.com/nandlabs/golly/cli"
//	)
//
//	func main() {
//	    app := cli.NewApp()
//	    app.Name = "myapp"
//	    app.Usage = "A simple CLI application"
//
//	    app.Action = func(c *cli.Context) error {
//	        fmt.Println("Hello, World!")
//	        return nil
//	    }
//
//	    err := app.Run(os.Args)
//	    if err != nil {
//	        fmt.Println(err)
//	        os.Exit(1)
//	    }
//	}
//
// For more information and examples, please refer to the package documentation at:
//
//	https://github.com/nandlabs/golly/cli
package cli
