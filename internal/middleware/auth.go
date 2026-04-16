package middleware

import (
	"strings"

	"github.com/falashlion/urban-sanctuary-api/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// JWTAuth validates Bearer tokens and extracts claims into context.
func JWTAuth(accessSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Error(c, 401, "UNAUTHORIZED", "Missing authorization header", nil)
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			response.Error(c, 401, "UNAUTHORIZED", "Invalid authorization header format", nil)
			return
		}

		tokenString := parts[1]

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(accessSecret), nil
		})

		if err != nil || !token.Valid {
			response.Error(c, 401, "UNAUTHORIZED", "Invalid or expired token", nil)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			response.Error(c, 401, "UNAUTHORIZED", "Invalid token claims", nil)
			return
		}

		// Extract and set claims in context
		if sub, ok := claims["sub"].(string); ok {
			c.Set("user_id", sub)
		}
		if role, ok := claims["role"].(string); ok {
			c.Set("user_role", role)
		}
		if email, ok := claims["email"].(string); ok {
			c.Set("user_email", email)
		}
		if verified, ok := claims["verified"].(bool); ok {
			c.Set("user_verified", verified)
		}

		c.Next()
	}
}

// OptionalAuth tries to validate a Bearer token but doesn't require it.
func OptionalAuth(accessSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.Next()
			return
		}

		token, err := jwt.Parse(parts[1], func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(accessSecret), nil
		})

		if err == nil && token.Valid {
			if claims, ok := token.Claims.(jwt.MapClaims); ok {
				if sub, ok := claims["sub"].(string); ok {
					c.Set("user_id", sub)
				}
				if role, ok := claims["role"].(string); ok {
					c.Set("user_role", role)
				}
			}
		}

		c.Next()
	}
}

// GetUserID extracts the user ID from the Gin context (set by JWTAuth).
func GetUserID(c *gin.Context) string {
	return c.GetString("user_id")
}

// GetUserRole extracts the user role from the Gin context.
func GetUserRole(c *gin.Context) string {
	return c.GetString("user_role")
}
