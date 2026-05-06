package router

import (
	"github.com/K8S-Graph-Explorer/backend/internal/config"
	"github.com/K8S-Graph-Explorer/backend/internal/infrastructure/database"
	"github.com/K8S-Graph-Explorer/backend/internal/infrastructure/kubernetes"
	"github.com/K8S-Graph-Explorer/backend/internal/infrastructure/logger"
	"github.com/K8S-Graph-Explorer/backend/internal/interface/http/controllers"
	"github.com/K8S-Graph-Explorer/backend/internal/interface/http/middleware"
	"github.com/K8S-Graph-Explorer/backend/internal/interface/http/websocket"
	"github.com/K8S-Graph-Explorer/backend/internal/repositories/neo4j"
	"github.com/K8S-Graph-Explorer/backend/internal/services"
	"github.com/K8S-Graph-Explorer/backend/internal/services/converter"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// NewRouter creates and configures the Gin router
func NewRouter(
	cfg *config.Config,
	log *logger.Logger,
	neo4jClient *database.Neo4jClient,
	k8sClient *kubernetes.Client,
	wsHub *websocket.Hub,
) *gin.Engine {
	// Set Gin mode
	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// Global middleware
	router.Use(gin.Recovery())
	router.Use(middleware.Logger(log))
	router.Use(middleware.RequestID())

	// CORS configuration
	corsConfig := cors.Config{
		AllowOrigins:     cfg.Server.AllowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Request-ID"},
		ExposeHeaders:    []string{"Content-Length", "X-Request-ID"},
		AllowCredentials: true,
	}
	router.Use(cors.New(corsConfig))

	// Initialize repositories
	resourceRepo := neo4j.NewResourceRepository(neo4jClient)
	graphRepo := neo4j.NewGraphRepository(neo4jClient)

	// Initialize services
	resourceService := services.NewResourceService(resourceRepo)
	graphService := services.NewGraphService(graphRepo)
	syncService := services.NewSyncService(k8sClient, resourceRepo, wsHub)
	graphConverter := converter.NewGraphConverter(k8sClient)

	// Initialize controllers
	healthController := controllers.NewHealthController()
	resourceController := controllers.NewResourceController(resourceService)
	graphController := controllers.NewGraphController(graphService)
	syncController := controllers.NewSyncController(syncService)
	schemaController := controllers.NewSchemaController(graphConverter)

	// Health check routes (no auth required)
	router.GET("/health", healthController.Health)
	router.GET("/ready", healthController.Ready)
	router.GET("/live", healthController.Live)

	// WebSocket route
	router.GET("/ws", func(c *gin.Context) {
		websocket.ServeWs(wsHub, c)
	})

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Resource routes
		resources := v1.Group("/resources")
		{
			resources.GET("", resourceController.ListResources)
			resources.GET("/search", resourceController.SearchResources)
			resources.GET("/:uid", resourceController.GetResource)
			resources.GET("/:uid/related", resourceController.GetRelatedResources)
		}

		// Graph routes
		graph := v1.Group("/graph")
		{
			graph.GET("", graphController.GetGraph)
			graph.GET("/topology/:uid", graphController.GetResourceTopology)
		}

		// Schema routes (normalized graph)
		schema := v1.Group("/schema")
		{
			schema.GET("", schemaController.GetGraphSchema)
			schema.GET("/json", schemaController.GetGraphSchemaJSON)
		}

		// Sync routes
		sync := v1.Group("/sync")
		{
			sync.POST("", syncController.SyncNamespace)
			sync.POST("/refresh", syncController.RefreshGraph)
		}
	}

	return router
}
