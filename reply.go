package bcr

import (
	"github.com/diamondburned/arikawa/v2/api"
	"github.com/diamondburned/arikawa/v2/discord"
)

// Send sends a message to the context channel
func (ctx *Context) Send(content string, embed *discord.Embed) (m *discord.Message, err error) {
	return ctx.Session.SendMessageComplex(ctx.Channel.ID, api.SendMessageData{
		Content:         content,
		Embed:           embed,
		AllowedMentions: ctx.Router.DefaultMentions,
	})
}

// Reply *replies* to the original message in the context channel
func (ctx *Context) Reply(content string, embed *discord.Embed) (m *discord.Message, err error) {
	return ctx.Session.SendMessageComplex(ctx.Channel.ID, api.SendMessageData{
		Content:         content,
		Embed:           embed,
		AllowedMentions: ctx.Router.DefaultMentions,

		Reference: &discord.MessageReference{
			MessageID: ctx.Message.ID,
		},
	})
}
