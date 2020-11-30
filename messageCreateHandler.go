package main

import (
	"github.com/bwmarrin/discordgo"
)

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Content == "ml!help" {
		s.ChannelMessageSendEmbed(m.ChannelID, &discordgo.MessageEmbed{
			Title:       "MeetingLogs Help",
			Description: "GitHub repo: [ethanent/discord-meetinglogs](https://github.com/ethanent/discord-meetinglogs)",
			Timestamp:   "",
			Color:       EmbedColor,
			Footer: &discordgo.MessageEmbedFooter{
				Text: "MeetingLogs Bot",
			},
			Fields: []*discordgo.MessageEmbedField{
				&discordgo.MessageEmbedField{
					Name:   "Setup",
					Value:  "To use the bot, create a channel called \"meeting-logs\" in your server. The bot will report meetings with at least two members that last at least a little while there.",
					Inline: false,
				},
				&discordgo.MessageEmbedField{
					Name:   "Troubleshooting",
					Value:  "Be sure that the bot can send messages in the output channel. Also, make sure that the bot can see your voice channels.",
					Inline: false,
				},
			},
		})
	}
}
