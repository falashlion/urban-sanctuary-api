package handler

import (
	"errors"
	"reflect"
	"strings"

	"github.com/falashlion/urban-sanctuary-api/internal/domain"
	"github.com/falashlion/urban-sanctuary-api/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog"
)

// parseValidationErrors converts go-playground validator errors into ErrorDetail slices.
func parseValidationErrors(err error) []response.ErrorDetail {
	var ve validator.ValidationErrors
	if !errors.As(err, &ve) {
		return []response.ErrorDetail{{Message: err.Error()}}
	}

	details := make([]response.ErrorDetail, 0, len(ve))
	for _, fe := range ve {
		field := toSnakeCase(fe.Field())
		msg := formatValidationMessage(fe)
		details = append(details, response.ErrorDetail{
			Field:   field,
			Message: msg,
		})
	}
	return details
}

// handleAppError unwraps domain.AppError and sends the appropriate error response.
func handleAppError(c *gin.Context, err error, log *zerolog.Logger) {
	var appErr *domain.AppError
	if errors.As(err, &appErr) {
		var details []response.ErrorDetail
		for _, d := range appErr.Details {
			details = append(details, response.ErrorDetail{
				Field:   d.Field,
				Message: d.Message,
			})
		}
		response.Error(c, appErr.StatusCode, appErr.Code, appErr.Message, details)
		return
	}

	log.Error().Err(err).Msg("unexpected error")
	response.Error(c, 500, "INTERNAL_ERROR", "An unexpected error occurred", nil)
}

func formatValidationMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "this field is required"
	case "email":
		return "must be a valid email address"
	case "min":
		return "minimum " + fe.Param() + " characters required"
	case "max":
		return "maximum " + fe.Param() + " characters allowed"
	case "len":
		return "must be exactly " + fe.Param() + " characters"
	case "gte":
		return "must be at least " + fe.Param()
	case "lte":
		return "must be at most " + fe.Param()
	case "gt":
		return "must be greater than " + fe.Param()
	case "oneof":
		return "must be one of: " + fe.Param()
	case "uuid":
		return "must be a valid UUID"
	case "url":
		return "must be a valid URL"
	case "e164":
		return "must be a valid phone number in E.164 format"
	case "latitude":
		return "must be a valid latitude"
	case "longitude":
		return "must be a valid longitude"
	default:
		return "validation failed on " + fe.Tag()
	}
}

func toSnakeCase(s string) string {
	var result []byte
	for i, c := range s {
		if c >= 'A' && c <= 'Z' {
			if i > 0 {
				result = append(result, '_')
			}
			result = append(result, byte(c+32))
		} else {
			result = append(result, byte(c))
		}
	}
	return string(result)
}

// Suppress unused imports
var _ = reflect.TypeOf
var _ = strings.TrimSpace
