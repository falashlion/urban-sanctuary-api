package dto

// --- Booking DTOs ---

// CreateBookingRequest is the request body for creating a booking.
type CreateBookingRequest struct {
	PropertyID       string `json:"property_id" validate:"required,uuid"`
	CheckIn          string `json:"check_in" validate:"required"`
	CheckOut         string `json:"check_out" validate:"required"`
	GuestsCount      int    `json:"guests_count" validate:"required,gte=1"`
	UseLoyaltyPoints int    `json:"use_loyalty_points" validate:"omitempty,gte=0"`
}

// BookingResponse is the public representation of a booking.
type BookingResponse struct {
	ID                  string            `json:"id"`
	PropertyID          string            `json:"property_id"`
	GuestID             string            `json:"guest_id"`
	CheckIn             string            `json:"check_in"`
	CheckOut            string            `json:"check_out"`
	Nights              int               `json:"nights"`
	GuestsCount         int               `json:"guests_count"`
	BaseAmount          float64           `json:"base_amount"`
	DiscountAmount      float64           `json:"discount_amount"`
	TotalAmount         float64           `json:"total_amount"`
	Status              string            `json:"status"`
	LoyaltyPointsEarned int               `json:"loyalty_points_earned"`
	LoyaltyPointsUsed   int               `json:"loyalty_points_used"`
	Property            *PropertyResponse `json:"property,omitempty"`
	Guest               *UserResponse     `json:"guest,omitempty"`
	CreatedAt           string            `json:"created_at"`
	UpdatedAt           string            `json:"updated_at"`
}

// BookingListQuery contains query parameters for listing bookings.
type BookingListQuery struct {
	Page    int    `form:"page,default=1" validate:"gte=1"`
	PerPage int    `form:"per_page,default=20" validate:"gte=1,lte=100"`
	Status  string `form:"status" validate:"omitempty,oneof=pending confirmed cancelled completed"`
}

// --- Review DTOs ---

// CreateReviewRequest is the request body for submitting a review.
type CreateReviewRequest struct {
	Rating  int     `json:"rating" validate:"required,gte=1,lte=5"`
	Comment *string `json:"comment" validate:"omitempty,max=2000"`
}

// ReviewResponse is the public representation of a review.
type ReviewResponse struct {
	ID         string        `json:"id"`
	BookingID  string        `json:"booking_id"`
	PropertyID string        `json:"property_id"`
	GuestID    string        `json:"guest_id"`
	Rating     int           `json:"rating"`
	Comment    *string       `json:"comment,omitempty"`
	IsVerified bool          `json:"is_verified"`
	Guest      *UserResponse `json:"guest,omitempty"`
	CreatedAt  string        `json:"created_at"`
}
