package bcr

import (
	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/utils/json/option"
)

// Edit the given message
func (ctx *Context) Edit(m *discord.Message, c string, editEmbeds bool, embeds ...discord.Embed) (msg *discord.Message, err error) {
	e := &embeds
	if !editEmbeds {
		e = nil
	}

	return ctx.State.EditMessageComplex(m.ChannelID, m.ID, api.EditMessageData{
		Content:         option.NewNullableString(c),
		Embeds:          e,
		AllowedMentions: ctx.Router.DefaultMentions,
	})
}
