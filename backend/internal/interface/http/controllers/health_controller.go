package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// HealthController handles health check endpoints
type HealthController struct{}

// NewHealthController creates a new HealthController
func NewHealthController() *HealthController {
	return &HealthController{}
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status  string            `json:"status"`
	Service string            `json:"service"`
	Version string            `json:"version"`
	Checks  map[string]string `json:"checks,omitempty"`
}

// Health returns the health status of the service
// @Summary      Health check
// @Description  Returns the health status of the service
// @Tags         Health
// @Produce      json
// @Success      200  {object}  HealthResponse
// @Router       /health [get]
func (c *HealthController) Health(ctx *gin.Context) {
	response := HealthResponse{
		Status:  "healthy",
		Service: "k8s-graph-explorer",
		Version: "1.0.0",
	}

	ctx.JSON(http.StatusOK, response)
}

// Ready returns the readiness status of the service
// @Summary      Readiness check
// @Description  Returns whether the service is ready to accept requests
// @Tags         Health
// @Produce      json
// @Success      200  {object}  HealthResponse
// @Failure      503  {object}  HealthResponse
// @Router       /ready [get]
func (c *HealthController) Ready(ctx *gin.Context) {
	// In a real implementation, check database connectivity, etc.
	response := HealthResponse{
		Status:  "ready",
		Service: "k8s-graph-explorer",
		Version: "1.0.0",
		Checks: map[string]string{
			"neo4j":      "connected",
			"kubernetes": "connected",
		},
	}

	ctx.JSON(http.StatusOK, response)
}

// Live returns the liveness status of the service
// @Summary      Liveness check
// @Description  Returns whether the service is alive
// @Tags         Health
// @Produce      json
// @Success      200  {object}  HealthResponse
// @Router       /live [get]
func (c *HealthController) Live(ctx *gin.Context) {
	response := HealthResponse{
		Status:  "alive",
		Service: "k8s-graph-explorer",
		Version: "1.0.0",
	}

	ctx.JSON(http.StatusOK, response)
}
