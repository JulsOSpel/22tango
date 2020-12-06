package main

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"regexp"
)

var channelspecMatchRegex = regexp.MustCompile(`\w+`)

const (
	genVoiceChannel byte = 0

	// genVoiceAndTextChannel = 1
)

func channelCreate(s *discordgo.Session, e *discordgo.ChannelCreate) {
	fmt.Println("Channel create:", e.Name)

	// Check if meets voice channel spec qualifications

	d := channelspecMatchRegex.FindAllString(e.Name, 20)

	if d != nil && len(d) >= 2 && d[0] == "mm" {
		fmt.Println("Matched channelspec regex.")

		switch d[1] {
		case "gen":
			if e.Channel.Type != discordgo.ChannelTypeGuildVoice {
				s.ChannelEdit(e.Channel.ID, "[MM] ERR NOT VOICE")
				return
			}

			s.ChannelEdit(e.Channel.ID, "[MM] Please wait...")

			fmt.Println("Init gen channel", e.Channel.ID)

			err := app.Set(e.GuildID, "chandat-v1-"+e.Channel.ID, []byte{genVoiceChannel})

			if err != nil {
				s.ChannelEdit(e.Channel.ID, "[MM] ERROR")
				return
			}

			s.ChannelEdit(e.Channel.ID, "Join to Create Channel")
		case "logmeet":
			if e.Channel.Type != discordgo.ChannelTypeGuildText {
				s.ChannelEdit(e.Channel.ID, "mm-error-not-text")
				return
			}

			s.ChannelEdit(e.Channel.ID, "[MM] Please wait...")

			fmt.Println("Init logmeet channel", e.Channel.ID)

			err := app.Set(e.GuildID, "meetingLogChannelID", []byte(e.Channel.ID))

			if err != nil {
				s.ChannelEdit(e.Channel.ID, "[MM] ERROR")
				return
			}

			s.ChannelEdit(e.Channel.ID, "meeting-logs")
		default:
			s.ChannelEdit(e.Channel.ID, "[MM] UNKNOWN DIRECTIVE")
			return
		}
	} else {
		fmt.Println("Did not match channelspec regex.")
	}
}

func channelDelete(s *discordgo.Session, e *discordgo.ChannelDelete) {
	// Clean up any chandat data for channel. (if exists)

	err := app.Del(e.GuildID, "chandat-v1-"+e.Channel.ID)

	if err != nil {
		fmt.Println(err)
		return
	}
}
