package main

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	queue "github.com/ethanent/discordgo_voicestateupdatequeue"
	"strings"
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
			// Join event

			joinedChannel, err := s.Channel(e.ChannelID)

			if err != nil {
				fmt.Println(err)
				continue
			}

			cname := joinedChannel.Name

			curGuildMeetings, ok := guildMeetings[e.GuildID]

			if !ok {
				// Init guild

				guildMeetings[e.GuildID] = map[string]*meeting{}
				curGuildMeetings = guildMeetings[e.GuildID]
			}

			// If this is a generator channel, generate...

			if strings.Contains(cname, "join") || strings.Contains(cname, "click") && strings.Contains(cname, "channel") || strings.Contains(cname, "room") || strings.Contains(cname, "meeting") {
				// This is a generator channel.

				// Get user name

				generatingUser, err := s.User(e.UserID)

				if err != nil {
					fmt.Println(err)

					continue
				}

				// Generate channel

				generatedChannel, err := s.GuildChannelCreateComplex(e.GuildID, discordgo.GuildChannelCreateData{
					Name:     "[Gen] " + generatingUser.Username,
					Type:     discordgo.ChannelTypeGuildVoice,
					Position: joinedChannel.Position + 1,
					PermissionOverwrites: []*discordgo.PermissionOverwrite{
						&discordgo.PermissionOverwrite{
							ID:    e.UserID,
							Type:  "1",
							Deny:  0,
							Allow: discordgo.PermissionManageChannels,
						},
					},
					ParentID: joinedChannel.ParentID,
				})

				if err != nil {
					fmt.Println(err)

					continue
				}

				// Init meeting and move user.

				curGuildMeetings[e.ChannelID] = &meeting{
					channelID:     generatedChannel.ID,
					began:         eventTime,
					ended:         nil,
					curMembers:    []string{},
					events:        []*meetingEvent{},
					isTempChannel: true,
				}

				s.GuildMemberMove(e.GuildID, e.UserID, &generatedChannel.ID)

				continue
			}

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
				kind:    meetingLeave,
				subject: e.UserID,
			})
		} else if e.Type == queue.VoiceChannelLeave {
			// Leave event

			// Locate meeting

			guild, ok := guildMeetings[e.GuildID]

			if !ok {
				fmt.Println("Untracked leave in guild " + e.GuildID + " (this shouldn't happen generally)")
				continue
			}

			curMeeting, ok := guild[e.ChannelID]

			if !ok {
				fmt.Println("Untracked leave in channel " + e.GuildID + "-" + e.ChannelID + " (this shouldn't happen generally)")
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
				curMeeting.curMembers = append(curMeeting.curMembers[:userIdxInMeeting], curMeeting.curMembers[userIdxInMeeting:]...)
			}

			curMeeting.events = append(curMeeting.events, &meetingEvent{
				at:      eventTime,
				kind:    meetingLeave,
				subject: e.UserID,
			})

			// If meeting is empty, delete it from memory and trigger potential message.

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

				// => Consider sending summary message.

				err := sendSummaryMessage(s, e.GuildID, curMeeting)

				if err != nil {
					fmt.Println(err)
				}

				// => If meeting channel is temp channel, delete it.

				if curMeeting.isTempChannel {
					if _, err := s.ChannelDelete(e.ChannelID); err != nil {
						fmt.Println(err)
					}
				}
			}
		}
	}
}
