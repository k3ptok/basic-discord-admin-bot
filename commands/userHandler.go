package commands

import (
	"fmt"
	"github.com/k3ptok/BasicDiscordBot/cmdctx"
)

func handleInfo(ctx *cmdctx.Context) error {
	user := ctx.Interaction.Member.User
	return ctx.Respond(fmt.Sprintf("Hello %s! Your ID is `%s`.", user.Username, user.ID))
}