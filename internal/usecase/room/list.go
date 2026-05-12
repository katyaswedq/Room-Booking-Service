package room

import (
	"context"
	"time"

	domainroom "github.com/avito-internships/test-backend-1-katyaswedq/internal/domain/room"
)

type ListRoomsUseCase struct {
	repo domainroom.Repository
}

type ListRoomOutput struct {
	ID          string
	Name        string
	Description string
	Capacity    int
	CreatedAt   time.Time
}

func NewListRoomsUseCase(repo domainroom.Repository) *ListRoomsUseCase {
	return &ListRoomsUseCase{
		repo: repo,
	}
}

func (uc *ListRoomsUseCase) List(ctx context.Context) ([]ListRoomOutput, error) {
	rooms, err := uc.repo.List(ctx)
	if err != nil {
		return nil, err
	}

	out := make([]ListRoomOutput, 0, len(rooms))
	for _, rm := range rooms {
		out = append(out, ListRoomOutput{
			ID:          rm.ID,
			Name:        rm.Name,
			Description: rm.Description,
			Capacity:    rm.Capacity,
			CreatedAt:   rm.CreatedAt,
		})
	}

	return out, nil
}