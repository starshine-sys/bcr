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
	return "0.9.1"
}

// RequiredIntents are the intents required for the command handler
const RequiredIntents = gateway.IntentGuildMessages | gateway.IntentGuildMessageReactions | gateway.IntentDirectMessages | gateway.IntentDirectMessageReactions | gateway.IntentGuilds

// Router is the command router
type Router struct {
	BotOwners []string

	Prefixes []string
	Prefixer Prefixer

	Session *state.State
	Bot     *discord.User

	BlacklistFunc   func(*Context) bool
	HelpCommand     func(*Context) error
	DefaultMentions *api.AllowedMentions
	EmbedColor      discord.Color

	cooldowns *CooldownCache
	cmds      map[string]*Command
	cmdMu     sync.RWMutex

	// maps + mutexes
	reactions  map[reactionKey]reactionInfo
	reactionMu sync.RWMutex
	messages   map[messageKey]messageInfo
	messageMu  sync.RWMutex
}

// New creates a new router object
func New(s *state.State, owners, prefixes []string) *Router {
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

	// set prefixer
	r.Prefixer = r.DefaultPrefixer

	// add required handlers
	r.Session.AddHandler(r.ReactionAdd)
	r.Session.AddHandler(r.ReactionMessageDelete)
	r.Session.AddHandler(r.MsgHandlerCreate)

	return r
}

// NewWithState creates a new router with a state.
// The token is automatically prefixed with `Bot `.
func NewWithState(token string, owners []discord.UserID, prefixes []string) (*Router, error) {
	return NewWithIntents(token, owners, prefixes, RequiredIntents)
}

// NewWithIntents creates a new router with a state, with the specified intents.
// The token is automatically prefixed with `Bot `.
func NewWithIntents(token string, owners []discord.UserID, prefixes []string, intents gateway.Intents) (*Router, error) {
	ownerStrings := make([]string, 0)
	for _, o := range owners {
		ownerStrings = append(ownerStrings, o.String())
	}
	s, err := state.NewWithIntents("Bot "+token, intents)
	if err != nil {
		return nil, err
	}

	r := New(s, ownerStrings, prefixes)
	return r, nil
}

// AddCommand adds a command to the router
func (r *Router) AddCommand(c *Command) *Command {
	c.id = sGen.Get()
	r.cmdMu.Lock()
	defer r.cmdMu.Unlock()
	r.cmds[strings.ToLower(c.Name)] = c

	for _, a := range c.Aliases {
		r.cmds[strings.ToLower(a)] = c
	}

	return c
}
