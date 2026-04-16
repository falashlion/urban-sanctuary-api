package email

import (
	"fmt"
	"net/smtp"
	"strings"

	"github.com/rs/zerolog"
)

// EmailClient defines the interface for sending emails.
type EmailClient interface {
	Send(to, subject, htmlBody string) error
}

// SMTPClient implements EmailClient using SMTP.
type SMTPClient struct {
	host     string
	port     int
	username string
	password string
	from     string
	log      zerolog.Logger
}

// NewSMTPClient creates a new SMTP email client.
func NewSMTPClient(host string, port int, username, password, from string, log zerolog.Logger) *SMTPClient {
	return &SMTPClient{
		host:     host,
		port:     port,
		username: username,
		password: password,
		from:     from,
		log:      log,
	}
}

// Send sends an email via SMTP.
func (c *SMTPClient) Send(to, subject, htmlBody string) error {
	c.log.Info().
		Str("to", to).
		Str("subject", subject).
		Msg("Sending email via SMTP")

	addr := fmt.Sprintf("%s:%d", c.host, c.port)
	auth := smtp.PlainAuth("", c.username, c.password, c.host)

	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	msg := []byte(strings.Join([]string{
		"From: " + c.from,
		"To: " + to,
		"Subject: " + subject,
		mime,
		htmlBody,
	}, "\r\n"))

	err := smtp.SendMail(addr, auth, c.from, []string{to}, msg)
	if err != nil {
		c.log.Error().Err(err).Str("to", to).Msg("Failed to send email")
		return fmt.Errorf("failed to send email: %w", err)
	}

	c.log.Info().Str("to", to).Msg("Email sent successfully")
	return nil
}

// NoopEmailClient is a no-op email client for development/testing.
type NoopEmailClient struct {
	log zerolog.Logger
}

// NewNoopEmailClient creates a no-op email client.
func NewNoopEmailClient(log zerolog.Logger) *NoopEmailClient {
	return &NoopEmailClient{log: log}
}

// Send logs the email but does not actually send it.
func (c *NoopEmailClient) Send(to, subject, htmlBody string) error {
	c.log.Info().
		Str("to", to).
		Str("subject", subject).
		Msg("Email (noop) - not sent")
	return nil
}
