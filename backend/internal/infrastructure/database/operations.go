package database

import (
	"context"
	"fmt"

	"github.com/K8S-Graph-Explorer/backend/internal/domain/models"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

// GraphOperations provides high-level graph operations for Neo4j
type GraphOperations struct {
	client *Neo4jClient
}

// NewGraphOperations creates a new GraphOperations instance
func NewGraphOperations(client *Neo4jClient) *GraphOperations {
	return &GraphOperations{client: client}
}

// CreateNode creates a node in the graph
func (g *GraphOperations) CreateNode(ctx context.Context, node models.Node) error {
	cypher := `
		MERGE (n:%s {uid: $uid})
		ON CREATE SET 
			n.name = $name,
			n.namespace = $namespace,
			n.labels = $labels,
			n.properties = $properties,
			n.createdAt = datetime()
		ON MATCH SET
			n.name = $name,
			n.namespace = $namespace,
			n.labels = $labels,
			n.properties = $properties,
			n.updatedAt = datetime()
	`

	_, err := g.client.RunWrite(ctx, fmt.Sprintf(cypher, node.Type), map[string]interface{}{
		"uid":        node.ID,
		"name":       node.Name,
		"namespace":  node.Namespace,
		"labels":     node.Labels,
		"properties": node.Properties,
	})

	return err
}

// CreateEdge creates an edge/relationship in the graph
func (g *GraphOperations) CreateEdge(ctx context.Context, edge models.Edge) error {
	cypher := `
		MATCH (source {uid: $sourceId})
		MATCH (target {uid: $targetId})
		MERGE (source)-[r:%s]->(target)
		ON CREATE SET
			r.id = $id,
			r.properties = $properties,
			r.createdAt = datetime()
		ON MATCH SET
			r.properties = $properties,
			r.updatedAt = datetime()
	`

	_, err := g.client.RunWrite(ctx, fmt.Sprintf(cypher, edge.Type), map[string]interface{}{
		"id":         edge.ID,
		"sourceId":   edge.SourceID,
		"targetId":   edge.TargetID,
		"properties": edge.Properties,
	})

	return err
}

// DeleteNode deletes a node and its relationships
func (g *GraphOperations) DeleteNode(ctx context.Context, uid string) error {
	cypher := `
		MATCH (n {uid: $uid})
		DETACH DELETE n
	`

	_, err := g.client.RunWrite(ctx, cypher, map[string]interface{}{
		"uid": uid,
	})

	return err
}

// GetNode retrieves a node by UID
func (g *GraphOperations) GetNode(ctx context.Context, uid string) (*models.Node, error) {
	cypher := `
		MATCH (n {uid: $uid})
		RETURN n, labels(n) as nodeLabels
	`

	records, err := g.client.RunQuery(ctx, cypher, map[string]interface{}{
		"uid": uid,
	})

	if err != nil {
		return nil, err
	}

	if len(records) == 0 {
		return nil, nil
	}

	return g.recordToNode(records[0])
}

// GetNodesByType retrieves all nodes of a specific type
func (g *GraphOperations) GetNodesByType(ctx context.Context, nodeType string, namespace string) ([]models.Node, error) {
	var cypher string
	params := map[string]interface{}{}

	if namespace != "" {
		cypher = fmt.Sprintf(`
			MATCH (n:%s {namespace: $namespace})
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

	records, err := g.client.RunQuery(ctx, cypher, params)
	if err != nil {
		return nil, err
	}

	nodes := make([]models.Node, 0, len(records))
	for _, record := range records {
		node, err := g.recordToNode(record)
		if err != nil {
			continue
		}
		nodes = append(nodes, *node)
	}

	return nodes, nil
}

// GetRelatedNodes gets nodes related to a given node
func (g *GraphOperations) GetRelatedNodes(ctx context.Context, uid string, depth int) ([]models.Node, []models.Edge, error) {
	cypher := `
		MATCH (start {uid: $uid})
		CALL apoc.path.subgraphAll(start, {maxLevel: $depth})
		YIELD nodes, relationships
		RETURN nodes, relationships
	`

	// Fallback if APOC is not available
	fallbackCypher := fmt.Sprintf(`
		MATCH (start {uid: $uid})-[r*1..%d]-(related)
		WITH start, collect(DISTINCT related) as nodes, collect(DISTINCT r) as rels
		RETURN start, nodes, rels
	`, depth)

	records, err := g.client.RunQuery(ctx, cypher, map[string]interface{}{
		"uid":   uid,
		"depth": depth,
	})

	if err != nil {
		// Try fallback
		records, err = g.client.RunQuery(ctx, fallbackCypher, map[string]interface{}{
			"uid": uid,
		})
		if err != nil {
			return nil, nil, err
		}
	}

	var nodes []models.Node
	var edges []models.Edge

	for _, record := range records {
		if nodesData, ok := record["nodes"].([]interface{}); ok {
			for _, n := range nodesData {
				if nodeMap, ok := n.(map[string]interface{}); ok {
					node, _ := g.mapToNode(nodeMap)
					if node != nil {
						nodes = append(nodes, *node)
					}
				}
			}
		}

		if relsData, ok := record["relationships"].([]interface{}); ok {
			for _, r := range relsData {
				if relMap, ok := r.(map[string]interface{}); ok {
					edge := g.mapToEdge(relMap)
					edges = append(edges, edge)
				}
			}
		}
	}

	return nodes, edges, nil
}

// SyncGraph syncs an entire graph schema to Neo4j
func (g *GraphOperations) SyncGraph(ctx context.Context, graph *models.GraphSchema) error {
	session := g.client.GetSession(ctx)
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		// Create/update all nodes
		for _, node := range graph.Nodes {
			cypher := fmt.Sprintf(`
				MERGE (n:%s {uid: $uid})
				SET n.name = $name, n.namespace = $namespace, n.updatedAt = datetime()
			`, node.Type)

			if _, err := tx.Run(ctx, cypher, map[string]interface{}{
				"uid":       node.ID,
				"name":      node.Name,
				"namespace": node.Namespace,
			}); err != nil {
				return nil, err
			}
		}

		// Create/update all edges
		for _, edge := range graph.Edges {
			cypher := fmt.Sprintf(`
				MATCH (source {uid: $sourceId})
				MATCH (target {uid: $targetId})
				MERGE (source)-[r:%s]->(target)
				SET r.updatedAt = datetime()
			`, edge.Type)

			if _, err := tx.Run(ctx, cypher, map[string]interface{}{
				"sourceId": edge.SourceID,
				"targetId": edge.TargetID,
			}); err != nil {
				return nil, err
			}
		}

		return nil, nil
	})

	return err
}

// ClearNamespace removes all nodes and relationships in a namespace
func (g *GraphOperations) ClearNamespace(ctx context.Context, namespace string) error {
	cypher := `
		MATCH (n {namespace: $namespace})
		DETACH DELETE n
	`

	_, err := g.client.RunWrite(ctx, cypher, map[string]interface{}{
		"namespace": namespace,
	})

	return err
}

// ClearAll removes all nodes and relationships
func (g *GraphOperations) ClearAll(ctx context.Context) error {
	cypher := `MATCH (n) DETACH DELETE n`
	_, err := g.client.RunWrite(ctx, cypher, nil)
	return err
}

// Helper methods

func (g *GraphOperations) recordToNode(record map[string]interface{}) (*models.Node, error) {
	nodeData, ok := record["n"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid node data")
	}
	return g.mapToNode(nodeData)
}

func (g *GraphOperations) mapToNode(data map[string]interface{}) (*models.Node, error) {
	props, _ := data["properties"].(map[string]interface{})

	node := &models.Node{
		Properties: make(map[string]interface{}),
	}

	if id, ok := props["uid"].(string); ok {
		node.ID = id
	}
	if name, ok := props["name"].(string); ok {
		node.Name = name
	}
	if namespace, ok := props["namespace"].(string); ok {
		node.Namespace = namespace
	}
	if labels, ok := data["labels"].([]interface{}); ok && len(labels) > 0 {
		node.Type = models.NodeType(labels[0].(string))
	}
	if labelsMap, ok := props["labels"].(map[string]interface{}); ok {
		node.Labels = make(map[string]string)
		for k, v := range labelsMap {
			if str, ok := v.(string); ok {
				node.Labels[k] = str
			}
		}
	}

	return node, nil
}

func (g *GraphOperations) mapToEdge(data map[string]interface{}) models.Edge {
	edge := models.Edge{}

	if id, ok := data["id"].(string); ok {
		edge.ID = id
	}
	if relType, ok := data["type"].(string); ok {
		edge.Type = models.EdgeType(relType)
	}
	if startNode, ok := data["startNode"].(string); ok {
		edge.SourceID = startNode
	}
	if endNode, ok := data["endNode"].(string); ok {
		edge.TargetID = endNode
	}
	if props, ok := data["properties"].(map[string]interface{}); ok {
		edge.Properties = props
	}

	return edge
}
