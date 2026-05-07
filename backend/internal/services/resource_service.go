package services

import (
	"context"

	"github.com/K8S-Graph-Explorer/backend/internal/domain/models"
	"github.com/K8S-Graph-Explorer/backend/internal/domain/repositories"
)

// ResourceService handles business logic for K8s resources
type ResourceService struct {
	resourceRepo repositories.ResourceRepository
}

// NewResourceService creates a new ResourceService
func NewResourceService(resourceRepo repositories.ResourceRepository) *ResourceService {
	return &ResourceService{
		resourceRepo: resourceRepo,
	}
}

// GetResource retrieves a resource by UID
func (s *ResourceService) GetResource(ctx context.Context, uid string) (*models.K8sResource, error) {
	return s.resourceRepo.GetByUID(ctx, uid)
}

// ListResources lists resources with optional filters
func (s *ResourceService) ListResources(ctx context.Context, namespace string, kind string) ([]*models.K8sResource, error) {
	if kind != "" {
		return s.resourceRepo.GetByKind(ctx, kind, namespace)
	}
	if namespace != "" {
		return s.resourceRepo.GetByNamespace(ctx, namespace)
	}
	// Return all resources (limited)
	return s.resourceRepo.GetByNamespace(ctx, "")
}

// CreateResource creates a new resource
func (s *ResourceService) CreateResource(ctx context.Context, resource *models.K8sResource) error {
	return s.resourceRepo.Create(ctx, resource)
}

// UpdateResource updates an existing resource
func (s *ResourceService) UpdateResource(ctx context.Context, resource *models.K8sResource) error {
	return s.resourceRepo.Update(ctx, resource)
}

// DeleteResource deletes a resource
func (s *ResourceService) DeleteResource(ctx context.Context, uid string) error {
	return s.resourceRepo.Delete(ctx, uid)
}

// SearchResources searches resources by query
func (s *ResourceService) SearchResources(ctx context.Context, query string, namespace string) ([]*models.K8sResource, error) {
	return s.resourceRepo.Search(ctx, query, namespace)
}

// GetRelatedResources gets resources related to a given resource
func (s *ResourceService) GetRelatedResources(ctx context.Context, uid string) ([]*models.K8sResource, error) {
	return s.resourceRepo.GetRelatedResources(ctx, uid)
}

// CreateRelationship creates a relationship between resources
func (s *ResourceService) CreateRelationship(ctx context.Context, fromUID, toUID string, relType models.RelationshipType) error {
	return s.resourceRepo.CreateRelationship(ctx, fromUID, toUID, relType)
}
