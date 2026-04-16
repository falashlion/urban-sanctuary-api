package dto

// --- Payment DTOs ---

// InitiatePaymentRequest is the request body for initiating a payment.
type InitiatePaymentRequest struct {
	BookingID   string `json:"booking_id" validate:"required,uuid"`
	Provider    string `json:"provider" validate:"required,oneof=mtn_momo orange_money"`
	PhoneNumber string `json:"phone_number" validate:"required"`
}

// PaymentResponse is the public representation of a payment.
type PaymentResponse struct {
	ID                string  `json:"id"`
	BookingID         string  `json:"booking_id"`
	Provider          string  `json:"provider"`
	ProviderReference *string `json:"provider_reference,omitempty"`
	PhoneNumber       string  `json:"phone_number"`
	Amount            float64 `json:"amount"`
	Currency          string  `json:"currency"`
	Status            string  `json:"status"`
	CreatedAt         string  `json:"created_at"`
	UpdatedAt         string  `json:"updated_at"`
}

// --- User DTOs ---

// UpdateProfileRequest is the request body for updating user profile.
type UpdateProfileRequest struct {
	FirstName *string `json:"first_name" validate:"omitempty,min=1,max=100"`
	LastName  *string `json:"last_name" validate:"omitempty,min=1,max=100"`
	AvatarURL *string `json:"avatar_url" validate:"omitempty,url"`
}

// ChangePasswordRequest is the request body for changing password.
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" validate:"required"`
	NewPassword     string `json:"new_password" validate:"required,min=8,max=128"`
	TOTPCode        string `json:"totp_code" validate:"omitempty,len=6"`
}

// LoyaltyResponse is the response for loyalty points.
type LoyaltyResponse struct {
	Balance     int               `json:"balance"`
	PointsValue float64           `json:"points_value_xaf"`
	History     []LoyaltyEntry    `json:"history,omitempty"`
}

// LoyaltyEntry represents a loyalty points transaction.
type LoyaltyEntry struct {
	BookingID string `json:"booking_id"`
	Type      string `json:"type"`
	Points    int    `json:"points"`
	Date      string `json:"date"`
}
