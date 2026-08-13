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
	
	err := godotenv.Load()
  	if err != nil {
    	log.Fatal("Error loading .env file")
  	}

	botToken := os.Getenv("DISCORD_KEY")
	//Load godotenv and grab the bot token from a .env file. If anyone actually downloads this, you will need to make your own from the discord developer portal.

	//Create log file
	logFile, err := os.OpenFile("bot.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
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


	slog.Info("Starting Discord Bot Session...")

	discordSession, err := discordgo.New("Bot " + botToken)
	if err != nil {
		slog.Error("Failed to create session", "error", err)
		return
	}

	//Register handlers
	dg.AddHandler(onConnect)
	dg.AddHandler(onDisconnect)
	dg.AddHandler(onMessageCreate)

	discordSession.Identify.Intents = discordgo.IntentsGuildMessages
	// Tell the bot to listen to messages

	err := discordSession.Open()
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