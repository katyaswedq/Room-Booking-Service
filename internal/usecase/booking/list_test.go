package booking

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	domainbooking "github.com/avito-internships/test-backend-1-katyaswedq/internal/domain/booking"
)

type mockListBookingRepository struct {
	listAllFn func(ctx context.Context) ([]domainbooking.Booking, error)
}

func (m *mockListBookingRepository) Create(ctx context.Context, b *domainbooking.Booking) error {
	return nil
}

func (m *mockListBookingRepository) GetByID(ctx context.Context, id string) (*domainbooking.Booking, error) {
	return nil, nil
}

func (m *mockListBookingRepository) ListByUser(ctx context.Context, userID string) ([]domainbooking.Booking, error) {
	return nil, nil
}

func (m *mockListBookingRepository) ListAll(ctx context.Context) ([]domainbooking.Booking, error) {
	if m.listAllFn != nil {
		return m.listAllFn(ctx)
	}
	return nil, nil
}

func (m *mockListBookingRepository) UpdateStatus(ctx context.Context, id string, status string) error {
	return nil
}

func TestListBookings_DefaultPagination_ReturnsAllWhenLessThanDefaultPageSize(t *testing.T) {
	createdAt := time.Now().UTC()

	repo := &mockListBookingRepository{
		listAllFn: func(ctx context.Context) ([]domainbooking.Booking, error) {
			return []domainbooking.Booking{
				{
					ID:             uuid.NewString(),
					SlotID:         uuid.NewString(),
					UserID:         uuid.NewString(),
					Status:         domainbooking.StatusActive,
					ConferenceLink: "",
					CreatedAt:      createdAt,
				},
				{
					ID:             uuid.NewString(),
					SlotID:         uuid.NewString(),
					UserID:         uuid.NewString(),
					Status:         domainbooking.StatusCancelled,
					ConferenceLink: "",
					CreatedAt:      createdAt,
				},
			}, nil
		},
	}

	uc := NewListUseCase(repo)

	out, err := uc.List(context.Background(), ListInput{})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if out == nil {
		t.Fatal("expected non-nil output")
	}

	if len(out.Bookings) != 2 {
		t.Fatalf("expected 2 bookings, got %d", len(out.Bookings))
	}

	if out.Pagination.Page != 1 {
		t.Fatalf("expected page 1, got %d", out.Pagination.Page)
	}

	if out.Pagination.PageSize != 20 {
		t.Fatalf("expected pageSize 20, got %d", out.Pagination.PageSize)
	}

	if out.Pagination.Total != 2 {
		t.Fatalf("expected total 2, got %d", out.Pagination.Total)
	}
}

func TestListBookings_InvalidPage_ReturnsError(t *testing.T) {
	uc := NewListUseCase(&mockListBookingRepository{})

	out, err := uc.List(context.Background(), ListInput{
		Page: "0",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, ErrInvalidPagination) {
		t.Fatalf("expected ErrInvalidPagination, got %v", err)
	}

	if out != nil {
		t.Fatalf("expected nil output, got %#v", out)
	}
}

func TestListBookings_InvalidPageSize_ReturnsError(t *testing.T) {
	uc := NewListUseCase(&mockListBookingRepository{})

	out, err := uc.List(context.Background(), ListInput{
		PageSize: "-1",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, ErrInvalidPagination) {
		t.Fatalf("expected ErrInvalidPagination, got %v", err)
	}

	if out != nil {
		t.Fatalf("expected nil output, got %#v", out)
	}
}

func TestListBookings_PageSizeGreaterThanMax_UsesMaxPageSize(t *testing.T) {
	bookings := make([]domainbooking.Booking, 0, 150)
	createdAt := time.Now().UTC()

	for i := 0; i < 150; i++ {
		bookings = append(bookings, domainbooking.Booking{
			ID:             uuid.NewString(),
			SlotID:         uuid.NewString(),
			UserID:         uuid.NewString(),
			Status:         domainbooking.StatusActive,
			ConferenceLink: "",
			CreatedAt:      createdAt,
		})
	}

	repo := &mockListBookingRepository{
		listAllFn: func(ctx context.Context) ([]domainbooking.Booking, error) {
			return bookings, nil
		},
	}

	uc := NewListUseCase(repo)

	out, err := uc.List(context.Background(), ListInput{
		Page:     "1",
		PageSize: "1000",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if out == nil {
		t.Fatal("expected non-nil output")
	}

	if out.Pagination.PageSize != 100 {
		t.Fatalf("expected pageSize 100, got %d", out.Pagination.PageSize)
	}

	if len(out.Bookings) != 100 {
		t.Fatalf("expected 100 bookings on page, got %d", len(out.Bookings))
	}

	if out.Pagination.Total != 150 {
		t.Fatalf("expected total 150, got %d", out.Pagination.Total)
	}
}

func TestListBookings_SecondPage_ReturnsCorrectSlice(t *testing.T) {
	createdAt := time.Now().UTC()

	bookings := []domainbooking.Booking{
		{
			ID:             "booking-1",
			SlotID:         uuid.NewString(),
			UserID:         uuid.NewString(),
			Status:         domainbooking.StatusActive,
			ConferenceLink: "",
			CreatedAt:      createdAt,
		},
		{
			ID:             "booking-2",
			SlotID:         uuid.NewString(),
			UserID:         uuid.NewString(),
			Status:         domainbooking.StatusActive,
			ConferenceLink: "",
			CreatedAt:      createdAt,
		},
		{
			ID:             "booking-3",
			SlotID:         uuid.NewString(),
			UserID:         uuid.NewString(),
			Status:         domainbooking.StatusCancelled,
			ConferenceLink: "",
			CreatedAt:      createdAt,
		},
	}

	repo := &mockListBookingRepository{
		listAllFn: func(ctx context.Context) ([]domainbooking.Booking, error) {
			return bookings, nil
		},
	}

	uc := NewListUseCase(repo)

	out, err := uc.List(context.Background(), ListInput{
		Page:     "2",
		PageSize: "2",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if out == nil {
		t.Fatal("expected non-nil output")
	}

	if len(out.Bookings) != 1 {
		t.Fatalf("expected 1 booking on second page, got %d", len(out.Bookings))
	}

	if out.Bookings[0].ID != "booking-3" {
		t.Fatalf("expected booking-3 on second page, got %s", out.Bookings[0].ID)
	}

	if out.Pagination.Page != 2 {
		t.Fatalf("expected page 2, got %d", out.Pagination.Page)
	}

	if out.Pagination.PageSize != 2 {
		t.Fatalf("expected pageSize 2, got %d", out.Pagination.PageSize)
	}

	if out.Pagination.Total != 3 {
		t.Fatalf("expected total 3, got %d", out.Pagination.Total)
	}
}

func TestListBookings_PageOutOfRange_ReturnsEmptyBookings(t *testing.T) {
	createdAt := time.Now().UTC()

	repo := &mockListBookingRepository{
		listAllFn: func(ctx context.Context) ([]domainbooking.Booking, error) {
			return []domainbooking.Booking{
				{
					ID:             uuid.NewString(),
					SlotID:         uuid.NewString(),
					UserID:         uuid.NewString(),
					Status:         domainbooking.StatusActive,
					ConferenceLink: "",
					CreatedAt:      createdAt,
				},
			}, nil
		},
	}

	uc := NewListUseCase(repo)

	out, err := uc.List(context.Background(), ListInput{
		Page:     "10",
		PageSize: "20",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if out == nil {
		t.Fatal("expected non-nil output")
	}

	if len(out.Bookings) != 0 {
		t.Fatalf("expected 0 bookings, got %d", len(out.Bookings))
	}

	if out.Pagination.Total != 1 {
		t.Fatalf("expected total 1, got %d", out.Pagination.Total)
	}
}