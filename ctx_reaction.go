package bcr

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/diamondburned/arikawa/v2/gateway"

	"github.com/diamondburned/arikawa/v2/discord"
)

// AddReactionHandlerWithTimeout is like AddReactionHandler but accepts a timeout
func (ctx *Context) AddReactionHandlerWithTimeout(
	msg discord.MessageID,
	user discord.UserID,
	reaction string,
	deleteOnTrigger, deleteReaction bool,
	timeout time.Duration,
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

	// delete handlers after the set time to stop them from building up
	time.AfterFunc(ctx.Router.ReactTimeout, func() {
		ctx.Router.DeleteReactions(msg)
	})
}

// AddReactionHandler adds a reaction handler for the given message
func (ctx *Context) AddReactionHandler(
	msg discord.MessageID,
	user discord.UserID,
	reaction string,
	deleteOnTrigger, deleteReaction bool,
	fn func(*Context),
) {
	ctx.AddReactionHandlerWithTimeout(msg, user, reaction, deleteOnTrigger, deleteReaction, ctx.Router.ReactTimeout, fn)
	return
}

// YesNoHandler adds a reaction handler for the given message.
// This handler times out after one minute. If it timed out, `false` and `true` are returned, respectively.
func (ctx *Context) YesNoHandler(msg discord.Message, user discord.UserID) (yes, timeout bool) {
	return ctx.YesNoHandlerWithTimeout(msg, user, time.Minute)
}

// YesNoHandlerWithTimeout is like YesNoHandler but lets you specify your own timeout.
func (ctx *Context) YesNoHandlerWithTimeout(msg discord.Message, user discord.UserID, t time.Duration) (yes, timeout bool) {
	c, cancel := context.WithTimeout(context.Background(), t)

	go func() {
		// react with the correct emojis
		// this is run in a goroutine to add the handler immediately
		ctx.State.React(msg.ChannelID, msg.ID, discord.APIEmoji("✅"))
		ctx.State.React(msg.ChannelID, msg.ID, discord.APIEmoji("❌"))
	}()

	defer cancel()
	ev := ctx.State.WaitFor(c, func(ev interface{}) bool {
		// first, try a reaction add event
		v, ok := ev.(*gateway.MessageReactionAddEvent)
		if !ok {
			// try a message create event
			m, ok := ev.(*gateway.MessageCreateEvent)
			if !ok {
				return false
			}

			// return true if
			// - the channel ID and user ID are correct
			// - the message is "yes", "y", "no", or "n"
			return m.ChannelID == msg.ChannelID && m.Author.ID == user &&
				(strings.ToLower(m.Content) == "yes" || strings.ToLower(m.Content) == "y" || strings.ToLower(m.Content) == "no" || strings.ToLower(m.Content) == "n")
		}

		// return true if
		// - the channel ID, message ID, and user ID are correct
		// - the emoji is the yes or no emoji
		return v.ChannelID == msg.ChannelID && v.MessageID == msg.ID && v.UserID == user &&
			(v.Emoji.APIString() == "✅" || v.Emoji.APIString() == "❌")
	})

	// if the event timed out, return
	if ev == nil {
		return false, true
	}

	// try a message reaction event
	v, ok := ev.(*gateway.MessageReactionAddEvent)
	if !ok {
		m, ok := ev.(*gateway.MessageCreateEvent)
		if !ok {
			return false, false
		}
		// if not, it's a message create event
		return strings.EqualFold(m.Content, "yes") || strings.EqualFold(m.Content, "y"), false
	}
	return v.Emoji.APIString() == "✅", false
}

var (
	// ErrorTimedOut is returned when WaitForReaction times out
	ErrorTimedOut = errors.New("context: timed out waiting for reaction")
	// ErrorFailedConversion is returned when WaitForReaction can't convert the interface{} to a MessageReactionAddEvent
	ErrorFailedConversion = errors.New("context: failed conversion in WaitForReaction")
)

// WaitForReaction calls WaitForReactionWithTimeout with a 3-minute timeout
func (ctx *Context) WaitForReaction(msg discord.Message, user discord.UserID) (ev *gateway.MessageReactionAddEvent, err error) {
	return ctx.WaitForReactionWithTimeout(msg, user, 5*time.Minute)
}

// WaitForReactionWithTimeout waits for a reaction with a user-given timeout
func (ctx *Context) WaitForReactionWithTimeout(msg discord.Message, user discord.UserID, timeout time.Duration) (ev *gateway.MessageReactionAddEvent, err error) {
	c, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	v := ctx.State.WaitFor(c, func(ev interface{}) bool {
		v, ok := ev.(*gateway.MessageReactionAddEvent)
		if !ok {
			return false
		}
		return v.ChannelID == msg.ChannelID && v.MessageID == msg.ID && v.UserID == user
	})

	if v == nil {
		return nil, ErrorTimedOut
	}

	ev, ok := v.(*gateway.MessageReactionAddEvent)
	if !ok {
		return nil, ErrorFailedConversion
	}
	return ev, nil
}

// Confirm confirms the given string to the context user
func (ctx *Context) Confirm(s string) (yes bool) {
	m, err := ctx.Sendf("Are you sure you want to %v?", s)
	if err != nil {
		ctx.Router.Logger.Error("sending message: %v", err)
		return false
	}

	yes, timeout := ctx.YesNoHandlerWithTimeout(*m, ctx.Author.ID, 3*time.Minute)
	if timeout {
		_, err = ctx.Send(":x: Operation timed out.", nil)
		if err != nil {
			ctx.Router.Logger.Error("sending message: %v", err)
		}
		return false
	}

	if !yes {
		_, err = ctx.Send(":x: Operation cancelled.", nil)
		if err != nil {
			ctx.Router.Logger.Error("sending message: %v", err)
		}
		return false
	}

	return true
}
