package bcr

import (
	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/diamondburned/arikawa/v3/utils/json/option"
)

// CommandContext is the context for an application command.
type CommandContext struct {
	State *state.State
	Rest  *api.Client

	Event *gateway.InteractionCreateEvent
	Data  *discord.CommandInteractionData

	User   discord.User
	Member *discord.Member

	Guild   *discord.Guild
	Channel *discord.Channel

	// internal fields
	deferred bool
}

// NewCommandContext creates a new command context.
func (r *Router) NewCommandContext(s *state.State, ev *gateway.InteractionCreateEvent) (ctx *CommandContext, err error) {
	ctx = &CommandContext{
		State: s,
		Rest:  r.Rest,

		Event: ev,
		Data:  ev.Data.(*discord.CommandInteractionData),

		Member: ev.Member,
	}

	u := ev.User
	if u == nil {
		u = &ev.Member.User
	}
	ctx.User = *u

	ctx.Guild, ctx.Channel, err = r.CollectContext(s, ev.GuildID, ev.ChannelID)
	if err != nil {
		return nil, err
	}

	return ctx, nil
}

// Defer defers a command response. The caller will have 15 minutes to respond using EditOriginal.
func (ctx *CommandContext) Defer(ephemeral bool) error {
	var dat *api.InteractionResponseData
	if ephemeral {
		dat = &api.InteractionResponseData{
			Flags: api.EphemeralResponse,
		}
	}

	ctx.deferred = true

	return ctx.Rest.RespondInteraction(ctx.Event.ID, ctx.Event.Token, api.InteractionResponse{
		Type: api.DeferredMessageInteractionWithSource,
		Data: dat,
	})
}

// Respond responds to the interaction.
func (ctx *CommandContext) Respond(content string, embeds ...discord.Embed) error {
	return ctx.respond(false, content, embeds)
}

// RespondEphemeral responds to the interaction with an ephemeral message.
func (ctx *CommandContext) RespondEphemeral(content string, embeds ...discord.Embed) error {
	return ctx.respond(true, content, embeds)
}

func (ctx *CommandContext) respond(ephemeral bool, content string, embeds []discord.Embed) error {
	if ctx.deferred {
		_, err := ctx.EditOriginal(api.EditInteractionResponseData{
			Content: option.NewNullableString(content),
			Embeds:  &embeds,
		})
		return err
	}

	flags := api.EphemeralResponse
	if !ephemeral {
		flags = 0
	}

	return ctx.Rest.RespondInteraction(ctx.Event.ID, ctx.Event.Token, api.InteractionResponse{
		Type: api.MessageInteractionWithSource,
		Data: &api.InteractionResponseData{
			Content: option.NewNullableString(content),
			Embeds:  &embeds,
			Flags:   flags,
		},
	})
}

// EditOriginal edits the original response.
func (ctx *CommandContext) EditOriginal(dat api.EditInteractionResponseData) (*discord.Message, error) {
	return ctx.Rest.EditInteractionResponse(ctx.Event.AppID, ctx.Event.Token, dat)
}
