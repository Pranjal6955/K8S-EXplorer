# K8S Graph Explorer - Go Backend

A clean architecture Go backend service using the Gin framework for the Kubernetes Resource Topology Visualizer.

## 🏗️ Architecture

This backend follows **Clean Architecture** principles with clear separation of concerns:

```
backend/
├── cmd/
│   └── server/              # Application entry point
│       └── main.go
├── internal/
│   ├── config/              # Configuration management
│   ├── domain/              # Business logic layer
│   │   ├── models/          # Domain models/entities
│   │   └── repositories/    # Repository interfaces
│   ├── infrastructure/      # External services
│   │   ├── database/        # Neo4j client
│   │   ├── kubernetes/      # K8s client
│   │   └── logger/          # Logging
│   ├── interface/           # Interface adapters
│   │   └── http/
│   │       ├── controllers/ # HTTP handlers
│   │       ├── middleware/  # HTTP middleware
│   │       └── router/      # Route definitions
│   ├── repositories/        # Repository implementations
│   │   └── neo4j/           # Neo4j repositories
│   └── services/            # Application services
├── graph/                   # GraphQL (optional)
├── go.mod
├── Makefile
└── .env.example
```

## 🚀 Quick Start

```bash
# Install dependencies
make deps

# Copy and configure environment
cp .env.example .env

# Run the server
make run

# Or run with hot reload
make dev

# This will start:
# - Backend (Go with Air) at http://localhost:8080
# - Dashboard (Next.js) at http://localhost:3000
```

## 📡 API Endpoints

### Health Checks

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/health` | Health check |
| GET | `/ready` | Readiness check |
| GET | `/live` | Liveness check |

### Resources

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/resources` | List resources |
| GET | `/api/v1/resources/:uid` | Get resource by UID |
| GET | `/api/v1/resources/:uid/related` | Get related resources |
| GET | `/api/v1/resources/search?q=...` | Search resources |

### Graph

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/graph` | Get complete graph |
| GET | `/api/v1/graph/topology/:uid` | Get resource topology |

### Sync

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/sync` | Sync namespace |
| POST | `/api/v1/sync/refresh` | Full refresh |

## 🔧 Configuration

| Variable | Description | Default |
|----------|-------------|---------|
| `SERVER_PORT` | Server port | `8080` |
| `APP_ENV` | Environment (development/production) | `development` |
| `NEO4J_URI` | Neo4j connection URI | `bolt://localhost:7687` |
| `NEO4J_USERNAME` | Neo4j username | `neo4j` |
| `NEO4J_PASSWORD` | Neo4j password | `password` |
| `KUBECONFIG` | Path to kubeconfig | `~/.kube/config` |
| `IN_CLUSTER` | Running inside K8s | `false` |

## 🛠️ Makefile Commands

```bash
make help          # Show all commands
make build         # Build binary
make run           # Run server
make dev           # Run with hot reload
make test          # Run tests
make test-coverage # Run tests with coverage
make lint          # Run linter
make fmt           # Format code
make docker-build  # Build Docker image
make tools         # Install dev tools
```

## 🧪 Testing

```bash
# Run all tests
make test

# Run with coverage report
make test-coverage
```

## 🐳 Docker

```bash
# Build image
make docker-build

# Run container
make docker-run
```
