package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"io"
	"log/slog"
	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

func main() {

	//Create log file
	logFile, err := os.OpenFile("bot.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		fmt.Println("failed to create log file", err)
		return
	}
	defer logFile.Close()

	//Configure logging to write to both the console (Stdout) AND the logfile
	multiWriter := io.MultiWriter(os.Stdout, logFile)
	logger := slog.New(slog.NewTextHandler(multiWriter, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	err = godotenv.Load()
  	if err != nil {
    	slog.Error("Failed to load godotenv", "error", err)
  	}

	botToken := os.Getenv("DISCORD_KEY")
	//Load godotenv and grab various auth codes from .env. If anyone actually downloads this, you will need to make/get your own from the discord developer portal
	//and your desired discord server

	//Load JSON config to set up notification channels from before shutdown
	LoadConfig()

	slog.Info("Starting Discord Bot Session...")

	discordSession, err := discordgo.New("Bot " + botToken)
	if err != nil {
		slog.Error("Failed to create session", "error", err)
		return
	}

	//Register handlers
	discordSession.AddHandler(InteractionHandler)
	discordSession.AddHandler(onConnect)
	discordSession.AddHandler(onDisconnect)
	discordSession.AddHandler(onMessageCreate)
	discordSession.AddHandler(guildScheduledEventCreate)


	discordSession.Identify.Intents = discordgo.IntentsGuildMessages | discordgo.IntentsGuilds | discordgo.IntentsGuildScheduledEvents
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

	// Register name spaced commands
	RegisterNamespacedCommand(discordSession)

	// Start global reminder scheduler
	StartGlobalReminderScheduler(discordSession)

	slog.Info("Discord Bot is currently running. Press CTRL + C to terminate session.")
	
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	slog.Warn("Termination signal recieved. Shutting down...")

}