package commands

import (
	"fmt"
	"github.com/k3ptok/BasicDiscordBot/cmdctx"
	"github.com/bwmarrin/discordgo"
)

func handleInfo(ctx *cmdctx.Context) error {
	var target *discordgo.User

	// 1. Try to get the target parameter
	if user, ok := ctx.GetUser("target"); ok { // Make sure this matches the name in Definition!
		target = user
	} else {
		// 2. Fall back to the user who executed the command
		if ctx.Interaction.Member != nil && ctx.Interaction.Member.User != nil {
			target = ctx.Interaction.Member.User
		} else {
			target = ctx.Interaction.User // For DM interactions
		}
	}

	return ctx.Respond(fmt.Sprintf("👤 **User Info:** %s (ID: `%s`)", target.Username, target.ID))
}