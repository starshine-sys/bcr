package bcr

import (
	"sync"
	"time"

	"emperror.dev/errors"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
)

type slashButtonInfo struct {
	ctx    *SlashContext
	fn     func(*SlashContext, *gateway.InteractionCreateEvent)
	delete bool
}

// AddButtonHandler adds a handler for the given message ID, user ID, and custom ID
func (ctx *SlashContext) AddButtonHandler(
	msg discord.MessageID,
	user discord.UserID,
	customID string,
	del bool,
	fn func(*SlashContext, *gateway.InteractionCreateEvent),
) ButtonRemoveFunc {
	ctx.Router.slashButtonMu.Lock()
	defer ctx.Router.slashButtonMu.Unlock()

	ctx.Router.slashButtons[buttonKey{msg, user, customID}] = slashButtonInfo{ctx, fn, del}

	return func() {
		ctx.Router.slashButtonMu.Lock()
		delete(ctx.Router.slashButtons, buttonKey{msg, user, customID})
		ctx.Router.slashButtonMu.Unlock()
	}
}

// ButtonPages is like PagedEmbed but uses buttons instead of reactions.
func (ctx *SlashContext) ButtonPages(embeds []discord.Embed, timeout time.Duration) (msg *discord.Message, rmFunc func(), err error) {
	rmFunc = func() {}

	if len(embeds) == 0 {
		return nil, func() {}, errors.New("no embeds")
	}

	if len(embeds) == 1 {
		err = ctx.SendX("", embeds[0])
		if err != nil {
			return
		}

		msg, err = ctx.Original()
		return
	}

	err = ctx.State.RespondInteraction(ctx.InteractionID, ctx.InteractionToken, api.InteractionResponse{
		Type: api.MessageInteractionWithSource,
		Data: &api.InteractionResponseData{
			Embeds: &[]discord.Embed{embeds[0]},
			Components: &[]discord.Component{discord.ActionRowComponent{
				Components: []discord.Component{
					discord.ButtonComponent{
						Emoji: &discord.ButtonEmoji{
							Name: "⏪",
						},
						Style:    discord.SecondaryButton,
						CustomID: "first",
					},
					discord.ButtonComponent{
						Emoji: &discord.ButtonEmoji{
							Name: "⬅️",
						},
						Style:    discord.SecondaryButton,
						CustomID: "prev",
					},
					discord.ButtonComponent{
						Emoji: &discord.ButtonEmoji{
							Name: "➡️",
						},
						Style:    discord.SecondaryButton,
						CustomID: "next",
					},
					discord.ButtonComponent{
						Emoji: &discord.ButtonEmoji{
							Name: "⏩",
						},
						Style:    discord.SecondaryButton,
						CustomID: "last",
					},
					discord.ButtonComponent{
						Emoji: &discord.ButtonEmoji{
							Name: "❌",
						},
						Style:    discord.SecondaryButton,
						CustomID: "cross",
					},
				},
			}},
		},
	})
	if err != nil {
		return
	}

	msg, err = ctx.Original()

	page := 0

	prev := ctx.AddButtonHandler(msg.ID, ctx.User.ID, "prev", false, func(ctx *SlashContext, ev *gateway.InteractionCreateEvent) {
		if page == 0 {
			page = len(embeds) - 1
		} else {
			page--
		}

		err = ctx.State.RespondInteraction(ev.ID, ev.Token, api.InteractionResponse{
			Type: api.UpdateMessage,
			Data: &api.InteractionResponseData{
				Embeds: &[]discord.Embed{embeds[page]},
			},
		})
		if err != nil {
			ctx.Router.Logger.Error("Editing message: %v", err)
		}
	})

	next := ctx.AddButtonHandler(msg.ID, ctx.User.ID, "next", false, func(ctx *SlashContext, ev *gateway.InteractionCreateEvent) {
		if page >= len(embeds)-1 {
			page = 0
		} else {
			page++
		}

		err = ctx.State.RespondInteraction(ev.ID, ev.Token, api.InteractionResponse{
			Type: api.UpdateMessage,
			Data: &api.InteractionResponseData{
				Embeds: &[]discord.Embed{embeds[page]},
			},
		})
		if err != nil {
			ctx.Router.Logger.Error("Editing message: %v", err)
		}
	})

	first := ctx.AddButtonHandler(msg.ID, ctx.User.ID, "first", false, func(ctx *SlashContext, ev *gateway.InteractionCreateEvent) {
		page = 0

		err = ctx.State.RespondInteraction(ev.ID, ev.Token, api.InteractionResponse{
			Type: api.UpdateMessage,
			Data: &api.InteractionResponseData{
				Embeds: &[]discord.Embed{embeds[page]},
			},
		})
		if err != nil {
			ctx.Router.Logger.Error("Editing message: %v", err)
		}
	})

	last := ctx.AddButtonHandler(msg.ID, ctx.User.ID, "last", false, func(ctx *SlashContext, ev *gateway.InteractionCreateEvent) {
		page = len(embeds) - 1

		err = ctx.State.RespondInteraction(ev.ID, ev.Token, api.InteractionResponse{
			Type: api.UpdateMessage,
			Data: &api.InteractionResponseData{
				Embeds: &[]discord.Embed{embeds[page]},
			},
		})
		if err != nil {
			ctx.Router.Logger.Error("Editing message: %v", err)
		}
	})

	var o sync.Once

	cross := ctx.AddButtonHandler(msg.ID, ctx.User.ID, "cross", false, func(ctx *SlashContext, ev *gateway.InteractionCreateEvent) {
		err = ctx.State.RespondInteraction(ev.ID, ev.Token, api.InteractionResponse{
			Type: api.UpdateMessage,
			Data: &api.InteractionResponseData{
				Components: &[]discord.Component{},
			},
		})
		if err != nil {
			ctx.Router.Logger.Error("Editing message: %v", err)
		}
	})

	rmFunc = func() {
		o.Do(func() {
			_, err = ctx.State.EditMessageComplex(msg.ChannelID, msg.ID, api.EditMessageData{
				Components: &[]discord.Component{},
			})
			if err != nil {
				ctx.Router.Logger.Error("Editing message: %v", err)
			}

			prev()
			next()
			first()
			last()
			cross()
		})
	}

	time.AfterFunc(timeout, rmFunc)
	return msg, rmFunc, err
}
