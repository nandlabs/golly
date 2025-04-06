package cli

// Context represents the context for the command-line interface.
type Context struct {
	CommandStack []string          // CommandStack stores the stack of executed commands.
	Flags        map[string]string // Flags stores the command-line flags and their values.
}

// NewCLIContext creates a new CLI context.
func NewCLIContext() *Context {
	return &Context{
		CommandStack: []string{},
		Flags:        make(map[string]string),
	}
}

// SetFlag sets the value of a command-line flag.
func (ctx *Context) SetFlag(name, value string) {
	ctx.Flags[name] = value
}

// GetFlag retrieves the value of a command-line flag.
// It returns the value and a boolean indicating whether the flag exists.
func (ctx *Context) GetFlag(name string) (string, bool) {
	value, exists := ctx.Flags[name]
	return value, exists
}
