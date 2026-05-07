package watcher

import (
	"context"
	"time"

	"github.com/K8S-Graph-Explorer/backend/internal/domain/models"
	"github.com/K8S-Graph-Explorer/backend/internal/domain/repositories"
	"github.com/K8S-Graph-Explorer/backend/internal/infrastructure/logger"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Broadcaster defines the interface for broadcasting messages
type Broadcaster interface {
	Broadcast(message interface{})
}

// EventProcessor processes Kubernetes events and syncs them to the database
type EventProcessor struct {
	watcher      *Watcher
	resourceRepo repositories.ResourceRepository
	logger       *logger.Logger
	broadcaster  Broadcaster
	eventChan    chan ResourceEvent
	workers      int
}

// NewEventProcessor creates a new event processor
func NewEventProcessor(
	watcher *Watcher,
	resourceRepo repositories.ResourceRepository,
	broadcaster Broadcaster,
	log *logger.Logger,
	workers int,
) *EventProcessor {
	if workers <= 0 {
		workers = 5
	}

	return &EventProcessor{
		watcher:      watcher,
		resourceRepo: resourceRepo,
		broadcaster:  broadcaster,
		logger:       log,
		eventChan:    make(chan ResourceEvent, 1000),
		workers:      workers,
	}
}

// Start starts the event processor
func (p *EventProcessor) Start(ctx context.Context) {
	// Register event handler
	p.watcher.AddEventHandler(func(event ResourceEvent) {
		select {
		case p.eventChan <- event:
		default:
			p.logger.Warn("Event channel full, dropping event",
				"kind", event.Kind,
				"name", event.Name,
				"type", event.Type,
			)
		}
	})

	// Start worker goroutines
	for i := 0; i < p.workers; i++ {
		go p.worker(ctx, i)
	}

	p.logger.Info("Event processor started", "workers", p.workers)
}

// worker processes events from the channel
func (p *EventProcessor) worker(ctx context.Context, id int) {
	for {
		select {
		case <-ctx.Done():
			p.logger.Debug("Worker stopping", "id", id)
			return
		case event := <-p.eventChan:
			if err := p.processEvent(ctx, event); err != nil {
				p.logger.Error("Failed to process event",
					"error", err,
					"kind", event.Kind,
					"name", event.Name,
					"type", event.Type,
				)
			}
		}
	}
}

// processEvent processes a single event
func (p *EventProcessor) processEvent(ctx context.Context, event ResourceEvent) error {
	var err error
	var resource *models.K8sResource

	switch event.Type {
	case EventAdd, EventUpdate:
		resource = p.convertToResource(event)
		if resource == nil {
			return nil
		}
		err = p.resourceRepo.Create(ctx, resource)

	case EventDelete:
		resource = &models.K8sResource{UID: event.UID, Name: event.Name, Kind: event.Kind, Namespace: event.Namespace}
		err = p.resourceRepo.Delete(ctx, event.UID)
	}

	if err != nil {
		return err
	}

	// Broadcast event
	if p.broadcaster != nil {
		p.broadcaster.Broadcast(map[string]interface{}{
			"type": "RESOURCE_EVENT",
			"payload": map[string]interface{}{
				"eventType": event.Type,
				"kind":      event.Kind,
				"name":      event.Name,
				"namespace": event.Namespace,
				"uid":       event.UID,
				"resource":  resource, // Might be nil or partial on delete
				"timestamp": time.Now(),
			},
		})
	}

	return nil
}

// convertToResource converts a ResourceEvent to a domain Resource
func (p *EventProcessor) convertToResource(event ResourceEvent) *models.K8sResource {
	resource := &models.K8sResource{
		UID:       event.UID,
		Name:      event.Name,
		Namespace: event.Namespace,
		Kind:      event.Kind,
		CreatedAt: time.Now(),
	}

	switch event.Kind {
	case "Pod":
		if pod, ok := event.Object.(*corev1.Pod); ok {
			resource.APIVersion = "v1"
			resource.Labels = pod.Labels
			resource.Annotations = pod.Annotations
			resource.Status = models.ResourceStatus{
				Phase: string(pod.Status.Phase),
				Ready: isPodReady(pod),
			}
			resource.OwnerRefs = convertOwnerRefs(pod.OwnerReferences)
		}

	case "Service":
		if svc, ok := event.Object.(*corev1.Service); ok {
			resource.APIVersion = "v1"
			resource.Labels = svc.Labels
			resource.Annotations = svc.Annotations
		}

	case "Deployment":
		if deploy, ok := event.Object.(*appsv1.Deployment); ok {
			resource.APIVersion = "apps/v1"
			resource.Labels = deploy.Labels
			resource.Annotations = deploy.Annotations
			resource.Status = models.ResourceStatus{
				Ready: isDeploymentReady(deploy),
			}
		}

	case "PersistentVolumeClaim":
		if pvc, ok := event.Object.(*corev1.PersistentVolumeClaim); ok {
			resource.APIVersion = "v1"
			resource.Labels = pvc.Labels
			resource.Annotations = pvc.Annotations
			resource.Status = models.ResourceStatus{
				Phase: string(pvc.Status.Phase),
			}
		}
	}

	return resource
}

// isPodReady checks if a pod is ready
func isPodReady(pod *corev1.Pod) bool {
	for _, condition := range pod.Status.Conditions {
		if condition.Type == corev1.PodReady && condition.Status == corev1.ConditionTrue {
			return true
		}
	}
	return false
}

// isDeploymentReady checks if a deployment is ready
func isDeploymentReady(deploy *appsv1.Deployment) bool {
	return deploy.Status.ReadyReplicas == deploy.Status.Replicas
}

// convertOwnerRefs converts Kubernetes owner references to domain model
func convertOwnerRefs(refs []metav1.OwnerReference) []models.OwnerReference {
	result := make([]models.OwnerReference, len(refs))
	for i, ref := range refs {
		result[i] = models.OwnerReference{
			UID:        string(ref.UID),
			Name:       ref.Name,
			Kind:       ref.Kind,
			APIVersion: ref.APIVersion,
		}
	}
	return result
}
