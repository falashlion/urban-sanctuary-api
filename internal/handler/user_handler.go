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

// UserHandler handles user profile HTTP endpoints.
type UserHandler struct {
	svc *service.UserService
	log zerolog.Logger
}

// NewUserHandler creates a new UserHandler.
func NewUserHandler(svc *service.UserService, log zerolog.Logger) *UserHandler {
	return &UserHandler{svc: svc, log: log}
}

// GetProfile handles GET /api/v1/users/me
func (h *UserHandler) GetProfile(c *gin.Context) {
	userID, err := uuid.Parse(middleware.GetUserID(c))
	if err != nil {
		response.Error(c, 401, "UNAUTHORIZED", "Invalid user ID", nil)
		return
	}

	result, err := h.svc.GetProfile(c.Request.Context(), userID)
	if err != nil {
		handleAppError(c, err, &h.log)
		return
	}

	response.Success(c, http.StatusOK, "Profile retrieved", result, nil)
}

// UpdateProfile handles PATCH /api/v1/users/me
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userID, err := uuid.Parse(middleware.GetUserID(c))
	if err != nil {
		response.Error(c, 401, "UNAUTHORIZED", "Invalid user ID", nil)
		return
	}

	var req dto.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 422, "VALIDATION_ERROR", "Invalid request body", parseValidationErrors(err))
		return
	}

	result, err := h.svc.UpdateProfile(c.Request.Context(), userID, req)
	if err != nil {
		handleAppError(c, err, &h.log)
		return
	}

	response.Success(c, http.StatusOK, "Profile updated", result, nil)
}

// ChangePassword handles PATCH /api/v1/users/me/password
func (h *UserHandler) ChangePassword(c *gin.Context) {
	userID, err := uuid.Parse(middleware.GetUserID(c))
	if err != nil {
		response.Error(c, 401, "UNAUTHORIZED", "Invalid user ID", nil)
		return
	}

	var req dto.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 422, "VALIDATION_ERROR", "Invalid request body", parseValidationErrors(err))
		return
	}

	if err := h.svc.ChangePassword(c.Request.Context(), userID, req); err != nil {
		handleAppError(c, err, &h.log)
		return
	}

	response.Success(c, http.StatusOK, "Password changed successfully", nil, nil)
}

// GetBookings handles GET /api/v1/users/me/bookings
func (h *UserHandler) GetBookings(c *gin.Context) {
	userID, err := uuid.Parse(middleware.GetUserID(c))
	if err != nil {
		response.Error(c, 401, "UNAUTHORIZED", "Invalid user ID", nil)
		return
	}

	var query dto.BookingListQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Error(c, 422, "VALIDATION_ERROR", "Invalid query parameters", parseValidationErrors(err))
		return
	}

	bookings, total, err := h.svc.GetBookings(c.Request.Context(), userID, query)
	if err != nil {
		handleAppError(c, err, &h.log)
		return
	}

	meta := response.PaginationMeta(query.Page, query.PerPage, total)
	response.Success(c, http.StatusOK, "Bookings retrieved", bookings, meta)
}

// GetLoyalty handles GET /api/v1/users/me/loyalty
func (h *UserHandler) GetLoyalty(c *gin.Context) {
	userID, err := uuid.Parse(middleware.GetUserID(c))
	if err != nil {
		response.Error(c, 401, "UNAUTHORIZED", "Invalid user ID", nil)
		return
	}

	result, err := h.svc.GetLoyalty(c.Request.Context(), userID)
	if err != nil {
		handleAppError(c, err, &h.log)
		return
	}

	response.Success(c, http.StatusOK, "Loyalty info retrieved", result, nil)
}

// GetNotifications handles GET /api/v1/users/me/notifications
func (h *UserHandler) GetNotifications(c *gin.Context) {
	userID, err := uuid.Parse(middleware.GetUserID(c))
	if err != nil {
		response.Error(c, 401, "UNAUTHORIZED", "Invalid user ID", nil)
		return
	}

	page := 1
	perPage := 20
	// Simple pagination parsing
	if p := c.Query("page"); p != "" {
		if _, err := c.GetQuery("page"); err {
			// use default
		}
	}

	notifications, total, err := h.svc.GetNotifications(c.Request.Context(), userID, page, perPage)
	if err != nil {
		handleAppError(c, err, &h.log)
		return
	}

	meta := response.PaginationMeta(page, perPage, total)
	response.Success(c, http.StatusOK, "Notifications retrieved", notifications, meta)
}
