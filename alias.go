package bcr

import (
	"errors"
	"fmt"
	"strings"

	"github.com/diamondburned/arikawa/v2/bot/extras/shellwords"
)

// Errors related to creating aliases
var (
	ErrNoPath     = errors.New("alias: no path supplied")
	ErrNilCommand = errors.New("alias: command was nil")
)

// ArgTransformer is used in Alias, passing in the context's RawArgs, which are then split again.
type ArgTransformer func(string) string

// Alias creates an alias to the command `path`, and transforms the arguments according to argTransform.
// argTransform is called with the context's RawArgs.
func (r *Router) Alias(name string, aliases, path []string, argTransform ArgTransformer) (*Command, error) {
	if len(path) == 0 {
		return nil, errors.New("no path supplied")
	}

	c, ok := r.cmds[path[0]]
	if !ok {
		return nil, ErrNilCommand
	}
	if len(path) > 1 {
		for _, step := range path[1:] {
			c, ok = c.subCmds[step]
			if !ok {
				return nil, ErrNilCommand
			}
		}
	}

	cmd := Command{
		Name:    name,
		Aliases: aliases,

		Summary:     fmt.Sprintf("Alias to `%v`:\n%v", strings.Join(path, " "), c.Summary),
		Description: c.Description,
		Usage:       c.Usage,

		Flags: c.Flags,

		Blacklistable:     c.Blacklistable,
		CustomPermissions: c.CustomPermissions,
		Permissions:       c.Permissions,

		subCmds: c.subCmds,

		GuildOnly: c.GuildOnly,
		OwnerOnly: c.OwnerOnly,
		Cooldown:  c.Cooldown,

		Command: func(ctx *Context) (err error) {
			if argTransform != nil {
				ctx.RawArgs = argTransform(ctx.RawArgs)

				ctx.Args, err = shellwords.Parse(ctx.RawArgs)
				if err != nil {
					ctx.Args = strings.Split(ctx.RawArgs, " ")
				}
			}

			return c.Command(ctx)
		},
	}

	return &cmd, nil
}

// AliasMust is a wrapper around Alias that panics if err is non-nil
func (r *Router) AliasMust(name string, aliases, path []string, argTransform ArgTransformer) *Command {
	a, err := r.Alias(name, aliases, path, argTransform)
	if err != nil {
		panic(err)
	}
	return a
}

// DefaultArgTransformer adds a prefix or suffix (or both!) to the current args
func DefaultArgTransformer(prefix, suffix string) ArgTransformer {
	t := strings.TrimSpace(fmt.Sprintf("%v $args %v", prefix, suffix))

	return func(args string) string {
		return strings.Replace(t, "$args", args, -1)
	}
}
