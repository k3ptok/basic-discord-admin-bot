package cmdctx

import (
	//"fmt"
	"github.com/bwmarrin/discordgo"
)

type Context struct {
	Session 	*discordgo.Session
	Interaction *discordgo.InteractionCreate
	Data 		 discordgo.ApplicationCommandInteractionData
}

func New(s *discordgo.Session, i *discordgo.InteractionCreate) (*Context, bool) {
	data, ok := i.Data.(discordgo.ApplicationCommandInteractionData)
	if !ok {
		return nil, false
	}
	return &Context{
		Session:	 s,
		Interaction: i,
		Data:		 data,
	}, true
}

func (ctx *Context) RespondText(content string) error {
	return ctx.Session.InteractionRespond(ctx.Interaction.Interaction, &discordgo.InteractionResponse{
		Type:	discordgo.InteractionResponseChannelMessageWithSource,
		Data:	&discordgo.InteractionResponseData{
			Content:	content,
		},
	})
}

func (ctx *Context) Defer(ephemeral bool) error {
	var flags discordgo.MessageFlags
	if ephemeral {
		flags = discordgo.MessageFlagsEphemeral
	}
	return ctx.Session.InteractionRespond(ctx.Interaction.Interaction, &discordgo.InteractionResponse{
		Type:	discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data:	&discordgo.InteractionResponseData{
			Flags: flags,
		},
	})
}