package factcheck

import (
	"github.com/bwmarrin/discordgo"
	"github.com/ethanent/discordkvs"
)

const HelpMessage = "factcheck subcommands:\n\tstatus\n\tenable\n\tdisable"
const FactCheckDisabledMessage = "Fact check is **disabled**."
const FactCheckEnabledMessage = "Fact check is **enabled**."
const DataWriteErrorMessage = "Error writing data. Ensure bot permissions are configured correctly."
const NonOwnerErrorMessage = "Only the server owner may perform that action."

func ensureOwner(s *discordgo.Session, m *discordgo.MessageCreate) bool {
	g, err := s.Guild(m.GuildID)

	if err != nil {
		return false
	}

	return g.OwnerID == m.Author.ID
}

func HandleCommand(s *discordgo.Session, m *discordgo.MessageCreate, app *discordkvs.Application, a []string) {
	if len(a) < 1 {
		s.ChannelMessageSend(m.ChannelID, HelpMessage)
		return
	}

	switch a[0] {
	case "status":
		fcv, err := app.Get(m.GuildID, "factcheck")

		if err != nil {
			s.ChannelMessageSend(m.ChannelID, FactCheckDisabledMessage)
			return
		}

		if len(fcv) < 1 {
			// Invalid value
			s.ChannelMessageSend(m.ChannelID, "Invalid value")

			return
		}

		switch fcv[0] {
		case 0:
			// Disabled
			s.ChannelMessageSend(m.ChannelID, FactCheckDisabledMessage)
		case 1:
			// Enabled standard (warn)
			s.ChannelMessageSend(m.ChannelID, FactCheckEnabledMessage)
		default:
			// Invalid
			s.ChannelMessageSend(m.ChannelID, "Invalid value")
		}
	case "enable":
		if !ensureOwner(s, m) {
			s.ChannelMessageSend(m.ChannelID, NonOwnerErrorMessage)
			return
		}

		if err := app.Set(m.GuildID, "factcheck", []byte{1}); err != nil {
			s.ChannelMessageSend(m.ChannelID, DataWriteErrorMessage)
		} else {
			s.ChannelMessageSend(m.ChannelID, FactCheckEnabledMessage)
		}
	case "disable":
		if !ensureOwner(s, m) {
			s.ChannelMessageSend(m.ChannelID, NonOwnerErrorMessage)
			return
		}

		if err := app.Set(m.GuildID, "factcheck", []byte{0}); err != nil {
			s.ChannelMessageSend(m.ChannelID, DataWriteErrorMessage)
		} else {
			s.ChannelMessageSend(m.ChannelID, FactCheckDisabledMessage)
		}
	default:
		s.ChannelMessageSend(m.ChannelID, HelpMessage)
	}
}
