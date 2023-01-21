package bcr

import (
	"strings"
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

	// CollectFunc is the function used to collect the guild and channel an interaction takes place in.
	// The default CollectFunc uses a *state.State, but this can be overridden for more specialized caches.
	CollectFunc CollectFunc

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

	r.CollectFunc = r.DefaultCollectFunc

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

	r.CollectFunc = r.DefaultCollectFunc

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
	b := &ModalBuilder{
		r:  r,
		id: id,
	}

	if strings.HasSuffix(string(id), "*") {
		b.suffixWildcard = true
		b.id = discord.ComponentID(strings.TrimSuffix(string(id), "*"))
	} else if strings.HasPrefix(string(id), "*") {
		b.prefixWildcard = true
		b.id = discord.ComponentID(strings.TrimPrefix(string(id), "*"))
	}

	return b
}

func (r *Router) Button(id discord.ComponentID) *ButtonBuilder {
	b := &ButtonBuilder{
		r:     r,
		id:    id,
		msgID: discord.NullMessageID,
	}

	if strings.HasSuffix(string(id), "*") {
		b.suffixWildcard = true
		b.id = discord.ComponentID(strings.TrimSuffix(string(id), "*"))
	} else if strings.HasPrefix(string(id), "*") {
		b.prefixWildcard = true
		b.id = discord.ComponentID(strings.TrimPrefix(string(id), "*"))
	}

	return b
}

func (r *Router) Select(id discord.ComponentID) *SelectBuilder {
	b := &SelectBuilder{
		r:     r,
		id:    id,
		msgID: discord.NullMessageID,
	}

	if strings.HasSuffix(string(id), "*") {
		b.suffixWildcard = true
		b.id = discord.ComponentID(strings.TrimSuffix(string(id), "*"))
	} else if strings.HasPrefix(string(id), "*") {
		b.prefixWildcard = true
		b.id = discord.ComponentID(strings.TrimPrefix(string(id), "*"))
	}

	return b
}

func (r *Router) DefaultCollectFunc(s *state.State, guildID discord.GuildID, channelID discord.ChannelID) (
	g *discord.Guild,
	ch, parentCh *discord.Channel,
	err error,
) {
	if guildID.IsValid() {
		g, err = s.Guild(guildID)
		if err != nil {
			return nil, nil, nil, err
		}
		g.Roles, err = s.Roles(guildID)
		if err != nil {
			return nil, nil, nil, err
		}
	}

	// get the channel
	ch, err = s.Channel(channelID)
	if err != nil {
		return nil, nil, nil, err
	}

	if IsThread(ch) {
		parentCh, err = s.Channel(ch.ParentID)
		if err != nil {
			return nil, nil, nil, err
		}
	}

	return g, ch, parentCh, nil
}
