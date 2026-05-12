package repo

import (
	"context"
	"errors"

	trmpgx "github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	domainschedule "github.com/avito-internships/test-backend-1-katyaswedq/internal/domain/schedule"
)

type ScheduleRepository struct {
	pool   *pgxpool.Pool
	getter *trmpgx.CtxGetter
}

func NewScheduleRepository(pool *pgxpool.Pool, getter *trmpgx.CtxGetter) *ScheduleRepository {
	return &ScheduleRepository{
		pool:   pool,
		getter: getter,
	}
}

func (r *ScheduleRepository) Create(ctx context.Context, s *domainschedule.Schedule) error {
	_, err := r.getter.DefaultTrOrDB(ctx, r.pool).Exec(
		ctx,
		`INSERT INTO schedules (id, room_id, days_of_week, start_time, end_time)
		 VALUES ($1, $2, $3, $4, $5)`,
		s.ID,
		s.RoomID,
		s.DaysOfWeek,
		s.StartTime,
		s.EndTime,
	)

	return err
}

func (r *ScheduleRepository) ExistsByRoomID(ctx context.Context, roomID string) (bool, error) {
	var exists int

	err := r.getter.DefaultTrOrDB(ctx, r.pool).QueryRow(
		ctx,
		`SELECT 1
		 FROM schedules
		 WHERE room_id = $1`, roomID).Scan(&exists)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}

		return false, err
	}

	return true, nil
}