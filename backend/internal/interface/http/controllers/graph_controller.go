package controllers

import (
	"net/http"
	"strconv"

	"github.com/K8S-Graph-Explorer/backend/internal/services"
	"github.com/gin-gonic/gin"
)

// GraphController handles graph-related HTTP requests
type GraphController struct {
	graphService *services.GraphService
}

// NewGraphController creates a new GraphController
func NewGraphController(graphService *services.GraphService) *GraphController {
	return &GraphController{
		graphService: graphService,
	}
}

// GetGraph retrieves the complete graph
// @Summary      Get graph
// @Description  Retrieves the complete resource topology graph
// @Tags         Graph
// @Produce      json
// @Param        namespace  query  string  false  "Namespace filter"
// @Success      200  {object}  models.Graph
// @Failure      500  {object}  ErrorResponse
// @Router       /api/v1/graph [get]
func (c *GraphController) GetGraph(ctx *gin.Context) {
	namespace := ctx.Query("namespace")

	graph, err := c.graphService.GetGraph(ctx.Request.Context(), namespace)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_error",
			Message: err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, graph)
}

// GetResourceTopology retrieves a subgraph for a specific resource
// @Summary      Get resource topology
// @Description  Retrieves the topology graph for a specific resource
// @Tags         Graph
// @Produce      json
// @Param        uid    path   string  true   "Resource UID"
// @Param        depth  query  int     false  "Graph depth (default: 3, max: 10)"
// @Success      200  {object}  models.Graph
// @Failure      400  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /api/v1/graph/topology/{uid} [get]
func (c *GraphController) GetResourceTopology(ctx *gin.Context) {
	uid := ctx.Param("uid")
	if uid == "" {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "bad_request",
			Message: "Resource UID is required",
		})
		return
	}

	depth := 3
	if depthStr := ctx.Query("depth"); depthStr != "" {
		if d, err := strconv.Atoi(depthStr); err == nil {
			depth = d
		}
	}

	graph, err := c.graphService.GetResourceTopology(ctx.Request.Context(), uid, depth)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_error",
			Message: err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, graph)
}
