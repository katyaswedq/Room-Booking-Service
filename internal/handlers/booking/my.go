package booking

import (
	"context"
	"errors"
	"net/http"

	handlers "github.com/avito-internships/test-backend-1-katyaswedq/internal/handlers"
	middleware "github.com/avito-internships/test-backend-1-katyaswedq/internal/handlers/middleware"
	bookingusecase "github.com/avito-internships/test-backend-1-katyaswedq/internal/usecase/booking"
)

type MyUseCase interface {
	List(ctx context.Context, input bookingusecase.MyInput) (*bookingusecase.MyOutput, error)
}

type MyHandler struct {
	usecase MyUseCase
}

func NewMyHandler(usecase MyUseCase) *MyHandler {
	return &MyHandler{usecase: usecase}
}

func (h *MyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok || userID == "" {
		handlers.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized")
		return
	}

	out, err := h.usecase.List(r.Context(), bookingusecase.MyInput{
		UserID: userID,
	})
	if err != nil {
		switch {
		case errors.Is(err, bookingusecase.ErrInvalidUserID):
			handlers.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		default:
			handlers.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		}
		return
	}

	resp := make([]bookingResponse, 0, len(out.Bookings))
	for _, b := range out.Bookings {
		resp = append(resp, bookingResponse{
			ID:             b.ID,
			SlotID:         b.SlotID,
			UserID:         b.UserID,
			Status:         b.Status,
			ConferenceLink: b.ConferenceLink,
			CreatedAt:      b.CreatedAt,
		})
	}

	handlers.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"bookings": resp,
	})
}