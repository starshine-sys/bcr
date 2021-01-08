package bcr

import (
	"errors"
	"regexp"
	"strings"

	"github.com/diamondburned/arikawa/v2/discord"
)

var (
	channelMentionRegex = regexp.MustCompile("<#(\\d+)>")
	userMentionRegex    = regexp.MustCompile("<@!?(\\d+)>")
	roleMentionRegex    = regexp.MustCompile("<@&(\\d+)>")

	idRegex = regexp.MustCompile("^\\d+$")
)

// Errors related to parsing
var (
	ErrInvalidMention  = errors.New("invalid mention")
	ErrChannelNotFound = errors.New("channel not found")
	ErrMemberNotFound  = errors.New("member not found")
	ErrUserNotFound    = errors.New("user not found")
	ErrRoleNotFound    = errors.New("role not found")
)

// ParseChannel parses a channel mention/id/name
func (ctx *Context) ParseChannel(s string) (c *discord.Channel, err error) {
	// check if it's an ID
	if idRegex.MatchString(s) {
		sf, err := discord.ParseSnowflake(s)
		if err != nil {
			return nil, err
		}
		return ctx.Session.Channel(discord.ChannelID(sf))
	}

	// check if it's a mention
	if channelMentionRegex.MatchString(s) {
		matches := channelMentionRegex.FindStringSubmatch(s)
		if len(matches) < 2 {
			return nil, ErrInvalidMention
		}
		sf, err := discord.ParseSnowflake(matches[1])
		if err != nil {
			return nil, err
		}
		return ctx.Session.Channel(discord.ChannelID(sf))
	}

	// otherwise, fall back to names
	channels, err := ctx.Session.Channels(ctx.Message.GuildID)
	if err != nil {
		return nil, err
	}

	for _, ch := range channels {
		if strings.ToLower(s) == strings.ToLower(ch.Name) {
			return &ch, nil
		}
	}
	return nil, ErrChannelNotFound
}

// ParseMember parses a member mention/id/name
func (ctx *Context) ParseMember(s string) (c *discord.Member, err error) {
	// check if it's an ID
	if idRegex.MatchString(s) {
		sf, err := discord.ParseSnowflake(s)
		if err != nil {
			return nil, err
		}
		return ctx.Session.Member(ctx.Message.GuildID, discord.UserID(sf))
	}

	// check if it's a mention
	if userMentionRegex.MatchString(s) {
		matches := userMentionRegex.FindStringSubmatch(s)
		if len(matches) < 2 {
			return nil, ErrInvalidMention
		}
		sf, err := discord.ParseSnowflake(matches[1])
		if err != nil {
			return nil, err
		}
		return ctx.Session.Member(ctx.Message.GuildID, discord.UserID(sf))
	}

	// otherwise, fall back to names
	members, err := ctx.Session.Members(ctx.Message.GuildID)
	if err != nil {
		return nil, err
	}

	for _, m := range members {
		// check full name
		if strings.ToLower(m.User.Username)+"#"+m.User.Discriminator == strings.ToLower(s) {
			return &m, nil
		}

		// check just username
		if strings.ToLower(m.User.Username) == strings.ToLower(s) {
			return &m, nil
		}

		// check nickname
		if strings.ToLower(m.Nick) == strings.ToLower(s) {
			return &m, nil
		}
	}
	return nil, ErrMemberNotFound
}

// ParseRole parses a role mention/id/name
func (ctx *Context) ParseRole(s string) (c *discord.Role, err error) {
	// check if it's an ID
	if idRegex.MatchString(s) {
		sf, err := discord.ParseSnowflake(s)
		if err != nil {
			return nil, err
		}
		return ctx.Session.Role(ctx.Message.GuildID, discord.RoleID(sf))
	}

	// check if it's a mention
	if roleMentionRegex.MatchString(s) {
		matches := roleMentionRegex.FindStringSubmatch(s)
		if len(matches) < 2 {
			return nil, ErrInvalidMention
		}
		sf, err := discord.ParseSnowflake(matches[1])
		if err != nil {
			return nil, err
		}
		return ctx.Session.Role(ctx.Message.GuildID, discord.RoleID(sf))
	}

	// otherwise, fall back to names
	roles, err := ctx.Session.Roles(ctx.Message.GuildID)
	if err != nil {
		return nil, err
	}

	for _, r := range roles {
		if strings.ToLower(s) == strings.ToLower(r.Name) {
			return &r, nil
		}
	}
	return nil, ErrChannelNotFound
}

// ParseUser finds a user by mention or ID
func (ctx *Context) ParseUser(s string) (u *discord.User, err error) {
	if idRegex.MatchString(s) {
		sf, err := discord.ParseSnowflake(s)
		if err != nil {
			return nil, err
		}
		return ctx.Session.User(discord.UserID(sf))
	}

	if userMentionRegex.MatchString(s) {
		matches := userMentionRegex.FindStringSubmatch(s)
		if len(matches) < 2 {
			return nil, ErrInvalidMention
		}
		sf, err := discord.ParseSnowflake(matches[1])
		if err != nil {
			return nil, err
		}
		return ctx.Session.User(discord.UserID(sf))
	}

	return nil, ErrUserNotFound
}
