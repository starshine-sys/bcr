package bcr

import (
	"strings"
	"sync"
	"time"

	"github.com/diamondburned/arikawa/v2/discord"
)

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

	subCmds map[string]*Command
	subMu   sync.RWMutex

	Permissions discord.Permissions
	Command     func(*Context) error

	GuildOnly bool
	OwnerOnly bool
	Cooldown  time.Duration
}

// AddSubcommand adds a subcommand to a command
func (c *Command) AddSubcommand(sub *Command) *Command {
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
