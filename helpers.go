package bcr

import "github.com/diamondburned/arikawa/v3/discord"

// IsThread returns true if the given channel is a thread channel.
func IsThread(ch *discord.Channel) bool {
	return ch.Type == discord.GuildNewsThread ||
		ch.Type == discord.GuildPublicThread ||
		ch.Type == discord.GuildPrivateThread
}
