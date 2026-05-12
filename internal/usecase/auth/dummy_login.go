package auth

import (
	"fmt"

	infraAuth "github.com/avito-internships/test-backend-1-katyaswedq/internal/infrastructure/auth"
)

const (
	adminID = "11111111-1111-1111-1111-111111111111"
	userID  = "22222222-2222-2222-2222-222222222222"
)

type DummyLoginUseCase struct {
	jwtSecret string
}

func NewDummyLoginUseCase(jwtSecret string) *DummyLoginUseCase {
	return &DummyLoginUseCase{
		jwtSecret: jwtSecret,
	}
}

func (u *DummyLoginUseCase) DummyLogin(role string) (string, error) {
	var id string

	switch role {
	case "admin":
		id = adminID
	case "user":
		id = userID
	default:
		return "", fmt.Errorf("invalid role")
	}

	return infraAuth.GenerateToken(id, role, u.jwtSecret)
}