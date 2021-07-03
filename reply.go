package bcr

import (
	"errors"
	"fmt"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
)

// Errors related to sending messages
var (
	ErrBotMissingPermissions = errors.New("bot is missing permissions")
)

// Send sends a message to the context channel
func (ctx *Context) Send(content string, embeds ...discord.Embed) (m *discord.Message, err error) {
	return ctx.State.SendMessageComplex(ctx.Channel.ID, api.SendMessageData{
		Content:         content,
		Embeds:          embeds,
		AllowedMentions: ctx.Router.DefaultMentions,
	})
}

// Sendf sends a message with Printf-like syntax
func (ctx *Context) Sendf(template string, args ...interface{}) (m *discord.Message, err error) {
	return ctx.Send(fmt.Sprintf(template, args...))
}

// Reply sends a message with Printf-like syntax, in an embed.
// Use Replyc to set the embed's colour.
func (ctx *Context) Reply(template string, args ...interface{}) (m *discord.Message, err error) {
	return ctx.Send("", discord.Embed{
		Description: fmt.Sprintf(template, args...),
		Color:       ctx.Router.EmbedColor,
	})
}

// Replyc sends a message with Printf-like syntax, in an embed. The first argument is the embed's colour.
func (ctx *Context) Replyc(colour discord.Color, template string, args ...interface{}) (m *discord.Message, err error) {
	return ctx.Send("", discord.Embed{
		Description: fmt.Sprintf(template, args...),
		Color:       colour,
	})
}
