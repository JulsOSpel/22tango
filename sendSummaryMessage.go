package main

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"sort"
	"time"
)

const EmbedColor = 823784
const TimeLayout = "2 Jan 2006 3:04 PM MST"
const MinMembersToSendSummary = 2
const MinDurationToSendSummary = 2 * time.Minute

type memberDuration struct {
	memberID string
	duration time.Duration
}

type memberDurationList []memberDuration

func (m memberDurationList) Len() int {
	return len(m)
}

func (m memberDurationList) Less(i, j int) bool {
	return m[i].duration < m[j].duration
}

func (m memberDurationList) Swap(i, j int) {
	hold := m[j]

	m[j] = m[i]
	m[i] = hold
}

func sendSummaryMessage(s *discordgo.Session, guildId string, meetingVoiceChannel *discordgo.Channel, m *meeting) error {
	memberDurations := meetingMemberDurations(m)

	// Don't send summary if qualifications not met.

	if len(memberDurations) < MinMembersToSendSummary {
		fmt.Println("Not sending summary. Min members.", len(memberDurations))
		return nil
	}

	if m.ended.Sub(m.began) < MinDurationToSendSummary {
		fmt.Println("Not sending summary. Min duration.")
		return nil
	}

	// Send.

	guildChannels, err := s.GuildChannels(guildId)

	if err != nil {
		return err
	}

	var sendSummaryChannel *discordgo.Channel

	for _, curChannel := range guildChannels {
		if curChannel.Type != discordgo.ChannelTypeGuildText {
			// Ignore non-text channel.

			continue
		}

		// Check if channel name in valid log channel names

		isLogChannelName := false

		for _, logChannelName := range LogChannelNames {
			if curChannel.Name == logChannelName {
				isLogChannelName = true
				break
			}
		}

		if isLogChannelName {
			sendSummaryChannel = curChannel
			break
		}
	}

	if sendSummaryChannel == nil {
		// Did not locate a meeting summary channel

		fmt.Println("Did not locate summary channel.")

		return nil
	}

	// Create summary message

	summaryFields := []*discordgo.MessageEmbedField{
		&discordgo.MessageEmbedField{
			Name:  "Duration",
			Value: m.ended.Sub(m.began).Round(time.Second).String(),
		},
		&discordgo.MessageEmbedField{
			Name:   "Began",
			Value:  m.began.Format(TimeLayout),
			Inline: true,
		},
		&discordgo.MessageEmbedField{
			Name:   "Ended",
			Value:  m.ended.Format(TimeLayout),
			Inline: true,
		},
	}

	// Sort member durations

	durations := memberDurationList{}

	for curMemID, curMemDur := range memberDurations {
		durations = append(durations, memberDuration{
			memberID: curMemID,
			duration: curMemDur,
		})
	}

	sort.Sort(durations)

	// Add sorted durations to new field (descending order)

	durationsText := ""

	for i := len(durations) - 1; i >= 0; i-- {
		d := durations[i]

		durationsText += "<@" + d.memberID + "> " + d.duration.Round(time.Second).String()

		if i != 0 {
			durationsText += "\n"
		}
	}

	summaryFields = append(summaryFields, &discordgo.MessageEmbedField{
		Name:   "Members",
		Value:  durationsText,
		Inline: false,
	})

	// Create final summary embed

	summary := &discordgo.MessageEmbed{
		Title: meetingVoiceChannel.Name + " Meeting Summary",
		Color: 7123569,
		Footer: &discordgo.MessageEmbedFooter{
			Text: "MeetingLogs Bot",
		},
		Fields: summaryFields,
	}

	// Send summary message

	if _, err := s.ChannelMessageSendEmbed(sendSummaryChannel.ID, summary); err != nil {
		return err
	}

	return nil
}
