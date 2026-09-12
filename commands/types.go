package commands

import (
	"github.com/bwmarrin/discordgo"
	"github.com/k3ptok/context/cmdctx"
)

type SubCommandHandler func(ctx *Context) error

type Command struct {
	Definition *discordgo.ApplicationCommand
	SubCommands map[string]SubCommandHandler
}
// Route incoming interactions to appropriate subcommand
func (c *Command) Handle(s *discordgo.Session, i *discordgo.InteractionCreate) {
	options := i.ApplicationCommandData().Options
	if len(options) == 0 {
		return
	}

	//determine subcommand name
	subcommandName := options[0].Name
	if handler, exists := c.SubCommands[subcommandName]; exists {
		handler(s, i)
	}
}