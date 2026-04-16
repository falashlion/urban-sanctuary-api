package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/falashlion/urban-sanctuary-api/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PaymentRepository handles database operations for payments.
type PaymentRepository struct {
	pool *pgxpool.Pool
}

// NewPaymentRepository creates a new PaymentRepository.
func NewPaymentRepository(pool *pgxpool.Pool) *PaymentRepository {
	return &PaymentRepository{pool: pool}
}

// Create inserts a new payment into the database.
func (r *PaymentRepository) Create(ctx context.Context, p *domain.Payment) error {
	query := `
		INSERT INTO payments (id, booking_id, user_id, provider, provider_reference,
			phone_number, amount, currency, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING created_at, updated_at`

	p.ID = uuid.New()
	return r.pool.QueryRow(ctx, query,
		p.ID, p.BookingID, p.UserID, p.Provider, p.ProviderReference,
		p.PhoneNumber, p.Amount, p.Currency, p.Status,
	).Scan(&p.CreatedAt, &p.UpdatedAt)
}

// GetByID retrieves a payment by its ID.
func (r *PaymentRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Payment, error) {
	query := `
		SELECT id, booking_id, user_id, provider, provider_reference,
			phone_number, amount, currency, status, webhook_payload,
			created_at, updated_at
		FROM payments WHERE id = $1`

	p := &domain.Payment{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&p.ID, &p.BookingID, &p.UserID, &p.Provider, &p.ProviderReference,
		&p.PhoneNumber, &p.Amount, &p.Currency, &p.Status, &p.WebhookPayload,
		&p.CreatedAt, &p.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get payment: %w", err)
	}
	return p, nil
}

// GetByBookingID retrieves payments for a booking.
func (r *PaymentRepository) GetByBookingID(ctx context.Context, bookingID uuid.UUID) ([]domain.Payment, error) {
	query := `
		SELECT id, booking_id, user_id, provider, provider_reference,
			phone_number, amount, currency, status, webhook_payload,
			created_at, updated_at
		FROM payments WHERE booking_id = $1 ORDER BY created_at DESC`

	rows, err := r.pool.Query(ctx, query, bookingID)
	if err != nil {
		return nil, fmt.Errorf("failed to get payments by booking: %w", err)
	}
	defer rows.Close()

	var payments []domain.Payment
	for rows.Next() {
		var p domain.Payment
		if err := rows.Scan(
			&p.ID, &p.BookingID, &p.UserID, &p.Provider, &p.ProviderReference,
			&p.PhoneNumber, &p.Amount, &p.Currency, &p.Status, &p.WebhookPayload,
			&p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, err
		}
		payments = append(payments, p)
	}
	return payments, nil
}

// GetByProviderReference retrieves a payment by its external provider reference.
func (r *PaymentRepository) GetByProviderReference(ctx context.Context, ref string) (*domain.Payment, error) {
	query := `
		SELECT id, booking_id, user_id, provider, provider_reference,
			phone_number, amount, currency, status, webhook_payload,
			created_at, updated_at
		FROM payments WHERE provider_reference = $1`

	p := &domain.Payment{}
	err := r.pool.QueryRow(ctx, query, ref).Scan(
		&p.ID, &p.BookingID, &p.UserID, &p.Provider, &p.ProviderReference,
		&p.PhoneNumber, &p.Amount, &p.Currency, &p.Status, &p.WebhookPayload,
		&p.CreatedAt, &p.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get payment by provider ref: %w", err)
	}
	return p, nil
}

// UpdateStatus updates a payment's status.
func (r *PaymentRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.PaymentStatus) error {
	query := `UPDATE payments SET status = $2, updated_at = NOW() WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id, status)
	return err
}

// UpdateWebhookPayload stores the raw webhook payload from the payment provider.
func (r *PaymentRepository) UpdateWebhookPayload(ctx context.Context, id uuid.UUID, payload json.RawMessage) error {
	query := `UPDATE payments SET webhook_payload = $2, updated_at = NOW() WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id, payload)
	return err
}

// UpdateProviderReference sets the external transaction reference.
func (r *PaymentRepository) UpdateProviderReference(ctx context.Context, id uuid.UUID, ref string) error {
	query := `UPDATE payments SET provider_reference = $2, updated_at = NOW() WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id, ref)
	return err
}
