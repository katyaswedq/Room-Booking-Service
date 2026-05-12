package slot

import (
	"errors"
	"time"

	"github.com/google/uuid"

	domainslot "github.com/avito-internships/test-backend-1-katyaswedq/internal/domain/slot"
)

const slotDuration = 30 * time.Minute

var ErrInvalidTimeRange = errors.New("invalid time range")

func GenerateForNextDays(roomID string, daysOfWeek []int, startTime time.Time, endTime time.Time, now time.Time, days int,) ([]domainslot.Slot, error) {
	if !startTime.Before(endTime) {
		return nil, ErrInvalidTimeRange
	}

	allowedDays := make(map[int]struct{}, len(daysOfWeek))
	for _, day := range daysOfWeek {
		allowedDays[day] = struct{}{}
	}

	startOfToday := time.Date(now.UTC().Year(),now.UTC().Month(), now.UTC().Day(), 0, 0, 0, 0, time.UTC)

	slots := make([]domainslot.Slot, 0)

	for i := 0; i < days; i++ {
		currentDate := startOfToday.AddDate(0, 0, i)

		apiDay := int(currentDate.Weekday())
		if apiDay == 0 {
			apiDay = 7
		}

		if _, ok := allowedDays[apiDay]; !ok {
			continue
		}

		windowStart := time.Date(currentDate.Year(), currentDate.Month(), currentDate.Day(), startTime.Hour(), startTime.Minute(), 0, 0, time.UTC)

		windowEnd := time.Date(currentDate.Year(), currentDate.Month(), currentDate.Day(), endTime.Hour(), endTime.Minute(), 0, 0, time.UTC)

		for slotStart := windowStart; !slotStart.Add(slotDuration).After(windowEnd); slotStart = slotStart.Add(slotDuration) {
			slotEnd := slotStart.Add(slotDuration)

			slots = append(slots, domainslot.Slot{
				ID:     uuid.NewString(),
				RoomID: roomID,
				Start:  slotStart,
				End:    slotEnd,
			})
		}
	}

	return slots, nil
}