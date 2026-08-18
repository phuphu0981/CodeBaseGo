package main

import (
	"context"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/gorm"

	_ "codebasego/docs"
	"codebasego/internal/common"
	"codebasego/internal/platform/eventbus"
	"codebasego/internal/platform/server"
)

//	@title						CodebaseGo API
//	@version					1.0
//	@description				Modular Go + Gin Backend API (REST & GraphQL)
//	@host						localhost:8080
//	@BasePath					/api/v1
//	@securityDefinitions.apikey	BearerAuth
//	@in							header
//	@name						Authorization

func main() {
	app, err := InitializeApp()
	if err != nil {
		log.Fatal("failed to initialize app: ", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	// Start all active module background workers (including Transactional Outbox)
	for _, worker := range app.Workers {
		worker.StartBackground(ctx, &wg)
	}

	defer func() {
		if app.DB != nil {
			if sqlDB, err := app.DB.DB(); err == nil {
				log.Println("closing database connections...")
				_ = sqlDB.Close()
			}
		}
	}()

	if err := app.Server.Run(); err != nil {
		log.Println("failed to run server: ", err)
	}

	// Graceful shutdown: cancel worker contexts and wait for active jobs to complete
	cancel()
	log.Println("waiting for background workers to shutdown gracefully...")
	workerDone := make(chan struct{})
	go func() {
		wg.Wait()
		close(workerDone)
	}()

	select {
	case <-workerDone:
		log.Println("all background workers stopped cleanly")
	case <-time.After(5 * time.Second):
		log.Println("background workers shutdown timed out")
	}
}

// App is the top-level composition root. NewApp wires all modules to the server.
type App struct {
	DB       *gorm.DB
	Server   *server.Server
	Workers  []common.BackgroundWorker
	EventBus common.EventBus
}

func NewApp(
	db *gorm.DB,
	srv *server.Server,
	outbox *eventbus.OutboxProcessor,
	migrators []common.Migrator,
	registrars []common.RouteRegistrar,
	workers []common.BackgroundWorker,
	bus common.EventBus,
) (*App, error) {
	// Database migration at startup (if enabled in config)
	if srv.Config.DB.AutoMigrate {
		log.Println("running database auto-migrations (Dev/Test mode)...")
		if outbox != nil {
			if err := outbox.AutoMigrate(db); err != nil {
				return nil, err
			}
		}
		for _, m := range migrators {
			if err := m.AutoMigrate(db); err != nil {
				return nil, err
			}
		}
	} else {
		log.Println("database auto-migration skipped (Staging/Production SQL migrations mode)")
	}

	// Include outbox processor in background workers
	allWorkers := workers
	if outbox != nil {
		allWorkers = append([]common.BackgroundWorker{outbox}, workers...)
	}

	// Swagger documentation (enabled in non-production mode)
	if srv.Config.Server.Mode != "release" && srv.Config.Server.Mode != "production" {
		srv.Engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	// Root endpoint
	srv.Engine.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"message": "CodebaseGo API is running",
			"docs":    "/swagger/index.html",
			"health":  "/api/v1/health",
		})
	})

	v1 := srv.Engine.Group("/api/v1")
	for _, r := range registrars {
		r.RegisterRoutes(v1)
	}

	return &App{
		DB:       db,
		Server:   srv,
		Workers:  allWorkers,
		EventBus: bus,
	}, nil
}
