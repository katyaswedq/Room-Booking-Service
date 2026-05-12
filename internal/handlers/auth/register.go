package auth

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/avito-internships/test-backend-1-katyaswedq/internal/handlers"
	usecaseauth "github.com/avito-internships/test-backend-1-katyaswedq/internal/usecase/auth"
)

type RegisterHandler struct {
	useCase *usecaseauth.RegisterUseCase
}

func NewRegisterHandler(useCase *usecaseauth.RegisterUseCase) *RegisterHandler {
	return &RegisterHandler{
		useCase: useCase,
	}
}

func (h *RegisterHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		handlers.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request")
		return
	}

	out, err := h.useCase.Execute(r.Context(), usecaseauth.RegisterInput{
		Email:    req.Email,
		Password: req.Password,
		Role:     req.Role,
	})
	if err != nil {
		switch {
		case errors.Is(err, usecaseauth.ErrInvalidEmail),
			errors.Is(err, usecaseauth.ErrInvalidPassword),
			errors.Is(err, usecaseauth.ErrInvalidRole),
			errors.Is(err, usecaseauth.ErrEmailAlreadyUsed):
			handlers.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		default:
			handlers.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		}
		return
	}

	handlers.WriteJSON(w, http.StatusCreated, RegisterResponse{
		User: RegisterUser{
			ID:        out.User.ID,
			Email:     out.User.Email,
			Role:      string(out.User.Role),
			CreatedAt: formatTimePtr(out.User.CreatedAt),
		},
	})
}

func formatTimePtr(t time.Time) *string {
	if t.IsZero() {
		return nil
	}

	value := t.UTC().Format(time.RFC3339)
	return &value
}