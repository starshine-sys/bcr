package bcr

import (
	"errors"
	"strings"

	"github.com/diamondburned/arikawa/v2/bot/extras/shellwords"
	"github.com/diamondburned/arikawa/v2/discord"
	"github.com/diamondburned/arikawa/v2/state"
)

// Errors related to getting the context
var (
	ErrChannel   = errors.New("context: couldn't get channel")
	ErrNoBotUser = errors.New("context: couldn't get bot user")
)

// Context is a command context
type Context struct {
	Command         string
	fullCommandPath []string

	Args    []string
	RawArgs string

	Session *state.State
	Bot     *discord.User

	Message discord.Message
	Channel *discord.Channel
	Author  discord.User

	Cmd    *Command
	Router *Router

	AdditionalParams map[string]interface{}
}

// NewContext returns a new message context
func (r *Router) NewContext(m discord.Message) (ctx *Context, err error) {
	messageContent := m.Content
	for _, prefix := range r.Prefixes {
		if strings.HasPrefix(strings.ToLower(messageContent), prefix) {
			messageContent = messageContent[len(prefix):]
		}
	}
	messageContent = strings.TrimSpace(messageContent)

	message, err := shellwords.Parse(messageContent)
	if err != nil {
		message = strings.Split(messageContent, " ")
	}
	command := strings.ToLower(message[0])
	args := []string{}
	if len(message) > 1 {
		args = message[1:]
	}

	raw := TrimPrefixesSpace(messageContent, message[0])

	// create the context
	ctx = &Context{
		Command:          command,
		Args:             args,
		Message:          m,
		Author:           m.Author,
		RawArgs:          raw,
		Router:           r,
		Session:          r.Session,
		Bot:              r.Bot,
		AdditionalParams: make(map[string]interface{}),
	}

	// get the channel
	channel, err := r.Session.Channel(m.ChannelID)
	if err != nil {
		return ctx, ErrChannel
	}
	ctx.Channel = channel

	return ctx, err
}
