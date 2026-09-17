package main

import (
	"embed"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	"github.com/k3ptok/BasicDiscordBot/automod"
	"github.com/k3ptok/BasicDiscordBot/cmdctx"
	"github.com/k3ptok/BasicDiscordBot/commands"
	"github.com/k3ptok/BasicDiscordBot/internal/database"
	"github.com/k3ptok/BasicDiscordBot/leveling"
	"github.com/k3ptok/BasicDiscordBot/logger"
	
)

//go:embed sql/schema/*.sql
var embedMigrations embed.FS

const DiscordScamFeedURL = "https://raw.githubusercontent.com/Phishing-Database/Phishing.Database/master/phishing-domains-ACTIVE.txt"

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

	
	db := database.InitDB(log, embedMigrations)
	dbQueries := database.New(db)

	xpManager := leveling.NewManager(dbQueries, log)
	modLogger := automod.NewModLogger(log, dbQueries)
	autoMod := automod.NewManager(log, modLogger)
	domainStore := automod.NewDomainStore(log)
	domainStore.StartAutoUpdater("banned-domains.txt", DiscordScamFeedURL, 6*time.Hour)

	allCommands := []commands.Command{
		commands.NewAdminCommandStructure(),
		//commands.NewUserCommandStructure(),
		commands.NewPingCommand(),
		commands.NewTagCommand(),
		commands.NewRankCommand(),
		commands.NewLeaderboardCommand(),
		commands.NewHelpCommand(),
	}

	commandMap := make(map[string]commands.Command)
	for _, cmd := range allCommands {
		commandMap[cmd.Definition.Name] = cmd
	}

	RegisterRules(autoMod, domainStore)
	discordSession.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if i.Type != discordgo.InteractionApplicationCommand {
			return
		}

		if ctx, ok := cmdctx.New(s, i, log, dbQueries, modLogger); ok {
			if cmd, exists := commandMap[ctx.Data.Name]; exists {
				commands.PanicWrapper(cmd, ctx)
			}
			log.Info("Received command interaction", slog.String("command", ctx.Data.Name))
		}

	})

	discordSession.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		autoMod.ProcessMessage(s, m)
		xpManager.ProcessMessage(s, m)
	})

	//set up discord intents
	discordSession.Identify.Intents = discordgo.IntentsGuildMessages |
		discordgo.IntentsGuilds |
		discordgo.IntentsMessageContent |
		discordgo.IntentsGuildScheduledEvents

	//start discord bot session
	err = discordSession.Open()
	if err != nil {
		log.Error("Failed to open connection to Discord", slog.Any("error", err))
		return
	}
	defer func() { //delay shutdown process until everything has finished cycling
		log.Info("Closing Discord Connection...")
		discordSession.Close()
	}()

	//***=== Command Registration ===***//

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
	//Overwrite currently loaded LOCAL commands on server to remove old/deleted commands
	registeredCmds, err := discordSession.ApplicationCommandBulkOverwrite(appID, guildID, definitions)
	if err != nil {
		log.Error("Failed to bulk overwrite commands", slog.Any("error", err))
	} else {
		log.Info("Successfully synced slash commands", slog.Int("Count:", len(registeredCmds)))
	}
	slog.Info("Discord Bot is currently running. Press CTRL + C to terminate session.")

	//***=== Shutdown Signal ===***//

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	slog.Warn("Termination signal recieved. Shutting down...")

}
