package auth

import (
	"context"
	"errors"
	"strings"

	"golang.org/x/crypto/bcrypt"

	domainuser "github.com/avito-internships/test-backend-1-katyaswedq/internal/domain/user"
	infraauth "github.com/avito-internships/test-backend-1-katyaswedq/internal/infrastructure/auth"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
)

type LoginUseCase struct {
	userRepo domainuser.Repository
	secret   string
}

func NewLoginUseCase(userRepo domainuser.Repository, secret string) *LoginUseCase {
	return &LoginUseCase{
		userRepo: userRepo,
		secret:   secret,
	}
}

type LoginInput struct {
	Email    string
	Password string
}

type LoginOutput struct {
	Token string
}

func (uc *LoginUseCase) Execute(ctx context.Context, input LoginInput) (LoginOutput, error) {
	email := strings.TrimSpace(input.Email)
	password := strings.TrimSpace(input.Password)

	if email == "" || password == "" {
		return LoginOutput{}, ErrInvalidCredentials
	}

	user, err := uc.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return LoginOutput{}, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return LoginOutput{}, ErrInvalidCredentials
	}

	token, err := infraauth.GenerateToken(user.ID, string(user.Role), uc.secret)
	if err != nil {
		return LoginOutput{}, err
	}

	return LoginOutput{
		Token: token,
	}, nil
}