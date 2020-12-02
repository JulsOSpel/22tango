package main

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	queue "github.com/ethanent/discordgo_voicestateupdatequeue"
	"os"
	"os/signal"
	"syscall"
)

var LogChannelNames = []string{"meeting-logs", "voice-logs", "voice-channel-logs", "conference-logs", "meetinglogs", "conferencelogs", "voicelogs", "conferences", "meeting-summaries", "meetingsummaries", "voicesummaries", "voice-summaries", "meetings", "meeting", "conference"}

func main() {
	s, err := discordgo.New("Bot " + os.Getenv("DISCORD_BOT_TOKEN"))

	if err != nil {
		panic(err)
	}

	eventChan := make(chan *queue.VoiceStateEvent)

	q := queue.NewVoiceStateEventQueue(eventChan)

	s.AddHandler(q.Handler)

	s.AddHandler(messageCreate)

	go beginAcceptingVoiceEvents(s, eventChan)

	fmt.Println("Opening session...")

	if err := s.Open(); err != nil {
		panic(err)
	}

	err = s.UpdateStatusComplex(discordgo.UpdateStatusData{
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
