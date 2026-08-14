package graphql

import (
	"strings"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gin-gonic/gin"
	"github.com/google/wire"

	"codebasego/internal/common"
	"codebasego/internal/platform/config"
)

var _ common.RouteRegistrar = (*Module)(nil)

// @wire:bind target=UserService source=*user.Service
// @wire:bind target=AuthService source=*auth.Service
// ProviderSet is the Wire provider set for the GraphQL module.
var ProviderSet = wire.NewSet(
	NewResolver,
	NewModule,
)

type Module struct {
	resolver    *Resolver
	cfg         *config.Config
	authService AuthService
}

func NewModule(r *Resolver, cfg *config.Config, authService AuthService) *Module {
	return &Module{
		resolver:    r,
		cfg:         cfg,
		authService: authService,
	}
}

// RegisterRoutes registers HTTP routes for the GraphQL module (/graphql and /playground).
func (m *Module) RegisterRoutes(rg *gin.RouterGroup) {
	srv := handler.NewDefaultServer(NewExecutableSchema(Config{Resolvers: m.resolver}))
	srv.Use(extension.FixedComplexityLimit(100))

	// Playground UI (GET /playground) - only enabled in non-production mode
	if m.cfg == nil || (m.cfg.Server.Mode != "release" && m.cfg.Server.Mode != "production") {
		rg.GET("/playground", gin.WrapH(playground.Handler("GraphQL Playground", "/api/v1/graphql")))
	}

	// GraphQL API (POST /graphql)
	rg.POST("/graphql", m.optionalAuthMiddleware(), GinContextToContextMiddleware(), gin.WrapH(srv))
}

func (m *Module) optionalAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if m.authService != nil {
			header := c.GetHeader("Authorization")
			if header != "" {
				parts := strings.SplitN(header, " ", 2)
				if len(parts) == 2 && parts[0] == "Bearer" {
					claims, err := m.authService.ValidateToken(parts[1])
					if err == nil {
						c.Set("user_id", claims.UserID)
						c.Set("email", claims.Email)
					}
				}
			}
		}
		c.Next()
	}
}
