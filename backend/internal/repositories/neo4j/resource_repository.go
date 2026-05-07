package neo4j

import (
	"context"
	"fmt"
	"time"

	"github.com/K8S-Graph-Explorer/backend/internal/domain/models"
	"github.com/K8S-Graph-Explorer/backend/internal/infrastructure/database"
)

// ResourceRepository implements the ResourceRepository interface using Neo4j
type ResourceRepository struct {
	client *database.Neo4jClient
}

// NewResourceRepository creates a new ResourceRepository
func NewResourceRepository(client *database.Neo4jClient) *ResourceRepository {
	return &ResourceRepository{client: client}
}

// Create creates a new resource in the graph database
func (r *ResourceRepository) Create(ctx context.Context, resource *models.K8sResource) error {
	query := `
		MERGE (n:K8sResource {uid: $uid})
		SET n.name = $name,
			n.namespace = $namespace,
			n.kind = $kind,
			n.apiVersion = $apiVersion,
			n.createdAt = $createdAt,
			n.labels = $labels
		RETURN n
	`
	params := map[string]interface{}{
		"uid":        resource.UID,
		"name":       resource.Name,
		"namespace":  resource.Namespace,
		"kind":       resource.Kind,
		"apiVersion": resource.APIVersion,
		"createdAt":  resource.CreatedAt.Format(time.RFC3339),
		"labels":     mapToString(resource.Labels),
	}

	_, err := r.client.RunWrite(ctx, query, params)
	return err
}

// GetByUID retrieves a resource by its UID
func (r *ResourceRepository) GetByUID(ctx context.Context, uid string) (*models.K8sResource, error) {
	query := `MATCH (n:K8sResource {uid: $uid}) RETURN n`
	results, err := r.client.RunQuery(ctx, query, map[string]interface{}{"uid": uid})
	if err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return nil, nil
	}

	return mapToResource(results[0]["n"])
}

// GetByNamespace retrieves all resources in a namespace
func (r *ResourceRepository) GetByNamespace(ctx context.Context, namespace string) ([]*models.K8sResource, error) {
	query := `MATCH (n:K8sResource {namespace: $namespace}) RETURN n`
	results, err := r.client.RunQuery(ctx, query, map[string]interface{}{"namespace": namespace})
	if err != nil {
		return nil, err
	}

	return mapToResources(results)
}

// GetByKind retrieves all resources of a specific kind
func (r *ResourceRepository) GetByKind(ctx context.Context, kind string, namespace string) ([]*models.K8sResource, error) {
	var query string
	params := map[string]interface{}{"kind": kind}

	if namespace != "" {
		query = `MATCH (n:K8sResource {kind: $kind, namespace: $namespace}) RETURN n`
		params["namespace"] = namespace
	} else {
		query = `MATCH (n:K8sResource {kind: $kind}) RETURN n`
	}

	results, err := r.client.RunQuery(ctx, query, params)
	if err != nil {
		return nil, err
	}

	return mapToResources(results)
}

// Update updates an existing resource
func (r *ResourceRepository) Update(ctx context.Context, resource *models.K8sResource) error {
	return r.Create(ctx, resource) // MERGE handles update
}

// Delete deletes a resource by UID
func (r *ResourceRepository) Delete(ctx context.Context, uid string) error {
	query := `MATCH (n:K8sResource {uid: $uid}) DETACH DELETE n`
	_, err := r.client.RunWrite(ctx, query, map[string]interface{}{"uid": uid})
	return err
}

// Search searches resources by name or label
func (r *ResourceRepository) Search(ctx context.Context, searchQuery string, namespace string) ([]*models.K8sResource, error) {
	cypher := `
		MATCH (n:K8sResource)
		WHERE n.name CONTAINS $query OR n.labels CONTAINS $query
	`
	params := map[string]interface{}{"query": searchQuery}

	if namespace != "" {
		cypher += ` AND n.namespace = $namespace`
		params["namespace"] = namespace
	}
	cypher += ` RETURN n LIMIT 50`

	results, err := r.client.RunQuery(ctx, cypher, params)
	if err != nil {
		return nil, err
	}

	return mapToResources(results)
}

// CreateRelationship creates a relationship between two resources
func (r *ResourceRepository) CreateRelationship(ctx context.Context, fromUID, toUID string, relType models.RelationshipType) error {
	query := fmt.Sprintf(`
		MATCH (a:K8sResource {uid: $fromUID}), (b:K8sResource {uid: $toUID})
		MERGE (a)-[r:%s]->(b)
		RETURN r
	`, relType)

	params := map[string]interface{}{
		"fromUID": fromUID,
		"toUID":   toUID,
	}

	_, err := r.client.RunWrite(ctx, query, params)
	return err
}

// GetRelatedResources retrieves resources related to a given resource
func (r *ResourceRepository) GetRelatedResources(ctx context.Context, uid string) ([]*models.K8sResource, error) {
	query := `
		MATCH (n:K8sResource {uid: $uid})-[r]-(related:K8sResource)
		RETURN DISTINCT related as n
	`
	results, err := r.client.RunQuery(ctx, query, map[string]interface{}{"uid": uid})
	if err != nil {
		return nil, err
	}

	return mapToResources(results)
}

// Helper functions
func mapToString(m map[string]string) string {
	if m == nil {
		return ""
	}
	result := ""
	for k, v := range m {
		if result != "" {
			result += ","
		}
		result += fmt.Sprintf("%s=%s", k, v)
	}
	return result
}

func mapToResource(data interface{}) (*models.K8sResource, error) {
	if data == nil {
		return nil, nil
	}

	// Convert Neo4j node to Resource
	nodeMap, ok := data.(map[string]interface{})
	if !ok {
		return nil, nil
	}

	props, ok := nodeMap["properties"].(map[string]interface{})
	if !ok {
		props = nodeMap // Sometimes properties are flat
	}

	resource := &models.K8sResource{}

	if uid, ok := props["uid"].(string); ok {
		resource.UID = uid
	}
	if name, ok := props["name"].(string); ok {
		resource.Name = name
	}
	if namespace, ok := props["namespace"].(string); ok {
		resource.Namespace = namespace
	}
	if kind, ok := props["kind"].(string); ok {
		resource.Kind = kind
	}
	if apiVersion, ok := props["apiVersion"].(string); ok {
		resource.APIVersion = apiVersion
	}
	if createdAt, ok := props["createdAt"].(string); ok {
		if t, err := time.Parse(time.RFC3339, createdAt); err == nil {
			resource.CreatedAt = t
		}
	}

	// Parse labels from string
	if labelsStr, ok := props["labels"].(string); ok && labelsStr != "" {
		resource.Labels = parseLabels(labelsStr)
	}

	return resource, nil
}

func parseLabels(labelsStr string) map[string]string {
	labels := make(map[string]string)
	if labelsStr == "" {
		return labels
	}

	// Parse "key1=value1,key2=value2" format
	pairs := splitString(labelsStr, ',')
	for _, pair := range pairs {
		kv := splitString(pair, '=')
		if len(kv) == 2 {
			labels[kv[0]] = kv[1]
		}
	}
	return labels
}

func splitString(s string, sep rune) []string {
	var result []string
	current := ""
	for _, c := range s {
		if c == sep {
			result = append(result, current)
			current = ""
		} else {
			current += string(c)
		}
	}
	if current != "" {
		result = append(result, current)
	}
	return result
}

func mapToResources(results []map[string]interface{}) ([]*models.K8sResource, error) {
	resources := make([]*models.K8sResource, 0, len(results))
	for _, result := range results {
		resource, err := mapToResource(result["n"])
		if err != nil {
			return nil, err
		}
		if resource != nil {
			resources = append(resources, resource)
		}
	}
	return resources, nil
}
