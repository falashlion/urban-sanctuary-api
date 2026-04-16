package dto

// --- Auth DTOs ---

// RegisterRequest is the request body for user registration.
type RegisterRequest struct {
	Email     *string `json:"email" validate:"omitempty,email"`
	Phone     *string `json:"phone" validate:"omitempty,e164"`
	Password  string  `json:"password" validate:"required,min=8,max=128"`
	FirstName string  `json:"first_name" validate:"required,min=1,max=100"`
	LastName  string  `json:"last_name" validate:"required,min=1,max=100"`
}

// LoginRequest is the request body for user login.
type LoginRequest struct {
	Email    *string `json:"email" validate:"omitempty,email"`
	Phone    *string `json:"phone" validate:"omitempty,e164"`
	Password string  `json:"password" validate:"required"`
}

// RefreshRequest is the request body for token refresh (token comes from cookie).
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// OTPRequestDTO is the request body for OTP request.
type OTPRequestDTO struct {
	Identifier string `json:"identifier" validate:"required"`
	Purpose    string `json:"purpose" validate:"required,oneof=login registration password_reset verification"`
}

// OTPVerifyRequest is the request body for OTP verification.
type OTPVerifyRequest struct {
	Identifier string `json:"identifier" validate:"required"`
	Code       string `json:"code" validate:"required,len=6"`
	Purpose    string `json:"purpose" validate:"required,oneof=login registration password_reset verification"`
}

// Enable2FAResponse is the response for 2FA enable request.
type Enable2FAResponse struct {
	Secret    string `json:"secret"`
	QRCodeURL string `json:"qr_code_url"`
}

// Confirm2FARequest is the request body for 2FA confirmation.
type Confirm2FARequest struct {
	Code string `json:"code" validate:"required,len=6"`
}

// Verify2FARequest is the request body for 2FA verification.
type Verify2FARequest struct {
	Code string `json:"code" validate:"required,len=6"`
}

// PasswordResetRequestDTO is the request body for password reset initiation.
type PasswordResetRequestDTO struct {
	Email *string `json:"email" validate:"omitempty,email"`
	Phone *string `json:"phone" validate:"omitempty,e164"`
}

// PasswordResetDTO is the request body for password reset.
type PasswordResetDTO struct {
	Token       string `json:"token" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=8,max=128"`
}

// AuthResponse is the response for successful authentication.
type AuthResponse struct {
	AccessToken  string       `json:"access_token"`
	TokenType    string       `json:"token_type"`
	ExpiresIn    int          `json:"expires_in"`
	User         UserResponse `json:"user"`
}

// UserResponse is the public representation of a user.
type UserResponse struct {
	ID            string  `json:"id"`
	Email         *string `json:"email,omitempty"`
	Phone         *string `json:"phone,omitempty"`
	FirstName     string  `json:"first_name"`
	LastName      string  `json:"last_name"`
	AvatarURL     *string `json:"avatar_url,omitempty"`
	Role          string  `json:"role"`
	IsVerified    bool    `json:"is_verified"`
	TOTPEnabled   bool    `json:"totp_enabled"`
	LoyaltyPoints int     `json:"loyalty_points"`
	CreatedAt     string  `json:"created_at"`
}
