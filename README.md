# K8S Graph Explorer 🌐

A Kubernetes Resource Topology Visualizer that provides an interactive graph-based view of your Kubernetes cluster resources and their relationships.

## 🚀 Overview

K8S Graph Explorer helps DevOps engineers and platform teams understand the complex relationships between Kubernetes resources through an intuitive graph visualization. It leverages Neo4j as a graph database to store and query resource relationships efficiently.

## 🏗️ Architecture

```
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│    Dashboard    │────▶│   GraphQL API   │────▶│     Neo4j       │
│   (Next.js)     │     │   (Go Backend)  │     │  (Graph DB)     │
└─────────────────┘     └─────────────────┘     └─────────────────┘
                               │
                               ▼
                        ┌─────────────────┐
                        │   Kubernetes    │
                        │    Cluster      │
                        └─────────────────┘
```

## 📁 Project Structure

```
K8S-Graph-Explorer/
├── backend/           # Go backend with GraphQL API
│   ├── cmd/           # Application entrypoints
│   ├── internal/      # Private application code
│   ├── pkg/           # Public libraries
│   └── graph/         # GraphQL schema and resolvers
├── dashboard/         # Next.js frontend application
├── docker/            # Docker configurations
│   ├── backend/       # Backend Dockerfile
│   ├── dashboard/     # Dashboard Dockerfile
│   └── neo4j/         # Neo4j configuration
├── k8s/               # Kubernetes manifests
│   ├── base/          # Base Kustomize configs
│   └── overlays/      # Environment-specific overlays
└── docs/              # Project documentation
```

## 🛠️ Tech Stack

| Component | Technology |
|-----------|------------|
| Backend | Go 1.21+ |
| Frontend | Next.js 14 |
| Graph Database | Neo4j 5.x |
| API | GraphQL |
| Containerization | Docker |
| Orchestration | Kubernetes |

## 🚦 Getting Started

### Prerequisites

- Go 1.21+
- Node.js 18+
- Docker & Docker Compose
- Access to a Kubernetes cluster
- Neo4j (via Docker or managed service)

### Quick Start with Docker Compose

```bash
# Clone the repository
git clone https://github.com/yourusername/K8S-Graph-Explorer.git
cd K8S-Graph-Explorer

# Start all services
docker-compose up -d

# Access the dashboard
open http://localhost:3000
```

### Development Setup

#### Backend

```bash
cd backend
go mod download
go run cmd/server/main.go
```

#### Dashboard

```bash
cd dashboard
npm install
npm run dev
```

## 🔧 Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `NEO4J_URI` | Neo4j connection URI | `bolt://localhost:7687` |
| `NEO4J_USER` | Neo4j username | `neo4j` |
| `NEO4J_PASSWORD` | Neo4j password | - |
| `GRAPHQL_PORT` | GraphQL server port | `8080` |
| `KUBECONFIG` | Path to kubeconfig | `~/.kube/config` |

## 📊 Features

- **Interactive Graph Visualization**: Explore K8s resources as connected nodes
- **Real-time Updates**: Watch resource changes in real-time
- **Relationship Discovery**: Automatically discover relationships between resources
- **Search & Filter**: Find resources quickly with powerful search
- **Multiple Cluster Support**: Connect to multiple clusters
- **Export Capabilities**: Export graphs as images or data

## 🤝 Contributing

Contributions are welcome! Please read our [Contributing Guide](docs/CONTRIBUTING.md) for details.

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 📞 Support

- 📧 Email: support@example.com
- 💬 Discord: [Join our server](#)
- 📖 Documentation: [docs/](docs/)
