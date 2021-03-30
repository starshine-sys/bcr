package bcr

import (
	"strings"

	"github.com/diamondburned/arikawa/v2/discord"
)

// PermError is a permission error
type PermError struct {
	perms discord.Permissions
	s     []string
}

func (p *PermError) Error() string {
	return strings.Join(p.s, ", ")
}

// CheckPerms checks the user's permissions in the current channel
func (ctx *Context) CheckPerms() (err error) {
	return ctx.perms(ctx.Author.ID, ctx.Cmd.Permissions)
}

// CheckBotPerms checks the bot's permissions in the current channel
func (ctx *Context) CheckBotPerms(p discord.Permissions) (err error) {
	return ctx.perms(ctx.Bot.ID, p)
}

func (ctx *Context) perms(user discord.UserID, p discord.Permissions) (err error) {
	perms, err := ctx.State.Permissions(ctx.Channel.ID, user)
	if err != nil {
		return err
	}

	b := perms.Has(p)
	if b {
		return nil
	}
	return &PermError{s: PermStrings(p), perms: p}
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

// checkBotSendPerms checks if the bot can send messages in a channel
func (ctx *Context) checkBotSendPerms(ch discord.ChannelID, e bool) bool {
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
		ctx.State.SendMessage(ch, ":x: I do not have permission to send embeds in this channel. Please ensure I have the `Embed Links` permission here.", nil)
		return false
	}

	return true
}
