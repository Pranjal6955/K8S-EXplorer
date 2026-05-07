package controllers

import (
	"net/http"

	"github.com/K8S-Graph-Explorer/backend/internal/domain/models"
	"github.com/K8S-Graph-Explorer/backend/internal/services/converter"
	"github.com/gin-gonic/gin"
)

// SchemaController handles graph schema conversion endpoints
type SchemaController struct {
	converter *converter.GraphConverter
}

// NewSchemaController creates a new SchemaController
func NewSchemaController(converter *converter.GraphConverter) *SchemaController {
	return &SchemaController{
		converter: converter,
	}
}

// GetGraphSchema returns the normalized graph schema for a namespace
// @Summary      Get graph schema
// @Description  Converts Kubernetes resources to a normalized graph schema with nodes and edges
// @Tags         Schema
// @Produce      json
// @Param        namespace  query  string  false  "Namespace (empty for all namespaces)"
// @Success      200  {object}  models.GraphSchema
// @Failure      500  {object}  ErrorResponse
// @Router       /api/v1/schema [get]
func (c *SchemaController) GetGraphSchema(ctx *gin.Context) {
	namespace := ctx.Query("namespace")

	var graph interface{}
	var err error

	if namespace == "" {
		// Get graph for all namespaces
		graph, err = c.converter.ConvertAllNamespaces(ctx.Request.Context())
	} else {
		// Get graph for specific namespace
		graph, err = c.converter.ConvertNamespace(ctx.Request.Context(), namespace)
	}

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "conversion_error",
			Message: err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, graph)
}

// GetGraphSchemaJSON returns the graph schema as formatted JSON
// @Summary      Get graph schema as JSON string
// @Description  Returns the graph schema as a formatted JSON string
// @Tags         Schema
// @Produce      json
// @Param        namespace  query  string  false  "Namespace (empty for all namespaces)"
// @Success      200  {object}  map[string]string
// @Failure      500  {object}  ErrorResponse
// @Router       /api/v1/schema/json [get]
func (c *SchemaController) GetGraphSchemaJSON(ctx *gin.Context) {
	namespace := ctx.Query("namespace")

	var graph interface{}
	var err error

	if namespace == "" {
		graph, err = c.converter.ConvertAllNamespaces(ctx.Request.Context())
	} else {
		graph, err = c.converter.ConvertNamespace(ctx.Request.Context(), namespace)
	}

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "conversion_error",
			Message: err.Error(),
		})
		return
	}

	// Return as formatted JSON string
	jsonStr, err := c.converter.ToJSONString(graph.(*models.GraphSchema))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "serialization_error",
			Message: err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"schema": jsonStr,
	})
}
