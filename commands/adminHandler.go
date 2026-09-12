package commands

import (
	"fmt"
	"github.com/k3ptok/BasicDiscordBot/cmdctx"
)

func handleKick(ctx *cmdctx.Context) error {
	// Subcommand options are located at Options[0].Options
	target := ctx.Data.Options[0].Options[0].UserValue(ctx.Session)
	return ctx.Respond(fmt.Sprintf("🔨 Successfully kicked %s (simulation).", target.Mention()))
}