package bcr

import (
	"emperror.dev/errors"
	"github.com/diamondburned/arikawa/v3/discord"
)

const ErrNotCommand = errors.Sentinel("not a command interaction")

// Checker is typically implemented by check errors.
type Checker interface {
	CheckResponse() (content string, embeds []discord.Embed)
}
