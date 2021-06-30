package bcr

import "github.com/diamondburned/arikawa/v2/discord"

// GreedyChannelParser parses all arguments until it finds an error.
// Returns the parsed channels and the position at which it stopped.
// If all arguments were parsed as channels, returns -1.
func (ctx *Context) GreedyChannelParser(args []string) (channels []*discord.Channel, n int) {
	for i, a := range args {
		c, err := ctx.ParseChannel(a)
		if err != nil {
			return channels, i
		}
		channels = append(channels, c)
	}
	return channels, -1
}

// GreedyMemberParser parses all arguments until it finds an error.
// Returns the parsed members and the position at which it stopped.
// If all arguments were parsed as members, returns -1.
func (ctx *Context) GreedyMemberParser(args []string) (members []*discord.Member, n int) {
	for i, a := range args {
		c, err := ctx.ParseMember(a)
		if err != nil {
			return members, i
		}
		members = append(members, c)
	}
	return members, -1
}

// GreedyRoleParser parses all arguments until it finds an error.
// Returns the parsed roles and the position at which it stopped.
// If all arguments were parsed as roles, returns -1.
func (ctx *Context) GreedyRoleParser(args []string) (roles []*discord.Role, n int) {
	for i, a := range args {
		c, err := ctx.ParseRole(a)
		if err != nil {
			return roles, i
		}
		roles = append(roles, c)
	}
	return roles, -1
}

// GreedyUserParser parses all arguments until it finds an error.
// Returns the parsed users and the position at which it stopped.
// If all arguments were parsed as users, returns -1.
func (ctx *Context) GreedyUserParser(args []string) (users []*discord.User, n int) {
	for i, a := range args {
		c, err := ctx.ParseUser(a)
		if err != nil {
			return users, i
		}
		users = append(users, c)
	}
	return users, -1
}
