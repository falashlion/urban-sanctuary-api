package domain

import (
	"time"

	"github.com/google/uuid"
)

// Review represents a guest's review of a property stay.
type Review struct {
	ID         uuid.UUID `json:"id"`
	BookingID  uuid.UUID `json:"booking_id"`
	PropertyID uuid.UUID `json:"property_id"`
	GuestID    uuid.UUID `json:"guest_id"`
	Rating     int       `json:"rating"`
	Comment    *string   `json:"comment,omitempty"`
	IsVerified bool      `json:"is_verified"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`

	// Joined data
	Guest *User `json:"guest,omitempty"`
}

// NotificationChannel represents a delivery channel for notifications.
type NotificationChannel string

const (
	NotificationChannelEmail NotificationChannel = "email"
	NotificationChannelSMS   NotificationChannel = "sms"
	NotificationChannelPush  NotificationChannel = "push"
)

// Notification represents a user notification.
type Notification struct {
	ID        uuid.UUID           `json:"id"`
	UserID    uuid.UUID           `json:"user_id"`
	Type      string              `json:"type"`
	Channel   NotificationChannel `json:"channel"`
	Title     string              `json:"title"`
	Content   string              `json:"content"`
	SentAt    *time.Time          `json:"sent_at,omitempty"`
	ReadAt    *time.Time          `json:"read_at,omitempty"`
	CreatedAt time.Time           `json:"created_at"`
}

// TicketStatus represents the status of a support ticket.
type TicketStatus string

const (
	TicketStatusOpen       TicketStatus = "open"
	TicketStatusInProgress TicketStatus = "in_progress"
	TicketStatusResolved   TicketStatus = "resolved"
	TicketStatusClosed     TicketStatus = "closed"
)

// TicketPriority represents the priority of a support ticket.
type TicketPriority string

const (
	TicketPriorityLow    TicketPriority = "low"
	TicketPriorityMedium TicketPriority = "medium"
	TicketPriorityHigh   TicketPriority = "high"
	TicketPriorityUrgent TicketPriority = "urgent"
)

// SupportTicket represents a customer support ticket.
type SupportTicket struct {
	ID        uuid.UUID      `json:"id"`
	UserID    uuid.UUID      `json:"user_id"`
	Subject   string         `json:"subject"`
	Status    TicketStatus   `json:"status"`
	Priority  TicketPriority `json:"priority"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`

	// Joined data
	User     *User           `json:"user,omitempty"`
	Messages []TicketMessage `json:"messages,omitempty"`
}

// TicketMessage represents a message within a support ticket.
type TicketMessage struct {
	ID         uuid.UUID `json:"id"`
	TicketID   uuid.UUID `json:"ticket_id"`
	SenderID   uuid.UUID `json:"sender_id"`
	Content    string    `json:"content"`
	IsInternal bool      `json:"is_internal"`
	CreatedAt  time.Time `json:"created_at"`

	// Joined data
	Sender *User `json:"sender,omitempty"`
}

// Permission represents an RBAC permission entry.
type Permission struct {
	ID         uuid.UUID `json:"id"`
	Role       Role      `json:"role"`
	Module     string    `json:"module"`
	CanRead    bool      `json:"can_read"`
	CanWrite   bool      `json:"can_write"`
	CanDelete  bool      `json:"can_delete"`
	CanApprove bool      `json:"can_approve"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// SiteConfig represents a site-wide configuration variable.
type SiteConfig struct {
	ID          uuid.UUID  `json:"id"`
	Key         string     `json:"key"`
	Value       string     `json:"value"`
	Description *string    `json:"description,omitempty"`
	UpdatedBy   *uuid.UUID `json:"updated_by,omitempty"`
	UpdatedAt   time.Time  `json:"updated_at"`
	CreatedAt   time.Time  `json:"created_at"`
}

// AuditLog represents an audit trail entry.
type AuditLog struct {
	ID           uuid.UUID  `json:"id"`
	UserID       *uuid.UUID `json:"user_id,omitempty"`
	Action       string     `json:"action"`
	ResourceType string     `json:"resource_type"`
	ResourceID   *uuid.UUID `json:"resource_id,omitempty"`
	Metadata     any        `json:"metadata,omitempty"`
	IPAddress    *string    `json:"ip_address,omitempty"`
	UserAgent    *string    `json:"user_agent,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
}
