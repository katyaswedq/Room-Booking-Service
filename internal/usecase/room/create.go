package room

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"

	domainRoom "github.com/avito-internships/test-backend-1-katyaswedq/internal/domain/room"
)

var (
	ErrInvalidName     = errors.New("invalid name")
	ErrInvalidCapacity = errors.New("invalid capacity")
)

type CreateRoomUseCase struct {
	repo domainRoom.Repository
}

type CreateRoomInput struct {
	Name        string
	Description string
	Capacity    int
}

type CreateRoomOutput struct {
	ID          string
	Name        string
	Description string
	Capacity    int
	CreatedAt   time.Time
}

func NewCreateRoomUseCase(repo domainRoom.Repository) *CreateRoomUseCase {
	return &CreateRoomUseCase{
		repo: repo,
	}
}

func (uc *CreateRoomUseCase) Create(ctx context.Context, input CreateRoomInput) (*CreateRoomOutput, error) {
	name := strings.TrimSpace(input.Name)
	description := strings.TrimSpace(input.Description)

	if name == "" {
		return nil, ErrInvalidName
	}

	if input.Capacity <= 0 {
		return nil, ErrInvalidCapacity
	}

	rm := &domainRoom.Room{
		ID:          uuid.NewString(),
		Name:        name,
		Description: description,
		Capacity:    input.Capacity,
		CreatedAt:   time.Now().UTC(),
	}

	if err := uc.repo.Create(ctx, rm); err != nil {
		return nil, err
	}

	return &CreateRoomOutput{
		ID:          rm.ID,
		Name:        rm.Name,
		Description: rm.Description,
		Capacity:    rm.Capacity,
		CreatedAt:   rm.CreatedAt,
	}, nil
}