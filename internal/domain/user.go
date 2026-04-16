package domain

import (
	"time"

	"github.com/google/uuid"
)

// Role represents a user's role in the system.
type Role string

const (
	RoleGuest     Role = "guest"
	RoleHomeowner Role = "homeowner"
	RoleAdmin     Role = "admin"
)

// ValidRoles returns all valid roles.
func ValidRoles() []Role {
	return []Role{RoleGuest, RoleHomeowner, RoleAdmin}
}

// IsValid checks if the role is valid.
func (r Role) IsValid() bool {
	for _, valid := range ValidRoles() {
		if r == valid {
			return true
		}
	}
	return false
}

// User represents a user in the system.
type User struct {
	ID            uuid.UUID `json:"id"`
	Email         *string   `json:"email,omitempty"`
	Phone         *string   `json:"phone,omitempty"`
	PasswordHash  string    `json:"-"`
	FirstName     string    `json:"first_name"`
	LastName      string    `json:"last_name"`
	AvatarURL     *string   `json:"avatar_url,omitempty"`
	Role          Role      `json:"role"`
	IsVerified    bool      `json:"is_verified"`
	IsActive      bool      `json:"is_active"`
	TOTPSecret    *string   `json:"-"`
	TOTPEnabled   bool      `json:"totp_enabled"`
	LoyaltyPoints int       `json:"loyalty_points"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// FullName returns the user's full name.
func (u *User) FullName() string {
	return u.FirstName + " " + u.LastName
}

// RefreshToken represents a stored refresh token.
type RefreshToken struct {
	ID         uuid.UUID `json:"id"`
	UserID     uuid.UUID `json:"user_id"`
	TokenHash  string    `json:"-"`
	ExpiresAt  time.Time `json:"expires_at"`
	Revoked    bool      `json:"revoked"`
	DeviceInfo *string   `json:"device_info,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}

// OTPPurpose represents the purpose of an OTP code.
type OTPPurpose string

const (
	OTPPurposeLogin         OTPPurpose = "login"
	OTPPurposeRegistration  OTPPurpose = "registration"
	OTPPurposePasswordReset OTPPurpose = "password_reset"
	OTPPurposeVerification  OTPPurpose = "verification"
)

// OTPCode represents a one-time password code.
type OTPCode struct {
	ID         uuid.UUID  `json:"id"`
	Identifier string     `json:"identifier"`
	CodeHash   string     `json:"-"`
	Purpose    OTPPurpose `json:"purpose"`
	ExpiresAt  time.Time  `json:"expires_at"`
	Attempts   int        `json:"attempts"`
	Used       bool       `json:"used"`
	CreatedAt  time.Time  `json:"created_at"`
}
