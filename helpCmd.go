package bcr

import (
	"fmt"
	"strings"

	"github.com/Starshine113/snowflake"
	"github.com/diamondburned/arikawa/v2/discord"
)

// CommandList shows a list of commands in embed form
func (r *Router) CommandList(ctx *Context) (err error) {
	// deduplicate commands
	sf := make([]snowflake.Snowflake, 0)
	cmds := make([]*Command, 0)
	for _, c := range ctx.Router.cmds {
		if !snowflakeInSlice(c.id, sf) {
			sf = append(sf, c.id)
			cmds = append(cmds, c)
		}
	}

	cmdSlices := make([][]*Command, 0)

	for i := 0; i < len(cmds); i += 15 {
		end := i + 15

		if end > len(cmds) {
			end = len(cmds)
		}

		cmdSlices = append(cmdSlices, cmds[i:end])
	}

	embeds := make([]discord.Embed, 0)

	for i, slice := range cmdSlices {
		var s strings.Builder
		for _, c := range slice {
			s.WriteString(fmt.Sprintf("`%v`: %v\n", c.Name, c.Summary))
		}

		embeds = append(embeds, discord.Embed{
			Title:       fmt.Sprintf("List of commands (%v)", len(cmdSlices)),
			Description: s.String(),
			Color:       r.EmbedColor,

			Footer: &discord.EmbedFooter{
				Text: fmt.Sprintf("Page %v/%v", i+1, len(cmdSlices)),
			},
		})
	}

	_, err = ctx.PagedEmbed(embeds)
	return err
}
