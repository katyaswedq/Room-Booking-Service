package auth

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/avito-internships/test-backend-1-katyaswedq/internal/handlers"
	usecaseauth "github.com/avito-internships/test-backend-1-katyaswedq/internal/usecase/auth"
)

type LoginHandler struct {
	useCase *usecaseauth.LoginUseCase
}

func NewLoginHandler(useCase *usecaseauth.LoginUseCase) *LoginHandler {
	return &LoginHandler{
		useCase: useCase,
	}
}

func (h *LoginHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		handlers.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request")
		return
	}

	out, err := h.useCase.Execute(r.Context(), usecaseauth.LoginInput{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		switch {
		case errors.Is(err, usecaseauth.ErrInvalidCredentials):
			handlers.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "invalid credentials")
		default:
			handlers.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		}
		return
	}

	handlers.WriteJSON(w, http.StatusOK, LoginResponse{
		Token: out.Token,
	})
}