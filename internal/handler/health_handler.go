package handler

import (
	"net/http"

	"github.com/falashlion/urban-sanctuary-api/internal/platform/cache"
	"github.com/falashlion/urban-sanctuary-api/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

// HealthHandler handles health check endpoints.
type HealthHandler struct {
	db    *pgxpool.Pool
	redis *cache.RedisClient
}

// NewHealthHandler creates a new HealthHandler.
func NewHealthHandler(db *pgxpool.Pool, redis *cache.RedisClient) *HealthHandler {
	return &HealthHandler{db: db, redis: redis}
}

// Health handles GET /api/v1/health
func (h *HealthHandler) Health(c *gin.Context) {
	status := "healthy"
	checks := gin.H{}

	// Check database
	if err := h.db.Ping(c.Request.Context()); err != nil {
		status = "degraded"
		checks["database"] = "unhealthy"
	} else {
		checks["database"] = "healthy"
	}

	// Check Redis
	if h.redis != nil {
		if err := h.redis.Health(c.Request.Context()); err != nil {
			status = "degraded"
			checks["redis"] = "unhealthy"
		} else {
			checks["redis"] = "healthy"
		}
	}

	httpStatus := http.StatusOK
	if status != "healthy" {
		httpStatus = http.StatusServiceUnavailable
	}

	response.Success(c, httpStatus, status, gin.H{
		"status": status,
		"checks": checks,
	}, nil)
}
