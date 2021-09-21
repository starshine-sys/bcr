package bcr

import (
	"fmt"
	"strings"
	"sync"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
)

// InteractionCreate is called when an interaction create event is received.
func (r *Router) InteractionCreate(ic *gateway.InteractionCreateEvent) {
	if ic.Type != gateway.CommandInteraction {
		return
	}

	ctx, err := r.NewSlashContext(ic)
	if err != nil {
		r.Logger.Error("Couldn't create slash context: %v", err)
		return
	}

	err = r.ExecuteSlash(ctx)
	if err != nil {
		r.Logger.Error("Couldn't create slash context: %v", err)
	}
}

// ExecuteSlash executes slash commands. Only one layer for now, so no subcommands, sorry :(
func (r *Router) ExecuteSlash(ctx *SlashContext) (err error) {
	err = r.executeSlash(true, ctx, r.cmds, &r.cmdMu)
	if err == errCommandRun {
		return nil
	}
	return
}

func errCommand(err error) error {
	if err == nil {
		return errCommandRun
	}
	return err
}

func (r *Router) executeSlash(isTopLevel bool, ctx *SlashContext, cmds map[string]*Command, mu *sync.RWMutex) (err error) {
	// first, check subcommands
	if len(ctx.CommandOptions) > 0 && isTopLevel {
		for _, g := range r.SlashGroups {
			if strings.EqualFold(g.Name, ctx.CommandName) {
				nctx := &SlashContext{}
				*nctx = *ctx
				nctx.CommandName = ctx.CommandOptions[0].Name
				nctx.CommandOptions = ctx.CommandOptions[0].Options

				// convert subcommands slice to a map
				m := map[string]*Command{}
				for _, cmd := range g.Subcommands {
					m[strings.ToLower(cmd.Name)] = cmd
				}
				var nmu sync.RWMutex // this doesn't matter so we just create a new one

				return r.executeSlash(false, nctx, m, &nmu)
			}
		}
	}

	// else, we try top-level commands (or skip to this immediately if it isn't the top level)
	mu.RLock()
	cmd, ok := cmds[ctx.CommandName]
	if !ok || cmd.SlashCommand == nil {
		mu.RUnlock()
		err = ctx.SendEphemeral(fmt.Sprintf("Looks like you found a command (``%v``) that's registered as a slash command, but doesn't work as one :(\nPlease report this to the bot developer as this is a bug!", EscapeBackticks(ctx.CommandName)))
		return errCommand(err)
	}
	mu.RUnlock()

	ctx.Command = cmd

	if (cmd.GuildOnly || cmd.Permissions != 0) && !ctx.Event.GuildID.IsValid() {
		err = ctx.SendEphemeral(":x: This command cannot be run in DMs.")
		return errCommand(err)
	}

	if r.BlacklistFunc != nil && cmd.Blacklistable {
		if r.BlacklistFunc(ctx) {
			err = ctx.SendEphemeral("This command can't be used here.")
			return errCommand(err)
		}
	}

	if cmd.GuildPermissions != 0 {
		if ctx.Guild == nil || ctx.Member == nil {
			err = ctx.SendEphemeral(":x: This command cannot be used in DMs.")
			return errCommand(err)
		}
		if !ctx.GuildPerms().Has(cmd.GuildPermissions) {
			err = ctx.SendEphemeral(fmt.Sprintf(":x: You are not allowed to use this command. You are missing the following permissions:\n> ```%v```", strings.Join(PermStrings(cmd.GuildPermissions), ", ")))
			return errCommand(err)
		}
	}

	if cmd.Permissions != 0 {
		if ctx.Member != nil && ctx.Guild != nil {
			if !discord.CalcOverwrites(*ctx.Guild, *ctx.Channel, *ctx.Member).Has(cmd.Permissions) {
				err = ctx.SendEphemeral(fmt.Sprintf(":x: You are not allowed to use this command. You are missing the following permissions:\n> ```%v```", strings.Join(PermStrings(cmd.Permissions), ", ")))
				return errCommand(err)
			}
		} else {
			err = ctx.SendEphemeral(":x: This command cannot be run in DMs.")
			return errCommand(err)
		}
	}

	if cmd.OwnerOnly {
		isOwner := false
		for _, u := range ctx.Router.BotOwners {
			if ctx.Author.ID.String() == u {
				isOwner = true
				break
			}
		}

		if !isOwner {
			err = ctx.SendEphemeral(":x: This command can only be used by a bot owner.")
			return errCommand(err)
		}
	}

	if cmd.CustomPermissions != nil {
		b, err := cmd.CustomPermissions.Check(ctx)
		// if it errored, send that error and return
		if err != nil {
			err = ctx.SendEphemeral(fmt.Sprintf(":x: An internal error occurred when checking your permissions.\nThe following permission(s) could not be checked:\n> ```%s```", cmd.CustomPermissions))
			return errCommand(err)
		}

		// else if it returned false, show that error and return
		if !b {
			err = ctx.SendEphemeral(fmt.Sprintf(":x: You are not allowed to use this command. You are missing the following permission(s):\n> ```%s```", cmd.CustomPermissions))
			return errCommand(err)
		}
	}

	err = cmd.SlashCommand(ctx)
	return errCommand(err)
}
