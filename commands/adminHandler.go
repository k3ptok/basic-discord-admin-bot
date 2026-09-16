package commands

import (
	"context"
	"time"
	"fmt"
	"log/slog"
	"database/sql"
	"github.com/bwmarrin/discordgo"
	"github.com/k3ptok/BasicDiscordBot/cmdctx"
	"github.com/k3ptok/BasicDiscordBot/internal/database"
)
// Kick a user using slash commands
func handleKick(ctx *cmdctx.Context) error {
	// extract required target user
	target, ok := ctx.GetUser("target")
	if !ok {
		return ctx.Respond("❌ You must specify a target user to kick.")
	}

	// Prevent kicking self or bot
	if target.ID == ctx.Interaction.Member.User.ID {
		return ctx.Respond("❌ You cannot kick yourself.")
	}

	// Extract optional reason (default to fallback if empty)
	reason, ok := ctx.GetString("reason")
	if !ok || reason == "" {
		reason = "No reason provided."
	}

	// Attempt to DM BEFORE kicking
	userChannel, err := ctx.Session.UserChannelCreate(target.ID)
	dmStatus := "✅ User was notified via DM."

	if err != nil {
		dmStatus = "⚠️ Could not open DM with user."
	} else {
		dmMsg := fmt.Sprintf("👢 You have been kicked from the server.\n**Reason:** %s", reason)
		_, err := ctx.Session.ChannelMessageSend(userChannel.ID, dmMsg)
		if err != nil {
			dmStatus = "⚠️ Could not DM user (server DMs disabled)."
		}
	}

	// execute the Kick in Discord
	err = ctx.Session.GuildMemberDeleteWithReason(ctx.Interaction.GuildID, target.ID, reason)
	if err != nil {
		ctx.Logger.Error("Failed to kick user", slog.Any("error", err))
		return ctx.RespondEphemeral("❌ Failed to kick the user. Check my permissions and role hierarchy.")
	}

	err = ctx.DB.InsertModLog(context.Background(), database.InsertModLogParams{
		GuildID:	ctx.Interaction.GuildID,
		UserID:		target.ID,
		Action:		"KICK",
		Reason:		sql.NullString{String: reason, Valid: true},
	})
	if err != nil {
		ctx.Logger.Error("Failed to log mod event to db", slog.Any("error", err))
		return ctx.RespondEphemeral("Failed to log mod event to db")
	}

	// Respond ephemerally to the admin
	embed := &discordgo.MessageEmbed{
		Title:       "👢 User Kicked",
		Color:       0xE67E22,
		Description: fmt.Sprintf("**Target:** %s (`%s`)\n**Reason:** %s\n\n*%s*", target.Mention(), target.ID, reason, dmStatus),
	}
	return ctx.RespondEmbedEphemeral(embed)
}

func handleTimeout(ctx *cmdctx.Context) error {
	target, ok := ctx.GetUser("user")
	if !ok || target == nil {
		return ctx.Respond("Must specify a target")
	}

	
	durationStr, ok := ctx.GetString("duration")
	if !ok {
		ctx.Respond("Must specify a valid time duration")
	}

	duration, err := time.ParseDuration(durationStr)
	if err != nil {
		return ctx.Respond("Invalid time format. Use '10m' or '1h' format")
	}

	// Discord enforces a maximum timeout time of 28 days
	if duration > 28*24*time.Hour {
		return ctx.Respond("Discord does not allow timeouts longer than 28 days")
	}

	reason, ok := ctx.GetString("reason")
	if !ok || reason == "" {
		reason = "No reason provided"
	}

	//calculate timeout end time
	until := time.Now().Add(duration)

	// Execute timeout
	err = ctx.Session.GuildMemberTimeout(ctx.Interaction.GuildID, target.ID, &until)
	if err != nil {
		ctx.Logger.Error("Failed totimeout user", slog.Any("error", err))
		return ctx.Respond("Failed to time out user. Check logs for info")
	}

	// log timeout event to db
	err = ctx.DB.InsertModLog(context.Background(), database.InsertModLogParams{
		GuildID: 	ctx.Interaction.GuildID,
		UserID:		target.ID,
		Action:		fmt.Sprintf("TIMEOUT (%s)", durationStr),
		Reason:		sql.NullString{String: reason, Valid: true},	
	})
	if err != nil {
		ctx.Logger.Error("Failed to log timeout event to db", slog.Any("error", err))
	}

	embed := &discordgo.MessageEmbed{
		Title:       "⏱️ User Timed Out",
		Color:       0xE67E22, // Orange
		Description: fmt.Sprintf("**Target:** %s (`%s`)\n**Duration:** %s\n**Reason:** %s", target.Mention(), target.ID, durationStr, reason),
	}
	return ctx.RespondEmbedEphemeral(embed)
}

func handleBan(ctx *cmdctx.Context) error {
	target, ok := ctx.GetUser("user")
	if !ok || target == nil {
		return ctx.RespondEphemeral("Must specify a target")
	}

	reason, ok := ctx.GetString("reason")
	if !ok || reason == "" {
		reason = "No reason provided"
	}

	// open dm with target
	userChannel, err := ctx.Session.UserChannelCreate(target.ID)
	dmStatus := "✅ User was notified via DM."

	if err != nil {
		dmStatus = "⚠️ Could not open DM with user."
	} else {
		dmMsg := fmt.Sprintf("🔨 You have been banned from the server.\n**Reason:** %s", reason)
		_, err := ctx.Session.ChannelMessageSend(userChannel.ID, dmMsg)
		if err != nil {
			dmStatus = "⚠️ Could not DM user (server DMs disabled)."
		}
	}

	// execute ban
	err = ctx.Session.GuildBanCreateWithReason(ctx.Interaction.GuildID, target.ID, reason, 0)
	if err != nil {
		ctx.Logger.Error("Failed to ban user", slog.Any("error", err))
		return ctx.RespondEphemeral("❌ Failed to ban the user. Check my permissions and role hierarchy.")
	}

	// log mod event in db
	err = ctx.DB.InsertModLog(context.Background(),database.InsertModLogParams{
		GuildID:	ctx.Interaction.GuildID,
		UserID: 	target.ID,
		Action:		"BAN",
		Reason:		sql.NullString{String: reason, Valid: true},
	})
	if err != nil {
		ctx.Logger.Error("Failed to log ban event to db", slog.Any("error", err))
	}

	embed := &discordgo.MessageEmbed{
		Title:       "🔨 User Banned",
		Color:       0xE74C3C, // Red
		Description: fmt.Sprintf("**Target:** %s (`%s`)\n**Reason:** %s\n\n*%s*", target.Mention(), target.ID, reason, dmStatus),
	}
	return ctx.RespondEmbedEphemeral(embed)
}

func handleUnban(ctx *cmdctx.Context) error {
	target, ok := ctx.GetUser("user")
	if !ok || target == nil {
		return ctx.RespondEphemeral("Must provide valid user ID to unban")
	}

	reason, ok := ctx.GetString("reason")
	if !ok || reason == "" {
		reason = "No reason provided"
	}

	// Execute the unban
	err := ctx.Session.GuildBanDelete(ctx.Interaction.GuildID, target.ID)
	if err != nil {
		ctx.Logger.Error("Failed to unban user", slog.Any("error", err))
		return ctx.RespondEphemeral("Failed to unban target. Check logs for info")
	}

	//Log to db
	err = ctx.DB.InsertModLog(context.Background(), database.InsertModLogParams{
		GuildID:	ctx.Interaction.GuildID,
		UserID:		target.ID,
		Action:		"UNBAN",
		Reason:		sql.NullString{String: reason, Valid: true},
	})
	if err != nil {
		ctx.Logger.Error("Failed to log unban event to db", slog.Any("error", err))
	}

	embed := &discordgo.MessageEmbed{
		Title:       "🔓 User Unbanned",
		Color:       0x2ECC71, // Green
		Description: fmt.Sprintf("**Target ID:** `%s`\n**Reason:** %s", target.ID, reason),
	}
	return ctx.RespondEmbedEphemeral(embed)

}

//Purge target users chat history within the scan limit entered
func handlePurgeUser(ctx *cmdctx.Context) error {
	target, ok := ctx.GetUser("user")
	if !ok || target == nil {
		return ctx.RespondEphemeral("Must specify a target")
	}

	// determine how far back to scan
	scanLimit := 50
	if limitOpt, ok := ctx.GetInt("scan-limit"); ok {
		if limitOpt > 100 {
			scanLimit = 100 // Discord's api limit for single fetch request
		} else if limitOpt > 0 {
			scanLimit = int(limitOpt)
		}
	}

	// fetch channels message history
	messages, err := ctx.Session.ChannelMessages(ctx.Interaction.ChannelID, scanLimit, "", "", "")
	if err != nil {
		ctx.Logger.Error("Failed to fetch channel messages", slog.Any("error", err))
		return ctx.RespondEphemeral("Failed to fetch channel history")
	}

	// filter collected messages for target user and messages younger than 14 days
	var messageIDs []string
	twoWeeksAgo := time.Now().Add(-14 * 24 * time.Hour)

	for _, m := range messages {
		if m.Author.ID == target.ID {
			if m.Timestamp.After(twoWeeksAgo) {
				messageIDs = append(messageIDs, m.ID)
			}
		}
	}

	if len(messageIDs) == 0 {
		return ctx.RespondEphemeral(fmt.Sprintf("⚠️ Found no recent deletable messages from %s in the last %d messages scanned.", target.Mention(), scanLimit))
	}

	if len(messageIDs) == 1 {
		// Discord's BulkDelete endpoint sometimes fails if there is only 1 message, so we use the single delete endpoint
		err = ctx.Session.ChannelMessageDelete(ctx.Interaction.ChannelID, messageIDs[0])
	} else {
		// Execute the Bulk Delete
		err = ctx.Session.ChannelMessagesBulkDelete(ctx.Interaction.ChannelID, messageIDs)
	}

	if err != nil {
		ctx.Logger.Error("Failed to delete messages", slog.Any("error", err))
		return ctx.RespondEphemeral("❌ Failed to delete messages.")
	}
	
	return ctx.RespondEphemeral(fmt.Sprintf("🧹 Successfully scrubbed **%d** messages from %s.", len(messageIDs), target.Mention()))
}

// Issue a logged warning to user
func handleWarn(ctx *cmdctx.Context) error {
	// Get target user
	target, ok := ctx.GetUser("user")
	if !ok || target == nil {
		return ctx.RespondEphemeral("❌ You must specify a target user to warn.")
	}
	//Get entered reason
	reason, ok := ctx.GetString("reason")
	if !ok || reason == "" {
		return ctx.RespondEphemeral("❌ A reason is required for a formal warning.")
	}
	// Save warning to database
	err := ctx.DB.InsertModLog(context.Background(), database.InsertModLogParams{
		GuildID: ctx.Interaction.GuildID,
		UserID:  target.ID,
		Action:  "WARN",
		Reason:  sql.NullString{String: reason, Valid: true},
	})
	if err != nil {
		ctx.Logger.Error("Failed to save warning to database", slog.Any("error", err))
		return ctx.RespondEphemeral("❌ An internal error occurred while saving the warning.")
	}

	//attempt to dm user
	userChannel, err := ctx.Session.UserChannelCreate(target.ID)
	dmStatus := "✅ User was notified via DM."

	if err != nil {
		// If we can't open a channel, they likely have the bot blocked
		dmStatus = "⚠️ Could not DM user (they may have DMs disabled)."
	} else {
		// send the warning to their dms
		dmMsg := fmt.Sprintf("⚠️ You have received a warning in the server.\n**Reason:** %s", reason)
		_, err := ctx.Session.ChannelMessageSend(userChannel.ID, dmMsg)
		
		if err != nil {
			// Sometimes opening the channel works, but sending fails due to privacy settings
			dmStatus = "⚠️ Could not DM user (they may have server DMs disabled)."
		}
	}

	// create message embed
	embed := &discordgo.MessageEmbed{
		Title:       "⚠️ User Warned",
		Color:       0xF1C40F,
		Description: fmt.Sprintf("**Target:** %s\n**Reason:** %s\n\n*%s*", target.Mention(), reason, dmStatus),
	}

	return ctx.RespondEmbedEphemeral(embed)


}

func handleHistory(ctx *cmdctx.Context) error {
	target, ok := ctx.GetUser("user")
	if !ok || target == nil {
		return ctx.RespondEphemeral("❌ You must specify a target user.")
	}

	logs, err := ctx.DB.GetUserModLogs(context.Background(), database.GetUserModLogsParams{
		GuildID:	ctx.Interaction.GuildID,
		UserID:		target.ID,
	})
	if err != nil {
		ctx.Logger.Error("Failed to retrieve users moderation history", slog.Any("error", err))
		return ctx.RespondEphemeral("❌ An internal error occurred while fetching history.")
	}
	if len(logs) == 0 {
		return ctx.RespondEphemeral(fmt.Sprintf("✅ **%s** has no moderation history.", target.Username))
	}
	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("📜 Infraction History: %s", target.Username),
		Color:       0x3498DB, // Blue
		Description: fmt.Sprintf("Found **%d** past infractions.", len(logs)),
	}
	for _, entry := range logs {
		reason := "No reason provided"
		if entry.Reason.Valid {
			reason = entry.Reason.String
		}
		dateStr := "Unknown Date"
		if entry.CreatedAt.Valid {
			dateStr = entry.CreatedAt.Time.Format("Jan 02, 2006")
		}
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   fmt.Sprintf("[%s] %s", dateStr, entry.Action),
			Value:  reason,
			Inline: false,
		})
	}
	return ctx.RespondEmbedEphemeral(embed)
}

func handleWhois(ctx *cmdctx.Context) error {
	target, ok := ctx.GetUser("user")
	if !ok || target == nil {
		return ctx.RespondEphemeral("Must specify a target")
	}

	member, err := ctx.Session.GuildMember(ctx.Interaction.GuildID, target.ID)
	if err != nil {
		ctx.Logger.Error("Failed to retrieve member data", slog.Any("error", err))
		return ctx.RespondEphemeral("Failed to retrieve member data")
	}

	// pull account creation data from snowflake
	accountCreated, err := discordgo.SnowflakeTimestamp(target.ID)
	createdStr := "Unknown"
	if err == nil {
		createdStr = accountCreated.Format("Jan 02, 2006 • 15:04 MST")
	}

	// format
	joinedStr := member.JoinedAt.Format("Jan 02, 2006 • 15:04 MST")

	
	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("🔍 Whois: %s", target.Username),
		Color:       0x9B59B6, // Purple
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: target.AvatarURL(""),
		},
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "ID",
				Value:  fmt.Sprintf("`%s`", target.ID),
				Inline: true,
			},
			{
				Name:   "Mention",
				Value:  target.Mention(),
				Inline: true,
			},
			{
				Name:   "Account Created",
				Value:  createdStr,
				Inline: false,
			},
			{
				Name:   "Joined Server",
				Value:  joinedStr,
				Inline: false,
			},
		},
	}

	return ctx.RespondEmbedEphemeral(embed)
}

// MakeLogChannelHandler returns a SubCommandHandler closure linked to your ModLogger
func MakeLogChannelHandler() SubCommandHandler {
	return func(ctx *cmdctx.Context) error {
		// Extract the "target" channel option from command context
		channel, ok := ctx.GetChannel("target")
		if !ok || channel == nil {
			return ctx.RespondEphemeral("❌ Invalid channel selected.")
		}

		err := ctx.DB.SetLogChannel(context.Background(), database.SetLogChannelParams{
			GuildID:		ctx.Interaction.GuildID,
			LogChannelID: 	channel.ID,
		})
		if err != nil {
			ctx.Logger.Error("Failed to save log channel to database", slog.Any("error", err))
			return ctx.RespondEphemeral("❌ Failed to save configuration to the database.")
		}

		ctx.ModLogger.ClearCache(ctx.Interaction.GuildID)
		return ctx.RespondEphemeral(fmt.Sprintf("✅ AutoMod violation alerts will now be sent to <#%s>.", channel.ID))
	}
	
}