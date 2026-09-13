package cmdctx

import (
	"log/slog"
	"github.com/bwmarrin/discordgo"
	//"github.com/k3ptok/BasicDiscordBot/logger"
)

type Context struct {
	Session 		*discordgo.Session
	Interaction 	*discordgo.InteractionCreate
	Data 		 	discordgo.ApplicationCommandInteractionData
	Logger			*slog.Logger
	Subcommand		string
	SubcommandGroup	string
	optionsMap		map[string]*discordgo.ApplicationCommandInteractionDataOption
}

// New creates a wrapped context with option pre-parsing and scoped logging
func New(s *discordgo.Session, i *discordgo.InteractionCreate, logger *slog.Logger) (*Context, bool) {
	data, ok := i.Data.(discordgo.ApplicationCommandInteractionData)
	if !ok {
		return nil, false
	}

	ctx := &Context{
		Session:		s,
		Interaction:	i,
		Data:			data,
		optionsMap:		make(map[string]*discordgo.ApplicationCommandInteractionDataOption),
	}

	// Walk options tree once to extract Subcommand name and flatten parameters
	ctx.parseOptions(data.Options)

	// Determine executor username safely for DM or Guild contexts
	var executorUsername string
	if i.Member != nil && i.Member.User != nil {
    	executorUsername = i.Member.User.Username
	} else if i.User != nil { // For DMs / user context
    	executorUsername = i.User.Username
	}

	//Create scoped logger with interaction metadata
	ctx.Logger = logger.With(
		slog.String("command", data.Name),
		slog.String("Subcommand", ctx.Subcommand),
		slog.String("executor_id", i.Member.User.ID),
		slog.String("executor_name", executorUsername),
		slog.String("guild_id", i.GuildID),
	)
	return ctx, true
}
// Recursively walks options to map subcommands and leaf arguments
func (c *Context) parseOptions(options []*discordgo.ApplicationCommandInteractionDataOption) {
	for _, opt := range options {
		switch opt.Type {
		case discordgo.ApplicationCommandOptionSubCommandGroup:
			c.SubcommandGroup = opt.Name
			c.parseOptions(opt.Options)
		case discordgo.ApplicationCommandOptionSubCommand:
			c.Subcommand = opt.Name
			c.parseOptions(opt.Options)
		default:
			//Leaf parameters (String, User, Int, Bool, etc.)
		}
	}
}



