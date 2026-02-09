package controllers

import (
	"net/http"

	"github.com/K8S-Graph-Explorer/backend/internal/services"
	"github.com/gin-gonic/gin"
)

// ResourceController handles resource-related HTTP requests
type ResourceController struct {
	resourceService *services.ResourceService
}

// NewResourceController creates a new ResourceController
func NewResourceController(resourceService *services.ResourceService) *ResourceController {
	return &ResourceController{
		resourceService: resourceService,
	}
}

// ListResourcesRequest represents the query parameters for listing resources
type ListResourcesRequest struct {
	Namespace string `form:"namespace"`
	Kind      string `form:"kind"`
}

// GetResource retrieves a resource by UID
// @Summary      Get resource by UID
// @Description  Retrieves a Kubernetes resource by its unique identifier
// @Tags         Resources
// @Produce      json
// @Param        uid  path  string  true  "Resource UID"
// @Success      200  {object}  models.K8sResource
// @Failure      404  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /api/v1/resources/{uid} [get]
func (c *ResourceController) GetResource(ctx *gin.Context) {
	uid := ctx.Param("uid")
	if uid == "" {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "bad_request",
			Message: "Resource UID is required",
		})
		return
	}

	resource, err := c.resourceService.GetResource(ctx.Request.Context(), uid)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_error",
			Message: err.Error(),
		})
		return
	}

	if resource == nil {
		ctx.JSON(http.StatusNotFound, ErrorResponse{
			Error:   "not_found",
			Message: "Resource not found",
		})
		return
	}

	ctx.JSON(http.StatusOK, resource)
}

// ListResources lists resources with optional filters
// @Summary      List resources
// @Description  Lists Kubernetes resources with optional namespace and kind filters
// @Tags         Resources
// @Produce      json
// @Param        namespace  query  string  false  "Namespace filter"
// @Param        kind       query  string  false  "Kind filter"
// @Success      200  {array}   models.K8sResource
// @Failure      500  {object}  ErrorResponse
// @Router       /api/v1/resources [get]
func (c *ResourceController) ListResources(ctx *gin.Context) {
	var req ListResourcesRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "bad_request",
			Message: err.Error(),
		})
		return
	}

	resources, err := c.resourceService.ListResources(ctx.Request.Context(), req.Namespace, req.Kind)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_error",
			Message: err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, resources)
}

// SearchResources searches resources by query
// @Summary      Search resources
// @Description  Searches for resources by name or label
// @Tags         Resources
// @Produce      json
// @Param        q          query  string  true   "Search query"
// @Param        namespace  query  string  false  "Namespace filter"
// @Success      200  {array}   models.K8sResource
// @Failure      400  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /api/v1/resources/search [get]
func (c *ResourceController) SearchResources(ctx *gin.Context) {
	query := ctx.Query("q")
	if query == "" {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "bad_request",
			Message: "Search query is required",
		})
		return
	}

	namespace := ctx.Query("namespace")

	resources, err := c.resourceService.SearchResources(ctx.Request.Context(), query, namespace)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_error",
			Message: err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, resources)
}

// GetRelatedResources retrieves resources related to a given resource
// @Summary      Get related resources
// @Description  Retrieves resources related to a given resource
// @Tags         Resources
// @Produce      json
// @Param        uid  path  string  true  "Resource UID"
// @Success      200  {array}   models.K8sResource
// @Failure      500  {object}  ErrorResponse
// @Router       /api/v1/resources/{uid}/related [get]
func (c *ResourceController) GetRelatedResources(ctx *gin.Context) {
	uid := ctx.Param("uid")
	if uid == "" {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "bad_request",
			Message: "Resource UID is required",
		})
		return
	}

	resources, err := c.resourceService.GetRelatedResources(ctx.Request.Context(), uid)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_error",
			Message: err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, resources)
}
