package bcr

import (
	"errors"
	"fmt"
	"strings"
	"sync"
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

	// append the current command to fullCommandPath, for help strings
	ctx.fullCommandPath = append(ctx.fullCommandPath, ctx.Command)
	// check if the second argument is `help` or `usage`, if so, show the command's help
	err = ctx.tryHelp()
	if err != nil {
		return err
	}

	// if the command has subcommands, try those
	if c.subCmds != nil && len(ctx.Args) > 0 {
		// shift all arguments over by 1
		ctx.Command = ctx.Args[0]
		if len(ctx.Args) > 0 {
			ctx.Args = ctx.Args[1:]
			ctx.RawArgs = TrimPrefixesSpace(ctx.RawArgs, ctx.Command)
		} else {
			ctx.Args = []string{}
			ctx.RawArgs = ""
		}
		err = r.execInner(ctx, c.subCmds, &c.subMu)
		// return all errors, including errCommandRun, so further layers stop executing as well
		if err != nil {
			return err
		}
	}

	// set the context's Cmd field to the command
	ctx.Cmd = c

	// if the command is guild-only, and this isn't a guild channel, error
	if c.GuildOnly && ctx.Message.GuildID == 0 {
		_, err = ctx.Send(":x: This command cannot be run in DMs.", nil)
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
		_, err = ctx.Send(":x: This command can only be used by a bot owner.", nil)
		if err != nil {
			return err
		}
		return errCommandRun
	}

	// if the command requires extra permissions
	if c.Permissions != 0 {
		// check command permissions
		if err = ctx.CheckPerms(); err != nil {
			_, err = ctx.Send(
				fmt.Sprintf(":x: You are not allowed to use this command. You are missing the following permissions:\n> ```%v```", err), nil,
			)
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
		s, b := c.CustomPermissions(ctx)
		// if it returned false, error and return
		if !b {
			_, err = ctx.Send(":x: You are not allowed to use this command:\n> "+s, nil)
			if err != nil {
				return err
			}
			return errCommandRun
		}
	}

	// check for a cooldown
	if r.cooldowns.Get(strings.Join(ctx.fullCommandPath, "-"), ctx.Author.ID, ctx.Channel.ID) {
		_, err = ctx.Send(fmt.Sprintf(":x: This command can only be run once every %v.", c.Cooldown), nil)
		if err != nil {
			return err
		}
		return errCommandRun
	}

	// execute the command
	err = c.Command(ctx)
	if err != nil {
		return err
	}
	// if there's a cooldown, set it
	if c.Cooldown != 0 {
		r.cooldowns.Set(strings.Join(ctx.fullCommandPath, "-"), ctx.Author.ID, ctx.Channel.ID, c.Cooldown)
	}

	// return with errCommandRun, which indicates to an outer layer (if any) that it should stop execution
	return errCommandRun
}
