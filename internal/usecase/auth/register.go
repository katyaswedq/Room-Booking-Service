package auth

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	domainuser "github.com/avito-internships/test-backend-1-katyaswedq/internal/domain/user"
)

var (
	ErrInvalidEmail    = errors.New("invalid email")
	ErrInvalidPassword = errors.New("invalid password")
	ErrInvalidRole     = errors.New("invalid role")
	ErrEmailAlreadyUsed = errors.New("email already used")
)

type RegisterUseCase struct {
	userRepo domainuser.Repository
}

func NewRegisterUseCase(userRepo domainuser.Repository) *RegisterUseCase {
	return &RegisterUseCase{
		userRepo: userRepo,
	}
}

type RegisterInput struct {
	Email    string
	Password string
	Role     string
}

type RegisterOutput struct {
	User domainuser.User
}

func (uc *RegisterUseCase) Execute(ctx context.Context, input RegisterInput) (RegisterOutput, error) {
	email := strings.TrimSpace(input.Email)
	password := strings.TrimSpace(input.Password)
	role := strings.TrimSpace(input.Role)

	if email == "" || !strings.Contains(email, "@") {
		return RegisterOutput{}, ErrInvalidEmail
	}

	if len(password) < 6 {
		return RegisterOutput{}, ErrInvalidPassword
	}

	if role != string(domainuser.RoleAdmin) && role != string(domainuser.RoleUser) {
		return RegisterOutput{}, ErrInvalidRole
	}

	exists, err := uc.userRepo.ExistsByEmail(ctx, email)
	if err != nil {
		return RegisterOutput{}, err
	}
	if exists {
		return RegisterOutput{}, ErrEmailAlreadyUsed
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return RegisterOutput{}, err
	}

	user := domainuser.User{
		ID:           uuid.New().String(),
		Email:        email,
		PasswordHash: string(hash),
		Role:         domainuser.Role(role),
		CreatedAt:    time.Time{}, 
	}

	createdUser, err := uc.userRepo.Create(ctx, user)
	if err != nil {
		return RegisterOutput{}, err
	}

	return RegisterOutput{
		User: createdUser,
	}, nil
}