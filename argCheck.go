package bcr

import "errors"

// Errors
var (
	ErrorNotEnoughArgs = errors.New("not enough arguments")
	ErrorTooManyArgs   = errors.New("too many arguments")
)

// CheckMinArgs checks if the argument count is less than the given count
func (ctx *Context) CheckMinArgs(c int) (err error) {
	if len(ctx.Args) < c {
		return ErrorNotEnoughArgs
	}
	return nil
}

// CheckRequiredArgs checks if the arg count is exactly the given count
func (ctx *Context) CheckRequiredArgs(c int) (err error) {
	if len(ctx.Args) != c {
		if len(ctx.Args) > c {
			return ErrorTooManyArgs
		}
		return ErrorNotEnoughArgs
	}
	return nil
}

// CheckArgRange checks if the number of arguments is within the given range
func (ctx *Context) CheckArgRange(min, max int) (err error) {
	if len(ctx.Args) > max {
		return ErrorTooManyArgs
	}
	if len(ctx.Args) < min {
		return ErrorNotEnoughArgs
	}
	return nil
}
