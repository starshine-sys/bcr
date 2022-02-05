package bcr

import (
	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/diamondburned/arikawa/v3/utils/json/option"
)

type CommandContext struct {
	Command []string
	Options []discord.CommandInteractionOption

	InteractionID    discord.InteractionID
	InteractionToken string

	State *state.State

	User          discord.User
	Member        *discord.Member
	Guild         *discord.Guild
	Channel       *discord.Channel
	ParentChannel *discord.Channel

	Event *gateway.InteractionCreateEvent
	Data  *discord.CommandInteraction

	deferred bool
}

// NewCommandContext creates a new command context.
func (r *Router) NewCommandContext(ic *gateway.InteractionCreateEvent) (ctx *CommandContext, err error) {
	data, ok := ic.Data.(*discord.CommandInteraction)
	if !ok {
		return nil, ErrNotCommand
	}

	sc := &CommandContext{
		Event:            ic,
		Data:             data,
		Command:          []string{data.Name},
		Options:          data.Options,
		InteractionID:    ic.ID,
		InteractionToken: ic.Token,
	}

	if ic.Member != nil {
		sc.Member = ic.Member
		sc.User = ic.Member.User
	} else {
		sc.User = *ic.User
	}

	state, _ := r.ShardFromGuildID(ic.GuildID)
	sc.State = state

	// get guild
	if ic.GuildID.IsValid() {
		sc.Guild, err = sc.State.Guild(ic.GuildID)
		if err != nil {
			return sc, err
		}
		sc.Guild.Roles, err = sc.State.Roles(ic.GuildID)
		if err != nil {
			return sc, err
		}
	}

	// get the channel
	sc.Channel, err = sc.State.Channel(ic.ChannelID)
	if err != nil {
		return sc, err
	}

	if IsThread(sc.Channel) {
		sc.ParentChannel, err = sc.State.Channel(sc.Channel.ParentID)
		if err != nil {
			return sc, err
		}
	}

	return sc, nil
}

func (ctx *CommandContext) Defer() error {
	err := ctx.State.RespondInteraction(ctx.InteractionID, ctx.InteractionToken, api.InteractionResponse{
		Type: api.DeferredMessageInteractionWithSource,
	})
	if err != nil {
		return err
	}

	ctx.deferred = true
	return nil
}

func (ctx *CommandContext) DeferEphemeral() error {
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

func (ctx *CommandContext) Reply(content string, embeds ...discord.Embed) error {
	return ctx.reply(0, content, embeds)
}

func (ctx *CommandContext) ReplyEphemeral(content string, embeds ...discord.Embed) error {
	return ctx.reply(api.EphemeralResponse, content, embeds)
}

func (ctx *CommandContext) reply(flags api.InteractionResponseFlags, content string, embeds []discord.Embed) error {
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
