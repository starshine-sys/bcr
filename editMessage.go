package bcr

import (
	"github.com/diamondburned/arikawa/v2/api"
	"github.com/diamondburned/arikawa/v2/discord"
	"github.com/diamondburned/arikawa/v2/utils/json/option"
)

// EditMessage is a struct used for preparing Edit commands.
// This struct's methods are made for chaining together, with a final Send() call.
type EditMessage struct {
	ctx *Context

	channelID discord.ChannelID
	messageID discord.MessageID

	Content string
	Embed   *discord.Embed

	AllowedMentions *api.AllowedMentions
}

// Edit creates a new EditMessage struct
func (ctx *Context) Edit(m *discord.Message) *EditMessage {
	return &EditMessage{
		ctx: ctx,

		channelID: m.ChannelID,
		messageID: m.ID,

		AllowedMentions: ctx.Router.DefaultMentions,
	}
}

// SetContent sets the content
func (e *EditMessage) SetContent(c string) *EditMessage {
	e.Content = c
	return e
}

// SetEmbed sets the embed
func (e *EditMessage) SetEmbed(embed *discord.Embed) *EditMessage {
	e.Embed = embed
	return e
}

// Send sends the edit
func (e *EditMessage) Send() (m *discord.Message, err error) {
	return e.ctx.Session.EditMessageComplex(e.channelID, e.messageID, api.EditMessageData{
		Content:         option.NewNullableString(e.Content),
		Embed:           e.Embed,
		AllowedMentions: e.AllowedMentions,
	})
}
