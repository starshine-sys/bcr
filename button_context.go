package bcr

import (
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
)

type ButtonContext struct {
	*Context

	Data     *discord.ButtonInteraction
	CustomID discord.ComponentID
}

// ButtonContext creates a new command context.
func (r *Router) NewButtonContext(ic *gateway.InteractionCreateEvent) (ctx *ButtonContext, err error) {
	data, ok := ic.Data.(*discord.ButtonInteraction)
	if !ok {
		return nil, ErrNotButton
	}

	root, err := r.NewRootContext(ic)
	if err != nil {
		return nil, err
	}

	ctx = &ButtonContext{
		Context:  root,
		Data:     data,
		CustomID: data.CustomID,
	}

	return ctx, nil
}
