package bcr

import (
	"context"
	"time"

	"emperror.dev/errors"
	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/utils/json/option"
)

func PaginateEmbeds(embeds ...discord.Embed) []PaginateData {
	data := make([]PaginateData, 0, len(embeds))
	for _, e := range embeds {
		data = append(data, PaginateData{
			Embeds: []discord.Embed{e},
		})
	}
	return data
}

func PaginateStrings(strings ...string) []PaginateData {
	data := make([]PaginateData, 0, len(strings))
	for _, s := range strings {
		data = append(data, PaginateData{
			Content: s,
		})
	}
	return data
}

type PaginateData struct {
	Content         string
	Embeds          []discord.Embed
	Components      discord.ContainerComponents
	AllowedMentions *api.AllowedMentions
}

func (data PaginateData) responseData(cs discord.ContainerComponent) *api.InteractionResponseData {
	mentions := &api.AllowedMentions{
		Parse: []api.AllowedMentionType{api.AllowUserMention},
	}
	if data.AllowedMentions != nil {
		mentions = data.AllowedMentions
	}

	var components discord.ContainerComponents
	if cs != nil {
		components = append(data.Components, cs)
	}

	return &api.InteractionResponseData{
		Content:         option.NewNullableString(data.Content),
		Embeds:          &data.Embeds,
		Components:      &components,
		AllowedMentions: mentions,
	}
}

const NoPaginateData = errors.Sentinel("no paginate data")

func (ctx *Context) Paginate(data []PaginateData, timeout time.Duration) (*discord.Message, context.CancelFunc, error) {
	switch len(data) {
	case 0:
		return nil, emptyFunc, NoPaginateData
	case 1:
		err := ctx.State.RespondInteraction(ctx.InteractionID, ctx.InteractionToken, api.InteractionResponse{
			Type: api.MessageInteractionWithSource,
			Data: data[0].responseData(nil),
		})
		if err != nil {
			return nil, emptyFunc, err
		}

		msg, err := ctx.Original()
		return msg, emptyFunc, err
	default:
		err := ctx.State.RespondInteraction(ctx.InteractionID, ctx.InteractionToken, api.InteractionResponse{
			Type: api.MessageInteractionWithSource,
			Data: data[0].responseData(paginateButtons),
		})
		if err != nil {
			return nil, emptyFunc, err
		}

		msg, err := ctx.Original()
		if err != nil {
			return nil, emptyFunc, err
		}

		cctx, cancel := context.WithTimeout(context.Background(), timeout)

		go ctx.paginateLoop(cctx, msg.ID, data)

		return msg, cancel, nil
	}
}

func (ctx *Context) paginateLoop(cctx context.Context, id discord.MessageID, data []PaginateData) {
	page := new(int)

	for {
		select {
		case <-cctx.Done():
			return
		default:
			v := ctx.State.WaitFor(cctx, func(i interface{}) bool {
				v, ok := i.(*gateway.InteractionCreateEvent)
				if !ok {
					return false
				}

				if v.Message == nil || v.Message.ID != id {
					return false
				}

				data, ok := v.Data.(*discord.ButtonInteraction)
				if !ok {
					return false
				}

				return contains([]discord.ComponentID{"next-page", "prev-page", "first-page", "last-page"}, data.CustomID)
			})
			if v == nil {
				return
			}

			ev, ok := v.(*gateway.InteractionCreateEvent)
			if ok {
				evData, ok := ev.Data.(*discord.ButtonInteraction)
				if ok {
					ctx.paginateInteraction(ev, evData, data, page)
				}
			}
		}
	}
}

func (ctx *Context) paginateInteraction(ev *gateway.InteractionCreateEvent, evData *discord.ButtonInteraction, data []PaginateData, page *int) error {
	switch evData.CustomID {
	case "first-page":
		*page = 0
	case "last-page":
		*page = len(data) - 1
	case "prev-page":
		if *page == 0 {
			*page = len(data) - 1
		} else {
			*page--
		}
	case "next-page":
		if *page == len(data)-1 {
			*page = 0
		} else {
			*page++
		}
	}

	return ctx.State.RespondInteraction(ev.ID, ev.Token, api.InteractionResponse{
		Type: api.UpdateMessage,
		Data: data[*page].responseData(paginateButtons),
	})
}

func contains[T comparable](slice []T, v T) bool {
	for _, s := range slice {
		if s == v {
			return true
		}
	}
	return false
}

func emptyFunc() {}

var paginateButtons discord.ContainerComponent = &discord.ActionRowComponent{
	&discord.ButtonComponent{
		Style:    discord.SecondaryButtonStyle(),
		CustomID: "first-page",
		Emoji: &discord.ComponentEmoji{
			Name: "⏪",
		},
	},
	&discord.ButtonComponent{
		Style:    discord.SecondaryButtonStyle(),
		CustomID: "prev-page",
		Emoji: &discord.ComponentEmoji{
			Name: "⬅️",
		},
	},
	&discord.ButtonComponent{
		Style:    discord.SecondaryButtonStyle(),
		CustomID: "next-page",
		Emoji: &discord.ComponentEmoji{
			Name: "➡️",
		},
	},
	&discord.ButtonComponent{
		Style:    discord.SecondaryButtonStyle(),
		CustomID: "last-page",
		Emoji: &discord.ComponentEmoji{
			Name: "⏩",
		},
	},
}
