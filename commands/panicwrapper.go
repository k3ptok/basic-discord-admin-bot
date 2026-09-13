package commands

import(
	"fmt"
	"log/slog"
	"github.com/k3ptok/BasicDiscordBot/cmdctx"
)

// PanicWrapper runs a command wrapped in panic recovery and error logging
func PanicWrapper(cmds Command, ctx *cmdctx.Context) {
	defer func() {
		if r := recover(); r != nil {
			ctx.Logger.Error("panic recovered during execution", slog.Any("panic", r))
			//Attempt to notify user
			_ = ctx.Respond("❌ An internal error occurred while processing this command.")
		}
	}()

	if err := cmds.Execute(ctx); err != nil {
		ctx.Logger.Error("Command returned error", slog.Any("error", err))
		_ = ctx.Respond(fmt.Sprintf("❌ Execution error: %s", err.Error()))
	}
}