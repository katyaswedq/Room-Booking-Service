package booking

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"

	domainbooking "github.com/avito-internships/test-backend-1-katyaswedq/internal/domain/booking"
	domainslot "github.com/avito-internships/test-backend-1-katyaswedq/internal/domain/slot"
)

type mockBookingRepository struct {
	createFn func(ctx context.Context, booking *domainbooking.Booking) error
}

func (m *mockBookingRepository) Create(ctx context.Context, booking *domainbooking.Booking) error {
	if m.createFn != nil {
		return m.createFn(ctx, booking)
	}
	return nil
}

func (m *mockBookingRepository) GetByID(ctx context.Context, bookingID string) (*domainbooking.Booking, error) {
	return nil, nil
}

func (m *mockBookingRepository) ListByUser(ctx context.Context, userID string) ([]domainbooking.Booking, error) {
	return nil, nil
}

func (m *mockBookingRepository) ListAll(ctx context.Context) ([]domainbooking.Booking, error) {
	return nil, nil
}

func (m *mockBookingRepository) UpdateStatus(ctx context.Context, bookingID string, status string) error {
	return nil
}

type mockSlotRepository struct {
	getByIDFn func(ctx context.Context, slotID string) (*domainslot.Slot, error)
}

func (m *mockSlotRepository) BulkCreate(ctx context.Context, slots []domainslot.Slot) error {
	return nil
}

func (m *mockSlotRepository) ListAvailableByRoomAndDate(ctx context.Context, roomID string, date time.Time) ([]domainslot.Slot, error) {
	return nil, nil
}

func (m *mockSlotRepository) ExistsByID(ctx context.Context, slotID string) (bool, error) {
	return false, nil
}

func (m *mockSlotRepository) GetByID(ctx context.Context, slotID string) (*domainslot.Slot, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, slotID)
	}
	return nil, nil
}

type mockConferenceService struct {
	createLinkFn func(ctx context.Context, slot domainslot.Slot) (string, error)
}

func (m *mockConferenceService) CreateLink(ctx context.Context, slot domainslot.Slot) (string, error) {
	if m.createLinkFn != nil {
		return m.createLinkFn(ctx, slot)
	}
	return "", nil
}

func TestCreateBooking_InvalidSlotID_ReturnsError(t *testing.T) {
	uc := NewCreateUseCase(
		&mockBookingRepository{},
		&mockSlotRepository{},
		nil,
	)

	out, err := uc.Create(context.Background(), CreateInput{
		SlotID: "bad-uuid",
		UserID: uuid.NewString(),
	})

	if !errors.Is(err, ErrInvalidSlotID) {
		t.Fatalf("expected ErrInvalidSlotID, got %v", err)
	}
	if out != nil {
		t.Fatalf("expected nil output, got %+v", out)
	}
}

func TestCreateBooking_InvalidUserID_ReturnsError(t *testing.T) {
	uc := NewCreateUseCase(
		&mockBookingRepository{},
		&mockSlotRepository{},
		nil,
	)

	out, err := uc.Create(context.Background(), CreateInput{
		SlotID: uuid.NewString(),
		UserID: "bad-uuid",
	})

	if !errors.Is(err, ErrInvalidUserID) {
		t.Fatalf("expected ErrInvalidUserID, got %v", err)
	}
	if out != nil {
		t.Fatalf("expected nil output, got %+v", out)
	}
}

func TestCreateBooking_SlotNotFound_ReturnsError(t *testing.T) {
	slotID := uuid.NewString()

	uc := NewCreateUseCase(
		&mockBookingRepository{},
		&mockSlotRepository{
			getByIDFn: func(ctx context.Context, id string) (*domainslot.Slot, error) {
				if id != slotID {
					t.Fatalf("unexpected slot id: %s", id)
				}
				return nil, nil
			},
		},
		nil,
	)

	out, err := uc.Create(context.Background(), CreateInput{
		SlotID: slotID,
		UserID: uuid.NewString(),
	})

	if !errors.Is(err, ErrSlotNotFound) {
		t.Fatalf("expected ErrSlotNotFound, got %v", err)
	}
	if out != nil {
		t.Fatalf("expected nil output, got %+v", out)
	}
}

func TestCreateBooking_SlotRepoError_ReturnsError(t *testing.T) {
	expectedErr := errors.New("slot repo error")

	uc := NewCreateUseCase(
		&mockBookingRepository{},
		&mockSlotRepository{
			getByIDFn: func(ctx context.Context, id string) (*domainslot.Slot, error) {
				return nil, expectedErr
			},
		},
		nil,
	)

	out, err := uc.Create(context.Background(), CreateInput{
		SlotID: uuid.NewString(),
		UserID: uuid.NewString(),
	})

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}
	if out != nil {
		t.Fatalf("expected nil output, got %+v", out)
	}
}

func TestCreateBooking_SlotAlreadyStarted_ReturnsError(t *testing.T) {
	slotID := uuid.NewString()
	userID := uuid.NewString()

	uc := NewCreateUseCase(
		&mockBookingRepository{},
		&mockSlotRepository{
			getByIDFn: func(ctx context.Context, id string) (*domainslot.Slot, error) {
				return &domainslot.Slot{
					ID:    slotID,
					Start: time.Now().UTC().Add(-time.Minute),
					End:   time.Now().UTC(),
				}, nil
			},
		},
		nil,
	)

	out, err := uc.Create(context.Background(), CreateInput{
		SlotID: slotID,
		UserID: userID,
	})

	if !errors.Is(err, ErrSlotAlreadyStarted) {
		t.Fatalf("expected ErrSlotAlreadyStarted, got %v", err)
	}
	if out != nil {
		t.Fatalf("expected nil output, got %+v", out)
	}
}

func TestCreateBooking_Success(t *testing.T) {
	slotID := uuid.NewString()
	userID := uuid.NewString()

	uc := NewCreateUseCase(
		&mockBookingRepository{
			createFn: func(ctx context.Context, booking *domainbooking.Booking) error {
				if booking.SlotID != slotID {
					t.Fatalf("expected slotID %s, got %s", slotID, booking.SlotID)
				}
				if booking.UserID != userID {
					t.Fatalf("expected userID %s, got %s", userID, booking.UserID)
				}
				if booking.Status != domainbooking.StatusActive {
					t.Fatalf("expected status %s, got %s", domainbooking.StatusActive, booking.Status)
				}
				if booking.ConferenceLink != "" {
					t.Fatalf("expected empty conference link, got %s", booking.ConferenceLink)
				}
				return nil
			},
		},
		&mockSlotRepository{
			getByIDFn: func(ctx context.Context, id string) (*domainslot.Slot, error) {
				return &domainslot.Slot{
					ID:    slotID,
					Start: time.Now().UTC().Add(time.Hour),
					End:   time.Now().UTC().Add(2 * time.Hour),
				}, nil
			},
		},
		nil,
	)

	out, err := uc.Create(context.Background(), CreateInput{
		SlotID: slotID,
		UserID: userID,
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if out == nil {
		t.Fatal("expected non-nil output")
	}
	if out.ID == "" {
		t.Fatal("expected generated booking id")
	}
	if out.SlotID != slotID {
		t.Fatalf("expected slotID %s, got %s", slotID, out.SlotID)
	}
	if out.UserID != userID {
		t.Fatalf("expected userID %s, got %s", userID, out.UserID)
	}
	if out.Status != domainbooking.StatusActive {
		t.Fatalf("expected status %s, got %s", domainbooking.StatusActive, out.Status)
	}
	if out.ConferenceLink != "" {
		t.Fatalf("expected empty conference link, got %s", out.ConferenceLink)
	}
}

func TestCreateBooking_SlotAlreadyBooked_ReturnsError(t *testing.T) {
	slotID := uuid.NewString()
	userID := uuid.NewString()

	uc := NewCreateUseCase(
		&mockBookingRepository{
			createFn: func(ctx context.Context, booking *domainbooking.Booking) error {
				return &pgconn.PgError{Code: "23505"}
			},
		},
		&mockSlotRepository{
			getByIDFn: func(ctx context.Context, id string) (*domainslot.Slot, error) {
				return &domainslot.Slot{
					ID:    slotID,
					Start: time.Now().UTC().Add(time.Hour),
					End:   time.Now().UTC().Add(2 * time.Hour),
				}, nil
			},
		},
		nil,
	)

	out, err := uc.Create(context.Background(), CreateInput{
		SlotID: slotID,
		UserID: userID,
	})

	if !errors.Is(err, ErrSlotAlreadyBooked) {
		t.Fatalf("expected ErrSlotAlreadyBooked, got %v", err)
	}
	if out != nil {
		t.Fatalf("expected nil output, got %+v", out)
	}
}

func TestCreateBooking_WithConferenceLink_Success(t *testing.T) {
	slotID := uuid.NewString()
	userID := uuid.NewString()
	expectedLink := "https://conference.local/rooms/r1/slots/" + slotID

	uc := NewCreateUseCase(
		&mockBookingRepository{
			createFn: func(ctx context.Context, booking *domainbooking.Booking) error {
				if booking.ConferenceLink != expectedLink {
					t.Fatalf("expected conference link %s, got %s", expectedLink, booking.ConferenceLink)
				}
				return nil
			},
		},
		&mockSlotRepository{
			getByIDFn: func(ctx context.Context, id string) (*domainslot.Slot, error) {
				return &domainslot.Slot{
					ID:     slotID,
					RoomID: "r1",
					Start:  time.Now().UTC().Add(time.Hour),
					End:    time.Now().UTC().Add(2 * time.Hour),
				}, nil
			},
		},
		&mockConferenceService{
			createLinkFn: func(ctx context.Context, slot domainslot.Slot) (string, error) {
				if slot.ID != slotID {
					t.Fatalf("expected slot id %s, got %s", slotID, slot.ID)
				}
				return expectedLink, nil
			},
		},
	)

	out, err := uc.Create(context.Background(), CreateInput{
		SlotID:               slotID,
		UserID:               userID,
		CreateConferenceLink: true,
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if out == nil {
		t.Fatal("expected non-nil output")
	}
	if out.ConferenceLink != expectedLink {
		t.Fatalf("expected conference link %s, got %s", expectedLink, out.ConferenceLink)
	}
}

func TestCreateBooking_WithConferenceLink_NoService_ReturnsError(t *testing.T) {
	slotID := uuid.NewString()
	userID := uuid.NewString()

	uc := NewCreateUseCase(
		&mockBookingRepository{},
		&mockSlotRepository{
			getByIDFn: func(ctx context.Context, id string) (*domainslot.Slot, error) {
				return &domainslot.Slot{
					ID:    slotID,
					Start: time.Now().UTC().Add(time.Hour),
					End:   time.Now().UTC().Add(2 * time.Hour),
				}, nil
			},
		},
		nil,
	)

	out, err := uc.Create(context.Background(), CreateInput{
		SlotID:               slotID,
		UserID:               userID,
		CreateConferenceLink: true,
	})

	if !errors.Is(err, ErrConferenceLinkError) {
		t.Fatalf("expected ErrConferenceLinkError, got %v", err)
	}
	if out != nil {
		t.Fatalf("expected nil output, got %+v", out)
	}
}

func TestCreateBooking_WithConferenceLink_ServiceError_ReturnsError(t *testing.T) {
	slotID := uuid.NewString()
	userID := uuid.NewString()
	expectedErr := errors.New("conference service error")

	uc := NewCreateUseCase(
		&mockBookingRepository{},
		&mockSlotRepository{
			getByIDFn: func(ctx context.Context, id string) (*domainslot.Slot, error) {
				return &domainslot.Slot{
					ID:    slotID,
					Start: time.Now().UTC().Add(time.Hour),
					End:   time.Now().UTC().Add(2 * time.Hour),
				}, nil
			},
		},
		&mockConferenceService{
			createLinkFn: func(ctx context.Context, slot domainslot.Slot) (string, error) {
				return "", expectedErr
			},
		},
	)

	out, err := uc.Create(context.Background(), CreateInput{
		SlotID:               slotID,
		UserID:               userID,
		CreateConferenceLink: true,
	})

	if !errors.Is(err, ErrConferenceLinkError) {
		t.Fatalf("expected ErrConferenceLinkError, got %v", err)
	}
	if out != nil {
		t.Fatalf("expected nil output, got %+v", out)
	}
}