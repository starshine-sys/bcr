package bcr

import (
	"fmt"
	"log"

	"github.com/diamondburned/arikawa/v2/gateway"
)

// MessageCreate gets called on new messages
// - makes sure the router has a bot user
// - checks if the message matches a prefix
// - runs commands
func (r *Router) MessageCreate(m *gateway.MessageCreateEvent) {
	// set the bot user if not done already
	if r.Bot == nil {
		r.mustSetBotUser()
		r.Prefixes = append(r.Prefixes, fmt.Sprintf("<@%v>", r.Bot.ID), fmt.Sprintf("<@!%v>", r.Bot.ID))
	}

	// if the author is a bot, return
	if m.Author.Bot {
		return
	}

	// if the message does not start with any of the bot's prefixes (including mentions), return
	if !r.MatchPrefix(m.Content) {
		return
	}

	// get the context
	ctx, err := r.NewContext(m.Message)
	if err != nil {
		log.Println("Error getting context:", err)
		return
	}

	err = r.Execute(ctx)
	if err != nil {
		log.Println("Error executing commands:", err)
		return
	}
}

// mustSetBotUser sets the bot user in the router, panicking if it fails.
// This is intended to be used in MessageCreate to simplify error handling
func (r *Router) mustSetBotUser() {
	err := r.SetBotUser()
	if err != nil {
		panic(err)
	}
}

// SetBotUser sets the router's bot user, returning any errors
func (r *Router) SetBotUser() error {
	me, err := r.Session.Me()
	if err != nil {
		return err
	}

	r.Bot = me
	return nil
}
