package commands

import (
	"github.com/bwmarrin/discordgo"
	//"github.com/k3ptok/Basic/Discord/Bot/cmdctx"
)

//define all public utility commands
func NewPingCommand() Command {
	return Command{
		Definition: &discordgo.ApplicationCommand{
			Name:        "ping",
			Description: "Check bot latency",
		},
		Action: HandlePing, 
	}
}

func NewTagCommand() Command {
	return Command{
		Definition: &discordgo.ApplicationCommand{
			Name:        "tag",
			Description: "Pull up a helpful resource",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "name",
					Description: "The name of the tag",
					Required:    true,
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{Name: "Rules", Value: "rules"},
					},
				},
			},
		},
		Action: HandleTag,
	}
}