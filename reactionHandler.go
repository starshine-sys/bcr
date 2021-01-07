package bcr

import (
	"github.com/diamondburned/arikawa/v2/discord"
	"github.com/diamondburned/arikawa/v2/gateway"
)

type reactionInfo struct {
	userID          discord.UserID
	ctx             *Context
	fn              func(*Context)
	deleteOnTrigger bool
	deleteReaction  bool
}

type reactionKey struct {
	messageID discord.MessageID
	emoji     discord.APIEmoji
}

// ReactionAdd runs when a reaction is added to a message
func (r *Router) ReactionAdd(e *gateway.MessageReactionAddEvent) {
	r.reactionMu.Lock()
	defer r.reactionMu.Unlock()
	if v, ok := r.reactions[reactionKey{
		messageID: e.MessageID,
		emoji:     e.Emoji.APIString(),
	}]; ok {
		// handle deleting the reaction
		// only delete if:
		// - the user isn't the user the reaction's for
		// - or the reaction is supposed to be deleted
		// - and the user is not the bot user
		if (v.userID != e.UserID || v.deleteReaction) && e.GuildID != 0 && e.UserID != r.Bot.ID {
			if p, err := r.Session.Permissions(e.ChannelID, r.Bot.ID); err == nil {
				if p.Has(discord.PermissionManageMessages) {
					r.Session.DeleteUserReaction(e.ChannelID, e.MessageID, e.UserID, e.Emoji.APIString())
				}
			}
		}
		// check if the reacting user is the same as the required user
		if v.userID != e.UserID {
			return
		}
		// run the handler
		v.fn(v.ctx)

		// if the handler should be deleted after running, do that
		if v.deleteOnTrigger {
			delete(r.reactions, reactionKey{
				messageID: e.MessageID,
				emoji:     e.Emoji.APIString(),
			})
		}
	}
}

// ReactionMessageDelete cleans up old handlers on deleted messages
func (r *Router) ReactionMessageDelete(m *gateway.MessageDeleteEvent) {
	r.reactionMu.Lock()
	defer r.reactionMu.Unlock()
	for k := range r.reactions {
		if k.messageID == m.ID {
			delete(r.reactions, k)
		}
	}
}

// DeleteReactions deletes all reactions for a message
func (r *Router) DeleteReactions(m discord.MessageID) {
	r.reactionMu.Lock()
	defer r.reactionMu.Unlock()
	for k := range r.reactions {
		if k.messageID == m {
			delete(r.reactions, k)
		}
	}
}
