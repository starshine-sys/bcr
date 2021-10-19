package bcr

// CommandBuilder is a builder for commands.
type CommandBuilder struct {
	name        string
	description string
	middlewares []CommandMiddleware
	exec        CommandFunc

	groups      []builderGroup
	subcommands []*Command
}

type builderGroup struct {
	middlewares []CommandMiddleware
	commands    []*Command
}

// On initializes a new command builder.
func On(name string) *CommandBuilder {
	return &CommandBuilder{name: name}
}

// Description sets the description for this command.
func (c *CommandBuilder) Description(desc string) *CommandBuilder {
	c.description = desc
	return c
}

// With adds middleware(s) to this command.
func (c *CommandBuilder) With(mw ...CommandMiddleware) *CommandBuilder {
	c.middlewares = append(c.middlewares, mw...)
	return c
}

// Do adds an execute function to this command.
func (c *CommandBuilder) Do(fn CommandFunc) *CommandBuilder {
	c.exec = fn
	return c
}

// Mount adds the command to the given router.
func (c *CommandBuilder) Mount(r *Router) {
	if len(c.groups) == 0 && len(c.subcommands) == 0 {
		c.mountRoot(r)
		return
	}

	panic("unimplemented")
}

func (c *CommandBuilder) mountRoot(r *Router) {
	if c.name == "" {
		panic("name cannot be empty")
	}

	if c.description == "" {
		panic("description cannot be empty")
	}

	if c.exec == nil {
		panic("function cannot be nil")
	}

	cmd := &Command{
		Name:        c.name,
		Description: c.description,
		Middlewares: c.middlewares,
		Execute:     c.exec,
	}

	r.Commands = append(r.Commands, cmd)
}
