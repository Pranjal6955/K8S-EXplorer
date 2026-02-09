package model

import (
	"time"
)

// Node represents a Kubernetes resource in the graph
type Node struct {
	ID         string                 `json:"id"`
	Type       NodeType               `json:"type"`
	Name       string                 `json:"name"`
	Namespace  *string                `json:"namespace,omitempty"`
	Labels     map[string]interface{} `json:"labels,omitempty"`
	Properties map[string]interface{} `json:"properties,omitempty"`
	CreatedAt  *time.Time             `json:"createdAt,omitempty"`
	Edges      []*Edge                `json:"edges,omitempty"`
}

// Edge represents a relationship between two nodes
type Edge struct {
	ID         string                 `json:"id"`
	Type       EdgeType               `json:"type"`
	SourceID   string                 `json:"sourceId"`
	TargetID   string                 `json:"targetId"`
	Source     *Node                  `json:"source,omitempty"`
	Target     *Node                  `json:"target,omitempty"`
	Properties map[string]interface{} `json:"properties,omitempty"`
}

// Topology represents the complete graph structure
type Topology struct {
	Nodes      []*Node             `json:"nodes"`
	Edges      []*Edge             `json:"edges"`
	Statistics *TopologyStatistics `json:"statistics"`
}

// TopologyStatistics contains summary information
type TopologyStatistics struct {
	TotalNodes  int                    `json:"totalNodes"`
	TotalEdges  int                    `json:"totalEdges"`
	NodesByType map[string]interface{} `json:"nodesByType"`
	EdgesByType map[string]interface{} `json:"edgesByType"`
	Namespaces  []string               `json:"namespaces"`
}

// DependencyGraph represents dependencies for a resource
type DependencyGraph struct {
	Root         *Node   `json:"root"`
	Dependencies []*Node `json:"dependencies"`
	Edges        []*Edge `json:"edges"`
	Depth        int     `json:"depth"`
}

// TopologyFilter for filtering topology queries
type TopologyFilter struct {
	Namespace *string    `json:"namespace,omitempty"`
	NodeTypes []NodeType `json:"nodeTypes,omitempty"`
	EdgeTypes []EdgeType `json:"edgeTypes,omitempty"`
	Limit     *int       `json:"limit,omitempty"`
}

// DependencyFilter for filtering dependency queries
type DependencyFilter struct {
	Direction *DependencyDirection `json:"direction,omitempty"`
	Depth     *int                 `json:"depth,omitempty"`
	EdgeTypes []EdgeType           `json:"edgeTypes,omitempty"`
}

// NodeInput for creating/updating nodes
type NodeInput struct {
	ID         string                 `json:"id"`
	Type       NodeType               `json:"type"`
	Name       string                 `json:"name"`
	Namespace  *string                `json:"namespace,omitempty"`
	Labels     map[string]interface{} `json:"labels,omitempty"`
	Properties map[string]interface{} `json:"properties,omitempty"`
}

// EdgeInput for creating/updating edges
type EdgeInput struct {
	ID         *string                `json:"id,omitempty"`
	Type       EdgeType               `json:"type"`
	SourceID   string                 `json:"sourceId"`
	TargetID   string                 `json:"targetId"`
	Properties map[string]interface{} `json:"properties,omitempty"`
}

// SyncResult contains the result of a sync operation
type SyncResult struct {
	Success      bool     `json:"success"`
	Message      string   `json:"message"`
	NodesCreated int      `json:"nodesCreated"`
	EdgesCreated int      `json:"edgesCreated"`
	Errors       []string `json:"errors,omitempty"`
}

// TopologyEvent represents a change in the topology
type TopologyEvent struct {
	Type          EventType `json:"type"`
	Timestamp     time.Time `json:"timestamp"`
	AffectedNodes []*Node   `json:"affectedNodes"`
	AffectedEdges []*Edge   `json:"affectedEdges"`
}

// NodeEvent represents a change to a single node
type NodeEvent struct {
	Type      EventType `json:"type"`
	Timestamp time.Time `json:"timestamp"`
	Node      *Node     `json:"node"`
}

// NodeType represents the type of a Kubernetes resource
type NodeType string

const (
	NodeTypePod                   NodeType = "Pod"
	NodeTypeService               NodeType = "Service"
	NodeTypeDeployment            NodeType = "Deployment"
	NodeTypeReplicaSet            NodeType = "ReplicaSet"
	NodeTypeConfigMap             NodeType = "ConfigMap"
	NodeTypeSecret                NodeType = "Secret"
	NodeTypePersistentVolumeClaim NodeType = "PersistentVolumeClaim"
	NodeTypeNamespace             NodeType = "Namespace"
	NodeTypeIngress               NodeType = "Ingress"
	NodeTypeStatefulSet           NodeType = "StatefulSet"
	NodeTypeDaemonSet             NodeType = "DaemonSet"
	NodeTypeJob                   NodeType = "Job"
	NodeTypeCronJob               NodeType = "CronJob"
)

// EdgeType represents the type of relationship
type EdgeType string

const (
	EdgeTypeOwns      EdgeType = "OWNS"
	EdgeTypeExposes   EdgeType = "EXPOSES"
	EdgeTypeDependsOn EdgeType = "DEPENDS_ON"
	EdgeTypeSelects   EdgeType = "SELECTS"
	EdgeTypeMounts    EdgeType = "MOUNTS"
	EdgeTypeUses      EdgeType = "USES"
	EdgeTypeRoutesTo  EdgeType = "ROUTES_TO"
)

// DependencyDirection specifies the direction for dependency traversal
type DependencyDirection string

const (
	DependencyDirectionUpstream   DependencyDirection = "UPSTREAM"
	DependencyDirectionDownstream DependencyDirection = "DOWNSTREAM"
	DependencyDirectionBoth       DependencyDirection = "BOTH"
)

// EventType for subscription events
type EventType string

const (
	EventTypeCreated EventType = "CREATED"
	EventTypeUpdated EventType = "UPDATED"
	EventTypeDeleted EventType = "DELETED"
)

// IsValid checks if NodeType is valid
func (e NodeType) IsValid() bool {
	switch e {
	case NodeTypePod, NodeTypeService, NodeTypeDeployment, NodeTypeReplicaSet,
		NodeTypeConfigMap, NodeTypeSecret, NodeTypePersistentVolumeClaim,
		NodeTypeNamespace, NodeTypeIngress, NodeTypeStatefulSet, NodeTypeDaemonSet,
		NodeTypeJob, NodeTypeCronJob:
		return true
	}
	return false
}

// String returns the string representation
func (e NodeType) String() string {
	return string(e)
}

// IsValid checks if EdgeType is valid
func (e EdgeType) IsValid() bool {
	switch e {
	case EdgeTypeOwns, EdgeTypeExposes, EdgeTypeDependsOn, EdgeTypeSelects,
		EdgeTypeMounts, EdgeTypeUses, EdgeTypeRoutesTo:
		return true
	}
	return false
}

// String returns the string representation
func (e EdgeType) String() string {
	return string(e)
}
