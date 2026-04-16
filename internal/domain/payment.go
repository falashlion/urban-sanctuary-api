package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// PaymentProvider represents a payment service provider.
type PaymentProvider string

const (
	PaymentProviderMTNMoMo      PaymentProvider = "mtn_momo"
	PaymentProviderOrangeMoney  PaymentProvider = "orange_money"
)

// PaymentStatus represents the status of a payment transaction.
type PaymentStatus string

const (
	PaymentStatusInitiated PaymentStatus = "initiated"
	PaymentStatusPending   PaymentStatus = "pending"
	PaymentStatusCompleted PaymentStatus = "completed"
	PaymentStatusFailed    PaymentStatus = "failed"
	PaymentStatusRefunded  PaymentStatus = "refunded"
)

// Payment represents a payment transaction.
type Payment struct {
	ID                uuid.UUID        `json:"id"`
	BookingID         uuid.UUID        `json:"booking_id"`
	UserID            uuid.UUID        `json:"user_id"`
	Provider          PaymentProvider  `json:"provider"`
	ProviderReference *string          `json:"provider_reference,omitempty"`
	PhoneNumber       string           `json:"phone_number"`
	Amount            float64          `json:"amount"`
	Currency          string           `json:"currency"`
	Status            PaymentStatus    `json:"status"`
	WebhookPayload    *json.RawMessage `json:"webhook_payload,omitempty"`
	CreatedAt         time.Time        `json:"created_at"`
	UpdatedAt         time.Time        `json:"updated_at"`
}
