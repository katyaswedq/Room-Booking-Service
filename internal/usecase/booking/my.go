package booking

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"

	domainbooking "github.com/avito-internships/test-backend-1-katyaswedq/internal/domain/booking"
	domainslot "github.com/avito-internships/test-backend-1-katyaswedq/internal/domain/slot"
)

var ErrInvalidUserID = errors.New("invalid user id")

type MyUseCase struct {
	bookingRepo domainbooking.Repository
	slotRepo    domainslot.Repository
}

type MyInput struct {
	UserID string
}

type MyOutput struct {
	Bookings []MyBookingOutput
}

type MyBookingOutput struct {
	ID             string
	SlotID         string
	UserID         string
	Status         string
	ConferenceLink string
	CreatedAt      time.Time
}

func NewMyUseCase(bookingRepo domainbooking.Repository, slotRepo domainslot.Repository) *MyUseCase {
	return &MyUseCase{
		bookingRepo: bookingRepo,
		slotRepo:    slotRepo,
	}
}

func (uc *MyUseCase) List(ctx context.Context, input MyInput) (*MyOutput, error) {
	userID := strings.TrimSpace(input.UserID)
	if _, err := uuid.Parse(userID); err != nil {
		return nil, ErrInvalidUserID
	}

	bookings, err := uc.bookingRepo.ListByUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	out := make([]MyBookingOutput, 0, len(bookings))

	for _, booking := range bookings {
		slot, err := uc.slotRepo.GetByID(ctx, booking.SlotID)
		if err != nil {
			return nil, err
		}
		if slot == nil {
			continue
		}

		if slot.Start.Before(now) {
			continue
		}

		out = append(out, MyBookingOutput{
			ID:             booking.ID,
			SlotID:         booking.SlotID,
			UserID:         booking.UserID,
			Status:         booking.Status,
			ConferenceLink: booking.ConferenceLink,
			CreatedAt:      booking.CreatedAt,
		})
	}

	return &MyOutput{
		Bookings: out,
	}, nil
}