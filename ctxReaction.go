package bcr

import (
	"context"
	"time"

	"github.com/diamondburned/arikawa/v2/gateway"

	"github.com/diamondburned/arikawa/v2/discord"
)

// AddReactionHandler adds a reaction handler for the given message
func (ctx *Context) AddReactionHandler(
	msg discord.MessageID,
	user discord.UserID,
	reaction string,
	deleteOnTrigger, deleteReaction bool,
	fn func(*Context),
) {
	ctx.Router.reactionMu.Lock()

	ctx.Router.reactions[reactionKey{
		messageID: msg,
		emoji:     discord.APIEmoji(reaction),
	}] = reactionInfo{
		userID:          user,
		ctx:             ctx,
		fn:              fn,
		deleteOnTrigger: deleteOnTrigger,
		deleteReaction:  deleteReaction,
	}

	ctx.Router.reactionMu.Unlock()

	// delete handlers after 15 minutes to stop them from building up
	time.AfterFunc(15*time.Minute, func() {
		ctx.Router.DeleteReactions(msg)
	})
}

// YesNoHandler adds a reaction handler for the given message.
// This handler times out after one minute. If it timed out, `false` and `true` are returned, respectively.
func (ctx *Context) YesNoHandler(msg discord.Message, user discord.UserID) (yes, timeout bool) {
	c, cancel := context.WithTimeout(context.Background(), time.Minute)

	go func() {
		// react with the correct emojis
		// this is run in a goroutine to add the handler immediately
		ctx.Session.React(msg.ChannelID, msg.ID, discord.APIEmoji("✅"))
		ctx.Session.React(msg.ChannelID, msg.ID, discord.APIEmoji("❌"))
	}()

	defer cancel()
	ev := ctx.Session.WaitFor(c, func(ev interface{}) bool {
		v, ok := ev.(*gateway.MessageReactionAddEvent)
		if !ok {
			return false
		}
		return v.ChannelID == msg.ChannelID && v.MessageID == msg.ID && v.UserID == user &&
			(v.Emoji.APIString() == "✅" || v.Emoji.APIString() == "❌")
	})

	if ev == nil {
		return false, true
	}
	v, ok := ev.(*gateway.MessageReactionAddEvent)
	if !ok {
		return false, false
	}
	return v.Emoji.APIString() == "✅", false
}
