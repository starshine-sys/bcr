package bcr

import "strings"

// MatchPrefix returns true if the message content contains any of the prefixes
func (r *Router) MatchPrefix(content string) bool {
	return HasAnyPrefix(strings.ToLower(content), r.Prefixes...)
}
