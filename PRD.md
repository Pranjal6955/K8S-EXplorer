# Product Requirements Document: K8S Graph Explorer (v2)

## 1. Executive Summary
K8S Graph Explorer is a specialized observability tool designed to visualize the complex web of relationships within a Kubernetes cluster. By representing resources (Pods, Services, Deployments, etc.) as nodes and their interactions (labels, ownership, mapping) as edges, it provides a mental model that standard dashboards lack.

## 2. Problem Statement
Kubernetes clusters are inherently complex. Understanding why a Service isn't reaching a Pod or identifying the blast radius of a failing Node often requires multiple `kubectl` commands and manual mental mapping. Existing tools are either too text-heavy or lack real-time relationship discovery.

## 3. Product Vision
To be the "Google Maps" for Kubernetes—allowing users to zoom into specific namespaces, trace traffic flows, and visualize resource dependencies in real-time with zero configuration.

## 4. Target Audience
- **DevOps Engineers**: Troubleshooting connectivity and resource mapping.
- **Platform Teams**: Auditing cluster architecture and resource usage.
- **Developers**: Understanding how their applications are deployed and exposed.

## 5. Functional Requirements

### 5.1 Resource Synchronization
- **Automatic Discovery**: Scan all standard K8s resources (Pods, Services, Deployments, ReplicaSets, ConfigMaps, Secrets, Ingresses, etc.).
- **Relationship Mapping**:
    - `OWNS`: Deployment -> ReplicaSet -> Pod.
    - `EXPOSES`: Service -> Pod.
    - `MOUNTS`: Pod -> ConfigMap/Secret/PVC.
    - `SELECTS`: Service -> Pod (via label selectors).
- **Real-time Sync**: Listen to K8s API events (Informer pattern) to update the graph instantly.

### 5.2 Visualization Dashboard
- **Interactive Graph**: Pan, zoom, and drag nodes.
- **Details Pane**: Click a node to see its full YAML, status, and metadata.
- **Filtering**: Filter by Namespace, Resource Type, or Label.
- **Search**: Global search for any resource by name or UID.
- **Layouts**: Multiple layout algorithms (Force-directed, Hierarchical, Circle).

### 5.3 Advanced Features
- **Pathfinding**: Find the shortest path between two resources (e.g., "How does this Ingress reach this Pod?").
- **Blast Radius Analysis**: Highlight all resources dependent on a selected node.
- **Multi-cluster Support**: Switch between different contexts/clusters seamlessly.

## 6. Technical Stack (Proposed)

| Layer | Technology | Rationale |
|-------|------------|-----------|
| **Frontend** | Next.js 14+ | Fast, SEO-friendly, robust ecosystem. |
| **Styling** | Tailwind CSS + Shadcn/UI | Premium, consistent design system. |
| **Graph Engine** | Cytoscape.js | High performance for complex graphs. |
| **API** | GraphQL (GQLen) | Efficiently fetch only the nodes/edges needed for the current viewport. |
| **Backend** | Go 1.22+ | Excellent concurrency for K8s Informers and high performance. |
| **Database** | Neo4j | Native graph storage for complex relationship queries. |
| **Deployment** | Docker / Helm | Industry standard for K8s tooling. |

## 7. Non-Functional Requirements
- **Scalability**: Support clusters with up to 5,000 nodes without UI lag.
- **Security**: Read-only access to K8s API by default (RBAC).
- **Usability**: First-class dark mode support and responsive design.

## 8. Success Metrics
- **Time to Insight**: Reduced time to identify resource relationships vs. `kubectl`.
- **User Engagement**: Frequency of use during troubleshooting sessions.
- **Accuracy**: 100% parity between the graph state and the actual cluster state.

## 9. Roadmap
1. **Phase 1 (MVP)**: Core sync engine + Basic graph visualization (Nodes & Edges).
2. **Phase 2 (UX)**: Details pane, search, and advanced filtering.
3. **Phase 3 (Analysis)**: Pathfinding, blast radius, and real-time event streaming.
4. **Phase 4 (Enterprise)**: Multi-cluster support and Auth/RBAC integration.
