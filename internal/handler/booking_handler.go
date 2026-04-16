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

// BookingHandler handles booking HTTP endpoints.
type BookingHandler struct {
	svc *service.BookingService
	log zerolog.Logger
}

// NewBookingHandler creates a new BookingHandler.
func NewBookingHandler(svc *service.BookingService, log zerolog.Logger) *BookingHandler {
	return &BookingHandler{svc: svc, log: log}
}

// Create handles POST /api/v1/bookings
func (h *BookingHandler) Create(c *gin.Context) {
	guestID, err := uuid.Parse(middleware.GetUserID(c))
	if err != nil {
		response.Error(c, 401, "UNAUTHORIZED", "Invalid user ID", nil)
		return
	}

	var req dto.CreateBookingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 422, "VALIDATION_ERROR", "Invalid request body", parseValidationErrors(err))
		return
	}

	booking, err := h.svc.Create(c.Request.Context(), guestID, req)
	if err != nil {
		handleAppError(c, err, &h.log)
		return
	}

	response.Success(c, http.StatusCreated, "Booking created", booking, nil)
}

// GetByID handles GET /api/v1/bookings/:id
func (h *BookingHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, 400, "BAD_REQUEST", "Invalid booking ID", nil)
		return
	}

	userID, _ := uuid.Parse(middleware.GetUserID(c))
	isAdmin := middleware.GetUserRole(c) == "admin"

	booking, err := h.svc.GetByID(c.Request.Context(), id, userID, isAdmin)
	if err != nil {
		handleAppError(c, err, &h.log)
		return
	}

	response.Success(c, http.StatusOK, "Booking retrieved", booking, nil)
}

// Cancel handles POST /api/v1/bookings/:id/cancel
func (h *BookingHandler) Cancel(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, 400, "BAD_REQUEST", "Invalid booking ID", nil)
		return
	}

	userID, _ := uuid.Parse(middleware.GetUserID(c))

	if err := h.svc.Cancel(c.Request.Context(), id, userID); err != nil {
		handleAppError(c, err, &h.log)
		return
	}

	response.Success(c, http.StatusOK, "Booking cancelled", nil, nil)
}

// SubmitReview handles POST /api/v1/bookings/:id/review
func (h *BookingHandler) SubmitReview(c *gin.Context) {
	bookingID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, 400, "BAD_REQUEST", "Invalid booking ID", nil)
		return
	}

	guestID, _ := uuid.Parse(middleware.GetUserID(c))

	var req dto.CreateReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 422, "VALIDATION_ERROR", "Invalid request body", parseValidationErrors(err))
		return
	}

	review, err := h.svc.SubmitReview(c.Request.Context(), bookingID, guestID, req)
	if err != nil {
		handleAppError(c, err, &h.log)
		return
	}

	response.Success(c, http.StatusCreated, "Review submitted", review, nil)
}
