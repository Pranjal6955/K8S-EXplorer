package neo4j

import (
	"context"
	"fmt"

	"github.com/K8S-Graph-Explorer/backend/internal/domain/models"
	"github.com/K8S-Graph-Explorer/backend/internal/infrastructure/database"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

// EdgeRepository handles edge/relationship operations in Neo4j
type EdgeRepository struct {
	client *database.Neo4jClient
}

// NewEdgeRepository creates a new EdgeRepository
func NewEdgeRepository(client *database.Neo4jClient) *EdgeRepository {
	return &EdgeRepository{client: client}
}

// CreateEdge creates or updates an edge between two nodes
func (r *EdgeRepository) CreateEdge(ctx context.Context, edge *models.Edge) error {
	session := r.client.GetSession(ctx)
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		// MERGE to create or update relationship
		cypher := fmt.Sprintf(`
			MATCH (source {uid: $sourceId})
			MATCH (target {uid: $targetId})
			MERGE (source)-[r:%s]->(target)
			ON CREATE SET
				r.id = $id,
				r.createdAt = datetime()
			ON MATCH SET
				r.updatedAt = datetime()
			RETURN r
		`, edge.Type)

		params := map[string]interface{}{
			"id":       edge.ID,
			"sourceId": edge.SourceID,
			"targetId": edge.TargetID,
		}

		_, err := tx.Run(ctx, cypher, params)
		return nil, err
	})

	return err
}

// CreateEdgeWithProperties creates an edge with additional properties
func (r *EdgeRepository) CreateEdgeWithProperties(ctx context.Context, edge *models.Edge) error {
	session := r.client.GetSession(ctx)
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		cypher := fmt.Sprintf(`
			MATCH (source {uid: $sourceId})
			MATCH (target {uid: $targetId})
			MERGE (source)-[r:%s]->(target)
			ON CREATE SET
				r = $props,
				r.id = $id,
				r.createdAt = datetime()
			ON MATCH SET
				r += $props,
				r.updatedAt = datetime()
			RETURN r
		`, edge.Type)

		params := map[string]interface{}{
			"id":       edge.ID,
			"sourceId": edge.SourceID,
			"targetId": edge.TargetID,
			"props":    edge.Properties,
		}

		_, err := tx.Run(ctx, cypher, params)
		return nil, err
	})

	return err
}

// DeleteEdge deletes an edge by its ID
func (r *EdgeRepository) DeleteEdge(ctx context.Context, edgeID string) error {
	session := r.client.GetSession(ctx)
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		cypher := `
			MATCH ()-[r {id: $id}]->()
			DELETE r
			RETURN count(r) as deleted
		`

		_, err := tx.Run(ctx, cypher, map[string]interface{}{
			"id": edgeID,
		})
		return nil, err
	})

	return err
}

// DeleteEdgeBetweenNodes deletes all edges of a specific type between two nodes
func (r *EdgeRepository) DeleteEdgeBetweenNodes(ctx context.Context, sourceID, targetID string, edgeType models.EdgeType) error {
	session := r.client.GetSession(ctx)
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		cypher := fmt.Sprintf(`
			MATCH (source {uid: $sourceId})-[r:%s]->(target {uid: $targetId})
			DELETE r
			RETURN count(r) as deleted
		`, edgeType)

		_, err := tx.Run(ctx, cypher, map[string]interface{}{
			"sourceId": sourceID,
			"targetId": targetID,
		})
		return nil, err
	})

	return err
}

// GetEdgesByNode retrieves all edges connected to a node
func (r *EdgeRepository) GetEdgesByNode(ctx context.Context, nodeUID string) ([]*models.Edge, error) {
	cypher := `
		MATCH (n {uid: $uid})-[r]-(m)
		RETURN 
			r.id as id,
			type(r) as type,
			startNode(r).uid as sourceId,
			endNode(r).uid as targetId,
			properties(r) as properties
	`

	records, err := r.client.RunQuery(ctx, cypher, map[string]interface{}{
		"uid": nodeUID,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get edges: %w", err)
	}

	edges := make([]*models.Edge, 0, len(records))
	for _, record := range records {
		edge := r.recordToEdge(record)
		edges = append(edges, edge)
	}

	return edges, nil
}

// GetEdgesByType retrieves all edges of a specific type
func (r *EdgeRepository) GetEdgesByType(ctx context.Context, edgeType models.EdgeType) ([]*models.Edge, error) {
	cypher := fmt.Sprintf(`
		MATCH (source)-[r:%s]->(target)
		RETURN 
			r.id as id,
			type(r) as type,
			source.uid as sourceId,
			target.uid as targetId,
			properties(r) as properties
		LIMIT 1000
	`, edgeType)

	records, err := r.client.RunQuery(ctx, cypher, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get edges by type: %w", err)
	}

	edges := make([]*models.Edge, 0, len(records))
	for _, record := range records {
		edge := r.recordToEdge(record)
		edges = append(edges, edge)
	}

	return edges, nil
}

// QueryRelations retrieves related nodes through specified relationship types
func (r *EdgeRepository) QueryRelations(ctx context.Context, nodeUID string, edgeTypes []models.EdgeType, direction string, depth int) ([]*models.Node, []*models.Edge, error) {
	// Build relationship type pattern
	var relPattern string
	if len(edgeTypes) > 0 {
		types := ""
		for i, t := range edgeTypes {
			if i > 0 {
				types += "|"
			}
			types += string(t)
		}
		relPattern = fmt.Sprintf("[r:%s*1..%d]", types, depth)
	} else {
		relPattern = fmt.Sprintf("[r*1..%d]", depth)
	}

	// Build direction pattern
	var cypher string
	switch direction {
	case "outgoing":
		cypher = fmt.Sprintf(`
			MATCH (start {uid: $uid})-%s->(related)
			WITH start, collect(DISTINCT related) as nodes, collect(DISTINCT r) as relationships
			UNWIND nodes as n
			UNWIND relationships as relList
			UNWIND relList as rel
			RETURN DISTINCT 
				n.uid as nodeUid, 
				n.name as nodeName, 
				n.namespace as nodeNamespace,
				labels(n) as nodeLabels,
				type(rel) as relType,
				startNode(rel).uid as relSource,
				endNode(rel).uid as relTarget
		`, relPattern)
	case "incoming":
		cypher = fmt.Sprintf(`
			MATCH (start {uid: $uid})<-%s-(related)
			WITH start, collect(DISTINCT related) as nodes, collect(DISTINCT r) as relationships
			UNWIND nodes as n
			UNWIND relationships as relList
			UNWIND relList as rel
			RETURN DISTINCT 
				n.uid as nodeUid, 
				n.name as nodeName, 
				n.namespace as nodeNamespace,
				labels(n) as nodeLabels,
				type(rel) as relType,
				startNode(rel).uid as relSource,
				endNode(rel).uid as relTarget
		`, relPattern)
	default: // both directions
		cypher = fmt.Sprintf(`
			MATCH (start {uid: $uid})-%s-(related)
			WITH start, collect(DISTINCT related) as nodes, collect(DISTINCT r) as relationships
			UNWIND nodes as n
			UNWIND relationships as relList
			UNWIND relList as rel
			RETURN DISTINCT 
				n.uid as nodeUid, 
				n.name as nodeName, 
				n.namespace as nodeNamespace,
				labels(n) as nodeLabels,
				type(rel) as relType,
				startNode(rel).uid as relSource,
				endNode(rel).uid as relTarget
		`, relPattern)
	}

	records, err := r.client.RunQuery(ctx, cypher, map[string]interface{}{
		"uid": nodeUID,
	})

	if err != nil {
		return nil, nil, fmt.Errorf("failed to query relations: %w", err)
	}

	// Deduplicate and convert
	nodeMap := make(map[string]*models.Node)
	edgeMap := make(map[string]*models.Edge)

	for _, record := range records {
		// Extract node
		if uid, ok := record["nodeUid"].(string); ok && uid != "" {
			if _, exists := nodeMap[uid]; !exists {
				node := &models.Node{ID: uid}
				if name, ok := record["nodeName"].(string); ok {
					node.Name = name
				}
				if namespace, ok := record["nodeNamespace"].(string); ok {
					node.Namespace = namespace
				}
				if labels, ok := record["nodeLabels"].([]interface{}); ok && len(labels) > 0 {
					node.Type = models.NodeType(labels[0].(string))
				}
				nodeMap[uid] = node
			}
		}

		// Extract edge
		if relType, ok := record["relType"].(string); ok {
			source, _ := record["relSource"].(string)
			target, _ := record["relTarget"].(string)
			edgeID := fmt.Sprintf("%s-%s-%s", source, relType, target)
			if _, exists := edgeMap[edgeID]; !exists {
				edgeMap[edgeID] = &models.Edge{
					ID:       edgeID,
					Type:     models.EdgeType(relType),
					SourceID: source,
					TargetID: target,
				}
			}
		}
	}

	// Convert maps to slices
	nodes := make([]*models.Node, 0, len(nodeMap))
	for _, node := range nodeMap {
		nodes = append(nodes, node)
	}

	edges := make([]*models.Edge, 0, len(edgeMap))
	for _, edge := range edgeMap {
		edges = append(edges, edge)
	}

	return nodes, edges, nil
}

// QueryPath finds the shortest path between two nodes
func (r *EdgeRepository) QueryPath(ctx context.Context, startUID, endUID string, maxDepth int) ([]*models.Node, []*models.Edge, error) {
	cypher := `
		MATCH path = shortestPath((start {uid: $startUid})-[*1..` + fmt.Sprintf("%d", maxDepth) + `]-(end {uid: $endUid}))
		UNWIND nodes(path) as n
		UNWIND relationships(path) as r
		RETURN DISTINCT
			n.uid as nodeUid,
			n.name as nodeName,
			n.namespace as nodeNamespace,
			labels(n) as nodeLabels,
			type(r) as relType,
			startNode(r).uid as relSource,
			endNode(r).uid as relTarget
	`

	records, err := r.client.RunQuery(ctx, cypher, map[string]interface{}{
		"startUid": startUID,
		"endUid":   endUID,
	})

	if err != nil {
		return nil, nil, fmt.Errorf("failed to query path: %w", err)
	}

	// Convert records to nodes and edges
	nodeMap := make(map[string]*models.Node)
	edgeMap := make(map[string]*models.Edge)

	for _, record := range records {
		if uid, ok := record["nodeUid"].(string); ok && uid != "" {
			if _, exists := nodeMap[uid]; !exists {
				node := &models.Node{ID: uid}
				if name, ok := record["nodeName"].(string); ok {
					node.Name = name
				}
				if namespace, ok := record["nodeNamespace"].(string); ok {
					node.Namespace = namespace
				}
				if labels, ok := record["nodeLabels"].([]interface{}); ok && len(labels) > 0 {
					node.Type = models.NodeType(labels[0].(string))
				}
				nodeMap[uid] = node
			}
		}

		if relType, ok := record["relType"].(string); ok {
			source, _ := record["relSource"].(string)
			target, _ := record["relTarget"].(string)
			edgeID := fmt.Sprintf("%s-%s-%s", source, relType, target)
			if _, exists := edgeMap[edgeID]; !exists {
				edgeMap[edgeID] = &models.Edge{
					ID:       edgeID,
					Type:     models.EdgeType(relType),
					SourceID: source,
					TargetID: target,
				}
			}
		}
	}

	nodes := make([]*models.Node, 0, len(nodeMap))
	for _, node := range nodeMap {
		nodes = append(nodes, node)
	}

	edges := make([]*models.Edge, 0, len(edgeMap))
	for _, edge := range edgeMap {
		edges = append(edges, edge)
	}

	return nodes, edges, nil
}

// recordToEdge converts a query record to an Edge model
func (r *EdgeRepository) recordToEdge(record map[string]interface{}) *models.Edge {
	edge := &models.Edge{}

	if id, ok := record["id"].(string); ok {
		edge.ID = id
	}
	if t, ok := record["type"].(string); ok {
		edge.Type = models.EdgeType(t)
	}
	if sourceId, ok := record["sourceId"].(string); ok {
		edge.SourceID = sourceId
	}
	if targetId, ok := record["targetId"].(string); ok {
		edge.TargetID = targetId
	}
	if props, ok := record["properties"].(map[string]interface{}); ok {
		edge.Properties = props
	}

	return edge
}
