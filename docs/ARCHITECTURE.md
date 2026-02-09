# Architecture Overview

## System Architecture

K8S Graph Explorer is designed as a modern, cloud-native application with a clear separation of concerns between the frontend, backend, and database layers.

```
┌──────────────────────────────────────────────────────────────────────┐
│                           User Browser                                │
└──────────────────────────────────┬───────────────────────────────────┘
                                   │
                                   ▼
┌──────────────────────────────────────────────────────────────────────┐
│                         Next.js Dashboard                             │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐       │
│  │   Graph View    │  │  Resource List  │  │    Topology     │       │
│  │   Component     │  │   Component     │  │   Component     │       │
│  └─────────────────┘  └─────────────────┘  └─────────────────┘       │
│                                                                       │
│  ┌─────────────────────────────────────────────────────────────┐     │
│  │                    Apollo GraphQL Client                     │     │
│  └─────────────────────────────────────────────────────────────┘     │
└──────────────────────────────────┬───────────────────────────────────┘
                                   │ GraphQL over HTTP
                                   ▼
┌──────────────────────────────────────────────────────────────────────┐
│                         Go Backend Server                             │
│  ┌─────────────────────────────────────────────────────────────┐     │
│  │                      GraphQL Handler                         │     │
│  │                      (gqlgen)                                │     │
│  └─────────────────────────────────────────────────────────────┘     │
│                                                                       │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐       │
│  │    Resolvers    │  │   K8s Client    │  │  Neo4j Client   │       │
│  └─────────────────┘  └─────────────────┘  └─────────────────┘       │
└───────────┬───────────────────────────────────────┬──────────────────┘
            │                                       │
            │ K8s API                               │ Bolt Protocol
            ▼                                       ▼
┌───────────────────────┐               ┌───────────────────────┐
│   Kubernetes Cluster  │               │        Neo4j          │
│  ┌─────────────────┐  │               │  ┌─────────────────┐  │
│  │   API Server    │  │               │  │   Graph Store   │  │
│  └─────────────────┘  │               │  └─────────────────┘  │
└───────────────────────┘               └───────────────────────┘
```

## Component Details

### Dashboard (Next.js)

The dashboard is a React-based single-page application built with Next.js 14. It provides:

- **Interactive Graph Visualization**: Using libraries like React Flow or D3.js
- **Real-time Updates**: WebSocket subscriptions for live resource changes
- **Search & Filtering**: Quick resource discovery
- **Responsive Design**: Works on desktop and tablet devices

### Backend (Go)

The backend is a stateless Go service that:

- **Exposes GraphQL API**: Using gqlgen for type-safe GraphQL
- **Connects to Kubernetes**: Using client-go for cluster access
- **Manages Graph Data**: Syncs K8s resources to Neo4j
- **Handles Authentication**: JWT-based auth (optional)

### Neo4j (Graph Database)

Neo4j stores the relationships between Kubernetes resources:

- **Nodes**: Each K8s resource is a node with properties
- **Relationships**: Connections like OWNS, SELECTS, EXPOSES
- **Cypher Queries**: Efficient graph traversals

## Data Flow

### Resource Sync Flow

```
1. Backend receives sync request
2. K8s Client lists resources from cluster
3. For each resource:
   a. Create/Update node in Neo4j
   b. Identify and create relationships
4. Return sync result to caller
```

### Query Flow

```
1. Dashboard sends GraphQL query
2. Backend resolves query
3. Neo4j returns graph data
4. Backend transforms to GraphQL response
5. Dashboard renders visualization
```

## Relationship Types

| Type | Description | Example |
|------|-------------|---------|
| OWNS | Owner reference | Deployment → ReplicaSet |
| SELECTS | Label selector | Service → Pod |
| MOUNTS | Volume mount | Pod → ConfigMap/Secret |
| EXPOSES | Network exposure | Ingress → Service |

## Security Considerations

- **RBAC**: Backend uses ServiceAccount with minimal permissions
- **Network Policies**: Restrict traffic between components
- **TLS**: All external traffic encrypted
- **Secrets**: Managed via K8s Secrets or external vault
