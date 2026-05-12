package room

import (
	"context"
	"net/http"

	handlers "github.com/avito-internships/test-backend-1-katyaswedq/internal/handlers"
	roomUsecase "github.com/avito-internships/test-backend-1-katyaswedq/internal/usecase/room"
)

type ListUseCase interface {
	List(ctx context.Context) ([]roomUsecase.ListRoomOutput, error)
}

type ListHandler struct {
	usecase ListUseCase
}

func NewListHandler(usecase ListUseCase) *ListHandler {
	return &ListHandler{
		usecase: usecase,
	}
}

func (h *ListHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	out, err := h.usecase.List(r.Context())
	if err != nil {
		handlers.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		return
	}

	rooms := make([]roomResponse, 0, len(out))
	for _, rm := range out {
		rooms = append(rooms, roomResponse{
			ID:          rm.ID,
			Name:        rm.Name,
			Description: rm.Description,
			Capacity:    rm.Capacity,
			CreatedAt:   rm.CreatedAt,
		})
	}

	handlers.WriteJSON(w, http.StatusOK, listRoomsResponse{
		Rooms: rooms,
	})
}