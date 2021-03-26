package bcr

import (
	"errors"
	"fmt"

	"github.com/diamondburned/arikawa/v2/api"
	"github.com/diamondburned/arikawa/v2/discord"
)

// Errors related to sending messages
var (
	ErrBotMissingPermissions = errors.New("bot is missing permissions")
)

// Send sends a message to the context channel
func (ctx *Context) Send(content string, embed *discord.Embed) (m *discord.Message, err error) {
	if !ctx.checkBotSendPerms(ctx.Channel.ID, embed != nil) {
		return nil, ErrBotMissingPermissions
	}

	return ctx.State.SendMessageComplex(ctx.Channel.ID, api.SendMessageData{
		Content:         content,
		Embed:           embed,
		AllowedMentions: ctx.Router.DefaultMentions,
	})
}

// Sendf sends a message with Printf-like syntax
func (ctx *Context) Sendf(template string, args ...interface{}) (m *discord.Message, err error) {
	if !ctx.checkBotSendPerms(ctx.Channel.ID, false) {
		return nil, ErrBotMissingPermissions
	}

	return ctx.Send(fmt.Sprintf(template, args...), nil)
}

// Reply *replies* to the original message in the context channel
func (ctx *Context) Reply(content string, embed *discord.Embed) (m *discord.Message, err error) {
	if !ctx.checkBotSendPerms(ctx.Channel.ID, embed != nil) {
		return nil, ErrBotMissingPermissions
	}

	return ctx.State.SendMessageComplex(ctx.Channel.ID, api.SendMessageData{
		Content:         content,
		Embed:           embed,
		AllowedMentions: ctx.Router.DefaultMentions,

		Reference: &discord.MessageReference{
			MessageID: ctx.Message.ID,
		},
	})
}
