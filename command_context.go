package bcr

import (
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
)

type CommandContext struct {
	*Context

	Command []string
	Options []discord.CommandInteractionOption
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
