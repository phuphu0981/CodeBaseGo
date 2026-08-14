package health

import (
	"github.com/gin-gonic/gin"
	"github.com/google/wire"

	"codebasego/internal/common"
)

var _ common.RouteRegistrar = (*Module)(nil)

// ProviderSet is the Wire provider set for the Health module.
var ProviderSet = wire.NewSet(
	NewHandler,
	NewModule,
)

type Module struct {
	handler *Handler
}

func NewModule(h *Handler) *Module {
	return &Module{handler: h}
}

// RegisterRoutes registers HTTP routes for the Health module.
func (m *Module) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("/health", m.handler.HealthCheck)
	rg.GET("/health/live", m.handler.Liveness)
	rg.GET("/health/ready", m.handler.Readiness)
	rg.GET("/livez", m.handler.Liveness)
	rg.GET("/readyz", m.handler.Readiness)
}
