package main

import (
	"strings"
	"bufio"
	"net/url"
	"log/slog"
	"os"
	"github.com/bwmarrin/discordgo"
	"github.com/k3ptok/BasicDiscordBot/automod"
)

var (
	//Trusted domains. Should Not be flagged.
	legitDomains = map[string]struct{}{
		"discord.com":       {},
		"discord.gg":        {},
		"discordapp.com":    {},
		"discord.media":     {},
		"steampowered.com":  {},
		"steamcommunity.com":{},
	}

	//Common name impersonations of Discord
	discordLookAlikes = []string{
		"discoord", "discorb", "dlscord", "discort", "discorx", "discord-", "dlscordapp",
	}

	// domain extensions commonly used by scammers
	susExtensions = []string{
		".xyz", ".top", ".info", ".site", ".ru", ".click", ".online", ".gift",
	}
)

func LoadScamDomains(filePath string, logger *slog.Logger) map[string]struct{} {
	domains := make(map[string]struct{})

	file, err := os.Open(filePath)
	if err != nil {
		logger.Error("Failed to open scam domains file", slog.String("path", filePath), slog.Any("error", err))
		return domains
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		//skip empty lines and commented out lines
		if line != "" && !strings.HasPrefix(line, "#") {
			domains[strings.ToLower(line)] = struct{}{}
		}
	}

	if err := scanner.Err(); err != nil {
		logger.Error("Error reading scam domains file", slog.Any("error", err))
	}

	logger.Info("Loaded scam domains into memory", slog.Int("Count:", len(domains)))
	return domains
}

//Attach rules to automod
func RegisterRules(am *automod.Manager, store *automod.DomainStore) {
	//Load scam domains
	//domains := LoadScamDomains("banned-domains.txt", logger)

	am.AddRule(func(s *discordgo.Session, m *discordgo.MessageCreate) (bool, string) {
		// Quick exit if message doesn't contain a URL scheme or dot
		if !strings.Contains(m.Content, ".") {
			return false, ""
		}
		words := strings.Fields(m.Content)
		for _, word := range words {
			cleanWord := strings.ToLower(word)

			// Add temporary scheme if missing so net/url can parse hostname correctly
			if !strings.HasPrefix(cleanWord, "http://") && !strings.HasPrefix(cleanWord, "https://") {
				cleanWord = "https://" + cleanWord
			}

			parsed, err := url.Parse(cleanWord)
			if err != nil {
				continue
			}

			hostname := strings.ToLower(parsed.Hostname())

			// Pass whitelisted official domains immediately
			if _, isLegit := legitDomains[hostname]; isLegit {
				continue
			}

			// Check map for exact domain match
			if store.Has(hostname) {
				return true, "Message contained a known scam/phishing domain"
			}

			for _, match := range discordLookAlikes {
				if strings.Contains(hostname, match) {
					return true, "Suspicious Discord domain impersonation"
				}
			}

			for _, match := range susExtensions {
				if strings.HasSuffix(hostname, match) {
					return true, "suspicious domain extension detected."
				}
			}
		}
		return false, ""
	})


	// **== Chat Rules ==**

	// Rule 1: Block Discord Invites
	am.AddRule(func(s *discordgo.Session, m *discordgo.MessageCreate) (bool, string) {
		content := strings.ToLower(m.Content)
		if strings.Contains(content, "discord.gg/") || strings.Contains(content, "discord.com/invite/") {
			return true, "Invite links are not allowed"
		}
		return false, ""
	})

	// Rule 2: Block Scam Keywords
	am.AddRule(func(s *discordgo.Session, m *discordgo.MessageCreate) (bool, string) {
		content := strings.ToLower(m.Content)
		badWords := []string{"free nitro", "steam gift", "claim crypto"}

		for _, word := range badWords {
			if strings.Contains(content, word) {
				return true, "Prohibited scam phrase detected"
			}
		}
		return false, ""
	})
}