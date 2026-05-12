package booking

import (
	"context"
	"errors"
	"net/http"

	handlers "github.com/avito-internships/test-backend-1-katyaswedq/internal/handlers"
	bookingusecase "github.com/avito-internships/test-backend-1-katyaswedq/internal/usecase/booking"
)

type ListUseCase interface {
	List(ctx context.Context, input bookingusecase.ListInput) (*bookingusecase.ListOutput, error)
}

type ListHandler struct {
	usecase ListUseCase
}

func NewListHandler(usecase ListUseCase) *ListHandler {
	return &ListHandler{usecase: usecase}
}

func (h *ListHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	page := r.URL.Query().Get("page")
	pageSize := r.URL.Query().Get("pageSize")

	out, err := h.usecase.List(r.Context(), bookingusecase.ListInput{
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		switch {
		case errors.Is(err, bookingusecase.ErrInvalidPagination):
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

	handlers.WriteJSON(w, http.StatusOK, listBookingsResponse{
		Bookings: resp,
		Pagination: paginationResponse{
			Page:     out.Pagination.Page,
			PageSize: out.Pagination.PageSize,
			Total:    out.Pagination.Total,
		},
	})
}