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

	deferred bool
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

	// get guild
	if ic.GuildID.IsValid() {
		ctx.Guild, err = ctx.State.Guild(ic.GuildID)
		if err != nil {
			return ctx, err
		}
		ctx.Guild.Roles, err = ctx.State.Roles(ic.GuildID)
		if err != nil {
			return ctx, err
		}
	}

	// get the channel
	ctx.Channel, err = ctx.State.Channel(ic.ChannelID)
	if err != nil {
		return ctx, err
	}

	if IsThread(ctx.Channel) {
		ctx.ParentChannel, err = ctx.State.Channel(ctx.Channel.ParentID)
		if err != nil {
			return ctx, err
		}
	}

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
			Flags: api.EphemeralResponse,
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
	return ctx.reply(api.EphemeralResponse, content, embeds)
}

func (ctx *Context) reply(flags api.InteractionResponseFlags, content string, embeds []discord.Embed) error {
	if ctx.deferred {
		_, err := ctx.State.EditInteractionResponse(ctx.Event.AppID, ctx.InteractionToken, api.EditInteractionResponseData{
			Content: option.NewNullableString(content),
			Embeds:  &embeds,
		})
		return err
	}

	return ctx.State.RespondInteraction(ctx.InteractionID, ctx.InteractionToken, api.InteractionResponse{
		Type: api.MessageInteractionWithSource,
		Data: &api.InteractionResponseData{
			Content: option.NewNullableString(content),
			Embeds:  &embeds,
			Flags:   flags,
		},
	})
}

// Ctx implements HasContext
func (ctx *Context) Ctx() *Context {
	return ctx
}

// Original returns the original response to an interaction, if any.
func (ctx *Context) Original() (msg *discord.Message, err error) {
	url := api.EndpointWebhooks + ctx.State.Ready().User.ID.String() + "/" + ctx.InteractionToken + "/messages/@original"

	return msg, ctx.State.RequestJSON(&msg, "GET", url)
}
