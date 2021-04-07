package bcr

import (
	"fmt"
	"strings"

	"github.com/diamondburned/arikawa/v2/discord"
	"github.com/spf13/pflag"
	"github.com/starshine-sys/snowflake/v2"
)

func (ctx *Context) tryHelp() error {
	// if there's no args return nil
	if len(ctx.Args) == 0 {
		return nil
	}

	// if the first argument isn't "help" or "usage" return nil
	if strings.ToLower(ctx.Args[0]) != "help" && strings.ToLower(ctx.Args[0]) != "usage" {
		return nil
	}

	// execute the help command
	err := ctx.Help(ctx.FullCommandPath)
	if err != nil {
		return err
	}
	return errCommandRun
}

// Help sends a help embed for the command
func (ctx *Context) Help(path []string) (err error) {
	// recurse into subcommands
	cmds := ctx.Router.cmds
	var cmd *Command
	for i, n := range path {
		var ok bool

		// we tried recursing, but the map is nil, so the command wasn't found
		if cmds == nil {
			_, err = ctx.Send(fmt.Sprintf(":x: Command ``%v`` not found.", EscapeBackticks(strings.Join(path, " "))), nil)
			return err
		}

		// the command name wasn't found
		if cmd, ok = cmds[n]; !ok {
			_, err = ctx.Send(fmt.Sprintf(":x: Command ``%v`` not found.", EscapeBackticks(strings.Join(path, " "))), nil)
			return err
		}

		// we've not reached the end of the loop, so try recursing
		if i != len(path)-1 {
			cmds = cmd.subCmds
		}
	}

	if cmd == nil {
		_, err = ctx.Send(fmt.Sprintf(":x: Command ``%v`` not found.", EscapeBackticks(strings.Join(path, " "))), nil)
		return err
	}

	var fs *pflag.FlagSet
	if cmd.Flags != nil {
		fs = cmd.Flags(pflag.NewFlagSet("", pflag.ContinueOnError))
	}

	fields := make([]discord.EmbedField, 0)

	if cmd.Description != "" {
		fields = append(fields, discord.EmbedField{
			Name:  "Description",
			Value: cmd.Description,
		})
	}

	// get full names for path
	var title []string
	c := ctx.Router.GetCommand(path[0])
	title = append(title, c.Name)
	if len(path) > 1 {
		for _, p := range path[1:] {
			c = c.GetCommand(p)
			title = append(title, c.Name)
		}
	}

	usage := strings.Join(title, " ")

	if fs != nil {
		usage += " "
		fs.VisitAll(func(f *pflag.Flag) {
			usage += fmt.Sprintf("[-%v %v] ", f.Shorthand, f.Value.Type())
		})
	}

	if cmd.Usage != "" {
		usage += " " + cmd.Usage
	}

	fields = append(fields, discord.EmbedField{
		Name:  "Usage",
		Value: "`" + strings.TrimSpace(usage) + "`",
	})

	if fs != nil {
		var desc string

		fs.VisitAll(func(f *pflag.Flag) {
			desc += fmt.Sprintf("`-%v, --%v`: %v\n", f.Shorthand, f.Name, f.Usage)
		})

		fields = append(fields, discord.EmbedField{
			Name:  "Flags",
			Value: desc,
		})
	}

	if cmd.Permissions != 0 {
		fields = append(fields, discord.EmbedField{
			Name:  "Required permissions",
			Value: fmt.Sprintf("`%v`", strings.Join(PermStrings(cmd.Permissions), ", ")),
		})
	}

	if len(cmd.Aliases) != 0 {
		fields = append(fields, discord.EmbedField{
			Name:  "Aliases",
			Value: fmt.Sprintf("`%v`", strings.Join(cmd.Aliases, ", ")),
		})
	}
	if cmd.subCmds != nil {
		// deduplicate subcommands
		sf := make([]snowflake.Snowflake, 0)
		subCmds := make([]*Command, 0)
		for _, c := range cmd.subCmds {
			if !snowflakeInSlice(c.id, sf) {
				sf = append(sf, c.id)
				subCmds = append(subCmds, c)
			}
		}

		var b strings.Builder
		var i int
		for _, v := range subCmds {
			i++
			// if this is the last command, add a *special* list thingy
			if i == len(subCmds) {
				b.WriteString("`└─ ")
			} else {
				b.WriteString("`├─ ")
			}
			b.WriteString(v.Name)
			b.WriteString("`")
			if i != len(subCmds) {
				b.WriteString("\n")
			}
		}
		fields = append(fields, discord.EmbedField{
			Name:  "Subcommand(s)",
			Value: b.String(),
		})
	}

	_, err = ctx.Send("", &discord.Embed{
		Title:       "`" + strings.ToUpper(strings.Join(title, " ")) + "`",
		Description: DefaultValue(cmd.Summary, "No summary provided"),
		Fields:      fields,
		Color:       ctx.Router.EmbedColor,
	})
	return err
}
