package middleware

import (
	"context"
	"net/http"
	"strings"

	handlers "github.com/avito-internships/test-backend-1-katyaswedq/internal/handlers"
	infraAuth "github.com/avito-internships/test-backend-1-katyaswedq/internal/infrastructure/auth"
)

type contextKey string

const (
	UserIDKey contextKey = "user_id"
	RoleKey   contextKey = "role"
)

func JWT(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				handlers.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing authorization header")
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				handlers.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "invalid authorization header")
				return
			}

			tokenString := parts[1]

			claims, err := infraAuth.ParseToken(tokenString, secret)
			if err != nil {
				handlers.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "invalid token")
				return
			}

			ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
			ctx = context.WithValue(ctx, RoleKey, claims.Role)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}