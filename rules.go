package main

import (
	"bufio"
	"github.com/bwmarrin/discordgo"
	"github.com/k3ptok/BasicDiscordBot/automod"
	"log/slog"
	"net/url"
	"os"
	"strings"
)

var (
	//Trusted domains. Should Not be flagged.
	legitDomains = map[string]struct{}{
		"discord.com":         {},
		"discord.gg":          {},
		"discordapp.com":      {},
		"discord.media":       {},
		"steampowered.com":    {},
		"steamcommunity.com":  {},
		"play.google.com":     {},
		"store.epicgames.com": {},
		"roblox.com":          {},
	}

	//Common name impersonations of major game clients
	brandLookalikes = []string{
		// Discord / Nitro
		"discoord", "discorb", "dlscord", "discort", "discorx", "discord-",
		"dlscordapp", "discrod", "discordgift", "discorid", "discordi",

		// Steam / Valve
		"steancommunity", "steamcomminuty", "steamcommunitu", "stearmcommunity",
		"steamcammunity", "steampowerd", "steam-promo", "steam-trade", "steamn",

		// Epic Games & Roblox (Common targets for young/mobile gamers)
		"epicgams", "epicgamse", "robloox", "roblx", "robux-",
	}

	//Common scam keywords
	scamBaitKeywords = []string{
		"nitro", "free-nitro", "steam-gift", "tradeoffer", "free-robux",
		"airdrop", "apk-mod", "unlimited-coins", "mod-apk", "free-gems",
	}

	// domain extensions commonly used by scammers
	susExtensions = []string{
		".xyz", ".top", ".info", ".site", ".ru", ".click", ".online",
		".gift", ".cfd", ".shop", ".zip", ".mov",
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

// Attach rules to automod
func RegisterRules(am *automod.Manager, store *automod.DomainStore, modLogger *automod.ModLogger) {

	am.AddRule(func(s *discordgo.Session, m *discordgo.MessageCreate) (bool, string) {
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
				_ = s.ChannelMessageDelete(m.ChannelID, m.ID)

				// Log action to log channel
				modLogger.LogViolation(s, m, "Blacklisted Scam Domain", hostname)
				return true, "Flagged by scam domain blocklist"
			}

			// catch brand mimic domains
			for _, match := range brandLookalikes {
				if strings.Contains(hostname, match) {
					_ = s.ChannelMessageDelete(m.ChannelID, m.ID)

					//Log action
					modLogger.LogViolation(s, m, "Company name impersonation", hostname)
					return true, "Flagged by domain impersonation check"
				}
			}

			//catch keywords paired with sus domain extensions
			for _, keyword := range scamBaitKeywords {
				if strings.Contains(hostname, keyword) {
					for _, tld := range susExtensions {
						if strings.HasSuffix(hostname, tld) {
							_ = s.ChannelMessageDelete(m.ChannelID, m.ID)

							//Log action
							modLogger.LogViolation(s, m, "scam keyword paired with sus domain extension", hostname)
							return true, "flagged keyword + sus domain extension"
						}
					}
				}
			}
		}
		return false, ""
	})

}
