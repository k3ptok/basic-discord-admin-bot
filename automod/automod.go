package automod

import (
	"log/slog"
	//"strings"
	"sync"
	"time"
	"fmt"

	"github.com/bwmarrin/discordgo"
)

//Defines a check that evaluates messages posted in chat
type Rule func(s *discordgo.Session, m *discordgo.MessageCreate) (violation bool, reason string)

type userHistory struct {
	lastMessage string
	timestamps 	[]time.Time
}

type Manager struct {
	logger *slog.Logger
	rules 	[]Rule
	mu 		sync.Mutex
	cache 	map[string]*userHistory
}

func NewManager(logger *slog.Logger) *Manager {
	m := &Manager{
		logger: logger,
		rules:	make([]Rule, 0),
		cache:	make(map[string]*userHistory),
	}

	go m.cleanupCache()
	
	return m
}

func (m *Manager) AddRule(rule Rule) {
	m.rules = append(m.rules, rule)
}

func (m *Manager) ProcessMessage(s *discordgo.Session, msg *discordgo.MessageCreate) {
	//Ignore bots and self
	if msg.Author.Bot || msg.WebhookID != "" || msg.GuildID == "" {
		return
	}

	//Spamming check
	if m.isSpamming(msg) {
		m.applyTimeout(s, msg, "Automated spam detection (rapid/duplicate messages)")
		return
	}
	
	// Grab current rule list
	for _, rule := range m.rules {
		if violation, reason := rule(s, msg); violation {
			m.handleViolation(s, msg, reason)
			break
		}
	}
}

//Currently returns true if user posts 5 duplicate messages or 5 rapid messages within 5 seconds
func (m *Manager) isSpamming(msg *discordgo.MessageCreate) bool {
	key := msg.GuildID + ":" + msg.Author.ID
	now := time.Now()

	m.mu.Lock()
	defer m.mu.Unlock()

	history, exists := m.cache[key]
	if !exists {
		history = &userHistory{}
		m.cache[key] = history
	}

	var recent []time.Time
	for _, t := range history.timestamps {
		if now.Sub(t) < 5 * time.Second {
			recent = append(recent, t)
		}
	}
	recent = append(recent, now)
	history.timestamps = recent

	isDuplicate := history.lastMessage == msg.Content && msg.Content != ""
	history.lastMessage = msg.Content

	return (isDuplicate && len(recent) >= 5) || len(recent) >= 5
}

func (m *Manager) applyTimeout(s *discordgo.Session, msg *discordgo.MessageCreate, reason string) {
	err := s.ChannelMessageDelete(msg.ChannelID, msg.ID)
	if err != nil {
		m.logger.Error("Failed to delete message after timeout", slog.Any("error", err))
		return
	}
	until := time.Now().Add(10 * time.Minute)
	err = s.GuildMemberTimeout(msg.GuildID, msg.Author.ID, &until)
	if err != nil {
		m.logger.Error("Failed to timeout offending user", slog.Any("error", err))
		return
	}

	m.logger.Warn("Timed out spammer", slog.String("user_id", msg.Author.ID))
	_, _ = s.ChannelMessageSend(msg.ChannelID, fmt.Sprintf("🚨 %s has been placed in timeout for 10 minutes. Reason: **%s**", msg.Author.Mention(), reason))
}

func (m *Manager) handleViolation(s *discordgo.Session, msg *discordgo.MessageCreate, reason string) {
	//attempt to delete message
	err := s.ChannelMessageDelete(msg.ChannelID, msg.ID)
	if err != nil {
		m.logger.Error("Failed to delete message", slog.Any("error", err))
		return
	}

	//temporary warning
	warnMsg, err := s.ChannelMessageSend(msg.ChannelID,
	"⚠️ " + msg.Author.Mention() + " your message was removed for: **" + reason + "**")
	if err != nil {
		m.logger.Error("failed to create warnmsg in violation handler", slog.Any("error", err))
		return
	}

	m.logger.Info("Automod action taken",
		slog.String("user_id", msg.Author.ID),
		slog.String("user_name", msg.Author.Username),
		slog.String("reason", reason))

	//warning message cleanup
	go func() {
		time.Sleep(7 * time.Second)
		_ = s.ChannelMessageDelete(msg.ChannelID, warnMsg.ID)
	}()
}

func (m *Manager) cleanupCache() {
	for {
		time.Sleep(5 * time.Minute)
		m.mu.Lock()
		now := time.Now()
		for key, history := range m.cache {
			if len(history.timestamps) == 0 || now.Sub(history.timestamps[len(history.timestamps)-1]) > 5*time.Minute {
				delete(m.cache, key)
			}
		}
		m.mu.Unlock()
	}
	
}