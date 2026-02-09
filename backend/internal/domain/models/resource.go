package models

import "time"

// K8sResource represents a Kubernetes resource entity
type K8sResource struct {
	UID         string            `json:"uid"`
	Name        string            `json:"name"`
	Namespace   string            `json:"namespace,omitempty"`
	Kind        string            `json:"kind"`
	APIVersion  string            `json:"apiVersion"`
	Labels      map[string]string `json:"labels,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
	CreatedAt   time.Time         `json:"createdAt"`
	Status      ResourceStatus    `json:"status,omitempty"`
	OwnerRefs   []OwnerReference  `json:"ownerReferences,omitempty"`
}

// OwnerReference represents an owner reference
type OwnerReference struct {
	UID        string `json:"uid"`
	Name       string `json:"name"`
	Kind       string `json:"kind"`
	APIVersion string `json:"apiVersion"`
}

// ResourceStatus represents the status of a resource
type ResourceStatus struct {
	Phase      string      `json:"phase,omitempty"`
	Ready      bool        `json:"ready"`
	Conditions []Condition `json:"conditions,omitempty"`
}

// Condition represents a condition on a resource
type Condition struct {
	Type               string    `json:"type"`
	Status             string    `json:"status"`
	Reason             string    `json:"reason,omitempty"`
	Message            string    `json:"message,omitempty"`
	LastTransitionTime time.Time `json:"lastTransitionTime,omitempty"`
}
