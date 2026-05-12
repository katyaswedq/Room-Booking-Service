package slot

import "time"

type slotResponse struct {
	ID     string    `json:"id"`
	RoomID string    `json:"roomId"`
	Start  time.Time `json:"start"`
	End    time.Time `json:"end"`
}

type listSlotsResponse struct {
	Slots []slotResponse `json:"slots"`
}