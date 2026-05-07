package repositories

import (
	"context"

	"github.com/K8S-Graph-Explorer/backend/internal/domain/models"
)

// NodeRepository defines operations for graph nodes
type NodeRepository interface {
	// CreateNode creates or updates a node in the graph
	CreateNode(ctx context.Context, node *models.Node) error

	// CreateNodeWithProperties creates a node with additional properties
	CreateNodeWithProperties(ctx context.Context, node *models.Node) error

	// GetNodeByUID retrieves a node by its UID
	GetNodeByUID(ctx context.Context, uid string) (*models.Node, error)

	// GetNodesByType retrieves all nodes of a specific type
	GetNodesByType(ctx context.Context, nodeType models.NodeType, namespace string) ([]*models.Node, error)

	// DeleteNode deletes a node and all its relationships
	DeleteNode(ctx context.Context, uid string) error

	// DeleteNodesByNamespace deletes all nodes in a namespace
	DeleteNodesByNamespace(ctx context.Context, namespace string) (int, error)

	// SearchNodes searches for nodes by name pattern
	SearchNodes(ctx context.Context, query string, namespace string) ([]*models.Node, error)
}

// EdgeRepository defines operations for graph edges/relationships
type EdgeRepository interface {
	// CreateEdge creates or updates an edge between two nodes
	CreateEdge(ctx context.Context, edge *models.Edge) error

	// CreateEdgeWithProperties creates an edge with additional properties
	CreateEdgeWithProperties(ctx context.Context, edge *models.Edge) error

	// DeleteEdge deletes an edge by its ID
	DeleteEdge(ctx context.Context, edgeID string) error

	// DeleteEdgeBetweenNodes deletes all edges of a specific type between two nodes
	DeleteEdgeBetweenNodes(ctx context.Context, sourceID, targetID string, edgeType models.EdgeType) error

	// GetEdgesByNode retrieves all edges connected to a node
	GetEdgesByNode(ctx context.Context, nodeUID string) ([]*models.Edge, error)

	// GetEdgesByType retrieves all edges of a specific type
	GetEdgesByType(ctx context.Context, edgeType models.EdgeType) ([]*models.Edge, error)

	// QueryRelations retrieves related nodes through specified relationship types
	QueryRelations(ctx context.Context, nodeUID string, edgeTypes []models.EdgeType, direction string, depth int) ([]*models.Node, []*models.Edge, error)

	// QueryPath finds the shortest path between two nodes
	QueryPath(ctx context.Context, startUID, endUID string, maxDepth int) ([]*models.Node, []*models.Edge, error)
}

// GraphRepository defines operations for the entire graph
type GraphRepository interface {
	// GetGraph retrieves the complete graph for a namespace
	GetGraph(ctx context.Context, namespace string) (*models.Graph, error)

	// GetSubGraph retrieves a subgraph starting from a specific node
	GetSubGraph(ctx context.Context, nodeUID string, depth int) (*models.Graph, error)

	// SyncGraph syncs an entire graph schema to the database
	SyncGraph(ctx context.Context, graph *models.GraphSchema) error

	// ClearGraph clears the graph data for a namespace
	ClearGraph(ctx context.Context, namespace string) error

	// GetStatistics returns graph statistics
	GetStatistics(ctx context.Context) (*GraphStatistics, error)
}

// ResourceRepository defines operations for K8s resources (legacy interface)
type ResourceRepository interface {
	// Create creates or updates a resource
	Create(ctx context.Context, resource *models.K8sResource) error

	// Update updates an existing resource
	Update(ctx context.Context, resource *models.K8sResource) error

	// Delete deletes a resource
	Delete(ctx context.Context, uid string) error

	// GetByUID retrieves a resource by UID
	GetByUID(ctx context.Context, uid string) (*models.K8sResource, error)

	// GetByNamespace retrieves all resources in a namespace
	GetByNamespace(ctx context.Context, namespace string) ([]*models.K8sResource, error)

	// GetByKind retrieves all resources of a specific kind
	GetByKind(ctx context.Context, kind string, namespace string) ([]*models.K8sResource, error)

	// Search searches for resources by query
	Search(ctx context.Context, query string, namespace string) ([]*models.K8sResource, error)

	// GetRelatedResources gets resources related to a given resource
	GetRelatedResources(ctx context.Context, uid string) ([]*models.K8sResource, error)

	// CreateRelationship creates a relationship between resources
	CreateRelationship(ctx context.Context, fromUID, toUID string, relType models.RelationshipType) error
}

// GraphStatistics contains graph statistics
type GraphStatistics struct {
	TotalNodes  int            `json:"totalNodes"`
	TotalEdges  int            `json:"totalEdges"`
	NodesByType map[string]int `json:"nodesByType"`
	EdgesByType map[string]int `json:"edgesByType"`
	Namespaces  []string       `json:"namespaces"`
}
