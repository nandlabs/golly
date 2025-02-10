# cli

This project is a Go library for building powerful and user-friendly command-line interfaces (CLIs). The library makes it easy to create and manage complex command structures, parse command-line arguments, and provide helpful error messages and usage information to the user.

---

- [Installation](#installation)
- [Features](#features)
- [Usage](#usage)
  - [Default Usage](#default)
  - [Subcommand Usage](#subcommands)
  - [Flags Usage](#flags)

---

## Installation

```bash
go get oss.nandlabs.io/golly/cli
```

## Features

- Easy to use API for building complex command structures
- Argument parsing and validation
- Automatically generates usage and help information
- Written in Go and follows best practices for Go programming

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
    app := cli.NewCLI()

    gollyCmd := &cli.Command{
        Name:        "welcome",
        Description: "Welcome command",
        Handler: func(ctx *cli.Context) error {
            fmt.Println("welcome to golly!")
            return nil
        },
    }

    app.AddCommand(gollyCmd)

    if err := app.Execute(); err != nil {
        log.Fatal(err)
    }
}
```

CLI Command and Output

```shell
~ % go run main.go welcome
welcome to golly!
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

    app := cli.NewCLI()

    welcomeCmd := &cli.Command{
        Name: "welcome",
        Description: "Welcome to golly",
        Handler: func(ctx *cli.Context) error {
            fmt.Println("welcome to golly")
            return nil
        },
        SubCommands: map[string]*cli.Command{
            "home": {
                Name:        "home",
                Description: "welcome home",
                Handler: func(ctx *cli.Context) error {
                    fmt.Println("welcome home")
                    return nil
                },
            },
            "office": {
                Name:        "level",
                Description: "level of the skill",
                Handler: func(ctx *cli.Context) error {
                    fmt.Println("welcome office")
                    return nil
                },
            },
        },
    }

    app.AddCommand(welcomeCmd)

    if err := app.Execute(); err != nil {
        fmt.Println("Error:", err)
    }
}
```

CLI Commands and Output

```shell
~ % go run main.go welcome
welcome to golly
```

```shell
~ % go run main.go welcome home
welcome home
```

```shell
~ % go run main.go welcome office
welcome office
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

func main() {
    app := cli.NewCLI()

    server := &cli.Command{
        Name:        "server",
        Description: "Server command",
        Handler: func(ctx *cli.Context) error {
        region, _ := ctx.GetFlag("region")
            fmt.Printf("IN REGION, %s\n", region)
            return nil
        },
        Flags: []cli.Flag{
            {
                Name:    "region",
                Aliases: []string{"r"},
                Usage:   "Provide region",
                Default: "",
            },
        },
        SubCommands: map[string]*cli.Command{
            "create": {
                Name:        "create",
                Description: "create",
                Handler: func(ctx *cli.Context) error {
                    typ, _ := ctx.GetFlag("type")
                    fmt.Printf("SERVER TYPE %s\n", typ)
                    return nil
                },
                Flags: []cli.Flag{
                    {
                        Name:    "type",
                        Aliases: []string{"t"},
                        Usage:   "server type",
                        Default: "",
                    },
                },
            },
        },
    }

    app.AddCommand(server)

    if err := app.Execute(); err != nil {
        fmt.Println("Error:", err)
    }
}
```

CLI Commands and Output

```shell
~ % go run main.go server --region="us-east-1"
IN REGION, us-east-1
```

```shell
~ % go run main.go server create --type="t3.medium"
SERVER TYPE t3.medium
```
