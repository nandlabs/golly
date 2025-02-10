package cli

type Context struct {
	CommandStack []string
	Flags        map[string]string
}

func NewCLIContext() *Context {
	return &Context{
		CommandStack: []string{},
		Flags:        make(map[string]string),
	}
}

func (ctx *Context) SetFlag(name, value string) {
	ctx.Flags[name] = value
}

func (ctx *Context) GetFlag(name string) (string, bool) {
	value, exists := ctx.Flags[name]
	return value, exists
}
