package booking

import "time"

const (
	StatusActive    = "active"
	StatusCancelled = "cancelled"
)

type Booking struct {
	ID             string
	SlotID         string
	UserID         string
	Status         string
	ConferenceLink string
	CreatedAt      time.Time
}