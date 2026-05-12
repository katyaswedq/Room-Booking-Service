package booking

import "time"

type CreateBookingRequest struct {
	SlotID               string `json:"slotId"`
	CreateConferenceLink bool   `json:"createConferenceLink"`
}

type bookingResponse struct {
	ID             string    `json:"id"`
	SlotID         string    `json:"slotId"`
	UserID         string    `json:"userId"`
	Status         string    `json:"status"`
	ConferenceLink string    `json:"conferenceLink"`
	CreatedAt      time.Time `json:"createdAt"`
}

type createBookingResponse struct {
	Booking bookingResponse `json:"booking"`
}

type paginationResponse struct {
	Page     int `json:"page"`
	PageSize int `json:"pageSize"`
	Total    int `json:"total"`
}

type listBookingsResponse struct {
	Bookings   []bookingResponse `json:"bookings"`
	Pagination paginationResponse `json:"pagination"`
}