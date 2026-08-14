package main

import (
	"encoding/json"
	"os"
	"log/slog"
)

const configFile = "config.json"

func SaveConfig() {
	configMu.RLock()
	defer configMu.RUnlock()

	data, err := json.MarshalIndent(customChannels, "", " ")
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
	configMu.RLock()
	defer configMu.RUnlock()

	// if the file doesn't exist (first run), skip loading
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		return
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		slog.Error("Failed to read config file", "error", err)
		return
	}

	err = json.Unmarshal(data, &customChannels)
	if err != nil {
		slog.Error("Failed to parse config", "error", err)
		return
	}
}