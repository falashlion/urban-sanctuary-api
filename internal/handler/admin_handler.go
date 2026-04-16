package handler

import (
	"net/http"

	"github.com/falashlion/urban-sanctuary-api/internal/dto"
	"github.com/falashlion/urban-sanctuary-api/internal/middleware"
	"github.com/falashlion/urban-sanctuary-api/internal/service"
	"github.com/falashlion/urban-sanctuary-api/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

// AdminHandler handles admin HTTP endpoints.
type AdminHandler struct {
	svc *service.AdminService
	log zerolog.Logger
}

// NewAdminHandler creates a new AdminHandler.
func NewAdminHandler(svc *service.AdminService, log zerolog.Logger) *AdminHandler {
	return &AdminHandler{svc: svc, log: log}
}

// ListUsers handles GET /api/v1/admin/users
func (h *AdminHandler) ListUsers(c *gin.Context) {
	var query dto.AdminUserListQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Error(c, 422, "VALIDATION_ERROR", "Invalid query parameters", parseValidationErrors(err))
		return
	}

	users, total, err := h.svc.ListUsers(c.Request.Context(), query)
	if err != nil {
		handleAppError(c, err, &h.log)
		return
	}

	meta := response.PaginationMeta(query.Page, query.PerPage, total)
	response.Success(c, http.StatusOK, "Users retrieved", users, meta)
}

// UpdateUserRole handles PATCH /api/v1/admin/users/:id/role
func (h *AdminHandler) UpdateUserRole(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, 400, "BAD_REQUEST", "Invalid user ID", nil)
		return
	}

	var req dto.UpdateUserRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 422, "VALIDATION_ERROR", "Invalid request body", parseValidationErrors(err))
		return
	}

	if err := h.svc.UpdateUserRole(c.Request.Context(), userID, req.Role); err != nil {
		handleAppError(c, err, &h.log)
		return
	}

	response.Success(c, http.StatusOK, "User role updated", nil, nil)
}

// UpdateUserStatus handles PATCH /api/v1/admin/users/:id/status
func (h *AdminHandler) UpdateUserStatus(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, 400, "BAD_REQUEST", "Invalid user ID", nil)
		return
	}

	var req dto.UpdateUserStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 422, "VALIDATION_ERROR", "Invalid request body", parseValidationErrors(err))
		return
	}

	if err := h.svc.UpdateUserStatus(c.Request.Context(), userID, req.IsActive); err != nil {
		handleAppError(c, err, &h.log)
		return
	}

	response.Success(c, http.StatusOK, "User status updated", nil, nil)
}

// ListProperties handles GET /api/v1/admin/properties
func (h *AdminHandler) ListProperties(c *gin.Context) {
	var query dto.AdminPropertyListQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Error(c, 422, "VALIDATION_ERROR", "Invalid query parameters", parseValidationErrors(err))
		return
	}

	props, total, err := h.svc.ListProperties(c.Request.Context(), query)
	if err != nil {
		handleAppError(c, err, &h.log)
		return
	}

	meta := response.PaginationMeta(query.Page, query.PerPage, total)
	response.Success(c, http.StatusOK, "Properties retrieved", props, meta)
}

// ListBookings handles GET /api/v1/admin/bookings
func (h *AdminHandler) ListBookings(c *gin.Context) {
	var query dto.AdminBookingListQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Error(c, 422, "VALIDATION_ERROR", "Invalid query parameters", parseValidationErrors(err))
		return
	}

	bookings, total, err := h.svc.ListBookings(c.Request.Context(), query)
	if err != nil {
		handleAppError(c, err, &h.log)
		return
	}

	meta := response.PaginationMeta(query.Page, query.PerPage, total)
	response.Success(c, http.StatusOK, "Bookings retrieved", bookings, meta)
}

// ListTickets handles GET /api/v1/admin/tickets
func (h *AdminHandler) ListTickets(c *gin.Context) {
	var query dto.TicketListQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Error(c, 422, "VALIDATION_ERROR", "Invalid query parameters", parseValidationErrors(err))
		return
	}

	tickets, total, err := h.svc.ListTickets(c.Request.Context(), query)
	if err != nil {
		handleAppError(c, err, &h.log)
		return
	}

	meta := response.PaginationMeta(query.Page, query.PerPage, total)
	response.Success(c, http.StatusOK, "Tickets retrieved", tickets, meta)
}

// GetTicket handles GET /api/v1/admin/tickets/:id
func (h *AdminHandler) GetTicket(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, 400, "BAD_REQUEST", "Invalid ticket ID", nil)
		return
	}

	ticket, err := h.svc.GetTicket(c.Request.Context(), id)
	if err != nil {
		handleAppError(c, err, &h.log)
		return
	}

	response.Success(c, http.StatusOK, "Ticket retrieved", ticket, nil)
}

// ReplyToTicket handles POST /api/v1/admin/tickets/:id/reply
func (h *AdminHandler) ReplyToTicket(c *gin.Context) {
	ticketID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, 400, "BAD_REQUEST", "Invalid ticket ID", nil)
		return
	}

	senderID, _ := uuid.Parse(middleware.GetUserID(c))

	var req dto.TicketReplyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 422, "VALIDATION_ERROR", "Invalid request body", parseValidationErrors(err))
		return
	}

	if err := h.svc.ReplyToTicket(c.Request.Context(), ticketID, senderID, req); err != nil {
		handleAppError(c, err, &h.log)
		return
	}

	response.Success(c, http.StatusCreated, "Reply sent", nil, nil)
}

// UpdateTicketStatus handles PATCH /api/v1/admin/tickets/:id/status
func (h *AdminHandler) UpdateTicketStatus(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, 400, "BAD_REQUEST", "Invalid ticket ID", nil)
		return
	}

	var req dto.UpdateTicketStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 422, "VALIDATION_ERROR", "Invalid request body", parseValidationErrors(err))
		return
	}

	if err := h.svc.UpdateTicketStatus(c.Request.Context(), id, req.Status); err != nil {
		handleAppError(c, err, &h.log)
		return
	}

	response.Success(c, http.StatusOK, "Ticket status updated", nil, nil)
}

// GetPermissions handles GET /api/v1/admin/permissions
func (h *AdminHandler) GetPermissions(c *gin.Context) {
	perms, err := h.svc.GetPermissions(c.Request.Context())
	if err != nil {
		handleAppError(c, err, &h.log)
		return
	}

	response.Success(c, http.StatusOK, "Permissions retrieved", perms, nil)
}

// UpdatePermissions handles PUT /api/v1/admin/permissions
func (h *AdminHandler) UpdatePermissions(c *gin.Context) {
	var req dto.UpdatePermissionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 422, "VALIDATION_ERROR", "Invalid request body", parseValidationErrors(err))
		return
	}

	if err := h.svc.UpdatePermissions(c.Request.Context(), req); err != nil {
		handleAppError(c, err, &h.log)
		return
	}

	response.Success(c, http.StatusOK, "Permissions updated", nil, nil)
}

// GetSiteConfig handles GET /api/v1/admin/config
func (h *AdminHandler) GetSiteConfig(c *gin.Context) {
	configs, err := h.svc.GetSiteConfig(c.Request.Context())
	if err != nil {
		handleAppError(c, err, &h.log)
		return
	}

	response.Success(c, http.StatusOK, "Site config retrieved", configs, nil)
}

// UpdateSiteConfig handles PUT /api/v1/admin/config/:key
func (h *AdminHandler) UpdateSiteConfig(c *gin.Context) {
	key := c.Param("key")
	updatedBy, _ := uuid.Parse(middleware.GetUserID(c))

	var req dto.UpdateSiteConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 422, "VALIDATION_ERROR", "Invalid request body", parseValidationErrors(err))
		return
	}

	if err := h.svc.UpdateSiteConfig(c.Request.Context(), key, req, updatedBy); err != nil {
		handleAppError(c, err, &h.log)
		return
	}

	response.Success(c, http.StatusOK, "Config updated", nil, nil)
}

// GetAuditLogs handles GET /api/v1/admin/audit-logs
func (h *AdminHandler) GetAuditLogs(c *gin.Context) {
	var query dto.AuditLogQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Error(c, 422, "VALIDATION_ERROR", "Invalid query parameters", parseValidationErrors(err))
		return
	}

	logs, total, err := h.svc.GetAuditLogs(c.Request.Context(), query)
	if err != nil {
		handleAppError(c, err, &h.log)
		return
	}

	meta := response.PaginationMeta(query.Page, query.PerPage, total)
	response.Success(c, http.StatusOK, "Audit logs retrieved", logs, meta)
}
