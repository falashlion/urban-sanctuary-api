package service

import (
	"context"
	"fmt"
	"time"

	"github.com/falashlion/urban-sanctuary-api/internal/domain"
	"github.com/falashlion/urban-sanctuary-api/internal/dto"
	"github.com/falashlion/urban-sanctuary-api/internal/repository"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

// BookingService handles booking-related business logic.
type BookingService struct {
	bookingRepo  *repository.BookingRepository
	propRepo     *repository.PropertyRepository
	userRepo     *repository.UserRepository
	authRepo     *repository.AuthRepository
	log          zerolog.Logger
}

// NewBookingService creates a new BookingService.
func NewBookingService(
	bookingRepo *repository.BookingRepository,
	propRepo *repository.PropertyRepository,
	userRepo *repository.UserRepository,
	authRepo *repository.AuthRepository,
	log zerolog.Logger,
) *BookingService {
	return &BookingService{
		bookingRepo: bookingRepo,
		propRepo:    propRepo,
		userRepo:    userRepo,
		authRepo:    authRepo,
		log:         log,
	}
}

// Create creates a new booking (step 1 of the checkout flow).
func (s *BookingService) Create(ctx context.Context, guestID uuid.UUID, req dto.CreateBookingRequest) (*dto.BookingResponse, error) {
	propertyID, err := uuid.Parse(req.PropertyID)
	if err != nil {
		return nil, domain.ErrValidation([]domain.ErrDetail{
			{Field: "property_id", Message: "invalid property ID"},
		})
	}

	// Parse dates
	checkIn, err := time.Parse("2006-01-02", req.CheckIn)
	if err != nil {
		return nil, domain.ErrValidation([]domain.ErrDetail{
			{Field: "check_in", Message: "invalid date format, use YYYY-MM-DD"},
		})
	}
	checkOut, err := time.Parse("2006-01-02", req.CheckOut)
	if err != nil {
		return nil, domain.ErrValidation([]domain.ErrDetail{
			{Field: "check_out", Message: "invalid date format, use YYYY-MM-DD"},
		})
	}

	// Validate dates
	if !checkOut.After(checkIn) {
		return nil, domain.ErrValidation([]domain.ErrDetail{
			{Field: "check_out", Message: "check-out must be after check-in"},
		})
	}
	if checkIn.Before(time.Now().Truncate(24 * time.Hour)) {
		return nil, domain.ErrValidation([]domain.ErrDetail{
			{Field: "check_in", Message: "check-in date cannot be in the past"},
		})
	}

	// Verify property exists and is published
	prop, err := s.propRepo.GetByID(ctx, propertyID)
	if err != nil {
		return nil, domain.ErrInternal(err)
	}
	if prop == nil {
		return nil, domain.ErrNotFound("Property")
	}
	if prop.Status != domain.PropertyStatusPublished {
		return nil, domain.ErrBadRequest("Property is not available for booking")
	}

	// Check guest count
	if req.GuestsCount > prop.MaxGuests {
		return nil, domain.ErrValidation([]domain.ErrDetail{
			{Field: "guests_count", Message: fmt.Sprintf("maximum %d guests allowed", prop.MaxGuests)},
		})
	}

	// Check for booking conflicts
	conflict, err := s.bookingRepo.CheckConflict(ctx, propertyID, checkIn, checkOut)
	if err != nil {
		return nil, domain.ErrInternal(err)
	}
	if conflict {
		return nil, domain.ErrBookingConflict()
	}

	// Calculate amounts
	nights := int(checkOut.Sub(checkIn).Hours() / 24)
	baseAmount := float64(nights) * prop.PricePerNight

	// Apply loyalty points discount
	var discountAmount float64
	var loyaltyPointsUsed int
	if req.UseLoyaltyPoints > 0 {
		guest, err := s.userRepo.GetByID(ctx, guestID)
		if err != nil {
			return nil, domain.ErrInternal(err)
		}
		if guest != nil && guest.LoyaltyPoints >= req.UseLoyaltyPoints {
			loyaltyPointsUsed = req.UseLoyaltyPoints
			discountAmount = float64(loyaltyPointsUsed) * 10.0 // 10 XAF per point
			if discountAmount > baseAmount {
				discountAmount = baseAmount
				loyaltyPointsUsed = int(baseAmount / 10.0)
			}
		}
	}

	totalAmount := baseAmount - discountAmount

	booking := &domain.Booking{
		PropertyID:        propertyID,
		GuestID:           guestID,
		CheckIn:           checkIn,
		CheckOut:          checkOut,
		GuestsCount:       req.GuestsCount,
		BaseAmount:        baseAmount,
		DiscountAmount:    discountAmount,
		TotalAmount:       totalAmount,
		Status:            domain.BookingStatusPending,
		LoyaltyPointsUsed: loyaltyPointsUsed,
	}

	if err := s.bookingRepo.Create(ctx, booking); err != nil {
		return nil, domain.ErrInternal(fmt.Errorf("failed to create booking: %w", err))
	}

	// Deduct loyalty points if used
	if loyaltyPointsUsed > 0 {
		_ = s.userRepo.DeductLoyaltyPoints(ctx, guestID, loyaltyPointsUsed)
	}

	resp := toBookingResponse(booking)
	return &resp, nil
}

// GetByID retrieves a booking by its ID.
func (s *BookingService) GetByID(ctx context.Context, id, userID uuid.UUID, isAdmin bool) (*dto.BookingResponse, error) {
	booking, err := s.bookingRepo.GetByID(ctx, id)
	if err != nil {
		return nil, domain.ErrInternal(err)
	}
	if booking == nil {
		return nil, domain.ErrNotFound("Booking")
	}

	// Authorization: only the guest, property owner, or admin can view
	if !isAdmin && booking.GuestID != userID {
		// Check if user is the property owner
		prop, err := s.propRepo.GetByID(ctx, booking.PropertyID)
		if err != nil || prop == nil || prop.OwnerID != userID {
			return nil, domain.ErrForbidden()
		}
	}

	resp := toBookingResponse(booking)
	return &resp, nil
}

// Cancel cancels a pending or confirmed booking.
func (s *BookingService) Cancel(ctx context.Context, id, userID uuid.UUID) error {
	booking, err := s.bookingRepo.GetByID(ctx, id)
	if err != nil {
		return domain.ErrInternal(err)
	}
	if booking == nil {
		return domain.ErrNotFound("Booking")
	}
	if booking.GuestID != userID {
		return domain.ErrForbidden()
	}
	if booking.Status != domain.BookingStatusPending && booking.Status != domain.BookingStatusConfirmed {
		return domain.ErrBookingNotPending()
	}

	if err := s.bookingRepo.UpdateStatus(ctx, id, domain.BookingStatusCancelled); err != nil {
		return domain.ErrInternal(err)
	}

	// Refund loyalty points if they were used
	if booking.LoyaltyPointsUsed > 0 {
		_ = s.userRepo.AddLoyaltyPoints(ctx, userID, booking.LoyaltyPointsUsed)
	}

	return nil
}

// SubmitReview creates a review for a completed booking.
func (s *BookingService) SubmitReview(ctx context.Context, bookingID, guestID uuid.UUID, req dto.CreateReviewRequest) (*dto.ReviewResponse, error) {
	booking, err := s.bookingRepo.GetByID(ctx, bookingID)
	if err != nil {
		return nil, domain.ErrInternal(err)
	}
	if booking == nil {
		return nil, domain.ErrNotFound("Booking")
	}
	if booking.GuestID != guestID {
		return nil, domain.ErrForbidden()
	}
	if booking.Status != domain.BookingStatusCompleted {
		return nil, domain.ErrBadRequest("Reviews can only be submitted for completed stays")
	}

	// Check if review already exists
	existing, _ := s.authRepo.GetReviewByBookingID(ctx, bookingID)
	if existing != nil {
		return nil, domain.ErrConflict("A review has already been submitted for this booking")
	}

	review := &domain.Review{
		BookingID:  bookingID,
		PropertyID: booking.PropertyID,
		GuestID:    guestID,
		Rating:     req.Rating,
		Comment:    req.Comment,
		IsVerified: true,
	}

	if err := s.authRepo.CreateReview(ctx, review); err != nil {
		return nil, domain.ErrInternal(err)
	}

	return &dto.ReviewResponse{
		ID:         review.ID.String(),
		BookingID:  review.BookingID.String(),
		PropertyID: review.PropertyID.String(),
		GuestID:    review.GuestID.String(),
		Rating:     review.Rating,
		Comment:    review.Comment,
		IsVerified: review.IsVerified,
		CreatedAt:  review.CreatedAt.Format(time.RFC3339),
	}, nil
}
