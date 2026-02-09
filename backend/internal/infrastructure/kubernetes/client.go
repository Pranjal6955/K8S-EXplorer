package kubernetes

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/K8S-Graph-Explorer/backend/internal/config"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// Client wraps the Kubernetes clientset
type Client struct {
	clientset *kubernetes.Clientset
}

// NewClient creates a new Kubernetes client
func NewClient(cfg config.KubernetesConfig) (*Client, error) {
	var k8sConfig *rest.Config
	var err error

	if cfg.InCluster {
		// Use in-cluster configuration
		k8sConfig, err = rest.InClusterConfig()
		if err != nil {
			return nil, fmt.Errorf("failed to create in-cluster config: %w", err)
		}
	} else {
		// Use kubeconfig file
		kubeconfigPath := cfg.KubeconfigPath
		if kubeconfigPath == "" {
			homeDir, _ := os.UserHomeDir()
			kubeconfigPath = filepath.Join(homeDir, ".kube", "config")
		}

		k8sConfig, err = clientcmd.BuildConfigFromFlags("", kubeconfigPath)
		if err != nil {
			return nil, fmt.Errorf("failed to build kubeconfig: %w", err)
		}
	}

	clientset, err := kubernetes.NewForConfig(k8sConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes clientset: %w", err)
	}

	return &Client{clientset: clientset}, nil
}

// GetClientset returns the underlying Kubernetes clientset
func (c *Client) GetClientset() *kubernetes.Clientset {
	return c.clientset
}

// Resource represents a Kubernetes resource
type Resource struct {
	UID        string
	Name       string
	Namespace  string
	Kind       string
	APIVersion string
	Labels     map[string]string
	OwnerRefs  []OwnerReference
}

// OwnerReference represents an owner reference
type OwnerReference struct {
	UID        string
	Name       string
	Kind       string
	APIVersion string
}

// ListNamespaces lists all namespaces
func (c *Client) ListNamespaces(ctx context.Context) ([]Resource, error) {
	namespaces, err := c.clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list namespaces: %w", err)
	}

	var resources []Resource
	for _, ns := range namespaces.Items {
		resources = append(resources, Resource{
			UID:        string(ns.UID),
			Name:       ns.Name,
			Namespace:  "",
			Kind:       "Namespace",
			APIVersion: "v1",
			Labels:     ns.Labels,
		})
	}

	return resources, nil
}

// ListPods lists pods in a namespace
func (c *Client) ListPods(ctx context.Context, namespace string) ([]Resource, error) {
	pods, err := c.clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list pods: %w", err)
	}

	var resources []Resource
	for _, pod := range pods.Items {
		var ownerRefs []OwnerReference
		for _, ref := range pod.OwnerReferences {
			ownerRefs = append(ownerRefs, OwnerReference{
				UID:        string(ref.UID),
				Name:       ref.Name,
				Kind:       ref.Kind,
				APIVersion: ref.APIVersion,
			})
		}

		resources = append(resources, Resource{
			UID:        string(pod.UID),
			Name:       pod.Name,
			Namespace:  pod.Namespace,
			Kind:       "Pod",
			APIVersion: "v1",
			Labels:     pod.Labels,
			OwnerRefs:  ownerRefs,
		})
	}

	return resources, nil
}

// ListDeployments lists deployments in a namespace
func (c *Client) ListDeployments(ctx context.Context, namespace string) ([]Resource, error) {
	deployments, err := c.clientset.AppsV1().Deployments(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list deployments: %w", err)
	}

	var resources []Resource
	for _, deployment := range deployments.Items {
		resources = append(resources, Resource{
			UID:        string(deployment.UID),
			Name:       deployment.Name,
			Namespace:  deployment.Namespace,
			Kind:       "Deployment",
			APIVersion: "apps/v1",
			Labels:     deployment.Labels,
		})
	}

	return resources, nil
}

// ListServices lists services in a namespace
func (c *Client) ListServices(ctx context.Context, namespace string) ([]Resource, error) {
	services, err := c.clientset.CoreV1().Services(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list services: %w", err)
	}

	var resources []Resource
	for _, svc := range services.Items {
		resources = append(resources, Resource{
			UID:        string(svc.UID),
			Name:       svc.Name,
			Namespace:  svc.Namespace,
			Kind:       "Service",
			APIVersion: "v1",
			Labels:     svc.Labels,
		})
	}

	return resources, nil
}

// ListReplicaSets lists ReplicaSets in a namespace
func (c *Client) ListReplicaSets(ctx context.Context, namespace string) ([]Resource, error) {
	replicaSets, err := c.clientset.AppsV1().ReplicaSets(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list replicasets: %w", err)
	}

	var resources []Resource
	for _, rs := range replicaSets.Items {
		var ownerRefs []OwnerReference
		for _, ref := range rs.OwnerReferences {
			ownerRefs = append(ownerRefs, OwnerReference{
				UID:        string(ref.UID),
				Name:       ref.Name,
				Kind:       ref.Kind,
				APIVersion: ref.APIVersion,
			})
		}

		resources = append(resources, Resource{
			UID:        string(rs.UID),
			Name:       rs.Name,
			Namespace:  rs.Namespace,
			Kind:       "ReplicaSet",
			APIVersion: "apps/v1",
			Labels:     rs.Labels,
			OwnerRefs:  ownerRefs,
		})
	}

	return resources, nil
}

// ListConfigMaps lists ConfigMaps in a namespace
func (c *Client) ListConfigMaps(ctx context.Context, namespace string) ([]Resource, error) {
	configMaps, err := c.clientset.CoreV1().ConfigMaps(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list configmaps: %w", err)
	}

	var resources []Resource
	for _, cm := range configMaps.Items {
		resources = append(resources, Resource{
			UID:        string(cm.UID),
			Name:       cm.Name,
			Namespace:  cm.Namespace,
			Kind:       "ConfigMap",
			APIVersion: "v1",
			Labels:     cm.Labels,
		})
	}

	return resources, nil
}
