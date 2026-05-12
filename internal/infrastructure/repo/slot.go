package repo

import (
	"context"
	"errors"
	"time"

	trmpgx "github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	domainslot "github.com/avito-internships/test-backend-1-katyaswedq/internal/domain/slot"
)

type SlotRepository struct {
	pool   *pgxpool.Pool
	getter *trmpgx.CtxGetter
}

func NewSlotRepository(pool *pgxpool.Pool, getter *trmpgx.CtxGetter) *SlotRepository {
	return &SlotRepository{
		pool:   pool,
		getter: getter,
	}
}

func (r *SlotRepository) BulkCreate(ctx context.Context, slots []domainslot.Slot) error {
	if len(slots) == 0 {
		return nil
	}

	batch := &pgx.Batch{}

	for _, slot := range slots {
		batch.Queue(
			`INSERT INTO slots (id, room_id, start_at, end_at)
			 VALUES ($1, $2, $3, $4)`, slot.ID, slot.RoomID, slot.Start, slot.End)
	}

	results := r.getter.DefaultTrOrDB(ctx, r.pool).SendBatch(ctx, batch)
	defer results.Close()

	for range slots {
		if _, err := results.Exec(); err != nil {
			return err
		}
	}

	return nil
}

func (r *SlotRepository) ListAvailableByRoomAndDate(ctx context.Context, roomID string, date time.Time) ([]domainslot.Slot, error) {
	startOfDay := time.Date(
		date.UTC().Year(),
		date.UTC().Month(),
		date.UTC().Day(),
		0, 0, 0, 0,
		time.UTC,
	)
	endOfDay := startOfDay.Add(24 * time.Hour)

	rows, err := r.getter.DefaultTrOrDB(ctx, r.pool).Query(
		ctx,
		`SELECT s.id, s.room_id, s.start_at, s.end_at
		 FROM slots s
		 LEFT JOIN bookings b ON b.slot_id = s.id AND b.status = 'active'
		 WHERE s.room_id = $1 AND s.start_at >= $2 AND s.start_at < $3 AND b.id IS NULL
		 ORDER BY s.start_at ASC`, roomID, startOfDay, endOfDay)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	slots := make([]domainslot.Slot, 0)
	for rows.Next() {
		var slot domainslot.Slot

		if err := rows.Scan(&slot.ID, &slot.RoomID, &slot.Start, &slot.End); err != nil {
			return nil, err
		}

		slots = append(slots, slot)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return slots, nil
}

func (r *SlotRepository) ExistsByID(ctx context.Context, id string) (bool, error) {
	var exists int

	err := r.getter.DefaultTrOrDB(ctx, r.pool).QueryRow(
		ctx,
		`SELECT 1
		 FROM slots
		 WHERE id = $1`, id).Scan(&exists)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}

		return false, err
	}

	return true, nil
}

func (r *SlotRepository) GetByID(ctx context.Context, id string) (*domainslot.Slot, error) {
	var slot domainslot.Slot

	err := r.getter.DefaultTrOrDB(ctx, r.pool).QueryRow(
		ctx,
		`SELECT id, room_id, start_at, end_at
		 FROM slots
		 WHERE id = $1`,
		id).Scan(&slot.ID, &slot.RoomID, &slot.Start, &slot.End)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &slot, nil
}