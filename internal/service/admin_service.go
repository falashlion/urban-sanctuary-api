package service

import (
	"context"
	"time"

	"github.com/falashlion/urban-sanctuary-api/internal/domain"
	"github.com/falashlion/urban-sanctuary-api/internal/dto"
	"github.com/falashlion/urban-sanctuary-api/internal/repository"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

// AdminService handles admin-specific business logic.
type AdminService struct {
	userRepo    *repository.UserRepository
	propRepo    *repository.PropertyRepository
	bookingRepo *repository.BookingRepository
	adminRepo   *repository.AdminRepository
	log         zerolog.Logger
}

// NewAdminService creates a new AdminService.
func NewAdminService(
	userRepo *repository.UserRepository,
	propRepo *repository.PropertyRepository,
	bookingRepo *repository.BookingRepository,
	adminRepo *repository.AdminRepository,
	log zerolog.Logger,
) *AdminService {
	return &AdminService{
		userRepo:    userRepo,
		propRepo:    propRepo,
		bookingRepo: bookingRepo,
		adminRepo:   adminRepo,
		log:         log,
	}
}

// ListUsers lists all users with optional role filter.
func (s *AdminService) ListUsers(ctx context.Context, query dto.AdminUserListQuery) ([]dto.UserResponse, int64, error) {
	if query.Page < 1 {
		query.Page = 1
	}
	if query.PerPage < 1 || query.PerPage > 100 {
		query.PerPage = 20
	}

	users, total, err := s.userRepo.List(ctx, query.Role, query.Search, query.Page, query.PerPage)
	if err != nil {
		return nil, 0, domain.ErrInternal(err)
	}

	var responses []dto.UserResponse
	for _, u := range users {
		responses = append(responses, toUserResponse(&u))
	}
	return responses, total, nil
}

// UpdateUserRole changes a user's role.
func (s *AdminService) UpdateUserRole(ctx context.Context, userID uuid.UUID, role string) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return domain.ErrInternal(err)
	}
	if user == nil {
		return domain.ErrNotFound("User")
	}
	return s.userRepo.UpdateRole(ctx, userID, domain.Role(role))
}

// UpdateUserStatus activates or deactivates a user.
func (s *AdminService) UpdateUserStatus(ctx context.Context, userID uuid.UUID, isActive bool) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return domain.ErrInternal(err)
	}
	if user == nil {
		return domain.ErrNotFound("User")
	}
	return s.userRepo.UpdateStatus(ctx, userID, isActive)
}

// ListProperties lists all properties including drafts and archived.
func (s *AdminService) ListProperties(ctx context.Context, query dto.AdminPropertyListQuery) ([]dto.PropertyResponse, int64, error) {
	if query.Page < 1 {
		query.Page = 1
	}
	if query.PerPage < 1 || query.PerPage > 100 {
		query.PerPage = 20
	}

	statusFilter := query.Status
	if statusFilter == "" {
		statusFilter = "" // empty means show all statuses for admin
	}

	props, total, err := s.propRepo.List(ctx, "", 0, 0, 0, 0, "", "created_at", "desc", query.Page, query.PerPage, statusFilter)
	if err != nil {
		return nil, 0, domain.ErrInternal(err)
	}

	var responses []dto.PropertyResponse
	for _, p := range props {
		responses = append(responses, toPropertyResponse(&p))
	}
	return responses, total, nil
}

// ListBookings lists all bookings with optional status filter.
func (s *AdminService) ListBookings(ctx context.Context, query dto.AdminBookingListQuery) ([]dto.BookingResponse, int64, error) {
	if query.Page < 1 {
		query.Page = 1
	}
	if query.PerPage < 1 || query.PerPage > 100 {
		query.PerPage = 20
	}

	bookings, total, err := s.bookingRepo.ListAll(ctx, query.Status, query.Page, query.PerPage)
	if err != nil {
		return nil, 0, domain.ErrInternal(err)
	}

	var responses []dto.BookingResponse
	for _, b := range bookings {
		responses = append(responses, toBookingResponse(&b))
	}
	return responses, total, nil
}

// ListTickets lists support tickets.
func (s *AdminService) ListTickets(ctx context.Context, query dto.TicketListQuery) ([]dto.TicketResponse, int64, error) {
	if query.Page < 1 {
		query.Page = 1
	}
	if query.PerPage < 1 || query.PerPage > 100 {
		query.PerPage = 20
	}

	tickets, total, err := s.adminRepo.ListTickets(ctx, query.Status, query.Priority, query.Page, query.PerPage)
	if err != nil {
		return nil, 0, domain.ErrInternal(err)
	}

	var responses []dto.TicketResponse
	for _, t := range tickets {
		responses = append(responses, toTicketResponse(&t))
	}
	return responses, total, nil
}

// GetTicket retrieves a ticket with its messages.
func (s *AdminService) GetTicket(ctx context.Context, id uuid.UUID) (*dto.TicketResponse, error) {
	ticket, err := s.adminRepo.GetTicketByID(ctx, id)
	if err != nil {
		return nil, domain.ErrInternal(err)
	}
	if ticket == nil {
		return nil, domain.ErrNotFound("Ticket")
	}

	resp := toTicketResponse(ticket)
	return &resp, nil
}

// ReplyToTicket adds a message to a ticket.
func (s *AdminService) ReplyToTicket(ctx context.Context, ticketID, senderID uuid.UUID, req dto.TicketReplyRequest) error {
	ticket, err := s.adminRepo.GetTicketByID(ctx, ticketID)
	if err != nil {
		return domain.ErrInternal(err)
	}
	if ticket == nil {
		return domain.ErrNotFound("Ticket")
	}

	msg := &domain.TicketMessage{
		TicketID:   ticketID,
		SenderID:   senderID,
		Content:    req.Content,
		IsInternal: req.IsInternal,
	}

	return s.adminRepo.CreateTicketMessage(ctx, msg)
}

// UpdateTicketStatus updates a ticket's status.
func (s *AdminService) UpdateTicketStatus(ctx context.Context, id uuid.UUID, status string) error {
	ticket, err := s.adminRepo.GetTicketByID(ctx, id)
	if err != nil {
		return domain.ErrInternal(err)
	}
	if ticket == nil {
		return domain.ErrNotFound("Ticket")
	}
	return s.adminRepo.UpdateTicketStatus(ctx, id, domain.TicketStatus(status))
}

// GetPermissions retrieves the full RBAC permission matrix.
func (s *AdminService) GetPermissions(ctx context.Context) ([]dto.PermissionResponse, error) {
	perms, err := s.adminRepo.ListPermissions(ctx)
	if err != nil {
		return nil, domain.ErrInternal(err)
	}

	var responses []dto.PermissionResponse
	for _, p := range perms {
		responses = append(responses, dto.PermissionResponse{
			ID:         p.ID.String(),
			Role:       string(p.Role),
			Module:     p.Module,
			CanRead:    p.CanRead,
			CanWrite:   p.CanWrite,
			CanDelete:  p.CanDelete,
			CanApprove: p.CanApprove,
		})
	}
	return responses, nil
}

// UpdatePermissions updates the permission matrix.
func (s *AdminService) UpdatePermissions(ctx context.Context, req dto.UpdatePermissionsRequest) error {
	for _, entry := range req.Permissions {
		perm := &domain.Permission{
			Role:       domain.Role(entry.Role),
			Module:     entry.Module,
			CanRead:    entry.CanRead,
			CanWrite:   entry.CanWrite,
			CanDelete:  entry.CanDelete,
			CanApprove: entry.CanApprove,
		}
		if err := s.adminRepo.UpsertPermission(ctx, perm); err != nil {
			return domain.ErrInternal(err)
		}
	}
	return nil
}

// GetSiteConfig retrieves all site configuration variables.
func (s *AdminService) GetSiteConfig(ctx context.Context) ([]dto.SiteConfigResponse, error) {
	configs, err := s.adminRepo.ListSiteConfig(ctx)
	if err != nil {
		return nil, domain.ErrInternal(err)
	}

	var responses []dto.SiteConfigResponse
	for _, c := range configs {
		responses = append(responses, dto.SiteConfigResponse{
			ID:          c.ID.String(),
			Key:         c.Key,
			Value:       c.Value,
			Description: c.Description,
			UpdatedAt:   c.UpdatedAt.Format(time.RFC3339),
		})
	}
	return responses, nil
}

// UpdateSiteConfig updates a site config variable.
func (s *AdminService) UpdateSiteConfig(ctx context.Context, key string, req dto.UpdateSiteConfigRequest, updatedBy uuid.UUID) error {
	return s.adminRepo.UpdateSiteConfig(ctx, key, req.Value, req.Description, updatedBy)
}

// GetAuditLogs retrieves audit logs with filters.
func (s *AdminService) GetAuditLogs(ctx context.Context, query dto.AuditLogQuery) ([]dto.AuditLogResponse, int64, error) {
	if query.Page < 1 {
		query.Page = 1
	}
	if query.PerPage < 1 || query.PerPage > 100 {
		query.PerPage = 20
	}

	logs, total, err := s.adminRepo.ListAuditLogs(ctx, query.UserID, query.Action, query.ResourceType, query.Page, query.PerPage)
	if err != nil {
		return nil, 0, domain.ErrInternal(err)
	}

	var responses []dto.AuditLogResponse
	for _, l := range logs {
		var userIDStr *string
		if l.UserID != nil {
			s := l.UserID.String()
			userIDStr = &s
		}
		var resIDStr *string
		if l.ResourceID != nil {
			s := l.ResourceID.String()
			resIDStr = &s
		}
		responses = append(responses, dto.AuditLogResponse{
			ID:           l.ID.String(),
			UserID:       userIDStr,
			Action:       l.Action,
			ResourceType: l.ResourceType,
			ResourceID:   resIDStr,
			Metadata:     l.Metadata,
			IPAddress:    l.IPAddress,
			CreatedAt:    l.CreatedAt.Format(time.RFC3339),
		})
	}
	return responses, total, nil
}

// CreateAuditLog creates an audit log entry.
func (s *AdminService) CreateAuditLog(ctx context.Context, log *domain.AuditLog) error {
	return s.adminRepo.CreateAuditLog(ctx, log)
}

func toTicketResponse(t *domain.SupportTicket) dto.TicketResponse {
	resp := dto.TicketResponse{
		ID:        t.ID.String(),
		UserID:    t.UserID.String(),
		Subject:   t.Subject,
		Status:    string(t.Status),
		Priority:  string(t.Priority),
		CreatedAt: t.CreatedAt.Format(time.RFC3339),
		UpdatedAt: t.UpdatedAt.Format(time.RFC3339),
	}

	if t.User != nil {
		userResp := toUserResponse(t.User)
		resp.User = &userResp
	}

	for _, m := range t.Messages {
		msgResp := dto.TicketMessageResponse{
			ID:         m.ID.String(),
			TicketID:   m.TicketID.String(),
			SenderID:   m.SenderID.String(),
			Content:    m.Content,
			IsInternal: m.IsInternal,
			CreatedAt:  m.CreatedAt.Format(time.RFC3339),
		}
		if m.Sender != nil {
			senderResp := toUserResponse(m.Sender)
			msgResp.Sender = &senderResp
		}
		resp.Messages = append(resp.Messages, msgResp)
	}

	return resp
}
