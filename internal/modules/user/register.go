package user

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

// ProviderSet is the Wire provider set for the User module.
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

// AutoMigrate performs database migrations for the User module entities.
func (m *Module) AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(&User{})
}

// RegisterRoutes registers HTTP routes for the User module under /users.
func (m *Module) RegisterRoutes(rg *gin.RouterGroup) {
	group := rg.Group("/users", gin.HandlerFunc(m.authMiddleware))
	group.GET("", m.handler.List)
	group.GET("/:id", m.handler.GetByID)
	group.PUT("/:id", m.handler.Update)
	group.DELETE("/:id", m.handler.Delete)
}
