package room

import (
	"context"
	"errors"
	"testing"
	"time"

	domainroom "github.com/avito-internships/test-backend-1-katyaswedq/internal/domain/room"
)

type mockListRoomRepository struct {
	listFn func(ctx context.Context) ([]domainroom.Room, error)
}

func (m *mockListRoomRepository) Create(ctx context.Context, room *domainroom.Room) error {
	return nil
}

func (m *mockListRoomRepository) List(ctx context.Context) ([]domainroom.Room, error) {
	if m.listFn != nil {
		return m.listFn(ctx)
	}
	return nil, nil
}

func (m *mockListRoomRepository) ExistsByID(ctx context.Context, id string) (bool, error) {
	return false, nil
}

func TestListRooms_Success_ReturnsMappedRooms(t *testing.T) {
	createdAt := time.Now().UTC()

	repo := &mockListRoomRepository{
		listFn: func(ctx context.Context) ([]domainroom.Room, error) {
			return []domainroom.Room{
				{
					ID:          "room-1",
					Name:        "Room A",
					Description: "First room",
					Capacity:    6,
					CreatedAt:   createdAt,
				},
				{
					ID:          "room-2",
					Name:        "Room B",
					Description: "Second room",
					Capacity:    8,
					CreatedAt:   createdAt,
				},
			}, nil
		},
	}

	uc := NewListRoomsUseCase(repo)

	out, err := uc.List(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(out) != 2 {
		t.Fatalf("expected 2 rooms, got %d", len(out))
	}

	if out[0].ID != "room-1" {
		t.Fatalf("expected first room id room-1, got %s", out[0].ID)
	}

	if out[0].Name != "Room A" {
		t.Fatalf("expected first room name Room A, got %s", out[0].Name)
	}

	if out[1].ID != "room-2" {
		t.Fatalf("expected second room id room-2, got %s", out[1].ID)
	}

	if out[1].Capacity != 8 {
		t.Fatalf("expected second room capacity 8, got %d", out[1].Capacity)
	}
}

func TestListRooms_Empty_ReturnsEmptySlice(t *testing.T) {
	repo := &mockListRoomRepository{
		listFn: func(ctx context.Context) ([]domainroom.Room, error) {
			return []domainroom.Room{}, nil
		},
	}

	uc := NewListRoomsUseCase(repo)

	out, err := uc.List(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(out) != 0 {
		t.Fatalf("expected 0 rooms, got %d", len(out))
	}
}

func TestListRooms_RepositoryError_ReturnsError(t *testing.T) {
	expectedErr := errors.New("repo error")

	repo := &mockListRoomRepository{
		listFn: func(ctx context.Context) ([]domainroom.Room, error) {
			return nil, expectedErr
		},
	}

	uc := NewListRoomsUseCase(repo)

	out, err := uc.List(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}

	if out != nil {
		t.Fatalf("expected nil output, got %#v", out)
	}
}