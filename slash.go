package bcr

import (
	"strings"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/starshine-sys/snowflake/v2"
)

// SyncCommands syncs slash commands in the given guilds.
// If no guilds are given, slash commands are synced globally.
// Router.Bot *must* be set before calling this function or it will panic!
func (r *Router) SyncCommands(guildIDs ...discord.GuildID) (err error) {
	r.cmdMu.Lock()
	cmds := []*Command{}
	for _, cmd := range r.cmds {
		if cmd.Options != nil && !inCmds(cmds, cmd.id) {
			cmds = append(cmds, cmd)
		}
	}
	r.cmdMu.Unlock()

	slashCmds := []discord.Command{}
	for _, cmd := range cmds {
		slashCmds = append(slashCmds, discord.Command{
			Name:        strings.ToLower(cmd.Name),
			Description: cmd.Summary,
			Options:     *cmd.Options,
		})
	}
	for _, g := range r.SlashGroups {
		slashCmds = append(slashCmds, g.Command())
	}

	if len(guildIDs) > 0 {
		return r.syncCommandsIn(slashCmds, guildIDs)
	}
	return r.syncCommandsGlobal(slashCmds)
}

func inCmds(cmds []*Command, id snowflake.ID) bool {
	for _, cmd := range cmds {
		if cmd.id == id {
			return true
		}
	}
	return false
}

func (r *Router) syncCommandsGlobal(cmds []discord.Command) (err error) {
	appID := discord.AppID(r.Bot.ID)
	s, _ := r.StateFromGuildID(0)

	deleted := []discord.CommandID{}
	current, err := s.Commands(appID)
	if err != nil {
		return err
	}

	for _, c := range current {
		if !in(cmds, c.Name) {
			deleted = append(deleted, c.ID)
		}
	}

	for _, id := range deleted {
		err = s.DeleteCommand(appID, id)
		if err != nil {
			return err
		}
	}

	_, err = s.BulkOverwriteCommands(appID, cmds)
	return
}

func in(cmds []discord.Command, name string) bool {
	for _, cmd := range cmds {
		if cmd.Name == name {
			return true
		}
	}
	return false
}

func (r *Router) syncCommandsIn(cmds []discord.Command, guildIDs []discord.GuildID) (err error) {
	appID := discord.AppID(r.Bot.ID)

	for _, guild := range guildIDs {
		s, _ := r.StateFromGuildID(guild)

		deleted := []discord.CommandID{}
		current, err := s.GuildCommands(appID, guild)
		if err != nil {
			return err
		}

		for _, c := range current {
			if !in(cmds, c.Name) {
				deleted = append(deleted, c.ID)
			}
		}

		for _, id := range deleted {
			err = s.DeleteGuildCommand(appID, guild, id)
			if err != nil {
				return err
			}
		}

		_, err = s.BulkOverwriteGuildCommands(appID, guild, cmds)
		if err != nil {
			return err
		}
	}

	return
}
