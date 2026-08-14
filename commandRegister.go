package main

import (
	"github.com/bwmarrin/discordgo"
	"log/slog"
	"time"
	"fmt"
)

type GuildConfig struct {
	GuildID           string
	ReminderChannelID string
	ReminderTimings   []string 
}

type ScheduledEvent struct {
	EventID   string
	GuildID   string
	Title     string
	StartTime time.Time
	// Tracks which offsets have already fired for this event ID.
	// Key: string (e.g., "15m"), Value: bool
	SentReminders map[string]bool 
}


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
								Type:         discordgo.ApplicationCommandOptionChannel,
								Name:         "channel",
								Description:  "The text channel for reminders",
								Required:     true,
								ChannelTypes: []discordgo.ChannelType{discordgo.ChannelTypeGuildText},
							},
						},
					},
					{ // Setup "view" command to view current config settings
						Name:        "view",
						Description: "View current notification and reminder settings",
						Type:        discordgo.ApplicationCommandOptionSubCommand,
					},
					{ // Reminders command to choose reminder times
						Name:        "reminders",
						Description: "Change times that event reminders are sent",
						Type:        discordgo.ApplicationCommandOptionSubCommand,
						Options: []*discordgo.ApplicationCommandOption{
							{ 
								Type:        discordgo.ApplicationCommandOptionString,
								Name:        "before",
								Description: "Time before event (e.g., 30m, 2h)",
								Required:    true,
							},
						},
					},
				},
			},
			{
				Name:        "mod",
				Description: "Moderation Utilities",
				Type:        discordgo.ApplicationCommandOptionSubCommandGroup,
				Options: []*discordgo.ApplicationCommandOption{ 
					{ 
						Name:        "reload-domains",
						Description: "Reload the banned-domains.txt file to update automod",
						Type:        discordgo.ApplicationCommandOptionSubCommand,
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

// Handles the command usage
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

	// Layer 1: Find the SubcommandGroup safely
	var activeGroup *discordgo.ApplicationCommandInteractionDataOption
	for _, opt := range data.Options {
		if opt.Name == "config" || opt.Name == "mod" {
			activeGroup = opt
			break
		}
	}
	if activeGroup == nil || len(activeGroup.Options) == 0 {
		return
	}

	// Route based on Group Name
	switch activeGroup.Name {
	case "config":
		// Layer 2: Determine which subcommand inside command was called
		subcommand := activeGroup.Options[0]

		switch subcommand.Name {
		case "set-channel":
			// Layer 3: Safely find the channel argument
			if len(subcommand.Options) == 0 || subcommand.Options[0].Name != "channel" {
				return
			}
			selectedChannel := subcommand.Options[0].ChannelValue(s)

			configMu.Lock()
			settings := botConfig[i.GuildID]
			settings.ChannelID = selectedChannel.ID
			botConfig[i.GuildID] = settings
			configMu.Unlock()

			SaveConfig()

			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "✅ Event reminders will now be sent to <#" + selectedChannel.ID + ">!",
				},
			})

		case "view": 
			configMu.Lock()
			settings, exists := botConfig[i.GuildID]
			configMu.Unlock()

			channelText := "Not set"
			if exists && settings.ChannelID != "" {
				channelText = "<#" + settings.ChannelID + ">"
			}

			timingsText := "None configured"
			if exists && len(settings.Timings) > 0 {
				timingsText = fmt.Sprintf("`%v`", settings.Timings)
			}

			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: fmt.Sprintf("📋 **Current Server Configuration:**\n• **Notification Channel:** %s\n• **Active Reminders:** %s", channelText, timingsText),
				},
			})

		case "reminders": 
			if len(subcommand.Options) == 0 || subcommand.Options[0].Name != "before" {
				return
			}
			userInput := subcommand.Options[0].StringValue()

			duration, err := time.ParseDuration(userInput)
			if err != nil {
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Flags:   discordgo.MessageFlagsEphemeral,
						Content: "❌ **Invalid time format!** Please use markers like `30m` or `2h`.",
					},
				})
				return
			}

			durationStr := duration.String()

			configMu.Lock()
			settings := botConfig[i.GuildID]

			isDuplicate := false
			for _, existingTime := range settings.Timings {
				if existingTime == durationStr {
					isDuplicate = true
					break
				}
			}

			if isDuplicate {
				configMu.Unlock()
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "⚠️ A reminder offset for `" + durationStr + "` already exists on this server.",
					},
				})
				return
			}

			settings.Timings = append(settings.Timings, durationStr)
			botConfig[i.GuildID] = settings
			configMu.Unlock()

			SaveConfig()

			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "✅ Successfully added a custom alert! Bot will now send notifications **" + durationStr + "** prior to scheduled events.",
				},
			})
		} 

	case "mod":
		subcommand := activeGroup.Options[0]

		switch subcommand.Name {
		case "reload-domains":
			LoadBannedDomains()

			slog.Info("Admin triggered on-the-fly domain blocklist refresh",
				"userID", i.Member.User.ID,
				"user", i.Member.User.Username,
				"guildID", i.GuildID,
			)

			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{ // FIXED: Changed InteractionResponse to InteractionRespond
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{ // FIXED: Removed duplicate 'Data: Data:' typo compile block
					Content: "🔄 **Blocklist Updated!** The `banned_domains.txt` file has been successfully reloaded into the bot's active memory.",
				},
			})
		}
	}
}

