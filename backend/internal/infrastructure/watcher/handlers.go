package watcher

import (
	"fmt"
)

// Example demonstrates how to use the Kubernetes watcher
func Example() {
	/*
		// Example usage of the Kubernetes watcher

		// 1. Create configuration
		cfg := config.KubernetesConfig{
			InCluster:      false,
			KubeconfigPath: "", // Uses default ~/.kube/config
		}

		// 2. Create logger
		log, _ := logger.NewLogger("development")

		// 3. Create watcher with default options (watches all resources)
		watcher, err := NewWatcher(cfg, log)
		if err != nil {
			log.Fatal("Failed to create watcher", "error", err)
		}

		// 4. Or create with custom options
		opts := WatcherOptions{
			Namespaces:       []string{"default", "kube-system"},
			ResyncPeriod:     1 * time.Minute,
			WatchPods:        true,
			WatchServices:    true,
			WatchDeployments: true,
			WatchPVCs:        false, // Don't watch PVCs
		}
		watcher, err = NewWatcherWithOptions(cfg, log, opts)

		// 5. Add event handlers
		watcher.AddEventHandler(func(event ResourceEvent) {
			fmt.Printf("[%s] %s: %s/%s (UID: %s)\n",
				event.Type,
				event.Kind,
				event.Namespace,
				event.Name,
				event.UID,
			)
		})

		// 6. You can add multiple handlers
		watcher.AddEventHandler(func(event ResourceEvent) {
			// Log to database, send webhook, etc.
			switch event.Type {
			case EventAdd:
				// Handle add
			case EventUpdate:
				// Handle update
			case EventDelete:
				// Handle delete
			}
		})

		// 7. Start the watcher (blocking)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		go func() {
			if err := watcher.Start(ctx); err != nil {
				log.Error("Watcher error", "error", err)
			}
		}()

		// 8. The watcher runs until context is cancelled
		// To stop: cancel()
	*/
}

// PrintEventHandler returns a handler that prints events
func PrintEventHandler() EventHandler {
	return func(event ResourceEvent) {
		fmt.Printf("[%s] %s %s/%s (UID: %s) at %s\n",
			event.Type,
			event.Kind,
			event.Namespace,
			event.Name,
			event.UID,
			event.Timestamp.Format("15:04:05"),
		)
	}
}

// LoggingEventHandler returns a handler that logs events
func LoggingEventHandler(log interface {
	Info(msg string, keysAndValues ...interface{})
}) EventHandler {
	return func(event ResourceEvent) {
		log.Info("Kubernetes resource event",
			"type", event.Type,
			"kind", event.Kind,
			"namespace", event.Namespace,
			"name", event.Name,
			"uid", event.UID,
		)
	}
}

// ChannelEventHandler returns a handler that sends events to a channel
func ChannelEventHandler(ch chan<- ResourceEvent) EventHandler {
	return func(event ResourceEvent) {
		select {
		case ch <- event:
		default:
			// Channel full, drop event
		}
	}
}

// FilteredEventHandler wraps a handler with a filter
func FilteredEventHandler(handler EventHandler, filter func(event ResourceEvent) bool) EventHandler {
	return func(event ResourceEvent) {
		if filter(event) {
			handler(event)
		}
	}
}

// KindFilter returns a filter that only passes specific kinds
func KindFilter(kinds ...string) func(event ResourceEvent) bool {
	kindSet := make(map[string]struct{}, len(kinds))
	for _, k := range kinds {
		kindSet[k] = struct{}{}
	}
	return func(event ResourceEvent) bool {
		_, ok := kindSet[event.Kind]
		return ok
	}
}

// NamespaceFilter returns a filter that only passes specific namespaces
func NamespaceFilter(namespaces ...string) func(event ResourceEvent) bool {
	nsSet := make(map[string]struct{}, len(namespaces))
	for _, ns := range namespaces {
		nsSet[ns] = struct{}{}
	}
	return func(event ResourceEvent) bool {
		_, ok := nsSet[event.Namespace]
		return ok
	}
}

// EventTypeFilter returns a filter that only passes specific event types
func EventTypeFilter(types ...EventType) func(event ResourceEvent) bool {
	typeSet := make(map[EventType]struct{}, len(types))
	for _, t := range types {
		typeSet[t] = struct{}{}
	}
	return func(event ResourceEvent) bool {
		_, ok := typeSet[event.Type]
		return ok
	}
}
