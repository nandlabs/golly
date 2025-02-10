package cli

import (
	"fmt"
	"testing"
)

func TestNewCommand(t *testing.T) {
	handler := func(ctx *Context) error { return nil }
	cmd := NewCommand("test", "A test command", handler)

	if cmd.Name != "test" {
		t.Errorf("Expected Name to be 'test', got '%s'", cmd.Name)
	}
	if cmd.Description != "A test command" {
		t.Errorf("Expected Description to be 'A test command', got '%s'", cmd.Description)
	}
	if cmd.Handler == nil {
		t.Error("Expected Handler to be non-nil")
	}
	if len(cmd.SubCommands) != 0 {
		t.Errorf("Expected SubCommands to be empty, got %d", len(cmd.SubCommands))
	}
	if len(cmd.Flags) != 0 {
		t.Errorf("Expected Flags to be empty, got %d", len(cmd.Flags))
	}
}

func TestAddSubCommand(t *testing.T) {
	parentCmd := NewCommand("parent", "Parent command", nil)
	subCmd := NewCommand("child", "Child command", nil)
	parentCmd.AddSubCommand(subCmd)

	if len(parentCmd.SubCommands) != 1 {
		t.Errorf("Expected 1 SubCommand, got %d", len(parentCmd.SubCommands))
	}
	if parentCmd.SubCommands["child"] != subCmd {
		t.Error("SubCommand 'child' not added correctly")
	}
}

func TestPrintCommandHelp(t *testing.T) {
	cli := NewCLI()
	cmd := NewCommand("test", "Test command", nil)
	cmd.Flags = []Flag{
		{
			Name:    "attach",
			Aliases: []string{"a"},
			Usage:   "Attach to STDIN, STDOUT or STDERR",
			Default: "[]",
		},
	}

	cmd.AddSubCommand(NewCommand("subcmd", "A subcommand", nil))

	cli.printCommandHelp(cmd, 0)
}

func TestPrintDetailedCommandHelp(t *testing.T) {
	cli := NewCLI()
	server := &Command{
		Name:        "server",
		Description: "Server command",
		Handler: func(ctx *Context) error {
			region, _ := ctx.GetFlag("region")
			fmt.Printf("IN REGION, %s\n", region)
			return nil
		},
		Flags: []Flag{
			{
				Name:    "region",
				Aliases: []string{"r"},
				Usage:   "Provide region",
				Default: "us-east-1",
			},
		},
		SubCommands: map[string]*Command{
			"create": {
				Name:        "create",
				Description: "create",
				Handler: func(ctx *Context) error {
					typ, _ := ctx.GetFlag("type")
					fmt.Printf("SERVER TYPE %s\n", typ)
					return nil
				},
				Flags: []Flag{
					{
						Name:    "type",
						Aliases: []string{"t"},
						Usage:   "server type",
						Default: "t2.micro",
					},
				},
			},
		},
	}

	cli.AddCommand(server)

	cli.printUsage()

	cli.printDetailedHelp([]string{"server", "create"}, server)
}
