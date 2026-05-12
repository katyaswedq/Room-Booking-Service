package schedule

import "context"

type Repository interface {
	Create(ctx context.Context, s *Schedule) error
	ExistsByRoomID(ctx context.Context, roomID string) (bool, error)
}