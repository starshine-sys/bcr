package bcr

import (
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/spf13/pflag"
)

var errCommandRun = errors.New("command run in layer")

// Execute executes the command router
func (r *Router) Execute(ctx *Context) (err error) {
	err = r.execInner(ctx, r.cmds, &r.cmdMu)
	if err == errCommandRun {
		return nil
	}
	return err
}

func (r *Router) execInner(ctx *Context, cmds map[string]*Command, mu *sync.RWMutex) (err error) {
	var (
		c  *Command
		ok bool
	)
	mu.RLock()
	// check if a command matches, if not, return
	if c, ok = cmds[ctx.Command]; !ok {
		mu.RUnlock()
		return
	}
	mu.RUnlock()

	// append the current command to FullCommandPath, for help strings
	ctx.FullCommandPath = append(ctx.FullCommandPath, ctx.Command)
	// check if the second argument is `help` or `usage`, if so, show the command's help
	err = ctx.tryHelp()
	if err != nil {
		return err
	}

	// if the command has subcommands, try those
	if c.subCmds != nil && len(ctx.Args) > 0 {
		if _, ok = c.subCmds[ctx.Peek()]; ok {
			ctx.Command = ctx.Pop()
			err = r.execInner(ctx, c.subCmds, &c.subMu)
			// return all errors, including errCommandRun, so further layers stop executing as well
			if err != nil {
				return err
			}
		}
	}

	// set the context's Cmd field to the command
	ctx.Cmd = c

	// if the command is guild-only or needs extra permissions, and this isn't a guild channel, error
	if (c.GuildOnly || c.Permissions != 0) && ctx.Message.GuildID == 0 {
		_, err = ctx.Send(":x: This command cannot be run in DMs.")
		if err != nil {
			return err
		}
		return errCommandRun
	}

	// check if the command can be blacklisted
	if r.BlacklistFunc != nil && c.Blacklistable {
		// if the channel's blacklisted, return
		if r.BlacklistFunc(ctx) {
			return errCommandRun
		}
	}

	// if the command requires bot owner to use, and the user isn't a bot owner, error
	if !ctx.checkOwner() {
		_, err = ctx.Send(":x: This command can only be used by a bot owner.")
		if err != nil {
			return err
		}
		return errCommandRun
	}

	if c.GuildPermissions != 0 {
		if ctx.Guild == nil || ctx.Member == nil {
			_, err = ctx.Send(":x: This command cannot be used in DMs.")
			return errCommandRun
		}
		if !ctx.GuildPerms().Has(c.GuildPermissions) {
			_, err = ctx.Sendf(":x: You are not allowed to use this command. You are missing the following permissions:\n> ```%v```", strings.Join(PermStrings(c.GuildPermissions), ", "))
			// if there's an error, return it
			if err != nil {
				return err
			}
			// but if not, return errCommandRun so we don't try running more
			return errCommandRun
		}
	}

	if c.Permissions != 0 {
		if ctx.Guild == nil || ctx.Channel == nil || ctx.Member == nil {
			_, err = ctx.Send(":x: This command cannot be used in DMs.")
			return errCommandRun
		}
		if !discord.CalcOverwrites(*ctx.Guild, *ctx.Channel, *ctx.Member).Has(c.Permissions) {
			_, err = ctx.Sendf(":x: You are not allowed to use this command. You are missing the following permissions:\n> ```%v```", strings.Join(PermStrings(c.Permissions), ", "))
			// if there's an error, return it
			if err != nil {
				return err
			}
			// but if not, return errCommandRun so we don't try running more
			return errCommandRun
		}
	}

	// if the command has a custom permission handler, check it
	if c.CustomPermissions != nil {
		b, err := c.CustomPermissions.Check(ctx)
		// if it errored, send that error and return
		if err != nil {
			_, err = ctx.Send(fmt.Sprintf(":x: An internal error occurred when checking your permissions.\nThe following permission(s) could not be checked:\n> ```%s```", c.CustomPermissions.String(ctx)))
			if err != nil {
				return err
			}
			return errCommandRun
		}

		// else if it returned false, show that error and return
		if !b {
			_, err = ctx.Send(fmt.Sprintf(":x: You are not allowed to use this command. You are missing the following permission(s):\n> ```%s```", c.CustomPermissions.String(ctx)))
			if err != nil {
				return err
			}
			return errCommandRun
		}
	}

	// check router-level permissions
	// (usually, custom permission systems)
	if r.PermissionCheck != nil {
		_, allowed, data := r.PermissionCheck(ctx, true)
		if !allowed {
			if _, err = ctx.State.SendMessageComplex(ctx.Channel.ID, data); err != nil {
				return err
			}
			return errCommandRun
		}
	}

	// check for a cooldown
	if r.cooldowns.Get(strings.Join(ctx.FullCommandPath, "-"), ctx.Author.ID, ctx.Channel.ID) {
		_, err = ctx.Sendf(":x: This command can only be run once every %v.", c.Cooldown)
		if err != nil {
			return err
		}
		return errCommandRun
	}

	// if the command has any flags set, parse those
	if c.Flags != nil {
		ctx.Flags = c.Flags(pflag.NewFlagSet("", pflag.ContinueOnError))
		ctx.Flags.ParseErrorsWhitelist.UnknownFlags = true

		err = ctx.Flags.Parse(ctx.Args)
		if err != nil {
			_, err = ctx.Send(":x: There was an error parsing your input. Try checking this command's help.")
			return
		}
		ctx.Args = ctx.Flags.Args()
	}

	// check arguments
	err = ctx.argCheck()
	if err != nil {
		return err
	}

	if c.Command != nil {
		err = c.Command(ctx)
	} else {
		err = c.SlashCommand(ctx)
	}
	if err != nil {
		return err
	}
	// if there's a cooldown, set it
	if c.Cooldown != 0 {
		r.cooldowns.Set(strings.Join(ctx.FullCommandPath, "-"), ctx.Author.ID, ctx.Channel.ID, c.Cooldown)
	}

	// return with errCommandRun, which indicates to an outer layer (if any) that it should stop execution
	return errCommandRun
}
