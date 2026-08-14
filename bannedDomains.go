package main

import (
	"os"
	
	"sync"
	"log/slog"
	"bufio"
	"strings"
)

var (
	bannedDomains = make(map[string]bool)
	domainsMu       sync.RWMutex
)

func LoadBannedDomains() {
	domainsMu.Lock()
	defer domainsMu.Unlock()

	file, err := os.Open("banned-domains.txt")
	if os.IsNotExist(err) {
		// Create an empty file if it doesn't exist yet
		_ = os.WriteFile("banned_domains.txt", []byte(""), 0644)
		return
	} else if err != nil {
		slog.Error("Failed to open banned domains file", "error", err)
		return
	}
	defer file.Close()

	// Clear current memory list before reloading
	bannedDomains = make(map[string]bool)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// Ignore empty lines or comment tags
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		bannedDomains[strings.ToLower(line)] = true
	}
	slog.Info("Successfully loaded banned domains", "count", len(bannedDomains))
}