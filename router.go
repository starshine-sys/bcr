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

	commands map[string]*handler[*CommandContext]
	modals   map[discord.ComponentID]*handler[*ModalContext]
}

func NewFromState(s *state.State) *Router {
	c := api.NewClient(s.Token)
	s.Client = c

	r := &Router{
		Rest:     c,
		State:    s,
		commands: make(map[string]*handler[*CommandContext]),
		modals:   make(map[discord.ComponentID]*handler[*ModalContext]),
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

func (r *Router) Command(path string) *commandBuilder {
	return &commandBuilder{
		r:    r,
		path: path,
	}
}

func (r *Router) Modal(id discord.ComponentID) *modalBuilder {
	return &modalBuilder{
		r:  r,
		id: id,
	}
}
