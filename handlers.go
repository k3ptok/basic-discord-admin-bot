package main

import (
	"github.com/bwmarrin/discordgo"
	"log/slog"
)

func onMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore all messages created by the bot itself
	if m.Author.ID == s.State.User.ID {
		return
	}

	// If the message is "ping" reply with "Pong!"
	if m.Content == "ping" {
		_, err := s.ChannelMessageSend(m.ChannelID, "Pong!")
		if err != nil {
			slog.Error("Failed to send reply message", "error", err)
			return
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