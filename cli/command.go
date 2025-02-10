package cli

type Command struct {
	Name        string
	Description string
	Handler     func(ctx *Context) error
	SubCommands map[string]*Command
	Flags       []Flag
}

func NewCommand(name, description string, handler func(ctx *Context) error) *Command {
	return &Command{
		Name:        name,
		Description: description,
		Handler:     handler,
		SubCommands: make(map[string]*Command),
		Flags:       []Flag{},
	}
}

func (cmd *Command) AddSubCommand(subCmd *Command) {
	cmd.SubCommands[subCmd.Name] = subCmd
}
