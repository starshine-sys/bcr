package bcr

import (
	"errors"
	"strings"
)

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

func (ctx *Context) argCheck() (err error) {
	// if there's no requirements, return
	if ctx.Cmd.Args == nil {
		return nil
	}

	// there's only a minimum number of arguments
	// if there's too few, show an error
	if ctx.Cmd.Args[1] == -1 && len(ctx.Args) < ctx.Cmd.Args[0] {
		_, err = ctx.Sendf(
			":x: You didn't give enough arguments: this command requires %v arguments, but you gave %v.\n> **Usage:**\n> ```%v%v %v```",
			ctx.Cmd.Args[0],
			len(ctx.Args),
			ctx.Router.Prefixes[0], strings.Join(ctx.fullCommandPath, " "), ctx.Cmd.Usage,
		)
		if err != nil {
			return err
		}
		return errCommandRun
	}

	// there's only a maximum number of arguments
	// if there's too many, show an error
	if ctx.Cmd.Args[0] == -1 && len(ctx.Args) > ctx.Cmd.Args[1] {
		_, err = ctx.Sendf(
			":x: You gave too many arguments: this command requires at most %v arguments, but you gave %v.\n> **Usage:**\n> ```%v%v %v```",
			ctx.Cmd.Args[1],
			len(ctx.Args),
			ctx.Router.Prefixes[0], strings.Join(ctx.fullCommandPath, " "), ctx.Cmd.Usage,
		)
		if err != nil {
			return err
		}
		return errCommandRun
	}

	// there's both a minimum and maximum number of arguments
	if ctx.Cmd.Args[0] != -1 && ctx.Cmd.Args[1] != -1 {
		if ctx.Cmd.Args[0] == ctx.Cmd.Args[1] && len(ctx.Args) != ctx.Cmd.Args[0] {
			_, err = ctx.Sendf(
				":x: This command requires exactly %v arguments, but you gave %v.\n> **Usage:**\n> ```%v%v %v```",
				ctx.Cmd.Args[0],
				len(ctx.Args),
				ctx.Router.Prefixes[0], strings.Join(ctx.fullCommandPath, " "), ctx.Cmd.Usage,
			)
		} else if len(ctx.Args) < ctx.Cmd.Args[0] {
			_, err = ctx.Sendf(
				":x: You didn't give enough arguments: this command requires %v arguments, but you gave %v.\n> **Usage:**\n> ```%v%v %v```",
				ctx.Cmd.Args[0],
				len(ctx.Args),
				ctx.Router.Prefixes[0], strings.Join(ctx.fullCommandPath, " "), ctx.Cmd.Usage,
			)
		} else if len(ctx.Args) > ctx.Cmd.Args[1] {
			_, err = ctx.Sendf(
				":x: You gave too many arguments: this command requires at most %v arguments, but you gave %v.\n> **Usage:**\n> ```%v%v %v```",
				ctx.Cmd.Args[1],
				len(ctx.Args),
				ctx.Router.Prefixes[0], strings.Join(ctx.fullCommandPath, " "), ctx.Cmd.Usage,
			)
		}
		if err != nil {
			return err
		}
		return errCommandRun
	}

	// everything's fine, return nil
	return nil
}
