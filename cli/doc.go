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
//  func main() {
// 	app := cli.NewCLI()

// 	server := &cli.Command{
// 		Name:        "server",
// 		Description: "Server command",
// 		Handler: func(ctx *cli.Context) error {
// 			region, _ := ctx.GetFlag("region")
// 			fmt.Printf("IN REGION, %s\n", region)
// 			return nil
// 		},
// 		Flags: []cli.Flag{
// 			{
// 				Name:    "region",
// 				Aliases: []string{"r"},
// 				Usage:   "Provide region",
// 				Default: "us-east-1",
// 			},
// 		},
// 		SubCommands: map[string]*cli.Command{
// 			"create": {
// 				Name:        "create",
// 				Description: "create",
// 				Handler: func(ctx *cli.Context) error {
// 					typ, _ := ctx.GetFlag("type")
// 					fmt.Printf("SERVER TYPE %s\n", typ)
// 					return nil
// 				},
// 				Flags: []cli.Flag{
// 					{
// 						Name:    "type",
// 						Aliases: []string{"t"},
// 						Usage:   "server type",
// 						Default: "t2.micro",
// 					},
// 				},
// 			},
// 		},
// 	}

// 	app.AddCommand(server)

//		if err := app.Execute(); err != nil {
//			fmt.Println("Error:", err)
//		}
//	}
//
// For more information and examples, please refer to the package documentation at:
//
//	https://github.com/nandlabs/golly/cli
package cli
