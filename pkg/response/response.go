package response

import (
	"github.com/gin-gonic/gin"
)

// Meta contains pagination metadata.
type Meta struct {
	Page       int   `json:"page,omitempty"`
	PerPage    int   `json:"per_page,omitempty"`
	Total      int64 `json:"total,omitempty"`
	TotalPages int   `json:"total_pages,omitempty"`
}

// ErrorDetail represents a field-level error detail.
type ErrorDetail struct {
	Field   string `json:"field,omitempty"`
	Message string `json:"message"`
}

// Success sends a successful JSON response with the standard envelope.
func Success(c *gin.Context, status int, msg string, data any, meta *Meta) {
	resp := gin.H{
		"success":    true,
		"status":     status,
		"message":    msg,
		"data":       data,
		"request_id": c.GetString("request_id"),
	}
	if meta != nil {
		resp["meta"] = meta
	}
	c.JSON(status, resp)
}

// Error sends an error JSON response with the standard envelope.
func Error(c *gin.Context, status int, code, msg string, details []ErrorDetail) {
	errBody := gin.H{
		"code":    code,
		"message": msg,
	}
	if len(details) > 0 {
		errBody["details"] = details
	}

	c.AbortWithStatusJSON(status, gin.H{
		"success":    false,
		"status":     status,
		"error":      errBody,
		"request_id": c.GetString("request_id"),
	})
}

// PaginationMeta creates a Meta struct from pagination parameters.
func PaginationMeta(page, perPage int, total int64) *Meta {
	totalPages := int(total) / perPage
	if int(total)%perPage > 0 {
		totalPages++
	}
	return &Meta{
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: totalPages,
	}
}
