package room

import (
	"context"
	"errors"
	"testing"

	domainroom "github.com/avito-internships/test-backend-1-katyaswedq/internal/domain/room"
)

type mockCreateRoomRepository struct {
	createFn func(ctx context.Context, room *domainroom.Room) error
}

func (m *mockCreateRoomRepository) Create(ctx context.Context, room *domainroom.Room) error {
	if m.createFn != nil {
		return m.createFn(ctx, room)
	}
	return nil
}

func (m *mockCreateRoomRepository) List(ctx context.Context) ([]domainroom.Room, error) {
	return nil, nil
}

func (m *mockCreateRoomRepository) ExistsByID(ctx context.Context, id string) (bool, error) {
	return false, nil
}

func TestCreateRoom_InvalidName_ReturnsError(t *testing.T) {
	uc := NewCreateRoomUseCase(&mockCreateRoomRepository{})

	out, err := uc.Create(context.Background(), CreateRoomInput{
		Name:        "   ",
		Description: "desc",
		Capacity:    6,
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, ErrInvalidName) {
		t.Fatalf("expected ErrInvalidName, got %v", err)
	}

	if out != nil {
		t.Fatalf("expected nil output, got %#v", out)
	}
}

func TestCreateRoom_InvalidCapacity_ReturnsError(t *testing.T) {
	uc := NewCreateRoomUseCase(&mockCreateRoomRepository{})

	out, err := uc.Create(context.Background(), CreateRoomInput{
		Name:        "Room A",
		Description: "desc",
		Capacity:    0,
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, ErrInvalidCapacity) {
		t.Fatalf("expected ErrInvalidCapacity, got %v", err)
	}

	if out != nil {
		t.Fatalf("expected nil output, got %#v", out)
	}
}

func TestCreateRoom_Success_TrimsFieldsAndReturnsOutput(t *testing.T) {
	var createdRoom *domainroom.Room

	repo := &mockCreateRoomRepository{
		createFn: func(ctx context.Context, room *domainroom.Room) error {
			createdRoom = room
			return nil
		},
	}

	uc := NewCreateRoomUseCase(repo)

	out, err := uc.Create(context.Background(), CreateRoomInput{
		Name:        "  Room A  ",
		Description: "  First room  ",
		Capacity:    6,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if createdRoom == nil {
		t.Fatal("expected repository Create to be called")
	}

	if createdRoom.ID == "" {
		t.Fatal("expected created room ID to be set")
	}

	if createdRoom.Name != "Room A" {
		t.Fatalf("expected trimmed name %q, got %q", "Room A", createdRoom.Name)
	}

	if createdRoom.Description != "First room" {
		t.Fatalf("expected trimmed description %q, got %q", "First room", createdRoom.Description)
	}

	if createdRoom.Capacity != 6 {
		t.Fatalf("expected capacity 6, got %d", createdRoom.Capacity)
	}

	if out == nil {
		t.Fatal("expected non-nil output")
	}

	if out.ID == "" {
		t.Fatal("expected output ID to be set")
	}

	if out.Name != "Room A" {
		t.Fatalf("expected output name %q, got %q", "Room A", out.Name)
	}

	if out.Description != "First room" {
		t.Fatalf("expected output description %q, got %q", "First room", out.Description)
	}

	if out.Capacity != 6 {
		t.Fatalf("expected output capacity 6, got %d", out.Capacity)
	}
}

func TestCreateRoom_RepositoryError_ReturnsError(t *testing.T) {
	expectedErr := errors.New("repo error")

	repo := &mockCreateRoomRepository{
		createFn: func(ctx context.Context, room *domainroom.Room) error {
			return expectedErr
		},
	}

	uc := NewCreateRoomUseCase(repo)

	out, err := uc.Create(context.Background(), CreateRoomInput{
		Name:        "Room A",
		Description: "desc",
		Capacity:    6,
	})
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