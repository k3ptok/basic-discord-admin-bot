package automod

import (
	"fmt"
	"context"
	"log/slog"
	"github.com/k3ptok/BasicDiscordBot/internal/database"
	"github.com/bwmarrin/discordgo"
)

type ModLogger struct {
	db 			*database.Queries
	logger      *slog.Logger
}

func NewModLogger(logger *slog.Logger, db *database.Queries) *ModLogger {
	return &ModLogger{
		db:		db,
		logger: logger,
	}
}

// LogViolation fetches the channel from the DB and dispatches the embed
func (ml *ModLogger) LogViolation(s *discordgo.Session, m *discordgo.MessageCreate, reason, matchedWord string) {
	channelID, err := ml.db.GetLogChannel(context.Background(), m.GuildID)
	if err != nil || channelID == "" {
		// Skip dispatch if no channel is configured or found
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
