package repo

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/avito-internships/test-backend-1-katyaswedq/internal/domain/room"
)

type RoomRepository struct {
	pool *pgxpool.Pool
}

func NewRoomRepository(pool *pgxpool.Pool) *RoomRepository {
	return &RoomRepository{
		pool: pool,
	}
}

func (r *RoomRepository) Create(ctx context.Context, rm *room.Room) error {
	query := `insert into rooms (id, name, description, capacity, created_at)
			  values ($1, $2, $3, $4, $5)`

	_, err := r.pool.Exec(ctx, query, rm.ID, rm.Name, rm.Description, rm.Capacity, rm.CreatedAt)
	if err != nil {
		return err
	}

	return nil
}

func (r *RoomRepository) List(ctx context.Context) ([]room.Room, error) {
	query := `select id, name, description, capacity, created_at
			  from rooms
		      order by created_at desc`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	rooms := make([]room.Room, 0)

	for rows.Next() {
		var rm room.Room
		if err := rows.Scan(&rm.ID, &rm.Name, &rm.Description, &rm.Capacity, &rm.CreatedAt); err != nil {
			return nil, err
		}

		rooms = append(rooms, rm)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return rooms, nil
}

func (r *RoomRepository) ExistsByID(ctx context.Context, id string) (bool, error) {
	query := `SELECT 1 FROM rooms WHERE id = $1`

	var exists int
	err := r.pool.QueryRow(ctx, query, id).Scan(&exists)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}