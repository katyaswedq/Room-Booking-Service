package booking

import (
	"context"
	"errors"
	"strconv"
	"time"

	domainbooking "github.com/avito-internships/test-backend-1-katyaswedq/internal/domain/booking"
)

const (
	defaultPage     = 1
	defaultPageSize = 20
	maxPageSize     = 100
)

var ErrInvalidPagination = errors.New("invalid pagination")

type ListUseCase struct {
	bookingRepo domainbooking.Repository
}

type ListInput struct {
	Page     string
	PageSize string
}

type ListOutput struct {
	Bookings   []ListBookingOutput
	Pagination PaginationOutput
}

type ListBookingOutput struct {
	ID             string
	SlotID         string
	UserID         string
	Status         string
	ConferenceLink string
	CreatedAt      time.Time
}

type PaginationOutput struct {
	Page     int
	PageSize int
	Total    int
}

func NewListUseCase(bookingRepo domainbooking.Repository) *ListUseCase {
	return &ListUseCase{
		bookingRepo: bookingRepo,
	}
}

func (uc *ListUseCase) List(ctx context.Context, input ListInput) (*ListOutput, error) {
	page := defaultPage
	pageSize := defaultPageSize

	if input.Page != "" {
		parsedPage, err := strconv.Atoi(input.Page)
		if err != nil || parsedPage <= 0 {
			return nil, ErrInvalidPagination
		}
		page = parsedPage
	}

	if input.PageSize != "" {
		parsedPageSize, err := strconv.Atoi(input.PageSize)
		if err != nil || parsedPageSize <= 0 {
			return nil, ErrInvalidPagination
		}
		if parsedPageSize > maxPageSize {
			parsedPageSize = maxPageSize
		}
		pageSize = parsedPageSize
	}

	bookings, err := uc.bookingRepo.ListAll(ctx)
	if err != nil {
		return nil, err
	}

	total := len(bookings)

	start := (page - 1) * pageSize
	if start > total {
		start = total
	}

	end := start + pageSize
	if end > total {
		end = total
	}

	pageBookings := bookings[start:end]

	out := make([]ListBookingOutput, 0, len(pageBookings))
	for _, booking := range pageBookings {
		out = append(out, ListBookingOutput{
			ID:             booking.ID,
			SlotID:         booking.SlotID,
			UserID:         booking.UserID,
			Status:         booking.Status,
			ConferenceLink: booking.ConferenceLink,
			CreatedAt:      booking.CreatedAt,
		})
	}

	return &ListOutput{
		Bookings: out,
		Pagination: PaginationOutput{
			Page:     page,
			PageSize: pageSize,
			Total:    total,
		},
	}, nil
}