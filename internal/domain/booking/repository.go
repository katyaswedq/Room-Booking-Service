package booking

import "context"

type Repository interface {
	Create(ctx context.Context, b *Booking) error
	GetByID(ctx context.Context, id string) (*Booking, error)
	ListByUser(ctx context.Context, userID string) ([]Booking, error)
	ListAll(ctx context.Context) ([]Booking, error)
	UpdateStatus(ctx context.Context, id string, status string) error
}