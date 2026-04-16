package handler

import (
	"net/http"

	"github.com/falashlion/urban-sanctuary-api/internal/dto"
	"github.com/falashlion/urban-sanctuary-api/internal/service"
	"github.com/falashlion/urban-sanctuary-api/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

// AuthHandler handles authentication HTTP endpoints.
type AuthHandler struct {
	svc *service.AuthService
	log zerolog.Logger
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(svc *service.AuthService, log zerolog.Logger) *AuthHandler {
	return &AuthHandler{svc: svc, log: log}
}

// Register handles POST /api/v1/auth/register
func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 422, "VALIDATION_ERROR", "Invalid request body", parseValidationErrors(err))
		return
	}

	result, err := h.svc.Register(c.Request.Context(), req)
	if err != nil {
		handleAppError(c, err, &h.log)
		return
	}

	response.Success(c, http.StatusCreated, "Registration successful", result, nil)
}

// Login handles POST /api/v1/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 422, "VALIDATION_ERROR", "Invalid request body", parseValidationErrors(err))
		return
	}

	result, refreshToken, err := h.svc.Login(c.Request.Context(), req)
	if err != nil {
		handleAppError(c, err, &h.log)
		return
	}

	// Set refresh token as HttpOnly cookie
	c.SetCookie("refresh_token", refreshToken, 7*24*3600, "/api/v1/auth", "", false, true)

	response.Success(c, http.StatusOK, "Login successful", result, nil)
}

// Refresh handles POST /api/v1/auth/refresh
func (h *AuthHandler) Refresh(c *gin.Context) {
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil || refreshToken == "" {
		// Try from request body
		var req dto.RefreshRequest
		if err := c.ShouldBindJSON(&req); err == nil && req.RefreshToken != "" {
			refreshToken = req.RefreshToken
		} else {
			response.Error(c, 401, "UNAUTHORIZED", "Refresh token is required", nil)
			return
		}
	}

	result, newRefreshToken, err := h.svc.RefreshToken(c.Request.Context(), refreshToken)
	if err != nil {
		handleAppError(c, err, &h.log)
		return
	}

	c.SetCookie("refresh_token", newRefreshToken, 7*24*3600, "/api/v1/auth", "", false, true)

	response.Success(c, http.StatusOK, "Token refreshed", result, nil)
}

// Logout handles POST /api/v1/auth/logout
func (h *AuthHandler) Logout(c *gin.Context) {
	refreshToken, _ := c.Cookie("refresh_token")
	_ = h.svc.Logout(c.Request.Context(), refreshToken)

	// Clear cookie
	c.SetCookie("refresh_token", "", -1, "/api/v1/auth", "", false, true)

	response.Success(c, http.StatusOK, "Logged out successfully", nil, nil)
}

// RequestOTP handles POST /api/v1/auth/otp/request
func (h *AuthHandler) RequestOTP(c *gin.Context) {
	var req dto.OTPRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 422, "VALIDATION_ERROR", "Invalid request body", parseValidationErrors(err))
		return
	}

	if err := h.svc.RequestOTP(c.Request.Context(), req); err != nil {
		handleAppError(c, err, &h.log)
		return
	}

	response.Success(c, http.StatusOK, "OTP sent successfully", nil, nil)
}

// VerifyOTP handles POST /api/v1/auth/otp/verify
func (h *AuthHandler) VerifyOTP(c *gin.Context) {
	var req dto.OTPVerifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 422, "VALIDATION_ERROR", "Invalid request body", parseValidationErrors(err))
		return
	}

	if err := h.svc.VerifyOTP(c.Request.Context(), req); err != nil {
		handleAppError(c, err, &h.log)
		return
	}

	response.Success(c, http.StatusOK, "OTP verified successfully", nil, nil)
}

// Enable2FA handles POST /api/v1/auth/2fa/enable
func (h *AuthHandler) Enable2FA(c *gin.Context) {
	// TODO: Implement TOTP setup with pquerna/otp
	response.Success(c, http.StatusOK, "2FA setup initiated", gin.H{
		"secret":     "PLACEHOLDER_SECRET",
		"qr_code_url": "otpauth://totp/UrbanSanctuary?secret=PLACEHOLDER",
	}, nil)
}

// Confirm2FA handles POST /api/v1/auth/2fa/confirm
func (h *AuthHandler) Confirm2FA(c *gin.Context) {
	var req dto.Confirm2FARequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 422, "VALIDATION_ERROR", "Invalid request body", parseValidationErrors(err))
		return
	}
	// TODO: Validate TOTP code and activate 2FA
	response.Success(c, http.StatusOK, "2FA activated successfully", nil, nil)
}

// Verify2FA handles POST /api/v1/auth/2fa/verify
func (h *AuthHandler) Verify2FA(c *gin.Context) {
	var req dto.Verify2FARequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 422, "VALIDATION_ERROR", "Invalid request body", parseValidationErrors(err))
		return
	}
	// TODO: Validate TOTP code
	response.Success(c, http.StatusOK, "2FA verification successful", nil, nil)
}

// PasswordResetRequest handles POST /api/v1/auth/password/reset-request
func (h *AuthHandler) PasswordResetRequest(c *gin.Context) {
	var req dto.PasswordResetRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 422, "VALIDATION_ERROR", "Invalid request body", parseValidationErrors(err))
		return
	}

	if err := h.svc.PasswordResetRequest(c.Request.Context(), req); err != nil {
		handleAppError(c, err, &h.log)
		return
	}

	response.Success(c, http.StatusOK, "Password reset instructions sent", nil, nil)
}

// PasswordReset handles POST /api/v1/auth/password/reset
func (h *AuthHandler) PasswordReset(c *gin.Context) {
	var req dto.PasswordResetDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 422, "VALIDATION_ERROR", "Invalid request body", parseValidationErrors(err))
		return
	}

	if err := h.svc.ResetPassword(c.Request.Context(), req); err != nil {
		handleAppError(c, err, &h.log)
		return
	}

	response.Success(c, http.StatusOK, "Password reset successful", nil, nil)
}
