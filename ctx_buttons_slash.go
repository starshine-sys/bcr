package bcr

import (
	"context"
	"sync"
	"time"

	"emperror.dev/errors"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/utils/json/option"
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
	return ctx.ButtonPagesWithComponents(embeds, timeout, nil)
}

// ButtonPagesWithComponents is like ButtonPages but adds the given components before the buttons used for pagination.
func (ctx *SlashContext) ButtonPagesWithComponents(embeds []discord.Embed, timeout time.Duration, components []discord.Component) (msg *discord.Message, rmFunc func(), err error) {
	rmFunc = func() {}

	if len(embeds) == 0 {
		return nil, func() {}, errors.New("no embeds")
	}

	ctx.AdditionalParams["page"] = 0

	if len(embeds) == 1 {
		err = ctx.State.RespondInteraction(ctx.InteractionID, ctx.InteractionToken, api.InteractionResponse{
			Type: api.MessageInteractionWithSource,
			Data: &api.InteractionResponseData{
				Embeds:     &[]discord.Embed{embeds[0]},
				Components: &components,
			},
		})
		if err != nil {
			return
		}

		msg, err = ctx.Original()
		return
	}

	components = append(components, discord.ActionRowComponent{
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
	})

	err = ctx.State.RespondInteraction(ctx.InteractionID, ctx.InteractionToken, api.InteractionResponse{
		Type: api.MessageInteractionWithSource,
		Data: &api.InteractionResponseData{
			Embeds:     &[]discord.Embed{embeds[0]},
			Components: &components,
		},
	})
	if err != nil {
		return
	}

	msg, err = ctx.Original()

	page := 0

	prev := ctx.AddButtonHandler(msg.ID, ctx.Author.ID, "prev", false, func(ctx *SlashContext, ev *gateway.InteractionCreateEvent) {
		if page == 0 {
			page = len(embeds) - 1
			ctx.AdditionalParams["page"] = len(embeds) - 1
		} else {
			page--
			ctx.AdditionalParams["page"] = page
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

	next := ctx.AddButtonHandler(msg.ID, ctx.Author.ID, "next", false, func(ctx *SlashContext, ev *gateway.InteractionCreateEvent) {
		if page >= len(embeds)-1 {
			page = 0
			ctx.AdditionalParams["page"] = 0
		} else {
			page++
			ctx.AdditionalParams["page"] = page
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

	first := ctx.AddButtonHandler(msg.ID, ctx.Author.ID, "first", false, func(ctx *SlashContext, ev *gateway.InteractionCreateEvent) {
		page = 0
		ctx.AdditionalParams["page"] = 0

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

	last := ctx.AddButtonHandler(msg.ID, ctx.Author.ID, "last", false, func(ctx *SlashContext, ev *gateway.InteractionCreateEvent) {
		page = len(embeds) - 1
		ctx.AdditionalParams["page"] = len(embeds) - 1

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

	cross := ctx.AddButtonHandler(msg.ID, ctx.Author.ID, "cross", false, func(ctx *SlashContext, ev *gateway.InteractionCreateEvent) {
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

// ConfirmButton confirms a prompt with buttons or "yes"/"no" messages.
func (ctx *SlashContext) ConfirmButton(userID discord.UserID, data ConfirmData) (yes, timeout bool) {
	if data.Message == "" && len(data.Embeds) == 0 {
		return
	}

	if data.YesPrompt == "" {
		data.YesPrompt = "Confirm"
	}
	if data.YesStyle == 0 {
		data.YesStyle = discord.PrimaryButton
	}
	if data.NoPrompt == "" {
		data.NoPrompt = "Cancel"
	}
	if data.NoStyle == 0 {
		data.NoStyle = discord.SecondaryButton
	}
	if data.Timeout == 0 {
		data.Timeout = time.Minute
	}

	con, cancel := context.WithTimeout(context.Background(), data.Timeout)
	defer cancel()

	err := ctx.State.RespondInteraction(ctx.InteractionID, ctx.InteractionToken, api.InteractionResponse{
		Data: &api.InteractionResponseData{
			Content: option.NewNullableString(data.Message),
			Embeds:  &data.Embeds,

			Components: &[]discord.Component{
				discord.ActionRowComponent{
					Components: []discord.Component{
						discord.ButtonComponent{
							Label:    data.YesPrompt,
							Style:    data.YesStyle,
							CustomID: "yes",
						},
						discord.ButtonComponent{
							Label:    data.NoPrompt,
							Style:    data.NoStyle,
							CustomID: "no",
						},
					},
				},
			},
		},
	})
	if err != nil {
		return
	}

	msg, err := ctx.Original()
	if err != nil {
		return
	}

	v := ctx.State.WaitFor(con, func(ev interface{}) bool {
		v, ok := ev.(*gateway.InteractionCreateEvent)
		if !ok {
			return false
		}

		if v.Message == nil || (v.Member == nil && v.User == nil) {
			return false
		}

		if v.Message.ID != msg.ID {
			return false
		}

		var uID discord.UserID
		if v.Member != nil {
			uID = v.Member.User.ID
		} else {
			uID = v.User.ID
		}

		if uID != userID {
			return false
		}

		if v.Data.CustomID == "" {
			return false
		}

		yes = v.Data.CustomID == "yes"
		timeout = false
		return true
	})

	if v == nil {
		return false, true
	}

	upd := &[]discord.Component{
		discord.ActionRowComponent{
			Components: []discord.Component{
				discord.ButtonComponent{
					Label:    data.YesPrompt,
					Style:    data.YesStyle,
					CustomID: "yes",
					Disabled: true,
				},
				discord.ButtonComponent{
					Label:    data.NoPrompt,
					Style:    data.NoStyle,
					CustomID: "no",
					Disabled: true,
				},
			},
		},
	}

	if ev, ok := v.(*gateway.InteractionCreateEvent); ok {
		ctx.State.RespondInteraction(ev.ID, ev.Token, api.InteractionResponse{
			Type: api.UpdateMessage,
			Data: &api.InteractionResponseData{
				Components: upd,
			},
		})
	}

	return
}
