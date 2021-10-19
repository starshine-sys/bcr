package bcr

type commandError struct {
	path []string
	err  error
}

func (c *commandError) Error() string {
	return c.err.Error()
}

// ExecuteSlashCommand is the slash command handler called by InteractionCreate.
func (r *Router) ExecuteSlashCommand(ctx *CommandContext) (err error) {
	// top level commands first
	for _, c := range r.Commands {
		if c.Name == ctx.Data.Name {
			fn := c.Execute

			for _, mw := range r.CommandMiddlewares {
				fn = mw(fn)
			}

			for _, mw := range c.Middlewares {
				fn = mw(fn)
			}

			err = fn(ctx)
			if err != nil {
				return &commandError{
					path: []string{ctx.Data.Name},
					err:  err,
				}
			}
		}
	}

	return
}
