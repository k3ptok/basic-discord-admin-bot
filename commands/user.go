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
				
			},
		},
	},
	SubCommands: map[string]SubCommandHandler{
		"info": handleInfo,
	},
}

}