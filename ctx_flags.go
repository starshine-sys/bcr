package bcr

import (
	"strings"

	"github.com/diamondburned/arikawa/v3/discord"
)

// Flags defines methods to retrieve flags from a command input.
type Flags interface {
	GetStringFlag(name string) string
	GetBoolFlag(name string) bool
	GetIntFlag(name string) int64
	GetFloatFlag(name string) float64

	GetUserFlag(name string) (*discord.User, error)
	GetMemberFlag(name string) (*discord.Member, error)
	GetRoleFlag(name string) (*discord.Role, error)
	GetChannelFlag(name string) (*discord.Channel, error)
}

// GetStringFlag gets the named flag as a string, or falls back to an empty string.
func (ctx *Context) GetStringFlag(name string) string {
	v, ok := ctx.FlagMap[strings.ToLower(name)]
	if !ok {
		return ""
	}

	s, ok := v.(*string)
	if !ok || s == nil {
		return ""
	}
	return *s
}

// GetBoolFlag gets the named flag as a bool, or falls back to false.
func (ctx *Context) GetBoolFlag(name string) bool {
	v, ok := ctx.FlagMap[strings.ToLower(name)]
	if !ok {
		return false
	}

	b, ok := v.(*bool)
	if !ok || b == nil {
		return false
	}
	return *b
}

// GetIntFlag gets the named flag as an int64, or falls back to 0.
func (ctx *Context) GetIntFlag(name string) int64 {
	v, ok := ctx.FlagMap[strings.ToLower(name)]
	if !ok {
		return 0
	}

	i, ok := v.(*int64)
	if !ok || i == nil {
		return 0
	}
	return *i
}

// GetFloatFlag gets the named flag as a float64, or falls back to 0.
func (ctx *Context) GetFloatFlag(name string) float64 {
	v, ok := ctx.FlagMap[strings.ToLower(name)]
	if !ok {
		return 0
	}

	f, ok := v.(*float64)
	if !ok || f == nil {
		return 0
	}
	return *f
}

// GetUserFlag gets the named flag as a user.
func (ctx *Context) GetUserFlag(name string) (*discord.User, error) {
	v, ok := ctx.FlagMap[strings.ToLower(name)]
	if !ok {
		return nil, ErrUserNotFound
	}

	s, ok := v.(*string)
	if !ok || s == nil {
		return nil, ErrUserNotFound
	}

	return ctx.ParseUser(*s)
}

// GetMemberFlag gets the named flag as a member.
func (ctx *Context) GetMemberFlag(name string) (*discord.Member, error) {
	v, ok := ctx.FlagMap[strings.ToLower(name)]
	if !ok {
		return nil, ErrMemberNotFound
	}

	s, ok := v.(*string)
	if !ok || s == nil {
		return nil, ErrMemberNotFound
	}

	return ctx.ParseMember(*s)
}

// GetRoleFlag gets the named flag as a role.
func (ctx *Context) GetRoleFlag(name string) (*discord.Role, error) {
	v, ok := ctx.FlagMap[strings.ToLower(name)]
	if !ok {
		return nil, ErrRoleNotFound
	}

	s, ok := v.(*string)
	if !ok || s == nil {
		return nil, ErrRoleNotFound
	}

	return ctx.ParseRole(*s)
}

// GetChannelFlag gets the named flag as a channel.
func (ctx *Context) GetChannelFlag(name string) (*discord.Channel, error) {
	v, ok := ctx.FlagMap[strings.ToLower(name)]
	if !ok {
		return nil, ErrChannelNotFound
	}

	s, ok := v.(*string)
	if !ok || s == nil {
		return nil, ErrChannelNotFound
	}

	return ctx.ParseChannel(*s)
}
