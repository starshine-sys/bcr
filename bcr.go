package bcr

import (
	"strings"
	"sync"

	"github.com/diamondburned/arikawa/v2/api"
	"github.com/diamondburned/arikawa/v2/discord"
	"github.com/diamondburned/arikawa/v2/gateway"
	"github.com/diamondburned/arikawa/v2/state"
)

// Version returns the current brc version
func Version() string {
	return "0.2.1"
}

// RequiredIntents are the intents required for the command handler
const RequiredIntents = gateway.IntentGuildMessages | gateway.IntentGuildMessageReactions | gateway.IntentDirectMessages | gateway.IntentDirectMessageReactions | gateway.IntentGuilds

// Router is the command router
type Router struct {
	BotOwners []string
	Prefixes  []string

	Session *state.State
	Bot     *discord.User

	BlacklistFunc   func(*Context) bool
	HelpCommand     func(*Context) error
	DefaultMentions *api.AllowedMentions
	EmbedColor      discord.Color

	cooldowns *CooldownCache
	cmds      map[string]*Command
	cmdMu     sync.RWMutex
	reactions map[reactionKey]reactionInfo
	messages  map[messageKey]messageInfo
}

// NewRouter creates a new router object
func NewRouter(s *state.State, owners, prefixes []string) *Router {
	r := &Router{
		Session:    s,
		BotOwners:  owners,
		Prefixes:   prefixes,
		EmbedColor: discord.DefaultEmbedColor,

		DefaultMentions: &api.AllowedMentions{
			Parse: []api.AllowedMentionType{api.AllowUserMention},
		},

		cmds:      make(map[string]*Command),
		reactions: make(map[reactionKey]reactionInfo),
		messages:  make(map[messageKey]messageInfo),
		cooldowns: newCooldownCache(),
	}

	// add required handlers
	r.Session.AddHandler(r.ReactionAdd)
	r.Session.AddHandler(r.ReactionMessageDelete)
	r.Session.AddHandler(r.MsgHandlerCreate)

	// set help command
	r.HelpCommand = r.DefaultHelpCommand
	return r
}

// AddCommand adds a command to the router
func (r *Router) AddCommand(c *Command) *Command {
	r.cmdMu.Lock()
	defer r.cmdMu.Unlock()
	r.cmds[strings.ToLower(c.Name)] = c

	for _, a := range c.Aliases {
		r.cmds[strings.ToLower(a)] = c
	}

	return c
}
