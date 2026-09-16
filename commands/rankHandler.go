package commands

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/bwmarrin/discordgo"
	"github.com/k3ptok/BasicDiscordBot/internal/database"
	"github.com/k3ptok/BasicDiscordBot/cmdctx"
	"github.com/k3ptok/BasicDiscordBot/leveling"
)

func HandleRank(ctx *cmdctx.Context) error {
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
			return ctx.RespondEphemeral("❌ Something went wrong fetching rank data.") 
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

	return ctx.RespondEmbed(embed)
}
