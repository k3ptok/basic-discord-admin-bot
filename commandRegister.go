package main

import (
	"github.com/bwmarrin/discordgo"
	"log/slog"
)

func RegisterNamespacedCommand(s *discordgo.Session) {
	cmd := &discordgo.ApplicationCommand{
		Name:        "mybot",
		Description: "Base command for all bot controls",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Name:        "config",
				Description: "Manage server notification configurations",
				Type:        discordgo.ApplicationCommandOptionSubCommandGroup,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Name:        "set-channel",
						Description: "Set the channel where event reminders will be sent",
						Type:        discordgo.ApplicationCommandOptionSubCommand,
						Options: []*discordgo.ApplicationCommandOption{
							{
								Type:        discordgo.ApplicationCommandOptionChannel,
								Name:        "channel",
								Description: "The text channel for reminders",
								Required:    true,
								ChannelTypes: []discordgo.ChannelType{discordgo.ChannelTypeGuildText},
							},
						},
					},
				},
			},
		},
	}

	_, err := s.ApplicationCommandCreate(s.State.User.ID, "", cmd)
	if err != nil {
		slog.Error("Failed to register namespaced command", "error", err)
	}
}

//Handles the command usage
func InteractionHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionApplicationCommand {
		return
	}

	data := i.ApplicationCommandData()
	if data.Name != "mybot" {
		return
	}

	// Ensure the user has admin/management rights before checking options
	if i.Member.Permissions&discordgo.PermissionManageGuild == 0 {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "❌ You need 'Manage Server' permissions to run this command.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	// Layer 1: Check for the 'config' SubcommandGroup
	if len(data.Options) == 0 || data.Options[0].Name != "config" {
		return
	}
	configGroup := data.Options[0]

	// Layer 2: Check for the 'set-channel' Subcommand
	if len(configGroup.Options) == 0 || configGroup.Options[0].Name != "set-channel" {
		return
	}
	setChannelSub := configGroup.Options[0]

	// Layer 3: Extract the channel argument
	if len(setChannelSub.Options) == 0 || setChannelSub.Options[0].Name != "channel" {
		return
	}
	
	// Safely retrieve the resolved channel object
	selectedChannel := setChannelSub.Options[0].ChannelValue(s)

	// Save to your system memory/database
	configMu.Lock()
	customChannels[i.GuildID] = selectedChannel.ID
	configMu.Unlock()

	//Save config to JSON so it survives a computer/server host restart
	SaveConfig()

	// Reply to confirmation
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "✅ Event reminders will now be sent to <#" + selectedChannel.ID + ">!",
		},
	})
}
