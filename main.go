package main

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"os"
	"os/signal"
	"syscall"
)

var LogChannelNames = []string{"meeting-logs", "voice-logs", "voice-channel-logs", "conference-logs", "meetinglogs", "conferencelogs", "voicelogs", "conferences", "meeting-summaries", "meetingsummaries", "voicesummaries", "voice-summaries", "meetings", "meeting", "conference"}

func main() {
	c, err := discordgo.New("Bot " + os.Getenv("DISCORD_BOT_TOKEN"))

	if err != nil {
		panic(err)
	}

	c.AddHandler(messageCreate)
	c.AddHandler(voiceStateUpdate)

	fmt.Println("Opening session...")

	if err := c.Open(); err != nil {
		panic(err)
	}

	err = c.UpdateStatusComplex(discordgo.UpdateStatusData{
		AFK:    false,
		Status: "ml!help",
		Game: &discordgo.Game{
			Name: "ml!help",
		},
	})

	if err != nil {
		panic(err)
	}

	fmt.Println("OK.")

	sc := make(chan os.Signal, 1)

	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)

	<-sc

	fmt.Println("Session ended.")
}
