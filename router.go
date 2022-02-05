package bcr

import (
	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/session/shard"
	"github.com/diamondburned/arikawa/v3/state"
)

type Router struct {
	Rest         *api.Client
	State        *state.State
	ShardManager *shard.Manager

	commands map[string]*command
}

func NewFromState(s *state.State) *Router {
	c := api.NewClient(s.Token)
	s.Client = c

	r := &Router{
		Rest:     c,
		State:    s,
		commands: map[string]*command{},
	}

	return r
}

func (r *Router) ShardFromGuildID(guildID discord.GuildID) (*state.State, int) {
	if r.ShardManager == nil {
		return r.State, 0
	}

	s, id := r.ShardManager.FromGuildID(guildID)
	return s.(*state.State), id
}
