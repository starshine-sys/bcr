package bcr

import "github.com/diamondburned/arikawa/v3/discord"

// GuildPerms returns the global (guild) permissions of this Context's user.
// If in DMs, it will return the permissions users have in DMs.
func (ctx *Context) GuildPerms() (perms discord.Permissions) {
	if ctx.Guild == nil || ctx.Member == nil {
		return discord.PermissionViewChannel | discord.PermissionSendMessages | discord.PermissionAddReactions | discord.PermissionReadMessageHistory
	}

	if ctx.Guild.OwnerID == ctx.Author.ID {
		return discord.PermissionAll
	}

	for _, id := range ctx.Member.RoleIDs {
		for _, r := range ctx.Guild.Roles {
			if id == r.ID {
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

// checkOwner returns true if the user can run the command
func (ctx *Context) checkOwner() bool {
	if !ctx.Cmd.OwnerOnly {
		return true
	}

	for _, i := range ctx.Router.BotOwners {
		if i == ctx.Author.ID.String() {
			return true
		}
	}

	return false
}

// CheckBotSendPerms checks if the bot can send messages in a channel
func (ctx *Context) CheckBotSendPerms(ch discord.ChannelID, e bool) bool {
	// if this is a DM channel we always have perms
	if ctx.Channel.ID == ch && ctx.Message.GuildID == 0 {
		return true
	}

	perms, err := ctx.State.Permissions(ch, ctx.Bot.ID)
	if err != nil {
		return false
	}

	// if the bot doesn't have permission to send messages to the channel, return
	if !perms.Has(discord.PermissionViewChannel) || !perms.Has(discord.PermissionSendMessages) {
		return false
	}

	// if the bot requires embed links but doesn't have it, return false
	if e && !perms.Has(discord.PermissionEmbedLinks) {
		// but we *can* send an error message (at least probably, we've checked for perms already)
		ctx.State.SendMessage(ch, ":x: I do not have permission to send embeds in this channel. Please ensure I have the `Embed Links` permission here.")
		return false
	}

	return true
}
