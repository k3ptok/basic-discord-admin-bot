package main

import (
	"github.com/bwmarrin/discordgo"
	"log/slog"
	"strings"
	"fmt"
	"time"
)

func containsURL(text string) bool {
	// A simple check for http:// or https:// patterns
	return strings.Contains(text, "http://") || strings.Contains(text, "https://")
}

func isUserStaff(s *discordgo.Session, guildID, userID string) bool {
	member, err := s.State.Member(guildID, userID)
	if err != nil {
		// Fallback to API if state cache doesn't have the user
		member, err = s.GuildMember(guildID, userID)
		if err != nil {
			return false
		}
	}

	// Check if they have admin permissions directly
	permissions, err := s.State.UserChannelPermissions(userID, member.GuildID) // simplified approach or use member roles
	if err == nil && permissions&discordgo.PermissionAdministrator != 0 {
		return true
	}

	// Loop through their assigned roles to check for specific staff flags
	for _, roleID := range member.Roles {
		role, err := s.State.Role(guildID, roleID)
		if err == nil {
			// Exempt anyone with "Manage Server" or "Kick/Ban" capabilities
			if role.Permissions&discordgo.PermissionManageGuild != 0 || 
			   role.Permissions&discordgo.PermissionBanMembers != 0 {
				return true
			}
		}
	}
	return false
}

func onMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore all messages created by a bot 
	if m.Author.Bot {
		return
	}

	if isUserStaff(s, m.GuildID, m.Author.ID) {
		return
	}

	contentLower:= strings.ToLower(m.Content)
	shouldTimeout := false
	reason := ""
	
	// if spamming
	if IsSpamming(m.GuildID, m.Author.ID) {
		shouldTimeout = true
		reason = "Rapid chat flooding (exceeded 5 messages in 5 seconds)"
	}

	//Keyword check
	if !shouldTimeout && containsURL(contentLower) {
		if strings.Contains(contentLower, "free nitro") || 
		   strings.Contains(contentLower, "gift for you") {
			shouldTimeout = true
			reason = "Phishing text match (Keywords + URL)"
		}
	}

	// Blocked domains check
	if !shouldTimeout && containsURL(contentLower) {
		domainsMu.RLock()
		for domain := range bannedDomains {
			if strings.Contains(contentLower, domain) {
				shouldTimeout = true
				reason = "Posted link matching blocklist domain: " + domain
				break
			}
		}
		domainsMu.RUnlock()
	}

	//if scam pattern is detected
	if shouldTimeout {
		// Immediately clear out the offensive message
		_ = s.ChannelMessageDelete(m.ChannelID, m.ID)

		timeoutDuration := 24 * time.Hour
		timeoutUntil := time.Now().Add(timeoutDuration)

		err := s.GuildMemberTimeout(m.GuildID, m.Author.ID, &timeoutUntil)
		if err != nil {
			slog.Error("AUTOMOD FAILURE: Action could not execute",
				"component", "moderator",
				"guildID", m.GuildID,
				"userID", m.Author.ID,
				"username", m.Author.Username,
				"reason", reason,
				"error", err.Error(),
			)
		} else {
			// Write the security action entry straight to bot.log file!
			slog.Warn("AUTOMOD SUCCESS: Isolated account",
				"component", "moderator",
				"guildID", m.GuildID,
				"userID", m.Author.ID,
				"username", m.Author.Username,
				"reason", reason,
			)

			// Inform the server logs channel
			logChannelID := findNotificationChannel(s, m.GuildID)
			if logChannelID != "" {
				alert := fmt.Sprintf("🚨 **Automod Action:** %s (%s) was timed out for 24 hours.\n**Reason:** %s", 
					m.Author.Username, m.Author.Mention(), reason)
				_, _ = s.ChannelMessageSend(logChannelID, alert)
			}
		}
	}
}


func onConnect(s *discordgo.Session, c *discordgo.Connect) {
	slog.Info("Websocket session established with Discord.")
}

func onDisconnect(s *discordgo.Session, d *discordgo.Disconnect) {
	slog.Warn("Bot disconnected from Discord gateway! Reconnecting automatically...")
}

func guildScheduledEventCreate(s *discordgo.Session, e *discordgo.GuildScheduledEventCreate) {
	slog.Info("New event detected", "Name", e.Name, "StartTime", e.ScheduledStartTime, "GuildID", e.GuildID)
}