// Package bot provides a basic embeddable Bot struct for more easily handling commands
package bot

import (
	"sort"

	"github.com/diamondburned/arikawa/v2/discord"
	"github.com/starshine-sys/bcr"
)

// Bot is the main bot struct
type Bot struct {
	Router *bcr.Router

	Modules []Module
}

// Module is a single module/category of commands
type Module interface {
	String() string
	Commands() []*bcr.Command
}

// New creates a new instance of Bot.
// The token will be prefixed with `Bot ` automatically.
func New(token string) (*Bot, error) {
	r, err := bcr.NewWithState(token, []discord.UserID{}, []string{})
	if err != nil {
		return nil, err
	}
	return NewWithRouter(r), nil
}

// NewWithRouter creates a new bot with the given router
func NewWithRouter(r *bcr.Router) *Bot {
	return &Bot{
		Router: r,
	}
}

// Prefix is a helper function to set the bot's router's prefixes
func (bot *Bot) Prefix(prefixes ...string) {
	bot.Router.Prefixes = prefixes
}

// Owner is a helper function to set the bot's owner(s)
func (bot *Bot) Owner(owners ...discord.UserID) {
	o := make([]string, 0)
	for _, owner := range owners {
		o = append(o, owner.String())
	}
	bot.Router.BotOwners = o
}

// Add adds a module to the bot
func (bot *Bot) Add(f func(*Bot) (string, []*bcr.Command)) {
	m, c := f(bot)

	// sort the list of commands
	sort.Sort(bcr.Commands(c))

	// add the module
	bot.Modules = append(bot.Modules, &botModule{
		name:     m,
		commands: c,
	})
}

// Start wraps around Router.State.Open()
func (bot *Bot) Start() error {
	return bot.Router.State.Open()
}
