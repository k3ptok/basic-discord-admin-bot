package cmdctx

import (
	"log/slog"
	"github.com/bwmarrin/discordgo"
	//"github.com/k3ptok/BasicDiscordBot/logger"
)

type Context struct {
	Session 	*discordgo.Session
	Interaction *discordgo.InteractionCreate
	Data 		 discordgo.ApplicationCommandInteractionData
	Logger		*slog.Logger
}

func New(s *discordgo.Session, i *discordgo.InteractionCreate, logger *slog.Logger) (*Context, bool) {
	data, ok := i.Data.(discordgo.ApplicationCommandInteractionData)
	if !ok {
		return nil, false
	}

	cmdLogger := logger.With(
		slog.String("command", data.Name),
		slog.String("user_id", i.Member.User.ID),
		slog.String("guild_id", i.GuildID),
	)
	return &Context{
		Session:	 s,
		Interaction: i,
		Data:		 data,
		Logger:		 cmdLogger,
	}, true
}

func (ctx *Context) Respond(content string) error {
	return ctx.Session.InteractionRespond(ctx.Interaction.Interaction, &discordgo.InteractionResponse{
		Type:	discordgo.InteractionResponseChannelMessageWithSource,
		Data:	&discordgo.InteractionResponseData{
			Content:	content,
		},
	})
}

