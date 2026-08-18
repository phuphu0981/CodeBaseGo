package seo

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

// ProviderSet is the Wire provider set for the SEO module.
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

// AutoMigrate performs database migrations for the SEO module entities.
func (m *Module) AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(&SEO{})
}

// RegisterRoutes registers HTTP routes for the SEO module under /seo.
func (m *Module) RegisterRoutes(rg *gin.RouterGroup) {
	group := rg.Group("/seo")
	group.GET("/by-slug", m.handler.GetBySlug)
	group.GET("", m.handler.List)
	group.GET("/:id", m.handler.GetByID)

	protected := group.Group("", gin.HandlerFunc(m.authMiddleware))
	protected.POST("", m.handler.Create)
	protected.PUT("/:id", m.handler.Update)
	protected.DELETE("/:id", m.handler.Delete)
}
