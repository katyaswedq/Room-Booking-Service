package slot

import (
	"context"
	"time"
)

type Repository interface {
	BulkCreate(ctx context.Context, slots []Slot) error
	ListAvailableByRoomAndDate(ctx context.Context, roomID string, date time.Time) ([]Slot, error)
	ExistsByID(ctx context.Context, id string) (bool, error)
	GetByID(ctx context.Context, id string) (*Slot, error)
}