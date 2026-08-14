package main

import (
	"time"
	"log/slog"
	"github.com/bwmarrin/discordgo"
)

//Background loop checking for running events and sending messages to server at specific times
func StartGlobalReminderScheduler(s *discordgo.Session) {
	ticker := time.NewTicker(1 * time.Minute)

	notifiedEvents := make(map[string]map[string]bool)
	// Set up two maps to add events that have been notified to avoid spamming the chat channel

	go func() {
		for range ticker.C {
			// Loop through all servers the bot is a member of
			for _, guild := range s.State.Guilds {
				events, err := s.GuildScheduledEvents(guild.ID, false)
				if err != nil {
					slog.Error("Failed to fetch events", "guildID", guild.ID, "error", err)
					continue
				}

				for _, event := range events {
					//skip if event is complete
					if event.Status == discordgo.GuildScheduledEventStatusCompleted {
						continue
					}
					
					//Initialize tracking map if it does not exist for some reason
					if _, exists := notifiedEvents[event.ID]; !exists {
						notifiedEvents[event.ID] = make(map[string]bool)
					}

					timeRemaining := event.ScheduledStartTime.Sub(time.Now())

					var milestone string
					var messageText string

					// MILESTONE 1: 1 Day Before (Between 23h 55m and 24h remaining)
					if timeRemaining > 23*time.Hour+55*time.Minute && timeRemaining <= 24*time.Hour {
						milestone = "24h"
						messageText = "📅 **Upcoming Event:** " + event.Name + " starts in 24 hours!"
					

					// MILESTONE 2: 12 Hours Before (Between 11h 55m and 12h remaining)
					} else if timeRemaining > 11*time.Hour+55*time.Minute && timeRemaining <= 12*time.Hour {
						milestone = "12h"
						messageText = "📅 **Upcoming Event:** " + event.Name + " starts in 12 hours!"
					
					
					// MILESTONE 3: 30 Minutes Before (Between 29m and 30m remaining)
					} else if timeRemaining > 29*time.Minute && timeRemaining <= 30*time.Minute {
						milestone = "30m"
						messageText = "📅 **Upcoming Event:** " + event.Name + " starts in 30 minutes!!"
					}

					//If a milestone matches and we HAVEN'T sent it yet for this specific event
					if milestone != "" && !notifiedEvents[event.ID][milestone] {
						targetChannelID := findNotificationChannel(s, guild.ID)
						
						if targetChannelID != "" {
							s.ChannelMessageSend(targetChannelID, messageText)

							// Mark only this specific milestone as completed
							notifiedEvents[event.ID][milestone] = true
							slog.Info("Sent timed reminder", "event", event.Name, "milestone", milestone)
						}
					}
					//Cleanup completed or canceled events
					if event.Status == discordgo.GuildScheduledEventStatusCompleted || event.Status == discordgo.GuildScheduledEventStatusCanceled {
						delete(notifiedEvents, event.ID)
					}
				}
			}

		}
	}()

}