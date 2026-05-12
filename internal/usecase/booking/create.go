package booking

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"

	domainbooking "github.com/avito-internships/test-backend-1-katyaswedq/internal/domain/booking"
	domainslot "github.com/avito-internships/test-backend-1-katyaswedq/internal/domain/slot"
)

var (
	ErrInvalidSlotID      = errors.New("invalid slot id")
	ErrSlotNotFound       = errors.New("slot not found")
	ErrSlotAlreadyBooked  = errors.New("slot already booked")
	ErrSlotAlreadyStarted = errors.New("slot already started")
	ErrConferenceLinkError = errors.New("failed to create conference link")
)

type CreateUseCase struct {
	bookingRepo       domainbooking.Repository
	slotRepo          domainslot.Repository
	conferenceService domainbooking.ConferenceService
}

type CreateInput struct {
	SlotID               string
	UserID               string
	CreateConferenceLink bool
}

type CreateOutput struct {
	ID             string
	SlotID         string
	UserID         string
	Status         string
	ConferenceLink string
	CreatedAt      time.Time
}

func NewCreateUseCase(
	bookingRepo domainbooking.Repository,
	slotRepo domainslot.Repository,
	conferenceService domainbooking.ConferenceService,
) *CreateUseCase {
	return &CreateUseCase{
		bookingRepo:       bookingRepo,
		slotRepo:          slotRepo,
		conferenceService: conferenceService,
	}
}

func (uc *CreateUseCase) Create(ctx context.Context, input CreateInput) (*CreateOutput, error) {
	slotID := strings.TrimSpace(input.SlotID)
	userID := strings.TrimSpace(input.UserID)

	if _, err := uuid.Parse(slotID); err != nil {
		return nil, ErrInvalidSlotID
	}

	if _, err := uuid.Parse(userID); err != nil {
		return nil, ErrInvalidUserID
	}

	slot, err := uc.slotRepo.GetByID(ctx, slotID)
	if err != nil {
		return nil, err
	}
	if slot == nil {
		return nil, ErrSlotNotFound
	}

	if !time.Now().UTC().Before(slot.Start) {
		return nil, ErrSlotAlreadyStarted
	}

	conferenceLink := ""
	if input.CreateConferenceLink {
		if uc.conferenceService == nil {
			return nil, ErrConferenceLinkError
		}

		conferenceLink, err = uc.conferenceService.CreateLink(ctx, *slot)
		if err != nil {
			return nil, ErrConferenceLinkError
		}
	}

	booking := &domainbooking.Booking{
		ID:             uuid.NewString(),
		SlotID:         slotID,
		UserID:         userID,
		Status:         domainbooking.StatusActive,
		ConferenceLink: conferenceLink,
		CreatedAt:      time.Now().UTC(),
	}

	if err := uc.bookingRepo.Create(ctx, booking); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrSlotAlreadyBooked
		}
		return nil, err
	}

	return &CreateOutput{
		ID:             booking.ID,
		SlotID:         booking.SlotID,
		UserID:         booking.UserID,
		Status:         booking.Status,
		ConferenceLink: booking.ConferenceLink,
		CreatedAt:      booking.CreatedAt,
	}, nil
}