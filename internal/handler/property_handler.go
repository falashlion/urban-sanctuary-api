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

// PropertyHandler handles property HTTP endpoints.
type PropertyHandler struct {
	svc *service.PropertyService
	log zerolog.Logger
}

// NewPropertyHandler creates a new PropertyHandler.
func NewPropertyHandler(svc *service.PropertyService, log zerolog.Logger) *PropertyHandler {
	return &PropertyHandler{svc: svc, log: log}
}

// List handles GET /api/v1/properties
func (h *PropertyHandler) List(c *gin.Context) {
	var query dto.PropertyListQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Error(c, 422, "VALIDATION_ERROR", "Invalid query parameters", parseValidationErrors(err))
		return
	}

	props, total, err := h.svc.List(c.Request.Context(), query)
	if err != nil {
		handleAppError(c, err, &h.log)
		return
	}

	meta := response.PaginationMeta(query.Page, query.PerPage, total)
	response.Success(c, http.StatusOK, "Properties retrieved", props, meta)
}

// GetByID handles GET /api/v1/properties/:id
func (h *PropertyHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, 400, "BAD_REQUEST", "Invalid property ID", nil)
		return
	}

	prop, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		handleAppError(c, err, &h.log)
		return
	}

	response.Success(c, http.StatusOK, "Property retrieved", prop, nil)
}

// GetAvailability handles GET /api/v1/properties/:id/availability
func (h *PropertyHandler) GetAvailability(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, 400, "BAD_REQUEST", "Invalid property ID", nil)
		return
	}

	ranges, err := h.svc.GetAvailability(c.Request.Context(), id)
	if err != nil {
		handleAppError(c, err, &h.log)
		return
	}

	response.Success(c, http.StatusOK, "Availability retrieved", ranges, nil)
}

// Create handles POST /api/v1/properties
func (h *PropertyHandler) Create(c *gin.Context) {
	ownerID, err := uuid.Parse(middleware.GetUserID(c))
	if err != nil {
		response.Error(c, 401, "UNAUTHORIZED", "Invalid user ID", nil)
		return
	}

	var req dto.CreatePropertyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 422, "VALIDATION_ERROR", "Invalid request body", parseValidationErrors(err))
		return
	}

	prop, err := h.svc.Create(c.Request.Context(), ownerID, req)
	if err != nil {
		handleAppError(c, err, &h.log)
		return
	}

	response.Success(c, http.StatusCreated, "Property created", prop, nil)
}

// Update handles PATCH /api/v1/properties/:id
func (h *PropertyHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, 400, "BAD_REQUEST", "Invalid property ID", nil)
		return
	}

	ownerID, _ := uuid.Parse(middleware.GetUserID(c))
	isAdmin := middleware.GetUserRole(c) == "admin"

	var req dto.UpdatePropertyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 422, "VALIDATION_ERROR", "Invalid request body", parseValidationErrors(err))
		return
	}

	prop, err := h.svc.Update(c.Request.Context(), id, ownerID, req, isAdmin)
	if err != nil {
		handleAppError(c, err, &h.log)
		return
	}

	response.Success(c, http.StatusOK, "Property updated", prop, nil)
}

// UploadImages handles POST /api/v1/properties/:id/images
func (h *PropertyHandler) UploadImages(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, 400, "BAD_REQUEST", "Invalid property ID", nil)
		return
	}

	ownerID, _ := uuid.Parse(middleware.GetUserID(c))
	isAdmin := middleware.GetUserRole(c) == "admin"

	file, header, err := c.Request.FormFile("image")
	if err != nil {
		response.Error(c, 400, "BAD_REQUEST", "Image file is required", nil)
		return
	}
	defer file.Close()

	img, err := h.svc.UploadImage(c.Request.Context(), id, ownerID, header.Filename, file, header.Header.Get("Content-Type"), isAdmin)
	if err != nil {
		handleAppError(c, err, &h.log)
		return
	}

	response.Success(c, http.StatusCreated, "Image uploaded", img, nil)
}

// DeleteImage handles DELETE /api/v1/properties/:id/images/:imageId
func (h *PropertyHandler) DeleteImage(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, 400, "BAD_REQUEST", "Invalid property ID", nil)
		return
	}

	imageID := c.Param("imageId")
	ownerID, _ := uuid.Parse(middleware.GetUserID(c))
	isAdmin := middleware.GetUserRole(c) == "admin"

	if err := h.svc.DeleteImage(c.Request.Context(), id, imageID, ownerID, isAdmin); err != nil {
		handleAppError(c, err, &h.log)
		return
	}

	response.Success(c, http.StatusOK, "Image deleted", nil, nil)
}

// UpdateStatus handles PATCH /api/v1/properties/:id/status
func (h *PropertyHandler) UpdateStatus(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, 400, "BAD_REQUEST", "Invalid property ID", nil)
		return
	}

	var req dto.UpdatePropertyStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 422, "VALIDATION_ERROR", "Invalid request body", parseValidationErrors(err))
		return
	}

	if err := h.svc.UpdateStatus(c.Request.Context(), id, req.Status); err != nil {
		handleAppError(c, err, &h.log)
		return
	}

	response.Success(c, http.StatusOK, "Property status updated", nil, nil)
}
