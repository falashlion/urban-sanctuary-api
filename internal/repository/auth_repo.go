package repository

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/falashlion/urban-sanctuary-api/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// AuthRepository handles database operations for auth-related tables.
type AuthRepository struct {
	pool *pgxpool.Pool
}

// NewAuthRepository creates a new AuthRepository.
func NewAuthRepository(pool *pgxpool.Pool) *AuthRepository {
	return &AuthRepository{pool: pool}
}

// --- Refresh Tokens ---

// CreateRefreshToken stores a hashed refresh token.
func (r *AuthRepository) CreateRefreshToken(ctx context.Context, token *domain.RefreshToken) error {
	query := `
		INSERT INTO refresh_tokens (id, user_id, token_hash, expires_at, device_info)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING created_at`

	token.ID = uuid.New()
	return r.pool.QueryRow(ctx, query,
		token.ID, token.UserID, token.TokenHash, token.ExpiresAt, token.DeviceInfo,
	).Scan(&token.CreatedAt)
}

// GetRefreshTokenByHash retrieves a refresh token by its hash.
func (r *AuthRepository) GetRefreshTokenByHash(ctx context.Context, tokenHash string) (*domain.RefreshToken, error) {
	query := `
		SELECT id, user_id, token_hash, expires_at, revoked, device_info, created_at
		FROM refresh_tokens WHERE token_hash = $1 AND revoked = false`

	t := &domain.RefreshToken{}
	err := r.pool.QueryRow(ctx, query, tokenHash).Scan(
		&t.ID, &t.UserID, &t.TokenHash, &t.ExpiresAt, &t.Revoked, &t.DeviceInfo, &t.CreatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get refresh token: %w", err)
	}
	return t, nil
}

// RevokeRefreshToken marks a refresh token as revoked.
func (r *AuthRepository) RevokeRefreshToken(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE refresh_tokens SET revoked = true WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	return err
}

// RevokeAllUserTokens revokes all refresh tokens for a user.
func (r *AuthRepository) RevokeAllUserTokens(ctx context.Context, userID uuid.UUID) error {
	query := `UPDATE refresh_tokens SET revoked = true WHERE user_id = $1 AND revoked = false`
	_, err := r.pool.Exec(ctx, query, userID)
	return err
}

// CleanupExpiredTokens removes expired refresh tokens.
func (r *AuthRepository) CleanupExpiredTokens(ctx context.Context) error {
	query := `DELETE FROM refresh_tokens WHERE expires_at < NOW() OR revoked = true`
	_, err := r.pool.Exec(ctx, query)
	return err
}

// HashToken creates a SHA-256 hash of a token string.
func HashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

// --- OTP Codes ---

// CreateOTPCode stores a new OTP code.
func (r *AuthRepository) CreateOTPCode(ctx context.Context, otp *domain.OTPCode) error {
	query := `
		INSERT INTO otp_codes (id, identifier, code_hash, purpose, expires_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING created_at`

	otp.ID = uuid.New()
	return r.pool.QueryRow(ctx, query,
		otp.ID, otp.Identifier, otp.CodeHash, otp.Purpose, otp.ExpiresAt,
	).Scan(&otp.CreatedAt)
}

// GetLatestOTP retrieves the latest unused OTP for an identifier and purpose.
func (r *AuthRepository) GetLatestOTP(ctx context.Context, identifier string, purpose domain.OTPPurpose) (*domain.OTPCode, error) {
	query := `
		SELECT id, identifier, code_hash, purpose, expires_at, attempts, used, created_at
		FROM otp_codes
		WHERE identifier = $1 AND purpose = $2 AND used = false AND expires_at > $3
		ORDER BY created_at DESC LIMIT 1`

	otp := &domain.OTPCode{}
	err := r.pool.QueryRow(ctx, query, identifier, purpose, time.Now()).Scan(
		&otp.ID, &otp.Identifier, &otp.CodeHash, &otp.Purpose,
		&otp.ExpiresAt, &otp.Attempts, &otp.Used, &otp.CreatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get OTP: %w", err)
	}
	return otp, nil
}

// IncrementOTPAttempts increments the attempt count for an OTP.
func (r *AuthRepository) IncrementOTPAttempts(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE otp_codes SET attempts = attempts + 1 WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	return err
}

// MarkOTPUsed marks an OTP code as used.
func (r *AuthRepository) MarkOTPUsed(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE otp_codes SET used = true WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	return err
}

// --- Reviews ---

// CreateReview inserts a new review.
func (r *AuthRepository) CreateReview(ctx context.Context, review *domain.Review) error {
	query := `
		INSERT INTO reviews (id, booking_id, property_id, guest_id, rating, comment, is_verified)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING created_at, updated_at`

	review.ID = uuid.New()
	return r.pool.QueryRow(ctx, query,
		review.ID, review.BookingID, review.PropertyID, review.GuestID,
		review.Rating, review.Comment, review.IsVerified,
	).Scan(&review.CreatedAt, &review.UpdatedAt)
}

// GetReviewByBookingID retrieves a review by booking ID.
func (r *AuthRepository) GetReviewByBookingID(ctx context.Context, bookingID uuid.UUID) (*domain.Review, error) {
	query := `
		SELECT id, booking_id, property_id, guest_id, rating, comment, is_verified, created_at, updated_at
		FROM reviews WHERE booking_id = $1`

	review := &domain.Review{}
	err := r.pool.QueryRow(ctx, query, bookingID).Scan(
		&review.ID, &review.BookingID, &review.PropertyID, &review.GuestID,
		&review.Rating, &review.Comment, &review.IsVerified,
		&review.CreatedAt, &review.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return review, nil
}

// ListReviewsByProperty retrieves reviews for a property.
func (r *AuthRepository) ListReviewsByProperty(ctx context.Context, propertyID uuid.UUID) ([]domain.Review, error) {
	query := `
		SELECT r.id, r.booking_id, r.property_id, r.guest_id, r.rating, r.comment,
			r.is_verified, r.created_at, r.updated_at,
			u.first_name, u.last_name, u.avatar_url
		FROM reviews r
		JOIN users u ON u.id = r.guest_id
		WHERE r.property_id = $1
		ORDER BY r.created_at DESC`

	rows, err := r.pool.Query(ctx, query, propertyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reviews []domain.Review
	for rows.Next() {
		var rv domain.Review
		var guestFirst, guestLast string
		var guestAvatar *string
		if err := rows.Scan(
			&rv.ID, &rv.BookingID, &rv.PropertyID, &rv.GuestID, &rv.Rating, &rv.Comment,
			&rv.IsVerified, &rv.CreatedAt, &rv.UpdatedAt,
			&guestFirst, &guestLast, &guestAvatar,
		); err != nil {
			return nil, err
		}
		rv.Guest = &domain.User{
			ID:        rv.GuestID,
			FirstName: guestFirst,
			LastName:  guestLast,
			AvatarURL: guestAvatar,
		}
		reviews = append(reviews, rv)
	}
	return reviews, nil
}
