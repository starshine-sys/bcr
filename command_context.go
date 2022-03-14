package bcr

import (
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
)

type CommandContext struct {
	*Context

	Command []string
	Options discord.CommandInteractionOptions
	Data    *discord.CommandInteraction
}

// NewCommandContext creates a new command context.
func (r *Router) NewCommandContext(ic *gateway.InteractionCreateEvent) (ctx *CommandContext, err error) {
	data, ok := ic.Data.(*discord.CommandInteraction)
	if !ok {
		return nil, ErrNotCommand
	}

	root, err := r.NewRootContext(ic)
	if err != nil {
		return nil, err
	}

	ctx = &CommandContext{
		Context: root,
		Data:    data,
		Command: []string{data.Name},
		Options: data.Options,
	}

	return ctx, nil
}

func (ctx *CommandContext) FirstUser() discord.User {
	for _, u := range ctx.Data.Resolved.Users {
		return u
	}
	return discord.User{}
}

func (ctx *CommandContext) FirstMessage() discord.Message {
	for _, u := range ctx.Data.Resolved.Messages {
		return u
	}
	return discord.Message{}
}

func (ctx *CommandContext) Option(name string) discord.CommandInteractionOption {
	return ctx.Options.Find(name)
}
