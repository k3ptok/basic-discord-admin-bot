package commands

import (
	"log/slog"
	"github.com/bwmarrin/discordgo"
	"github.com/k3ptok/BasicDiscordBot/cmdctx"
)

func RequirePermissions(permission int64, handler SubCommandHandler) SubCommandHandler {
	return func(ctx *cmdctx.Context) error {
		if ctx.Interaction.Member == nil {
			return ctx.Respond("❌ This command can only be used in a server.")
		}

		userPerms := ctx.Interaction.Member.Permissions

		//check if user is an admin or has the permission
		isAdmin := (userPerms & discordgo.PermissionAdministrator) == discordgo.PermissionAdministrator
		hasPerm := (userPerms & permission) == permission

		if !isAdmin && !hasPerm {
			ctx.Logger.Warn("Unauthorized Command Attempt",
			slog.String("user_id", ctx.Interaction.Member.User.ID),
			slog.String("subcommand", ctx.Subcommand),
		)
		return ctx.Respond("⛔ **Access Denied:** You do not have permission to execute this command.")
		}
		// Permission granted
		return handler(ctx)
	}
}