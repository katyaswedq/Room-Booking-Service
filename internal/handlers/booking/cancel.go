package booking

import (
	"context"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	handlers "github.com/avito-internships/test-backend-1-katyaswedq/internal/handlers"
	middleware "github.com/avito-internships/test-backend-1-katyaswedq/internal/handlers/middleware"
	bookingusecase "github.com/avito-internships/test-backend-1-katyaswedq/internal/usecase/booking"
)

type CancelUseCase interface {
	Cancel(ctx context.Context, input bookingusecase.CancelInput) (*bookingusecase.CancelOutput, error)
}

type CancelHandler struct {
	usecase CancelUseCase
}

func NewCancelHandler(usecase CancelUseCase) *CancelHandler {
	return &CancelHandler{usecase: usecase}
}

func (h *CancelHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	bookingID := chi.URLParam(r, "bookingId")
	if bookingID == "" {
		handlers.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "booking id is required")
		return
	}

	userID, _ := r.Context().Value(middleware.UserIDKey).(string)

	out, err := h.usecase.Cancel(r.Context(), bookingusecase.CancelInput{
		BookingID: bookingID,
		UserID:    userID,
	})
	if err != nil {
		switch {
		case errors.Is(err, bookingusecase.ErrInvalidBookingID):
			handlers.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		case errors.Is(err, bookingusecase.ErrBookingNotFound):
			handlers.WriteError(w, http.StatusNotFound, "BOOKING_NOT_FOUND", err.Error())
		case errors.Is(err, bookingusecase.ErrForbidden):
			handlers.WriteError(w, http.StatusForbidden, "FORBIDDEN", "forbidden")
		default:
			handlers.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		}
		return
	}

	handlers.WriteJSON(w, http.StatusOK, createBookingResponse{
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