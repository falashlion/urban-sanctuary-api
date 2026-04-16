package service

import (
	"github.com/falashlion/urban-sanctuary-api/internal/domain"
	"github.com/falashlion/urban-sanctuary-api/internal/platform/email"
	"github.com/falashlion/urban-sanctuary-api/internal/platform/sms"
	"github.com/falashlion/urban-sanctuary-api/internal/repository"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"time"
)

// NotificationService handles sending notifications via email and SMS.
type NotificationService struct {
	adminRepo   *repository.AdminRepository
	emailClient email.EmailClient
	smsClient   sms.SMSClient
	log         zerolog.Logger
}

// NewNotificationService creates a new NotificationService.
func NewNotificationService(
	adminRepo *repository.AdminRepository,
	emailClient email.EmailClient,
	smsClient sms.SMSClient,
	log zerolog.Logger,
) *NotificationService {
	return &NotificationService{
		adminRepo:   adminRepo,
		emailClient: emailClient,
		smsClient:   smsClient,
		log:         log,
	}
}

// SendEmail sends an email notification and records it.
func (s *NotificationService) SendEmail(userID uuid.UUID, to, subject, htmlBody string) {
	if err := s.emailClient.Send(to, subject, htmlBody); err != nil {
		s.log.Error().Err(err).Str("to", to).Msg("Failed to send email notification")
		return
	}

	now := time.Now()
	_ = s.adminRepo.CreateNotification(nil, &domain.Notification{
		UserID:  userID,
		Type:    "email",
		Channel: domain.NotificationChannelEmail,
		Title:   subject,
		Content: htmlBody,
		SentAt:  &now,
	})
}

// SendSMS sends an SMS notification and records it.
func (s *NotificationService) SendSMS(userID uuid.UUID, to, message string) {
	if err := s.smsClient.Send(to, message); err != nil {
		s.log.Error().Err(err).Str("to", to).Msg("Failed to send SMS notification")
		return
	}

	now := time.Now()
	_ = s.adminRepo.CreateNotification(nil, &domain.Notification{
		UserID:  userID,
		Type:    "sms",
		Channel: domain.NotificationChannelSMS,
		Title:   "SMS Notification",
		Content: message,
		SentAt:  &now,
	})
}
