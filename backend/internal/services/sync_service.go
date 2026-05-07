package services

import (
	"context"
	"fmt"
	"time"

	"github.com/K8S-Graph-Explorer/backend/internal/domain/models"
	"github.com/K8S-Graph-Explorer/backend/internal/domain/repositories"
	"github.com/K8S-Graph-Explorer/backend/internal/infrastructure/kubernetes"
)

// Broadcaster defines the interface for broadcasting messages
type Broadcaster interface {
	Broadcast(message interface{})
}

// SyncService handles synchronization between K8s and Neo4j
type SyncService struct {
	k8sClient    *kubernetes.Client
	resourceRepo repositories.ResourceRepository
	broadcaster  Broadcaster
}

// NewSyncService creates a new SyncService
func NewSyncService(k8sClient *kubernetes.Client, resourceRepo repositories.ResourceRepository, broadcaster Broadcaster) *SyncService {
	return &SyncService{
		k8sClient:    k8sClient,
		resourceRepo: resourceRepo,
		broadcaster:  broadcaster,
	}
}

// SyncResult represents the result of a sync operation
type SyncResult struct {
	Success     bool     `json:"success"`
	Message     string   `json:"message"`
	SyncedCount int      `json:"syncedCount"`
	Errors      []string `json:"errors,omitempty"`
}

// SyncNamespace syncs resources from a specific namespace
func (s *SyncService) SyncNamespace(ctx context.Context, namespace string) (*SyncResult, error) {
	result := &SyncResult{
		Success: true,
		Errors:  []string{},
	}

	// Sync Pods
	pods, err := s.k8sClient.ListPods(ctx, namespace)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("failed to list pods: %v", err))
	} else {
		for _, pod := range pods {
			resource := convertK8sResource(pod)
			if err := s.resourceRepo.Create(ctx, resource); err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("failed to create pod %s: %v", pod.Name, err))
			} else {
				result.SyncedCount++
			}
			// Create owner relationships
			for _, owner := range pod.OwnerRefs {
				if err := s.resourceRepo.CreateRelationship(ctx, owner.UID, pod.UID, models.RelationshipOwns); err != nil {
					result.Errors = append(result.Errors, fmt.Sprintf("failed to create relationship: %v", err))
				}
			}
		}
	}

	// Sync Deployments
	deployments, err := s.k8sClient.ListDeployments(ctx, namespace)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("failed to list deployments: %v", err))
	} else {
		for _, deployment := range deployments {
			resource := convertK8sResource(deployment)
			if err := s.resourceRepo.Create(ctx, resource); err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("failed to create deployment %s: %v", deployment.Name, err))
			} else {
				result.SyncedCount++
			}
		}
	}

	// Sync Services
	services, err := s.k8sClient.ListServices(ctx, namespace)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("failed to list services: %v", err))
	} else {
		for _, svc := range services {
			resource := convertK8sResource(svc)
			if err := s.resourceRepo.Create(ctx, resource); err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("failed to create service %s: %v", svc.Name, err))
			} else {
				result.SyncedCount++
			}
		}
	}

	if len(result.Errors) > 0 {
		result.Success = false
		result.Message = fmt.Sprintf("Sync completed with %d errors", len(result.Errors))
	} else {
		result.Message = fmt.Sprintf("Successfully synced %d resources", result.SyncedCount)

		// Broadcast update event
		if s.broadcaster != nil {
			s.broadcaster.Broadcast(map[string]interface{}{
				"type": "TOPOLOGY_UPDATE",
				"payload": map[string]interface{}{
					"namespace":   namespace,
					"syncedCount": result.SyncedCount,
					"timestamp":   time.Now(),
				},
			})
		}
	}

	return result, nil
}

// SyncAllNamespaces syncs resources from all namespaces
func (s *SyncService) SyncAllNamespaces(ctx context.Context) (*SyncResult, error) {
	result := &SyncResult{
		Success: true,
		Errors:  []string{},
	}

	// Get all namespaces
	namespaces, err := s.k8sClient.ListNamespaces(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list namespaces: %w", err)
	}

	// Sync namespace resources
	for _, ns := range namespaces {
		resource := convertK8sResource(ns)
		if err := s.resourceRepo.Create(ctx, resource); err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("failed to create namespace %s: %v", ns.Name, err))
		} else {
			result.SyncedCount++
		}
	}

	// Sync resources in each namespace concurrently
	type nsSyncResult struct {
		result *SyncResult
		err    error
	}
	resChan := make(chan nsSyncResult, len(namespaces))

	for _, ns := range namespaces {
		go func(namespaceName string) {
			nsResult, err := s.SyncNamespace(ctx, namespaceName)
			resChan <- nsSyncResult{nsResult, err}
		}(ns.Name)
	}

	for i := 0; i < len(namespaces); i++ {
		res := <-resChan
		if res.err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("failed to sync namespace: %v", res.err))
		} else if res.result != nil {
			result.SyncedCount += res.result.SyncedCount
			result.Errors = append(result.Errors, res.result.Errors...)
		}
	}

	if len(result.Errors) > 0 {
		result.Success = false
		result.Message = fmt.Sprintf("Sync completed with %d errors", len(result.Errors))
	} else {
		result.Message = fmt.Sprintf("Successfully synced %d resources", result.SyncedCount)

		// Broadcast update event for all namespaces
		if s.broadcaster != nil {
			s.broadcaster.Broadcast(map[string]interface{}{
				"type": "TOPOLOGY_UPDATE_ALL", // Different type or same?
				"payload": map[string]interface{}{
					"syncedCount": result.SyncedCount,
					"timestamp":   time.Now(),
				},
			})
		}
	}

	return result, nil
}

// convertK8sResource converts a kubernetes.Resource to models.K8sResource
func convertK8sResource(r kubernetes.Resource) *models.K8sResource {
	ownerRefs := make([]models.OwnerReference, len(r.OwnerRefs))
	for i, ref := range r.OwnerRefs {
		ownerRefs[i] = models.OwnerReference{
			UID:        ref.UID,
			Name:       ref.Name,
			Kind:       ref.Kind,
			APIVersion: ref.APIVersion,
		}
	}

	return &models.K8sResource{
		UID:        r.UID,
		Name:       r.Name,
		Namespace:  r.Namespace,
		Kind:       r.Kind,
		APIVersion: r.APIVersion,
		Labels:     r.Labels,
		CreatedAt:  time.Now(),
		OwnerRefs:  ownerRefs,
	}
}
