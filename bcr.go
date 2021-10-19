// Package bcr provides a command router for Arikawa v3. It does *not* support text commands.
package bcr

import (
	"log"
	"strings"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/gateway/shard"
	"github.com/diamondburned/arikawa/v3/state"
)

var (
	// Debug is used for debug log calls
	Debug = log.Printf
	// Error is used for error log calls
	Error = log.Printf
)

// Router is the main router struct. To have your bot listen to commands, make sure to add the InteractionCreate method on this struct as a handler.
type Router struct {
	// Global middlewares used for all commands.
	CommandMiddlewares []CommandMiddleware
	// Commands are all top-level application commands.
	Commands []*Command

	// ShardManager is the shard manager containing all of this bot's shards, used in InteractionCreate to populate the context.
	ShardManager *shard.Manager
	// State is the bot's state, used in InteractionCreate to populate the context.
	// If both State and ShardManager are non-nil, ShardManager is used instead.
	State *state.State
	// Rest is the global API client. All Contexts, regardless of their shard, use this client.
	Rest *api.Client

	// OnPanic is called when an interaction create event panics.
	OnPanic func(v interface{})
	// OnCommandError is called when a command errors.
	OnCommandError func(ctx *CommandContext, path []string, err error)
	// OnComponentError is called when a component errors.
	OnComponentError func(ctx string, err error)

	// CollectContext is used to collect data for command/component contexts.
	CollectContext func(s *state.State, guildID discord.GuildID, channelID discord.ChannelID) (*discord.Guild, *discord.Channel, error)
}

// New creates a new Router.
func New() *Router {
	r := &Router{
		OnCommandError: func(ctx *CommandContext, path []string, err error) {
			Error("Error in command %v: %v", strings.Join(path, " "), err)
		},

		CollectContext: DefaultCollectContext,
	}

	return r
}

// InteractionCreate is the main handler used by Router, which delegates off to command or component interactions.
func (r *Router) InteractionCreate(ev *gateway.InteractionCreateEvent) {
	if r.OnPanic != nil {
		defer func() {
			rev := recover()
			if rev != nil {
				r.OnPanic(rev)
			}
		}()
	}

	s := r.State
	if r.ShardManager != nil {
		sh, _ := r.ShardManager.FromGuildID(ev.GuildID)
		s = sh.(shard.ShardState).Shard.(*state.State)
	}

	switch ev.Type {
	case discord.CommandInteraction:
		ctx, err := r.NewCommandContext(s, ev)
		if err != nil {
			Error("Error creating context: %v", err)
		}

		err = r.ExecuteSlashCommand(ctx)
		if err != nil {
			e := err.(*commandError)
			r.OnCommandError(ctx, e.path, e.err)
		}
	}
}

// DefaultCollectContext is the default value for r.CollectContext
func DefaultCollectContext(s *state.State, guildID discord.GuildID, channelID discord.ChannelID) (*discord.Guild, *discord.Channel, error) {
	g, err := s.Guild(guildID)
	if err == nil {
		g.Roles, err = s.Roles(guildID)
		if err != nil {
			return nil, nil, err
		}
	}

	ch, err := s.Channel(channelID)
	if err != nil {
		return nil, nil, err
	}
	return g, ch, err
}
