package bcr

import "github.com/diamondburned/arikawa/v2/discord"

// AddReactionHandler adds a reaction handler for the given message
func (ctx *Context) AddReactionHandler(
	msg discord.MessageID,
	user discord.UserID,
	reaction string,
	deleteOnTrigger, deleteReaction bool,
	fn func(*Context),
) {
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
}

// AddYesNoHandler adds a reaction handler for the given message
func (ctx *Context) AddYesNoHandler(
	msg discord.Message,
	user discord.UserID,
	yesFn func(*Context),
	noFn func(*Context),
) {
	// react with the correct emojis
	ctx.Session.React(msg.ChannelID, msg.ID, discord.APIEmoji("✅"))
	ctx.Session.React(msg.ChannelID, msg.ID, discord.APIEmoji("❌"))

	// yes handler
	ctx.Router.reactions[reactionKey{
		messageID: msg.ID,
		emoji:     discord.APIEmoji("✅"),
	}] = reactionInfo{
		userID: user,
		ctx:    ctx,
		fn: func(ctx *Context) {
			yesFn(ctx)

			ctx.Router.DeleteReactions(msg.ID)
		},
		deleteOnTrigger: false,
		deleteReaction:  false,
	}

	// no handler
	ctx.Router.reactions[reactionKey{
		messageID: msg.ID,
		emoji:     discord.APIEmoji("❌"),
	}] = reactionInfo{
		userID: user,
		ctx:    ctx,
		fn: func(ctx *Context) {
			noFn(ctx)

			ctx.Router.DeleteReactions(msg.ID)
		},
		deleteOnTrigger: false,
		deleteReaction:  false,
	}
}
