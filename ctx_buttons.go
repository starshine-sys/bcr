package bcr

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
)

// ConfirmData is the data for ctx.ConfirmButton()
type ConfirmData struct {
	Message string
	Embeds  []discord.Embed

	// Defaults to "Confirm"
	YesPrompt string
	// Defaults to a primary button
	YesStyle discord.ButtonStyle
	// Defaults to "Cancel"
	NoPrompt string
	// Defaults to a secondary button
	NoStyle discord.ButtonStyle

	// Defaults to one minute
	Timeout time.Duration
}

// ConfirmButton confirms a prompt with buttons or "yes"/"no" messages.
func (ctx *Context) ConfirmButton(userID discord.UserID, data ConfirmData) (yes, timeout bool) {
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

	msg, err := ctx.State.SendMessageComplex(ctx.Message.ChannelID, api.SendMessageData{
		Content: data.Message,
		Embeds:  data.Embeds,

		Components: []discord.Component{
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
	})
	if err != nil {
		return
	}

	v := ctx.State.WaitFor(con, func(ev interface{}) bool {
		v, ok := ev.(*gateway.InteractionCreateEvent)
		if ok {
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
		}

		m, ok := ev.(*gateway.MessageCreateEvent)
		if ok {
			if m.ChannelID != msg.ChannelID || m.Author.ID != ctx.Author.ID {
				return false
			}

			switch strings.ToLower(m.Content) {
			case "yes", "y", strings.ToLower(data.YesPrompt):
				yes = true
				timeout = false
				return true
			case "no", "n", strings.ToLower(data.NoPrompt):
				yes = false
				timeout = false
				return true
			default:
				return false
			}
		}

		return false
	})

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

	ctx.State.EditMessageComplex(msg.ChannelID, msg.ID, api.EditMessageData{
		Components: upd,
	})

	if v == nil {
		return false, true
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

type buttonKey struct {
	msg      discord.MessageID
	user     discord.UserID
	customID string
}

type buttonInfo struct {
	ctx    *Context
	fn     func(*Context, *gateway.InteractionCreateEvent)
	delete bool
}

// ButtonRemoveFunc is returned by AddButtonHandler
type ButtonRemoveFunc func()

// AddButtonHandler adds a handler for the given message ID, user ID, and custom ID
func (ctx *Context) AddButtonHandler(
	msg discord.MessageID,
	user discord.UserID,
	customID string,
	del bool,
	fn func(*Context, *gateway.InteractionCreateEvent),
) ButtonRemoveFunc {
	ctx.Router.buttonMu.Lock()
	defer ctx.Router.buttonMu.Unlock()

	ctx.Router.buttons[buttonKey{msg, user, customID}] = buttonInfo{ctx, fn, del}

	return func() {
		ctx.Router.buttonMu.Lock()
		delete(ctx.Router.buttons, buttonKey{msg, user, customID})
		ctx.Router.buttonMu.Unlock()
	}
}

// ButtonHandler handles buttons added by ctx.AddButtonHandler
func (r *Router) ButtonHandler(ev *gateway.InteractionCreateEvent) {
	if ev.Type != gateway.ButtonInteraction {
		return
	}

	if ev.Message == nil ||
		(ev.Member == nil && ev.User == nil) ||
		ev.Data == nil {
		return
	}
	if ev.Data.CustomID == "" {
		return
	}

	var user discord.UserID
	if ev.Member != nil {
		user = ev.Member.User.ID
	} else {
		user = ev.User.ID
	}

	r.buttonMu.RLock()
	info, ok := r.buttons[buttonKey{ev.Message.ID, user, ev.Data.CustomID}]
	r.buttonMu.RUnlock()

	if !ok {
		r.slashButton(ev, user)
		return
	}

	info.fn(info.ctx, ev)

	if info.delete {
		r.buttonMu.Lock()
		delete(r.buttons, buttonKey{ev.Message.ID, user, ev.Data.CustomID})
		r.buttonMu.Unlock()
	}
}

func (r *Router) slashButton(ev *gateway.InteractionCreateEvent, user discord.UserID) {
	r.slashButtonMu.RLock()
	info, ok := r.slashButtons[buttonKey{ev.Message.ID, user, ev.Data.CustomID}]
	r.slashButtonMu.RUnlock()

	if !ok {
		return
	}

	info.fn(info.ctx, ev)

	if info.delete {
		r.slashButtonMu.Lock()
		delete(r.slashButtons, buttonKey{ev.Message.ID, user, ev.Data.CustomID})
		r.slashButtonMu.Unlock()
	}
}

// ButtonPages is like PagedEmbed but uses buttons instead of reactions.
func (ctx *Context) ButtonPages(embeds []discord.Embed, timeout time.Duration) (msg *discord.Message, rmFunc func(), err error) {
	return ctx.ButtonPagesWithComponents(embeds, timeout, nil)
}

// ButtonPagesWithComponents is like ButtonPages but adds the given components before the buttons used for pagination.
func (ctx *Context) ButtonPagesWithComponents(embeds []discord.Embed, timeout time.Duration, components []discord.Component) (msg *discord.Message, rmFunc func(), err error) {
	rmFunc = func() {}

	if len(embeds) == 0 {
		return nil, func() {}, errors.New("no embeds")
	}

	if len(embeds) == 1 {
		msg, err = ctx.State.SendEmbeds(ctx.Message.ChannelID, embeds[0])
		return
	}

	components = append(components, []discord.Component{discord.ActionRowComponent{
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
	}}...)

	msg, err = ctx.State.SendMessageComplex(ctx.Message.ChannelID, api.SendMessageData{
		Embeds:     []discord.Embed{embeds[0]},
		Components: components,
	})
	if err != nil {
		return
	}

	page := 0

	prev := ctx.AddButtonHandler(msg.ID, ctx.Author.ID, "prev", false, func(ctx *Context, ev *gateway.InteractionCreateEvent) {
		if page == 0 {
			page = len(embeds) - 1
		} else {
			page--
		}

		ctx.State.RespondInteraction(ev.ID, ev.Token, api.InteractionResponse{
			Type: api.UpdateMessage,
			Data: &api.InteractionResponseData{
				Embeds: &[]discord.Embed{embeds[page]},
			},
		})
	})

	next := ctx.AddButtonHandler(msg.ID, ctx.Author.ID, "next", false, func(ctx *Context, ev *gateway.InteractionCreateEvent) {
		if page >= len(embeds)-1 {
			page = 0
		} else {
			page++
		}

		ctx.State.RespondInteraction(ev.ID, ev.Token, api.InteractionResponse{
			Type: api.UpdateMessage,
			Data: &api.InteractionResponseData{
				Embeds: &[]discord.Embed{embeds[page]},
			},
		})
	})

	first := ctx.AddButtonHandler(msg.ID, ctx.Author.ID, "first", false, func(ctx *Context, ev *gateway.InteractionCreateEvent) {
		page = 0

		ctx.State.RespondInteraction(ev.ID, ev.Token, api.InteractionResponse{
			Type: api.UpdateMessage,
			Data: &api.InteractionResponseData{
				Embeds: &[]discord.Embed{embeds[page]},
			},
		})
	})

	last := ctx.AddButtonHandler(msg.ID, ctx.Author.ID, "last", false, func(ctx *Context, ev *gateway.InteractionCreateEvent) {
		page = len(embeds) - 1

		ctx.State.RespondInteraction(ev.ID, ev.Token, api.InteractionResponse{
			Type: api.UpdateMessage,
			Data: &api.InteractionResponseData{
				Embeds: &[]discord.Embed{embeds[page]},
			},
		})
	})

	var o sync.Once

	cross := ctx.AddButtonHandler(msg.ID, ctx.Author.ID, "cross", false, func(ctx *Context, ev *gateway.InteractionCreateEvent) {
		ctx.State.EditMessageComplex(msg.ChannelID, msg.ID, api.EditMessageData{
			Components: &[]discord.Component{},
		})
	})

	rmFunc = func() {
		o.Do(func() {
			ctx.State.EditMessageComplex(msg.ChannelID, msg.ID, api.EditMessageData{
				Components: &[]discord.Component{},
			})

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
