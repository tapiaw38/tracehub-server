package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"

	"github.com/tapiaw38/tracehub-server/internal/adapters/datasources/repositories/project"
	"github.com/tapiaw38/tracehub-server/internal/adapters/datasources/repositories/trace"
	"github.com/tapiaw38/tracehub-server/internal/adapters/web/handlers/health"
	projectHandler "github.com/tapiaw38/tracehub-server/internal/adapters/web/handlers/project"
	traceHandler "github.com/tapiaw38/tracehub-server/internal/adapters/web/handlers/trace"
	"github.com/tapiaw38/tracehub-server/internal/adapters/web/middlewares"
	"github.com/tapiaw38/tracehub-server/internal/platform/config"
	"github.com/tapiaw38/tracehub-server/internal/platform/database"
	projectUsecase "github.com/tapiaw38/tracehub-server/internal/usecases/project"
	traceUsecase "github.com/tapiaw38/tracehub-server/internal/usecases/trace"
)

func main() {
	// Load configuration
	if err := config.Load(); err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	cfg := config.Get()
	log.Println("✓ Configuration loaded")

	// Connect to database
	db, err := database.Connect()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	// Run migrations
	if err := runMigrations(db); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}
	log.Println("✓ Migrations completed")

	// Initialize repositories
	projectRepo := project.NewRepository(db)
	traceRepo := trace.NewRepository(db)

	// Initialize use cases
	createProjectUC := projectUsecase.NewCreateUsecase(projectRepo)
	getProjectUC := projectUsecase.NewGetUsecase(projectRepo)
	listProjectsUC := projectUsecase.NewListUsecase(projectRepo)

	ingestTraceUC := traceUsecase.NewIngestUsecase(traceRepo, projectRepo)
	queryTracesUC := traceUsecase.NewQueryUsecase(traceRepo)

	// Initialize handlers
	healthHandler := health.NewHandler()
	projectCreateHandler := projectHandler.NewCreateHandler(createProjectUC)
	projectGetHandler := projectHandler.NewGetHandler(getProjectUC)
	projectListHandler := projectHandler.NewListHandler(listProjectsUC)
	traceIngestHandler := traceHandler.NewIngestHandler(ingestTraceUC)
	traceQueryHandler := traceHandler.NewQueryHandler(queryTracesUC)

	// Setup Gin
	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	router := gin.Default()

	// CORS middleware
	router.Use(cors.New(cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// Routes
	api := router.Group("/api/v1")
	{
		// Health check
		api.GET("/health", healthHandler.Handle)

		// Public routes - Project management (for now, no auth required for demo)
		api.POST("/projects", projectCreateHandler.Handle)
		api.GET("/projects", projectListHandler.Handle)
		api.GET("/projects/:id", projectGetHandler.Handle)

		// Protected routes - require API key
		authenticated := api.Group("")
		authenticated.Use(middlewares.ApiKeyAuth(projectRepo))
		{
			// Trace ingestion
			authenticated.POST("/traces", traceIngestHandler.Handle)
			authenticated.POST("/traces/batch", traceIngestHandler.HandleBatch)

			// Trace querying
			authenticated.GET("/traces", traceQueryHandler.Handle)
		}
	}

	// Start server
	addr := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)
	log.Printf("🚀 TraceHub Server started on %s", addr)

	// Graceful shutdown
	go func() {
		if err := router.Run(addr); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
}

func runMigrations(db *sql.DB) error {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create migration driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"postgres",
		driver,
	)
	if err != nil {
		return fmt.Errorf("failed to create migration instance: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}
