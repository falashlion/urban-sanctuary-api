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
	"golang.org/x/crypto/bcrypt"
)

// UserService handles user-related business logic.
type UserService struct {
	userRepo    *repository.UserRepository
	bookingRepo *repository.BookingRepository
	adminRepo   *repository.AdminRepository
	log         zerolog.Logger
}

// NewUserService creates a new UserService.
func NewUserService(
	userRepo *repository.UserRepository,
	bookingRepo *repository.BookingRepository,
	adminRepo *repository.AdminRepository,
	log zerolog.Logger,
) *UserService {
	return &UserService{
		userRepo:    userRepo,
		bookingRepo: bookingRepo,
		adminRepo:   adminRepo,
		log:         log,
	}
}

// GetProfile retrieves the authenticated user's profile.
func (s *UserService) GetProfile(ctx context.Context, userID uuid.UUID) (*dto.UserResponse, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, domain.ErrInternal(err)
	}
	if user == nil {
		return nil, domain.ErrNotFound("User")
	}

	resp := toUserResponse(user)
	return &resp, nil
}

// UpdateProfile updates the authenticated user's profile.
func (s *UserService) UpdateProfile(ctx context.Context, userID uuid.UUID, req dto.UpdateProfileRequest) (*dto.UserResponse, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, domain.ErrInternal(err)
	}
	if user == nil {
		return nil, domain.ErrNotFound("User")
	}

	if req.FirstName != nil {
		user.FirstName = *req.FirstName
	}
	if req.LastName != nil {
		user.LastName = *req.LastName
	}
	if req.AvatarURL != nil {
		user.AvatarURL = req.AvatarURL
	}

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, domain.ErrInternal(err)
	}

	resp := toUserResponse(user)
	return &resp, nil
}

// ChangePassword changes the authenticated user's password.
func (s *UserService) ChangePassword(ctx context.Context, userID uuid.UUID, req dto.ChangePasswordRequest) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return domain.ErrInternal(err)
	}
	if user == nil {
		return domain.ErrNotFound("User")
	}

	// Verify current password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.CurrentPassword)); err != nil {
		return domain.ErrValidation([]domain.ErrDetail{
			{Field: "current_password", Message: "current password is incorrect"},
		})
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), 12)
	if err != nil {
		return domain.ErrInternal(fmt.Errorf("failed to hash password: %w", err))
	}

	return s.userRepo.UpdatePassword(ctx, userID, string(hashedPassword))
}

// GetBookings retrieves the authenticated user's booking history.
func (s *UserService) GetBookings(ctx context.Context, userID uuid.UUID, query dto.BookingListQuery) ([]dto.BookingResponse, int64, error) {
	if query.Page < 1 {
		query.Page = 1
	}
	if query.PerPage < 1 || query.PerPage > 100 {
		query.PerPage = 20
	}

	bookings, total, err := s.bookingRepo.ListByGuest(ctx, userID, query.Status, query.Page, query.PerPage)
	if err != nil {
		return nil, 0, domain.ErrInternal(err)
	}

	var responses []dto.BookingResponse
	for _, b := range bookings {
		responses = append(responses, toBookingResponse(&b))
	}

	return responses, total, nil
}

// GetLoyalty retrieves the authenticated user's loyalty points info.
func (s *UserService) GetLoyalty(ctx context.Context, userID uuid.UUID) (*dto.LoyaltyResponse, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, domain.ErrInternal(err)
	}
	if user == nil {
		return nil, domain.ErrNotFound("User")
	}

	// Points value: each point = 10 XAF
	pointsValue := float64(user.LoyaltyPoints) * 10.0

	return &dto.LoyaltyResponse{
		Balance:     user.LoyaltyPoints,
		PointsValue: pointsValue,
	}, nil
}

// GetNotifications retrieves the authenticated user's notifications.
func (s *UserService) GetNotifications(ctx context.Context, userID uuid.UUID, page, perPage int) ([]domain.Notification, int64, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	return s.adminRepo.ListNotifications(ctx, userID, page, perPage)
}

func toBookingResponse(b *domain.Booking) dto.BookingResponse {
	return dto.BookingResponse{
		ID:                  b.ID.String(),
		PropertyID:          b.PropertyID.String(),
		GuestID:             b.GuestID.String(),
		CheckIn:             b.CheckIn.Format("2006-01-02"),
		CheckOut:            b.CheckOut.Format("2006-01-02"),
		Nights:              b.Nights,
		GuestsCount:         b.GuestsCount,
		BaseAmount:          b.BaseAmount,
		DiscountAmount:      b.DiscountAmount,
		TotalAmount:         b.TotalAmount,
		Status:              string(b.Status),
		LoyaltyPointsEarned: b.LoyaltyPointsEarned,
		LoyaltyPointsUsed:   b.LoyaltyPointsUsed,
		CreatedAt:           b.CreatedAt.Format(time.RFC3339),
		UpdatedAt:           b.UpdatedAt.Format(time.RFC3339),
	}
}
