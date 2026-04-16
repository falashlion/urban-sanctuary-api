package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/falashlion/urban-sanctuary-api/internal/domain"
	"github.com/falashlion/urban-sanctuary-api/internal/dto"
	"github.com/falashlion/urban-sanctuary-api/internal/platform/payment"
	"github.com/falashlion/urban-sanctuary-api/internal/repository"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

// PaymentService handles payment-related business logic.
type PaymentService struct {
	paymentRepo *repository.PaymentRepository
	bookingRepo *repository.BookingRepository
	userRepo    *repository.UserRepository
	mtnClient   *payment.MTNMoMoClient
	orangeClient *payment.OrangeMoneyClient
	log         zerolog.Logger
}

// NewPaymentService creates a new PaymentService.
func NewPaymentService(
	paymentRepo *repository.PaymentRepository,
	bookingRepo *repository.BookingRepository,
	userRepo *repository.UserRepository,
	mtnClient *payment.MTNMoMoClient,
	orangeClient *payment.OrangeMoneyClient,
	log zerolog.Logger,
) *PaymentService {
	return &PaymentService{
		paymentRepo:  paymentRepo,
		bookingRepo:  bookingRepo,
		userRepo:     userRepo,
		mtnClient:    mtnClient,
		orangeClient: orangeClient,
		log:          log,
	}
}

// Initiate starts a payment for a booking.
func (s *PaymentService) Initiate(ctx context.Context, userID uuid.UUID, req dto.InitiatePaymentRequest) (*dto.PaymentResponse, error) {
	bookingID, err := uuid.Parse(req.BookingID)
	if err != nil {
		return nil, domain.ErrValidation([]domain.ErrDetail{
			{Field: "booking_id", Message: "invalid booking ID"},
		})
	}

	// Verify booking exists and is pending
	booking, err := s.bookingRepo.GetByID(ctx, bookingID)
	if err != nil {
		return nil, domain.ErrInternal(err)
	}
	if booking == nil {
		return nil, domain.ErrNotFound("Booking")
	}
	if booking.GuestID != userID {
		return nil, domain.ErrForbidden()
	}
	if booking.Status != domain.BookingStatusPending {
		return nil, domain.ErrBookingNotPending()
	}

	// Get the appropriate payment client
	client, err := payment.GetClient(req.Provider, s.mtnClient, s.orangeClient)
	if err != nil {
		return nil, domain.ErrValidation([]domain.ErrDetail{
			{Field: "provider", Message: err.Error()},
		})
	}

	// Create payment record
	p := &domain.Payment{
		BookingID:   bookingID,
		UserID:      userID,
		Provider:    domain.PaymentProvider(req.Provider),
		PhoneNumber: req.PhoneNumber,
		Amount:      booking.TotalAmount,
		Currency:    "XAF",
		Status:      domain.PaymentStatusInitiated,
	}

	if err := s.paymentRepo.Create(ctx, p); err != nil {
		return nil, domain.ErrInternal(fmt.Errorf("failed to create payment: %w", err))
	}

	// Initiate payment with provider
	result, err := client.Initiate(ctx, payment.PaymentRequest{
		Amount:      booking.TotalAmount,
		Currency:    "XAF",
		PhoneNumber: req.PhoneNumber,
		Description: fmt.Sprintf("Urban Sanctuary Booking %s", bookingID.String()[:8]),
		ExternalID:  p.ID.String(),
	})
	if err != nil {
		_ = s.paymentRepo.UpdateStatus(ctx, p.ID, domain.PaymentStatusFailed)
		return nil, domain.ErrPaymentFailed(fmt.Sprintf("Payment initiation failed: %v", err))
	}

	// Update payment with provider reference
	_ = s.paymentRepo.UpdateProviderReference(ctx, p.ID, result.ProviderReference)
	_ = s.paymentRepo.UpdateStatus(ctx, p.ID, domain.PaymentStatusPending)

	provRef := result.ProviderReference
	return &dto.PaymentResponse{
		ID:                p.ID.String(),
		BookingID:         p.BookingID.String(),
		Provider:          string(p.Provider),
		ProviderReference: &provRef,
		PhoneNumber:       p.PhoneNumber,
		Amount:            p.Amount,
		Currency:          p.Currency,
		Status:            string(domain.PaymentStatusPending),
		CreatedAt:         p.CreatedAt.Format(time.RFC3339),
		UpdatedAt:         p.UpdatedAt.Format(time.RFC3339),
	}, nil
}

// GetStatus retrieves the current status of a payment.
func (s *PaymentService) GetStatus(ctx context.Context, paymentID, userID uuid.UUID) (*dto.PaymentResponse, error) {
	p, err := s.paymentRepo.GetByID(ctx, paymentID)
	if err != nil {
		return nil, domain.ErrInternal(err)
	}
	if p == nil {
		return nil, domain.ErrNotFound("Payment")
	}
	if p.UserID != userID {
		return nil, domain.ErrForbidden()
	}

	var provRef *string
	if p.ProviderReference != nil {
		provRef = p.ProviderReference
	}

	return &dto.PaymentResponse{
		ID:                p.ID.String(),
		BookingID:         p.BookingID.String(),
		Provider:          string(p.Provider),
		ProviderReference: provRef,
		PhoneNumber:       p.PhoneNumber,
		Amount:            p.Amount,
		Currency:          p.Currency,
		Status:            string(p.Status),
		CreatedAt:         p.CreatedAt.Format(time.RFC3339),
		UpdatedAt:         p.UpdatedAt.Format(time.RFC3339),
	}, nil
}

// HandleWebhook processes a payment provider webhook callback.
func (s *PaymentService) HandleWebhook(ctx context.Context, provider string, signature string, body io.Reader) error {
	bodyBytes, err := io.ReadAll(body)
	if err != nil {
		return domain.ErrInternal(fmt.Errorf("failed to read webhook body: %w", err))
	}

	// Validate webhook signature
	var client payment.PaymentClient
	if provider == "mtn" {
		client = s.mtnClient
	} else if provider == "orange" {
		client = s.orangeClient
	} else {
		return domain.ErrBadRequest("Unknown payment provider")
	}

	if !client.ValidateWebhook(signature, bodyBytes) {
		return domain.ErrUnauthorized()
	}

	// Parse the webhook payload (structure varies by provider)
	var payload struct {
		ExternalID string `json:"externalId"`
		Status     string `json:"status"`
		Reference  string `json:"financialTransactionId"`
	}
	if err := json.Unmarshal(bodyBytes, &payload); err != nil {
		s.log.Error().Err(err).Msg("Failed to parse webhook payload")
		return domain.ErrBadRequest("Invalid webhook payload")
	}

	// Find the payment by external ID (our payment ID)
	paymentID, err := uuid.Parse(payload.ExternalID)
	if err != nil {
		return domain.ErrBadRequest("Invalid external ID in webhook")
	}

	p, err := s.paymentRepo.GetByID(ctx, paymentID)
	if err != nil || p == nil {
		return domain.ErrNotFound("Payment")
	}

	// Store raw webhook payload
	rawPayload := json.RawMessage(bodyBytes)
	_ = s.paymentRepo.UpdateWebhookPayload(ctx, p.ID, rawPayload)

	// Update payment status based on webhook
	var newStatus domain.PaymentStatus
	switch payload.Status {
	case "SUCCESSFUL", "completed":
		newStatus = domain.PaymentStatusCompleted
	case "FAILED", "failed":
		newStatus = domain.PaymentStatusFailed
	default:
		newStatus = domain.PaymentStatusPending
	}

	if err := s.paymentRepo.UpdateStatus(ctx, p.ID, newStatus); err != nil {
		return domain.ErrInternal(err)
	}

	// If payment completed, update booking status and award loyalty points
	if newStatus == domain.PaymentStatusCompleted {
		_ = s.bookingRepo.UpdateStatus(ctx, p.BookingID, domain.BookingStatusConfirmed)

		// Award loyalty points (100 points per booking)
		_ = s.bookingRepo.UpdateLoyaltyPoints(ctx, p.BookingID, 100)
		_ = s.userRepo.AddLoyaltyPoints(ctx, p.UserID, 100)

		s.log.Info().
			Str("payment_id", p.ID.String()).
			Str("booking_id", p.BookingID.String()).
			Msg("Payment completed, booking confirmed")
	}

	return nil
}
