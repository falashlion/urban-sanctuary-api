package sms

import (
	"fmt"

	"github.com/rs/zerolog"
)

// SMSClient defines the interface for sending SMS messages.
type SMSClient interface {
	Send(to, message string) error
}

// AfricasTalkingClient implements SMSClient using Africa's Talking API.
type AfricasTalkingClient struct {
	apiKey   string
	username string
	senderID string
	log      zerolog.Logger
}

// NewAfricasTalkingClient creates a new Africa's Talking SMS client.
func NewAfricasTalkingClient(apiKey, username, senderID string, log zerolog.Logger) *AfricasTalkingClient {
	return &AfricasTalkingClient{
		apiKey:   apiKey,
		username: username,
		senderID: senderID,
		log:      log,
	}
}

// Send sends an SMS message via Africa's Talking.
func (c *AfricasTalkingClient) Send(to, message string) error {
	c.log.Info().
		Str("to", to).
		Str("sender_id", c.senderID).
		Msg("Sending SMS via Africa's Talking")

	// In production, this would make an HTTP call to the Africa's Talking API.
	// For now, we log the message in development mode.
	if c.username == "sandbox" {
		c.log.Info().
			Str("to", to).
			Str("message", message).
			Msg("SMS sent (sandbox mode - not actually delivered)")
		return nil
	}

	// TODO: Implement actual Africa's Talking API call
	// POST https://api.africastalking.com/version1/messaging
	// Headers: apiKey, Accept: application/json
	// Body: username, to, message, from (senderID)

	return fmt.Errorf("production SMS sending not yet implemented")
}

// NoopSMSClient is a no-op SMS client for testing.
type NoopSMSClient struct {
	log zerolog.Logger
}

// NewNoopSMSClient creates a no-op SMS client.
func NewNoopSMSClient(log zerolog.Logger) *NoopSMSClient {
	return &NoopSMSClient{log: log}
}

// Send logs the message but does not actually send it.
func (c *NoopSMSClient) Send(to, message string) error {
	c.log.Info().
		Str("to", to).
		Str("message", message).
		Msg("SMS (noop) - message not sent")
	return nil
}
