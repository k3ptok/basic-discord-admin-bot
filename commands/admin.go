package commands

import (
	"github.com/bwmarrin/discordgo"
)

var adminPerms int64 = discordgo.PermissionAdministrator

var Admin = Command{
	Definition:	&discordgo.ApplicationCommand{
		Name:						"admin",
		Description:				"Administration Commands",
		DefaultMemberPermissions: 	&adminPerms,
		Options:	[]*discordgo.ApplicationCommandOption{
			{
			Type:			discordgo.ApplicationCommandOptionSubCommand,
			Name:			"kick",
			Description:	"Kick a user from the server",
			Options:	[]*discordgo.ApplicationCommandOption{
				{
				Type:			discordgo.ApplicationCommandOptionUser,
				Name:			"user",
				Description:	"The target user",
				Required:		true,
			},
			{
				Type:			discordgo.ApplicationCommandOptionString,
				Name:			"reason",
				Description:	"Reason for kicking user",
				Required:		false,
			},
		},
	},
	},
	},
	SubCommands: map[string]SubCommandHandler{
		"kick": RequirePermissions(discordgo.PermissionKickMembers, handleKick),
	},
}
