package bcr

import (
	"context"
	"time"

	"github.com/diamondburned/arikawa/v2/discord"
	"github.com/diamondburned/arikawa/v2/gateway"
)

// AddMessageHandler adds a message handler for the given user/channel
func (ctx *Context) AddMessageHandler(
	c discord.ChannelID,
	user discord.UserID,
	fn func(*Context, discord.Message),
) {
	ctx.Router.messageMu.Lock()
	defer ctx.Router.messageMu.Unlock()
	ctx.Router.messages[messageKey{
		channelID: c,
		userID:    user,
	}] = messageInfo{
		ctx: ctx,
		fn:  fn,
	}
}

// WaitForMessage waits for a message that matches the given channel ID, user ID, and filter function.
// If filter is nil, only checks for the channel and user matching.
func (ctx *Context) WaitForMessage(ch discord.ChannelID, user discord.UserID, timeout time.Duration, filter func(*gateway.MessageCreateEvent) bool) (msg *gateway.MessageCreateEvent, timedOut bool) {
	c, cancel := context.WithTimeout(context.Background(), timeout)

	defer cancel()

	ev := ctx.State.WaitFor(c, func(ev interface{}) bool {
		v, ok := ev.(*gateway.MessageCreateEvent)
		if !ok {
			return false
		}

		if filter != nil {
			if !filter(v) {
				return false
			}
		}

		return v.ChannelID == ch && v.Author.ID == user
	})

	if ev == nil {
		return nil, true
	}

	return ev.(*gateway.MessageCreateEvent), false
}
