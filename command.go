package bcr

import (
	"strings"
	"sync"
	"time"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/spf13/pflag"
	"github.com/starshine-sys/snowflake/v2"
)

// CustomPerms is a custom permission checker
type CustomPerms interface {
	// The string used for the permissions if the check fails
	String() string

	// Returns true if the user has permission to run the command
	Check(*Context) (bool, error)
}

// Command is a single command, or a group
type Command struct {
	Name    string
	Aliases []string

	// Blacklistable commands use the router's blacklist function to check if they can be run
	Blacklistable bool

	// Summary is used in the command list
	Summary string
	// Description is used in the help command
	Description string
	// Usage is appended to the command name in help commands
	Usage string

	// Hidden commands are not returned from (*Router).Commands()
	Hidden bool

	Args *Args

	CustomPermissions CustomPerms

	// Flags is used to create a new flag set, which is then parsed before the command is run.
	// These can then be retrieved with the (*FlagSet).Get*() methods.
	Flags func(fs *pflag.FlagSet) *pflag.FlagSet

	subCmds map[string]*Command
	subMu   sync.RWMutex

	// GuildPermissions is the required *global* permissions
	GuildPermissions discord.Permissions
	// Permissions is the required permissions in the *context channel*
	Permissions discord.Permissions
	GuildOnly   bool
	OwnerOnly   bool

	Command  func(*Context) error
	Cooldown time.Duration

	// id is a unique ID. This is automatically generated on startup and is (pretty much) guaranteed to be unique *per session*. This ID will *not* be consistent between restarts.
	id snowflake.Snowflake
}

// AddSubcommand adds a subcommand to a command
func (c *Command) AddSubcommand(sub *Command) *Command {
	sub.id = sGen.Get()
	c.subMu.Lock()
	defer c.subMu.Unlock()
	if c.subCmds == nil {
		c.subCmds = make(map[string]*Command)
	}

	c.subCmds[strings.ToLower(sub.Name)] = sub
	for _, a := range sub.Aliases {
		c.subCmds[strings.ToLower(a)] = sub
	}

	return sub
}

// GetCommand gets a command by name
func (r *Router) GetCommand(name string) *Command {
	r.cmdMu.RLock()
	defer r.cmdMu.RUnlock()
	if v, ok := r.cmds[strings.ToLower(name)]; ok {
		return v
	}
	return nil
}

// GetCommand gets a command by name
func (c *Command) GetCommand(name string) *Command {
	c.subMu.RLock()
	defer c.subMu.RUnlock()
	if v, ok := c.subCmds[strings.ToLower(name)]; ok {
		return v
	}
	return nil
}

// Args is a minimum/maximum argument count.
// If either is -1, it's treated as "no minimum" or "no maximum".
// This replaces the Check* functions in Context.
type Args [2]int

// MinArgs returns an *Args with only a minimum number of arguments.
func MinArgs(i int) *Args {
	return &Args{i, -1}
}

// MaxArgs returns an *Args with only a maximum number of arguments.
func MaxArgs(i int) *Args {
	return &Args{-1, i}
}

// ArgRange returns an *Args with both a minimum and maximum number of arguments.
func ArgRange(i, j int) *Args {
	return &Args{i, j}
}

// ExactArgs returns an *Args with an exact number of required arguments.
func ExactArgs(i int) *Args {
	return &Args{i, i}
}

type requireRole struct {
	name string

	// owners override the role check
	owners []discord.UserID
	// any of these roles is required for the check to succeed
	roles []discord.RoleID
}

var _ CustomPerms = (*requireRole)(nil)

func (r *requireRole) String() string {
	return r.name
}

func (r *requireRole) Check(ctx *Context) (bool, error) {
	for _, u := range r.owners {
		if u == ctx.Author.ID {
			return true, nil
		}
	}

	if ctx.Member == nil {
		return false, nil
	}

	for _, r := range r.roles {
		for _, mr := range ctx.Member.RoleIDs {
			if r == mr {
				return true, nil
			}
		}
	}

	return false, nil
}

// RequireRole returns a CustomPerms that requires the given roles.
// If any of r's owner IDs are not valid snowflakes, this function will panic!
func (r *Router) RequireRole(name string, roles ...discord.RoleID) CustomPerms {
	owners := []discord.UserID{}

	for _, u := range r.BotOwners {
		id, err := discord.ParseSnowflake(u)
		if err != nil {
			panic(err)
		}

		owners = append(owners, discord.UserID(id))
	}

	return &requireRole{
		name:   name,
		owners: owners,
		roles:  roles,
	}
}
