package bcr

import (
	"emperror.dev/errors"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/state"
)

const (
	ErrNotCommand = errors.Sentinel("not a command interaction")
	ErrNotModal   = errors.Sentinel("not a modal interaction")
	ErrNotButton  = errors.Sentinel("not a button interaction")
)

const ErrUnknownCommand = errors.Sentinel("no command with that path found")

type HasContext interface {
	Ctx() *Context
}

type HandlerFunc[T HasContext] func(T) error
type CollectFunc func(s *state.State, guildID discord.GuildID, channelID discord.ChannelID) (
	g *discord.Guild,
	ch, parentCh *discord.Channel,
	err error,
)

type handler[T HasContext] struct {
	check   Check[T]
	handler HandlerFunc[T]
	// only used in button/select interactions
	once                           bool
	prefixWildcard, suffixWildcard bool
}

func (hn *handler[T]) doCheck(ctx T) error {
	if hn.check == nil {
		return nil
	}
	return hn.check(ctx)
}
