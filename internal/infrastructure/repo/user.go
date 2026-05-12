package repo

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	domainuser "github.com/avito-internships/test-backend-1-katyaswedq/internal/domain/user"
)

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

func (r *UserRepository) Create(ctx context.Context, user domainuser.User) (domainuser.User, error) {
	query := `
		INSERT INTO users (id, email, password_hash, role)
		VALUES ($1, $2, $3, $4)
		RETURNING created_at
	`

	err := r.pool.QueryRow(ctx, query, user.ID, user.Email, user.PasswordHash, user.Role,).Scan(&user.CreatedAt)

	if err != nil {
		return domainuser.User{}, err
	}

	return user, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (domainuser.User, error) {
	query := `SELECT id, email, password_hash, role, created_at
		      FROM users
		      WHERE email = $1`

	var u domainuser.User

	err := r.pool.QueryRow(ctx, query, email).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Role, &u.CreatedAt)

	if err != nil {
		return domainuser.User{}, err
	}

	return u, nil
}

func (r *UserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	query := `SELECT EXISTS (SELECT 1 FROM users WHERE email = $1)`

	var exists bool

	err := r.pool.QueryRow(ctx, query, email).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}