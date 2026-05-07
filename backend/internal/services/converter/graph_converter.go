package converter

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/K8S-Graph-Explorer/backend/internal/domain/models"
	"github.com/K8S-Graph-Explorer/backend/internal/infrastructure/kubernetes"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	k8sclient "k8s.io/client-go/kubernetes"
)

// GraphConverter converts Kubernetes resources to a normalized graph schema
type GraphConverter struct {
	clientset *k8sclient.Clientset
}

// NewGraphConverter creates a new GraphConverter
func NewGraphConverter(client *kubernetes.Client) *GraphConverter {
	return &GraphConverter{
		clientset: client.GetClientset(),
	}
}

// ConvertNamespace converts all resources in a namespace to a graph schema
func (c *GraphConverter) ConvertNamespace(ctx context.Context, namespace string) (*models.GraphSchema, error) {
	graph := models.NewGraphSchema()

	// Fetch all resources
	pods, err := c.clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list pods: %w", err)
	}

	services, err := c.clientset.CoreV1().Services(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list services: %w", err)
	}

	deployments, err := c.clientset.AppsV1().Deployments(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list deployments: %w", err)
	}

	replicaSets, err := c.clientset.AppsV1().ReplicaSets(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list replicasets: %w", err)
	}

	// Convert Deployments to nodes
	for _, deploy := range deployments.Items {
		graph.AddNode(c.deploymentToNode(&deploy))
	}

	// Convert ReplicaSets to nodes and create OWNS edges from Deployments
	rsMap := make(map[string]*appsv1.ReplicaSet)
	for i := range replicaSets.Items {
		rs := &replicaSets.Items[i]
		rsMap[string(rs.UID)] = rs

		// Create OWNS edge from Deployment to ReplicaSet
		for _, ownerRef := range rs.OwnerReferences {
			if ownerRef.Kind == "Deployment" {
				graph.AddEdge(models.Edge{
					ID:       fmt.Sprintf("%s-owns-%s", ownerRef.UID, rs.UID),
					Type:     models.EdgeTypeOwns,
					SourceID: string(ownerRef.UID),
					TargetID: string(rs.UID),
					Properties: map[string]interface{}{
						"controller": *ownerRef.Controller,
					},
				})
			}
		}
	}

	// Convert Pods to nodes and create edges
	for _, pod := range pods.Items {
		graph.AddNode(c.podToNode(&pod))

		// Create OWNS edge from ReplicaSet/Deployment to Pod
		for _, ownerRef := range pod.OwnerReferences {
			graph.AddEdge(models.Edge{
				ID:       fmt.Sprintf("%s-owns-%s", ownerRef.UID, pod.UID),
				Type:     models.EdgeTypeOwns,
				SourceID: string(ownerRef.UID),
				TargetID: string(pod.UID),
				Properties: map[string]interface{}{
					"ownerKind": ownerRef.Kind,
				},
			})
		}

		// Create DEPENDS_ON edges for volume mounts (ConfigMaps, Secrets, PVCs)
		c.addVolumeDependencies(graph, &pod)
	}

	// Convert Services to nodes and create EXPOSES edges
	for _, svc := range services.Items {
		graph.AddNode(c.serviceToNode(&svc))

		// Find pods that this service selects
		if svc.Spec.Selector != nil && len(svc.Spec.Selector) > 0 {
			selector := labels.SelectorFromSet(svc.Spec.Selector)
			for _, pod := range pods.Items {
				if selector.Matches(labels.Set(pod.Labels)) {
					graph.AddEdge(models.Edge{
						ID:       fmt.Sprintf("%s-exposes-%s", svc.UID, pod.UID),
						Type:     models.EdgeTypeExposes,
						SourceID: string(svc.UID),
						TargetID: string(pod.UID),
						Properties: map[string]interface{}{
							"ports": c.getServicePorts(&svc),
						},
					})
				}
			}
		}
	}

	return graph, nil
}

// ConvertAllNamespaces converts resources from all namespaces
func (c *GraphConverter) ConvertAllNamespaces(ctx context.Context) (*models.GraphSchema, error) {
	graph := models.NewGraphSchema()

	namespaces, err := c.clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list namespaces: %w", err)
	}

	for _, ns := range namespaces.Items {
		nsGraph, err := c.ConvertNamespace(ctx, ns.Name)
		if err != nil {
			// Log error but continue with other namespaces
			continue
		}

		// Merge graphs
		graph.Nodes = append(graph.Nodes, nsGraph.Nodes...)
		graph.Edges = append(graph.Edges, nsGraph.Edges...)
	}

	return graph, nil
}

// podToNode converts a Pod to a graph node
func (c *GraphConverter) podToNode(pod *corev1.Pod) models.Node {
	return models.Node{
		ID:        string(pod.UID),
		Type:      models.NodeTypePod,
		Name:      pod.Name,
		Namespace: pod.Namespace,
		Labels:    pod.Labels,
		Properties: map[string]interface{}{
			"phase":          string(pod.Status.Phase),
			"ready":          c.isPodReady(pod),
			"restarts":       c.getPodRestarts(pod),
			"nodeName":       pod.Spec.NodeName,
			"ip":             pod.Status.PodIP,
			"containers":     c.getContainerNames(pod),
			"createdAt":      pod.CreationTimestamp.Time,
			"serviceAccount": pod.Spec.ServiceAccountName,
		},
	}
}

// serviceToNode converts a Service to a graph node
func (c *GraphConverter) serviceToNode(svc *corev1.Service) models.Node {
	return models.Node{
		ID:        string(svc.UID),
		Type:      models.NodeTypeService,
		Name:      svc.Name,
		Namespace: svc.Namespace,
		Labels:    svc.Labels,
		Properties: map[string]interface{}{
			"type":        string(svc.Spec.Type),
			"clusterIP":   svc.Spec.ClusterIP,
			"ports":       c.getServicePorts(svc),
			"selector":    svc.Spec.Selector,
			"externalIPs": svc.Spec.ExternalIPs,
			"createdAt":   svc.CreationTimestamp.Time,
		},
	}
}

// deploymentToNode converts a Deployment to a graph node
func (c *GraphConverter) deploymentToNode(deploy *appsv1.Deployment) models.Node {
	return models.Node{
		ID:        string(deploy.UID),
		Type:      models.NodeTypeDeployment,
		Name:      deploy.Name,
		Namespace: deploy.Namespace,
		Labels:    deploy.Labels,
		Properties: map[string]interface{}{
			"replicas":          *deploy.Spec.Replicas,
			"readyReplicas":     deploy.Status.ReadyReplicas,
			"availableReplicas": deploy.Status.AvailableReplicas,
			"strategy":          string(deploy.Spec.Strategy.Type),
			"selector":          deploy.Spec.Selector.MatchLabels,
			"createdAt":         deploy.CreationTimestamp.Time,
		},
	}
}

// addVolumeDependencies adds DEPENDS_ON edges for pod volume dependencies
func (c *GraphConverter) addVolumeDependencies(graph *models.GraphSchema, pod *corev1.Pod) {
	for _, volume := range pod.Spec.Volumes {
		var targetID, targetKind string

		if volume.ConfigMap != nil {
			targetID = fmt.Sprintf("%s/%s/ConfigMap", pod.Namespace, volume.ConfigMap.Name)
			targetKind = "ConfigMap"
		} else if volume.Secret != nil {
			targetID = fmt.Sprintf("%s/%s/Secret", pod.Namespace, volume.Secret.SecretName)
			targetKind = "Secret"
		} else if volume.PersistentVolumeClaim != nil {
			targetID = fmt.Sprintf("%s/%s/PVC", pod.Namespace, volume.PersistentVolumeClaim.ClaimName)
			targetKind = "PersistentVolumeClaim"
		}

		if targetID != "" {
			graph.AddEdge(models.Edge{
				ID:       fmt.Sprintf("%s-depends-%s", pod.UID, targetID),
				Type:     models.EdgeTypeDependsOn,
				SourceID: string(pod.UID),
				TargetID: targetID,
				Properties: map[string]interface{}{
					"volumeName":   volume.Name,
					"resourceKind": targetKind,
				},
			})
		}
	}
}

// Helper functions

func (c *GraphConverter) isPodReady(pod *corev1.Pod) bool {
	for _, condition := range pod.Status.Conditions {
		if condition.Type == corev1.PodReady && condition.Status == corev1.ConditionTrue {
			return true
		}
	}
	return false
}

func (c *GraphConverter) getPodRestarts(pod *corev1.Pod) int32 {
	var restarts int32
	for _, cs := range pod.Status.ContainerStatuses {
		restarts += cs.RestartCount
	}
	return restarts
}

func (c *GraphConverter) getContainerNames(pod *corev1.Pod) []string {
	names := make([]string, len(pod.Spec.Containers))
	for i, container := range pod.Spec.Containers {
		names[i] = container.Name
	}
	return names
}

func (c *GraphConverter) getServicePorts(svc *corev1.Service) []map[string]interface{} {
	ports := make([]map[string]interface{}, len(svc.Spec.Ports))
	for i, port := range svc.Spec.Ports {
		ports[i] = map[string]interface{}{
			"name":       port.Name,
			"port":       port.Port,
			"targetPort": port.TargetPort.String(),
			"protocol":   string(port.Protocol),
		}
	}
	return ports
}

// ToJSON returns the graph schema as JSON
func (c *GraphConverter) ToJSON(graph *models.GraphSchema) ([]byte, error) {
	return json.MarshalIndent(graph, "", "  ")
}

// ToJSONString returns the graph schema as a JSON string
func (c *GraphConverter) ToJSONString(graph *models.GraphSchema) (string, error) {
	data, err := c.ToJSON(graph)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
