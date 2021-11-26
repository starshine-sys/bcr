package bcr

import (
	"fmt"
	"time"

	"emperror.dev/errors"

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
	Flags

	// SendX sends a message without returning the created discord.Message
	SendX(string, ...discord.Embed) error
	SendfX(string, ...interface{}) error

	// Send sends a message, returning the created message.
	Send(string, ...discord.Embed) (*discord.Message, error)
	Sendf(string, ...interface{}) (*discord.Message, error)

	SendComponents(discord.ContainerComponents, string, ...discord.Embed) (*discord.Message, error)

	// SendFiles sends a message with attachments
	SendFiles(string, ...sendpart.File) error

	// SendEphemeral sends an ephemeral message (or falls back to a normal message without slash commands)
	SendEphemeral(string, ...discord.Embed) error

	EditOriginal(api.EditInteractionResponseData) (*discord.Message, error)

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
	// GetMember returns this context's Member
	GetMember() *discord.Member

	// ButtonPages paginates a slice of embeds using buttons
	ButtonPages(embeds []discord.Embed, timeout time.Duration) (msg *discord.Message, rmFunc func(), err error)
	ButtonPagesWithComponents(embeds []discord.Embed, timeout time.Duration, components discord.ContainerComponents) (msg *discord.Message, rmFunc func(), err error)

	// ConfirmButton confirms a prompt with buttons or "yes"/"no" messages.
	ConfirmButton(userID discord.UserID, data ConfirmData) (yes, timeout bool)
}

var _ Contexter = (*SlashContext)(nil)

// SlashContext is the Contexter passed to a slash command function.
type SlashContext struct {
	CommandID      discord.CommandID
	CommandName    string
	CommandOptions []discord.CommandInteractionOption

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
	Data  *discord.CommandInteraction

	AdditionalParams map[string]interface{}
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

	if ic.Data.InteractionType() != discord.CommandInteractionType {
		return nil, ErrNotCommand
	}

	data, ok := ic.Data.(*discord.CommandInteraction)
	if !ok {
		return nil, ErrNotCommand
	}

	sc := &SlashContext{
		Router:           r,
		Event:            ic,
		Data:             data,
		CommandName:      data.Name,
		CommandID:        data.ID,
		CommandOptions:   data.Options,
		InteractionID:    ic.ID,
		InteractionToken: ic.Token,
		AdditionalParams: map[string]interface{}{},
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
		sc.ParentChannel, err = sc.State.Channel(sc.Channel.ParentID)
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
func (ctx *SlashContext) EditOriginal(data api.EditInteractionResponseData) (*discord.Message, error) {
	return ctx.State.EditInteractionResponse(discord.AppID(ctx.Router.Bot.ID), ctx.Event.Token, data)
}

// GetGuild ...
func (ctx *SlashContext) GetGuild() *discord.Guild { return ctx.Guild }

// GetChannel ...
func (ctx *SlashContext) GetChannel() *discord.Channel { return ctx.Channel }

// GetParentChannel ...
func (ctx *SlashContext) GetParentChannel() *discord.Channel { return ctx.ParentChannel }

// User ...
func (ctx *SlashContext) User() discord.User { return ctx.Author }

// GetMember ...
func (ctx *SlashContext) GetMember() *discord.Member { return ctx.Member }

// Send ...
func (ctx *SlashContext) Send(content string, embeds ...discord.Embed) (msg *discord.Message, err error) {
	err = ctx.SendX(content, embeds...)
	if err != nil {
		return
	}

	return ctx.Original()
}

// Sendf ...
func (ctx *SlashContext) Sendf(tmpl string, args ...interface{}) (msg *discord.Message, err error) {
	err = ctx.SendfX(tmpl, args...)
	if err != nil {
		return
	}

	return ctx.Original()
}

// GuildPerms returns the global (guild) permissions of this Context's user.
// If in DMs, it will return the permissions users have in DMs.
func (ctx *SlashContext) GuildPerms() (perms discord.Permissions) {
	if ctx.Guild == nil || ctx.Member == nil {
		return discord.PermissionViewChannel | discord.PermissionSendMessages | discord.PermissionAddReactions | discord.PermissionReadMessageHistory
	}

	if ctx.Guild.OwnerID == ctx.Author.ID {
		return discord.PermissionAll
	}

	for _, id := range ctx.Member.RoleIDs {
		for _, r := range ctx.Guild.Roles {
			if id == r.ID {
				if r.Permissions.Has(discord.PermissionAdministrator) {
					return discord.PermissionAll
				}

				perms |= r.Permissions
				break
			}
		}
	}

	return perms
}

// SendComponents sends a message with components
func (ctx *SlashContext) SendComponents(components discord.ContainerComponents, content string, embeds ...discord.Embed) (*discord.Message, error) {
	data := api.InteractionResponse{
		Type: api.MessageInteractionWithSource,
		Data: &api.InteractionResponseData{
			AllowedMentions: ctx.Router.DefaultMentions,
			Content:         option.NewNullableString(content),
			Embeds:          &embeds,
			Components:      &components,
		},
	}

	err := ctx.State.RespondInteraction(ctx.InteractionID, ctx.InteractionToken, data)
	if err != nil {
		return nil, err
	}

	return ctx.Original()
}
