package cli

import (
	"context"
	"flag"
)

type Context struct {
	context.Context
	App     *App
	Command *Command
	//flagsSet      *flag.FlagSet
	parentContext *Context
}

func NewContext(app *App, parentCtx *Context) *Context {
	c := &Context{
		App:           app,
		parentContext: parentCtx,
	}
	if parentCtx != nil {
		c.Context = parentCtx.Context
	}
	c.Command = &Command{}
	if c.Context == nil {
		c.Context = context.Background()
	}
	return c
}

func (conTxt *Context) Args() Args {
	res := args(flag.Args())
	return &res
}

func (conTxt *Context) GetFlag(name string) interface{} {
	return mappedFlags[name]
}
