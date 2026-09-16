package automod

import (
	"log/slog"
	"strings"
	"sync"
	"time"
	"fmt"
	"unicode"
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
	modLogger *ModLogger
	rules 	[]Rule
	mu 		sync.Mutex
	cache 	map[string]*userHistory
}

func NewManager(logger *slog.Logger, modLogger *ModLogger) *Manager {
	m := &Manager{
		logger: logger,
		modLogger: modLogger,
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

	normLast := normalizeText(history.lastMessage)
	normCurrent := normalizeText(msg.Content)
	isDuplicate := normCurrent != "" && isSimilar(normLast, normCurrent)

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
	// log violation before it is deleted
	m.modLogger.LogViolation(s, msg, reason, "N/A")

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

// normalizeText strips spaces and converts to lowercase for consistent matching
func normalizeText(s string) string {
	var builder strings.Builder
	for _, r := range s {
		if !unicode.IsSpace(r) {
			builder.WriteRune(unicode.ToLower(r))
		}
	}
	return builder.String()
}

// isSimilar checks if two strings are highly similar (e.g., > 85% match)
func isSimilar(s1, s2 string) bool {
	if s1 == s2 {
		return true
	}
	
	r1, r2 := []rune(s1), []rune(s2)
	len1, len2 := len(r1), len(r2)

	if len1 == 0 || len2 == 0 {
		return false
	}

	// Calculate Levenshtein distance
	d := make([][]int, len1+1)
	for i := range d {
		d[i] = make([]int, len2+1)
		d[i][0] = i
	}
	for j := 0; j <= len2; j++ {
		d[0][j] = j
	}

	for i := 1; i <= len1; i++ {
		for j := 1; j <= len2; j++ {
			cost := 0
			if r1[i-1] != r2[j-1] {
				cost = 1
			}
			d[i][j] = min(d[i-1][j]+1, min(d[i][j-1]+1, d[i-1][j-1]+cost))
		}
	}

	distance := d[len1][len2]
	maxLen := max(len1, len2)

	// Calculate similarity percentage (0.0 to 1.0)
	similarity := 1.0 - (float64(distance) / float64(maxLen))

	return similarity > 0.85
}