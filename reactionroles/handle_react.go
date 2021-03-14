package reactionroles

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/ethanent/discordkvs"
)

func MessageReactionAdd(s *discordgo.Session, e *discordgo.MessageReactionAdd, app *discordkvs.Application) {
	if e.UserID == s.State.User.ID {
		// Ignore self react
		return
	}

	useEmoji, _, err := NormalizeDiscordEmoji(e.Emoji.MessageFormat())

	if err != nil {
		return
	}

	assoscKeyName := e.ChannelID + "-" + e.MessageID + "-" + useEmoji

	useRoleIDB, err := app.Get(e.GuildID, assoscKeyName)

	if err != nil {
		// Not a role-associated reaction
		fmt.Println("not role-assosc", assoscKeyName)
		return
	}

	if err := s.GuildMemberRoleAdd(e.GuildID, e.UserID, string(useRoleIDB)); err != nil {
		fmt.Println(err)
		if dmc, err := s.UserChannelCreate(e.UserID); err == nil {
			s.ChannelMessageSend(dmc.ID, "Failed to add a role. Check the bot permissions.")
		}
		return
	}
}

func MessageReactionRemove(s *discordgo.Session, e *discordgo.MessageReactionRemove, app *discordkvs.Application) {
	if e.UserID == s.State.User.ID {
		// Ignore self react removal
		return
	}

	useEmoji, _, err := NormalizeDiscordEmoji(e.Emoji.MessageFormat())

	if err != nil {
		return
	}

	assoscKeyName := e.ChannelID + "-" + e.MessageID + "-" + useEmoji

	useRoleIDB, err := app.Get(e.GuildID, assoscKeyName)

	if err != nil {
		// Not a role-associated reaction
		return
	}

	if err := s.GuildMemberRoleRemove(e.GuildID, e.UserID, string(useRoleIDB)); err != nil {
		fmt.Println(err)
		if dmc, err := s.UserChannelCreate(e.UserID); err == nil {
			s.ChannelMessageSend(dmc.ID, "Failed to remove a role. Check the bot permissions.")
		}
		return
	}
}
