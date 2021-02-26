package bcr

import (
	"github.com/diamondburned/arikawa/v2/discord"
)

// MatchPrefix returns true if the message content contains any of the prefixes
func (r *Router) MatchPrefix(m discord.Message) bool {
	return r.Prefixer(m) != -1
}
