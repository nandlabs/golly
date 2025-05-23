package cli

import (
	"os"
	"strings"
	"testing"
)

func TestNewCLI(t *testing.T) {
	cli := NewCLI()
	if cli == nil {
		t.Fatal("Expected NewCLI to return a non-nil CLI instance")
	}
	if len(cli.rootCommands) != 0 {
		t.Errorf("Expected rootCommands to be empty, got %d", len(cli.rootCommands))
	}
}

func TestAddCommand(t *testing.T) {
	cli := NewCLI()
	cmd := NewCommand("test", "Test command", "v0.0.1", nil)
	cli.AddCommand(cmd)

	if len(cli.rootCommands) != 1 {
		t.Errorf("Expected 1 root command, got %d", len(cli.rootCommands))
	}
	if cli.rootCommands["test"] != cmd {
		t.Error("Command not added correctly to rootCommands")
	}
}

func TestExecute_NoCommandProvided(t *testing.T) {
	cli := NewCLI()
	os.Args = []string{"cli"} // Simulate running the CLI without a command

	err := cli.Execute()
	if err == nil || !strings.Contains(err.Error(), "no command provided") {
		t.Errorf("Expected error for no command provided, got %v", err)
	}
}

func TestExecute_UnknownCommand(t *testing.T) {
	cli := NewCLI()
	os.Args = []string{"cli", "unknown"} // Simulate running the CLI with an unknown command

	err := cli.Execute()
	if err == nil || !strings.Contains(err.Error(), "unknown command") {
		t.Errorf("Expected error for unknown command, got %v", err)
	}
}

func TestExecute_GlobalHelp(t *testing.T) {
	cli := NewCLI()
	os.Args = []string{"cli", "--help"} // Simulate running the CLI with global help flag

	err := cli.Execute()
	if err != nil {
		t.Errorf("Expected no error for --help flag, got %v", err)
	}
}

func TestExecute_CommandHelp(t *testing.T) {
	cli := NewCLI()
	cmd := NewCommand("test", "Test command", "v0.0.1", nil)
	cli.AddCommand(cmd)
	os.Args = []string{"cli", "test", "--help"} // Simulate running help for a specific command

	err := cli.Execute()
	if err != nil {
		t.Errorf("Expected no error for command help, got %v", err)
	}
}

func TestExecute_CommandExecution(t *testing.T) {
	cli := NewCLI()
	called := false
	handler := func(ctx *Context) error {
		called = true
		return nil
	}

	cmd := NewCommand("test", "Test command", "v0.0.1", handler)
	cli.AddCommand(cmd)
	os.Args = []string{"cli", "test"} // Simulate running the "test" command

	err := cli.Execute()
	if err != nil {
		t.Errorf("Expected no error for command execution, got %v", err)
	}
	if !called {
		t.Error("Expected command handler to be called")
	}
}

func TestExecute_FlagParsing(t *testing.T) {
	cli := NewCLI()
	called := false
	handler := func(ctx *Context) error {
		if ctx.Flags["name"] != "value" {
			t.Errorf("Expected flag 'name' to be 'value', got '%s'", ctx.Flags["name"])
		}
		called = true
		return nil
	}

	cmd := NewCommand("test", "Test command", "v0.0.1", handler)
	cmd.Flags = []*Flag{
		{Name: "name", Aliases: []string{"n"}, Usage: "Specify a name", Default: ""},
	}
	cli.AddCommand(cmd)
	os.Args = []string{"cli", "test", "--name=value"} // Simulate running with flags

	err := cli.Execute()
	if err != nil {
		t.Errorf("Expected no error for command execution with flags, got %v", err)
	}
	if !called {
		t.Error("Expected command handler to be called")
	}
}

func TestExecute_FlagParsingAlias(t *testing.T) {
	cli := NewCLI()
	called := false
	handler := func(ctx *Context) error {
		if ctx.Flags["name"] != "value" {
			t.Errorf("Expected flag 'name' to be 'value', got '%s'", ctx.Flags["name"])
		}
		called = true
		return nil
	}

	cmd := NewCommand("test", "Test command", "v0.0.1", handler)
	cmd.Flags = []*Flag{
		{Name: "name", Aliases: []string{"n"}, Usage: "Specify a name", Default: ""},
	}
	cli.AddCommand(cmd)
	os.Args = []string{"cli", "test", "-n=value"} // Simulate running with flags

	err := cli.Execute()
	if err != nil {
		t.Errorf("Expected no error for command execution with flags, got %v", err)
	}
	if !called {
		t.Error("Expected command handler to be called")
	}
}

func TestExecute_DefaultFlagParsingAlias(t *testing.T) {
	cli := NewCLI()
	called := false
	handler := func(ctx *Context) error {
		if ctx.Flags["name"] != "user" {
			t.Errorf("Expected Default value for the flag 'name' to be 'user', got '%s", ctx.Flags["name"])
		}
		called = true
		return nil
	}

	cmd := NewCommand("test", "Test Command", "v0.0.1", handler)
	cmd.Flags = []*Flag{
		{Name: "name", Aliases: []string{"n"}, Usage: "specify name", Default: "user"},
	}
	cli.AddCommand(cmd)
	os.Args = []string{"cli", "test", "-n"}

	err := cli.Execute()
	if err != nil {
		t.Errorf("Expected no error from command execution with default flags, got %v", err)
	}
	if !called {
		t.Error("Expected command handler to be called")
	}
}

func TestExecute_DefaultFlagParsing(t *testing.T) {
	cli := NewCLI()
	called := false
	handler := func(ctx *Context) error {
		if ctx.Flags["name"] != "user" {
			t.Errorf("Expected Default value for the flag 'name' to be 'user', got '%s", ctx.Flags["name"])
		}
		called = true
		return nil
	}

	cmd := NewCommand("test", "Test Command", "v0.0.1", handler)
	cmd.Flags = []*Flag{
		{Name: "name", Aliases: []string{"n"}, Usage: "specify name", Default: "user"},
	}
	cli.AddCommand(cmd)
	os.Args = []string{"cli", "test", "--name"}

	err := cli.Execute()
	if err != nil {
		t.Errorf("Expected no error from command execution with default flags, got %v", err)
	}
	if !called {
		t.Error("Expected command handler to be called")
	}
}

func TestExecute_CLIVersionFlag(t *testing.T) {
	cli := NewCLI()
	cli.AddVersion("v1.0.0")
	os.Args = []string{"cli", "--version"} // Simulate running with version flag

	err := cli.Execute()
	if err != nil {
		t.Errorf("Expected no error for version flag, got %v", err)
	}
}

func TestExecute_CommandVersionFlag(t *testing.T) {
	cli := NewCLI()
	cmd := NewCommand("test", "Test command", "v0.0.1", nil)
	cli.AddCommand(cmd)
	os.Args = []string{"cli", "test", "--version"} // Simulate running with version flag

	err := cli.Execute()
	if err != nil {
		t.Errorf("Expected no error for version flag, got %v", err)
	}
}

func TestExecute_SubCommandVersionFlag(t *testing.T) {
	cli := NewCLI()
	cmd := NewCommand("test", "Test command", "v0.0.1", nil)
	subCmd := NewCommand("sub", "Sub command", "v1.0.0", nil)
	cmd.AddSubCommand(subCmd)
	cli.AddCommand(cmd)
	os.Args = []string{"cli", "test", "sub", "--version"} // Simulate running with version flag

	err := cli.Execute()
	if err != nil {
		t.Errorf("Expected no error for version flag, got %v", err)
	}
}

func TestExecute_CommandAlias(t *testing.T) {
	cli := NewCLI()
	called := false
	handler := func(ctx *Context) error {
		called = true
		return nil
	}

	cmd := NewCommand("test", "Test command", "v0.0.1", handler)
	cmd.Aliases = []string{"t"}
	cli.AddCommand(cmd)
	os.Args = []string{"cli", "t"} // Simulate running the command with alias

	err := cli.Execute()
	if err != nil {
		t.Errorf("Expected no error for command execution with alias, got %v", err)
	}
	if !called {
		t.Error("Expected command handler to be called")
	}
}

func TestExecute_SubCommandAlias(t *testing.T) {
	cli := NewCLI()
	called := false
	handler := func(ctx *Context) error {
		called = true
		return nil
	}

	cmd := NewCommand("test", "Test command", "v0.0.1", nil)
	subCmd := NewCommand("sub", "Sub command", "v1.0.0", handler)
	subCmd.Aliases = []string{"s"}
	cmd.AddSubCommand(subCmd)
	cli.AddCommand(cmd)
	os.Args = []string{"cli", "test", "s"} // Simulate running the subcommand with alias

	err := cli.Execute()
	if err != nil {
		t.Errorf("Expected no error for subcommand execution with alias, got %v", err)
	}
	if !called {
		t.Error("Expected subcommand handler to be called")
	}
}

func TestExecute_InvalidSubCommand(t *testing.T) {
	cli := NewCLI()
	cmd := NewCommand("test", "Test command", "v0.0.1", nil)
	cli.AddCommand(cmd)
	os.Args = []string{"cli", "test", "invalid"} // Simulate running an invalid subcommand

	err := cli.Execute()
	if err == nil || !strings.Contains(err.Error(), "unknown command") {
		t.Errorf("Expected error for unknown subcommand, got %v", err)
	}
}
