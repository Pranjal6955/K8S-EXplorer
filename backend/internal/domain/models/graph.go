package models

// Graph represents the graph visualization data
type Graph struct {
	Nodes    []GraphNode   `json:"nodes"`
	Edges    []GraphEdge   `json:"edges"`
	Metadata GraphMetadata `json:"metadata"`
}

// GraphNode represents a node in the graph visualization
type GraphNode struct {
	ID        string                 `json:"id"`
	Label     string                 `json:"label"`
	Kind      string                 `json:"kind"`
	Namespace string                 `json:"namespace,omitempty"`
	Data      map[string]interface{} `json:"data,omitempty"`
}

// GraphEdge represents an edge in the graph visualization
type GraphEdge struct {
	ID     string           `json:"id"`
	Source string           `json:"source"`
	Target string           `json:"target"`
	Type   RelationshipType `json:"type"`
	Label  string           `json:"label,omitempty"`
}

// GraphMetadata represents metadata about the graph
type GraphMetadata struct {
	NodeCount   int    `json:"nodeCount"`
	EdgeCount   int    `json:"edgeCount"`
	GeneratedAt string `json:"generatedAt"`
}

// RelationshipType represents types of relationships between resources
type RelationshipType string

const (
	RelationshipOwns         RelationshipType = "OWNS"
	RelationshipOwnedBy      RelationshipType = "OWNED_BY"
	RelationshipSelects      RelationshipType = "SELECTS"
	RelationshipSelectedBy   RelationshipType = "SELECTED_BY"
	RelationshipExposes      RelationshipType = "EXPOSES"
	RelationshipExposedBy    RelationshipType = "EXPOSED_BY"
	RelationshipMounts       RelationshipType = "MOUNTS"
	RelationshipMountedBy    RelationshipType = "MOUNTED_BY"
	RelationshipReferences   RelationshipType = "REFERENCES"
	RelationshipReferencedBy RelationshipType = "REFERENCED_BY"
)
