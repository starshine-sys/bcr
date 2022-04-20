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
		return r.executeAutocomplete(ic)
	case *discord.SelectInteraction:
		return r.executeSelect(ic)
	case *discord.ButtonInteraction:
		return r.executeButton(ic)
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
		hn, ok = r.findWildcardModal(ctx.CustomID)
		if !ok {
			return nil
		}
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
	hn, msgScoped, ok := findHandler(r.buttons, ctx.CustomID, ctx.Event.Message.ID)
	r.componentsMu.RUnlock()
	if !ok {
		return nil
	}

	if !msgScoped {
		msgID = discord.NullMessageID
	}

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
		delete(r.buttons, componentKey{ctx.CustomID, msgID})
		r.componentsMu.Unlock()
	}

	return hn.handler(ctx)
}

func (r *Router) executeSelect(ic *gateway.InteractionCreateEvent) error {
	ctx, err := r.NewSelectContext(ic)
	if err != nil {
		return err
	}

	// check for message-scoped select
	msgID := ctx.Event.Message.ID

	r.componentsMu.RLock()
	hn, msgScoped, ok := findHandler(r.selects, ctx.CustomID, ctx.Event.Message.ID)
	r.componentsMu.RUnlock()
	if !ok {
		return nil
	}

	if !msgScoped {
		msgID = discord.NullMessageID
	}

	err = hn.doCheck(ctx)
	if err != nil {
		if hn.once {
			return nil
		}

		if v, ok := err.(CheckError[*SelectContext]); ok {
			s, e := v.CheckError(ctx)
			return ctx.Reply(s, e...)
		}
		return ctx.Reply(
			fmt.Sprintf("Unable to handle select:\n%v", err),
		)
	}

	if hn.once {
		r.componentsMu.Lock()
		delete(r.selects, componentKey{ctx.CustomID, msgID})
		r.componentsMu.Unlock()
	}

	return hn.handler(ctx)
}

func (r *Router) executeAutocomplete(ic *gateway.InteractionCreateEvent) error {
	ctx, err := r.NewAutocompleteContext(ic)
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

	hn, ok := r.autocompletes[strings.Join(ctx.Command, "/")]
	if !ok {
		return ErrUnknownCommand
	}

	return hn.handler(ctx)
}

func (r *Router) findWildcardModal(id discord.ComponentID) (*handler[*ModalContext], bool) {
	for k, v := range r.modals {
		if v.prefixWildcard {
			if strings.HasPrefix(string(id), string(k)) {
				return v, true
			}
		} else if v.suffixWildcard {
			if strings.HasSuffix(string(id), string(k)) {
				return v, true
			}
		}
	}
	return nil, false
}

// findHandler finds a handler in the given map.
// It assumes m is already locked for reading.
func findHandler[T HasContext](
	m map[componentKey]*handler[T],
	customID discord.ComponentID,
	msgID discord.MessageID,
) (h *handler[T], msgScoped bool, ok bool) {

	hn, ok := m[componentKey{customID, msgID}]
	if ok {
		return hn, true, true
	}

	hn, ok = m[componentKey{customID, discord.NullMessageID}]
	if ok {
		return hn, false, true
	}

	// check wildcards
	for k, v := range m {
		// message-scoped components cannot be wildcards
		if k.msgID.IsValid() {
			continue
		}

		if v.prefixWildcard {
			if strings.HasPrefix(string(customID), string(k.id)) {
				return v, false, true
			}
		} else if v.suffixWildcard {
			if strings.HasSuffix(string(customID), string(k.id)) {
				return v, false, true
			}
		}
	}
	return nil, false, false
}
