package commands

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/k3ptok/BasicDiscordBot/cmdctx"
	
)

type SubCommandHandler func(ctx *cmdctx.Context) error

type Command struct {
	Definition *discordgo.ApplicationCommand
	Action      func(ctx *cmdctx.Context) error    // NEW: For top-level commands (like /ping)
	SubCommands map[string]SubCommandHandler
}
// Route incoming interactions to appropriate subcommand
func (c *Command) Execute(ctx *cmdctx.Context) error {
	if c.Action != nil {
		return c.Action(ctx)
	}
	if handler, ok := c.SubCommands[ctx.Subcommand]; ok {
		return handler(ctx)
	}
	return fmt.Errorf("unknown subcommand: %s", ctx.Subcommand)
}