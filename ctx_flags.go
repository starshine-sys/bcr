package bcr

import (
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
	if ctx.Flags == nil {
		return ""
	}

	v, err := ctx.Flags.GetString(name)
	if err != nil {
		return ""
	}
	return v
}

// GetBoolFlag gets the named flag as a bool, or falls back to false.
func (ctx *Context) GetBoolFlag(name string) bool {
	if ctx.Flags == nil {
		return false
	}

	v, err := ctx.Flags.GetBool(name)
	if err != nil {
		return false
	}
	return v
}

// GetIntFlag gets the named flag as an int64, or falls back to 0.
func (ctx *Context) GetIntFlag(name string) int64 {
	if ctx.Flags == nil {
		return 0
	}

	v, err := ctx.Flags.GetInt64(name)
	if err != nil {
		return 0
	}
	return v
}

// GetFloatFlag gets the named flag as a float64, or falls back to 0.
func (ctx *Context) GetFloatFlag(name string) float64 {
	if ctx.Flags == nil {
		return 0
	}

	v, err := ctx.Flags.GetFloat64(name)
	if err != nil {
		return 0
	}
	return v
}

// GetUserFlag gets the named flag as a user.
func (ctx *Context) GetUserFlag(name string) (*discord.User, error) {
	if ctx.Flags == nil {
		return nil, ErrUserNotFound
	}

	v, err := ctx.Flags.GetString(name)
	if err != nil {
		return nil, ErrUserNotFound
	}
	return ctx.ParseUser(v)
}

// GetMemberFlag gets the named flag as a member.
func (ctx *Context) GetMemberFlag(name string) (*discord.Member, error) {
	if ctx.Flags == nil {
		return nil, ErrMemberNotFound
	}

	v, err := ctx.Flags.GetString(name)
	if err != nil {
		return nil, ErrMemberNotFound
	}
	return ctx.ParseMember(v)
}

// GetRoleFlag gets the named flag as a role.
func (ctx *Context) GetRoleFlag(name string) (*discord.Role, error) {
	if ctx.Flags == nil {
		return nil, ErrRoleNotFound
	}

	v, err := ctx.Flags.GetString(name)
	if err != nil {
		return nil, ErrRoleNotFound
	}
	return ctx.ParseRole(v)
}

// GetChannelFlag gets the named flag as a channel.
func (ctx *Context) GetChannelFlag(name string) (*discord.Channel, error) {
	if ctx.Flags == nil {
		return nil, ErrChannelNotFound
	}

	v, err := ctx.Flags.GetString(name)
	if err != nil {
		return nil, ErrChannelNotFound
	}
	return ctx.ParseChannel(v)
}
