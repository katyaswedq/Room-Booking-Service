package room

import "context"

type Repository interface {
	Create(ctx context.Context, room *Room) error
	List(ctx context.Context) ([]Room, error)
	ExistsByID(ctx context.Context, id string) (bool, error)
}