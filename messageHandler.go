package bcr

import (
	"github.com/diamondburned/arikawa/v2/discord"
	"github.com/diamondburned/arikawa/v2/gateway"
)

type messageInfo struct {
	ctx *Context
	fn  func(*Context, discord.Message)
}

type messageKey struct {
	channelID discord.ChannelID
	userID    discord.UserID
}

// MsgHandlerCreate runs when a new message is sent
func (r *Router) MsgHandlerCreate(e *gateway.MessageCreateEvent) {
	// if the author is a bot, return
	if e.Author.Bot {
		return
	}
	if v, ok := r.messages[messageKey{
		channelID: e.ChannelID,
		userID:    e.Author.ID,
	}]; ok {
		// run the handler
		v.fn(v.ctx, e.Message)

		// delete the handler
		delete(r.messages, messageKey{
			channelID: e.ChannelID,
			userID:    e.Author.ID,
		})
	}
}
