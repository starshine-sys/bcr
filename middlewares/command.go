package middlewares

import (
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/starshine-sys/bcr/v2"
)

// RequireNSFW is a command middleware that requires the command to be run in an NSFW channel.
func RequireNSFW(next bcr.CommandFunc) bcr.CommandFunc {
	return func(ctx *bcr.CommandContext) (err error) {
		if !ctx.Channel.NSFW {
			return ctx.RespondEphemeral("This command can only be used in NSFW channels.")
		}

		return next(ctx)
	}
}

// RequireGuildPermission is a command middleware that requires the given permissions for a command.
func RequireGuildPermission(perm discord.Permissions) bcr.CommandMiddleware {
	return func(next bcr.CommandFunc) bcr.CommandFunc {
		return func(ctx *bcr.CommandContext) (err error) {
			if !guildPerms(ctx).Has(perm) {
				return ctx.RespondEphemeral("You are not allowed to use this command, you are missing required permissions.")
			}

			return next(ctx)
		}
	}
}

func guildPerms(ctx *bcr.CommandContext) (perms discord.Permissions) {
	if ctx.Guild == nil || ctx.Member == nil {
		return 0
	}

	if ctx.Guild.OwnerID == ctx.User.ID {
		return discord.PermissionAll
	}

	for _, r := range ctx.Guild.Roles {
		for _, id := range ctx.Member.RoleIDs {
			if r.ID == id {
				perms |= r.Permissions
			}
		}
	}

	if perms.Has(discord.PermissionAdministrator) {
		return discord.PermissionAll
	}

	return perms
}
