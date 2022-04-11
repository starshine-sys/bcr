package bcr

import (
	"github.com/starshine-sys/snowflake/v2"
)

// Commands returns a list of commands
func (r *Router) Commands() []*Command {
	// deduplicate commands
	sf := make([]snowflake.Snowflake, 0)
	cmds := make([]*Command, 0)
	for _, c := range r.cmds {
		if !snowflakeInSlice(c.id, sf) {
			sf = append(sf, c.id)
			cmds = append(cmds, c)
		}
	}

	return cmds
}
