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

// PaymentHandler handles payment HTTP endpoints.
type PaymentHandler struct {
	svc *service.PaymentService
	log zerolog.Logger
}

// NewPaymentHandler creates a new PaymentHandler.
func NewPaymentHandler(svc *service.PaymentService, log zerolog.Logger) *PaymentHandler {
	return &PaymentHandler{svc: svc, log: log}
}

// Initiate handles POST /api/v1/payments/initiate
func (h *PaymentHandler) Initiate(c *gin.Context) {
	userID, err := uuid.Parse(middleware.GetUserID(c))
	if err != nil {
		response.Error(c, 401, "UNAUTHORIZED", "Invalid user ID", nil)
		return
	}

	var req dto.InitiatePaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 422, "VALIDATION_ERROR", "Invalid request body", parseValidationErrors(err))
		return
	}

	result, err := h.svc.Initiate(c.Request.Context(), userID, req)
	if err != nil {
		handleAppError(c, err, &h.log)
		return
	}

	response.Success(c, http.StatusCreated, "Payment initiated", result, nil)
}

// GetStatus handles GET /api/v1/payments/:id/status
func (h *PaymentHandler) GetStatus(c *gin.Context) {
	paymentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, 400, "BAD_REQUEST", "Invalid payment ID", nil)
		return
	}

	userID, _ := uuid.Parse(middleware.GetUserID(c))

	result, err := h.svc.GetStatus(c.Request.Context(), paymentID, userID)
	if err != nil {
		handleAppError(c, err, &h.log)
		return
	}

	response.Success(c, http.StatusOK, "Payment status retrieved", result, nil)
}

// MTNWebhook handles POST /api/v1/webhooks/mtn
func (h *PaymentHandler) MTNWebhook(c *gin.Context) {
	signature := c.GetHeader("X-Callback-Signature")

	if err := h.svc.HandleWebhook(c.Request.Context(), "mtn", signature, c.Request.Body); err != nil {
		h.log.Error().Err(err).Msg("MTN webhook processing failed")
		// Always return 200 to webhooks to prevent retries on validation errors
		c.JSON(http.StatusOK, gin.H{"status": "received"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "processed"})
}

// OrangeWebhook handles POST /api/v1/webhooks/orange
func (h *PaymentHandler) OrangeWebhook(c *gin.Context) {
	signature := c.GetHeader("X-Callback-Signature")

	if err := h.svc.HandleWebhook(c.Request.Context(), "orange", signature, c.Request.Body); err != nil {
		h.log.Error().Err(err).Msg("Orange webhook processing failed")
		c.JSON(http.StatusOK, gin.H{"status": "received"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "processed"})
}
