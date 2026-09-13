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

		if ctx, ok := cmdctx.New(s, i, log); ok {
			if cmd, exists := commandMap[ctx.Data.Name]; exists {
				commands.PanicWrapper(cmd, ctx)
			}
			log.Info("Received command interaction", slog.String("command", ctx.Data.Name))
		}
		
	})

	discordSession.Identify.Intents = discordgo.IntentsGuildMessages |
		discordgo.IntentsGuilds |
		discordgo.IntentsMessageContent |
		discordgo.IntentsGuildScheduledEvents
	// Set up bot Intents

	err = discordSession.Open()
	if err != nil {
		log.Error("Failed to open connection to Discord", slog.Any("error", err))
		return
	}
	defer func() {
		log.Info("Closing Discord Connection...")
		discordSession.Close()
	}()

	log.Info("Registering slash commands...")
	//Load active commands into slice
	var definitions []*discordgo.ApplicationCommand
	for _, cmd := range allCommands {
		definitions = append(definitions, cmd.Definition)
	}
	//Clear any global commands I may have set
	_, err = discordSession.ApplicationCommandBulkOverwrite(appID, "", []*discordgo.ApplicationCommand{})
	if err != nil {
		log.Error("Failed to clear global commands", slog.Any("error", err))
	} else {
		log.Info("Successfully cleared legacy global commands")
}
	//Overwrite currently loaded LOCAL commands on server to audit old/deleted commands
	registeredCmds, err := discordSession.ApplicationCommandBulkOverwrite(appID, guildID, definitions)
	if err != nil {
		log.Error("Failed to bulk overwrite commands", slog.Any("error", err))
	} else {
		log.Info("Successfully synced slash commands", slog.Int("Count:", len(registeredCmds)))
	}
	slog.Info("Discord Bot is currently running. Press CTRL + C to terminate session.")

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	slog.Warn("Termination signal recieved. Shutting down...")

}
