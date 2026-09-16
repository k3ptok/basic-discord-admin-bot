package commands

import (
	"github.com/bwmarrin/discordgo"
)

// Declare Admin Command
func NewAdminCommandStructure() Command {
	var adminPerms int64 = discordgo.PermissionAdministrator
	return Command{
		Definition: &discordgo.ApplicationCommand{
			Name:                     "admin",
			Description:              "Administration Commands",
			DefaultMemberPermissions: &adminPerms,
			Options: []*discordgo.ApplicationCommandOption{
				{	//New command under /admin
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "kick",
					Description: "Kick a user from the server",
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionUser,
							Name:        "user",
							Description: "The target user",
							Required:    true,
						},
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "reason",
							Description: "Reason for kicking user",
							Required:    false,
						},
					},
				},
				// Timeout subcommand
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "timeout",
					Description: "Temporarily mute a user in the server",
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionUser,
							Name:        "user",
							Description: "The user to timeout",
							Required:    true,
						},
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "duration",
							Description: "Duration (e.g., 10m, 1h, 24h)",
							Required:    true,
						},
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "reason",
							Description: "Reason for the timeout",
							Required:    false,
						},
					},
				},
				// Ban Subcommand
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "ban",
					Description: "Permanently ban a user from the server",
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionUser,
							Name:        "user",
							Description: "The user to ban",
							Required:    true,
						},
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "reason",
							Description: "Reason for the ban",
							Required:    false,
						},
					},
				},
				// Unban Subcommand
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "unban",
					Description: "Revoke a ban using a User ID",
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "user_id",
							Description: "The ID of the user to unban",
							Required:    true,
						},
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "reason",
							Description: "Reason for unbanning",
							Required:    false,
						},
					},
				},
				// user-purge command
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "purge-user",
					Description: "Delete recent messages from a specific user only",
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionUser,
							Name:        "user",
							Description: "The user whose messages you want to delete",
							Required:    true,
						},
						{
							Type:        discordgo.ApplicationCommandOptionInteger,
							Name:        "scan_limit",
							Description: "How many messages to scan back (default 50, max 100)",
							Required:    false,
						},
					},
				},
				// Warn subcommand
				{	
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "warn",
					Description: "Issue a formal warning to a user",
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionUser,
							Name:        "user",
							Description: "The user to warn",
							Required:    true,
						},
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "reason",
							Description: "Reason for the warning",
							Required:    true, // Making reason required for warnings
						},
					},
				},
				// log-channel command
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "log-channel",
					Description: "Set the channel for AutoMod violation alerts",
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionChannel,
							Name:        "target",
							Description: "Select the target text channel",
							Required:    true,
							ChannelTypes: []discordgo.ChannelType{
								discordgo.ChannelTypeGuildText,
							},
						},
					},
				},
				// history command
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "history",
					Description: "View a user's infraction history",
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionUser,
							Name:        "user",
							Description: "The user to investigate",
							Required:    true,
						},
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "whois",
					Description: "View detailed account information about a user",
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionUser,
							Name:        "user",
							Description: "The user to investigate",
							Required:    true,
						},
					},
				},
			},
		},
		SubCommands: map[string]SubCommandHandler{
			"kick":        RequirePermissions(discordgo.PermissionKickMembers, handleKick),
			"timeout": RequirePermissions(discordgo.PermissionModerateMembers, handleTimeout),
			"ban":   RequirePermissions(discordgo.PermissionBanMembers, handleBan),
			"unban": RequirePermissions(discordgo.PermissionBanMembers, handleUnban),
			"log-channel": RequirePermissions(discordgo.PermissionAdministrator, MakeLogChannelHandler()),
			"warn": RequirePermissions(discordgo.PermissionKickMembers, handleWarn),
			"history": RequirePermissions(discordgo.PermissionKickMembers, handleHistory),
			"purge-user": RequirePermissions(discordgo.PermissionManageMessages, handlePurgeUser),
			"whois": RequirePermissions(discordgo.PermissionKickMembers, handleWhois),
		},
	}
}
