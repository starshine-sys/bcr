package bcr

import (
	"errors"
	"strings"

	"github.com/diamondburned/arikawa/v2/bot/extras/shellwords"
	"github.com/diamondburned/arikawa/v2/discord"
	"github.com/diamondburned/arikawa/v2/gateway"
	"github.com/diamondburned/arikawa/v2/state"
)

// Errors related to getting the context
var (
	ErrChannel   = errors.New("context: couldn't get channel")
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

// Context is a command context
type Context struct {
	// Command and Prefix contain the invoked command's name and prefix, respectively.
	// Note that Command won't be accurate if the invoked command was a subcommand, use FullCommandPath for that.
	Command string
	Prefix  string

	FullCommandPath []string

	Args    []string
	RawArgs string

	internalArgs []string
	pos          int

	State *state.State
	Bot   *discord.User

	// Info about the message
	Message discord.Message
	Channel *discord.Channel
	Author  discord.User

	// Note: Member is nil for non-guild messages
	Member *discord.Member

	// The command and the router used
	Cmd    *Command
	Router *Router

	AdditionalParams map[string]interface{}
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

		internalArgs:     args,
		Args:             args,
		Message:          m.Message,
		Author:           m.Author,
		Member:           m.Member,
		RawArgs:          raw,
		Router:           r,
		State:            r.State,
		Bot:              r.Bot,
		AdditionalParams: make(map[string]interface{}),
	}

	// get the channel
	channel, err := r.State.Channel(m.ChannelID)
	if err != nil {
		return ctx, ErrChannel
	}
	ctx.Channel = channel

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
