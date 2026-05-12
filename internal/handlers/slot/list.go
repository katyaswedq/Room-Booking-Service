package slot

import (
	"context"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	handlers "github.com/avito-internships/test-backend-1-katyaswedq/internal/handlers"
	slotusecase "github.com/avito-internships/test-backend-1-katyaswedq/internal/usecase/slot"
)

type ListUseCase interface {
	List(ctx context.Context, input slotusecase.ListInput) (*slotusecase.ListOutput, error)
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
	roomID := chi.URLParam(r, "roomId")
	if roomID == "" {
		handlers.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "room id is required")
		return
	}

	date := r.URL.Query().Get("date")

	out, err := h.usecase.List(r.Context(), slotusecase.ListInput{
		RoomID: roomID,
		Date:   date,
	})
	if err != nil {
		switch {
		case errors.Is(err, slotusecase.ErrInvalidDate):
			handlers.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		case errors.Is(err, slotusecase.ErrRoomNotFound):
			handlers.WriteError(w, http.StatusNotFound, "ROOM_NOT_FOUND", err.Error())
		default:
			handlers.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		}
		return
	}

	slots := make([]slotResponse, 0, len(out.Slots))
	for _, slot := range out.Slots {
		slots = append(slots, slotResponse{
			ID:     slot.ID,
			RoomID: slot.RoomID,
			Start:  slot.Start,
			End:    slot.End,
		})
	}

	handlers.WriteJSON(w, http.StatusOK, listSlotsResponse{
		Slots: slots,
	})
}