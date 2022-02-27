package bcr

import (
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
)

type SelectContext struct {
	*Context

	CustomID discord.ComponentID
	Values   []string
	Data     *discord.SelectInteraction
}

// NewSelectContext creates a new select context.
func (r *Router) NewSelectContext(ic *gateway.InteractionCreateEvent) (ctx *SelectContext, err error) {
	data, ok := ic.Data.(*discord.SelectInteraction)
	if !ok {
		return nil, ErrNotModal
	}

	root, err := r.NewRootContext(ic)
	if err != nil {
		return nil, err
	}

	ctx = &SelectContext{
		Context:  root,
		Data:     data,
		CustomID: data.CustomID,
		Values:   data.Values,
	}

	return ctx, nil
}
