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

// UserRepository handles database operations for users.
type UserRepository struct {
	pool *pgxpool.Pool
}

// NewUserRepository creates a new UserRepository.
func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

// Create inserts a new user into the database.
func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	query := `
		INSERT INTO users (id, email, phone, password_hash, first_name, last_name, role, is_verified)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING created_at, updated_at`

	user.ID = uuid.New()
	return r.pool.QueryRow(ctx, query,
		user.ID, user.Email, user.Phone, user.PasswordHash,
		user.FirstName, user.LastName, user.Role, user.IsVerified,
	).Scan(&user.CreatedAt, &user.UpdatedAt)
}

// GetByID retrieves a user by their ID.
func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	query := `
		SELECT id, email, phone, password_hash, first_name, last_name, avatar_url,
			role, is_verified, is_active, totp_secret, totp_enabled, loyalty_points,
			created_at, updated_at
		FROM users WHERE id = $1`

	user := &domain.User{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&user.ID, &user.Email, &user.Phone, &user.PasswordHash,
		&user.FirstName, &user.LastName, &user.AvatarURL,
		&user.Role, &user.IsVerified, &user.IsActive,
		&user.TOTPSecret, &user.TOTPEnabled, &user.LoyaltyPoints,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}
	return user, nil
}

// GetByEmail retrieves a user by their email address.
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `
		SELECT id, email, phone, password_hash, first_name, last_name, avatar_url,
			role, is_verified, is_active, totp_secret, totp_enabled, loyalty_points,
			created_at, updated_at
		FROM users WHERE email = $1`

	user := &domain.User{}
	err := r.pool.QueryRow(ctx, query, email).Scan(
		&user.ID, &user.Email, &user.Phone, &user.PasswordHash,
		&user.FirstName, &user.LastName, &user.AvatarURL,
		&user.Role, &user.IsVerified, &user.IsActive,
		&user.TOTPSecret, &user.TOTPEnabled, &user.LoyaltyPoints,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}
	return user, nil
}

// GetByPhone retrieves a user by their phone number.
func (r *UserRepository) GetByPhone(ctx context.Context, phone string) (*domain.User, error) {
	query := `
		SELECT id, email, phone, password_hash, first_name, last_name, avatar_url,
			role, is_verified, is_active, totp_secret, totp_enabled, loyalty_points,
			created_at, updated_at
		FROM users WHERE phone = $1`

	user := &domain.User{}
	err := r.pool.QueryRow(ctx, query, phone).Scan(
		&user.ID, &user.Email, &user.Phone, &user.PasswordHash,
		&user.FirstName, &user.LastName, &user.AvatarURL,
		&user.Role, &user.IsVerified, &user.IsActive,
		&user.TOTPSecret, &user.TOTPEnabled, &user.LoyaltyPoints,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user by phone: %w", err)
	}
	return user, nil
}

// Update updates a user's profile fields.
func (r *UserRepository) Update(ctx context.Context, user *domain.User) error {
	query := `
		UPDATE users SET first_name = $2, last_name = $3, avatar_url = $4, updated_at = NOW()
		WHERE id = $1`

	_, err := r.pool.Exec(ctx, query, user.ID, user.FirstName, user.LastName, user.AvatarURL)
	return err
}

// UpdatePassword updates a user's password hash.
func (r *UserRepository) UpdatePassword(ctx context.Context, userID uuid.UUID, passwordHash string) error {
	query := `UPDATE users SET password_hash = $2, updated_at = NOW() WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, userID, passwordHash)
	return err
}

// UpdateRole updates a user's role.
func (r *UserRepository) UpdateRole(ctx context.Context, userID uuid.UUID, role domain.Role) error {
	query := `UPDATE users SET role = $2, updated_at = NOW() WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, userID, role)
	return err
}

// UpdateStatus updates a user's active status.
func (r *UserRepository) UpdateStatus(ctx context.Context, userID uuid.UUID, isActive bool) error {
	query := `UPDATE users SET is_active = $2, updated_at = NOW() WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, userID, isActive)
	return err
}

// UpdateVerified marks a user as verified.
func (r *UserRepository) UpdateVerified(ctx context.Context, userID uuid.UUID, verified bool) error {
	query := `UPDATE users SET is_verified = $2, updated_at = NOW() WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, userID, verified)
	return err
}

// UpdateTOTP updates a user's TOTP settings.
func (r *UserRepository) UpdateTOTP(ctx context.Context, userID uuid.UUID, secret *string, enabled bool) error {
	query := `UPDATE users SET totp_secret = $2, totp_enabled = $3, updated_at = NOW() WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, userID, secret, enabled)
	return err
}

// UpdateLoyaltyPoints updates a user's loyalty points balance.
func (r *UserRepository) UpdateLoyaltyPoints(ctx context.Context, userID uuid.UUID, points int) error {
	query := `UPDATE users SET loyalty_points = $2, updated_at = NOW() WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, userID, points)
	return err
}

// AddLoyaltyPoints adds loyalty points to a user's balance.
func (r *UserRepository) AddLoyaltyPoints(ctx context.Context, userID uuid.UUID, points int) error {
	query := `UPDATE users SET loyalty_points = loyalty_points + $2, updated_at = NOW() WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, userID, points)
	return err
}

// DeductLoyaltyPoints deducts loyalty points from a user's balance.
func (r *UserRepository) DeductLoyaltyPoints(ctx context.Context, userID uuid.UUID, points int) error {
	query := `UPDATE users SET loyalty_points = loyalty_points - $2, updated_at = NOW() WHERE id = $1 AND loyalty_points >= $2`
	result, err := r.pool.Exec(ctx, query, userID, points)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("insufficient loyalty points")
	}
	return nil
}

// List retrieves users with filters and pagination.
func (r *UserRepository) List(ctx context.Context, role string, search string, page, perPage int) ([]domain.User, int64, error) {
	countQuery := `SELECT COUNT(*) FROM users WHERE 1=1`
	dataQuery := `
		SELECT id, email, phone, password_hash, first_name, last_name, avatar_url,
			role, is_verified, is_active, totp_secret, totp_enabled, loyalty_points,
			created_at, updated_at
		FROM users WHERE 1=1`

	args := []interface{}{}
	argIdx := 1

	if role != "" {
		filter := fmt.Sprintf(" AND role = $%d", argIdx)
		countQuery += filter
		dataQuery += filter
		args = append(args, role)
		argIdx++
	}
	if search != "" {
		filter := fmt.Sprintf(" AND (first_name ILIKE $%d OR last_name ILIKE $%d OR email ILIKE $%d)", argIdx, argIdx, argIdx)
		countQuery += filter
		dataQuery += filter
		args = append(args, "%"+search+"%")
		argIdx++
	}

	var total int64
	err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
	}

	dataQuery += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, perPage, (page-1)*perPage)

	rows, err := r.pool.Query(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list users: %w", err)
	}
	defer rows.Close()

	var users []domain.User
	for rows.Next() {
		var user domain.User
		if err := rows.Scan(
			&user.ID, &user.Email, &user.Phone, &user.PasswordHash,
			&user.FirstName, &user.LastName, &user.AvatarURL,
			&user.Role, &user.IsVerified, &user.IsActive,
			&user.TOTPSecret, &user.TOTPEnabled, &user.LoyaltyPoints,
			&user.CreatedAt, &user.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}

	return users, total, nil
}

// Suppress unused import warning
var _ = json.Marshal
