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

// checkError is a basic implementation of CheckError that simply returns the given content and string.
type checkError[T HasContext] struct {
	content string
	embeds  []discord.Embed
}

func (c *checkError[T]) Error() string                          { return c.content }
func (c *checkError[T]) CheckError(T) (string, []discord.Embed) { return c.content, c.embeds }

// NewCheckError returns a simple, static CheckError.
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

// HasChannelPermissions returns a check that requires the given permissions on a channel level (taking into account overwrites).
func HasChannelPermissions[T HasContext](perm discord.Permissions) Check[T] {
	return func(ctx T) error {
		cctx := ctx.Ctx()
		if cctx.Member == nil || cctx.Guild == nil || cctx.Channel == nil {
			return NewCheckError[T]("This command cannot be run in DMs.")
		}

		if !overwrites(cctx.Guild, cctx.Channel, cctx.Member).Has(perm) {
			permStrings := PermStrings(perm)
			tmpl := "You must have the `%v` permission to use this command."
			if len(permStrings) > 1 {
				tmpl = "You must have following permissions to run this command: `%v`"
			}

			return NewCheckError[T](
				fmt.Sprintf(tmpl, strings.Join(permStrings, ", ")))
		}

		return nil
	}
}

// HasGuildPermissions returns a check that requires the given permissions on a guild level.
func HasGuildPermissions[T HasContext](perm discord.Permissions) Check[T] {
	return func(ctx T) error {
		cctx := ctx.Ctx()
		if cctx.Member == nil || cctx.Guild == nil {
			return NewCheckError[T]("This command cannot be run in DMs.")
		}

		if !guildPerms(*cctx.Guild, *cctx.Member).Has(perm) {
			permStrings := PermStrings(perm)
			tmpl := "You must have the `%v` permission to use this command."
			if len(permStrings) > 1 {
				tmpl = "You must have following permissions to run this command: `%v`"
			}

			return NewCheckError[T](
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
	return discord.CalcOverrides(*g, *ch, *m, g.Roles)
}

func guildPerms(g discord.Guild, m discord.Member) discord.Permissions {
	if g.OwnerID == m.User.ID {
		return discord.PermissionAll
	}

	var perms discord.Permissions

	for _, r := range g.Roles {
		for _, id := range m.RoleIDs {
			if r.ID == id {
				if r.Permissions.Has(discord.PermissionAdministrator) {
					return discord.PermissionAll
				}

				perms |= r.Permissions
				break
			}
		}
	}
	return perms
}
