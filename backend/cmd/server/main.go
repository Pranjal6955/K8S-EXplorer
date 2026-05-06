package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/K8S-Graph-Explorer/backend/internal/config"
	"github.com/K8S-Graph-Explorer/backend/internal/infrastructure/database"
	"github.com/K8S-Graph-Explorer/backend/internal/infrastructure/kubernetes"
	"github.com/K8S-Graph-Explorer/backend/internal/infrastructure/logger"
	"github.com/K8S-Graph-Explorer/backend/internal/infrastructure/watcher"
	"github.com/K8S-Graph-Explorer/backend/internal/interface/http/router"
	"github.com/K8S-Graph-Explorer/backend/internal/interface/http/websocket"
	neo4jrepo "github.com/K8S-Graph-Explorer/backend/internal/repositories/neo4j"
	"github.com/K8S-Graph-Explorer/backend/internal/services"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize logger
	appLogger, err := logger.NewLogger(cfg.Server.Environment)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer appLogger.Sync()

	appLogger.Info("Starting K8S Graph Explorer Backend...")

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize Neo4j connection
	neo4jClient, err := database.NewNeo4jClient(cfg.Neo4j)
	if err != nil {
		appLogger.Fatal("Failed to connect to Neo4j", "error", err)
	}
	defer neo4jClient.Close()
	appLogger.Info("Connected to Neo4j successfully")

	// Initialize Kubernetes client
	k8sClient, err := kubernetes.NewClient(cfg.Kubernetes)
	if err != nil {
		appLogger.Fatal("Failed to create Kubernetes client", "error", err)
	}
	appLogger.Info("Kubernetes client initialized")

	// Initialize WebSocket Hub
	wsHub := websocket.NewHub()
	go wsHub.Run()
	
	// Initialize repositories
	resourceRepo := neo4jrepo.NewResourceRepository(neo4jClient)

	// Initialize Kubernetes watcher
	k8sWatcher, err := watcher.NewWatcher(cfg.Kubernetes, appLogger)
	if err != nil {
		appLogger.Warn("Failed to create Kubernetes watcher, real-time updates disabled", "error", err)
	} else {
		// Create and start event processor
		eventProcessor := watcher.NewEventProcessor(k8sWatcher, resourceRepo, wsHub, appLogger, 5)
		eventProcessor.Start(ctx)

		// Add logging handler for debugging
		k8sWatcher.AddEventHandler(watcher.LoggingEventHandler(appLogger))

		// Start watcher in background
		go func() {
			if err := k8sWatcher.Start(ctx); err != nil {
				appLogger.Error("Kubernetes watcher error", "error", err)
			}
		}()
		appLogger.Info("Kubernetes resource watcher started")
	}

	// Initialize SyncService
	syncService := services.NewSyncService(k8sClient, resourceRepo, wsHub)

	// Perform initial sync in background
	go func() {
		appLogger.Info("Performing initial cluster synchronization...")
		res, err := syncService.SyncAllNamespaces(ctx)
		if err != nil {
			appLogger.Error("Initial sync failed", "error", err)
		} else {
			appLogger.Info("Initial sync completed", "resources", res.SyncedCount)
		}
	}()

	// Setup router
	r := router.NewRouter(cfg, appLogger, neo4jClient, k8sClient, wsHub, syncService)

	// Create HTTP server
	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		appLogger.Info("Server starting", "port", cfg.Server.Port, "environment", cfg.Server.Environment)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			appLogger.Fatal("Server failed to start", "error", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	appLogger.Info("Shutting down server...")

	// Cancel context to stop watcher and other background tasks
	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		appLogger.Fatal("Server forced to shutdown", "error", err)
	}

	appLogger.Info("Server exited gracefully")
}
