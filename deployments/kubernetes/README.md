# Kubernetes Deployment Guide

This directory contains Kubernetes manifests for deploying the Gorax workflow automation platform across different environments.

## Table of Contents

- [Architecture Overview](#architecture-overview)
- [Prerequisites](#prerequisites)
- [Quick Start](#quick-start)
- [Deployment Structure](#deployment-structure)
- [Configuration](#configuration)
- [Environments](#environments)
- [Secrets Management](#secrets-management)
- [Scaling](#scaling)
- [Monitoring and Observability](#monitoring-and-observability)
- [Troubleshooting](#troubleshooting)
- [Security](#security)
- [Maintenance](#maintenance)

## Architecture Overview

The Gorax platform consists of two main components:

1. **API Service**: Handles HTTP requests, webhook ingestion, and workflow management
2. **Worker Service**: Executes workflow tasks asynchronously

Both components are designed for horizontal scaling and high availability.

```
┌─────────────────┐
│   Ingress       │ (TLS, Rate Limiting, CORS)
└────────┬────────┘
         │
    ┌────┴────┐
    │         │
┌───▼───┐ ┌──▼─────┐
│  API  │ │Webhooks│
│Service│ │Endpoint│
└───┬───┘ └───┬────┘
    │         │
    └────┬────┘
         │
    ┌────▼────────┐
    │  Database   │
    │ (Postgres)  │
    └────┬────────┘
         │
    ┌────▼────┐
    │ Workers │ (Scaling 2-30)
    └─────────┘
```

## Prerequisites

- Kubernetes cluster (1.24+)
- kubectl (configured to access your cluster)
- kustomize (v4.0+) or kubectl with built-in kustomize support
- PostgreSQL database (can be deployed separately or use managed service)
- Ingress controller (nginx-ingress recommended)
- cert-manager (for TLS certificate management)

### Optional but Recommended

- Prometheus and Grafana for metrics
- Jaeger or OpenTelemetry for distributed tracing
- External secrets operator or sealed-secrets for secret management

## Quick Start

### 1. Deploy to Development Environment

```bash
# Review what will be deployed
kubectl kustomize deployments/kubernetes/overlays/development

# Apply the manifests
kubectl apply -k deployments/kubernetes/overlays/development

# Verify deployment
kubectl get pods -n gorax-dev
kubectl get svc -n gorax-dev
kubectl get ingress -n gorax-dev
```

### 2. Create Secrets

```bash
# Generate secrets (example)
export DB_URL="postgresql://user:password@postgres:5432/gorax?sslmode=require"
export JWT_SECRET=$(openssl rand -base64 32)
export ENCRYPTION_KEY=$(openssl rand -base64 32)

# Create secret
kubectl create secret generic dev-gorax-secrets \
  --from-literal=database-url="$DB_URL" \
  --from-literal=jwt-secret="$JWT_SECRET" \
  --from-literal=encryption-key="$ENCRYPTION_KEY" \
  --namespace=gorax-dev

# Verify
kubectl get secret dev-gorax-secrets -n gorax-dev
```

### 3. Access the Application

```bash
# Port forward for local testing
kubectl port-forward -n gorax-dev svc/dev-gorax-api 8080:80

# Test health endpoint
curl http://localhost:8080/health/live
```

## Deployment Structure

```
deployments/kubernetes/
├── base/                           # Base configurations
│   ├── kustomization.yaml         # Base kustomization
│   ├── namespace.yaml             # Namespace definition
│   ├── serviceaccount.yaml        # Service account for pods
│   ├── configmap.yaml             # Application configuration
│   ├── secrets.yaml               # Secret template (DO NOT commit real secrets)
│   ├── ingress.yaml               # Ingress configuration
│   ├── api/                       # API service manifests
│   │   ├── deployment.yaml        # API deployment
│   │   ├── service.yaml           # API service
│   │   ├── hpa.yaml               # Horizontal Pod Autoscaler
│   │   └── pdb.yaml               # Pod Disruption Budget
│   └── worker/                    # Worker service manifests
│       ├── deployment.yaml        # Worker deployment
│       ├── hpa.yaml               # Horizontal Pod Autoscaler
│       └── pdb.yaml               # Pod Disruption Budget
├── overlays/                      # Environment-specific overlays
│   ├── development/              # Development environment
│   │   ├── kustomization.yaml   # Dev customizations
│   │   └── patches/             # Dev-specific patches
│   ├── staging/                 # Staging environment
│   │   ├── kustomization.yaml   # Staging customizations
│   │   └── patches/             # Staging-specific patches
│   └── production/              # Production environment
│       ├── kustomization.yaml   # Production customizations
│       └── patches/             # Production-specific patches
└── README.md                    # This file
```

## Configuration

### ConfigMap Settings

The application is configured through the `gorax-config` ConfigMap. Key settings include:

| Setting | Default | Description |
|---------|---------|-------------|
| `log-level` | `info` | Logging level (debug, info, warn, error) |
| `worker-concurrency` | `10` | Number of concurrent workflows per worker |
| `max-workflow-execution-time` | `3600` | Maximum workflow execution time (seconds) |
| `workflow-timeout` | `300` | Individual step timeout (seconds) |
| `cors-allowed-origins` | (varies) | Allowed CORS origins |
| `feature-webhook-replay` | `true` | Enable webhook replay feature |
| `db-max-open-conns` | `25` | Max database connections |

### Environment Variables

Sensitive configuration is provided via environment variables sourced from Secrets:

- `DATABASE_URL`: PostgreSQL connection string
- `JWT_SECRET`: Secret for signing JWT tokens
- `ENCRYPTION_KEY`: 32-byte key for encrypting sensitive data

## Environments

### Development

- **Namespace**: `gorax-dev`
- **Replicas**: API=1, Workers=1
- **Resources**: Minimal (128-256Mi RAM, 100-250m CPU)
- **Scaling**: Disabled/minimal HPA
- **Domain**: `api.dev.gorax.io`, `webhooks.dev.gorax.io`
- **Logging**: Debug level
- **Features**: All enabled for testing

```bash
# Deploy to development
kubectl apply -k deployments/kubernetes/overlays/development

# Watch logs
kubectl logs -f -n gorax-dev -l component=api
kubectl logs -f -n gorax-dev -l component=worker
```

### Staging

- **Namespace**: `gorax-staging`
- **Replicas**: API=2, Workers=2
- **Resources**: Moderate (192-384Mi RAM, 200-400m CPU)
- **Scaling**: HPA 2-6 (API), 2-10 (Workers)
- **Domain**: `api.staging.gorax.io`, `webhooks.staging.gorax.io`
- **Logging**: Info level
- **Features**: All enabled, production-like configuration

```bash
# Deploy to staging
kubectl apply -k deployments/kubernetes/overlays/staging

# Check status
kubectl get pods -n gorax-staging -o wide
kubectl get hpa -n gorax-staging
```

### Production

- **Namespace**: `gorax-prod`
- **Replicas**: API=3, Workers=3
- **Resources**: Full (256-512Mi RAM for API, 512Mi-1Gi for Workers)
- **Scaling**: Aggressive HPA 3-20 (API), 3-30 (Workers)
- **Domain**: `api.gorax.io`, `webhooks.gorax.io`
- **Logging**: Warn level
- **Features**: All enabled
- **Topology**: Spread across availability zones
- **PDB**: Minimum 2 pods available during disruptions

```bash
# Deploy to production (use GitOps or CI/CD in real scenarios)
kubectl apply -k deployments/kubernetes/overlays/production

# Verify deployment
kubectl get all -n gorax-prod
kubectl get hpa -n gorax-prod -w
```

## Secrets Management

### Development

For development, you can create secrets directly:

```bash
kubectl create secret generic dev-gorax-secrets \
  --from-literal=database-url='postgresql://localhost:5432/gorax' \
  --from-literal=jwt-secret='dev-secret' \
  --from-literal=encryption-key='01234567890123456789012345678901' \
  -n gorax-dev
```

### Staging and Production

**DO NOT create secrets manually in staging/production!**

Use one of these approaches:

#### Option 1: Sealed Secrets

```bash
# Install sealed-secrets controller
kubectl apply -f https://github.com/bitnami-labs/sealed-secrets/releases/download/v0.24.0/controller.yaml

# Create sealed secret
echo -n 'postgresql://...' | kubectl create secret generic gorax-secrets \
  --dry-run=client \
  --from-file=database-url=/dev/stdin \
  -o yaml | \
  kubeseal -o yaml > sealed-secrets.yaml

# Apply sealed secret
kubectl apply -f sealed-secrets.yaml -n gorax-prod
```

#### Option 2: External Secrets Operator

```yaml
apiVersion: external-secrets.io/v1beta1
kind: ExternalSecret
metadata:
  name: gorax-secrets
  namespace: gorax-prod
spec:
  refreshInterval: 1h
  secretStoreRef:
    name: aws-secrets-manager
    kind: SecretStore
  target:
    name: gorax-secrets
  data:
  - secretKey: database-url
    remoteRef:
      key: gorax/prod/database-url
  - secretKey: jwt-secret
    remoteRef:
      key: gorax/prod/jwt-secret
  - secretKey: encryption-key
    remoteRef:
      key: gorax/prod/encryption-key
```

#### Option 3: Cloud Provider Secret Managers

**AWS Secrets Manager (EKS):**

```yaml
# ServiceAccount with IRSA
apiVersion: v1
kind: ServiceAccount
metadata:
  name: gorax
  annotations:
    eks.amazonaws.com/role-arn: arn:aws:iam::ACCOUNT_ID:role/gorax-secrets-role
```

**GCP Secret Manager (GKE):**

```yaml
# ServiceAccount with Workload Identity
apiVersion: v1
kind: ServiceAccount
metadata:
  name: gorax
  annotations:
    iam.gke.io/gcp-service-account: gorax@PROJECT_ID.iam.gserviceaccount.com
```

## Scaling

### Manual Scaling

```bash
# Scale API pods
kubectl scale deployment prod-gorax-api -n gorax-prod --replicas=5

# Scale worker pods
kubectl scale deployment prod-gorax-worker -n gorax-prod --replicas=10
```

### Horizontal Pod Autoscaler (HPA)

HPA is configured for both API and worker components.

**API Scaling Triggers:**
- CPU utilization > 70%
- Memory utilization > 80%

**Worker Scaling Triggers:**
- CPU utilization > 75%
- Memory utilization > 85%

```bash
# View HPA status
kubectl get hpa -n gorax-prod

# Describe HPA for details
kubectl describe hpa prod-gorax-api-hpa -n gorax-prod
kubectl describe hpa prod-gorax-worker-hpa -n gorax-prod

# Watch autoscaling in action
kubectl get hpa -n gorax-prod -w
```

### Custom Metrics (Advanced)

To scale based on custom metrics (e.g., workflow queue depth):

1. Deploy Prometheus and Prometheus Adapter
2. Configure custom metrics in HPA:

```yaml
metrics:
- type: Pods
  pods:
    metric:
      name: workflow_queue_depth
    target:
      type: AverageValue
      averageValue: "10"
```

## Monitoring and Observability

### Health Checks

Both API and workers expose health endpoints:

- `/health/live` - Liveness probe (is the app running?)
- `/health/ready` - Readiness probe (can the app serve traffic?)

```bash
# Check health directly
kubectl exec -it -n gorax-prod pod/prod-gorax-api-xxx -- wget -qO- localhost:8080/health/live
```

### Metrics

Prometheus metrics are exposed on:
- API: `:8080/metrics`
- Workers: `:8081/metrics`

Enable scraping with Prometheus ServiceMonitor:

```yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: gorax-api
  namespace: gorax-prod
spec:
  selector:
    matchLabels:
      app: gorax
      component: api
  endpoints:
  - port: http
    path: /metrics
```

### Logging

View logs:

```bash
# API logs
kubectl logs -f -n gorax-prod -l component=api --tail=100

# Worker logs
kubectl logs -f -n gorax-prod -l component=worker --tail=100

# All gorax logs
kubectl logs -f -n gorax-prod -l app=gorax

# Specific pod
kubectl logs -f -n gorax-prod prod-gorax-api-12345-abcde
```

### Tracing

Distributed tracing is enabled via OpenTelemetry. Configure the tracing endpoint in the ConfigMap:

```yaml
data:
  tracing-enabled: "true"
  tracing-sample-rate: "0.1"
  otel-exporter-endpoint: "http://jaeger-collector:14268/api/traces"
```

## Troubleshooting

### Pods Not Starting

```bash
# Check pod status
kubectl get pods -n gorax-prod

# Describe pod for events
kubectl describe pod -n gorax-prod prod-gorax-api-xxx

# Check logs
kubectl logs -n gorax-prod prod-gorax-api-xxx
```

**Common Issues:**

1. **ImagePullBackOff**: Image not found or registry auth issues
2. **CrashLoopBackOff**: Application failing to start (check logs)
3. **Pending**: Insufficient resources or scheduling constraints

### Database Connection Issues

```bash
# Test database connectivity
kubectl run -it --rm debug --image=postgres:15 --restart=Never -n gorax-prod -- \
  psql "postgresql://user:password@postgres:5432/gorax?sslmode=require"

# Check secrets
kubectl get secret prod-gorax-secrets -n gorax-prod -o jsonpath='{.data.database-url}' | base64 -d
```

### High Memory/CPU Usage

```bash
# Check resource usage
kubectl top pods -n gorax-prod

# Check HPA status
kubectl get hpa -n gorax-prod

# Increase resources if needed
kubectl set resources deployment prod-gorax-worker -n gorax-prod \
  --limits=cpu=2000m,memory=2Gi \
  --requests=cpu=1000m,memory=1Gi
```

### Ingress Not Working

```bash
# Check ingress status
kubectl get ingress -n gorax-prod
kubectl describe ingress prod-gorax-ingress -n gorax-prod

# Check ingress controller logs
kubectl logs -n ingress-nginx -l app.kubernetes.io/component=controller

# Verify DNS
nslookup api.gorax.io

# Check certificate
kubectl get certificate -n gorax-prod
kubectl describe certificate prod-gorax-tls -n gorax-prod
```

### Workers Not Processing Workflows

```bash
# Check worker logs
kubectl logs -n gorax-prod -l component=worker --tail=100

# Check worker count
kubectl get pods -n gorax-prod -l component=worker

# Check database for pending workflows
kubectl exec -it deploy/prod-gorax-api -n gorax-prod -- \
  psql $DATABASE_URL -c "SELECT status, COUNT(*) FROM workflow_executions GROUP BY status;"
```

## Security

### Security Best Practices

1. **Network Policies**: Restrict pod-to-pod communication

```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: gorax-api-policy
  namespace: gorax-prod
spec:
  podSelector:
    matchLabels:
      component: api
  policyTypes:
  - Ingress
  - Egress
  ingress:
  - from:
    - namespaceSelector:
        matchLabels:
          name: ingress-nginx
    ports:
    - protocol: TCP
      port: 8080
  egress:
  - to:
    - namespaceSelector:
        matchLabels:
          name: gorax-prod
    ports:
    - protocol: TCP
      port: 5432  # PostgreSQL
```

2. **Pod Security Standards**: Enforce restricted pod security

```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: gorax-prod
  labels:
    pod-security.kubernetes.io/enforce: restricted
    pod-security.kubernetes.io/audit: restricted
    pod-security.kubernetes.io/warn: restricted
```

3. **RBAC**: Limit ServiceAccount permissions

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: gorax-role
  namespace: gorax-prod
rules:
- apiGroups: [""]
  resources: ["configmaps", "secrets"]
  verbs: ["get", "list"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: gorax-rolebinding
  namespace: gorax-prod
subjects:
- kind: ServiceAccount
  name: gorax
roleRef:
  kind: Role
  name: gorax-role
  apiGroup: rbac.authorization.k8s.io
```

4. **Image Security**: Use specific image tags and scan for vulnerabilities

```bash
# Scan images before deployment
docker scan gorax/api:v1.2.3
docker scan gorax/worker:v1.2.3
```

5. **Secret Encryption**: Enable encryption at rest for secrets

```yaml
# Encryption configuration for kube-apiserver
apiVersion: apiserver.config.k8s.io/v1
kind: EncryptionConfiguration
resources:
  - resources:
    - secrets
    providers:
    - aescbc:
        keys:
        - name: key1
          secret: <base64-encoded-32-byte-key>
    - identity: {}
```

## Maintenance

### Updates and Rollouts

```bash
# Update API image
kubectl set image deployment/prod-gorax-api api=gorax/api:v1.2.3 -n gorax-prod

# Update worker image
kubectl set image deployment/prod-gorax-worker worker=gorax/worker:v1.2.3 -n gorax-prod

# Watch rollout
kubectl rollout status deployment/prod-gorax-api -n gorax-prod

# View rollout history
kubectl rollout history deployment/prod-gorax-api -n gorax-prod

# Rollback if needed
kubectl rollout undo deployment/prod-gorax-api -n gorax-prod
```

### Draining Nodes

```bash
# Safely drain a node before maintenance
kubectl drain node-name --ignore-daemonsets --delete-emptydir-data

# Uncordon after maintenance
kubectl uncordon node-name
```

### Database Migrations

```bash
# Run migrations as a Kubernetes Job
kubectl apply -f - <<EOF
apiVersion: batch/v1
kind: Job
metadata:
  name: gorax-migration-$(date +%s)
  namespace: gorax-prod
spec:
  template:
    spec:
      restartPolicy: Never
      containers:
      - name: migrate
        image: gorax/api:v1.2.3
        command: ["/app/migrate"]
        env:
        - name: DATABASE_URL
          valueFrom:
            secretKeyRef:
              name: gorax-secrets
              key: database-url
EOF

# Check job status
kubectl get jobs -n gorax-prod
kubectl logs -n gorax-prod job/gorax-migration-xxx
```

### Backup and Restore

```bash
# Backup ConfigMap
kubectl get configmap prod-gorax-config -n gorax-prod -o yaml > backup-config.yaml

# Backup all manifests
kubectl get all -n gorax-prod -o yaml > backup-all.yaml

# Database backup (external to Kubernetes)
# Use your database provider's backup solution
```

## Additional Resources

- [Kubernetes Documentation](https://kubernetes.io/docs/)
- [Kustomize Documentation](https://kustomize.io/)
- [nginx-ingress Documentation](https://kubernetes.github.io/ingress-nginx/)
- [cert-manager Documentation](https://cert-manager.io/docs/)
- [Prometheus Operator](https://prometheus-operator.dev/)

## Support

For issues and questions:
- GitHub Issues: https://github.com/gorax/gorax/issues
- Documentation: https://docs.gorax.io
- Community: https://discord.gg/gorax
