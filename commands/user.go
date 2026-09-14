package commands

import (
	"github.com/bwmarrin/discordgo"
)

func NewUserCommandStructure() Command {
	return Command{
	Definition: &discordgo.ApplicationCommand{
		Name:        "user",
		Description: "General user commands",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "info",
				Description: "Displays user info",
				Options:	[]*discordgo.ApplicationCommandOption{
					{
						Type:			discordgo.ApplicationCommandOptionUser,
						Name:			"target",
						Description:	"choose target to grab information",
						Required:		false,
					},
				},
			},
		},
	},
	SubCommands: map[string]SubCommandHandler{
		"info": handleInfo,
	},
}

}