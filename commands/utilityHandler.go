package commands

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/k3ptok/BasicDiscordBot/cmdctx"
)

func HandlePing(ctx *cmdctx.Context) error {
	// returns how long it took for discord to reply to ping
	latency := ctx.Session.HeartbeatLatency()

	embed := &discordgo.MessageEmbed{
		Title:       "🏓 Pong!",
		Description: fmt.Sprintf("Latency: **%v**", latency),
		Color:       0x2ECC71, // Green
	}
	
	// respond so everyone can see it
	return ctx.RespondEmbed(embed)
}

func HandleTag(ctx *cmdctx.Context) error {
	tagName, ok := ctx.GetString("name")
	if !ok {
		return ctx.RespondEphemeral("❌ Please select a tag from the menu.")
	}

	var response string

	// A simple switch statement handles the static tags
	switch tagName {
	case "rules":
		response = "📖 **Server Rules:** Follow the rules we definitely have in the rules channel we definitely have ##TO-DO##"
	default:
		return ctx.RespondEphemeral("❌ Unknown tag.")
	}

	// Tags should be public so everyone can see the answer
	return ctx.Respond(response)
}