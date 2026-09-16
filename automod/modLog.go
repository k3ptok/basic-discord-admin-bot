package automod

import (
	"fmt"
	"context"
	"log/slog"
	"github.com/k3ptok/BasicDiscordBot/internal/database"
	"github.com/bwmarrin/discordgo"
	"sync"
)

type ModLogger struct {
	db 			*database.Queries
	logger      *slog.Logger
	cache       map[string]string // Maps GuildID to ChannelID
	mu          sync.RWMutex
}

func NewModLogger(logger *slog.Logger, db *database.Queries) *ModLogger {
	return &ModLogger{
		db:		db,
		logger: logger,
		cache:	make(map[string]string),
	}
}

// LogViolation fetches the channel from the DB and dispatches the embed
func (ml *ModLogger) LogViolation(s *discordgo.Session, m *discordgo.MessageCreate, reason, matchedWord string) {
// check the memory cache first 
	ml.mu.RLock()
	channelID, exists := ml.cache[m.GuildID]
	ml.mu.RUnlock()

	// if it's not in the cache, query the database 
	if !exists {
		var err error
		channelID, err = ml.db.GetLogChannel(context.Background(), m.GuildID)
		if err != nil || channelID == "" {
			// If it fails or isn't set, cache an empty string so we don't spam the DB asking for it again
			ml.mu.Lock()
			ml.cache[m.GuildID] = ""
			ml.mu.Unlock()
			return
		}

		// save the newly found ID into the cache
		ml.mu.Lock()
		ml.cache[m.GuildID] = channelID
		ml.mu.Unlock()
	}

	// If the cached value is empty (no log channel configured), exit
	if channelID == "" {
		return
	}

embed := &discordgo.MessageEmbed{
		Title:       "🚨 AutoMod Action Taken",
		Color:       0xE74C3C,
		Description: fmt.Sprintf("A message by <@%s> was deleted in <#%s>.", m.Author.ID, m.ChannelID),
		Fields: []*discordgo.MessageEmbedField{
			{Name: "User", Value: fmt.Sprintf("%s (`%s`)", m.Author.Username, m.Author.ID), Inline: true},
			{Name: "Reason", Value: reason, Inline: true},
			{Name: "Trigger Token/Domain", Value: fmt.Sprintf("`%s`", matchedWord), Inline: false},
			{Name: "Flagged Message Content", Value: fmt.Sprintf("```%s```", m.Content), Inline: false},
		},
	}

	if _, err := s.ChannelMessageSendEmbed(channelID, embed); err != nil {
		ml.logger.Error("Failed to send mod-log embed", slog.String("guild_id", m.GuildID), slog.Any("error", err))
	}
}

func (ml *ModLogger) ClearCache(guildID string) {
	ml.mu.Lock()
	delete(ml.cache, guildID)
	ml.mu.Unlock()
}
