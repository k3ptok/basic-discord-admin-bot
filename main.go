package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"

	"github.com/k3ptok/BasicDiscordBot/cmdctx"
	"github.com/k3ptok/BasicDiscordBot/commands"
	"github.com/k3ptok/BasicDiscordBot/logger"
)

var allCommands = []commands.Command{
	commands.Admin,
	commands.User,
}

func main() {
	log := logger.InitLogger()
	log.Info("Starting discord bot spinup cycle...")

	err := godotenv.Load()
	if err != nil {
		slog.Error("Failed to load godotenv", "error", err)
	}

	botToken := os.Getenv("DISCORD_KEY")
	appID := os.Getenv("APPLICATION_ID")
	guildID := os.Getenv("SERVER_ID")
	//Load godotenv and grab various auth codes from .env. If anyone actually downloads this, you will need to make/get your own from the discord developer portal
	//and your desired discord server

	discordSession, err := discordgo.New("Bot " + botToken)
	if err != nil {
		log.Error("Failed to create session", slog.Any("error", err))
		os.Exit(1)
	}

	//Index commands by top level
	commandMap := make(map[string]commands.Command)
	for _, cmd := range allCommands {
		commandMap[cmd.Definition.Name] = cmd
	}

	discordSession.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if i.Type != discordgo.InteractionApplicationCommand {
			return
		}

		ctx, ok := cmdctx.New(s, i, log)
		if !ok {
			return
		}
		if cmd, exists := commandMap[ctx.Data.Name]; !exists {
			if err := cmd.Execute(ctx); err != nil {
				log.Error("Error executing command", slog.Any("error", err))
				return
			}
		}
		log.Info("Received command interaction", slog.String("command", ctx.Data.Name))
	})

	discordSession.Identify.Intents = discordgo.IntentsGuildMessages |
		discordgo.IntentsGuilds |
		discordgo.IntentsMessageContent |
		discordgo.IntentsGuildScheduledEvents
	// Set up bot Intents

	err = discordSession.Open()
	if err != nil {
		slog.Error("Failed to open connection to Discord", "error", err)
		return
	}
	defer func() {
		slog.Info("Closing Discord Connection...")
		discordSession.Close()
	}()

	slog.Info("Registering slash commands...")
	for _, cmd := range allCommands {
		_, err := discordSession.ApplicationCommandCreate(appID, guildID, cmd.Definition)
		if err != nil {
			slog.Error("Cannot create command", "command", "error", cmd.Definition.Name, err)
		}
	}

	slog.Info("Discord Bot is currently running. Press CTRL + C to terminate session.")

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	slog.Warn("Termination signal recieved. Shutting down...")

}
