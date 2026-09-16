package commands

import (
	"github.com/bwmarrin/discordgo"
)

func NewRankCommand() Command {
	return Command{
		Definition: &discordgo.ApplicationCommand{
			Name:        "rank",
			Description: "Display a user's rank in the server",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionUser,
					Name:        "user", // Matches ctx.GetUser("user")
					Description: "Check another user's rank (optional)",
					Required:    false,
				},
			},
		},
		Action: HandleRank,
	}
}