package booking

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	domainbooking "github.com/avito-internships/test-backend-1-katyaswedq/internal/domain/booking"
	domainslot "github.com/avito-internships/test-backend-1-katyaswedq/internal/domain/slot"
)

type mockMyBookingRepository struct {
	listByUserFn func(ctx context.Context, userID string) ([]domainbooking.Booking, error)
}

func (m *mockMyBookingRepository) Create(ctx context.Context, b *domainbooking.Booking) error {
	return nil
}

func (m *mockMyBookingRepository) GetByID(ctx context.Context, id string) (*domainbooking.Booking, error) {
	return nil, nil
}

func (m *mockMyBookingRepository) ListByUser(ctx context.Context, userID string) ([]domainbooking.Booking, error) {
	if m.listByUserFn != nil {
		return m.listByUserFn(ctx, userID)
	}
	return nil, nil
}

func (m *mockMyBookingRepository) ListAll(ctx context.Context) ([]domainbooking.Booking, error) {
	return nil, nil
}

func (m *mockMyBookingRepository) UpdateStatus(ctx context.Context, id string, status string) error {
	return nil
}

type mockMySlotRepository struct {
	getByIDFn func(ctx context.Context, id string) (*domainslot.Slot, error)
}

func (m *mockMySlotRepository) BulkCreate(ctx context.Context, slots []domainslot.Slot) error {
	return nil
}

func (m *mockMySlotRepository) ListAvailableByRoomAndDate(ctx context.Context, roomID string, date time.Time) ([]domainslot.Slot, error) {
	return nil, nil
}

func (m *mockMySlotRepository) ExistsByID(ctx context.Context, id string) (bool, error) {
	return false, nil
}

func (m *mockMySlotRepository) GetByID(ctx context.Context, id string) (*domainslot.Slot, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, nil
}

func TestMyBookings_InvalidUserID_ReturnsError(t *testing.T) {
	uc := NewMyUseCase(&mockMyBookingRepository{}, &mockMySlotRepository{})

	out, err := uc.List(context.Background(), MyInput{
		UserID: "bad-user-id",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, ErrInvalidUserID) {
		t.Fatalf("expected ErrInvalidUserID, got %v", err)
	}

	if out != nil {
		t.Fatalf("expected nil output, got %#v", out)
	}
}

func TestMyBookings_ReturnsOnlyFutureBookings(t *testing.T) {
	userID := uuid.NewString()

	futureSlotID := uuid.NewString()
	pastSlotID := uuid.NewString()
	missingSlotID := uuid.NewString()

	now := time.Now().UTC()

	repo := &mockMyBookingRepository{
		listByUserFn: func(ctx context.Context, inputUserID string) ([]domainbooking.Booking, error) {
			return []domainbooking.Booking{
				{
					ID:             "future-booking",
					SlotID:         futureSlotID,
					UserID:         inputUserID,
					Status:         domainbooking.StatusActive,
					ConferenceLink: "",
					CreatedAt:      now,
				},
				{
					ID:             "past-booking",
					SlotID:         pastSlotID,
					UserID:         inputUserID,
					Status:         domainbooking.StatusCancelled,
					ConferenceLink: "",
					CreatedAt:      now,
				},
				{
					ID:             "missing-slot-booking",
					SlotID:         missingSlotID,
					UserID:         inputUserID,
					Status:         domainbooking.StatusActive,
					ConferenceLink: "",
					CreatedAt:      now,
				},
			}, nil
		},
	}

	slotRepo := &mockMySlotRepository{
		getByIDFn: func(ctx context.Context, id string) (*domainslot.Slot, error) {
			switch id {
			case futureSlotID:
				return &domainslot.Slot{
					ID:     id,
					RoomID: uuid.NewString(),
					Start:  now.Add(2 * time.Hour),
					End:    now.Add(150 * time.Minute),
				}, nil
			case pastSlotID:
				return &domainslot.Slot{
					ID:     id,
					RoomID: uuid.NewString(),
					Start:  now.Add(-2 * time.Hour),
					End:    now.Add(-90 * time.Minute),
				}, nil
			case missingSlotID:
				return nil, nil
			default:
				return nil, nil
			}
		},
	}

	uc := NewMyUseCase(repo, slotRepo)

	out, err := uc.List(context.Background(), MyInput{
		UserID: userID,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if out == nil {
		t.Fatal("expected non-nil output")
	}

	if len(out.Bookings) != 1 {
		t.Fatalf("expected 1 future booking, got %d", len(out.Bookings))
	}

	if out.Bookings[0].ID != "future-booking" {
		t.Fatalf("expected future-booking, got %s", out.Bookings[0].ID)
	}

	if out.Bookings[0].SlotID != futureSlotID {
		t.Fatalf("expected slot id %s, got %s", futureSlotID, out.Bookings[0].SlotID)
	}
}

func TestMyBookings_SlotRepoError_ReturnsError(t *testing.T) {
	userID := uuid.NewString()
	slotID := uuid.NewString()
	expectedErr := errors.New("slot repo error")

	repo := &mockMyBookingRepository{
		listByUserFn: func(ctx context.Context, inputUserID string) ([]domainbooking.Booking, error) {
			return []domainbooking.Booking{
				{
					ID:             uuid.NewString(),
					SlotID:         slotID,
					UserID:         inputUserID,
					Status:         domainbooking.StatusActive,
					ConferenceLink: "",
					CreatedAt:      time.Now().UTC(),
				},
			}, nil
		},
	}

	slotRepo := &mockMySlotRepository{
		getByIDFn: func(ctx context.Context, id string) (*domainslot.Slot, error) {
			return nil, expectedErr
		},
	}

	uc := NewMyUseCase(repo, slotRepo)

	out, err := uc.List(context.Background(), MyInput{
		UserID: userID,
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

func TestMyBookings_RepositoryError_ReturnsError(t *testing.T) {
	expectedErr := errors.New("booking repo error")

	repo := &mockMyBookingRepository{
		listByUserFn: func(ctx context.Context, inputUserID string) ([]domainbooking.Booking, error) {
			return nil, expectedErr
		},
	}

	uc := NewMyUseCase(repo, &mockMySlotRepository{})

	out, err := uc.List(context.Background(), MyInput{
		UserID: uuid.NewString(),
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