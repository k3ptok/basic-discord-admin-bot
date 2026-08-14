package main

import (
	"time"
	"sync"
)

type UserMessageTracker struct {
	Timestamps []time.Time
}

var (
	// Key: GuildID + ":" + UserID
	rateLimitMap = make(map[string]*UserMessageTracker)
	rateLimitMu  sync.Mutex
)

// IsSpamming checks if a user sent more than 5 messages within 5 seconds
func IsSpamming(guildID, userID string) bool {
	rateLimitMu.Lock()
	defer rateLimitMu.Unlock()

	key := guildID + ":" + userID
	now := time.Now()

	tracker, exists := rateLimitMap[key]
	if !exists {
		rateLimitMap[key] = &UserMessageTracker{Timestamps: []time.Time{now}}
		return false
	}

	// Append current message time
	tracker.Timestamps = append(tracker.Timestamps, now)

	// Clean out old timestamps that are older than 5 seconds
	validWindow := now.Add(-5 * time.Second)
	var freshTimestamps []time.Time
	for _, t := range tracker.Timestamps {
		if t.After(validWindow) {
			freshTimestamps = append(freshTimestamps, t)
		}
	}
	tracker.Timestamps = freshTimestamps

	// Flag as spam if history count within 5 seconds exceeds 5 messages
	if len(tracker.Timestamps) > 5 {
		return true
	}
	return false
}