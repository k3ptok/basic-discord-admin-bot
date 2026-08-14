package main

import (
	"encoding/json"
	"os"
	"log/slog"
)

const configFile = "config.json" 

type ServerSettings struct {
	ChannelID string   `json:"channel_id"`
	Timings   []string `json:"timings"`
}

func SaveConfig() {
	configMu.RLock()
	defer configMu.RUnlock()

	data, err := json.MarshalIndent(botConfig, "", " ")
	if err != nil {
		slog.Error("Failed to marshal config to JSON", "error", err)
		return
	}

	err = os.WriteFile(configFile, data, 0644)
	if err != nil {
		slog.Error("Failed to write config file", "error", err)
		return
	}

}

//Read the json file back into config when the bot starts
func LoadConfig() {
	configMu.Lock()
	defer configMu.Unlock()

	// if the file doesn't exist (first run), skip loading
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		return
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		slog.Error("Failed to read config file", "error", err)
		return
	}

	err = json.Unmarshal(data, &botConfig)
	if err != nil {
		slog.Error("Failed to parse config", "error", err)
		return
	}
}