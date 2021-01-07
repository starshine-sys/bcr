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

	r.messageMu.RLock()
	defer r.messageMu.RUnlock()
	if v, ok := r.messages[messageKey{
		channelID: e.ChannelID,
		userID:    e.Author.ID,
	}]; ok {
		r.messageMu.RUnlock()
		// delete the handler
		r.messageMu.Lock()
		delete(r.messages, messageKey{
			channelID: e.ChannelID,
			userID:    e.Author.ID,
		})
		r.messageMu.Unlock()

		// run the handler
		v.fn(v.ctx, e.Message)
	}
}
