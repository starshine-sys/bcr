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
