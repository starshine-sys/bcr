package bcr

import (
	"fmt"
	"time"

	"emperror.dev/errors"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/api/webhook"
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
	Flags

	// SendX sends a message without returning the created discord.Message
	SendX(string, ...discord.Embed) error
	SendfX(string, ...interface{}) error

	// SendFiles sends a message with attachments
	SendFiles(string, ...sendpart.File) error

	// Session returns this context's *state.State
	Session() *state.State
	// User returns this context's Author
	User() discord.User
	// GetGuild returns this context's Guild
	GetGuild() *discord.Guild
	// GetChannel returns this context's Channel
	GetChannel() *discord.Channel
	// GetParentChannel returns this context's ParentChannel
	GetParentChannel() *discord.Channel

	// ButtonPages paginates a slice of embeds using buttons
	ButtonPages(embeds []discord.Embed, timeout time.Duration) (msg *discord.Message, rmFunc func(), err error)
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

	Author discord.User
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
		sc.Author = ic.Member.User
	} else {
		sc.Author = *ic.User
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

// Original returns the original response to an interaction, if any.
func (ctx *SlashContext) Original() (msg *discord.Message, err error) {
	url := api.EndpointWebhooks + ctx.Router.Bot.ID.String() + "/" + ctx.InteractionToken + "/messages/@original"

	return msg, ctx.State.RequestJSON(&msg, "GET", url)
}

// EditOriginal edits the original response.
// This is yoinked from arikawa/v3/api/webhook because that doesn't accept @original as a message ID.
func (ctx *SlashContext) EditOriginal(data webhook.EditMessageData) (*discord.Message, error) {
	if data.AllowedMentions != nil {
		if err := data.AllowedMentions.Verify(); err != nil {
			return nil, errors.Wrap(err, "allowedMentions error")
		}
	}
	if data.Embeds != nil {
		sum := 0
		for _, e := range *data.Embeds {
			if err := e.Validate(); err != nil {
				return nil, errors.Wrap(err, "embed error")
			}
			sum += e.Length()
			if sum > 6000 {
				return nil, &discord.OverboundError{Count: sum, Max: 6000, Thing: "sum of text in embeds"}
			}
		}
	}
	var msg *discord.Message
	return msg, sendpart.PATCH(ctx.State.Client.Client, data, &msg,
		api.EndpointWebhooks+ctx.Router.Bot.ID.String()+"/"+ctx.InteractionToken+"/messages/@original")

}

// GetGuild ...
func (ctx *SlashContext) GetGuild() *discord.Guild { return ctx.Guild }

// GetChannel ...
func (ctx *SlashContext) GetChannel() *discord.Channel { return ctx.Channel }

// GetParentChannel ...
func (ctx *SlashContext) GetParentChannel() *discord.Channel { return ctx.ParentChannel }

// User ...
func (ctx *SlashContext) User() discord.User { return ctx.Author }
