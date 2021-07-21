package bcr

import (
	"errors"
	"fmt"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/diamondburned/arikawa/v3/utils/json/option"
	"github.com/diamondburned/arikawa/v3/utils/sendpart"
)

// Contexter is the type passed to (*Command).SlashCommand.
// This includes basic methods implemented by both Context and SlashContext; for all of their respective methods and fields, convert the Contexter to a Context or SlashContext.
// As SlashCommand will be called with a basic *Context if Command is nil, this allows for some code deduplication if the command only uses basic methods.
type Contexter interface {
	// SendX sends a message without returning the created discord.Message
	SendX(string, ...discord.Embed) error
	SendfX(string, ...interface{}) error

	// SendFiles sends a message with attachments
	SendFiles(string, ...sendpart.File) error

	// Session returns this context's *state.State
	Session() *state.State
}

var _ Contexter = (*SlashContext)(nil)

// SlashContext is the Contexter passed to a slash command function.
type SlashContext struct {
	CommandID      discord.CommandID
	CommandName    string
	CommandOptions []gateway.InteractionOption

	InteractionID    discord.InteractionID
	InteractionToken string

	Command *Command
	Router  *Router
	State   *state.State

	// Event is the original raw event
	Event *gateway.InteractionCreateEvent
}

// Session returns this SlashContext's state.
func (ctx *SlashContext) Session() *state.State {
	return ctx.State
}

// Errors related to slash contexts
var (
	ErrNotCommand = errors.New("not a command interaction")
)

// NewSlashContext creates a new slash command context.
func (r *Router) NewSlashContext(ic *gateway.InteractionCreateEvent) (*SlashContext, error) {
	if ic.Type != gateway.CommandInteraction {
		return nil, ErrNotCommand
	}

	sc := &SlashContext{
		Router:           r,
		Event:            ic,
		CommandName:      ic.Data.Name,
		CommandID:        ic.Data.ID,
		CommandOptions:   ic.Data.Options,
		InteractionID:    ic.ID,
		InteractionToken: ic.Token,
	}

	state, _ := r.StateFromGuildID(ic.GuildID)
	sc.State = state

	return sc, nil
}

// SendX sends a message without returning the created discord.Message
func (ctx *SlashContext) SendX(content string, embeds ...discord.Embed) (err error) {
	data := api.InteractionResponse{
		Type: api.MessageInteractionWithSource,
		Data: &api.InteractionResponseData{
			AllowedMentions: ctx.Router.DefaultMentions,
		},
	}

	if len(embeds) != 0 {
		data.Data.Embeds = &embeds
	}
	if content != "" {
		data.Data.Content = option.NewNullableString(content)
	}

	err = ctx.State.RespondInteraction(ctx.InteractionID, ctx.InteractionToken, data)
	return
}

// SendfX ...
func (ctx *SlashContext) SendfX(format string, args ...interface{}) (err error) {
	return ctx.SendX(fmt.Sprintf(format, args...))
}

// SendFiles sends a message with attachments
func (ctx *SlashContext) SendFiles(content string, files ...sendpart.File) (err error) {
	data := api.InteractionResponse{
		Type: api.MessageInteractionWithSource,
		Data: &api.InteractionResponseData{
			AllowedMentions: ctx.Router.DefaultMentions,
		},
	}

	if len(files) != 0 {
		data.Data.Files = files
	}
	if content != "" {
		data.Data.Content = option.NewNullableString(content)
	}

	err = ctx.State.RespondInteraction(ctx.InteractionID, ctx.InteractionToken, data)
	return
}

// SendEphemeral sends an ephemeral message.
func (ctx *SlashContext) SendEphemeral(content string, embeds ...discord.Embed) (err error) {
	data := api.InteractionResponse{
		Type: api.MessageInteractionWithSource,
		Data: &api.InteractionResponseData{
			AllowedMentions: ctx.Router.DefaultMentions,
			Flags:           api.EphemeralResponse,
		},
	}

	if len(embeds) != 0 {
		data.Data.Embeds = &embeds
	}
	if content != "" {
		data.Data.Content = option.NewNullableString(content)
	}

	err = ctx.State.RespondInteraction(ctx.InteractionID, ctx.InteractionToken, data)
	return
}
