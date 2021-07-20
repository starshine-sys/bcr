package bcr

import (
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/diamondburned/arikawa/v3/discord"
)

// ErrNoEmbeds is returned if PagedEmbed() is called without any embeds
var ErrNoEmbeds = errors.New("PagedEmbed: no embeds")

// PagedEmbed sends a slice of embeds, and attaches reaction handlers to flip through them.
// if extendedReactions is true, also add delete, first page, and last page reactions.
func (ctx *Context) PagedEmbed(embeds []discord.Embed, extendedReactions bool) (msg *discord.Message, err error) {
	msg, _, err = ctx.PagedEmbedTimeout(embeds, extendedReactions, 15*time.Minute)
	return
}

// PagedEmbedTimeout creates a paged embed (see PagedEmbed) that times out after the given time.
// It also returns a timer that can be used to cancel the attached reaction-clearing timer.
func (ctx *Context) PagedEmbedTimeout(embeds []discord.Embed, extendedReactions bool, timeout time.Duration) (msg *discord.Message, timer *time.Timer, err error) {
	// if there's no embeds, return
	if len(embeds) == 0 {
		return nil, nil, ErrNoEmbeds
	}

	// set additional parameters, used for pagination
	ctx.AdditionalParams["page"] = 0
	ctx.AdditionalParams["embeds"] = embeds

	// send the first embed
	msg, err = ctx.Send("", embeds[0])
	if err != nil {
		return
	}

	// this doesn't seem to work rn for some reason
	// timer = time.AfterFunc(timeout, func() {
	// 	ctx.State.DeleteAllReactions(msg.ChannelID, msg.ID)
	// })

	// add :x: handler
	ctx.AddReactionHandlerWithTimeout(msg.ID, ctx.Author.ID, "❌", true, false, timeout, func(*Context) {
		// timer.Stop()
		err = ctx.State.DeleteMessage(ctx.Channel.ID, msg.ID)
		if err != nil {
			ctx.Router.Logger.Error("deleting message: %v", err)
		}
	})

	// if there's only one embed, that's it! no pager emoji needed
	if len(embeds) == 1 {
		return
	}

	// react with all required emoji--afawk there's no more concise way to do this
	if extendedReactions {
		if err = ctx.State.React(ctx.Channel.ID, msg.ID, "❌"); err != nil {
			return
		}
		if err = ctx.State.React(ctx.Channel.ID, msg.ID, "⏪"); err != nil {
			return
		}
	}
	if err = ctx.State.React(ctx.Channel.ID, msg.ID, "⬅️"); err != nil {
		return
	}
	if err = ctx.State.React(ctx.Channel.ID, msg.ID, "➡️"); err != nil {
		return
	}
	if extendedReactions {
		if err = ctx.State.React(ctx.Channel.ID, msg.ID, "⏩"); err != nil {
			return
		}
	}

	// add handlers for the reactions
	ctx.AddReactionHandlerWithTimeoutRemove(msg.ID, ctx.Author.ID, "⬅️", false, true, timeout, func(ctx *Context) {
		page := ctx.AdditionalParams["page"].(int)
		embeds := ctx.AdditionalParams["embeds"].([]discord.Embed)

		if page == 0 {
			ctx.State.EditEmbeds(ctx.Channel.ID, msg.ID, embeds[len(embeds)-1])
			ctx.AdditionalParams["page"] = len(embeds) - 1
			return
		}
		ctx.State.EditEmbeds(ctx.Channel.ID, msg.ID, embeds[page-1])
		ctx.AdditionalParams["page"] = page - 1
	})

	ctx.AddReactionHandlerWithTimeoutRemove(msg.ID, ctx.Author.ID, "➡️", false, true, timeout, func(ctx *Context) {
		page := ctx.AdditionalParams["page"].(int)
		embeds := ctx.AdditionalParams["embeds"].([]discord.Embed)

		if page >= len(embeds)-1 {
			ctx.State.EditEmbeds(ctx.Channel.ID, msg.ID, embeds[0])
			ctx.AdditionalParams["page"] = 0
			return
		}
		ctx.State.EditEmbeds(ctx.Channel.ID, msg.ID, embeds[page+1])
		ctx.AdditionalParams["page"] = page + 1
	})

	if extendedReactions {
		ctx.AddReactionHandlerWithTimeoutRemove(msg.ID, ctx.Author.ID, "⏪", false, true, timeout, func(ctx *Context) {
			embeds := ctx.AdditionalParams["embeds"].([]discord.Embed)

			ctx.State.EditEmbeds(ctx.Channel.ID, msg.ID, embeds[0])
			ctx.AdditionalParams["page"] = 0
		})

		ctx.AddReactionHandlerWithTimeoutRemove(msg.ID, ctx.Author.ID, "⏩", false, true, timeout, func(ctx *Context) {
			embeds := ctx.AdditionalParams["embeds"].([]discord.Embed)

			ctx.State.EditEmbeds(ctx.Channel.ID, msg.ID, embeds[len(embeds)-1])
			ctx.AdditionalParams["page"] = len(embeds) - 1
		})
	}
	return
}

// FieldPaginator paginates embed fields, for use in ctx.PagedEmbed
func FieldPaginator(title, description string, colour discord.Color, fields []discord.EmbedField, perPage int) []discord.Embed {
	var (
		embeds []discord.Embed
		count  int

		pages = 1
		buf   = discord.Embed{
			Title:       title,
			Description: description,
			Color:       colour,
			Footer: &discord.EmbedFooter{
				Text: fmt.Sprintf("Page 1/%v", math.Ceil(float64(len(fields))/float64(perPage))),
			},
		}
	)

	for _, field := range fields {
		if count >= perPage {
			embeds = append(embeds, buf)
			buf = discord.Embed{
				Title:       title,
				Description: description,
				Color:       colour,
				Footer: &discord.EmbedFooter{
					Text: fmt.Sprintf("Page %v/%v", pages+1, math.Ceil(float64(len(fields))/float64(perPage))),
				},
			}
			count = 0
			pages++
		}
		buf.Fields = append(buf.Fields, field)
		count++
	}

	embeds = append(embeds, buf)

	return embeds
}

// StringPaginator paginates strings, for use in ctx.PagedEmbed
func StringPaginator(title string, colour discord.Color, slice []string, perPage int) []discord.Embed {
	var (
		embeds []discord.Embed
		count  int

		pages = 1
		buf   = discord.Embed{
			Title: title,
			Color: colour,
			Footer: &discord.EmbedFooter{
				Text: fmt.Sprintf("Page 1/%v", math.Ceil(float64(len(slice))/float64(perPage))),
			},
		}
	)

	for _, s := range slice {
		if count >= perPage {
			embeds = append(embeds, buf)
			buf = discord.Embed{
				Title: title,
				Color: colour,
				Footer: &discord.EmbedFooter{
					Text: fmt.Sprintf("Page %v/%v", pages+1, math.Ceil(float64(len(slice))/float64(perPage))),
				},
			}
			count = 0
			pages++
		}
		buf.Description += s
		count++
	}

	embeds = append(embeds, buf)

	return embeds
}
