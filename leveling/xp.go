package leveling

import (
	"context"
	"fmt"
	"math/rand"
	"log/slog"
	"sync"
	"time"
	"github.com/bwmarrin/discordgo"
	"github.com/k3ptok/BasicDiscordBot/internal/database"
	
)

type Manager struct {
	db		*database.Queries
	cache   map[string]time.Time
	logger *slog.Logger
	mu 		sync.Mutex
}

func NewManager(db *database.Queries, logger *slog.Logger) *Manager {
	m := &Manager{
		db:		db,
		logger:	logger,
		cache:	make(map[string]time.Time),
	}
	go m.cleanupCache()
	return m
}

func CalculateLevel(totalXP int64) int64 {
	var level int64 = 0
	var threshold int64 = 0
	
	for totalXP >= threshold {
		level++
		threshold += 100 * (level + 1)
	}
	return level
}

func CalculateNextLevelXP(currentLevel int64) int64 {
	var level int64 = 0
	var threshold int64 = 0
	
	// Fast-forward the threshold to the current level
	for level <= currentLevel {
		level++
		threshold += 100 * (level + 1)
	}
	return threshold
}

func (m *Manager) ProcessMessage(s *discordgo.Session, msg *discordgo.MessageCreate) {
	// Ignore self, other bots, webhooks, and DM's
	if msg.Author.Bot || msg.WebhookID != "" || msg.GuildID == "" {
		return
	}

	key := msg.GuildID + ":" + msg.Author.ID
	now := time.Now()

	// check cooldown (15 seconds)
	m.mu.Lock()
	lastSeen, exists := m.cache[key]
	if exists && now.Sub(lastSeen) < 15*time.Second {
		m.mu.Unlock()
		return
	}
	m.cache[key] = now
	m.mu.Unlock()

	// fetch current stats ( ignore error if a user does not exist. Technically should not happen)
	currentStats, _ := m.db.GetUserRank(context.Background(), database.GetUserRankParams{
		GuildID:	msg.GuildID,
		UserID:		msg.Author.ID,
	})

	// issue random xp - 15-25
	xpGained := int64(rand.Intn(11) + 15)
	newTotalXP := currentStats.Xp + xpGained
	newLevel := CalculateLevel(newTotalXP)

	// save to database
	_, err := m.db.UpdateUserXP(context.Background(), database.UpdateUserXPParams{
		GuildID: 	msg.GuildID,
		UserID:		msg.Author.ID,
		Xp:			newTotalXP,
		Level:		newLevel,
		LastXpGain: now,
	})
	if err != nil {
		m.logger.Error("Failed to update user xp in database",
		slog.String("guild_id", msg.GuildID),
		slog.String("user_id", msg.Author.ID), 
		slog.Any("error", err),
	)
	return
	}

	// announce level up
	if newLevel > currentStats.Level && currentStats.Xp > 0 {
		_, _ = s.ChannelMessageSend(msg.ChannelID, 
			fmt.Sprintf("🎉 Congratulations %s! You just advanced to **Level %d**!", msg.Author.Mention(), newLevel))
	}
	
}

func (m *Manager) cleanupCache() {
	for {
		time.Sleep(10 * time.Minute)
		m.mu.Lock()
		now := time.Now()
		for key, lastSeen := range m.cache {
			if now.Sub(lastSeen) > 5*time.Minute {
				delete(m.cache, key)
			}
		}
		m.mu.Unlock()
	}
}