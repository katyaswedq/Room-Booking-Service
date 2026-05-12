package room

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	handlers "github.com/avito-internships/test-backend-1-katyaswedq/internal/handlers"
	roomUsecase "github.com/avito-internships/test-backend-1-katyaswedq/internal/usecase/room"
)

type CreateUseCase interface {
	Create(ctx context.Context, input roomUsecase.CreateRoomInput) (*roomUsecase.CreateRoomOutput, error)
}

type CreateHandler struct {
	usecase CreateUseCase
}

func NewCreateHandler(usecase CreateUseCase) *CreateHandler {
	return &CreateHandler{
		usecase: usecase,
	}
}

func (h *CreateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var req CreateRoomRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		handlers.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
		return
	}

	out, err := h.usecase.Create(r.Context(), roomUsecase.CreateRoomInput{
		Name:        req.Name,
		Description: req.Description,
		Capacity:    req.Capacity,
	})
	if err != nil {
		switch {
		case errors.Is(err, roomUsecase.ErrInvalidName),
			errors.Is(err, roomUsecase.ErrInvalidCapacity):
			handlers.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		default:
			handlers.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		}
		return
	}

	handlers.WriteJSON(w, http.StatusCreated, createRoomResponse{
		Room: roomResponse{
			ID:          out.ID,
			Name:        out.Name,
			Description: out.Description,
			Capacity:    out.Capacity,
			CreatedAt:   out.CreatedAt,
		},
	})
}