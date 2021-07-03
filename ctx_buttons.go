package bcr

import (
	"context"
	"strings"
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

	ctx.State.EditMessageComplex(msg.ChannelID, msg.ID, api.EditMessageData{
		Components: &[]discord.Component{},
	})

	if v == nil {
		return false, true
	}

	return
}
