package main

import (
	"time"
	"log/slog"
	"github.com/bwmarrin/discordgo"
	"fmt"
)

// Background loop checking for running events and sending messages to server at specific times
func StartGlobalReminderScheduler(s *discordgo.Session) {
	ticker := time.NewTicker(1 * time.Minute)

	// Key 1: EventID, Key 2: Timing String (e.g., "15m", "1h"), Value: Has sent
	notifiedEvents := make(map[string]map[string]bool)

	go func() {
		for range ticker.C {
			now := time.Now()

			// Loop through all servers the bot is a member of
			for _, guild := range s.State.Guilds {
				// 1. Safe Lock: Pull this specific server's custom timings slice
				configMu.RLock()
				serverSettings, hasSettings := botConfig[guild.ID]
				configMu.RUnlock()

				// Fallback: If the server hasn't set any custom timings, use a default list
				activeTimings := []string{"30m", "12h", "24h"} // Matches your old default rules
				if hasSettings && len(serverSettings.Timings) > 0 {
					activeTimings = serverSettings.Timings
				}

				events, err := s.GuildScheduledEvents(guild.ID, false)
				if err != nil {
					slog.Error("Failed to fetch events", "guildID", guild.ID, "error", err)
					continue
				}

				for _, event := range events {
					// Cleanup completed or canceled events from tracking memory
					if event.Status == discordgo.GuildScheduledEventStatusCompleted || event.Status == discordgo.GuildScheduledEventStatusCanceled {
						delete(notifiedEvents, event.ID)
						continue
					}

					// Initialize tracking map for this event if it doesn't exist yet
					if _, exists := notifiedEvents[event.ID]; !exists {
						notifiedEvents[event.ID] = make(map[string]bool)
					}

					// 2. Dynamic Evaluator: Loop through every custom timing the server requested
					for _, timingStr := range activeTimings {
						// Skip if we already sent this specific reminder milestone for this event
						if notifiedEvents[event.ID][timingStr] {
							continue
						}

						offset, err := time.ParseDuration(timingStr)
						if err != nil {
							continue // Skip corrupted configurations safely
						}

						// Calculate the exact target time this reminder should trigger
						// Example: StartTime is 5:00 PM, offset is 30m -> target is 4:30 PM
						reminderTargetTime := event.ScheduledStartTime.Add(-offset)

						// 3. Robust Time Check: Ensure the target time has arrived,
						// but don't alert if the event start time has already passed completely.
						if now.After(reminderTargetTime) && now.Before(event.ScheduledStartTime) {
							
							targetChannelID := findNotificationChannel(s, guild.ID)
							if targetChannelID == "" {
								continue
							}

							// Format a clean, human-readable reminder message dynamically
							messageText := fmt.Sprintf("📅 **Upcoming Event:** %s starts in %s!", event.Name, timingStr)
							if timingStr == "0s" || timingStr == "0m" {
								messageText = fmt.Sprintf("🎉 **Event Alert:** %s is starting right now!", event.Name)
							}

							_, err = s.ChannelMessageSend(targetChannelID, messageText)
							if err == nil {
								// Mark only this specific milestone string as completed
								notifiedEvents[event.ID][timingStr] = true
								slog.Info("Sent custom timed reminder", "event", event.Name, "milestone", timingStr, "guild", guild.ID)
							} else {
								slog.Error("Failed to send scheduled reminder", "channel", targetChannelID, "error", err)
							}
						}
					}
				}
			}
		}
	}()
}
