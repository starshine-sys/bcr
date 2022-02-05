package bcr

import (
	"emperror.dev/errors"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
)

func (r *Router) Execute(ic *gateway.InteractionCreateEvent) error {
	switch ic.Data.(type) {
	case *discord.CommandInteraction:
		return r.executeCommand(ic)
	case *discord.AutocompleteInteraction:

	case *discord.SelectInteraction:

	case *discord.ButtonInteraction:
	default:
		return errors.Sentinel("unhandled interaction type")
	}
}

func (r *Router) executeCommand(ic *gateway.InteractionCreateEvent) error {
	ctx, err := r.NewCommandContext(ic)
	if err != nil {
		return err
	}

	ctx.Command = append(ctx.Command, ctx.Data.Name)

	if len(ctx.Options) > 0 {

	}
}
