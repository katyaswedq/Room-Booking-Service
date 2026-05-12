package middleware

import (
	"net/http"

	"github.com/avito-internships/test-backend-1-katyaswedq/internal/handlers"
)

func RequireUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		role, ok := r.Context().Value(RoleKey).(string)
		if !ok || role != "user" {
			handlers.WriteError(w, http.StatusForbidden, "FORBIDDEN", "user role required")
			return
		}

		next.ServeHTTP(w, r)
	})
}