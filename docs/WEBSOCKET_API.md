# WebSocket API

This document describes the WebSocket API for receiving real-time topology updates from the backend.

## Connection

Connect to the WebSocket endpoint at:
`ws://localhost:8080/ws`

## Message Format

All messages sent from the server are JSON objects with the following structure:

```json
{
  "type": "MESSAGE_TYPE",
  "payload": { ... }
}
```

## Message Types

### TOPOLOGY_UPDATE

Sent when a namespace synchronization process completes (e.g., triggered manually or scheduled).

**Payload:**

```json
{
  "namespace": "string",
  "syncedCount": "number",
  "timestamp": "ISO8601 string"
}
```

### TOPOLOGY_UPDATE_ALL

Sent when a full synchronization of all namespaces completes.

**Payload:**

```json
{
  "syncedCount": "number",
  "timestamp": "ISO8601 string"
}
```

### RESOURCE_EVENT

Sent in real-time when a Kubernetes resource is added, updated, or deleted. This event reflects immediate changes detected by the K8s watcher.

**Payload:**

```json
{
  "eventType": "ADDED | MODIFIED | DELETED",
  "kind": "string (e.g. Pod, Deployment)",
  "name": "string",
  "namespace": "string",
  "uid": "string",
  "resource": "Object (The full K8s resource model)",
  "timestamp": "ISO8601 string"
}
```
**Note**: For `DELETED` events, the `resource` field might be partial or contain only metadata necessary to identify the deleted entity.

## Usage Example (Client-side)
```javascript
const ws = new WebSocket('ws://localhost:8080/ws');

ws.onopen = () => {
    console.log('Connected to WebSocket');
};

ws.onmessage = (event) => {
    const message = JSON.parse(event.data);
    switch (message.type) {
        case 'TOPOLOGY_UPDATE':
            console.log('Topology updated for namespace:', message.payload.namespace);
            // Refresh graph or fetch new data
            break;
        case 'RESOURCE_EVENT':
            console.log('Real-time update:', message.payload.kind, message.payload.name);
            // Update specific node in graph
            break;
    }
};
```
