package user

import "context"

type Repository interface {
	Create(ctx context.Context, user User) (User, error)
	GetByEmail(ctx context.Context, email string) (User, error)
	ExistsByEmail(ctx context.Context, email string) (bool, error)
}