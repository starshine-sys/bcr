package bcr

import (
	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/diamondburned/arikawa/v3/utils/json/option"
)

// IsThread returns true if the given channel is a thread channel.
func IsThread(ch *discord.Channel) bool {
	return ch.Type == discord.GuildNewsThread ||
		ch.Type == discord.GuildPublicThread ||
		ch.Type == discord.GuildPrivateThread
}

// ModalResponse creates a modal interaction response.
func ModalResponse(
	id discord.InteractionID, token string,
	s *state.State,
	customID, title string,
	components ...discord.InteractiveComponent,
) error {
	ccs := make(discord.ContainerComponents, 0, len(components))
	for _, c := range components {
		ccs = append(ccs, &discord.ActionRowComponent{c})
	}

	return s.RespondInteraction(id, token, api.InteractionResponse{
		Type: api.ModalResponse,
		Data: &api.InteractionResponseData{
			CustomID:   option.NewNullableString(customID),
			Title:      option.NewNullableString(title),
			Components: &ccs,
		},
	})
}
