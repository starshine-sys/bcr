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
	return ctx.Send(fmt.Sprintf(template, args...), nil)
}

// Reply sends a message with Printf-like syntax, in an embed.
// Use Replyc to set the embed's colour.
func (ctx *Context) Reply(template string, args ...interface{}) (m *discord.Message, err error) {
	return ctx.Send("", &discord.Embed{
		Description: fmt.Sprintf(template, args...),
		Color:       ctx.Router.EmbedColor,
	})
}

// Replyc sends a message with Printf-like syntax, in an embed. The first argument is the embed's colour.
func (ctx *Context) Replyc(colour discord.Color, template string, args ...interface{}) (m *discord.Message, err error) {
	return ctx.Send("", &discord.Embed{
		Description: fmt.Sprintf(template, args...),
		Color:       colour,
	})
}

// SendEmbedData is data for SendEmbed. All these fields can be kept empty.
type SendEmbedData struct {
	Title   string
	Message string
	Footer  string

	Color discord.Color
}

// SED is an alias for SendEmbedData
type SED = SendEmbedData

// SendEmbed sends a message, formatted as an embed.
func (ctx *Context) SendEmbed(data SendEmbedData) (m *discord.Message, err error) {
	var (
		footer *discord.EmbedFooter
		color  = data.Color
	)

	if data.Footer != "" {
		footer = &discord.EmbedFooter{
			Text: data.Footer,
		}
	}

	if color == 0 {
		color = ctx.Router.EmbedColor
	}

	return ctx.Send("", &discord.Embed{
		Title:       data.Title,
		Description: data.Message,
		Footer:      footer,
		Color:       color,
	})
}
