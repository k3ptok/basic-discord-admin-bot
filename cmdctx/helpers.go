package cmdctx

import(
	"github.com/bwmarrin/discordgo"
)

// --- Helper Response Methods ---

// Respond sends an immediate channel response
func (c *Context) Respond(content string) error {
	return c.Session.InteractionRespond(c.Interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: content,
		},
	})
}

// RespondEmbed sends a rich embed response
func (c *Context) RespondEmbed(embed *discordgo.MessageEmbed) error {
	return c.Session.InteractionRespond(c.Interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
		},
	})
}

// Defer informs Discord that the bot is processing (prevents 3s interaction timeout)
func (c *Context) Defer(ephemeral bool) error {
	var flags discordgo.MessageFlags
	if ephemeral {
		flags = discordgo.MessageFlagsEphemeral
	}
	return c.Session.InteractionRespond(c.Interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags: flags,
		},
	})
}

// EditFollowup updates a deferred response message
func (c *Context) EditFollowup(content string) error {
	_, err := c.Session.InteractionResponseEdit(c.Interaction.Interaction, &discordgo.WebhookEdit{
		Content: &content,
	})
	return err
}

// --- Type-Safe Option Extractors ---

//GetUser helper extracts user option safely
func (c *Context) GetUser(name string) (*discordgo.User, bool) {
	if opt, ok := c.optionsMap[name]; ok && opt.Type == discordgo.ApplicationCommandOptionUser {
		return opt.UserValue(c.Session), true
	}
	return nil, false
}

//GetString helper extracts string option safely
func (c *Context) GetString(name string) (string, bool) {
	if opt, ok := c.optionsMap[name]; ok && opt.Type == discordgo.ApplicationCommandOptionString {
		return opt.StringValue(), true
	}
	return "", false
}

//GetInt helper extracts int options safely
func (c *Context) GetInt(name string) (int64, bool) {
	if opt, ok := c.optionsMap[name]; ok && opt.Type == discordgo.ApplicationCommandOptionInteger {
		return opt.IntValue(), true
	}
	return 0, false
}

// GetBool safely extracts a boolean parameter
func (c *Context) GetBool(name string) (bool, bool) {
	if opt, ok := c.optionsMap[name]; ok && opt.Type == discordgo.ApplicationCommandOptionBoolean {
		return opt.BoolValue(), true
	}
	return false, false
}

// GetChannel safely extracts a channel object
func (c *Context) GetChannel(name string) (*discordgo.Channel, bool) {
	if opt, ok := c.optionsMap[name]; ok && opt.Type == discordgo.ApplicationCommandOptionChannel {
		return opt.ChannelValue(c.Session), true
	}
	return nil, false
}

// GetRole safely extracts a role object
func (c *Context) GetRole(name string) (*discordgo.Role, bool) {
	if opt, ok := c.optionsMap[name]; ok && opt.Type == discordgo.ApplicationCommandOptionRole {
		return opt.RoleValue(c.Session, c.Interaction.GuildID), true
	}
	return nil, false
}