package bcr

import (
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
)

type ModalContext struct {
	*Context

	CustomID   discord.ComponentID
	Components []discord.InteractiveComponent
	TextInputs []discord.TextInputComponent
	Data       *discord.ModalInteraction
}

// NewModalContext creates a new modal context.
func (r *Router) NewModalContext(ic *gateway.InteractionCreateEvent) (ctx *ModalContext, err error) {
	data, ok := ic.Data.(*discord.ModalInteraction)
	if !ok {
		return nil, ErrNotModal
	}

	root, err := r.NewRootContext(ic)
	if err != nil {
		return nil, err
	}

	ctx = &ModalContext{
		Context:  root,
		Data:     data,
		CustomID: data.CustomID,
	}

	// extract text input components
	for _, cc := range data.Components {
		v, ok := cc.(*discord.ActionRowComponent)
		if ok {
			ctx.Components = append(ctx.Components, *v...)

			for _, c := range *v {
				switch v := c.(type) {
				case *discord.TextInputComponent:
					ctx.TextInputs = append(ctx.TextInputs, *v)
				}
			}
		}
	}

	return ctx, nil
}
