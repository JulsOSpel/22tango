package main

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	queue "github.com/ethanent/discordgo_voicestateupdatequeue"
	"time"
)

type meetingEventKind int

const (
	meetingJoin meetingEventKind = iota
	meetingLeave
)

type meetingEvent struct {
	at      time.Time
	kind    meetingEventKind
	subject string
}

type meeting struct {
	channelID string

	began time.Time
	ended *time.Time

	curMembers []string

	events []*meetingEvent

	// Was this channel created using a channel generator channel?
	// If so, be sure to clean it up once meeting ends.
	isTempChannel bool
}

var guildMeetings = map[string]map[string]*meeting{}

// No mux needed because this program is single-threaded.

func beginAcceptingVoiceEvents(s *discordgo.Session, c chan *queue.VoiceStateEvent) {
	for e := range c {
		eventTime := time.Now()

		if e.Type == queue.VoiceChannelJoin {
			fmt.Println("Join event received", e.ChannelID)

			// Join event

			joinedChannel, err := s.Channel(e.ChannelID)

			if err != nil {
				fmt.Println(err)
				continue
			}

			curGuildMeetings, ok := guildMeetings[e.GuildID]

			if !ok {
				// Init guild

				guildMeetings[e.GuildID] = map[string]*meeting{}
				curGuildMeetings = guildMeetings[e.GuildID]
			}

			// Check if it is a generator channel...

			chanDat, err := app.Get(e.GuildID, "chandat-v1-"+e.ChannelID)
			var isGenChannel bool

			if err != nil {
				fmt.Println(err)

				isGenChannel = false
			} else {
				isGenChannel = chanDat[0] == genVoiceChannel
			}

			if isGenChannel {
				// This is a generator channel.

				fmt.Println("Gen channel")

				// Get user name

				generatingUser, err := s.User(e.UserID)

				if err != nil {
					fmt.Println(err)

					continue
				}

				// Generate channel

				generatedChannel, err := s.GuildChannelCreateComplex(e.GuildID, discordgo.GuildChannelCreateData{
					Name:     "[Voice] " + generatingUser.Username,
					Type:     discordgo.ChannelTypeGuildVoice,
					Position: joinedChannel.Position + 1,
					PermissionOverwrites: []*discordgo.PermissionOverwrite{
						&discordgo.PermissionOverwrite{
							ID:    e.UserID,
							Type:  "1",
							Deny:  0,
							Allow: discordgo.PermissionManageChannels | discordgo.PermissionAllVoice,
						},
					},
					ParentID: joinedChannel.ParentID,
				})

				if err != nil {
					fmt.Println(err)

					continue
				}

				// Init meeting and move user.

				curGuildMeetings[generatedChannel.ID] = &meeting{
					channelID:     generatedChannel.ID,
					began:         eventTime,
					ended:         nil,
					curMembers:    []string{},
					events:        []*meetingEvent{},
					isTempChannel: true,
				}

				if err := s.GuildMemberMove(e.GuildID, e.UserID, &generatedChannel.ID); err != nil {
					fmt.Println(err)
				}

				continue
			}

			fmt.Println("Not a gen channel")

			// Otherwise, this is a normal join event.

			joinedMeet, ok := curGuildMeetings[e.ChannelID]

			if !ok {
				// New meeting. Create it.

				curGuildMeetings[e.ChannelID] = &meeting{
					channelID:  e.ChannelID,
					began:      eventTime,
					ended:      nil,
					curMembers: []string{e.UserID},
					events: []*meetingEvent{
						&meetingEvent{
							at:      eventTime,
							kind:    meetingJoin,
							subject: e.UserID,
						},
					},
					isTempChannel: false,
				}

				continue
			}

			// Add user to meeting.

			joinedMeet.curMembers = append(joinedMeet.curMembers, e.UserID)
			joinedMeet.events = append(joinedMeet.events, &meetingEvent{
				at:      eventTime,
				kind:    meetingJoin,
				subject: e.UserID,
			})
		} else if e.Type == queue.VoiceChannelLeave {
			fmt.Println("Leave event received", e.ChannelID)
			// Leave event

			// Locate meeting

			guild, ok := guildMeetings[e.GuildID]

			if !ok {
				fmt.Println("Untracked leave in guild " + e.GuildID)
				continue
			}

			curMeeting, ok := guild[e.ChannelID]

			if !ok {
				fmt.Println("Untracked leave in channel " + e.GuildID + "-" + e.ChannelID)
				continue
			}

			// Remove user from meeting.

			userIdxInMeeting := -1

			for i, uid := range curMeeting.curMembers {
				if uid == e.UserID {
					userIdxInMeeting = i
					break
				}
			}

			if userIdxInMeeting != -1 {
				curMeeting.curMembers = append(curMeeting.curMembers[:userIdxInMeeting], curMeeting.curMembers[userIdxInMeeting+1:]...)
			}

			curMeeting.events = append(curMeeting.events, &meetingEvent{
				at:      eventTime,
				kind:    meetingLeave,
				subject: e.UserID,
			})

			// If meeting is empty, delete it from memory and trigger potential message.

			fmt.Println("Still left:", len(curMeeting.curMembers))

			if len(curMeeting.curMembers) == 0 {
				// Meeting is empty.

				// => End meeting

				curMeeting.ended = &eventTime

				// => Delete from memory.

				delete(guild, e.ChannelID)

				if len(guild) == 0 {
					delete(guildMeetings, e.GuildID)
				}

				fmt.Println("Deleted from memory. Guilds:", len(guildMeetings))

				// => If meeting channel is temp channel, delete it.

				var meetingChan *discordgo.Channel
				var err error

				if curMeeting.isTempChannel {
					meetingChan, err = s.ChannelDelete(e.ChannelID)

					if err != nil {
						fmt.Println(err)
					}
				} else {
					meetingChan, err = s.Channel(e.ChannelID)

					if err != nil {
						fmt.Println(err)
						continue
					}
				}

				// => Consider sending summary message.

				if err := sendSummaryMessage(s, e.GuildID, meetingChan, curMeeting); err != nil {
					fmt.Println(err)
				}
			}
		}
	}
}
