package bcr

import (
	"strings"

	"emperror.dev/errors"
	"github.com/diamondburned/arikawa/v3/discord"
)

// SlashCommandOption is a single slash command option with a name and value.
type SlashCommandOption struct {
	ctx *SlashContext
	discord.InteractionOption
}

// Option returns an option by name, or an empty option if it's not found.
func (ctx *SlashContext) Option(name string) (option SlashCommandOption) {
	for _, o := range ctx.CommandOptions {
		if strings.EqualFold(o.Name, name) {
			return SlashCommandOption{
				ctx:               ctx,
				InteractionOption: o,
			}
		}
	}
	return SlashCommandOption{ctx: ctx}
}

// SlashCommandOptions is a slice of slash command options.
type SlashCommandOptions struct {
	ctx *SlashContext
	s   []discord.InteractionOption
}

// NewSlashCommandOptions returns a new SlashCommandOptions.
func NewSlashCommandOptions(ctx *SlashContext, s []discord.InteractionOption) SlashCommandOptions {
	return SlashCommandOptions{
		ctx: ctx, s: s,
	}
}

// Get returns an option by name, or an empty option if it's not found.
func (options SlashCommandOptions) Get(name string) SlashCommandOption {
	for _, o := range options.s {
		if strings.EqualFold(o.Name, name) {
			return SlashCommandOption{
				ctx:               options.ctx,
				InteractionOption: o,
			}
		}
	}
	return SlashCommandOption{ctx: options.ctx}
}

// Option returns an option by name, or an empty option if it's not found.
func (o SlashCommandOption) Option(name string) SlashCommandOption {
	for _, option := range o.Options {
		if strings.EqualFold(option.Name, name) {
			return SlashCommandOption{
				ctx:               o.ctx,
				InteractionOption: option,
			}
		}
	}
	return SlashCommandOption{}
}

// Int returns the option as an integer, or 0 if it can't be converted.
func (o SlashCommandOption) Int() int64 {
	i, err := o.InteractionOption.Int()
	if err != nil {
		return 0
	}
	return i
}

// Float returns the option as a float, or 0 if it can't be converted.
func (o SlashCommandOption) Float() float64 {
	i, err := o.InteractionOption.Float()
	if err != nil {
		return 0
	}
	return i
}

// Bool returns the option as a bool, or false if it can't be converted.
func (o SlashCommandOption) Bool() bool {
	b, err := o.InteractionOption.Bool()
	if err != nil {
		return false
	}
	return b
}

// User returns the option as a user.
func (o SlashCommandOption) User() (*discord.User, error) {
	id, err := o.Snowflake()
	if err != nil {
		return nil, err
	}

	return o.ctx.State.User(discord.UserID(id))
}

// Member returns the option as a member.
func (o SlashCommandOption) Member() (*discord.Member, error) {
	if o.ctx.Guild == nil {
		return nil, errors.Sentinel("not in a guild")
	}

	id, err := o.Snowflake()
	if err != nil {
		return nil, err
	}

	return o.ctx.State.Member(o.ctx.Guild.ID, discord.UserID(id))
}

// Role returns the option as a role.
func (o SlashCommandOption) Role() (*discord.Role, error) {
	if o.ctx.Guild == nil {
		return nil, errors.Sentinel("not in a guild")
	}

	id, err := o.Snowflake()
	if err != nil {
		return nil, err
	}

	return o.ctx.State.Role(o.ctx.Guild.ID, discord.RoleID(id))
}

// Channel returns the option as a channel.
func (o SlashCommandOption) Channel() (*discord.Channel, error) {
	id, err := o.Snowflake()
	if err != nil {
		return nil, err
	}

	return o.ctx.State.Channel(discord.ChannelID(id))
}

// GetStringFlag gets the named flag as a string, or falls back to an empty string.
func (ctx *SlashContext) GetStringFlag(name string) string {
	return ctx.Option(name).String()
}

// GetBoolFlag gets the named flag as a bool, or falls back to false.
func (ctx *SlashContext) GetBoolFlag(name string) bool {
	return ctx.Option(name).Bool()
}

// GetIntFlag gets the named flag as an int64, or falls back to 0.
func (ctx *SlashContext) GetIntFlag(name string) int64 {
	return ctx.Option(name).Int()
}

// GetFloatFlag gets the named flag as a float64, or falls back to 0.
func (ctx *SlashContext) GetFloatFlag(name string) float64 {
	return ctx.Option(name).Float()
}

// GetUserFlag gets the named flag as a user.
func (ctx *SlashContext) GetUserFlag(name string) (*discord.User, error) {
	return ctx.Option(name).User()
}

// GetMemberFlag gets the named flag as a member.
func (ctx *SlashContext) GetMemberFlag(name string) (*discord.Member, error) {
	return ctx.Option(name).Member()
}

// GetRoleFlag gets the named flag as a role.
func (ctx *SlashContext) GetRoleFlag(name string) (*discord.Role, error) {
	return ctx.Option(name).Role()
}

// GetChannelFlag gets the named flag as a channel.
func (ctx *SlashContext) GetChannelFlag(name string) (*discord.Channel, error) {
	return ctx.Option(name).Channel()
}
