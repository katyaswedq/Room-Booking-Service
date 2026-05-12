package schedule

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	handlers "github.com/avito-internships/test-backend-1-katyaswedq/internal/handlers"
	scheduleUsecase "github.com/avito-internships/test-backend-1-katyaswedq/internal/usecase/schedule"
)

type CreateUseCase interface {
	Create(ctx context.Context, input scheduleUsecase.CreateInput) (*scheduleUsecase.CreateOutput, error)
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
	roomID := chi.URLParam(r, "roomId")
	if roomID == "" {
		handlers.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "room id is required")
		return
	}

	var req CreateScheduleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		handlers.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
		return
	}

	out, err := h.usecase.Create(r.Context(), scheduleUsecase.CreateInput{
		RoomID:     roomID,
		DaysOfWeek: req.DaysOfWeek,
		StartTime:  req.StartTime,
		EndTime:    req.EndTime,
	})
	if err != nil {
		switch {
		case errors.Is(err, scheduleUsecase.ErrInvalidDaysOfWeek),
			errors.Is(err, scheduleUsecase.ErrInvalidTimeRange):
			handlers.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		case errors.Is(err, scheduleUsecase.ErrRoomNotFound):
			handlers.WriteError(w, http.StatusNotFound, "ROOM_NOT_FOUND", err.Error())
		case errors.Is(err, scheduleUsecase.ErrScheduleExists):
			handlers.WriteError(w, http.StatusConflict, "SCHEDULE_EXISTS", "schedule for this room already exists and cannot be changed")
		default:
			handlers.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		}
		return
	}

	handlers.WriteJSON(w, http.StatusCreated, createScheduleResponse{
		Schedule: scheduleResponse{
			ID:         out.ID,
			RoomID:     out.RoomID,
			DaysOfWeek: out.DaysOfWeek,
			StartTime:  out.StartTime,
			EndTime:    out.EndTime,
		},
	})
}