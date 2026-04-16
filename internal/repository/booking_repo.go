package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/falashlion/urban-sanctuary-api/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// BookingRepository handles database operations for bookings.
type BookingRepository struct {
	pool *pgxpool.Pool
}

// NewBookingRepository creates a new BookingRepository.
func NewBookingRepository(pool *pgxpool.Pool) *BookingRepository {
	return &BookingRepository{pool: pool}
}

// Create inserts a new booking into the database.
func (r *BookingRepository) Create(ctx context.Context, b *domain.Booking) error {
	query := `
		INSERT INTO bookings (id, property_id, guest_id, check_in, check_out, guests_count,
			base_amount, discount_amount, total_amount, status, loyalty_points_earned, loyalty_points_used)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING nights, created_at, updated_at`

	b.ID = uuid.New()
	return r.pool.QueryRow(ctx, query,
		b.ID, b.PropertyID, b.GuestID, b.CheckIn, b.CheckOut, b.GuestsCount,
		b.BaseAmount, b.DiscountAmount, b.TotalAmount, b.Status,
		b.LoyaltyPointsEarned, b.LoyaltyPointsUsed,
	).Scan(&b.Nights, &b.CreatedAt, &b.UpdatedAt)
}

// GetByID retrieves a booking by its ID.
func (r *BookingRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Booking, error) {
	query := `
		SELECT id, property_id, guest_id, check_in, check_out, nights, guests_count,
			base_amount, discount_amount, total_amount, status,
			loyalty_points_earned, loyalty_points_used, created_at, updated_at
		FROM bookings WHERE id = $1`

	b := &domain.Booking{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&b.ID, &b.PropertyID, &b.GuestID, &b.CheckIn, &b.CheckOut, &b.Nights, &b.GuestsCount,
		&b.BaseAmount, &b.DiscountAmount, &b.TotalAmount, &b.Status,
		&b.LoyaltyPointsEarned, &b.LoyaltyPointsUsed, &b.CreatedAt, &b.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get booking: %w", err)
	}
	return b, nil
}

// UpdateStatus updates a booking's status.
func (r *BookingRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.BookingStatus) error {
	query := `UPDATE bookings SET status = $2, updated_at = NOW() WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id, status)
	return err
}

// CheckConflict checks if there are overlapping bookings for a property.
func (r *BookingRepository) CheckConflict(ctx context.Context, propertyID uuid.UUID, checkIn, checkOut time.Time) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM bookings
			WHERE property_id = $1
			AND status IN ('pending', 'confirmed')
			AND check_in < $3 AND check_out > $2
		)`

	var exists bool
	err := r.pool.QueryRow(ctx, query, propertyID, checkIn, checkOut).Scan(&exists)
	return exists, err
}

// ListByGuest retrieves bookings for a specific guest with pagination.
func (r *BookingRepository) ListByGuest(ctx context.Context, guestID uuid.UUID, status string, page, perPage int) ([]domain.Booking, int64, error) {
	countQuery := `SELECT COUNT(*) FROM bookings WHERE guest_id = $1`
	dataQuery := `
		SELECT id, property_id, guest_id, check_in, check_out, nights, guests_count,
			base_amount, discount_amount, total_amount, status,
			loyalty_points_earned, loyalty_points_used, created_at, updated_at
		FROM bookings WHERE guest_id = $1`

	args := []interface{}{guestID}
	argIdx := 2

	if status != "" {
		filter := fmt.Sprintf(" AND status = $%d", argIdx)
		countQuery += filter
		dataQuery += filter
		args = append(args, status)
		argIdx++
	}

	var total int64
	err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	dataQuery += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, perPage, (page-1)*perPage)

	rows, err := r.pool.Query(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var bookings []domain.Booking
	for rows.Next() {
		var b domain.Booking
		if err := rows.Scan(
			&b.ID, &b.PropertyID, &b.GuestID, &b.CheckIn, &b.CheckOut, &b.Nights, &b.GuestsCount,
			&b.BaseAmount, &b.DiscountAmount, &b.TotalAmount, &b.Status,
			&b.LoyaltyPointsEarned, &b.LoyaltyPointsUsed, &b.CreatedAt, &b.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		bookings = append(bookings, b)
	}
	return bookings, total, nil
}

// ListAll retrieves all bookings with optional status filter and pagination (admin).
func (r *BookingRepository) ListAll(ctx context.Context, status string, page, perPage int) ([]domain.Booking, int64, error) {
	countQuery := `SELECT COUNT(*) FROM bookings WHERE 1=1`
	dataQuery := `
		SELECT id, property_id, guest_id, check_in, check_out, nights, guests_count,
			base_amount, discount_amount, total_amount, status,
			loyalty_points_earned, loyalty_points_used, created_at, updated_at
		FROM bookings WHERE 1=1`

	args := []interface{}{}
	argIdx := 1

	if status != "" {
		filter := fmt.Sprintf(" AND status = $%d", argIdx)
		countQuery += filter
		dataQuery += filter
		args = append(args, status)
		argIdx++
	}

	var total int64
	err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	dataQuery += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, perPage, (page-1)*perPage)

	rows, err := r.pool.Query(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var bookings []domain.Booking
	for rows.Next() {
		var b domain.Booking
		if err := rows.Scan(
			&b.ID, &b.PropertyID, &b.GuestID, &b.CheckIn, &b.CheckOut, &b.Nights, &b.GuestsCount,
			&b.BaseAmount, &b.DiscountAmount, &b.TotalAmount, &b.Status,
			&b.LoyaltyPointsEarned, &b.LoyaltyPointsUsed, &b.CreatedAt, &b.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		bookings = append(bookings, b)
	}
	return bookings, total, nil
}

// UpdateLoyaltyPoints sets the loyalty points earned for a booking.
func (r *BookingRepository) UpdateLoyaltyPoints(ctx context.Context, id uuid.UUID, earned int) error {
	query := `UPDATE bookings SET loyalty_points_earned = $2, updated_at = NOW() WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id, earned)
	return err
}
