package bcr

import (
	"fmt"
	"strings"

	"github.com/diamondburned/arikawa/v3/discord"
)

// Check is a check for slash commands.
// If err != nil:
// - if err implements CheckError, respond with the content and embeds from that method
// - otherwise, print the string representation of the error
type Check[T HasContext] func(ctx T) (err error)

type CheckError[T HasContext] interface {
	CheckError(T) (string, []discord.Embed)
}

// checkError is a basic implementation of CheckError that simply returns the given content and strings
type checkError[T HasContext] struct {
	content string
	embeds  []discord.Embed
}

func (c *checkError[T]) Error() string                          { return c.content }
func (c *checkError[T]) CheckError(T) (string, []discord.Embed) { return c.content, c.embeds }

func NewCheckError[T HasContext](content string, embeds ...discord.Embed) error {
	return &checkError[T]{
		content: content,
		embeds:  embeds,
	}
}

// And combines all the given checks into a single check.
// The first one to fail is returned.
func And[T HasContext](checks ...Check[T]) Check[T] {
	return func(ctx T) error {
		for _, check := range checks {
			if err := check(ctx); err != nil {
				return err
			}
		}
		return nil
	}
}

// Or checks all given checks and returns nil if at least one of them succeeds.
func Or[T HasContext](checks ...Check[T]) Check[T] {
	return func(ctx T) (err error) {
		for _, check := range checks {
			err = check(ctx)
			if err == nil {
				return nil
			}
		}
		return err
	}
}

func HasChannelPermissions(perm discord.Permissions) Check[*CommandContext] {
	return func(ctx *CommandContext) error {
		if ctx.Member == nil || ctx.Guild == nil || ctx.Channel == nil {
			return NewCheckError[*CommandContext]("This command cannot be run in DMs.")
		}

		if !overwrites(ctx.Guild, ctx.Channel, ctx.Member).Has(perm) {
			permStrings := PermStrings(perm)
			tmpl := "You must have the `%v` permission to use this command."
			if len(permStrings) > 1 {
				tmpl = "You must have following permissions to run this command: `%v`"
			}

			return NewCheckError[*CommandContext](
				fmt.Sprintf(tmpl, strings.Join(permStrings, ", ")))
		}

		return nil
	}
}

const dmPermissions = discord.PermissionViewChannel | discord.PermissionAddReactions | discord.PermissionAttachFiles | discord.PermissionEmbedLinks | discord.PermissionSendMessages | discord.PermissionReadMessageHistory

func overwrites(g *discord.Guild, ch *discord.Channel, m *discord.Member) discord.Permissions {
	if g == nil || ch == nil || m == nil {
		return dmPermissions
	}
	return discord.CalcOverwrites(*g, *ch, *m)
}
