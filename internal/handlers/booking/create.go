package booking

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	handlers "github.com/avito-internships/test-backend-1-katyaswedq/internal/handlers"
	middleware "github.com/avito-internships/test-backend-1-katyaswedq/internal/handlers/middleware"
	bookingusecase "github.com/avito-internships/test-backend-1-katyaswedq/internal/usecase/booking"
)

type CreateUseCase interface {
	Create(ctx context.Context, input bookingusecase.CreateInput) (*bookingusecase.CreateOutput, error)
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
	var req CreateBookingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		handlers.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
		return
	}

	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok || userID == "" {
		handlers.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized")
		return
	}

	out, err := h.usecase.Create(r.Context(), bookingusecase.CreateInput{
		SlotID:               req.SlotID,
		UserID:               userID,
		CreateConferenceLink: req.CreateConferenceLink,
	})
	if err != nil {
		switch {
		case errors.Is(err, bookingusecase.ErrInvalidSlotID),
			errors.Is(err, bookingusecase.ErrSlotAlreadyStarted):
			handlers.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		case errors.Is(err, bookingusecase.ErrSlotNotFound):
			handlers.WriteError(w, http.StatusNotFound, "SLOT_NOT_FOUND", err.Error())
		case errors.Is(err, bookingusecase.ErrSlotAlreadyBooked):
			handlers.WriteError(w, http.StatusConflict, "SLOT_ALREADY_BOOKED", err.Error())
		case errors.Is(err, bookingusecase.ErrConferenceLinkError):
			handlers.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		default:
			handlers.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		}
		return
	}

	handlers.WriteJSON(w, http.StatusCreated, createBookingResponse{
		Booking: bookingResponse{
			ID:             out.ID,
			SlotID:         out.SlotID,
			UserID:         out.UserID,
			Status:         out.Status,
			ConferenceLink: out.ConferenceLink,
			CreatedAt:      out.CreatedAt,
		},
	})
}