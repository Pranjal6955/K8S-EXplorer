package watcher

import (
	"context"
	"fmt"
	"sync"
	"time"

	"path/filepath"

	"github.com/K8S-Graph-Explorer/backend/internal/config"
	"github.com/K8S-Graph-Explorer/backend/internal/infrastructure/logger"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

// EventType represents the type of Kubernetes event
type EventType string

const (
	EventAdd    EventType = "ADD"
	EventUpdate EventType = "UPDATE"
	EventDelete EventType = "DELETE"
)

// ResourceEvent represents a Kubernetes resource event
type ResourceEvent struct {
	Type      EventType   `json:"type"`
	Kind      string      `json:"kind"`
	Name      string      `json:"name"`
	Namespace string      `json:"namespace"`
	UID       string      `json:"uid"`
	Timestamp time.Time   `json:"timestamp"`
	Object    interface{} `json:"-"`
}

// EventHandler is a function that handles resource events
type EventHandler func(event ResourceEvent)

// WatcherOptions configures the watcher
type WatcherOptions struct {
	// Namespaces to watch (empty means all namespaces)
	Namespaces []string
	// ResyncPeriod for informers
	ResyncPeriod time.Duration
	// WatchPods enables Pod watching
	WatchPods bool
	// WatchServices enables Service watching
	WatchServices bool
	// WatchDeployments enables Deployment watching
	WatchDeployments bool
	// WatchPVCs enables PVC watching
	WatchPVCs bool
}

// DefaultWatcherOptions returns default options
func DefaultWatcherOptions() WatcherOptions {
	return WatcherOptions{
		Namespaces:       []string{}, // All namespaces
		ResyncPeriod:     30 * time.Second,
		WatchPods:        true,
		WatchServices:    true,
		WatchDeployments: true,
		WatchPVCs:        true,
	}
}

// Watcher watches Kubernetes resources and emits events
type Watcher struct {
	clientset       *kubernetes.Clientset
	informerFactory informers.SharedInformerFactory
	logger          *logger.Logger
	handlers        []EventHandler
	handlersMu      sync.RWMutex
	stopCh          chan struct{}
	started         bool
	startedMu       sync.Mutex
	options         WatcherOptions
}

// NewWatcher creates a new Kubernetes resource watcher
func NewWatcher(cfg config.KubernetesConfig, log *logger.Logger) (*Watcher, error) {
	return NewWatcherWithOptions(cfg, log, DefaultWatcherOptions())
}

// NewWatcherWithOptions creates a new watcher with custom options
func NewWatcherWithOptions(cfg config.KubernetesConfig, log *logger.Logger, opts WatcherOptions) (*Watcher, error) {
	var k8sConfig *rest.Config
	var err error

	if cfg.InCluster {
		k8sConfig, err = rest.InClusterConfig()
		if err != nil {
			return nil, fmt.Errorf("failed to create in-cluster config: %w", err)
		}
		log.Info("Using in-cluster Kubernetes configuration")
	} else {
		kubeconfigPath := cfg.KubeconfigPath
		if kubeconfigPath == "" {
			if home := homedir.HomeDir(); home != "" {
				kubeconfigPath = filepath.Join(home, ".kube", "config")
			}
		}

		k8sConfig, err = clientcmd.BuildConfigFromFlags("", kubeconfigPath)
		if err != nil {
			return nil, fmt.Errorf("failed to build kubeconfig: %w", err)
		}
		log.Info("Using kubeconfig", "path", kubeconfigPath)
	}

	clientset, err := kubernetes.NewForConfig(k8sConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes clientset: %w", err)
	}

	// Create informer factory
	var informerFactory informers.SharedInformerFactory
	if len(opts.Namespaces) == 1 {
		// Watch specific namespace
		informerFactory = informers.NewSharedInformerFactoryWithOptions(
			clientset,
			opts.ResyncPeriod,
			informers.WithNamespace(opts.Namespaces[0]),
		)
	} else {
		// Watch all namespaces
		informerFactory = informers.NewSharedInformerFactory(clientset, opts.ResyncPeriod)
	}

	return &Watcher{
		clientset:       clientset,
		informerFactory: informerFactory,
		logger:          log,
		handlers:        make([]EventHandler, 0),
		stopCh:          make(chan struct{}),
		options:         opts,
	}, nil
}

// AddEventHandler adds an event handler
func (w *Watcher) AddEventHandler(handler EventHandler) {
	w.handlersMu.Lock()
	defer w.handlersMu.Unlock()
	w.handlers = append(w.handlers, handler)
}

// emit sends an event to all registered handlers
func (w *Watcher) emit(event ResourceEvent) {
	w.handlersMu.RLock()
	defer w.handlersMu.RUnlock()

	for _, handler := range w.handlers {
		go handler(event)
	}
}

// Start starts watching all configured resources
func (w *Watcher) Start(ctx context.Context) error {
	w.startedMu.Lock()
	if w.started {
		w.startedMu.Unlock()
		return fmt.Errorf("watcher already started")
	}
	w.started = true
	w.startedMu.Unlock()

	defer runtime.HandleCrash()

	w.logger.Info("Starting Kubernetes resource watcher...")

	var syncers []cache.InformerSynced

	// Set up informers based on options
	if w.options.WatchPods {
		w.setupPodInformer()
		syncers = append(syncers, w.informerFactory.Core().V1().Pods().Informer().HasSynced)
	}

	if w.options.WatchServices {
		w.setupServiceInformer()
		syncers = append(syncers, w.informerFactory.Core().V1().Services().Informer().HasSynced)
	}

	if w.options.WatchDeployments {
		w.setupDeploymentInformer()
		syncers = append(syncers, w.informerFactory.Apps().V1().Deployments().Informer().HasSynced)
	}

	if w.options.WatchPVCs {
		w.setupPVCInformer()
		syncers = append(syncers, w.informerFactory.Core().V1().PersistentVolumeClaims().Informer().HasSynced)
	}

	// Start the informer factory
	w.informerFactory.Start(w.stopCh)

	// Wait for cache sync
	w.logger.Info("Waiting for informer caches to sync...")
	if !cache.WaitForCacheSync(w.stopCh, syncers...) {
		return fmt.Errorf("failed to sync informer caches")
	}

	w.logger.Info("Informer caches synced, watching for events...")

	// Wait for context cancellation
	<-ctx.Done()
	w.Stop()

	return nil
}

// Stop stops the watcher
func (w *Watcher) Stop() {
	w.startedMu.Lock()
	defer w.startedMu.Unlock()

	if w.started {
		close(w.stopCh)
		w.started = false
		w.logger.Info("Kubernetes watcher stopped")
	}
}

// setupPodInformer sets up the Pod informer
func (w *Watcher) setupPodInformer() {
	podInformer := w.informerFactory.Core().V1().Pods().Informer()

	podInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			pod, ok := obj.(*corev1.Pod)
			if !ok {
				return
			}
			w.logger.Debug("Pod ADD", "name", pod.Name, "namespace", pod.Namespace)
			w.emit(w.createEvent(EventAdd, "Pod", pod.ObjectMeta, pod))
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			pod, ok := newObj.(*corev1.Pod)
			if !ok {
				return
			}
			w.logger.Debug("Pod UPDATE", "name", pod.Name, "namespace", pod.Namespace)
			w.emit(w.createEvent(EventUpdate, "Pod", pod.ObjectMeta, pod))
		},
		DeleteFunc: func(obj interface{}) {
			pod, ok := obj.(*corev1.Pod)
			if !ok {
				// Handle DeletedFinalStateUnknown
				tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
				if !ok {
					return
				}
				pod, ok = tombstone.Obj.(*corev1.Pod)
				if !ok {
					return
				}
			}
			w.logger.Debug("Pod DELETE", "name", pod.Name, "namespace", pod.Namespace)
			w.emit(w.createEvent(EventDelete, "Pod", pod.ObjectMeta, pod))
		},
	})
}

// setupServiceInformer sets up the Service informer
func (w *Watcher) setupServiceInformer() {
	serviceInformer := w.informerFactory.Core().V1().Services().Informer()

	serviceInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			svc, ok := obj.(*corev1.Service)
			if !ok {
				return
			}
			w.logger.Debug("Service ADD", "name", svc.Name, "namespace", svc.Namespace)
			w.emit(w.createEvent(EventAdd, "Service", svc.ObjectMeta, svc))
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			svc, ok := newObj.(*corev1.Service)
			if !ok {
				return
			}
			w.logger.Debug("Service UPDATE", "name", svc.Name, "namespace", svc.Namespace)
			w.emit(w.createEvent(EventUpdate, "Service", svc.ObjectMeta, svc))
		},
		DeleteFunc: func(obj interface{}) {
			svc, ok := obj.(*corev1.Service)
			if !ok {
				tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
				if !ok {
					return
				}
				svc, ok = tombstone.Obj.(*corev1.Service)
				if !ok {
					return
				}
			}
			w.logger.Debug("Service DELETE", "name", svc.Name, "namespace", svc.Namespace)
			w.emit(w.createEvent(EventDelete, "Service", svc.ObjectMeta, svc))
		},
	})
}

// setupDeploymentInformer sets up the Deployment informer
func (w *Watcher) setupDeploymentInformer() {
	deploymentInformer := w.informerFactory.Apps().V1().Deployments().Informer()

	deploymentInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			deploy, ok := obj.(*appsv1.Deployment)
			if !ok {
				return
			}
			w.logger.Debug("Deployment ADD", "name", deploy.Name, "namespace", deploy.Namespace)
			w.emit(w.createEvent(EventAdd, "Deployment", deploy.ObjectMeta, deploy))
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			deploy, ok := newObj.(*appsv1.Deployment)
			if !ok {
				return
			}
			w.logger.Debug("Deployment UPDATE", "name", deploy.Name, "namespace", deploy.Namespace)
			w.emit(w.createEvent(EventUpdate, "Deployment", deploy.ObjectMeta, deploy))
		},
		DeleteFunc: func(obj interface{}) {
			deploy, ok := obj.(*appsv1.Deployment)
			if !ok {
				tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
				if !ok {
					return
				}
				deploy, ok = tombstone.Obj.(*appsv1.Deployment)
				if !ok {
					return
				}
			}
			w.logger.Debug("Deployment DELETE", "name", deploy.Name, "namespace", deploy.Namespace)
			w.emit(w.createEvent(EventDelete, "Deployment", deploy.ObjectMeta, deploy))
		},
	})
}

// setupPVCInformer sets up the PersistentVolumeClaim informer
func (w *Watcher) setupPVCInformer() {
	pvcInformer := w.informerFactory.Core().V1().PersistentVolumeClaims().Informer()

	pvcInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			pvc, ok := obj.(*corev1.PersistentVolumeClaim)
			if !ok {
				return
			}
			w.logger.Debug("PVC ADD", "name", pvc.Name, "namespace", pvc.Namespace)
			w.emit(w.createEvent(EventAdd, "PersistentVolumeClaim", pvc.ObjectMeta, pvc))
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			pvc, ok := newObj.(*corev1.PersistentVolumeClaim)
			if !ok {
				return
			}
			w.logger.Debug("PVC UPDATE", "name", pvc.Name, "namespace", pvc.Namespace)
			w.emit(w.createEvent(EventUpdate, "PersistentVolumeClaim", pvc.ObjectMeta, pvc))
		},
		DeleteFunc: func(obj interface{}) {
			pvc, ok := obj.(*corev1.PersistentVolumeClaim)
			if !ok {
				tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
				if !ok {
					return
				}
				pvc, ok = tombstone.Obj.(*corev1.PersistentVolumeClaim)
				if !ok {
					return
				}
			}
			w.logger.Debug("PVC DELETE", "name", pvc.Name, "namespace", pvc.Namespace)
			w.emit(w.createEvent(EventDelete, "PersistentVolumeClaim", pvc.ObjectMeta, pvc))
		},
	})
}

// createEvent creates a ResourceEvent from Kubernetes object metadata
func (w *Watcher) createEvent(eventType EventType, kind string, meta metav1.ObjectMeta, obj interface{}) ResourceEvent {
	return ResourceEvent{
		Type:      eventType,
		Kind:      kind,
		Name:      meta.Name,
		Namespace: meta.Namespace,
		UID:       string(meta.UID),
		Timestamp: time.Now(),
		Object:    obj,
	}
}

// GetClientset returns the Kubernetes clientset
func (w *Watcher) GetClientset() *kubernetes.Clientset {
	return w.clientset
}

// IsStarted returns whether the watcher is running
func (w *Watcher) IsStarted() bool {
	w.startedMu.Lock()
	defer w.startedMu.Unlock()
	return w.started
}
