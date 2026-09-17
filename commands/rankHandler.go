package commands

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/k3ptok/BasicDiscordBot/internal/database"
	"github.com/k3ptok/BasicDiscordBot/cmdctx"
	"github.com/k3ptok/BasicDiscordBot/leveling"
)

func HandleRank(ctx *cmdctx.Context) error {
	// defer to prevent timeout from discords API
	if err := ctx.Defer(false); err != nil {
		return err
	}

	targetUser, ok := ctx.GetUser("user")
	if !ok {
		// Fallback to the person executing the command
		if ctx.Interaction.Member != nil && ctx.Interaction.Member.User != nil {
			targetUser = ctx.Interaction.Member.User
		} else if ctx.Interaction.User != nil {
			targetUser = ctx.Interaction.User
		} else {
			return fmt.Errorf("unable to determine user")
		}
	}

	stats, err := ctx.DB.GetUserRank(context.Background(), database.GetUserRankParams{
		GuildID: ctx.Interaction.GuildID,
		UserID:  targetUser.ID,
	})

	if err != nil {
		if err == sql.ErrNoRows {
			stats = database.GetUserRankRow{Level: 0, Xp: 0} // Default for new users
		} else {
			ctx.Logger.Error("Failed to fetch user rank", slog.Any("error", err))
			return ctx.EditFollowup("Something went wrong fetching rank data.") 
		}
	}

	
	nextLevelXP := leveling.CalculateNextLevelXP(stats.Level)

	embed := &discordgo.MessageEmbed{
		Title: fmt.Sprintf("📊 Rank: %s", targetUser.Username),
		Color: 0x00FF00,
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: targetUser.AvatarURL(""),
		},
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Current Level",
				Value:  fmt.Sprintf("**%d**", stats.Level),
				Inline: true,
			},
			{
				Name:   "Total XP",
				Value:  fmt.Sprintf("**%d** / %d", stats.Xp, nextLevelXP),
				Inline: true,
			},
		},
	}

	return ctx.EditFollowupEmbed(embed)
}

func HandleLeaderboard(ctx *cmdctx.Context) error {
	// defer to avoid timeout
	if err := ctx.Defer(false); err != nil {
		return err
	}
	// get current leaderboard rankings
	ranks, err := ctx.DB.GetLeaderboard(context.Background(), ctx.Interaction.GuildID)
	if err != nil {
		ctx.Logger.Error("Failed to retrieve ranks from database", slog.Any("error", err))
		return ctx.EditFollowup("Failed to retrieve ranks from database")
	}

	if len(ranks) == 0 {
		return ctx.EditFollowup("Leaderboard is empty!")
	}

	var description strings.Builder
	for i, user := range ranks {
		medal := ""
		switch i {
		case 0:
			medal = "🥇"
		case 1:
			medal = "🥈"
		case 2:
			medal = "🥉"
		default:
			medal = fmt.Sprintf("**#%d**", i+1)
		}

		// <@UserID> makes Discord natively render the user's name!
		description.WriteString(fmt.Sprintf("%s <@%s> — **Level %d** (%d XP)\n\n", medal, user.UserID, user.Level, user.Xp))
	}

	// 3. Build the embed
	embed := &discordgo.MessageEmbed{
		Title:       "🏆 Server Leaderboard",
		Description: description.String(),
		Color:       0xFFD700, // Gold color
	}

	// 4. Send using your clean embed helper
	return ctx.EditFollowupEmbed(embed)
}
