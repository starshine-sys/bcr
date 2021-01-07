package bcr

import (
	"errors"
	"log"

	"github.com/diamondburned/arikawa/v2/discord"
)

// ErrNoEmbeds is returned if PagedEmbed() is called without any embeds
var ErrNoEmbeds = errors.New("PagedEmbed: no embeds")

// PagedEmbed sends a slice of embeds, and attaches reaction handlers to flip through them.
func (ctx *Context) PagedEmbed(embeds []discord.Embed) (msg *discord.Message, err error) {
	// if there's no embeds, return
	if len(embeds) == 0 {
		return nil, ErrNoEmbeds
	}

	// set additional parameters, used for pagination
	ctx.AdditionalParams["page"] = 0
	ctx.AdditionalParams["embeds"] = embeds

	// send the first embed
	msg, err = ctx.Session.SendEmbed(ctx.Channel.ID, embeds[0])
	if err != nil {
		return
	}

	// add :x: handler
	ctx.AddReactionHandler(msg.ID, ctx.Author.ID, "❌", true, false, func(*Context) {
		err = ctx.Session.DeleteMessage(ctx.Channel.ID, msg.ID)
		if err != nil {
			log.Printf("Error deleting message %v: %v", msg.ID, err)
		}
	})

	// if there's only one embed, that's it! no pager emoji needed
	if len(embeds) == 1 {
		return
	}

	// react with all required emoji--afawk there's no more concise way to do this
	if err = ctx.Session.React(ctx.Channel.ID, msg.ID, "❌"); err != nil {
		return
	}
	if err = ctx.Session.React(ctx.Channel.ID, msg.ID, "⏪"); err != nil {
		return
	}
	if err = ctx.Session.React(ctx.Channel.ID, msg.ID, "⬅️"); err != nil {
		return
	}
	if err = ctx.Session.React(ctx.Channel.ID, msg.ID, "➡️"); err != nil {
		return
	}
	if err = ctx.Session.React(ctx.Channel.ID, msg.ID, "⏩"); err != nil {
		return
	}

	// add handlers for the reactions
	ctx.AddReactionHandler(msg.ID, ctx.Author.ID, "⬅️", false, true, func(ctx *Context) {
		page := ctx.AdditionalParams["page"].(int)
		embeds := ctx.AdditionalParams["embeds"].([]discord.Embed)

		if page == 0 {
			return
		}
		ctx.Session.EditEmbed(ctx.Channel.ID, msg.ID, embeds[page-1])
		ctx.AdditionalParams["page"] = page - 1
	})

	ctx.AddReactionHandler(msg.ID, ctx.Author.ID, "➡️", false, true, func(ctx *Context) {
		page := ctx.AdditionalParams["page"].(int)
		embeds := ctx.AdditionalParams["embeds"].([]discord.Embed)

		if page >= len(embeds)-1 {
			return
		}
		ctx.Session.EditEmbed(ctx.Channel.ID, msg.ID, embeds[page+1])
		ctx.AdditionalParams["page"] = page + 1
	})

	ctx.AddReactionHandler(msg.ID, ctx.Author.ID, "⏪", false, true, func(ctx *Context) {
		embeds := ctx.AdditionalParams["embeds"].([]discord.Embed)

		ctx.Session.EditEmbed(ctx.Channel.ID, msg.ID, embeds[0])
		ctx.AdditionalParams["page"] = 0
	})

	ctx.AddReactionHandler(msg.ID, ctx.Author.ID, "⏩", false, true, func(ctx *Context) {
		embeds := ctx.AdditionalParams["embeds"].([]discord.Embed)

		ctx.Session.EditEmbed(ctx.Channel.ID, msg.ID, embeds[len(embeds)-1])
		ctx.AdditionalParams["page"] = len(embeds) - 1
	})
	return
}
