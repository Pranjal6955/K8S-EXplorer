package services

import (
	"context"

	"github.com/K8S-Graph-Explorer/backend/internal/domain/models"
	"github.com/K8S-Graph-Explorer/backend/internal/domain/repositories"
)

// GraphService handles business logic for graph operations
type GraphService struct {
	graphRepo repositories.GraphRepository
}

// NewGraphService creates a new GraphService
func NewGraphService(graphRepo repositories.GraphRepository) *GraphService {
	return &GraphService{
		graphRepo: graphRepo,
	}
}

// GetGraph retrieves the complete graph for a namespace
func (s *GraphService) GetGraph(ctx context.Context, namespace string) (*models.Graph, error) {
	return s.graphRepo.GetGraph(ctx, namespace)
}

// GetResourceTopology retrieves a subgraph starting from a specific resource
func (s *GraphService) GetResourceTopology(ctx context.Context, uid string, depth int) (*models.Graph, error) {
	if depth <= 0 {
		depth = 3 // Default depth
	}
	if depth > 10 {
		depth = 10 // Max depth
	}
	return s.graphRepo.GetSubGraph(ctx, uid, depth)
}

// ClearGraph clears the graph data
func (s *GraphService) ClearGraph(ctx context.Context, namespace string) error {
	return s.graphRepo.ClearGraph(ctx, namespace)
}
