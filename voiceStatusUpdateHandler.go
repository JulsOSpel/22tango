package main

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"strings"
	"sync"
	"time"
)

const MinimumMeetingDurationForPosting = time.Minute * 2

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
var guildsMux = &sync.RWMutex{}

func voiceStateUpdate(s *discordgo.Session, e *discordgo.VoiceStateUpdate) {
	fmt.Println("Receive VoiceState", e.GuildID, e.ChannelID, e.UserID)

	eventTime := time.Now()

	// Check if user joined a channel generator channel.

	if e.ChannelID != "" {
		// This potentially is a join. Let's check if the channel is indeed a channel generator channel.
		channel, err := s.Channel(e.ChannelID)

		if err != nil {
			fmt.Println(err)

			return
		}

		// => Check if name meets qualifications.

		cname := strings.ToLower(channel.Name)

		if strings.Contains(cname, "join") || strings.Contains(cname, "click") && strings.Contains(cname, "channel") || strings.Contains(cname, "room") || strings.Contains(cname, "meeting") {
			// This is a join to create channel.
			// Let's set up the new channel and create the meeting.

			fmt.Println("Channel generator detected. Preparing to bounce user to new temp channel.")

			guildsMux.Lock()
			defer guildsMux.Unlock()

			joinChannel, err := s.Channel(e.ChannelID)

			if err != nil {
				fmt.Println(err)

				return
			}

			member, err := s.GuildMember(e.GuildID, e.UserID)

			if err != nil {
				fmt.Println(err)

				return
			}

			var displayName string

			if member.Nick != "" {
				displayName = member.Nick
			} else {
				displayName = member.User.Username
			}

			createdChannel, err := s.GuildChannelCreateComplex(e.GuildID, discordgo.GuildChannelCreateData{
				Name:     "[MM] " + displayName,
				Type:     discordgo.ChannelTypeGuildVoice,
				Position: joinChannel.Position + 1,
				PermissionOverwrites: []*discordgo.PermissionOverwrite{
					&discordgo.PermissionOverwrite{
						ID:    e.UserID,
						Type:  "1",
						Deny:  0,
						Allow: discordgo.PermissionManageChannels,
					},
				},
				ParentID: joinChannel.ParentID,
			})

			if err != nil {
				fmt.Println(err)

				return
			}

			if err := s.GuildMemberMove(e.GuildID, e.UserID, &createdChannel.ID); err != nil {
				fmt.Println(err)

				return
			}

			// Initialize the meeting (without the user in it, because the user will be added due to the join event being provided for the generated channel as well)

			gm, ok := guildMeetings[e.GuildID]

			if !ok {
				guildMeetings[e.GuildID] = map[string]*meeting{}
				gm = guildMeetings[e.GuildID]
			}

			gm[createdChannel.ID] = &meeting{
				channelID:     createdChannel.ID,
				began:         eventTime,
				ended:         nil,
				curMembers:    []string{},
				events:        []*meetingEvent{},
				isTempChannel: true,
			}

			return
		}
	}

	// Otherwise continue. Check if user is joining / leaving.

	if e.ChannelID == "" {
		// Leave event

		guildsMux.Lock()
		defer guildsMux.Unlock()

		guild, ok := guildMeetings[e.GuildID]

		if !ok {
			// Untracked meeting (guild has not been initialized)

			fmt.Println("Leave for untracked guild. Ignoring.")

			return
		}

		// Find exited channel

		var exitedMeeting *meeting
		var memberIdxInCurMembers int

		for _, curMeeting := range guild {
			meetingHasMember := false

			for memberIdx, memberID := range curMeeting.curMembers {
				if memberID == e.UserID {
					meetingHasMember = true
					memberIdxInCurMembers = memberIdx
					break
				}
			}

			if meetingHasMember {
				exitedMeeting = curMeeting
				break
			}
		}

		if exitedMeeting == nil {
			// Untracked meeting (did not locate the meeting user left)

			fmt.Println("Leave for untracked meeting. Ignoring.")

			return
		}

		// Remove member from meeting curMembers

		exitedMeeting.curMembers = append(exitedMeeting.curMembers[:memberIdxInCurMembers], exitedMeeting.curMembers[memberIdxInCurMembers+1:]...)

		// Add leave event to meeting

		exitedMeeting.events = append(exitedMeeting.events, &meetingEvent{
			at:      eventTime,
			kind:    meetingLeave,
			subject: e.UserID,
		})

		fmt.Println("Logged leave", e.UserID, "(", len(exitedMeeting.curMembers), "members left in meeting", ")")

		// Check if this is final user to leave. If so, consider sending final message and clean up memory.

		if len(exitedMeeting.curMembers) == 0 {
			// Meeting concluded:
			// This is final user to leave meeting. Collect stats, clean up memory, and log to chat if it had multiple members.

			// Set meeting end time.

			exitedMeeting.ended = &eventTime

			// Discard meeting from memory.

			delete(guildMeetings[e.GuildID], exitedMeeting.channelID)

			if len(guildMeetings[e.GuildID]) == 0 {
				delete(guildMeetings, e.GuildID)
			}

			fmt.Println("Meeting concluded. Removed from memory. Guilds:", len(guildMeetings))

			// Collect statistics

			memberDurations := meetingMemberDurations(exitedMeeting)

			fmt.Println("Finalized meeting", e.GuildID, exitedMeeting.channelID, exitedMeeting.began, *exitedMeeting.ended, memberDurations)

			// If it is a temp channel (eg. made using generator), delete it.

			s.ChannelDelete(exitedMeeting.channelID)

			// If it was just one member, let's ignore it. Barely a real meeting.

			if len(memberDurations) < 2 {
				fmt.Println("Meeting did not have enough members to post")
				return
			}

			// If the meeting didn't last long, ignore it.

			if exitedMeeting.ended.Sub(exitedMeeting.began) < MinimumMeetingDurationForPosting {
				fmt.Println("Meeting not long enough to post")
				return
			}

			// Now let's post a message with the meeting summary.

			// Locate meeting logs channel in guild

			if err := sendSummaryMessage(s, e.GuildID, exitedMeeting, memberDurations); err != nil {
				fmt.Println("Error. Failed to send summary message:", err)
			}
		}
	} else {
		// Check if this is a user join.

		guildsMux.Lock()
		defer guildsMux.Unlock()

		meetingIfUninitialized := &meeting{
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

		guild, ok := guildMeetings[e.GuildID]

		if !ok {
			// This is a new meeting. Guild is not stored. Initialize guild and meeting.

			guildMeetings[e.GuildID] = map[string]*meeting{
				e.ChannelID: meetingIfUninitialized,
			}

			fmt.Println("Init meeting & guild")

			return
		}

		curMeeting, ok := guild[e.ChannelID]

		if !ok {
			// Meeting has not been initialized for guild. Initialize meeting.

			guildMeetings[e.GuildID][e.ChannelID] = meetingIfUninitialized

			fmt.Println("Init meeting")

			return
		}

		// Meeting is already initialized. Check if the member is already in the curMembers list to see if this is a join event.

		eventUserIsInCurMembers := false

		for _, curMember := range curMeeting.curMembers {
			if curMember == e.UserID {
				eventUserIsInCurMembers = true
				break
			}
		}

		if eventUserIsInCurMembers == true {
			// This is an update event. Discard it.

			fmt.Println("Ignoring update event.")

			return
		}

		curMeeting.events = append(curMeeting.events, &meetingEvent{
			at:      eventTime,
			kind:    meetingJoin,
			subject: e.UserID,
		})

		curMeeting.curMembers = append(curMeeting.curMembers, e.UserID)

		fmt.Println("Logged join.", e.UserID, "(", len(curMeeting.curMembers), "in meeting", ")")
	}
}
