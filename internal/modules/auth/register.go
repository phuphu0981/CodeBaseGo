package auth

import (
	"context"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	"gorm.io/gorm"

	"codebasego/internal/common"
)

var (
	_ common.RouteRegistrar   = (*Module)(nil)
	_ common.Migrator         = (*Module)(nil)
	_ common.BackgroundWorker = (*Module)(nil)
)

// @wire:bind target=UserService source=*user.Service
// ProviderSet is the Wire provider set for the Auth module.
var ProviderSet = wire.NewSet(
	NewGormRefreshTokenRepository,
	wire.Bind(new(RefreshTokenRepository), new(*GormRefreshTokenRepository)),
	NewService,
	NewHandler,
	NewModule,
	ProvideAuthMiddleware,
)

// ProvideAuthMiddleware provides common.AuthMiddleware authentication middleware for Wire DI.
func ProvideAuthMiddleware(m *Module) common.AuthMiddleware {
	return common.AuthMiddleware(m.AuthMiddleware())
}

type Module struct {
	handler *Handler
	service *Service
}

func NewModule(h *Handler, s *Service) *Module {
	return &Module{handler: h, service: s}
}

// AutoMigrate performs database migrations for the Auth module entities.
func (m *Module) AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(&RefreshToken{})
}

// RegisterRoutes registers HTTP routes for the Auth module under /auth.
func (m *Module) RegisterRoutes(rg *gin.RouterGroup) {
	group := rg.Group("/auth")
	group.POST("/register", m.handler.Register)
	group.POST("/login", m.handler.Login)
	group.POST("/refresh", m.handler.Refresh)
	group.POST("/logout", m.handler.Logout)
	group.POST("/logout-all", m.AuthMiddleware(), m.handler.LogoutAll)
}

// StartBackground runs periodic background cleanup tasks with graceful shutdown tracking.
func (m *Module) StartBackground(ctx context.Context, wg *sync.WaitGroup) {
	m.StartTokenCleanup(ctx, wg, 24*time.Hour)
}

func (m *Module) StartTokenCleanup(ctx context.Context, wg *sync.WaitGroup, interval time.Duration) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				_, _ = m.service.DeleteExpiredTokens(ctx)
			}
		}
	}()
}
