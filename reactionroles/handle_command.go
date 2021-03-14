package reactionroles

import (
	"errors"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/ethanent/22tango/factcheck"
	"github.com/ethanent/discordkvs"
	"regexp"
)

const HelpMessage = "reactionroles subcommands:\n\tconfig <channel ID> <message ID> <emoji> <role ID>\n\trm <channel ID> <message ID> <emoji>"

var customEmojiRegex = regexp.MustCompile(`^<a?:(.+:([0-9]{18}))>$`)

// Bad "emoji" regex:
var standardEmojiRegex = regexp.MustCompile(`^[^A-Za-z\.\,\\\/]{1,8}$`)

func NormalizeDiscordEmoji(e string) (std string, api string, err error) {
	if standardEmojiRegex.MatchString(e) {
		return e, e, nil
	} else if customEmojiRegex.MatchString(e) {
		submatches := customEmojiRegex.FindStringSubmatch(e)

		return submatches[2], submatches[1], nil
	}

	return "", "", errors.New("invalid emoji")
}

func HandleCommand(s *discordgo.Session, m *discordgo.MessageCreate, app *discordkvs.Application, a []string) {
	if len(a) < 1 {
		s.ChannelMessageSend(m.ChannelID, HelpMessage)
		return
	}

	switch a[0] {
	case "rm":
		fallthrough
	case "config":
		if !factcheck.EnsureOwner(s, m) {
			s.ChannelMessageSend(m.ChannelID, "Only the server owner may configure reaction roles.")
			return
		}

		if a[0] == "config" {
			if len(a) < 5 {
				s.ChannelMessageSend(m.ChannelID, "Usage: 2!rr config <channel ID> <message ID> <emoji> <role ID>")
				return
			}
		} else if len(a) < 4 {
			s.ChannelMessageSend(m.ChannelID, "Usage: 2!rr rm <channel ID> <message ID> <emoji>")
			return
		}

		// Check message exists

		if _, err := s.ChannelMessage(a[1], a[2]); err != nil {
			s.ChannelMessageSend(m.ChannelID, "The provided channel / message do not exist.")
			return
		}

		if a[0] == "config" {
			// Check role exists

			if _, err := s.State.Role(m.GuildID, a[4]); err != nil {
				s.ChannelMessageSend(m.ChannelID, "No role exists with that ID.")
				return
			}
		}

		// Check emoji

		_, apiEmoji, err := NormalizeDiscordEmoji(a[3])

		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "The emoji provided is invalid.")
			return
		}

		// Done

		assoscKeyName := a[1] + "-" + a[2] + "-" + apiEmoji

		fmt.Println(assoscKeyName)

		if a[0] == "config" {
			if err := app.Set(m.GuildID, assoscKeyName, []byte(a[4])); err != nil {
				s.ChannelMessageSend(m.ChannelID, "Failed to save. Check bot permissions.")
				return
			}

			// Add bot reaction

			s.MessageReactionAdd(a[1], a[2], apiEmoji)

			// Inform user

			s.ChannelMessageSend(m.ChannelID, "The configuration has been updated.")
		} else {
			if err := app.Del(m.GuildID, assoscKeyName); err != nil {
				s.ChannelMessageSend(m.ChannelID, "Failed to remove association. Check bot permissions.")
				return
			}

			// Remove bot reaction if one exists

			s.MessageReactionRemove(a[1], a[2], apiEmoji, s.State.User.ID)

			// Inform user

			s.ChannelMessageSend(m.ChannelID, "The association has been removed.")
		}
	case "help":
		fallthrough
	default:
		s.ChannelMessageSend(m.ChannelID, HelpMessage)
	}
}
