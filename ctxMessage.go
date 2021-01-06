package bcr

import "github.com/diamondburned/arikawa/v2/discord"

// AddMessageHandler adds a message handler for the given user/channel
func (ctx *Context) AddMessageHandler(
	c discord.ChannelID,
	user discord.UserID,
	fn func(*Context, discord.Message),
) {
	ctx.Router.messages[messageKey{
		channelID: c,
		userID:    user,
	}] = messageInfo{
		ctx: ctx,
		fn:  fn,
	}
}
