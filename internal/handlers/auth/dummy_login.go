package auth

import (
	"encoding/json"
	"net/http"

	handlers "github.com/avito-internships/test-backend-1-katyaswedq/internal/handlers"
	usecaseAuth "github.com/avito-internships/test-backend-1-katyaswedq/internal/usecase/auth"
)

type Handler struct {
	usecase *usecaseAuth.DummyLoginUseCase
}

func NewHandler(usecase *usecaseAuth.DummyLoginUseCase) *Handler {
	return &Handler{
		usecase: usecase,
	}
}

func (h *Handler) DummyLogin(w http.ResponseWriter, r *http.Request) {
	var req DummyLoginRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		handlers.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
		return
	}

	token, err := h.usecase.DummyLogin(req.Role)
	if err != nil {
		handlers.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid role")
		return
	}

	handlers.WriteJSON(w, http.StatusOK, DummyLoginResponse{
		Token: token,
	})
}