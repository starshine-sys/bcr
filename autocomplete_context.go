package bcr

import (
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
)

type AutocompleteContext struct {
	*Context

	Command []string
	Options []discord.AutocompleteOption
	Data    *discord.AutocompleteInteraction
}

// NewAutocompleteContext creates a new autocomplete context.
func (r *Router) NewAutocompleteContext(ic *gateway.InteractionCreateEvent) (ctx *AutocompleteContext, err error) {
	data, ok := ic.Data.(*discord.AutocompleteInteraction)
	if !ok {
		return nil, ErrNotCommand
	}

	root, err := r.NewRootContext(ic)
	if err != nil {
		return nil, err
	}

	ctx = &AutocompleteContext{
		Context: root,
		Data:    data,
		Command: []string{data.Name},
		Options: data.Options,
	}

	return ctx, nil
}
