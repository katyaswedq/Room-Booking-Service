package booking

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"

	domainbooking "github.com/avito-internships/test-backend-1-katyaswedq/internal/domain/booking"
)

var (
	ErrInvalidBookingID = errors.New("invalid booking id")
	ErrBookingNotFound  = errors.New("booking not found")
	ErrForbidden        = errors.New("forbidden")
)

type CancelUseCase struct {
	bookingRepo domainbooking.Repository
}

type CancelInput struct {
	BookingID string
	UserID    string
	Role      string
}

type CancelOutput struct {
	ID             string
	SlotID         string
	UserID         string
	Status         string
	ConferenceLink string
	CreatedAt      time.Time
}

func NewCancelUseCase(repo domainbooking.Repository) *CancelUseCase {
	return &CancelUseCase{
		bookingRepo: repo,
	}
}

func (uc *CancelUseCase) Cancel(ctx context.Context, input CancelInput) (*CancelOutput, error) {
	bookingID := strings.TrimSpace(input.BookingID)

	if _, err := uuid.Parse(bookingID); err != nil {
		return nil, ErrInvalidBookingID
	}

	b, err := uc.bookingRepo.GetByID(ctx, bookingID)
	if err != nil {
		return nil, err
	}
	if b == nil {
		return nil, ErrBookingNotFound
	}

	if b.UserID != input.UserID {
		return nil, ErrForbidden
	}

	if b.Status != domainbooking.StatusCancelled {
		if err := uc.bookingRepo.UpdateStatus(ctx, bookingID, domainbooking.StatusCancelled); err != nil {
			return nil, err
		}
		b.Status = domainbooking.StatusCancelled
	}

	return &CancelOutput{
		ID:             b.ID,
		SlotID:         b.SlotID,
		UserID:         b.UserID,
		Status:         b.Status,
		ConferenceLink: b.ConferenceLink,
		CreatedAt:      b.CreatedAt,
	}, nil
}