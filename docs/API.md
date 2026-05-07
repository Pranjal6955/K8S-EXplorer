# GraphQL API Reference

This document provides a complete reference for the K8S Graph Explorer GraphQL API.

## Endpoint

- **Development**: `http://localhost:8080/graphql`
- **Playground**: `http://localhost:8080/`

## Authentication

Currently, the API does not require authentication. In production, implement JWT-based authentication.

## Types

### K8sResource

Represents a Kubernetes resource node in the graph.

```graphql
type K8sResource {
  uid: ID!
  name: String!
  namespace: String
  kind: String!
  apiVersion: String!
  labels: [Label!]
  annotations: [Annotation!]
  owners: [K8sResource!]
  owned: [K8sResource!]
  relatedTo: [Relationship!]
  status: ResourceStatus
  createdAt: String
}
```

| Field | Type | Description |
|-------|------|-------------|
| `uid` | ID! | Unique identifier |
| `name` | String! | Resource name |
| `namespace` | String | Namespace (null for cluster-scoped) |
| `kind` | String! | Resource kind (Pod, Deployment, etc.) |
| `apiVersion` | String! | API version (v1, apps/v1, etc.) |
| `labels` | [Label!] | Key-value labels |
| `annotations` | [Annotation!] | Key-value annotations |
| `owners` | [K8sResource!] | Owner resources |
| `owned` | [K8sResource!] | Owned resources |
| `relatedTo` | [Relationship!] | Related resources |
| `status` | ResourceStatus | Current status |
| `createdAt` | String | Creation timestamp |

### Graph

Represents the graph visualization data.

```graphql
type Graph {
  nodes: [GraphNode!]!
  edges: [GraphEdge!]!
  metadata: GraphMetadata
}

type GraphNode {
  id: ID!
  label: String!
  kind: String!
  namespace: String
  data: K8sResource!
}

type GraphEdge {
  id: ID!
  source: ID!
  target: ID!
  type: RelationshipType!
  label: String
}
```

### RelationshipType

```graphql
enum RelationshipType {
  OWNS
  OWNED_BY
  SELECTS
  SELECTED_BY
  EXPOSES
  EXPOSED_BY
  MOUNTS
  MOUNTED_BY
  REFERENCES
  REFERENCED_BY
}
```

## Queries

### Get Single Resource

```graphql
query GetResource($uid: ID!) {
  resource(uid: $uid) {
    uid
    name
    namespace
    kind
    labels {
      key
      value
    }
    owners {
      uid
      name
      kind
    }
  }
}
```

**Variables:**
```json
{
  "uid": "abc-123-def"
}
```

### List Resources

```graphql
query ListResources($filter: ResourceFilter) {
  resources(filter: $filter) {
    uid
    name
    namespace
    kind
    apiVersion
  }
}
```

**Variables:**
```json
{
  "filter": {
    "namespace": "default",
    "kinds": ["Pod", "Deployment"],
    "allNamespaces": false
  }
}
```

### Get Cluster Info

```graphql
query GetCluster {
  cluster {
    name
    version
    nodeCount
    namespaces {
      uid
      name
    }
  }
}
```

### Get Graph

```graphql
query GetGraph($filter: ResourceFilter) {
  graph(filter: $filter) {
    nodes {
      id
      label
      kind
      namespace
    }
    edges {
      id
      source
      target
      type
    }
    metadata {
      nodeCount
      edgeCount
      generatedAt
    }
  }
}
```

**Variables:**
```json
{
  "filter": {
    "namespace": "kube-system"
  }
}
```

### List Namespaces

```graphql
query ListNamespaces {
  namespaces {
    uid
    name
    labels {
      key
      value
    }
  }
}
```

### Search Resources

```graphql
query SearchResources($query: String!, $filter: ResourceFilter) {
  search(query: $query, filter: $filter) {
    uid
    name
    namespace
    kind
  }
}
```

**Variables:**
```json
{
  "query": "nginx",
  "filter": {
    "allNamespaces": true
  }
}
```

## Mutations

### Sync Resources

Synchronize resources from Kubernetes to Neo4j.

```graphql
mutation SyncResources($namespace: String) {
  syncResources(namespace: $namespace) {
    success
    message
    syncedCount
    errors
  }
}
```

**Variables:**
```json
{
  "namespace": "default"
}
```

### Refresh Graph

Refresh the entire graph from the cluster.

```graphql
mutation RefreshGraph {
  refreshGraph {
    success
    message
    syncedCount
    errors
  }
}
```

## Subscriptions

### Watch Resource Changes

```graphql
subscription WatchResources($namespace: String, $kinds: [String!]) {
  resourceChanged(namespace: $namespace, kinds: $kinds) {
    uid
    name
    kind
    status {
      phase
      ready
    }
  }
}
```

**Variables:**
```json
{
  "namespace": "default",
  "kinds": ["Pod"]
}
```

## Examples

### Get Deployment with All Related Resources

```graphql
query GetDeploymentTopology($name: String!, $namespace: String!) {
  resources(filter: { 
    namespace: $namespace, 
    kinds: ["Deployment"], 
    namePattern: $name 
  }) {
    uid
    name
    owned {
      uid
      name
      kind
      owned {
        uid
        name
        kind
        status {
          phase
          ready
        }
      }
    }
    relatedTo {
      type
      resource {
        uid
        name
        kind
      }
    }
  }
}
```

### Get Service Endpoints

```graphql
query GetServiceEndpoints($serviceName: String!, $namespace: String!) {
  resources(filter: {
    namespace: $namespace,
    kinds: ["Service"],
    namePattern: $serviceName
  }) {
    uid
    name
    relatedTo {
      type
      resource {
        uid
        name
        kind
        ... on K8sResource {
          status {
            phase
            ready
          }
        }
      }
    }
  }
}
```

## Error Handling

Errors are returned in the standard GraphQL error format:

```json
{
  "errors": [
    {
      "message": "Resource not found",
      "path": ["resource"],
      "extensions": {
        "code": "NOT_FOUND"
      }
    }
  ],
  "data": {
    "resource": null
  }
}
```

### Error Codes

| Code | Description |
|------|-------------|
| `NOT_FOUND` | Resource does not exist |
| `FORBIDDEN` | Permission denied |
| `BAD_REQUEST` | Invalid input |
| `INTERNAL_ERROR` | Server error |
| `NEO4J_ERROR` | Database error |
| `K8S_ERROR` | Kubernetes API error |
