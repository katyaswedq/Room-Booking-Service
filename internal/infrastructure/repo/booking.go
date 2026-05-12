package repo

import (
	"context"
	"errors"

	trmpgx "github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	domainbooking "github.com/avito-internships/test-backend-1-katyaswedq/internal/domain/booking"
)

type BookingRepository struct {
	pool   *pgxpool.Pool
	getter *trmpgx.CtxGetter
}

func NewBookingRepository(pool *pgxpool.Pool, getter *trmpgx.CtxGetter) *BookingRepository {
	return &BookingRepository{
		pool:   pool,
		getter: getter,
	}
}

func (r *BookingRepository) Create(ctx context.Context, b *domainbooking.Booking) error {
	_, err := r.getter.DefaultTrOrDB(ctx, r.pool).Exec(
		ctx,
		`INSERT INTO bookings (id, slot_id, user_id, status, conference_link)
		 VALUES ($1, $2, $3, $4, $5)`, 
		b.ID,
		b.SlotID,
		b.UserID,
		b.Status,
		b.ConferenceLink)
	return err
}

func (r *BookingRepository) GetByID(ctx context.Context, id string) (*domainbooking.Booking, error) {
	var b domainbooking.Booking

	err := r.getter.DefaultTrOrDB(ctx, r.pool).QueryRow(
		ctx,
		`SELECT id, slot_id, user_id, status, conference_link, created_at
		 FROM bookings
		 WHERE id = $1`,
		id).Scan(&b.ID, &b.SlotID, &b.UserID, &b.Status, &b.ConferenceLink, &b.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &b, nil
}

func (r *BookingRepository) ListByUser(ctx context.Context, userID string) ([]domainbooking.Booking, error) {
	rows, err := r.getter.DefaultTrOrDB(ctx, r.pool).Query(
		ctx,
		`SELECT id, slot_id, user_id, status, conference_link, created_at
		 FROM bookings
		 WHERE user_id = $1
		 ORDER BY created_at DESC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	bookings := make([]domainbooking.Booking, 0)
	for rows.Next() {
		var b domainbooking.Booking

		if err := rows.Scan(&b.ID, &b.SlotID, &b.UserID, &b.Status, &b.ConferenceLink, &b.CreatedAt); err != nil {
			return nil, err
		}

		bookings = append(bookings, b)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return bookings, nil
}

func (r *BookingRepository) ListAll(ctx context.Context) ([]domainbooking.Booking, error) {
	rows, err := r.getter.DefaultTrOrDB(ctx, r.pool).Query(
		ctx,
		`SELECT id, slot_id, user_id, status, conference_link, created_at
		 FROM bookings
		 ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	bookings := make([]domainbooking.Booking, 0)
	for rows.Next() {
		var b domainbooking.Booking

		if err := rows.Scan(&b.ID, &b.SlotID, &b.UserID, &b.Status, &b.ConferenceLink, &b.CreatedAt); err != nil {
			return nil, err
		}

		bookings = append(bookings, b)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return bookings, nil
}

func (r *BookingRepository) UpdateStatus(ctx context.Context, id string, status string) error {
	cmd, err := r.getter.DefaultTrOrDB(ctx, r.pool).Exec(
		ctx,
		`UPDATE bookings
		 SET status = $1
		 WHERE id = $2`,
		status,
		id,
	)
	if err != nil {
		return err
	}

	if cmd.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}