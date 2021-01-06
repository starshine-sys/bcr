package bcr

// DefaultHelpCommand is the default command called when using [prefix]help.
// This can be changed by setting r.HelpCommand.
func (r *Router) DefaultHelpCommand(ctx *Context) (err error) {
	return
}
