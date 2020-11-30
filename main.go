package main

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"os"
	"os/signal"
	"syscall"
)

// Admin: https://discord.com/api/oauth2/authorize?client_id=782730468156112957&permissions=8&scope=bot
// Somewhat restrictive: https://discord.com/api/oauth2/authorize?client_id=782730468156112957&permissions=70590016&scope=bot

var LogChannelNames = []string{"meeting-logs", "voice-logs", "voice-channel-logs", "conference-logs", "meetinglogs", "conferencelogs", "voicelogs", "conferences", "meeting-summaries", "meetingsummaries", "voicesummaries", "voice-summaries", "meetings", "meeting", "conference"}

func main() {
	c, err := discordgo.New("Bot " + os.Getenv("DISCORD_BOT_TOKEN"))

	if err != nil {
		panic(err)
	}

	c.AddHandler(voiceStateUpdate)

	fmt.Println("Opening session...")

	if err := c.Open(); err != nil {
		panic(err)
	}

	sc := make(chan os.Signal, 1)

	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)

	<- sc

	fmt.Println("Session ended.")
}
