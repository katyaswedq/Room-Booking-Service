package booking

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	domainbooking "github.com/avito-internships/test-backend-1-katyaswedq/internal/domain/booking"
)

type mockCancelBookingRepository struct {
	getByIDFn      func(ctx context.Context, id string) (*domainbooking.Booking, error)
	updateStatusFn func(ctx context.Context, id string, status string) error
}

func (m *mockCancelBookingRepository) Create(ctx context.Context, b *domainbooking.Booking) error {
	return nil
}

func (m *mockCancelBookingRepository) GetByID(ctx context.Context, id string) (*domainbooking.Booking, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, nil
}

func (m *mockCancelBookingRepository) ListByUser(ctx context.Context, userID string) ([]domainbooking.Booking, error) {
	return nil, nil
}

func (m *mockCancelBookingRepository) ListAll(ctx context.Context) ([]domainbooking.Booking, error) {
	return nil, nil
}

func (m *mockCancelBookingRepository) UpdateStatus(ctx context.Context, id string, status string) error {
	if m.updateStatusFn != nil {
		return m.updateStatusFn(ctx, id, status)
	}
	return nil
}

func TestCancelBooking_InvalidBookingID_ReturnsError(t *testing.T) {
	uc := NewCancelUseCase(&mockCancelBookingRepository{})

	out, err := uc.Cancel(context.Background(), CancelInput{
		BookingID: "bad-booking-id",
		UserID:    uuid.NewString(),
		Role:      "user",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, ErrInvalidBookingID) {
		t.Fatalf("expected ErrInvalidBookingID, got %v", err)
	}

	if out != nil {
		t.Fatalf("expected nil output, got %#v", out)
	}
}

func TestCancelBooking_BookingNotFound_ReturnsError(t *testing.T) {
	repo := &mockCancelBookingRepository{
		getByIDFn: func(ctx context.Context, id string) (*domainbooking.Booking, error) {
			return nil, nil
		},
	}

	uc := NewCancelUseCase(repo)

	out, err := uc.Cancel(context.Background(), CancelInput{
		BookingID: uuid.NewString(),
		UserID:    uuid.NewString(),
		Role:      "user",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, ErrBookingNotFound) {
		t.Fatalf("expected ErrBookingNotFound, got %v", err)
	}

	if out != nil {
		t.Fatalf("expected nil output, got %#v", out)
	}
}

func TestCancelBooking_Forbidden_ReturnsError(t *testing.T) {
	bookingID := uuid.NewString()

	repo := &mockCancelBookingRepository{
		getByIDFn: func(ctx context.Context, id string) (*domainbooking.Booking, error) {
			return &domainbooking.Booking{
				ID:             bookingID,
				SlotID:         uuid.NewString(),
				UserID:         uuid.NewString(),
				Status:         domainbooking.StatusActive,
				ConferenceLink: "",
				CreatedAt:      time.Now().UTC(),
			}, nil
		},
	}

	uc := NewCancelUseCase(repo)

	out, err := uc.Cancel(context.Background(), CancelInput{
		BookingID: bookingID,
		UserID:    uuid.NewString(),
		Role:      "user",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}

	if out != nil {
		t.Fatalf("expected nil output, got %#v", out)
	}
}

func TestCancelBooking_AlreadyCancelled_ReturnsCurrentBooking(t *testing.T) {
	bookingID := uuid.NewString()
	userID := uuid.NewString()
	createdAt := time.Now().UTC()

	updateCalled := false

	repo := &mockCancelBookingRepository{
		getByIDFn: func(ctx context.Context, id string) (*domainbooking.Booking, error) {
			return &domainbooking.Booking{
				ID:             bookingID,
				SlotID:         uuid.NewString(),
				UserID:         userID,
				Status:         domainbooking.StatusCancelled,
				ConferenceLink: "",
				CreatedAt:      createdAt,
			}, nil
		},
		updateStatusFn: func(ctx context.Context, id string, status string) error {
			updateCalled = true
			return nil
		},
	}

	uc := NewCancelUseCase(repo)

	out, err := uc.Cancel(context.Background(), CancelInput{
		BookingID: bookingID,
		UserID:    userID,
		Role:      "user",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if out == nil {
		t.Fatal("expected non-nil output")
	}

	if out.ID != bookingID {
		t.Fatalf("expected booking id %s, got %s", bookingID, out.ID)
	}

	if out.Status != domainbooking.StatusCancelled {
		t.Fatalf("expected status %s, got %s", domainbooking.StatusCancelled, out.Status)
	}

	if updateCalled {
		t.Fatal("expected UpdateStatus not to be called for already cancelled booking")
	}
}

func TestCancelBooking_Success(t *testing.T) {
	bookingID := uuid.NewString()
	slotID := uuid.NewString()
	userID := uuid.NewString()
	createdAt := time.Now().UTC()

	updateCalled := false
	updatedStatus := ""

	repo := &mockCancelBookingRepository{
		getByIDFn: func(ctx context.Context, id string) (*domainbooking.Booking, error) {
			return &domainbooking.Booking{
				ID:             bookingID,
				SlotID:         slotID,
				UserID:         userID,
				Status:         domainbooking.StatusActive,
				ConferenceLink: "",
				CreatedAt:      createdAt,
			}, nil
		},
		updateStatusFn: func(ctx context.Context, id string, status string) error {
			updateCalled = true
			updatedStatus = status
			return nil
		},
	}

	uc := NewCancelUseCase(repo)

	out, err := uc.Cancel(context.Background(), CancelInput{
		BookingID: bookingID,
		UserID:    userID,
		Role:      "user",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !updateCalled {
		t.Fatal("expected UpdateStatus to be called")
	}

	if updatedStatus != domainbooking.StatusCancelled {
		t.Fatalf("expected updated status %s, got %s", domainbooking.StatusCancelled, updatedStatus)
	}

	if out == nil {
		t.Fatal("expected non-nil output")
	}

	if out.ID != bookingID {
		t.Fatalf("expected booking id %s, got %s", bookingID, out.ID)
	}

	if out.SlotID != slotID {
		t.Fatalf("expected slot id %s, got %s", slotID, out.SlotID)
	}

	if out.UserID != userID {
		t.Fatalf("expected user id %s, got %s", userID, out.UserID)
	}

	if out.Status != domainbooking.StatusCancelled {
		t.Fatalf("expected status %s, got %s", domainbooking.StatusCancelled, out.Status)
	}
}