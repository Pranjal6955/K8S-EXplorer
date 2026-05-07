# Development Guide

## Prerequisites

Before you begin, ensure you have the following installed:

- **Go 1.21+**: [Download Go](https://golang.org/dl/)
- **Node.js 18+**: [Download Node.js](https://nodejs.org/)
- **Docker & Docker Compose**: [Install Docker](https://docs.docker.com/get-docker/)
- **kubectl**: [Install kubectl](https://kubernetes.io/docs/tasks/tools/)
- **A Kubernetes cluster**: Local (minikube, kind) or remote

## Quick Start

### 1. Clone the Repository

```bash
git clone https://github.com/yourusername/K8S-Graph-Explorer.git
cd K8S-Graph-Explorer
```

### 2. Initialize the Project

The easiest way to set up everything is using the root `Makefile`:

```bash
# Initialize environment (.env) and install all dependencies
make setup
```

### 3. Start Dependencies (Neo4j)

```bash
# Start Neo4j in the background
make up-deps
```

### 4. Run in Development Mode

You can run both backend and dashboard in development mode:

```bash
# In one terminal, run the backend with hot reload
make dev-backend

# In another terminal, run the dashboard
make dev-dashboard
```

Alternatively, you can run the entire stack using Docker Compose:

```bash
make up
```

## Project Structure

```
K8S-Graph-Explorer/
├── backend/
│   ├── cmd/
│   │   └── server/           # Application entry point
│   │       └── main.go
│   ├── internal/
│   │   ├── config/           # Configuration handling
│   │   ├── database/         # Neo4j client
│   │   └── k8s/              # Kubernetes client
│   ├── graph/
│   │   ├── model/            # GraphQL models
│   │   ├── resolver.go       # Root resolver
│   │   ├── schema.graphqls   # GraphQL schema
│   │   └── schema.resolvers.go
│   ├── pkg/                  # Public packages (if any)
│   ├── go.mod
│   ├── gqlgen.yml            # gqlgen configuration
│   └── Makefile
├── dashboard/
│   ├── app/                  # Next.js App Router
│   ├── components/           # React components
│   ├── lib/                  # Utility functions
│   └── public/               # Static assets
├── docker/
│   ├── backend/
│   ├── dashboard/
│   └── neo4j/
├── k8s/
│   ├── base/                 # Base Kustomize configs
│   └── overlays/             # Environment overlays
└── docs/                     # Documentation
```

## Backend Development

### GraphQL Schema

The GraphQL schema is defined in `backend/graph/schema.graphqls`. After modifying the schema, regenerate the code:

```bash
cd backend
make generate
```

### Adding a New Query/Mutation

1. Update `schema.graphqls` with the new operation
2. Run `make generate`
3. Implement the resolver in `schema.resolvers.go`

### Testing

```bash
cd backend
make test
make test-coverage  # With coverage report
```

### Code Linting

```bash
cd backend
make lint
```

## Dashboard Development

### Component Structure

Components are organized by feature:

```
components/
├── graph/              # Graph visualization
├── resources/          # Resource list/details
├── layout/             # Layout components
└── ui/                 # Reusable UI components
```

### Adding a New Page

Create a new file in `app/` directory following Next.js App Router conventions:

```typescript
// app/resources/page.tsx
export default function ResourcesPage() {
  return <div>Resources</div>
}
```

### GraphQL Integration

Use Apollo Client for GraphQL queries:

```typescript
import { useQuery, gql } from '@apollo/client';

const GET_NAMESPACES = gql`
  query GetNamespaces {
    namespaces {
      uid
      name
    }
  }
`;

function NamespaceList() {
  const { loading, error, data } = useQuery(GET_NAMESPACES);
  // ...
}
```

## Database

### Connecting to Neo4j

Access Neo4j Browser at http://localhost:7474

Default credentials:
- Username: `neo4j`
- Password: `password` (or as set in `.env`)

### Useful Cypher Queries

```cypher
// View all nodes
MATCH (n) RETURN n LIMIT 25

// View all relationships
MATCH (a)-[r]->(b) RETURN a, r, b LIMIT 50

// Find pods in a namespace
MATCH (p:Pod {namespace: 'default'}) RETURN p

// Find deployment hierarchy
MATCH (d:Deployment)-[:OWNS]->(rs:ReplicaSet)-[:OWNS]->(p:Pod)
RETURN d, rs, p
```

## Common Tasks

### Reset the Database

```bash
# Stop all services
docker-compose down -v

# Start fresh
docker-compose up -d
```

### Update Dependencies

```bash
# Backend
cd backend
go get -u ./...
go mod tidy

# Dashboard
cd dashboard
npm update
```

### Build for Production

```bash
# Backend
cd backend
make build

# Dashboard
cd dashboard
npm run build
```

## Troubleshooting

### Neo4j Connection Failed

1. Ensure Neo4j is running: `docker-compose ps`
2. Check Neo4j logs: `docker-compose logs neo4j`
3. Verify connection settings in `.env`

### Kubernetes Connection Failed

1. Check kubeconfig: `kubectl cluster-info`
2. Ensure `KUBECONFIG` env var is set correctly
3. Verify RBAC permissions

### GraphQL Errors

1. Check the GraphQL Playground for errors
2. Review backend logs
3. Ensure schema and resolvers are in sync: `make generate`
