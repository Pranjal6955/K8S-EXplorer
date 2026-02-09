# Deployment Guide

This guide covers deploying K8S Graph Explorer to a Kubernetes cluster.

## Prerequisites

- A Kubernetes cluster (1.25+)
- kubectl configured with cluster access
- Container registry access (Docker Hub, GCR, ECR, etc.)
- Ingress controller installed (for production)
- cert-manager installed (for TLS, optional)

## Quick Deploy with Docker Compose

For local testing or development, use Docker Compose:

```bash
# Start all services
docker-compose up -d

# View logs
docker-compose logs -f

# Stop services
docker-compose down
```

## Kubernetes Deployment

### 1. Build and Push Images

```bash
# Build backend image
docker build -t your-registry/k8s-graph-backend:latest -f docker/backend/Dockerfile backend/

# Build dashboard image
docker build -t your-registry/k8s-graph-dashboard:latest -f docker/dashboard/Dockerfile dashboard/

# Push images
docker push your-registry/k8s-graph-backend:latest
docker push your-registry/k8s-graph-dashboard:latest
```

### 2. Create Secrets

Create the Neo4j credentials secret:

```bash
kubectl create namespace k8s-graph-explorer

kubectl create secret generic neo4j-credentials \
  --namespace k8s-graph-explorer \
  --from-literal=username=neo4j \
  --from-literal=password=your-secure-password \
  --from-literal=auth=neo4j/your-secure-password
```

### 3. Deploy Using Kustomize

#### Development Environment

```bash
kubectl apply -k k8s/overlays/development/
```

#### Production Environment

First, update the image references in the production overlay:

```yaml
# k8s/overlays/production/kustomization.yaml
images:
  - name: k8s-graph-explorer-backend
    newName: your-registry/k8s-graph-backend
    newTag: v1.0.0
  - name: k8s-graph-explorer-dashboard
    newName: your-registry/k8s-graph-dashboard
    newTag: v1.0.0
```

Then deploy:

```bash
kubectl apply -k k8s/overlays/production/
```

### 4. Verify Deployment

```bash
# Check all pods are running
kubectl get pods -n k8s-graph-explorer

# Check services
kubectl get svc -n k8s-graph-explorer

# View backend logs
kubectl logs -f deployment/backend -n k8s-graph-explorer

# View dashboard logs
kubectl logs -f deployment/dashboard -n k8s-graph-explorer
```

## Configuration

### Environment Variables

#### Backend

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | Server port | `8080` |
| `NEO4J_URI` | Neo4j connection URI | `bolt://neo4j:7687` |
| `NEO4J_USERNAME` | Neo4j username | `neo4j` |
| `NEO4J_PASSWORD` | Neo4j password | - |
| `IN_CLUSTER` | Running inside K8s | `true` |

#### Dashboard

| Variable | Description | Default |
|----------|-------------|---------|
| `NEXT_PUBLIC_GRAPHQL_ENDPOINT` | GraphQL API URL | `http://backend:8080/graphql` |

### Resource Requirements

Adjust resources based on your cluster size:

| Component | Dev Requests | Dev Limits | Prod Requests | Prod Limits |
|-----------|--------------|------------|---------------|-------------|
| Backend | 64Mi/50m | 256Mi/250m | 256Mi/250m | 1Gi/1000m |
| Dashboard | 64Mi/50m | 128Mi/100m | 128Mi/100m | 512Mi/500m |
| Neo4j | 512Mi/500m | 2Gi/1000m | 2Gi/1000m | 4Gi/2000m |

## Ingress Configuration

### NGINX Ingress

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: k8s-graph-explorer
  annotations:
    kubernetes.io/ingress.class: nginx
    cert-manager.io/cluster-issuer: letsencrypt-prod
spec:
  tls:
    - hosts:
        - k8s-graph.yourdomain.com
      secretName: k8s-graph-tls
  rules:
    - host: k8s-graph.yourdomain.com
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: dashboard
                port:
                  number: 3000
          - path: /graphql
            pathType: Prefix
            backend:
              service:
                name: backend
                port:
                  number: 8080
```

## Monitoring

### Health Checks

All components expose health check endpoints:

- **Backend**: `GET /health`
- **Dashboard**: `GET /`
- **Neo4j**: `GET :7474/`

### Prometheus Metrics

Add Prometheus annotations to scrape metrics:

```yaml
metadata:
  annotations:
    prometheus.io/scrape: "true"
    prometheus.io/port: "8080"
    prometheus.io/path: "/metrics"
```

## Backup and Recovery

### Neo4j Backup

```bash
# Create backup
kubectl exec -it neo4j-0 -n k8s-graph-explorer -- \
  neo4j-admin database dump neo4j --to-path=/backups/

# Copy backup locally
kubectl cp k8s-graph-explorer/neo4j-0:/backups/neo4j.dump ./neo4j.dump
```

### Restore Neo4j

```bash
# Copy backup to pod
kubectl cp ./neo4j.dump k8s-graph-explorer/neo4j-0:/backups/

# Restore
kubectl exec -it neo4j-0 -n k8s-graph-explorer -- \
  neo4j-admin database load neo4j --from-path=/backups/
```

## Scaling

### Horizontal Pod Autoscaler

```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: backend-hpa
  namespace: k8s-graph-explorer
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: backend
  minReplicas: 2
  maxReplicas: 10
  metrics:
    - type: Resource
      resource:
        name: cpu
        target:
          type: Utilization
          averageUtilization: 70
```

## Troubleshooting

### Common Issues

1. **Neo4j not starting**: Check PVC is bound and has enough space
2. **Backend can't connect to Neo4j**: Verify service DNS and credentials
3. **Dashboard shows connection error**: Check NEXT_PUBLIC_GRAPHQL_ENDPOINT
4. **RBAC errors**: Ensure ServiceAccount has correct ClusterRole binding

### Debug Commands

```bash
# View all events
kubectl get events -n k8s-graph-explorer --sort-by='.lastTimestamp'

# Describe failing pod
kubectl describe pod <pod-name> -n k8s-graph-explorer

# Check network connectivity
kubectl run debug --rm -it --image=busybox -n k8s-graph-explorer -- \
  wget -qO- http://backend:8080/health
```
