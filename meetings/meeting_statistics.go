package meetings

import "time"

func meetingMemberDurations(m *meeting) map[string]time.Duration {
	memberDurations := map[string]time.Duration{}

	participatingMembers := map[string]*meetingEvent{}

	for _, curEvent := range m.events {
		if curEvent.kind == meetingJoin {
			participatingMembers[curEvent.subject] = curEvent
		} else if curEvent.kind == meetingLeave {
			// Locate user in participating for removal

			memAccessJoinEvent, ok := participatingMembers[curEvent.subject]

			if !ok {
				// Ignore. No earlier join exists for leave event.
				continue
			}

			// Get access duration for this member (for current time in channel)

			curMemberAccessDuration := curEvent.at.Sub(memAccessJoinEvent.at)

			// Add access duration to member total

			curMemDuration, ok := memberDurations[curEvent.subject]

			if !ok {
				memberDurations[curEvent.subject] = curMemberAccessDuration
			} else {
				curMemDuration = curMemDuration + curMemberAccessDuration
			}

			// Delete previous join event

			delete(participatingMembers, curEvent.subject)
		}
	}

	return memberDurations
}
