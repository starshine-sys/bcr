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
	Guild   *discord.Guild

	User   discord.User
	Member *discord.Member

	Channel       *discord.Channel
	ParentChannel *discord.Channel

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
	var err error

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

	if ic.Member != nil {
		sc.Member = ic.Member
		sc.User = ic.Member.User
	} else {
		sc.User = *ic.User
	}

	state, _ := r.StateFromGuildID(ic.GuildID)
	sc.State = state

	// get guild
	if ic.GuildID.IsValid() {
		sc.Guild, err = sc.State.Guild(ic.GuildID)
		if err != nil {
			return sc, ErrGuild
		}
		sc.Guild.Roles, err = sc.State.Roles(ic.GuildID)
		if err != nil {
			return sc, ErrGuild
		}
	}

	// get the channel
	sc.Channel, err = sc.State.Channel(ic.ChannelID)
	if err != nil {
		return sc, ErrChannel
	}

	if sc.Thread() {
		sc.ParentChannel, err = sc.State.Channel(sc.Channel.CategoryID)
		if err != nil {
			return sc, ErrChannel
		}
	}

	return sc, nil
}

// Thread returns true if the context is in a thread channel.
// If this function returns true, ctx.ParentChannel will be non-nil.
func (ctx *SlashContext) Thread() bool {
	return ctx.Channel.Type == discord.GuildNewsThread || ctx.Channel.Type == discord.GuildPublicThread || ctx.Channel.Type == discord.GuildPrivateThread
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
