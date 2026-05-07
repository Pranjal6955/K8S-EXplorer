package neo4j

import (
	"context"
	"fmt"
	"time"

	"github.com/K8S-Graph-Explorer/backend/internal/domain/models"
	"github.com/K8S-Graph-Explorer/backend/internal/domain/repositories"
	"github.com/K8S-Graph-Explorer/backend/internal/infrastructure/database"
)

// GraphRepository implements the GraphRepository interface using Neo4j
type GraphRepository struct {
	client *database.Neo4jClient
}

// NewGraphRepository creates a new GraphRepository
func NewGraphRepository(client *database.Neo4jClient) *GraphRepository {
	return &GraphRepository{client: client}
}

// GetGraph retrieves the complete graph data
func (r *GraphRepository) GetGraph(ctx context.Context, namespace string) (*models.Graph, error) {
	// Query for nodes
	nodesQuery := `MATCH (n:K8sResource)`
	params := map[string]interface{}{}

	if namespace != "" {
		nodesQuery += ` WHERE n.namespace = $namespace`
		params["namespace"] = namespace
	}
	nodesQuery += ` RETURN n`

	nodeResults, err := r.client.RunQuery(ctx, nodesQuery, params)
	if err != nil {
		return nil, err
	}

	// Query for edges
	edgesQuery := `MATCH (a:K8sResource)-[r]->(b:K8sResource)`
	if namespace != "" {
		edgesQuery += ` WHERE a.namespace = $namespace OR b.namespace = $namespace`
	}
	edgesQuery += ` RETURN a.uid as source, b.uid as target, type(r) as relType`

	edgeResults, err := r.client.RunQuery(ctx, edgesQuery, params)
	if err != nil {
		return nil, err
	}

	// Build graph
	nodes := make([]models.GraphNode, 0, len(nodeResults))
	for _, result := range nodeResults {
		node := convertToGraphNode(result["n"])
		if node != nil {
			nodes = append(nodes, *node)
		}
	}

	edges := make([]models.GraphEdge, 0, len(edgeResults))
	for i, result := range edgeResults {
		edge := models.GraphEdge{
			ID:     fmt.Sprintf("edge-%d", i),
			Source: getString(result["source"]),
			Target: getString(result["target"]),
			Type:   models.RelationshipType(getString(result["relType"])),
		}
		edges = append(edges, edge)
	}

	return &models.Graph{
		Nodes: nodes,
		Edges: edges,
		Metadata: models.GraphMetadata{
			NodeCount:   len(nodes),
			EdgeCount:   len(edges),
			GeneratedAt: time.Now().Format(time.RFC3339),
		},
	}, nil
}

// GetSubGraph retrieves a subgraph starting from a specific resource
func (r *GraphRepository) GetSubGraph(ctx context.Context, rootUID string, depth int) (*models.Graph, error) {
	query := fmt.Sprintf(`
		MATCH path = (root:K8sResource {uid: $uid})-[*0..%d]-(related:K8sResource)
		WITH nodes(path) as nodes, relationships(path) as rels
		UNWIND nodes as n
		RETURN DISTINCT n
	`, depth)

	nodeResults, err := r.client.RunQuery(ctx, query, map[string]interface{}{"uid": rootUID})
	if err != nil {
		return nil, err
	}

	nodes := make([]models.GraphNode, 0, len(nodeResults))
	uids := make([]string, 0, len(nodeResults))
	for _, result := range nodeResults {
		node := convertToGraphNode(result["n"])
		if node != nil {
			nodes = append(nodes, *node)
			uids = append(uids, node.ID)
		}
	}

	// Get edges between these nodes
	edgesQuery := `
		MATCH (a:K8sResource)-[r]->(b:K8sResource)
		WHERE a.uid IN $uids AND b.uid IN $uids
		RETURN a.uid as source, b.uid as target, type(r) as relType
	`

	edgeResults, err := r.client.RunQuery(ctx, edgesQuery, map[string]interface{}{"uids": uids})
	if err != nil {
		return nil, err
	}

	edges := make([]models.GraphEdge, 0, len(edgeResults))
	for i, result := range edgeResults {
		edge := models.GraphEdge{
			ID:     fmt.Sprintf("edge-%d", i),
			Source: getString(result["source"]),
			Target: getString(result["target"]),
			Type:   models.RelationshipType(getString(result["relType"])),
		}
		edges = append(edges, edge)
	}

	return &models.Graph{
		Nodes: nodes,
		Edges: edges,
		Metadata: models.GraphMetadata{
			NodeCount:   len(nodes),
			EdgeCount:   len(edges),
			GeneratedAt: time.Now().Format(time.RFC3339),
		},
	}, nil
}

// ClearGraph clears all data from the graph
func (r *GraphRepository) ClearGraph(ctx context.Context, namespace string) error {
	query := `MATCH (n:K8sResource)`
	params := map[string]interface{}{}

	if namespace != "" {
		query += ` WHERE n.namespace = $namespace`
		params["namespace"] = namespace
	}
	query += ` DETACH DELETE n`

	_, err := r.client.RunWrite(ctx, query, params)
	return err
}

// SyncGraph syncs an entire graph schema to the database
func (r *GraphRepository) SyncGraph(ctx context.Context, graph *models.GraphSchema) error {
	// Create nodes
	for _, node := range graph.Nodes {
		query := fmt.Sprintf(`
			MERGE (n:%s {uid: $uid})
			SET n.name = $name, n.namespace = $namespace, n.updatedAt = datetime()
		`, node.Type)

		_, err := r.client.RunWrite(ctx, query, map[string]interface{}{
			"uid":       node.ID,
			"name":      node.Name,
			"namespace": node.Namespace,
		})
		if err != nil {
			return fmt.Errorf("failed to create node %s: %w", node.Name, err)
		}
	}

	// Create edges
	for _, edge := range graph.Edges {
		query := fmt.Sprintf(`
			MATCH (source {uid: $sourceId})
			MATCH (target {uid: $targetId})
			MERGE (source)-[r:%s]->(target)
			SET r.updatedAt = datetime()
		`, edge.Type)

		_, err := r.client.RunWrite(ctx, query, map[string]interface{}{
			"sourceId": edge.SourceID,
			"targetId": edge.TargetID,
		})
		if err != nil {
			return fmt.Errorf("failed to create edge: %w", err)
		}
	}

	return nil
}

// GetStatistics returns graph statistics
func (r *GraphRepository) GetStatistics(ctx context.Context) (*repositories.GraphStatistics, error) {
	// Get node counts by type
	nodeQuery := `
		MATCH (n:K8sResource)
		RETURN n.kind as kind, count(n) as count
	`
	nodeResults, err := r.client.RunQuery(ctx, nodeQuery, nil)
	if err != nil {
		return nil, err
	}

	nodesByType := make(map[string]int)
	totalNodes := 0
	for _, result := range nodeResults {
		kind := getString(result["kind"])
		count := getInt(result["count"])
		nodesByType[kind] = count
		totalNodes += count
	}

	// Get edge counts by type
	edgeQuery := `
		MATCH ()-[r]->()
		RETURN type(r) as relType, count(r) as count
	`
	edgeResults, err := r.client.RunQuery(ctx, edgeQuery, nil)
	if err != nil {
		return nil, err
	}

	edgesByType := make(map[string]int)
	totalEdges := 0
	for _, result := range edgeResults {
		relType := getString(result["relType"])
		count := getInt(result["count"])
		edgesByType[relType] = count
		totalEdges += count
	}

	// Get namespaces
	nsQuery := `
		MATCH (n:K8sResource)
		RETURN DISTINCT n.namespace as namespace
	`
	nsResults, err := r.client.RunQuery(ctx, nsQuery, nil)
	if err != nil {
		return nil, err
	}

	namespaces := make([]string, 0, len(nsResults))
	for _, result := range nsResults {
		if ns := getString(result["namespace"]); ns != "" {
			namespaces = append(namespaces, ns)
		}
	}

	return &repositories.GraphStatistics{
		TotalNodes:  totalNodes,
		TotalEdges:  totalEdges,
		NodesByType: nodesByType,
		EdgesByType: edgesByType,
		Namespaces:  namespaces,
	}, nil
}

// Helper functions
func convertToGraphNode(data interface{}) *models.GraphNode {
	if data == nil {
		return nil
	}

	nodeMap, ok := data.(map[string]interface{})
	if !ok {
		return nil
	}

	props, ok := nodeMap["properties"].(map[string]interface{})
	if !ok {
		props = nodeMap
	}

	node := &models.GraphNode{
		Data: map[string]interface{}{},
	}

	if uid, ok := props["uid"].(string); ok {
		node.ID = uid
	}
	if name, ok := props["name"].(string); ok {
		node.Label = name
	}
	if kind, ok := props["kind"].(string); ok {
		node.Kind = kind
	}

	// Copy other properties to Data
	for k, v := range props {
		if k != "uid" && k != "name" && k != "kind" {
			node.Data[k] = v
		}
	}

	return node
}

func getString(v interface{}) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

func getInt(v interface{}) int {
	if v == nil {
		return 0
	}
	switch n := v.(type) {
	case int:
		return n
	case int64:
		return int(n)
	case float64:
		return int(n)
	}
	return 0
}
