package converter

import (
	"encoding/json"

	"github.com/K8S-Graph-Explorer/backend/internal/domain/models"
)

// BuildGraphFromResources builds a graph from a list of resources and their relationships
func BuildGraphFromResources(resources []*models.K8sResource) *models.GraphSchema {
	graph := models.NewGraphSchema()
	nodeMap := make(map[string]bool)

	// First pass: create all nodes
	for _, resource := range resources {
		if nodeMap[resource.UID] {
			continue
		}

		node := models.Node{
			ID:        resource.UID,
			Type:      models.NodeType(resource.Kind),
			Name:      resource.Name,
			Namespace: resource.Namespace,
			Labels:    resource.Labels,
			Properties: map[string]interface{}{
				"apiVersion": resource.APIVersion,
				"createdAt":  resource.CreatedAt,
			},
		}

		// Add status properties if available
		if resource.Status.Phase != "" {
			node.Properties["phase"] = resource.Status.Phase
		}
		node.Properties["ready"] = resource.Status.Ready

		graph.AddNode(node)
		nodeMap[resource.UID] = true
	}

	// Second pass: create edges from owner references
	for _, resource := range resources {
		for _, ownerRef := range resource.OwnerRefs {
			graph.AddEdge(models.Edge{
				ID:       ownerRef.UID + "-owns-" + resource.UID,
				Type:     models.EdgeTypeOwns,
				SourceID: ownerRef.UID,
				TargetID: resource.UID,
				Properties: map[string]interface{}{
					"ownerKind": ownerRef.Kind,
				},
			})
		}
	}

	return graph
}

// GraphSchemaResponse represents the JSON response for graph schema
type GraphSchemaResponse struct {
	Nodes      []NodeResponse `json:"nodes"`
	Edges      []EdgeResponse `json:"edges"`
	Statistics Statistics     `json:"statistics"`
}

// NodeResponse represents a node in the JSON response
type NodeResponse struct {
	ID         string                 `json:"id"`
	Type       string                 `json:"type"`
	Name       string                 `json:"name"`
	Namespace  string                 `json:"namespace,omitempty"`
	Labels     map[string]string      `json:"labels,omitempty"`
	Properties map[string]interface{} `json:"properties,omitempty"`
}

// EdgeResponse represents an edge in the JSON response
type EdgeResponse struct {
	ID         string                 `json:"id"`
	Type       string                 `json:"type"`
	Source     string                 `json:"source"`
	Target     string                 `json:"target"`
	Properties map[string]interface{} `json:"properties,omitempty"`
}

// Statistics contains graph statistics
type Statistics struct {
	TotalNodes        int            `json:"totalNodes"`
	TotalEdges        int            `json:"totalEdges"`
	NodesByType       map[string]int `json:"nodesByType"`
	EdgesByType       map[string]int `json:"edgesByType"`
	NamespacesCovered []string       `json:"namespacesCovered"`
}

// ToResponse converts a GraphSchema to a response format with statistics
func ToResponse(graph *models.GraphSchema) *GraphSchemaResponse {
	response := &GraphSchemaResponse{
		Nodes: make([]NodeResponse, len(graph.Nodes)),
		Edges: make([]EdgeResponse, len(graph.Edges)),
		Statistics: Statistics{
			TotalNodes:        len(graph.Nodes),
			TotalEdges:        len(graph.Edges),
			NodesByType:       make(map[string]int),
			EdgesByType:       make(map[string]int),
			NamespacesCovered: []string{},
		},
	}

	namespaces := make(map[string]bool)

	// Convert nodes
	for i, node := range graph.Nodes {
		response.Nodes[i] = NodeResponse{
			ID:         node.ID,
			Type:       string(node.Type),
			Name:       node.Name,
			Namespace:  node.Namespace,
			Labels:     node.Labels,
			Properties: node.Properties,
		}

		// Update statistics
		response.Statistics.NodesByType[string(node.Type)]++
		if node.Namespace != "" {
			namespaces[node.Namespace] = true
		}
	}

	// Convert edges
	for i, edge := range graph.Edges {
		response.Edges[i] = EdgeResponse{
			ID:         edge.ID,
			Type:       string(edge.Type),
			Source:     edge.SourceID,
			Target:     edge.TargetID,
			Properties: edge.Properties,
		}

		// Update statistics
		response.Statistics.EdgesByType[string(edge.Type)]++
	}

	// Convert namespaces map to slice
	for ns := range namespaces {
		response.Statistics.NamespacesCovered = append(response.Statistics.NamespacesCovered, ns)
	}

	return response
}

// ToJSON converts the response to JSON bytes
func ToJSON(response *GraphSchemaResponse) ([]byte, error) {
	return json.MarshalIndent(response, "", "  ")
}

// ExampleOutput returns an example of the graph schema output
func ExampleOutput() string {
	return `{
  "nodes": [
    {
      "id": "abc-123",
      "type": "Deployment",
      "name": "nginx-deployment",
      "namespace": "default",
      "labels": {
        "app": "nginx"
      },
      "properties": {
        "replicas": 3,
        "readyReplicas": 3,
        "strategy": "RollingUpdate"
      }
    },
    {
      "id": "def-456",
      "type": "Pod",
      "name": "nginx-pod-1",
      "namespace": "default",
      "labels": {
        "app": "nginx"
      },
      "properties": {
        "phase": "Running",
        "ready": true,
        "nodeName": "worker-1"
      }
    },
    {
      "id": "ghi-789",
      "type": "Service",
      "name": "nginx-service",
      "namespace": "default",
      "labels": {
        "app": "nginx"
      },
      "properties": {
        "type": "ClusterIP",
        "clusterIP": "10.0.0.1",
        "ports": [{"port": 80, "targetPort": "8080"}]
      }
    }
  ],
  "edges": [
    {
      "id": "abc-123-owns-def-456",
      "type": "OWNS",
      "source": "abc-123",
      "target": "def-456",
      "properties": {
        "ownerKind": "ReplicaSet"
      }
    },
    {
      "id": "ghi-789-exposes-def-456",
      "type": "EXPOSES",
      "source": "ghi-789",
      "target": "def-456",
      "properties": {
        "ports": [{"port": 80}]
      }
    },
    {
      "id": "def-456-depends-configmap",
      "type": "DEPENDS_ON",
      "source": "def-456",
      "target": "default/nginx-config/ConfigMap",
      "properties": {
        "volumeName": "config",
        "resourceKind": "ConfigMap"
      }
    }
  ],
  "statistics": {
    "totalNodes": 3,
    "totalEdges": 3,
    "nodesByType": {
      "Deployment": 1,
      "Pod": 1,
      "Service": 1
    },
    "edgesByType": {
      "OWNS": 1,
      "EXPOSES": 1,
      "DEPENDS_ON": 1
    },
    "namespacesCovered": ["default"]
  }
}`
}
