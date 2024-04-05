package cli

import (
	"context"
	"io"
	"os"
	"path/filepath"
)

// App represents a CLI application.
type App struct {
	// Name is the application name.
	Name string
	// Usage is the application usage information.
	Usage string
	// HelpName is the name used in the help command.
	HelpName string
	// ArgsUsage is the usage information for command arguments.
	ArgsUsage string
	// UsageText is the custom usage text for the application.
	UsageText string
	// Version is the application version.
	Version string
	// HideVersion determines whether to hide the version information.
	HideVersion bool
	// Action is the function to be invoked on default execution.
	Action ActionFunc
	// Flags are the global flags for the application.
	Flags []*Flag
	// Commands are the application commands.
	Commands []*Command
	// Writer is the output writer for the application.
	Writer io.Writer
	// HideHelp determines whether to hide the help command.
	HideHelp bool
	// HideHelpCommand determines whether to hide the help command in the list of commands.
	HideHelpCommand bool
	// CommandVisible determines whether the commands are visible.
	CommandVisible bool
	// setupComplete determines whether the application setup is complete.
	setupComplete bool
	// rootCommand is the root command of the application.
	rootCommand *Command
}

// initialize initializes the application.
func (app *App) initialize() {
	if app.setupComplete {
		return
	}

	app.setupComplete = true

	if app.Name == "" {
		app.Name = filepath.Base(os.Args[0])
	}

	if app.HelpName == "" {
		app.HelpName = app.Name
	}

	if app.Usage == "" {
		app.Usage = "CLI App 101"
	}

	if app.Version == "" {
		app.HideVersion = true
	}

	var newCommands []*Command
	for _, c := range app.Commands {
		if c.HelpName == "" {
			c.HelpName = c.Name
		}
		newCommands = append(newCommands, c)
	}
	app.Commands = newCommands

	if app.Command(helpCommand.Name) == nil && !app.HideHelp {
		if HelpFlag != nil {
			app.appendFlag(HelpFlag)
		}
	}

	if len(app.Commands) > 0 {
		app.CommandVisible = true
	}

	if app.Action == nil {
		app.Action = helpCommand.Action
	}

	if app.Writer == nil {
		app.Writer = os.Stdout
	}
}

// Execute executes the application with the given arguments.
func (app *App) Execute(arguments []string) error {
	return app.ExecuteContext(context.Background(), arguments)
}

// ExecuteContext executes the application with the given context and arguments.
func (app *App) ExecuteContext(ctx context.Context, arguments []string) error {
	app.initialize()

	conTxt := NewContext(app, &Context{Context: ctx})

	app.rootCommand = app.newRootCommand()
	conTxt.Command = app.rootCommand

	return app.rootCommand.Run(conTxt, arguments...)
}

// newRootCommand creates a new root command for the application.
func (app *App) newRootCommand() *Command {
	return &Command{
		Name:      app.Name,
		Usage:     app.Usage,
		Action:    app.Action,
		Flags:     app.Flags,
		Commands:  app.Commands,
		ArgsUsage: app.ArgsUsage,
	}
}

// writer returns the output writer for the application.
func (app *App) writer() io.Writer {
	return app.Writer
}

// Command returns the command with the given name.
func (app *App) Command(name string) *Command {
	for _, c := range app.Commands {
		if c.HasName(name) {
			return c
		}
	}
	return nil
}

// appendCommand appends a command to the application.
func (app *App) appendCommand(c *Command) {
	if !hasCommand(app.Commands, c) {
		app.Commands = append(app.Commands, c)
	}
}

// appendFlag appends a flag to the application.
func (app *App) appendFlag(flag *Flag) {
	if !hasFlag(app.Flags, flag) {
		app.Flags = append(app.Flags, flag)
	}
}

// VisibleCommands returns the visible commands of the application.
func (app *App) VisibleCommands() []*Command {
	return app.Commands
}
