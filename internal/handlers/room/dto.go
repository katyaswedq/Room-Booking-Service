package room

import "time"

type CreateRoomRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Capacity    int    `json:"capacity"`
}

type roomResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Capacity    int       `json:"capacity"`
	CreatedAt   time.Time `json:"created_at"`
}

type createRoomResponse struct {
	Room roomResponse `json:"room"`
}

type listRoomsResponse struct {
	Rooms []roomResponse `json:"rooms"`
}