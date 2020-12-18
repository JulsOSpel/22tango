package main

import (
	"github.com/bwmarrin/discordgo"
	"github.com/ethanent/22tango/meetings"
)

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Content == "ml!help" {
		s.ChannelMessageSendEmbed(m.ChannelID, &discordgo.MessageEmbed{
			Title:       "MeetingManager Help",
			Description: "MeetingManager bot created by ethanent.\nGitHub repo: [ethanent/discord-meetinglogs](https://github.com/ethanent/discord-meetinglogs)",
			Timestamp:   "",
			Color:       meetings.EmbedColor,
			Footer: &discordgo.MessageEmbedFooter{
				Text: "MeetingLogs Bot",
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
					Name:   "Troubleshooting",
					Value:  "Recreate channels that misbehave and make sure that the bot can see your voice and text channels.",
					Inline: false,
				},
			},
		})
	}
}
