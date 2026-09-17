package commands

import (
	"github.com/bwmarrin/discordgo"
	"github.com/k3ptok/BasicDiscordBot/cmdctx"
)

func NewHelpCommand() Command {
	return Command{
		Definition:	&discordgo.ApplicationCommand{
			Name:			"help",
			Description:	"View features and available commands",
		},
		Action:	handleHelp,
	}
}

func handleHelp(ctx *cmdctx.Context) error {
	embed := &discordgo.MessageEmbed{
		Title:       "🤖 Server Bot Guide",
		Description: "Welcome! As you chat in the server, you will naturally earn XP and level up. *(Note: XP can only be earned once every 15 seconds to prevent spam).* \n\nHere are the commands you can use:",
		Color:       0x3498DB, // Blue
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "📊 Leveling & Ranks",
				Value:  "`/rank [user]` - View your current level and XP progress.\n`/leaderboard` - See the top-ranked members in the server.",
				Inline: false,
			},
			{
				Name:   "🛠️ Utilities",
				Value:  "`/ping` - Check the bot's current response time.\n`/tag <name>` - Recall a saved text snippet or guide.",
				Inline: false,
			},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Server Moderation and AutoMod systems are running in the background.",
		},
	}

	return ctx.RespondEmbed(embed)
}