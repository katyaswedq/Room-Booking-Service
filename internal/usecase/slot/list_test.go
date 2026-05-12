package slot

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	domainroom "github.com/avito-internships/test-backend-1-katyaswedq/internal/domain/room"
	domainslot "github.com/avito-internships/test-backend-1-katyaswedq/internal/domain/slot"
)

type mockListSlotRepository struct {
	listAvailableByRoomAndDateFn func(ctx context.Context, roomID string, date time.Time) ([]domainslot.Slot, error)
}

func (m *mockListSlotRepository) BulkCreate(ctx context.Context, slots []domainslot.Slot) error {
	return nil
}

func (m *mockListSlotRepository) ListAvailableByRoomAndDate(ctx context.Context, roomID string, date time.Time) ([]domainslot.Slot, error) {
	if m.listAvailableByRoomAndDateFn != nil {
		return m.listAvailableByRoomAndDateFn(ctx, roomID, date)
	}
	return nil, nil
}

func (m *mockListSlotRepository) ExistsByID(ctx context.Context, id string) (bool, error) {
	return false, nil
}

func (m *mockListSlotRepository) GetByID(ctx context.Context, id string) (*domainslot.Slot, error) {
	return nil, nil
}

type mockListRoomRepository struct {
	existsByIDFn func(ctx context.Context, id string) (bool, error)
}

func (m *mockListRoomRepository) Create(ctx context.Context, room *domainroom.Room) error {
	return nil
}

func (m *mockListRoomRepository) List(ctx context.Context) ([]domainroom.Room, error) {
	return nil, nil
}

func (m *mockListRoomRepository) ExistsByID(ctx context.Context, id string) (bool, error) {
	if m.existsByIDFn != nil {
		return m.existsByIDFn(ctx, id)
	}
	return false, nil
}

func TestListSlots_EmptyRoomID_ReturnsRoomNotFound(t *testing.T) {
	uc := NewListUseCase(&mockListSlotRepository{}, &mockListRoomRepository{})

	out, err := uc.List(context.Background(), ListInput{
		RoomID: "",
		Date:   "2026-03-26",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, ErrRoomNotFound) {
		t.Fatalf("expected ErrRoomNotFound, got %v", err)
	}

	if out != nil {
		t.Fatalf("expected nil output, got %#v", out)
	}
}

func TestListSlots_InvalidRoomID_ReturnsRoomNotFound(t *testing.T) {
	uc := NewListUseCase(&mockListSlotRepository{}, &mockListRoomRepository{})

	out, err := uc.List(context.Background(), ListInput{
		RoomID: "bad-room-id",
		Date:   "2026-03-26",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, ErrRoomNotFound) {
		t.Fatalf("expected ErrRoomNotFound, got %v", err)
	}

	if out != nil {
		t.Fatalf("expected nil output, got %#v", out)
	}
}

func TestListSlots_EmptyDate_ReturnsInvalidDate(t *testing.T) {
	uc := NewListUseCase(&mockListSlotRepository{}, &mockListRoomRepository{})

	out, err := uc.List(context.Background(), ListInput{
		RoomID: uuid.NewString(),
		Date:   "",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, ErrInvalidDate) {
		t.Fatalf("expected ErrInvalidDate, got %v", err)
	}

	if out != nil {
		t.Fatalf("expected nil output, got %#v", out)
	}
}

func TestListSlots_InvalidDateFormat_ReturnsInvalidDate(t *testing.T) {
	uc := NewListUseCase(&mockListSlotRepository{}, &mockListRoomRepository{})

	out, err := uc.List(context.Background(), ListInput{
		RoomID: uuid.NewString(),
		Date:   "26-03-2026",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, ErrInvalidDate) {
		t.Fatalf("expected ErrInvalidDate, got %v", err)
	}

	if out != nil {
		t.Fatalf("expected nil output, got %#v", out)
	}
}

func TestListSlots_RoomNotFound_ReturnsError(t *testing.T) {
	roomID := uuid.NewString()

	roomRepo := &mockListRoomRepository{
		existsByIDFn: func(ctx context.Context, id string) (bool, error) {
			return false, nil
		},
	}

	uc := NewListUseCase(&mockListSlotRepository{}, roomRepo)

	out, err := uc.List(context.Background(), ListInput{
		RoomID: roomID,
		Date:   "2026-03-26",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, ErrRoomNotFound) {
		t.Fatalf("expected ErrRoomNotFound, got %v", err)
	}

	if out != nil {
		t.Fatalf("expected nil output, got %#v", out)
	}
}

func TestListSlots_RoomRepoError_ReturnsError(t *testing.T) {
	expectedErr := errors.New("room repo error")
	roomID := uuid.NewString()

	roomRepo := &mockListRoomRepository{
		existsByIDFn: func(ctx context.Context, id string) (bool, error) {
			return false, expectedErr
		},
	}

	uc := NewListUseCase(&mockListSlotRepository{}, roomRepo)

	out, err := uc.List(context.Background(), ListInput{
		RoomID: roomID,
		Date:   "2026-03-26",
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

func TestListSlots_SlotRepoError_ReturnsError(t *testing.T) {
	expectedErr := errors.New("slot repo error")
	roomID := uuid.NewString()

	roomRepo := &mockListRoomRepository{
		existsByIDFn: func(ctx context.Context, id string) (bool, error) {
			return true, nil
		},
	}

	slotRepo := &mockListSlotRepository{
		listAvailableByRoomAndDateFn: func(ctx context.Context, gotRoomID string, date time.Time) ([]domainslot.Slot, error) {
			return nil, expectedErr
		},
	}

	uc := NewListUseCase(slotRepo, roomRepo)

	out, err := uc.List(context.Background(), ListInput{
		RoomID: roomID,
		Date:   "2026-03-26",
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

func TestListSlots_Success_ReturnsMappedSlots(t *testing.T) {
	roomID := uuid.NewString()
	dateStr := "2026-03-26"

	start1 := time.Date(2026, 3, 26, 9, 0, 0, 0, time.UTC)
	end1 := start1.Add(30 * time.Minute)
	start2 := time.Date(2026, 3, 26, 9, 30, 0, 0, time.UTC)
	end2 := start2.Add(30 * time.Minute)

	roomRepo := &mockListRoomRepository{
		existsByIDFn: func(ctx context.Context, id string) (bool, error) {
			return true, nil
		},
	}

	slotRepo := &mockListSlotRepository{
		listAvailableByRoomAndDateFn: func(ctx context.Context, gotRoomID string, date time.Time) ([]domainslot.Slot, error) {
			return []domainslot.Slot{
				{
					ID:     "slot-1",
					RoomID: roomID,
					Start:  start1,
					End:    end1,
				},
				{
					ID:     "slot-2",
					RoomID: roomID,
					Start:  start2,
					End:    end2,
				},
			}, nil
		},
	}

	uc := NewListUseCase(slotRepo, roomRepo)

	out, err := uc.List(context.Background(), ListInput{
		RoomID: roomID,
		Date:   dateStr,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if out == nil {
		t.Fatal("expected non-nil output")
	}

	if len(out.Slots) != 2 {
		t.Fatalf("expected 2 slots, got %d", len(out.Slots))
	}

	if out.Slots[0].ID != "slot-1" {
		t.Fatalf("expected first slot id slot-1, got %s", out.Slots[0].ID)
	}

	if out.Slots[0].RoomID != roomID {
		t.Fatalf("expected first slot roomID %s, got %s", roomID, out.Slots[0].RoomID)
	}

	if !out.Slots[0].Start.Equal(start1) {
		t.Fatalf("expected first slot start %v, got %v", start1, out.Slots[0].Start)
	}

	if !out.Slots[1].End.Equal(end2) {
		t.Fatalf("expected second slot end %v, got %v", end2, out.Slots[1].End)
	}
}