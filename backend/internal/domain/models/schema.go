package models

// GraphSchema represents a normalized graph with nodes and edges
type GraphSchema struct {
	Nodes []Node `json:"nodes"`
	Edges []Edge `json:"edges"`
}

// Node represents a node in the graph
type Node struct {
	ID         string                 `json:"id"`
	Type       NodeType               `json:"type"`
	Name       string                 `json:"name"`
	Namespace  string                 `json:"namespace,omitempty"`
	Labels     map[string]string      `json:"labels,omitempty"`
	Properties map[string]interface{} `json:"properties,omitempty"`
}

// Edge represents an edge/relationship in the graph
type Edge struct {
	ID         string                 `json:"id"`
	Type       EdgeType               `json:"type"`
	SourceID   string                 `json:"sourceId"`
	TargetID   string                 `json:"targetId"`
	Properties map[string]interface{} `json:"properties,omitempty"`
}

// NodeType represents the type of a node
type NodeType string

const (
	NodeTypePod        NodeType = "Pod"
	NodeTypeService    NodeType = "Service"
	NodeTypeDeployment NodeType = "Deployment"
	NodeTypeReplicaSet NodeType = "ReplicaSet"
	NodeTypeConfigMap  NodeType = "ConfigMap"
	NodeTypeSecret     NodeType = "Secret"
	NodeTypePVC        NodeType = "PersistentVolumeClaim"
	NodeTypeNamespace  NodeType = "Namespace"
)

// EdgeType represents the type of an edge/relationship
type EdgeType string

const (
	EdgeTypeExposes   EdgeType = "EXPOSES"
	EdgeTypeOwns      EdgeType = "OWNS"
	EdgeTypeDependsOn EdgeType = "DEPENDS_ON"
	EdgeTypeSelects   EdgeType = "SELECTS"
	EdgeTypeMounts    EdgeType = "MOUNTS"
	EdgeTypeUses      EdgeType = "USES"
)

// NewGraphSchema creates a new empty graph schema
func NewGraphSchema() *GraphSchema {
	return &GraphSchema{
		Nodes: make([]Node, 0),
		Edges: make([]Edge, 0),
	}
}

// AddNode adds a node to the graph
func (g *GraphSchema) AddNode(node Node) {
	g.Nodes = append(g.Nodes, node)
}

// AddEdge adds an edge to the graph
func (g *GraphSchema) AddEdge(edge Edge) {
	g.Edges = append(g.Edges, edge)
}

// GetNodeByID returns a node by ID
func (g *GraphSchema) GetNodeByID(id string) *Node {
	for i := range g.Nodes {
		if g.Nodes[i].ID == id {
			return &g.Nodes[i]
		}
	}
	return nil
}

// HasNode checks if a node exists
func (g *GraphSchema) HasNode(id string) bool {
	return g.GetNodeByID(id) != nil
}

// HasEdge checks if an edge exists between two nodes
func (g *GraphSchema) HasEdge(sourceID, targetID string, edgeType EdgeType) bool {
	for _, edge := range g.Edges {
		if edge.SourceID == sourceID && edge.TargetID == targetID && edge.Type == edgeType {
			return true
		}
	}
	return false
}
