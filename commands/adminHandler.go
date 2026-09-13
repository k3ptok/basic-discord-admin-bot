package commands

import (
	"fmt"
	"github.com/k3ptok/BasicDiscordBot/cmdctx"
	"github.com/bwmarrin/discordgo"
	"log/slog"
)

func handleKick(ctx *cmdctx.Context) error {
	// 1. Extract required target user
	target, ok := ctx.GetUser("target")
	if !ok {
		return ctx.Respond("❌ You must specify a target user to kick.")
	}

	// Prevent kicking self or bot
	if target.ID == ctx.Interaction.Member.User.ID {
		return ctx.Respond("❌ You cannot kick yourself.")
	}

	// 2. Extract optional reason (default to fallback if empty)
	reason, ok := ctx.GetString("reason")
	if !ok || reason == "" {
		reason = "No reason provided."
	}

	// Construct full audit log message showing who initiated the action
	auditReason := fmt.Sprintf("Kicked by %s: %s", ctx.Interaction.Member.User.Username, reason)

	// 3. Perform Discord API Kick Action
	err := ctx.Session.GuildMemberDeleteWithReason(
		ctx.Interaction.GuildID,
		target.ID,
		auditReason,
	)
	if err != nil {
		ctx.Logger.Error("Failed to kick member",
			slog.String("target_id", target.ID),
			slog.Any("error", err),
		)
		return ctx.Respond(fmt.Sprintf("❌ Failed to kick **%s**: `%s`", target.Username, err.Error()))
	}

	ctx.Logger.Info("Member kicked successfully",
		slog.String("target_id", target.ID),
		slog.String("moderator_id", ctx.Interaction.Member.User.ID),
		slog.String("reason", reason),
	)

	// 4. Return formatted response
	embed := &discordgo.MessageEmbed{
		Title:       "🔨 Member Kicked",
		Color:       0xE74C3C, // Red
		Description: fmt.Sprintf("**Target:** %s (`%s`)\n**Reason:** %s", target.Mention(), target.ID, reason),
		Footer: &discordgo.MessageEmbedFooter{
			Text: fmt.Sprintf("Actioned by %s", ctx.Interaction.Member.User.Username),
		},
	}

	return ctx.RespondEmbed(embed)
}