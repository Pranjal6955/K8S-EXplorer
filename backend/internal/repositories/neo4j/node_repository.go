package neo4j

import (
	"context"
	"fmt"

	"github.com/K8S-Graph-Explorer/backend/internal/domain/models"
	"github.com/K8S-Graph-Explorer/backend/internal/infrastructure/database"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

// NodeRepository handles node operations in Neo4j
type NodeRepository struct {
	client *database.Neo4jClient
}

// NewNodeRepository creates a new NodeRepository
func NewNodeRepository(client *database.Neo4jClient) *NodeRepository {
	return &NodeRepository{client: client}
}

// CreateNode creates or updates a node in the graph
func (r *NodeRepository) CreateNode(ctx context.Context, node *models.Node) error {
	session := r.client.GetSession(ctx)
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		// Use MERGE to create or update node
		cypher := fmt.Sprintf(`
			MERGE (n:%s {uid: $uid})
			ON CREATE SET
				n.name = $name,
				n.namespace = $namespace,
				n.labels = $labels,
				n.createdAt = datetime(),
				n.updatedAt = datetime()
			ON MATCH SET
				n.name = $name,
				n.namespace = $namespace,
				n.labels = $labels,
				n.updatedAt = datetime()
			RETURN n
		`, node.Type)

		params := map[string]interface{}{
			"uid":       node.ID,
			"name":      node.Name,
			"namespace": node.Namespace,
			"labels":    node.Labels,
		}

		_, err := tx.Run(ctx, cypher, params)
		return nil, err
	})

	return err
}

// CreateNodeWithProperties creates a node with additional properties
func (r *NodeRepository) CreateNodeWithProperties(ctx context.Context, node *models.Node) error {
	session := r.client.GetSession(ctx)
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		cypher := fmt.Sprintf(`
			MERGE (n:%s {uid: $uid})
			ON CREATE SET
				n = $props,
				n.uid = $uid,
				n.createdAt = datetime()
			ON MATCH SET
				n += $props,
				n.updatedAt = datetime()
			RETURN n
		`, node.Type)

		// Prepare properties map
		props := map[string]interface{}{
			"name":      node.Name,
			"namespace": node.Namespace,
		}

		// Add labels as individual properties
		for k, v := range node.Labels {
			props["label_"+k] = v
		}

		// Add custom properties
		for k, v := range node.Properties {
			props[k] = v
		}

		params := map[string]interface{}{
			"uid":   node.ID,
			"props": props,
		}

		_, err := tx.Run(ctx, cypher, params)
		return nil, err
	})

	return err
}

// GetNodeByUID retrieves a node by its UID
func (r *NodeRepository) GetNodeByUID(ctx context.Context, uid string) (*models.Node, error) {
	cypher := `
		MATCH (n {uid: $uid})
		RETURN n, labels(n) as nodeLabels
	`

	records, err := r.client.RunQuery(ctx, cypher, map[string]interface{}{
		"uid": uid,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get node: %w", err)
	}

	if len(records) == 0 {
		return nil, nil
	}

	return r.recordToNode(records[0])
}

// GetNodesByType retrieves all nodes of a specific type
func (r *NodeRepository) GetNodesByType(ctx context.Context, nodeType models.NodeType, namespace string) ([]*models.Node, error) {
	var cypher string
	params := map[string]interface{}{}

	if namespace != "" {
		cypher = fmt.Sprintf(`
			MATCH (n:%s)
			WHERE n.namespace = $namespace
			RETURN n, labels(n) as nodeLabels
			ORDER BY n.name
		`, nodeType)
		params["namespace"] = namespace
	} else {
		cypher = fmt.Sprintf(`
			MATCH (n:%s)
			RETURN n, labels(n) as nodeLabels
			ORDER BY n.name
		`, nodeType)
	}

	records, err := r.client.RunQuery(ctx, cypher, params)
	if err != nil {
		return nil, fmt.Errorf("failed to get nodes by type: %w", err)
	}

	nodes := make([]*models.Node, 0, len(records))
	for _, record := range records {
		node, err := r.recordToNode(record)
		if err != nil {
			continue
		}
		nodes = append(nodes, node)
	}

	return nodes, nil
}

// DeleteNode deletes a node and all its relationships
func (r *NodeRepository) DeleteNode(ctx context.Context, uid string) error {
	session := r.client.GetSession(ctx)
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		// DETACH DELETE removes the node and all its relationships
		cypher := `
			MATCH (n {uid: $uid})
			DETACH DELETE n
			RETURN count(n) as deleted
		`

		result, err := tx.Run(ctx, cypher, map[string]interface{}{
			"uid": uid,
		})
		if err != nil {
			return nil, err
		}

		// Consume the result
		_, err = result.Consume(ctx)
		return nil, err
	})

	return err
}

// DeleteNodesByNamespace deletes all nodes in a namespace
func (r *NodeRepository) DeleteNodesByNamespace(ctx context.Context, namespace string) (int, error) {
	session := r.client.GetSession(ctx)
	defer session.Close(ctx)

	result, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		cypher := `
			MATCH (n {namespace: $namespace})
			WITH n, count(n) as total
			DETACH DELETE n
			RETURN total
		`

		res, err := tx.Run(ctx, cypher, map[string]interface{}{
			"namespace": namespace,
		})
		if err != nil {
			return 0, err
		}

		record, err := res.Single(ctx)
		if err != nil {
			return 0, nil
		}

		count, _ := record.Get("total")
		if c, ok := count.(int64); ok {
			return int(c), nil
		}
		return 0, nil
	})

	if err != nil {
		return 0, err
	}

	return result.(int), nil
}

// SearchNodes searches for nodes by name pattern
func (r *NodeRepository) SearchNodes(ctx context.Context, query string, namespace string) ([]*models.Node, error) {
	var cypher string
	params := map[string]interface{}{
		"query": "(?i).*" + query + ".*", // Case-insensitive regex
	}

	if namespace != "" {
		cypher = `
			MATCH (n)
			WHERE n.name =~ $query AND n.namespace = $namespace
			RETURN n, labels(n) as nodeLabels
			LIMIT 50
		`
		params["namespace"] = namespace
	} else {
		cypher = `
			MATCH (n)
			WHERE n.name =~ $query
			RETURN n, labels(n) as nodeLabels
			LIMIT 50
		`
	}

	records, err := r.client.RunQuery(ctx, cypher, params)
	if err != nil {
		return nil, fmt.Errorf("failed to search nodes: %w", err)
	}

	nodes := make([]*models.Node, 0, len(records))
	for _, record := range records {
		node, err := r.recordToNode(record)
		if err != nil {
			continue
		}
		nodes = append(nodes, node)
	}

	return nodes, nil
}

// recordToNode converts a Neo4j record to a Node model
func (r *NodeRepository) recordToNode(record map[string]interface{}) (*models.Node, error) {
	nodeData, ok := record["n"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid node data")
	}

	props, ok := nodeData["properties"].(map[string]interface{})
	if !ok {
		props = nodeData // Sometimes properties are flat
	}

	node := &models.Node{
		Labels:     make(map[string]string),
		Properties: make(map[string]interface{}),
	}

	// Extract standard fields
	if uid, ok := props["uid"].(string); ok {
		node.ID = uid
	}
	if name, ok := props["name"].(string); ok {
		node.Name = name
	}
	if namespace, ok := props["namespace"].(string); ok {
		node.Namespace = namespace
	}

	// Extract node type from labels
	if labels, ok := record["nodeLabels"].([]interface{}); ok && len(labels) > 0 {
		node.Type = models.NodeType(labels[0].(string))
	}

	// Extract K8s labels (stored with label_ prefix)
	for k, v := range props {
		if len(k) > 6 && k[:6] == "label_" {
			if str, ok := v.(string); ok {
				node.Labels[k[6:]] = str
			}
		}
	}

	return node, nil
}
