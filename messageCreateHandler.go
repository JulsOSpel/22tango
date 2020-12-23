package main

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/ethanent/22tango/factcheck"
	"github.com/ethanent/22tango/meetings"
	"github.com/ethanent/discordkvs"
	"strings"
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
			Name:   "Troubleshooting",
			Value:  "Recreate channels that misbehave and make sure that the bot can see your voice and text channels.",
			Inline: false,
		},
	},
}
