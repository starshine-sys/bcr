package bcr

import (
	"errors"
	"strings"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/diamondburned/arikawa/v3/utils/bot/extras/shellwords"
	"github.com/spf13/pflag"
)

// Errors related to getting the context
var (
	ErrChannel   = errors.New("context: couldn't get channel")
	ErrGuild     = errors.New("context: couldn't get guild")
	ErrNoBotUser = errors.New("context: couldn't get bot user")

	ErrEmptyMessage = errors.New("context: message was empty")
)

// Prefixer returns the prefix used and the length. If the message doesn't start with a valid prefix, it returns -1.
// Note that this function should still use the built-in r.Prefixes for mention prefixes
type Prefixer func(m discord.Message) int

// DefaultPrefixer ...
func (r *Router) DefaultPrefixer(m discord.Message) int {
	for _, p := range r.Prefixes {
		if strings.HasPrefix(strings.ToLower(m.Content), p) {
			return len(p)
		}
	}
	return -1
}

var _ Contexter = (*Context)(nil)

// Context is a command context
type Context struct {
	// Command and Prefix contain the invoked command's name and prefix, respectively.
	// Note that Command won't be accurate if the invoked command was a subcommand, use FullCommandPath for that.
	Command string
	Prefix  string

	FullCommandPath []string

	Args    []string
	RawArgs string

	Flags *pflag.FlagSet

	InternalArgs []string
	pos          int

	State   *state.State
	ShardID int

	Bot *discord.User

	// Info about the message
	Message discord.Message
	Channel *discord.Channel
	Guild   *discord.Guild
	Author  discord.User

	// ParentChannel is only filled if ctx.Channel is a thread
	ParentChannel *discord.Channel

	// Note: Member is nil for non-guild messages
	Member *discord.Member

	// The command and the router used
	Cmd    *Command
	Router *Router

	AdditionalParams map[string]interface{}

	// Internal use for the Get* methods.
	// Not intended to be changed by the end user, exported so it can be created if context is not made through NewContext.
	FlagMap map[string]interface{}

	origMessage *discord.Message
}

// NewContext returns a new message context
func (r *Router) NewContext(m *gateway.MessageCreateEvent) (ctx *Context, err error) {
	messageContent := m.Content

	var p int
	if p = r.Prefixer(m.Message); p != -1 {
		messageContent = messageContent[p:]
	} else {
		return nil, ErrEmptyMessage
	}
	messageContent = strings.TrimSpace(messageContent)

	message, err := shellwords.Parse(messageContent)
	if err != nil {
		message = strings.Split(messageContent, " ")
	}
	if len(message) == 0 {
		return nil, ErrEmptyMessage
	}
	command := strings.ToLower(message[0])
	args := []string{}
	if len(message) > 1 {
		args = message[1:]
	}

	raw := TrimPrefixesSpace(messageContent, message[0])

	// create the context
	ctx = &Context{
		Command: command,
		Prefix:  m.Content[:p],

		InternalArgs:     args,
		Args:             args,
		Message:          m.Message,
		Author:           m.Author,
		Member:           m.Member,
		RawArgs:          raw,
		Router:           r,
		Bot:              r.Bot,
		AdditionalParams: make(map[string]interface{}),
		FlagMap:          make(map[string]interface{}),
	}

	ctx.State, ctx.ShardID = r.StateFromGuildID(m.GuildID)

	// get the channel
	ctx.Channel, err = ctx.State.Channel(m.ChannelID)
	if err != nil {
		return ctx, ErrChannel
	}

	if ctx.Thread() {
		ctx.ParentChannel, err = ctx.State.Channel(ctx.Channel.ParentID)
		if err != nil {
			return ctx, ErrChannel
		}
	}

	// get guild
	if m.GuildID.IsValid() {
		ctx.Guild, err = ctx.State.Guild(m.GuildID)
		if err != nil {
			return ctx, ErrGuild
		}
		ctx.Guild.Roles, err = ctx.State.Roles(m.GuildID)
		if err != nil {
			return ctx, ErrGuild
		}
	}

	return ctx, err
}

// DisplayName returns the context user's displayed name (either username without discriminator, or nickname)
func (ctx *Context) DisplayName() string {
	if ctx.Member == nil {
		return ctx.Author.Username
	}
	if ctx.Member.Nick == "" {
		return ctx.Author.Username
	}
	return ctx.Member.Nick
}

// Thread returns true if the context is in a thread channel.
// If this function returns true, ctx.ParentChannel will be non-nil.
func (ctx *Context) Thread() bool {
	return ctx.Channel.Type == discord.GuildNewsThread || ctx.Channel.Type == discord.GuildPublicThread || ctx.Channel.Type == discord.GuildPrivateThread
}

// Session returns this context's state.
func (ctx *Context) Session() *state.State {
	return ctx.State
}

// GetGuild ...
func (ctx *Context) GetGuild() *discord.Guild { return ctx.Guild }

// GetChannel ...
func (ctx *Context) GetChannel() *discord.Channel { return ctx.Channel }

// GetParentChannel ...
func (ctx *Context) GetParentChannel() *discord.Channel { return ctx.ParentChannel }

// User ...
func (ctx *Context) User() discord.User { return ctx.Author }

// GetMember ...
func (ctx *Context) GetMember() *discord.Member { return ctx.Member }

// EditOriginal edits the original response message.
func (ctx *Context) EditOriginal(data api.EditInteractionResponseData) (*discord.Message, error) {
	if ctx.origMessage == nil {
		return nil, errors.New("no original message to edit")
	}

	emd := api.EditMessageData{
		Content:         data.Content,
		Embeds:          data.Embeds,
		Components:      data.Components,
		AllowedMentions: data.AllowedMentions,
		Attachments:     data.Attachments,
	}

	return ctx.State.EditMessageComplex(ctx.origMessage.ChannelID, ctx.origMessage.ID, emd)
}
