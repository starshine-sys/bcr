package bcr

import (
	"github.com/diamondburned/arikawa/v2/api"
	"github.com/diamondburned/arikawa/v2/discord"
	"github.com/diamondburned/arikawa/v2/utils/json/option"
)

// Edit the given message
func (ctx *Context) Edit(m *discord.Message, c string, embed *discord.Embed) (msg *discord.Message, err error) {
	return ctx.State.EditMessageComplex(m.ChannelID, m.ID, api.EditMessageData{
		Content:         option.NewNullableString(c),
		Embed:           embed,
		AllowedMentions: ctx.Router.DefaultMentions,
	})
}
