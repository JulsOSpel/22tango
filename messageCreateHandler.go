package main

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/ethanent/22tango/factcheck"
	"github.com/ethanent/22tango/meetings"
	"github.com/ethanent/discordkvs"
	"strings"
	"time"
)

var subcommandHandlersMap = map[string]func(*discordgo.Session, *discordgo.MessageCreate, *discordkvs.Application, []string){
	"factcheck": factcheck.HandleCommand,
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if strings.Index(m.Content, "2!") == 0 && len(m.Content) >= 3 {
		// Command
		arg := strings.Split(string([]rune(m.Content)[2:]), " ")

		if len(arg) < 1 {
			return
		}

		fmt.Println("Command from", m.Author.ID, arg)

		switch arg[0] {
		case "help":
			s.ChannelMessageSendEmbed(m.ChannelID, helpEmbed)
		case "tz":
			if len(arg) < 2 {
				tzName := "Default (" + meetings.DefaultTZName + ")"

				if tzB, err := app.Get(m.GuildID, "timezone"); err == nil {
					tzName = string(tzB)
				}

				s.ChannelMessageSend(m.ChannelID, "Server preferred timezone is currently set to `" + tzName + "`.\n\nTo set timezone: `2!tz [timezone]`\nExample: `2!tz America/Los_Angeles`")
				return
			}

			// Prepare to update timezone

			// Check that user is owner

			if g, err := s.Guild(m.GuildID); err != nil {
				return
			} else {
				if g.OwnerID != m.Author.ID {
					s.ChannelMessageSend(m.ChannelID, "Only the server owner may update the preferred timezone.")
					return
				}
			}

			// Check that timezone exists

			_, err := time.LoadLocation(arg[1])

			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "Unknown timezone. Example: `America/Los_Angeles`\nList of timezones: <https://en.wikipedia.org/wiki/List_of_tz_database_time_zones>")
				return
			}

			if err := app.Set(m.GuildID, "timezone", []byte(arg[1])); err != nil {
				fmt.Println(err)
				s.ChannelMessageSend(m.ChannelID, "Failed to save timezone. Check bot permissions.")

				return
			}

			s.ChannelMessageSend(m.ChannelID, "Server preferred timezone has has been set to `" + arg[1] + "`")
		default:
			// Find correct handler for subcommand if exists

			useHandler, ok := subcommandHandlersMap[arg[0]]

			if ok {
				useHandler(s, m, app, arg[1:])
			}
		}
	}
}

var helpEmbed = &discordgo.MessageEmbed{
	Title:       "22tango Help",
	Description: "22tango bot created by ethanent.\nGitHub repo: [ethanent/22tango](https://github.com/ethanent/22tango)",
	Color:       meetings.EmbedColor,
	Footer: &discordgo.MessageEmbedFooter{
		Text: "22tango Bot",
	},
	Fields: []*discordgo.MessageEmbedField{
		&discordgo.MessageEmbedField{
			Name:   "Log Meetings",
			Value:  "Create a text channel called \"mm-logmeet\". The bot will rename it once it is up and running (if it can manage the channel.) The bot will report meetings with at least two members that last at least a little while there.",
			Inline: false,
		},
		&discordgo.MessageEmbedField{
			Name:   "Generator Channels",
			Value:  "Create a voice channel called \"mm-gen\". The bot will rename it once it is up and running (if it can manage the channel.)\nWhen someone joins this channel, a personal meeting channel is generated that they can manage.",
			Inline: false,
		},
		&discordgo.MessageEmbedField{
			Name:   "Fact Checking",
			Value:  "See fact checking features using the `2!factcheck` subcommand.",
			Inline: false,
		},
		&discordgo.MessageEmbedField{
			Name:   "Set Timezone",
			Value:  "Set timezone with 2!tz command. eg. `2!tz America/Los_Angeles`",
			Inline: false,
		},
		&discordgo.MessageEmbedField{
			Name:   "Troubleshooting",
			Value:  "Recreate channels that misbehave and make sure that the bot can see your voice and text channels.",
			Inline: false,
		},
	},
}
