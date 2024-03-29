package bcr

import (
	"strings"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
)

// Group is used for creating slash subcommands.
// No, we can't use the normal system,
// because a command with subcommands can't *itself* be invoked as a command.
// Also, no subcommand groups because those make everything more complicated
// and shouldn't be needed at this scale. Might change in the future, who knows!
type Group struct {
	Name        string
	Description string
	Subcommands []*Command
}

// Add adds a subcommand to the group.
func (g *Group) Add(cmd *Command) *Group {
	g.Subcommands = append(g.Subcommands, cmd)
	return g
}

// Command returns the group as a discord.Command.
func (g Group) Command() api.CreateCommandData {
	c := api.CreateCommandData{
		Type:        discord.ChatInputCommand,
		Name:        strings.ToLower(g.Name),
		Description: g.Description,
	}

	for _, cmd := range g.Subcommands {
		if cmd.SlashCommand == nil {
			continue
		}

		options := []discord.CommandOptionValue(nil)
		if cmd.Options != nil {
			for _, o := range *cmd.Options {
				v, ok := o.(discord.CommandOptionValue)
				if ok {
					options = append(options, v)
				}
			}
		}

		c.Options = append(c.Options, &discord.SubcommandOption{
			OptionName:  strings.ToLower(cmd.Name),
			Description: cmd.Summary,
			Options:     options,
		})
	}

	return c
}

// AddGroup adds a slash command group. Will panic if the group's name already exists as a slash command!
func (r *Router) AddGroup(g *Group) {
	r.cmdMu.RLock()
	defer r.cmdMu.RUnlock()
	for _, cmd := range r.cmds {
		if strings.EqualFold(cmd.Name, g.Name) && cmd.Options != nil && cmd.SlashCommand != nil {
			panic("slash command with name " + g.Name + " already exists!")
		}
	}

	r.SlashGroups = append(r.SlashGroups, g)
}
