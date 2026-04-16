package domain

import (
	"time"

	"github.com/google/uuid"
)

// BookingStatus represents the status of a booking.
type BookingStatus string

const (
	BookingStatusPending   BookingStatus = "pending"
	BookingStatusConfirmed BookingStatus = "confirmed"
	BookingStatusCancelled BookingStatus = "cancelled"
	BookingStatusCompleted BookingStatus = "completed"
)

// Booking represents a property reservation.
type Booking struct {
	ID                  uuid.UUID     `json:"id"`
	PropertyID          uuid.UUID     `json:"property_id"`
	GuestID             uuid.UUID     `json:"guest_id"`
	CheckIn             time.Time     `json:"check_in"`
	CheckOut            time.Time     `json:"check_out"`
	Nights              int           `json:"nights"`
	GuestsCount         int           `json:"guests_count"`
	BaseAmount          float64       `json:"base_amount"`
	DiscountAmount      float64       `json:"discount_amount"`
	TotalAmount         float64       `json:"total_amount"`
	Status              BookingStatus `json:"status"`
	LoyaltyPointsEarned int           `json:"loyalty_points_earned"`
	LoyaltyPointsUsed   int           `json:"loyalty_points_used"`
	CreatedAt           time.Time     `json:"created_at"`
	UpdatedAt           time.Time     `json:"updated_at"`

	// Joined data (not always populated)
	Property *Property `json:"property,omitempty"`
	Guest    *User     `json:"guest,omitempty"`
}
