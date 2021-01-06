package bcr

import (
	"time"

	"github.com/ReneKroon/ttlcache/v2"
	"github.com/diamondburned/arikawa/v2/discord"
)

// CooldownCache holds cooldowns for commands
type CooldownCache struct {
	c *ttlcache.Cache
}

func newCooldownCache() *CooldownCache {
	cache := ttlcache.NewCache()
	cache.SkipTTLExtensionOnHit(true)

	return &CooldownCache{c: cache}
}

// Set sets a cooldown for a command
func (c *CooldownCache) Set(cmdName string, userID discord.UserID, channelID discord.ChannelID, cooldown time.Duration) {
	// if the command's cooldown is 0, return
	if cooldown == 0 {
		return
	}

	c.c.SetWithTTL(cmdName+userID.String()+channelID.String(), true, cooldown)
	return
}

// Get returns true if the command is on cooldown
func (c *CooldownCache) Get(cmdName string, userID discord.UserID, channelID discord.ChannelID) bool {
	if _, e := c.c.Get(cmdName + userID.String() + channelID.String()); e == nil {
		return true
	}

	return false
}
