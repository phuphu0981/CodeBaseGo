package setting

import (
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	"gorm.io/gorm"

	"codebasego/internal/common"
)

var (
	_ common.RouteRegistrar = (*Module)(nil)
	_ common.Migrator       = (*Module)(nil)
)

// ProviderSet is the Wire provider set for the Setting module.
var ProviderSet = wire.NewSet(
	NewGormRepository,
	wire.Bind(new(Repository), new(*GormRepository)),
	NewService,
	NewHandler,
	NewModule,
)

type Module struct {
	handler        *Handler
	authMiddleware common.AuthMiddleware
}

func NewModule(h *Handler, authMiddleware common.AuthMiddleware) *Module {
	return &Module{
		handler:        h,
		authMiddleware: authMiddleware,
	}
}

// AutoMigrate performs database migrations for the Setting module entities.
func (m *Module) AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(&CoreConfig{})
}

// RegisterRoutes registers HTTP routes for the Setting module under /settings.
func (m *Module) RegisterRoutes(rg *gin.RouterGroup) {
	group := rg.Group("/settings")

	// Public routes
	group.GET("/public", m.handler.GetPublic)
	group.GET("/by-path", m.handler.GetByPath)

	// Protected management routes (Admin only)
	protected := group.Group("", gin.HandlerFunc(m.authMiddleware), common.RequireRole("admin"))
	protected.GET("", m.handler.List)
	protected.GET("/:id", m.handler.GetByID)
	protected.POST("", m.handler.Set)
	protected.DELETE("/:id", m.handler.Delete)
}
