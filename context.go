package bcr

import (
	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/diamondburned/arikawa/v3/utils/json/option"
)

// Context is the root context embedded into CommandContext, ModalContext etc
type Context struct {
	InteractionID    discord.InteractionID
	InteractionToken string
	Event            *gateway.InteractionCreateEvent

	State *state.State

	User          discord.User
	Member        *discord.Member
	Guild         *discord.Guild
	Channel       *discord.Channel
	ParentChannel *discord.Channel

	deferred  bool
	Responded bool
}

func (r *Router) NewRootContext(ic *gateway.InteractionCreateEvent) (ctx *Context, err error) {
	ctx = &Context{
		InteractionID:    ic.ID,
		InteractionToken: ic.Token,
		Event:            ic,
	}

	if ic.Member != nil {
		ctx.Member = ic.Member
		ctx.User = ic.Member.User
	} else {
		ctx.User = *ic.User
	}

	state, _ := r.ShardFromGuildID(ic.GuildID)
	ctx.State = state

	g, ch, parentCh, err := r.CollectFunc(state, ic.GuildID, ic.ChannelID)
	if err != nil {
		return ctx, err
	}

	ctx.Guild = g
	ctx.Channel = ch
	ctx.ParentChannel = parentCh

	return ctx, nil
}

func (ctx *Context) Defer() error {
	err := ctx.State.RespondInteraction(ctx.InteractionID, ctx.InteractionToken, api.InteractionResponse{
		Type: api.DeferredMessageInteractionWithSource,
	})
	if err != nil {
		return err
	}

	ctx.deferred = true
	return nil
}

func (ctx *Context) DeferEphemeral() error {
	err := ctx.State.RespondInteraction(ctx.InteractionID, ctx.InteractionToken, api.InteractionResponse{
		Type: api.DeferredMessageInteractionWithSource,
		Data: &api.InteractionResponseData{
			Flags: discord.EphemeralMessage,
		},
	})
	if err != nil {
		return err
	}

	ctx.deferred = true
	return nil
}

func (ctx *Context) Reply(content string, embeds ...discord.Embed) error {
	return ctx.reply(0, content, embeds)
}

func (ctx *Context) ReplyEphemeral(content string, embeds ...discord.Embed) error {
	return ctx.reply(discord.EphemeralMessage, content, embeds)
}

func (ctx *Context) reply(flags discord.MessageFlags, content string, embeds []discord.Embed) error {
	if ctx.deferred {
		_, err := ctx.State.EditInteractionResponse(ctx.Event.AppID, ctx.InteractionToken, api.EditInteractionResponseData{
			Content: option.NewNullableString(content),
			Embeds:  &embeds,
		})
		if err != nil {
			return err
		}

		ctx.Responded = true
		return nil
	}

	err := ctx.State.RespondInteraction(ctx.InteractionID, ctx.InteractionToken, api.InteractionResponse{
		Type: api.MessageInteractionWithSource,
		Data: &api.InteractionResponseData{
			Content: option.NewNullableString(content),
			Embeds:  &embeds,
			Flags:   flags,
		},
	})
	if err != nil {
		return err
	}

	ctx.Responded = true
	return nil
}

// Ctx implements HasContext
func (ctx *Context) Ctx() *Context {
	return ctx
}

// Original returns the original response to an interaction, if any.
func (ctx *Context) Original() (msg *discord.Message, err error) {
	url := api.EndpointWebhooks + ctx.Event.AppID.String() + "/" + ctx.InteractionToken + "/messages/@original"

	return msg, ctx.State.RequestJSON(&msg, "GET", url)
}

func (ctx *Context) ReplyComplex(data api.InteractionResponseData) error {
	if ctx.deferred {
		_, err := ctx.State.EditInteractionResponse(ctx.Event.AppID, ctx.InteractionToken, api.EditInteractionResponseData{
			Content:         data.Content,
			Embeds:          data.Embeds,
			Components:      data.Components,
			AllowedMentions: data.AllowedMentions,
			Files:           data.Files,
		})
		if err != nil {
			return err
		}

		ctx.Responded = true
		return nil
	}

	err := ctx.State.RespondInteraction(ctx.InteractionID, ctx.InteractionToken, api.InteractionResponse{
		Type: api.MessageInteractionWithSource,
		Data: &data,
	})
	if err != nil {
		return err
	}

	ctx.Responded = true
	return nil
}

func (ctx *Context) Followup(appID discord.AppID, data api.InteractionResponseData) (*discord.Message, error) {
	msg, err := ctx.State.FollowUpInteraction(appID, ctx.InteractionToken, data)
	return msg, err
}
