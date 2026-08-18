package handler

import (
	"log"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "codebasego/docs"
	"codebasego/internal/modules/auth"
	"codebasego/internal/modules/graphql"
	"codebasego/internal/modules/health"
	"codebasego/internal/modules/page"
	"codebasego/internal/modules/seo"
	"codebasego/internal/modules/setting"
	"codebasego/internal/modules/user"
	"codebasego/internal/platform/config"
	"codebasego/internal/platform/database"
	"codebasego/internal/platform/eventbus"
	"codebasego/internal/platform/logger"
	"codebasego/internal/platform/redis"
	"codebasego/internal/platform/server"
)

var (
	srv     *server.Server
	initErr error
	once    sync.Once
)

func initServer() {
	cfg, err := config.NewConfig()
	if err != nil {
		initErr = err
		log.Printf("failed to load config: %v", err)
		return
	}

	db, err := database.NewDB(cfg)
	if err != nil {
		initErr = err
		log.Printf("failed to connect database: %v", err)
		return
	}

	appLogger := logger.NewLogger(cfg)
	rdb, err := redis.NewClient(cfg, appLogger)
	if err != nil {
		initErr = err
		log.Printf("failed to connect redis: %v", err)
		return
	}

	srv = server.NewServer(cfg, appLogger, rdb)
	bus := eventbus.NewEventBus(cfg, rdb, appLogger)
	outbox := eventbus.NewOutboxProcessor(db, bus, appLogger)

	userRepo := user.NewGormRepository(db)
	userService := user.NewService(userRepo)

	authRepo := auth.NewGormRefreshTokenRepository(db)
	authService := auth.NewService(cfg, userService, authRepo, bus)
	authHandler := auth.NewHandler(authService)
	authMod := auth.NewModule(authHandler, authService)

	userHandler := user.NewHandler(userService)
	authMiddleware := auth.ProvideAuthMiddleware(authMod)
	userMod := user.NewModule(userHandler, authMiddleware)

	gqlResolver := graphql.NewResolver(userService, authService)
	gqlMod := graphql.NewModule(gqlResolver, cfg, authService)

	healthHandler := health.NewHandler(db, rdb)
	healthMod := health.NewModule(healthHandler)

	pageRepo := page.NewGormRepository(db)
	pageService := page.NewService(pageRepo)
	pageHandler := page.NewHandler(pageService)
	pageMod := page.NewModule(pageHandler, authMiddleware)

	seoRepo := seo.NewGormRepository(db)
	seoService := seo.NewService(seoRepo)
	seoHandler := seo.NewHandler(seoService)
	seoMod := seo.NewModule(seoHandler, authMiddleware)

	settingRepo := setting.NewGormRepository(db)
	settingService := setting.NewService(settingRepo)
	settingHandler := setting.NewHandler(settingService)
	settingMod := setting.NewModule(settingHandler, authMiddleware)

	if cfg.DB.AutoMigrate {
		if outbox != nil {
			_ = outbox.AutoMigrate(db)
		}
		_ = authMod.AutoMigrate(db)
		_ = userMod.AutoMigrate(db)
		_ = pageMod.AutoMigrate(db)
		_ = seoMod.AutoMigrate(db)
		_ = settingMod.AutoMigrate(db)
	}

	if cfg.Server.Mode != "release" && cfg.Server.Mode != "production" {
		srv.Engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	srv.Engine.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"message": "CodebaseGo API Serverless is running",
			"docs":    "/swagger/index.html",
			"health":  "/api/v1/health",
		})
	})

	v1 := srv.Engine.Group("/api/v1")
	authMod.RegisterRoutes(v1)
	gqlMod.RegisterRoutes(v1)
	healthMod.RegisterRoutes(v1)
	userMod.RegisterRoutes(v1)
	pageMod.RegisterRoutes(v1)
	seoMod.RegisterRoutes(v1)
	settingMod.RegisterRoutes(v1)
}

// Handler is the entrypoint for Vercel Serverless Functions
func Handler(w http.ResponseWriter, r *http.Request) {
	once.Do(initServer)

	if initErr != nil {
		http.Error(w, "Failed to initialize serverless application: "+initErr.Error(), http.StatusInternalServerError)
		return
	}

	if srv == nil || srv.Engine == nil {
		http.Error(w, "Server engine not initialized", http.StatusInternalServerError)
		return
	}

	srv.Engine.ServeHTTP(w, r)
}
