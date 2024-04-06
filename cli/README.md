## cli

This project is a Go library for building powerful and user-friendly command-line interfaces (CLIs). The library makes it easy to create and manage complex command structures, parse command-line arguments, and provide helpful error messages and usage information to the user.

---
- [Installation](#installation)
- [Features](#features)
- [Usage](#usage)
  - [Default Usage](#default)
  - [Subcommand Usage](#subcommands)
  - [Flags Usage](#flags)
---

### Installation

```bash
go get oss.nandlabs.io/golly/cli
```

### Features

* Easy to use API for building complex command structures 
* Argument parsing and validation 
* Automatically generates usage and help information 
* Written in Go and follows best practices for Go programming

### Usage

#### Default
```go
package main

import (
	"fmt"
	"log"
	"os"

	"oss.nandlabs.io/golly/cli"
)

func main() {
	app := &cli.App{
		Version: "v0.0.1",
		Action: func(ctx *cli.Context) error {
			fmt.Printf("Hello, Golang!")
			return nil
		},
	}

	if err := app.Execute(os.Args); err != nil {
		log.Fatal(err)
	}
}
```

CLI Command and Output
```shell
~ % go run main.go greet
Hello, Golang!
```

#### Subcommands

```go
package main

import (
	"fmt"
	"log"
	"os"

	"oss.nandlabs.io/golly/cli"
)

func main() {
	app := &cli.App{
		Version: "v0.0.1",
		Action: func(ctx *cli.Context) error {
			fmt.Printf("Hello, Golang!")
			return nil
		},
		Commands: []*cli.Command{
			{
				Name:    "test",
				Usage:   "this is a test command",
				Aliases: []string{"t"},
				Action: func(ctx *cli.Context) error {
					fmt.Println("hello from test command")
					return nil
				},
			},
			{
				Name:    "run",
				Usage:   "time to run",
				Aliases: []string{"r"},
				Action: func(ctx *cli.Context) error {
					fmt.Println("time to run away")
					return nil
				},
				Commands: []*cli.Command{
					{
						Name:  "slow",
						Usage: "run slow",
						Action: func(ctx *cli.Context) error {
							fmt.Println("time to run slow")
							return nil
						},
					},
					{
						Name:  "fast",
						Usage: "run fast",
						Action: func(ctx *cli.Context) error {
							fmt.Println("time to run fast")
							return nil
						},
					},
				},
			},
		},
	}

	if err := app.Execute(os.Args); err != nil {
		log.Fatal(err)
	}
}
```

CLI Commands and Output
```shell
~ % go run main.go test
hello from test command
```
```shell
~ % go run main.go run
time to run away
```
```shell
~ % go run main.go run fast
time to run fast
```

#### Flags

```go
package main

import (
	"fmt"
	"log"
	"os"

	"oss.nandlabs.io/golly/cli"
)

const (
	ProjectDir  = "pd"
	ProfileFile = "pf"
)

func main() {
	app := &cli.App{
		Version: "v0.0.1",
		Action: func(ctx *cli.Context) error {
			fmt.Printf("Hello, Golang!\n")
			fmt.Println(ctx.GetFlag(ProjectDir))
			fmt.Println(ctx.GetFlag(ProfileFile))
			return nil
		},
		Commands: []*cli.Command{
			{
				Name:    "test",
				Usage:   "this is a test command",
				Aliases: []string{"t"},
				Action: func(ctx *cli.Context) error {
					fmt.Println("hello from test command")
					fmt.Println(ctx.GetFlag(ProjectDir))
					fmt.Println(ctx.GetFlag(ProfileFile))
					return nil
				},
			},
			{
				Name:    "run",
				Usage:   "time to run",
				Aliases: []string{"r"},
				Action: func(ctx *cli.Context) error {
					fmt.Println("time to run away")
					fmt.Println(ctx.GetFlag(ProjectDir))
					fmt.Println(ctx.GetFlag(ProfileFile))
					return nil
				},
				Commands: []*cli.Command{
					{
						Name:  "slow",
						Usage: "run slow",
						Action: func(ctx *cli.Context) error {
							fmt.Println("time to run slow")
							fmt.Println(ctx.GetFlag(ProjectDir))
							fmt.Println(ctx.GetFlag(ProfileFile))
							return nil
						},
					},
					{
						Name:  "fast",
						Usage: "run fast",
						Action: func(ctx *cli.Context) error {
							fmt.Println("time to run fast")
							fmt.Println(ctx.GetFlag(ProjectDir))
							fmt.Println(ctx.GetFlag(ProfileFile))
							return nil
						},
					},
				},
			},
		},
		// global app flags
		Flags: []*cli.Flag{
			{
				Name:    ProjectDir,
				Aliases: []string{"pd"},
				Default: "",
				Usage:   "Directory of the project to be built",
			},
			{
				Name:    ProfileFile,
				Aliases: []string{"pf"},
				Default: "",
				Usage:   "Profile file name to be used",
			},
		},
	}

	if err := app.Execute(os.Args); err != nil {
		log.Fatal(err)
	}
}
```

CLI Commands and Output
```shell
~ % go run main.go test -pd="test" -pf="dev"
Hello, Golang!
test
dev
```
```shell
~ % go run main.go run -pd="test" -pf="dev"
time to run away
test
dev
```
```shell
~ % go run main.go run fast -pd="test" -pf="dev"
time to run fast
test
dev
```