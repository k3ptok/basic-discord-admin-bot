package main

import (
	"log/slog"
	"sync"
	"github.com/bwmarrin/discordgo"
)

// Thread-safe map to store config. Keys are GuildIDs, values are ChannelIDs.
// Note: In production, save this to a database so it survives restarts.
var (
	configMu       sync.RWMutex
	customChannels = make(map[string]string)
)

// Gets custom channel or falls back to server default
func findNotificationChannel(s *discordgo.Session, guildID string) string{
	configMu.RLock()
	customID, exists := customChannels[guildID]
	configMu.RUnlock()

	if exists && customID != "" {
		return customID
	}

	guild, err := s.Guild(guildID)
	if err != nil {
		slog.Error("Failed to fetch guild details for fallback", "error", err)
		return ""
	}

	// Use systemchannelid if available
	if guild.SystemChannelID != "" {
		return guild.SystemChannelID
	}

	// Final Fallback. Find the first text channel the bot can find
	channels, _ := s.GuildChannels(guildID)
	for _, ch := range channels {
		if ch.Type == discordgo.ChannelTypeGuildText {
			return ch.ID
		}
	}
	return ""
}