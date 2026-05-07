package models

// Cluster represents a Kubernetes cluster
type Cluster struct {
	Name       string        `json:"name"`
	Endpoint   string        `json:"endpoint,omitempty"`
	Version    string        `json:"version,omitempty"`
	NodeCount  int           `json:"nodeCount,omitempty"`
	Namespaces []K8sResource `json:"namespaces,omitempty"`
}

// Namespace represents a Kubernetes namespace
type Namespace struct {
	UID    string            `json:"uid"`
	Name   string            `json:"name"`
	Labels map[string]string `json:"labels,omitempty"`
	Status string            `json:"status,omitempty"`
}
