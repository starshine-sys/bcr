package bcr

import (
	"errors"
	"fmt"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/utils/sendpart"
)

// Errors related to sending messages
var (
	ErrBotMissingPermissions = errors.New("bot is missing permissions")
)

// SendX sends a message without returning the created discord.Message
func (ctx *Context) SendX(content string, embeds ...discord.Embed) (err error) {
	msg, err := ctx.Send(content, embeds...)
	if err != nil {
		return
	}

	ctx.origMessage = msg
	return
}

// Send sends a message to the context channel
func (ctx *Context) Send(content string, embeds ...discord.Embed) (m *discord.Message, err error) {
	m, err = ctx.State.SendMessageComplex(ctx.Channel.ID, api.SendMessageData{
		Content:         content,
		Embeds:          embeds,
		AllowedMentions: ctx.Router.DefaultMentions,
	})
	if err != nil {
		return
	}

	ctx.origMessage = m
	return
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

// SendFiles sends a message with attachments
func (ctx *Context) SendFiles(content string, files ...sendpart.File) (err error) {
	msg, err := ctx.State.SendMessageComplex(ctx.Channel.ID, api.SendMessageData{
		Content:         content,
		Files:           files,
		AllowedMentions: ctx.Router.DefaultMentions,
	})
	if err != nil {
		return
	}

	ctx.origMessage = msg
	return
}

// SendfX ...
func (ctx *Context) SendfX(format string, args ...interface{}) (err error) {
	return ctx.SendX(fmt.Sprintf(format, args...))
}

// SendEphemeral ...
func (ctx *Context) SendEphemeral(content string, embeds ...discord.Embed) (err error) {
	_, err = ctx.Send(content, embeds...)
	return
}

// SendComponents sends a message with components
func (ctx *Context) SendComponents(components discord.ContainerComponents, content string, embeds ...discord.Embed) (*discord.Message, error) {
	return ctx.State.SendMessageComplex(ctx.Message.ChannelID, api.SendMessageData{
		Content:         content,
		Embeds:          embeds,
		Components:      components,
		AllowedMentions: ctx.Router.DefaultMentions,
	})
}
