package main

import (
	"os"
	"os/signal"
	"syscall"
	"io"
	"log/slog"
	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	"gopkg.in/natefinch/lumberjack.v2"
)

func main() {

	// Configure the rotating log file inside the "logs" directory
	logRotator := &lumberjack.Logger{
		Filename:   "logs/bot.log", // Puts bot.log and all rotated backups inside /logs
		MaxSize:    10,             // Megabytes before rotating
		MaxBackups: 20,              // Number of old log files to keep
		MaxAge:     180,             // Number of days to keep old files
		Compress:   true,           // Compresses old logs to save disk space
	}

	// Configure logging to write to both the console (Stdout) AND the log rotator
	multiWriter := io.MultiWriter(os.Stdout, logRotator)
	logger := slog.New(slog.NewTextHandler(multiWriter, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	err := godotenv.Load()
  	if err != nil {
    	slog.Error("Failed to load godotenv", "error", err)
  	}

	botToken := os.Getenv("DISCORD_KEY")
	//Load godotenv and grab various auth codes from .env. If anyone actually downloads this, you will need to make/get your own from the discord developer portal
	//and your desired discord server

	//Load JSON config to set up notification channels from before shutdown
	LoadBannedDomains()

	slog.Info("Starting Discord Bot Session...")

	discordSession, err := discordgo.New("Bot " + botToken)
	if err != nil {
		slog.Error("Failed to create session", "error", err)
		return
	}

	//Register handlers
	


	discordSession.Identify.Intents = discordgo.IntentsGuildMessages |
									  discordgo.IntentsGuilds | 
									  discordgo.IntentsMessageContent |
									  discordgo.IntentsGuildScheduledEvents
	// Set up bot Intents
	 

	err = discordSession.Open()
	if err != nil {
		slog.Error("Failed to open connection to Discord", "error",err)
		return
	}
	defer func() {
		slog.Info("Closing Discord Connection...")
		discordSession.Close()
	}()

	slog.Info("Discord Bot is currently running. Press CTRL + C to terminate session.")
	
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	slog.Warn("Termination signal recieved. Shutting down...")

}