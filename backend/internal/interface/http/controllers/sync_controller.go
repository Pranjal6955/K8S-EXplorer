package controllers

import (
	"net/http"

	"github.com/K8S-Graph-Explorer/backend/internal/services"
	"github.com/gin-gonic/gin"
)

// SyncController handles sync-related HTTP requests
type SyncController struct {
	syncService *services.SyncService
}

// NewSyncController creates a new SyncController
func NewSyncController(syncService *services.SyncService) *SyncController {
	return &SyncController{
		syncService: syncService,
	}
}

// SyncRequest represents the request body for sync operations
type SyncRequest struct {
	Namespace string `json:"namespace,omitempty"`
}

// SyncNamespace syncs resources from a specific namespace
// @Summary      Sync namespace
// @Description  Syncs Kubernetes resources from a namespace to the graph database
// @Tags         Sync
// @Accept       json
// @Produce      json
// @Param        request  body  SyncRequest  true  "Sync request"
// @Success      200  {object}  services.SyncResult
// @Failure      400  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /api/v1/sync [post]
func (c *SyncController) SyncNamespace(ctx *gin.Context) {
	var req SyncRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		// If no body, sync all namespaces
		result, err := c.syncService.SyncAllNamespaces(ctx.Request.Context())
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "internal_error",
				Message: err.Error(),
			})
			return
		}
		ctx.JSON(http.StatusOK, result)
		return
	}

	if req.Namespace == "" {
		// Sync all namespaces
		result, err := c.syncService.SyncAllNamespaces(ctx.Request.Context())
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "internal_error",
				Message: err.Error(),
			})
			return
		}
		ctx.JSON(http.StatusOK, result)
		return
	}

	// Sync specific namespace
	result, err := c.syncService.SyncNamespace(ctx.Request.Context(), req.Namespace)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_error",
			Message: err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, result)
}

// RefreshGraph triggers a full refresh of the graph
// @Summary      Refresh graph
// @Description  Triggers a full refresh of the graph from all Kubernetes resources
// @Tags         Sync
// @Produce      json
// @Success      200  {object}  services.SyncResult
// @Failure      500  {object}  ErrorResponse
// @Router       /api/v1/sync/refresh [post]
func (c *SyncController) RefreshGraph(ctx *gin.Context) {
	result, err := c.syncService.SyncAllNamespaces(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_error",
			Message: err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, result)
}
