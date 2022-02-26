package bcr

import (
	"fmt"
	"strings"

	"emperror.dev/errors"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
)

func (r *Router) Execute(ic *gateway.InteractionCreateEvent) (err error) {
	defer func() {
		if rev := recover(); rev != nil {
			switch v := rev.(type) {
			case error:
				err = v
			case string:
				err = errors.New(v)
			default:
				err = fmt.Errorf("%v", v)
			}
		}
	}()

	switch ic.Data.(type) {
	case *discord.CommandInteraction:
		return r.executeCommand(ic)
	case *discord.AutocompleteInteraction:

	case *discord.SelectInteraction:

	case *discord.ButtonInteraction:

	case *discord.ModalInteraction:
		return r.executeModal(ic)
	}

	return errors.Sentinel("unhandled interaction type")
}

func (r *Router) executeCommand(ic *gateway.InteractionCreateEvent) error {
	ctx, err := r.NewCommandContext(ic)
	if err != nil {
		return err
	}

	options := ctx.Options
	if len(ctx.Options) > 0 {
		if ctx.Options[0].Type == discord.SubcommandOptionType {
			ctx.Command = append(ctx.Command, ctx.Options[0].Name)
			options = ctx.Options[0].Options
		} else if ctx.Options[0].Type == discord.SubcommandGroupOptionType {
			ctx.Command = append(ctx.Command, ctx.Options[0].Name, ctx.Options[0].Options[0].Name)
			options = ctx.Options[0].Options[0].Options
		}
	}

	ctx.Options = options

	hn, ok := r.commands[strings.Join(ctx.Command, "/")]
	if !ok {
		return ErrUnknownCommand
	}

	err = hn.doCheck(ctx)
	if err != nil {
		if v, ok := err.(CheckError[*CommandContext]); ok {
			s, e := v.CheckError(ctx)
			return ctx.Reply(s, e...)
		}
		return ctx.Reply(
			fmt.Sprintf("You are not allowed to execute this command:\n%v", err),
		)
	}

	return hn.handler(ctx)
}

func (r *Router) executeModal(ic *gateway.InteractionCreateEvent) error {
	ctx, err := r.NewModalContext(ic)
	if err != nil {
		return err
	}

	hn, ok := r.modals[ctx.CustomID]
	if !ok {
		return nil
	}

	err = hn.doCheck(ctx)
	if err != nil {
		if v, ok := err.(CheckError[*ModalContext]); ok {
			s, e := v.CheckError(ctx)
			return ctx.Reply(s, e...)
		}
		return ctx.Reply(
			fmt.Sprintf("Unable to submit this modal:\n%v", err),
		)
	}

	return hn.handler(ctx)
}

func (r *Router) executeButton(ic *gateway.InteractionCreateEvent) error {
	ctx, err := r.NewButtonContext(ic)
	if err != nil {
		return err
	}

	// check for message-scoped button
	msgID := ctx.Event.Message.ID

	r.componentsMu.RLock()
	hn, ok := r.buttons[componentKey{ctx.ID, ctx.Event.Message.ID}]
	if !ok {
		hn, ok = r.buttons[componentKey{ctx.ID, discord.NullMessageID}]
		if !ok {
			r.componentsMu.RUnlock()

			return nil
		}
		msgID = discord.NullMessageID
	}
	r.componentsMu.RUnlock()

	err = hn.doCheck(ctx)
	if err != nil {
		if hn.once {
			return nil
		}

		if v, ok := err.(CheckError[*ButtonContext]); ok {
			s, e := v.CheckError(ctx)
			return ctx.Reply(s, e...)
		}
		return ctx.Reply(
			fmt.Sprintf("Unable to handle button:\n%v", err),
		)
	}

	if hn.once {
		r.componentsMu.Lock()
		delete(r.buttons, componentKey{ctx.ID, msgID})
		r.componentsMu.Unlock()
	}

	return hn.handler(ctx)
}
