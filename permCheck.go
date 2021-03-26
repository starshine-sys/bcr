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

// PermissionUseSlashCommands is the permission bit to use slash commands in a server
const PermissionUseSlashCommands = 1 << 31

var permNames = map[discord.Permissions]string{
	discord.PermissionCreateInstantInvite: "Create Instant Invite",
	discord.PermissionKickMembers:         "Kick Members",
	discord.PermissionBanMembers:          "Ban Members",
	discord.PermissionAdministrator:       "Administrator",
	discord.PermissionManageChannels:      "Manage Channels",
	discord.PermissionManageGuild:         "Manage Server",
	discord.PermissionAddReactions:        "Add Reactions",
	discord.PermissionViewAuditLog:        "View Audit Log",
	discord.PermissionPrioritySpeaker:     "Priority Speaker",
	discord.PermissionStream:              "Stream",
	discord.PermissionViewChannel:         "View Channel",
	discord.PermissionSendMessages:        "Send Messages",
	discord.PermissionSendTTSMessages:     "Send TTS Messages",
	discord.PermissionManageMessages:      "Manage Messages",
	discord.PermissionEmbedLinks:          "Embed Links",
	discord.PermissionAttachFiles:         "Attach Files",
	discord.PermissionReadMessageHistory:  "Read Message History",
	discord.PermissionMentionEveryone:     "Mention Everyone",
	discord.PermissionUseExternalEmojis:   "Use External Emojis",
	discord.PermissionConnect:             "Connect",
	discord.PermissionSpeak:               "Speak",
	discord.PermissionMuteMembers:         "Mute Members",
	discord.PermissionDeafenMembers:       "Deafen Members",
	discord.PermissionMoveMembers:         "Move Members",
	discord.PermissionUseVAD:              "Use VAD",
	discord.PermissionChangeNickname:      "Change Nickname",
	discord.PermissionManageNicknames:     "Manage Nicknames",
	discord.PermissionManageRoles:         "Manage Roles",
	discord.PermissionManageWebhooks:      "Manage Webhooks",
	discord.PermissionManageEmojis:        "Manage Emojis",
	// This one isn't in arikawa or the developer docs yet, but *has* rolled out to roles
	1 << 31: "Use Slash Commands",
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

// PermStrings gets the permission strings for all required permissions
func PermStrings(p discord.Permissions) []string {
	return PermStringsFor(permNames, p)
}

// PermStringsFor is like PermStrings but lets you specify the specific map used
func PermStringsFor(m map[discord.Permissions]string, p discord.Permissions) []string {
	var out = make([]string, 0, 32)
	for perm, name := range m {
		if p&perm == perm {
			out = append(out, name)
		}
	}

	return out
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
