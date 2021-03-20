package bcr

import "github.com/diamondburned/arikawa/v2/discord"

// Roles are a sortable collection of discord.Role
type Roles []discord.Role

func (r Roles) Len() int {
	return len(r)
}

func (r Roles) Less(i, j int) bool {
	return r[i].Position > r[j].Position
}

func (r Roles) Swap(i, j int) {
	r[i], r[j] = r[j], r[i]
}
