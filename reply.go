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
	if !ctx.checkBotSendPerms(ctx.Channel.ID, true) {
		return nil, ErrBotMissingPermissions
	}

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

	return ctx.State.SendMessageComplex(ctx.Channel.ID, api.SendMessageData{
		Embed: &discord.Embed{
			Title:       data.Title,
			Description: data.Message,
			Footer:      footer,
			Color:       color,
		},
	})
}
