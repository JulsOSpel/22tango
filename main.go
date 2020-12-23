package main

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/ethanent/22tango/factcheck"
	"github.com/ethanent/22tango/meetings"
	queue "github.com/ethanent/discordgo_voicestateupdatequeue"
	"github.com/ethanent/discordkvs"
	"os"
	"os/signal"
	"syscall"
)

var LogChannelNames = []string{"meeting-logs", "voice-logs", "voice-channel-logs", "conference-logs", "meetinglogs", "conferencelogs", "voicelogs", "conferences", "meeting-summaries", "meetingsummaries", "voicesummaries", "voice-summaries", "meetings", "meeting", "conference"}

var app *discordkvs.Application

func main() {
	s, err := discordgo.New("Bot " + os.Getenv("DISCORD_BOT_TOKEN"))

	if err != nil {
		panic(err)
	}

	app, err = discordkvs.NewApplication(s, "MeetingManager-4341238733892")

	if err != nil {
		panic(err)
	}

	// => Set up meetings

	voiceEventChan := make(chan *queue.VoiceStateEvent)

	q := queue.NewVoiceStateEventQueue(voiceEventChan)

	s.AddHandler(q.Handler)

	// Add handlers wrapping to provide app to meetings package.

	s.AddHandler(func (s *discordgo.Session, e *discordgo.ChannelCreate) {
		meetings.ChannelCreate(s, e, app)
	})

	s.AddHandler(func (s *discordgo.Session, e *discordgo.ChannelDelete) {
		meetings.ChannelDelete(s, e, app)
	})

	go meetings.BeginAcceptingVoiceEvents(s, app, voiceEventChan)

	// => Set up factcheck

	if err := factcheck.LoadData(); err != nil {
		panic(err)
	}

	s.AddHandler(func (s *discordgo.Session, e *discordgo.MessageCreate) {
		factcheck.MessageSend(s, e, app)
	})

	// => Set up command handler

	s.AddHandler(messageCreate)

	// => Open session

	fmt.Println("Opening session...")

	if err := s.Open(); err != nil {
		panic(err)
	}

	err = s.UpdateStatusComplex(discordgo.UpdateStatusData{
		AFK:    false,
		Status: "2!help",
		Game: &discordgo.Game{
			Name: "2!help",
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
