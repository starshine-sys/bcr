package bcr

// Peek gets the next argument from the context's Args without removing it
func (ctx *Context) Peek() string {
	if len(ctx.InternalArgs) <= ctx.pos {
		return ""
	}
	return ctx.InternalArgs[ctx.pos]
}

// Pop gets the next argument from the context's Args and removes it from the slice
func (ctx *Context) Pop() string {
	if len(ctx.InternalArgs) <= ctx.pos {
		return ""
	}
	arg := ctx.InternalArgs[ctx.pos]
	ctx.pos++
	ctx.Args = ctx.InternalArgs[ctx.pos:]
	ctx.RawArgs = TrimPrefixesSpace(ctx.RawArgs, arg)
	return arg
}
