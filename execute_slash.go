package bcr

import "github.com/diamondburned/arikawa/v3/gateway"

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
	err = r.executeSlash(ctx)
	if err == errCommandRun {
		return nil
	}
	return
}

func (r *Router) executeSlash(ctx *SlashContext) (err error) {
	cmd, ok := r.cmds[ctx.CommandName]
	if !ok || cmd.SlashCommand == nil {
		err = ctx.SendfX("Looks like you found a command (``%v``) that's registered as a slash command, but doesn't work as one :(\nPlease report this to the bot developer as this is a bug!", EscapeBackticks(ctx.CommandName))
		if err == nil {
			return errCommandRun
		}
		return err
	}

	if (cmd.GuildOnly || cmd.Permissions != 0) && !ctx.Event.GuildID.IsValid() {
		err = ctx.SendEphemeral(":x: This command cannot be run in DMs.")
		if err != nil {
			return err
		}
		return errCommandRun
	}

	err = cmd.SlashCommand(ctx)
	if err == nil {
		return errCommandRun
	}
	return err
}
