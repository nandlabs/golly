package cli

// Command represents a command in the CLI.
type Command struct {
	Name        string
	Usage       string
	Version     string
	Aliases     []string
	Action      func(ctx *Context) error
	SubCommands map[string]*Command
	Flags       []*Flag
}

// NewCommand creates a new command with the given name, description, and handler function.
func NewCommand(name, description, version string, action func(ctx *Context) error) *Command {
	return &Command{
		Name:        name,
		Usage:       description,
		Version:     version,
		Action:      action,
		Aliases:     make([]string, 0),
		SubCommands: make(map[string]*Command),
		Flags:       make([]*Flag, 0),
	}
}

// AddSubCommand adds a subcommand to the command.
func (cmd *Command) AddSubCommand(subCmd *Command) {
	cmd.SubCommands[subCmd.Name] = subCmd
}
