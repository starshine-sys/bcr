package bcr

import (
	"io"

	"github.com/diamondburned/arikawa/v2/api"
	"github.com/diamondburned/arikawa/v2/discord"
	"github.com/diamondburned/arikawa/v2/utils/sendpart"
)

// MessageSend is a helper struct for sending messages.
// By default, it will send a message to the current channel, and check permissions (unless the target channel is the current channel and is a DM channel).
// These can be overridden with the Channel(id) and TogglePermCheck() methods.
// Alternatively, you can get the base SendMessageData struct and use that manually.
type MessageSend struct {
	Data api.SendMessageData

	user    discord.UserID
	channel discord.ChannelID

	checkPerms bool
	ctx        *Context
}

// NewMessage creates a new MessageSend object.
// Only the *first* channel argument is used, if omitted it's set to the current channel.
func (ctx *Context) NewMessage(channel ...discord.ChannelID) *MessageSend {
	ch := ctx.Channel.ID
	if len(channel) > 0 {
		ch = channel[0]
	}

	return &MessageSend{
		Data:       api.SendMessageData{},
		ctx:        ctx,
		checkPerms: true,
		channel:    ch,
	}
}

// NewDM creates a new MessageSend object for the given user.
// If the user has closed DMs, this will not error until the Send() call.
func (ctx *Context) NewDM(user discord.UserID) *MessageSend {
	return &MessageSend{
		Data:       api.SendMessageData{},
		ctx:        ctx,
		checkPerms: false,
		user:       user,
	}
}

// Channel sets the channel to send the message to
func (m *MessageSend) Channel(c discord.ChannelID) *MessageSend {
	m.channel = c
	return m
}

// Content sets the message content
func (m *MessageSend) Content(c string) *MessageSend {
	m.Data.Content = c
	return m
}

// Embed sets the message embed
func (m *MessageSend) Embed(e *discord.Embed) *MessageSend {
	m.Data.Embed = e
	return m
}

// AddFile adds a file to the message
func (m *MessageSend) AddFile(name string, reader io.Reader) *MessageSend {
	m.Data.Files = append(m.Data.Files, sendpart.File{
		Name:   name,
		Reader: reader,
	})
	return m
}

// BlockMentions blocks all mentions from this message
func (m *MessageSend) BlockMentions() *MessageSend {
	m.Data.AllowedMentions = &api.AllowedMentions{Parse: nil}
	return m
}

// AllowedMentions sets the message's allowed mentions
func (m *MessageSend) AllowedMentions(a *api.AllowedMentions) *MessageSend {
	m.Data.AllowedMentions = a
	return m
}

// TogglePermCheck toggles whether or not to check permissions for the destination channel
func (m *MessageSend) TogglePermCheck() *MessageSend {
	if m.checkPerms {
		m.checkPerms = false
	} else {
		m.checkPerms = true
	}
	return m
}

// Reference sets the message this message will reply to
func (m *MessageSend) Reference(id discord.MessageID) *MessageSend {
	m.Data.Reference = &discord.MessageReference{
		MessageID: id,
	}
	return m
}

// Send sends the message
func (m *MessageSend) Send() (msg *discord.Message, err error) {
	// if it's a user, send a DM
	if m.user != 0 {
		return m.sendDM()
	}
	// otherwise send the message normally
	return m.send()
}

func (m *MessageSend) send() (msg *discord.Message, err error) {
	if m.checkPerms {
		if !m.ctx.checkBotSendPerms(m.channel, m.Data.Embed != nil) {
			return nil, ErrBotMissingPermissions
		}
	}

	return m.ctx.State.SendMessageComplex(m.channel, m.Data)
}

func (m *MessageSend) sendDM() (msg *discord.Message, err error) {
	ch, err := m.ctx.State.CreatePrivateChannel(m.user)
	if err != nil {
		return nil, err
	}

	return m.ctx.State.SendMessageComplex(ch.ID, m.Data)
}
