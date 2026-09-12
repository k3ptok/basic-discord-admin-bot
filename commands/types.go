package commands

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/k3ptok/BasicDiscordBot/cmdctx"
	
)

type SubCommandHandler func(ctx *cmdctx.Context) error

type Command struct {
	Definition *discordgo.ApplicationCommand
	SubCommands map[string]SubCommandHandler
}
// Route incoming interactions to appropriate subcommand
func (c *Command) Execute(ctx *cmdctx.Context) error {
	if len(ctx.Data.Options) == 0 {
		return fmt.Errorf("No command provided")
	}

	subName := ctx.Data.Options[0].Name
	if handler, ok := c.SubCommands[subName]; ok {
		return handler(ctx)
	}

	return fmt.Errorf("Unknown subcommand: %s", subName)
}