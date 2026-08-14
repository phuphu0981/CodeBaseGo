package health

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	platformRedis "codebasego/internal/platform/redis"
	"codebasego/internal/platform/response"
)

type Handler struct {
	db  *gorm.DB
	rdb *platformRedis.Client
}

func NewHandler(db *gorm.DB, rdb *platformRedis.Client) *Handler {
	return &Handler{
		db:  db,
		rdb: rdb,
	}
}

// HealthCheck returns full server health status with DB & Redis connectivity.
//
//	@Summary	Comprehensive health check
//	@Tags		health
//	@Produce	json
//	@Success	200	{object}	response.Body
//	@Failure	503	{object}	response.Body
//	@Router		/health [get]
func (h *Handler) HealthCheck(c *gin.Context) {
	details := gin.H{
		"status":   "ok",
		"database": "up",
		"redis":    "disabled",
	}

	if h.db != nil {
		sqlDB, err := h.db.DB()
		if err != nil || sqlDB.PingContext(c.Request.Context()) != nil {
			response.Error(c, http.StatusServiceUnavailable, "database connection failed")
			return
		}
	}

	if h.rdb != nil && h.rdb.Client != nil {
		if err := h.rdb.Ping(c.Request.Context()).Err(); err != nil {
			response.Error(c, http.StatusServiceUnavailable, "redis connection failed")
			return
		}
		details["redis"] = "up"
	}

	response.Success(c, details)
}

// Liveness returns 200 OK if the application server process is running.
//
//	@Summary	Liveness probe (Kubernetes /load balancer)
//	@Tags		health
//	@Produce	json
//	@Success	200	{object}	response.Body
//	@Router		/health/live [get]
func (h *Handler) Liveness(c *gin.Context) {
	response.Success(c, gin.H{
		"status": "alive",
	})
}

// Readiness checks whether downstream dependencies (Database, Redis) are ready to accept traffic.
//
//	@Summary	Readiness probe (Kubernetes /load balancer)
//	@Tags		health
//	@Produce	json
//	@Success	200	{object}	response.Body
//	@Failure	503	{object}	response.Body
//	@Router		/health/ready [get]
func (h *Handler) Readiness(c *gin.Context) {
	readiness := gin.H{
		"status":   "ready",
		"database": "up",
		"redis":    "disabled",
	}

	if h.db != nil {
		sqlDB, err := h.db.DB()
		if err != nil || sqlDB.PingContext(c.Request.Context()) != nil {
			c.JSON(http.StatusServiceUnavailable, response.Body{
				Success: false,
				Error:   "database is not ready",
				Data:    gin.H{"status": "not_ready", "database": "down"},
			})
			return
		}
	}

	if h.rdb != nil && h.rdb.Client != nil {
		if err := h.rdb.Ping(c.Request.Context()).Err(); err != nil {
			c.JSON(http.StatusServiceUnavailable, response.Body{
				Success: false,
				Error:   "redis is not ready",
				Data:    gin.H{"status": "not_ready", "redis": "down"},
			})
			return
		}
		readiness["redis"] = "up"
	}

	response.Success(c, readiness)
}
