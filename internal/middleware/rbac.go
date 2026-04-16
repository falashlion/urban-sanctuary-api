package middleware

import (
	"github.com/falashlion/urban-sanctuary-api/pkg/response"
	"github.com/gin-gonic/gin"
)

// RequireRole checks that the authenticated user has one of the specified roles.
func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole := GetUserRole(c)
		if userRole == "" {
			response.Error(c, 401, "UNAUTHORIZED", "Authentication required", nil)
			return
		}

		for _, role := range roles {
			if userRole == role {
				c.Next()
				return
			}
		}

		response.Error(c, 403, "FORBIDDEN", "You do not have permission to perform this action", nil)
	}
}

// RequireAdmin checks that the authenticated user is an admin.
func RequireAdmin() gin.HandlerFunc {
	return RequireRole("admin")
}

// RequireOwnerOrAdmin checks that the user is a homeowner or admin.
func RequireOwnerOrAdmin() gin.HandlerFunc {
	return RequireRole("homeowner", "admin")
}
