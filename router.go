package bcr

import (
	"sync"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/session/shard"
	"github.com/diamondburned/arikawa/v3/state"
)

type Router struct {
	Rest         *api.Client
	State        *state.State
	ShardManager *shard.Manager

	commands      map[string]*handler[*CommandContext]
	autocompletes map[string]*handler[*AutocompleteContext]
	modals        map[discord.ComponentID]*handler[*ModalContext]

	componentsMu sync.RWMutex
	buttons      map[componentKey]*handler[*ButtonContext]
	selects      map[componentKey]*handler[*SelectContext]
}

type componentKey struct {
	id    discord.ComponentID
	msgID discord.MessageID
}

func NewFromState(s *state.State) *Router {
	c := api.NewClient(s.Token)
	s.Client = c

	r := &Router{
		Rest:          c,
		State:         s,
		commands:      make(map[string]*handler[*CommandContext]),
		autocompletes: make(map[string]*handler[*AutocompleteContext]),
		modals:        make(map[discord.ComponentID]*handler[*ModalContext]),
		buttons:       make(map[componentKey]*handler[*ButtonContext]),
		selects:       make(map[componentKey]*handler[*SelectContext]),
	}

	return r
}

func NewFromShardManager(token string, m *shard.Manager) *Router {
	c := api.NewClient(token)

	m.ForEach(func(shard shard.Shard) {
		s := shard.(*state.State)
		s.Client = c
	})

	r := &Router{
		Rest:          c,
		ShardManager:  m,
		commands:      make(map[string]*handler[*CommandContext]),
		autocompletes: make(map[string]*handler[*AutocompleteContext]),
		modals:        make(map[discord.ComponentID]*handler[*ModalContext]),
		buttons:       make(map[componentKey]*handler[*ButtonContext]),
		selects:       make(map[componentKey]*handler[*SelectContext]),
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

func (r *Router) Command(path string) *CommandBuilder {
	return &CommandBuilder{
		r:    r,
		path: path,
	}
}

func (r *Router) Autocomplete(path string) *AutocompleteBuilder {
	return &AutocompleteBuilder{
		r:    r,
		path: path,
	}
}

func (r *Router) Modal(id discord.ComponentID) *ModalBuilder {
	return &ModalBuilder{
		r:  r,
		id: id,
	}
}

func (r *Router) Button(id discord.ComponentID) *ButtonBuilder {
	return &ButtonBuilder{
		r:     r,
		id:    id,
		msgID: discord.NullMessageID,
	}
}

func (r *Router) Select(id discord.ComponentID) *SelectBuilder {
	return &SelectBuilder{
		r:     r,
		id:    id,
		msgID: discord.NullMessageID,
	}
}
