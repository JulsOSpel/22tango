package main

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"sync"
	"time"
)

type meetingEventKind int

const (
	meetingJoin meetingEventKind = iota
	meetingLeave
)

type meetingEvent struct {
	at time.Time
	kind meetingEventKind
	subject string
}

type meeting struct {
	channelID string

	began time.Time
	ended *time.Time

	curMembers []string

	events []*meetingEvent
}

var guildMeetings = map[string]map[string]*meeting{}
var guildsMux = &sync.RWMutex{}

func voiceStateUpdate(s *discordgo.Session, e *discordgo.VoiceStateUpdate) {
	fmt.Println("Receive VoiceState", e.GuildID, e.ChannelID, e.UserID)

	eventTime := time.Now()

	if e.ChannelID == "" {
		// Leave event

		guildsMux.Lock()
		defer guildsMux.Unlock()

		guild, ok := guildMeetings[e.GuildID]

		if !ok {
			// Untracked meeting (guild has not been initialized)

			fmt.Println("Untracked guild leave. Ignoring.")

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

			fmt.Println("Untracked meeting leave. Ignoring.")

			return
		}

		// Remove member from meeting curMembers

		exitedMeeting.curMembers = append(exitedMeeting.curMembers[:memberIdxInCurMembers], exitedMeeting.curMembers[memberIdxInCurMembers+1:]...)

		// Add leave event to meeting

		exitedMeeting.events = append(exitedMeeting.events, &meetingEvent{
			at:   eventTime,
			kind: meetingLeave,
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

			fmt.Println("memberDurations", memberDurations)

			// If it was just one member, let's ignore it. Barely a real meeting.

			if len(memberDurations) < 2 {
				return
			}

			// Now let's post a message with the meeting summary.

			// TODO :)
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
			events:     []*meetingEvent{
				&meetingEvent{
					at:      eventTime,
					kind:    meetingJoin,
					subject: e.UserID,
				},
			},
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
