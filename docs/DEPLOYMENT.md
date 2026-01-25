# Gorax Deployment Guide

## Table of Contents

1. [Deployment Overview](#deployment-overview)
2. [Infrastructure Requirements](#infrastructure-requirements)
3. [Docker Deployment](#docker-deployment)
4. [Kubernetes Deployment](#kubernetes-deployment)
5. [Database Setup](#database-setup)
6. [Redis Setup](#redis-setup)
7. [Environment Configuration](#environment-configuration)
8. [Load Balancer & Ingress](#load-balancer--ingress)
9. [Monitoring & Observability](#monitoring--observability)
10. [CI/CD Pipeline](#cicd-pipeline)
11. [Production Checklist](#production-checklist)
12. [Scaling Strategies](#scaling-strategies)
13. [Troubleshooting](#troubleshooting)

---

## Deployment Overview

### Architecture Options

Gorax supports multiple deployment architectures:

```
┌─────────────────────────────────────────────────────────┐
│                    Load Balancer                        │
│                 (NGINX/ALB/Ingress)                     │
└──────────────┬────────────────────┬─────────────────────┘
               │                    │
       ┌───────▼──────┐     ┌──────▼──────┐
       │  API Server  │     │  API Server │  (Horizontal scaling)
       │   (Go/Chi)   │     │   (Go/Chi)  │
       └───────┬──────┘     └──────┬──────┘
               │                   │
       ┌───────┴───────────────────┴─────────────┐
       │                                          │
   ┌───▼─────┐  ┌─────────┐  ┌──────────┐  ┌───▼────┐
   │PostgreSQL│  │  Redis  │  │   Ory    │  │ Worker │
   │(Primary) │  │ (Cache) │  │  Kratos  │  │ Pool   │
   └──────────┘  └─────────┘  └──────────┘  └────────┘
        │
   ┌────▼────┐
   │PostgreSQL│
   │(Replica) │
   └─────────┘
```

### Deployment Options

1. **Docker Compose** (Development/Small deployments)
   - Single-node deployment
   - Easy to set up
   - Limited scalability

2. **Kubernetes** (Production/Enterprise)
   - High availability
   - Auto-scaling
   - Service mesh integration

3. **Bare Metal** (Custom environments)
   - Maximum control
   - Complex setup
   - Manual scaling

---

## Infrastructure Requirements

### Minimum Requirements (Development)

| Component | CPU | Memory | Storage | Notes |
|-----------|-----|--------|---------|-------|
| API Server | 0.5 cores | 512 MB | 1 GB | Single instance |
| Worker | 0.5 cores | 512 MB | 1 GB | Single instance |
| PostgreSQL | 1 core | 1 GB | 20 GB | With local SSD |
| Redis | 0.5 cores | 256 MB | 1 GB | In-memory cache |
| Ory Kratos | 0.5 cores | 256 MB | 1 GB | Auth server |

**Total:** 3 cores, 2.5 GB RAM, 25 GB storage

### Recommended Production Setup

| Component | CPU | Memory | Storage | Notes |
|-----------|-----|--------|---------|-------|
| API Server (×3) | 2 cores ea. | 2 GB ea. | 10 GB ea. | Load balanced |
| Worker (×5) | 2 cores ea. | 4 GB ea. | 10 GB ea. | Auto-scaling |
| PostgreSQL | 4 cores | 16 GB | 200 GB | SSD, IOPS optimized |
| Redis | 2 cores | 8 GB | 20 GB | Persistence enabled |
| Ory Kratos (×2) | 1 core ea. | 1 GB ea. | 5 GB ea. | HA setup |

**Total:** 32+ cores, 72+ GB RAM, 350+ GB storage

### Cloud Provider Recommendations

#### AWS
- **API/Worker**: ECS Fargate or EKS
- **Database**: RDS PostgreSQL (db.r6g.xlarge or larger)
- **Cache**: ElastiCache Redis (cache.r6g.large)
- **Storage**: S3 for artifacts
- **Queue**: SQS for workflow execution queue
- **Secrets**: AWS Secrets Manager or KMS

#### Google Cloud Platform
- **API/Worker**: GKE or Cloud Run
- **Database**: Cloud SQL PostgreSQL
- **Cache**: Memorystore Redis
- **Storage**: Cloud Storage
- **Queue**: Cloud Tasks or Pub/Sub
- **Secrets**: Secret Manager

#### Azure
- **API/Worker**: AKS or Container Instances
- **Database**: Azure Database for PostgreSQL
- **Cache**: Azure Cache for Redis
- **Storage**: Blob Storage
- **Queue**: Service Bus or Storage Queues
- **Secrets**: Key Vault

---

## Docker Deployment

### Production Docker Compose

Create `docker-compose.prod.yml`:

```yaml
version: '3.8'

services:
  # NGINX Reverse Proxy
  nginx:
    image: nginx:alpine
    container_name: gorax-nginx
    restart: unless-stopped
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx/nginx.conf:/etc/nginx/nginx.conf:ro
      - ./nginx/ssl:/etc/nginx/ssl:ro
      - ./web/dist:/usr/share/nginx/html:ro
    depends_on:
      - api
    networks:
      - gorax-network
    healthcheck:
      test: ["CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://localhost/health"]
      interval: 30s
      timeout: 10s
      retries: 3

  # PostgreSQL Database
  postgres:
    image: postgres:16-alpine
    container_name: gorax-postgres
    restart: unless-stopped
    environment:
      POSTGRES_USER: ${DB_USER:-gorax}
      POSTGRES_PASSWORD: ${DB_PASSWORD}
      POSTGRES_DB: ${DB_NAME:-gorax}
      POSTGRES_INITDB_ARGS: "--encoding=UTF8 --locale=en_US.UTF-8"
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./postgres/postgresql.conf:/etc/postgresql/postgresql.conf:ro
      - ./migrations:/docker-entrypoint-initdb.d:ro
    command: postgres -c config_file=/etc/postgresql/postgresql.conf
    networks:
      - gorax-network
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U ${DB_USER:-gorax}"]
      interval: 10s
      timeout: 5s
      retries: 5
    logging:
      driver: "json-file"
      options:
        max-size: "100m"
        max-file: "3"

  # Redis Cache
  redis:
    image: redis:7-alpine
    container_name: gorax-redis
    restart: unless-stopped
    command: >
      redis-server
      --requirepass ${REDIS_PASSWORD}
      --maxmemory 2gb
      --maxmemory-policy allkeys-lru
      --appendonly yes
      --appendfsync everysec
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data
    networks:
      - gorax-network
    healthcheck:
      test: ["CMD", "redis-cli", "--raw", "incr", "ping"]
      interval: 10s
      timeout: 5s
      retries: 5
    logging:
      driver: "json-file"
      options:
        max-size: "50m"
        max-file: "3"

  # Ory Kratos - Database Migration
  kratos-migrate:
    image: oryd/kratos:v1.1.0
    container_name: gorax-kratos-migrate
    environment:
      DSN: postgres://${DB_USER}:${DB_PASSWORD}@postgres:5432/${DB_NAME}?sslmode=require
    volumes:
      - ./kratos:/etc/config/kratos:ro
    command: migrate sql -e --yes
    depends_on:
      postgres:
        condition: service_healthy
    networks:
      - gorax-network
    restart: on-failure

  # Ory Kratos - Identity Server
  kratos:
    image: oryd/kratos:v1.1.0
    container_name: gorax-kratos
    restart: unless-stopped
    environment:
      DSN: postgres://${DB_USER}:${DB_PASSWORD}@postgres:5432/${DB_NAME}?sslmode=require
      LOG_LEVEL: ${KRATOS_LOG_LEVEL:-info}
      SERVE_PUBLIC_BASE_URL: ${KRATOS_PUBLIC_URL}
      SERVE_ADMIN_BASE_URL: ${KRATOS_ADMIN_URL}
    volumes:
      - ./kratos:/etc/config/kratos:ro
    command: serve --config /etc/config/kratos/kratos.yml
    ports:
      - "4433:4433"  # Public API
      - "4434:4434"  # Admin API
    depends_on:
      kratos-migrate:
        condition: service_completed_successfully
    networks:
      - gorax-network
    healthcheck:
      test: ["CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://localhost:4433/health/ready"]
      interval: 10s
      timeout: 5s
      retries: 5
    logging:
      driver: "json-file"
      options:
        max-size: "100m"
        max-file: "3"

  # Gorax API Server
  api:
    image: ghcr.io/stherrien/gorax:latest
    container_name: gorax-api
    restart: unless-stopped
    environment:
      APP_ENV: production
      SERVER_ADDRESS: :8080
      DB_HOST: postgres
      DB_PORT: 5432
      DB_USER: ${DB_USER}
      DB_PASSWORD: ${DB_PASSWORD}
      DB_NAME: ${DB_NAME}
      DB_SSLMODE: require
      REDIS_ADDRESS: redis:6379
      REDIS_PASSWORD: ${REDIS_PASSWORD}
      KRATOS_PUBLIC_URL: ${KRATOS_PUBLIC_URL}
      KRATOS_ADMIN_URL: http://kratos:4434
      KRATOS_WEBHOOK_SECRET: ${KRATOS_WEBHOOK_SECRET}
      CREDENTIAL_USE_KMS: ${CREDENTIAL_USE_KMS:-false}
      CREDENTIAL_MASTER_KEY: ${CREDENTIAL_MASTER_KEY}
      CREDENTIAL_KMS_KEY_ID: ${CREDENTIAL_KMS_KEY_ID}
      AWS_REGION: ${AWS_REGION}
      AWS_ACCESS_KEY_ID: ${AWS_ACCESS_KEY_ID}
      AWS_SECRET_ACCESS_KEY: ${AWS_SECRET_ACCESS_KEY}
      METRICS_ENABLED: true
      METRICS_PORT: 9090
      TRACING_ENABLED: ${TRACING_ENABLED:-true}
      TRACING_ENDPOINT: ${TRACING_ENDPOINT}
      SENTRY_ENABLED: ${SENTRY_ENABLED:-true}
      SENTRY_DSN: ${SENTRY_DSN}
      SENTRY_ENVIRONMENT: production
      CORS_ALLOWED_ORIGINS: ${CORS_ALLOWED_ORIGINS}
    ports:
      - "8080:8080"
      - "9090:9090"  # Metrics
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
      kratos:
        condition: service_healthy
    networks:
      - gorax-network
    healthcheck:
      test: ["CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s
    deploy:
      resources:
        limits:
          cpus: '2'
          memory: 2G
        reservations:
          cpus: '0.5'
          memory: 512M
    logging:
      driver: "json-file"
      options:
        max-size: "100m"
        max-file: "5"

  # Gorax Worker Pool (scaled with replicas)
  worker:
    image: ghcr.io/stherrien/gorax-worker:latest
    restart: unless-stopped
    environment:
      APP_ENV: production
      DB_HOST: postgres
      DB_PORT: 5432
      DB_USER: ${DB_USER}
      DB_PASSWORD: ${DB_PASSWORD}
      DB_NAME: ${DB_NAME}
      DB_SSLMODE: require
      REDIS_ADDRESS: redis:6379
      REDIS_PASSWORD: ${REDIS_PASSWORD}
      WORKER_CONCURRENCY: 10
      WORKER_MAX_CONCURRENCY_PER_TENANT: 5
      WORKER_HEALTH_PORT: 8081
      AWS_REGION: ${AWS_REGION}
      AWS_SQS_QUEUE_URL: ${AWS_SQS_QUEUE_URL}
      QUEUE_ENABLED: true
      METRICS_ENABLED: true
      SENTRY_ENABLED: ${SENTRY_ENABLED:-true}
      SENTRY_DSN: ${SENTRY_DSN}
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
    networks:
      - gorax-network
    healthcheck:
      test: ["CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://localhost:8081/health"]
      interval: 30s
      timeout: 10s
      retries: 3
    deploy:
      replicas: 3
      resources:
        limits:
          cpus: '2'
          memory: 4G
        reservations:
          cpus: '1'
          memory: 1G
    logging:
      driver: "json-file"
      options:
        max-size: "100m"
        max-file: "5"

volumes:
  postgres_data:
    driver: local
  redis_data:
    driver: local

networks:
  gorax-network:
    driver: bridge
```

### Building and Deploying

```bash
# Clone repository
git clone https://github.com/stherrien/gorax.git
cd gorax

# Create environment file
cp .env.example .env.prod
# Edit .env.prod with production values

# Build images
docker build -t gorax-api:latest -f Dockerfile .
docker build -t gorax-worker:latest -f deployments/docker/Dockerfile.worker .

# Start services
docker-compose -f docker-compose.prod.yml --env-file .env.prod up -d

# Check logs
docker-compose -f docker-compose.prod.yml logs -f api

# Run database migrations (if not auto-applied)
docker-compose -f docker-compose.prod.yml exec api /app/gorax migrate up

# Health check
curl http://localhost:8080/health
```

### Docker Image Optimization

**Multi-stage Build Example:**

```dockerfile
# Stage 1: Build frontend
FROM node:18-alpine AS frontend-builder
WORKDIR /app/web
COPY web/package*.json ./
RUN npm ci --only=production
COPY web/ ./
RUN npm run build

# Stage 2: Build Go backend
FROM golang:1.23-alpine AS backend-builder
RUN apk add --no-cache git make gcc musl-dev
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build \
    -a -installsuffix cgo \
    -ldflags="-w -s -X main.version=${VERSION} -X main.buildTime=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
    -o /bin/gorax ./cmd/api

# Stage 3: Minimal runtime image
FROM gcr.io/distroless/static-debian11
COPY --from=backend-builder /bin/gorax /app/gorax
COPY --from=frontend-builder /app/web/dist /app/web/dist
USER nonroot:nonroot
EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=3s CMD ["/app/gorax", "health"]
ENTRYPOINT ["/app/gorax"]
```

---

## Kubernetes Deployment

### Namespace Setup

```yaml
# namespace.yaml
apiVersion: v1
kind: Namespace
metadata:
  name: gorax
  labels:
    name: gorax
```

### ConfigMap

```yaml
# configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: gorax-config
  namespace: gorax
data:
  APP_ENV: "production"
  SERVER_ADDRESS: ":8080"
  DB_HOST: "postgres.gorax.svc.cluster.local"
  DB_PORT: "5432"
  DB_NAME: "gorax"
  DB_SSLMODE: "require"
  REDIS_ADDRESS: "redis.gorax.svc.cluster.local:6379"
  KRATOS_PUBLIC_URL: "https://auth.gorax.io"
  KRATOS_ADMIN_URL: "http://kratos.gorax.svc.cluster.local:4434"
  METRICS_ENABLED: "true"
  METRICS_PORT: "9090"
  TRACING_ENABLED: "true"
  TRACING_ENDPOINT: "jaeger-collector.observability.svc.cluster.local:4317"
  SENTRY_ENABLED: "true"
  SENTRY_ENVIRONMENT: "production"
  CORS_ALLOWED_ORIGINS: "https://app.gorax.io,https://admin.gorax.io"
  WORKER_CONCURRENCY: "10"
  WORKER_MAX_CONCURRENCY_PER_TENANT: "5"
  QUEUE_ENABLED: "true"
```

### Secrets

```yaml
# secrets.yaml
apiVersion: v1
kind: Secret
metadata:
  name: gorax-secrets
  namespace: gorax
type: Opaque
stringData:
  DB_USER: "gorax"
  DB_PASSWORD: "<strong-database-password>"
  REDIS_PASSWORD: "<strong-redis-password>"
  CREDENTIAL_MASTER_KEY: "<base64-encoded-32-byte-key>"
  CREDENTIAL_KMS_KEY_ID: "arn:aws:kms:us-east-1:123456789:key/uuid"
  KRATOS_WEBHOOK_SECRET: "<random-webhook-secret>"
  AWS_ACCESS_KEY_ID: "<aws-access-key>"
  AWS_SECRET_ACCESS_KEY: "<aws-secret-key>"
  SENTRY_DSN: "<sentry-dsn-url>"
```

**Create secrets from environment:**

```bash
# Create secret from .env file
kubectl create secret generic gorax-secrets \
  --from-env-file=.env.prod \
  --namespace=gorax \
  --dry-run=client -o yaml | kubectl apply -f -
```

### API Server Deployment

```yaml
# api-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: gorax-api
  namespace: gorax
  labels:
    app: gorax-api
    version: v1.0.0
spec:
  replicas: 3
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxSurge: 1
      maxUnavailable: 0
  selector:
    matchLabels:
      app: gorax-api
  template:
    metadata:
      labels:
        app: gorax-api
        version: v1.0.0
      annotations:
        prometheus.io/scrape: "true"
        prometheus.io/port: "9090"
        prometheus.io/path: "/metrics"
    spec:
      serviceAccountName: gorax-api
      securityContext:
        runAsNonRoot: true
        runAsUser: 1000
        fsGroup: 1000
      containers:
      - name: api
        image: ghcr.io/stherrien/gorax:v1.0.0
        imagePullPolicy: IfNotPresent
        ports:
        - name: http
          containerPort: 8080
          protocol: TCP
        - name: metrics
          containerPort: 9090
          protocol: TCP
        envFrom:
        - configMapRef:
            name: gorax-config
        - secretRef:
            name: gorax-secrets
        resources:
          requests:
            cpu: 500m
            memory: 512Mi
          limits:
            cpu: 2000m
            memory: 2Gi
        livenessProbe:
          httpGet:
            path: /health
            port: http
          initialDelaySeconds: 30
          periodSeconds: 10
          timeoutSeconds: 5
          failureThreshold: 3
        readinessProbe:
          httpGet:
            path: /health/ready
            port: http
          initialDelaySeconds: 10
          periodSeconds: 5
          timeoutSeconds: 3
          failureThreshold: 2
        lifecycle:
          preStop:
            exec:
              command: ["/bin/sh", "-c", "sleep 15"]
        securityContext:
          allowPrivilegeEscalation: false
          readOnlyRootFilesystem: true
          capabilities:
            drop:
            - ALL
        volumeMounts:
        - name: tmp
          mountPath: /tmp
      volumes:
      - name: tmp
        emptyDir: {}
      affinity:
        podAntiAffinity:
          preferredDuringSchedulingIgnoredDuringExecution:
          - weight: 100
            podAffinityTerm:
              labelSelector:
                matchExpressions:
                - key: app
                  operator: In
                  values:
                  - gorax-api
              topologyKey: kubernetes.io/hostname
```

### API Service

```yaml
# api-service.yaml
apiVersion: v1
kind: Service
metadata:
  name: gorax-api
  namespace: gorax
  labels:
    app: gorax-api
spec:
  type: ClusterIP
  ports:
  - name: http
    port: 8080
    targetPort: http
    protocol: TCP
  - name: metrics
    port: 9090
    targetPort: metrics
    protocol: TCP
  selector:
    app: gorax-api
```

### Worker Deployment

```yaml
# worker-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: gorax-worker
  namespace: gorax
  labels:
    app: gorax-worker
spec:
  replicas: 5
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxSurge: 2
      maxUnavailable: 1
  selector:
    matchLabels:
      app: gorax-worker
  template:
    metadata:
      labels:
        app: gorax-worker
    spec:
      serviceAccountName: gorax-worker
      securityContext:
        runAsNonRoot: true
        runAsUser: 1000
        fsGroup: 1000
      containers:
      - name: worker
        image: ghcr.io/stherrien/gorax-worker:v1.0.0
        imagePullPolicy: IfNotPresent
        ports:
        - name: health
          containerPort: 8081
          protocol: TCP
        envFrom:
        - configMapRef:
            name: gorax-config
        - secretRef:
            name: gorax-secrets
        env:
        - name: AWS_SQS_QUEUE_URL
          value: "https://sqs.us-east-1.amazonaws.com/123456789/gorax-executions"
        resources:
          requests:
            cpu: 1000m
            memory: 1Gi
          limits:
            cpu: 2000m
            memory: 4Gi
        livenessProbe:
          httpGet:
            path: /health
            port: health
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health/ready
            port: health
          initialDelaySeconds: 10
          periodSeconds: 5
        securityContext:
          allowPrivilegeEscalation: false
          readOnlyRootFilesystem: true
          capabilities:
            drop:
            - ALL
```

### Horizontal Pod Autoscaler

```yaml
# hpa.yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: gorax-api-hpa
  namespace: gorax
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: gorax-api
  minReplicas: 3
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 80
  behavior:
    scaleDown:
      stabilizationWindowSeconds: 300
      policies:
      - type: Percent
        value: 50
        periodSeconds: 60
    scaleUp:
      stabilizationWindowSeconds: 0
      policies:
      - type: Percent
        value: 100
        periodSeconds: 30
      - type: Pods
        value: 2
        periodSeconds: 60
      selectPolicy: Max

---
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: gorax-worker-hpa
  namespace: gorax
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: gorax-worker
  minReplicas: 5
  maxReplicas: 20
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 75
  - type: External
    external:
      metric:
        name: sqs_messages_visible
        selector:
          matchLabels:
            queue_name: gorax-executions
      target:
        type: AverageValue
        averageValue: "30"
```

### Ingress

```yaml
# ingress.yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: gorax-ingress
  namespace: gorax
  annotations:
    kubernetes.io/ingress.class: "nginx"
    cert-manager.io/cluster-issuer: "letsencrypt-prod"
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
    nginx.ingress.kubernetes.io/force-ssl-redirect: "true"
    nginx.ingress.kubernetes.io/rate-limit: "100"
    nginx.ingress.kubernetes.io/limit-rps: "10"
    nginx.ingress.kubernetes.io/proxy-body-size: "10m"
    nginx.ingress.kubernetes.io/proxy-connect-timeout: "30"
    nginx.ingress.kubernetes.io/proxy-send-timeout: "60"
    nginx.ingress.kubernetes.io/proxy-read-timeout: "60"
spec:
  tls:
  - hosts:
    - api.gorax.io
    - app.gorax.io
    secretName: gorax-tls
  rules:
  - host: api.gorax.io
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: gorax-api
            port:
              number: 8080
  - host: app.gorax.io
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: gorax-frontend
            port:
              number: 80
```

### Persistent Volume Claims

```yaml
# pvc.yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: postgres-data
  namespace: gorax
spec:
  accessModes:
  - ReadWriteOnce
  storageClassName: fast-ssd
  resources:
    requests:
      storage: 200Gi

---
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: redis-data
  namespace: gorax
spec:
  accessModes:
  - ReadWriteOnce
  storageClassName: fast-ssd
  resources:
    requests:
      storage: 20Gi
```

### Deploy to Kubernetes

```bash
# Create namespace
kubectl apply -f namespace.yaml

# Create ConfigMap and Secrets
kubectl apply -f configmap.yaml
kubectl apply -f secrets.yaml

# Create PVCs
kubectl apply -f pvc.yaml

# Deploy databases (or use managed services)
kubectl apply -f postgres-deployment.yaml
kubectl apply -f redis-deployment.yaml

# Deploy Kratos
kubectl apply -f kratos-deployment.yaml

# Deploy API and Workers
kubectl apply -f api-deployment.yaml
kubectl apply -f api-service.yaml
kubectl apply -f worker-deployment.yaml

# Deploy HPAs
kubectl apply -f hpa.yaml

# Deploy Ingress
kubectl apply -f ingress.yaml

# Verify deployment
kubectl get pods -n gorax
kubectl get svc -n gorax
kubectl get ingress -n gorax

# Check logs
kubectl logs -f -n gorax -l app=gorax-api
kubectl logs -f -n gorax -l app=gorax-worker

# Scale manually if needed
kubectl scale deployment gorax-api --replicas=5 -n gorax
kubectl scale deployment gorax-worker --replicas=10 -n gorax
```

---

## Database Setup

### PostgreSQL Production Configuration

**PostgreSQL Configuration File** (`postgresql.conf`):

```conf
# Connection settings
listen_addresses = '*'
port = 5432
max_connections = 200

# Memory settings
shared_buffers = 4GB
effective_cache_size = 12GB
maintenance_work_mem = 1GB
work_mem = 32MB

# Write-Ahead Logging
wal_level = replica
max_wal_size = 2GB
min_wal_size = 1GB
wal_buffers = 16MB
checkpoint_completion_target = 0.9

# Replication
max_wal_senders = 10
max_replication_slots = 10
hot_standby = on

# Query tuning
random_page_cost = 1.1  # For SSD
effective_io_concurrency = 200

# Logging
log_destination = 'stderr'
logging_collector = on
log_directory = 'log'
log_filename = 'postgresql-%Y-%m-%d_%H%M%S.log'
log_rotation_age = 1d
log_rotation_size = 100MB
log_line_prefix = '%t [%p]: [%l-1] user=%u,db=%d,app=%a,client=%h '
log_checkpoints = on
log_connections = on
log_disconnections = on
log_duration = off
log_lock_waits = on
log_statement = 'ddl'
log_temp_files = 0

# Performance monitoring
shared_preload_libraries = 'pg_stat_statements'
pg_stat_statements.track = all
```

### Database Migration

**Using golang-migrate:**

```bash
# Install golang-migrate
brew install golang-migrate  # macOS
# or
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Run migrations
migrate -path migrations \
  -database "postgres://gorax:password@localhost:5432/gorax?sslmode=require" \
  up

# Check migration version
migrate -path migrations \
  -database "postgres://gorax:password@localhost:5432/gorax?sslmode=require" \
  version

# Rollback migration
migrate -path migrations \
  -database "postgres://gorax:password@localhost:5432/gorax?sslmode=require" \
  down 1
```

### Database Backup Strategy

**Automated Backup Script** (`backup-db.sh`):

```bash
#!/bin/bash
set -euo pipefail

# Configuration
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_NAME="${DB_NAME:-gorax}"
DB_USER="${DB_USER:-gorax}"
BACKUP_DIR="${BACKUP_DIR:-/backups/postgres}"
S3_BUCKET="${S3_BUCKET:-s3://gorax-backups/postgres}"
RETENTION_DAYS="${RETENTION_DAYS:-30}"

# Timestamp
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="gorax_${TIMESTAMP}.sql.gz"

# Create backup directory
mkdir -p "$BACKUP_DIR"

# Perform backup
echo "Starting backup of ${DB_NAME}..."
PGPASSWORD="$DB_PASSWORD" pg_dump \
  -h "$DB_HOST" \
  -p "$DB_PORT" \
  -U "$DB_USER" \
  -d "$DB_NAME" \
  --format=custom \
  --compress=9 \
  --file="${BACKUP_DIR}/${BACKUP_FILE}"

# Upload to S3
echo "Uploading to S3..."
aws s3 cp "${BACKUP_DIR}/${BACKUP_FILE}" "${S3_BUCKET}/${BACKUP_FILE}"

# Cleanup old backups
echo "Cleaning up old backups..."
find "$BACKUP_DIR" -name "gorax_*.sql.gz" -mtime +${RETENTION_DAYS} -delete

# Verify backup
if [ -f "${BACKUP_DIR}/${BACKUP_FILE}" ]; then
  SIZE=$(du -h "${BACKUP_DIR}/${BACKUP_FILE}" | cut -f1)
  echo "Backup completed successfully: ${BACKUP_FILE} (${SIZE})"
else
  echo "ERROR: Backup failed!"
  exit 1
fi
```

**Schedule via cron:**

```cron
# Run daily at 2 AM
0 2 * * * /path/to/backup-db.sh >> /var/log/backup.log 2>&1
```

### Connection Pooling with PgBouncer

```ini
# pgbouncer.ini
[databases]
gorax = host=postgres.gorax.svc.cluster.local port=5432 dbname=gorax

[pgbouncer]
listen_addr = 0.0.0.0
listen_port = 6432
auth_type = scram-sha-256
auth_file = /etc/pgbouncer/userlist.txt
pool_mode = transaction
max_client_conn = 1000
default_pool_size = 25
reserve_pool_size = 5
reserve_pool_timeout = 3
server_lifetime = 3600
server_idle_timeout = 600
log_connections = 1
log_disconnections = 1
log_pooler_errors = 1
```

**Deploy PgBouncer in Kubernetes:**

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: pgbouncer
  namespace: gorax
spec:
  replicas: 2
  selector:
    matchLabels:
      app: pgbouncer
  template:
    metadata:
      labels:
        app: pgbouncer
    spec:
      containers:
      - name: pgbouncer
        image: pgbouncer/pgbouncer:latest
        ports:
        - containerPort: 6432
        volumeMounts:
        - name: config
          mountPath: /etc/pgbouncer
      volumes:
      - name: config
        configMap:
          name: pgbouncer-config
```

---

## Redis Setup

### Redis Production Configuration

**Redis Configuration** (`redis.conf`):

```conf
# Network
bind 0.0.0.0
protected-mode yes
port 6379
tcp-backlog 511
timeout 0
tcp-keepalive 300

# Authentication
requirepass <strong-redis-password>

# Persistence
save 900 1
save 300 10
save 60 10000
stop-writes-on-bgsave-error yes
rdbcompression yes
rdbchecksum yes
dbfilename dump.rdb
dir /data

# AOF Persistence (recommended for durability)
appendonly yes
appendfilename "appendonly.aof"
appendfsync everysec
no-appendfsync-on-rewrite no
auto-aof-rewrite-percentage 100
auto-aof-rewrite-min-size 64mb

# Memory Management
maxmemory 2gb
maxmemory-policy allkeys-lru
maxmemory-samples 5

# Lazy Freeing
lazyfree-lazy-eviction yes
lazyfree-lazy-expire yes
lazyfree-lazy-server-del yes
replica-lazy-flush yes

# Logging
loglevel notice
logfile "/var/log/redis/redis.log"

# Slow Log
slowlog-log-slower-than 10000
slowlog-max-len 128

# Performance
hz 10
dynamic-hz yes

# Security
rename-command FLUSHDB ""
rename-command FLUSHALL ""
rename-command CONFIG ""
```

### Redis Sentinel for High Availability

```yaml
# redis-sentinel.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: redis-sentinel-config
  namespace: gorax
data:
  sentinel.conf: |
    port 26379
    dir /data
    sentinel monitor mymaster redis-master 6379 2
    sentinel down-after-milliseconds mymaster 5000
    sentinel parallel-syncs mymaster 1
    sentinel failover-timeout mymaster 10000
    sentinel auth-pass mymaster <redis-password>

---
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: redis-sentinel
  namespace: gorax
spec:
  serviceName: redis-sentinel
  replicas: 3
  selector:
    matchLabels:
      app: redis-sentinel
  template:
    metadata:
      labels:
        app: redis-sentinel
    spec:
      containers:
      - name: sentinel
        image: redis:7-alpine
        command:
        - redis-sentinel
        - /etc/redis/sentinel.conf
        ports:
        - containerPort: 26379
        volumeMounts:
        - name: config
          mountPath: /etc/redis
        - name: data
          mountPath: /data
      volumes:
      - name: config
        configMap:
          name: redis-sentinel-config
  volumeClaimTemplates:
  - metadata:
      name: data
    spec:
      accessModes: ["ReadWriteOnce"]
      resources:
        requests:
          storage: 5Gi
```

### Redis Cluster (for large scale)

```bash
# Create Redis Cluster with 6 nodes (3 masters + 3 replicas)
kubectl apply -f redis-cluster-statefulset.yaml

# Initialize cluster
kubectl exec -it redis-cluster-0 -n gorax -- redis-cli \
  --cluster create \
  $(kubectl get pods -l app=redis-cluster -o jsonpath='{range .items[*]}{.status.podIP}:6379 {end}') \
  --cluster-replicas 1 \
  --cluster-yes
```

---

## Environment Configuration

### Production Environment Variables

**Required Variables:**

```bash
# Application
APP_ENV=production
SERVER_ADDRESS=:8080

# Database
DB_HOST=postgres.production.rds.amazonaws.com
DB_PORT=5432
DB_USER=gorax
DB_PASSWORD=<strong-unique-password>
DB_NAME=gorax
DB_SSLMODE=require

# Redis
REDIS_ADDRESS=redis.production.cache.amazonaws.com:6379
REDIS_PASSWORD=<strong-redis-password>
REDIS_DB=0

# Ory Kratos
KRATOS_PUBLIC_URL=https://auth.gorax.io
KRATOS_ADMIN_URL=http://kratos.gorax.svc.cluster.local:4434
KRATOS_WEBHOOK_SECRET=<random-webhook-secret>

# Credentials
CREDENTIAL_USE_KMS=true
CREDENTIAL_KMS_KEY_ID=arn:aws:kms:us-east-1:123456789:key/uuid
CREDENTIAL_KMS_REGION=us-east-1

# AWS
AWS_REGION=us-east-1
AWS_ACCESS_KEY_ID=<access-key>
AWS_SECRET_ACCESS_KEY=<secret-key>
AWS_S3_BUCKET=gorax-artifacts-prod
AWS_SQS_QUEUE_URL=https://sqs.us-east-1.amazonaws.com/123456789/gorax-executions

# Observability
METRICS_ENABLED=true
METRICS_PORT=9090
TRACING_ENABLED=true
TRACING_ENDPOINT=jaeger-collector.observability.svc.cluster.local:4317
TRACING_SAMPLE_RATE=0.1
SENTRY_ENABLED=true
SENTRY_DSN=https://xxx@sentry.io/123456
SENTRY_ENVIRONMENT=production

# Security
CORS_ALLOWED_ORIGINS=https://app.gorax.io,https://admin.gorax.io
WEBSOCKET_ALLOWED_ORIGINS=https://app.gorax.io
SECURITY_HEADER_ENABLE_HSTS=true
SECURITY_HEADER_HSTS_MAX_AGE=63072000

# Worker
WORKER_CONCURRENCY=10
WORKER_MAX_CONCURRENCY_PER_TENANT=5
QUEUE_ENABLED=true
```

### Secret Management

#### Using AWS Secrets Manager

```bash
# Store secret
aws secretsmanager create-secret \
  --name gorax/production/db-password \
  --secret-string "your-secure-password" \
  --region us-east-1

# Retrieve secret
aws secretsmanager get-secret-value \
  --secret-id gorax/production/db-password \
  --query SecretString \
  --output text
```

#### Using HashiCorp Vault

```bash
# Enable KV secrets engine
vault secrets enable -path=gorax kv-v2

# Store secrets
vault kv put gorax/production \
  db_password="secure-password" \
  redis_password="redis-secure-password" \
  credential_master_key="base64-encoded-key"

# Retrieve secrets
vault kv get -field=db_password gorax/production
```

#### Kubernetes External Secrets Operator

```yaml
# external-secret.yaml
apiVersion: external-secrets.io/v1beta1
kind: ExternalSecret
metadata:
  name: gorax-secrets
  namespace: gorax
spec:
  refreshInterval: 1h
  secretStoreRef:
    name: aws-secrets-manager
    kind: SecretStore
  target:
    name: gorax-secrets
    creationPolicy: Owner
  data:
  - secretKey: DB_PASSWORD
    remoteRef:
      key: gorax/production/db-password
  - secretKey: REDIS_PASSWORD
    remoteRef:
      key: gorax/production/redis-password
  - secretKey: CREDENTIAL_MASTER_KEY
    remoteRef:
      key: gorax/production/credential-master-key
```

---

## Load Balancer & Ingress

### NGINX Configuration

**NGINX Config** (`nginx.conf`):

```nginx
user nginx;
worker_processes auto;
error_log /var/log/nginx/error.log warn;
pid /var/run/nginx.pid;

events {
    worker_connections 4096;
    use epoll;
    multi_accept on;
}

http {
    include /etc/nginx/mime.types;
    default_type application/octet-stream;

    # Logging
    log_format main '$remote_addr - $remote_user [$time_local] "$request" '
                    '$status $body_bytes_sent "$http_referer" '
                    '"$http_user_agent" "$http_x_forwarded_for" '
                    'rt=$request_time uct="$upstream_connect_time" '
                    'uht="$upstream_header_time" urt="$upstream_response_time"';

    access_log /var/log/nginx/access.log main;

    # Performance
    sendfile on;
    tcp_nopush on;
    tcp_nodelay on;
    keepalive_timeout 65;
    types_hash_max_size 2048;
    client_max_body_size 10m;

    # Gzip compression
    gzip on;
    gzip_vary on;
    gzip_proxied any;
    gzip_comp_level 6;
    gzip_types text/plain text/css text/xml text/javascript
               application/json application/javascript application/xml+rss
               application/rss+xml font/truetype font/opentype
               application/vnd.ms-fontobject image/svg+xml;

    # Rate limiting
    limit_req_zone $binary_remote_addr zone=api_limit:10m rate=100r/m;
    limit_req_zone $binary_remote_addr zone=login_limit:10m rate=5r/m;
    limit_conn_zone $binary_remote_addr zone=conn_limit:10m;

    # Upstream API servers
    upstream gorax_api {
        least_conn;
        server api-1.gorax.internal:8080 max_fails=3 fail_timeout=30s;
        server api-2.gorax.internal:8080 max_fails=3 fail_timeout=30s;
        server api-3.gorax.internal:8080 max_fails=3 fail_timeout=30s;
        keepalive 32;
    }

    # HTTP to HTTPS redirect
    server {
        listen 80 default_server;
        listen [::]:80 default_server;
        server_name _;
        return 301 https://$host$request_uri;
    }

    # HTTPS API Server
    server {
        listen 443 ssl http2;
        listen [::]:443 ssl http2;
        server_name api.gorax.io;

        # SSL Configuration
        ssl_certificate /etc/nginx/ssl/cert.pem;
        ssl_certificate_key /etc/nginx/ssl/key.pem;
        ssl_protocols TLSv1.2 TLSv1.3;
        ssl_ciphers 'ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256:ECDHE-ECDSA-AES256-GCM-SHA384:ECDHE-RSA-AES256-GCM-SHA384';
        ssl_prefer_server_ciphers off;
        ssl_session_cache shared:SSL:10m;
        ssl_session_timeout 10m;
        ssl_stapling on;
        ssl_stapling_verify on;

        # Security Headers
        add_header Strict-Transport-Security "max-age=63072000; includeSubDomains; preload" always;
        add_header X-Frame-Options "DENY" always;
        add_header X-Content-Type-Options "nosniff" always;
        add_header X-XSS-Protection "1; mode=block" always;
        add_header Referrer-Policy "strict-origin-when-cross-origin" always;
        add_header Content-Security-Policy "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; connect-src 'self' wss:;" always;

        # Rate limiting
        limit_req zone=api_limit burst=20 nodelay;
        limit_conn conn_limit 10;

        # Health check (no rate limiting)
        location = /health {
            access_log off;
            proxy_pass http://gorax_api;
            proxy_http_version 1.1;
            proxy_set_header Connection "";
        }

        # Login endpoint (stricter rate limit)
        location /api/auth/login {
            limit_req zone=login_limit burst=3 nodelay;
            proxy_pass http://gorax_api;
            proxy_http_version 1.1;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
            proxy_set_header Connection "";
        }

        # API endpoints
        location /api {
            proxy_pass http://gorax_api;
            proxy_http_version 1.1;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
            proxy_set_header Connection "";

            # Timeouts
            proxy_connect_timeout 30s;
            proxy_send_timeout 60s;
            proxy_read_timeout 60s;

            # Buffering
            proxy_buffering on;
            proxy_buffer_size 4k;
            proxy_buffers 8 4k;
            proxy_busy_buffers_size 8k;
        }

        # WebSocket support
        location /ws {
            proxy_pass http://gorax_api;
            proxy_http_version 1.1;
            proxy_set_header Upgrade $http_upgrade;
            proxy_set_header Connection "upgrade";
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;

            # WebSocket timeouts
            proxy_connect_timeout 7d;
            proxy_send_timeout 7d;
            proxy_read_timeout 7d;
        }
    }

    # Frontend Application
    server {
        listen 443 ssl http2;
        listen [::]:443 ssl http2;
        server_name app.gorax.io;

        ssl_certificate /etc/nginx/ssl/cert.pem;
        ssl_certificate_key /etc/nginx/ssl/key.pem;
        ssl_protocols TLSv1.2 TLSv1.3;

        root /usr/share/nginx/html;
        index index.html;

        # Security headers
        add_header Strict-Transport-Security "max-age=63072000; includeSubDomains; preload" always;
        add_header X-Frame-Options "DENY" always;
        add_header X-Content-Type-Options "nosniff" always;

        # Static assets caching
        location ~* \.(js|css|png|jpg|jpeg|gif|ico|svg|woff|woff2|ttf|eot)$ {
            expires 1y;
            add_header Cache-Control "public, immutable";
        }

        # SPA routing
        location / {
            try_files $uri $uri/ /index.html;
        }
    }
}
```

### AWS Application Load Balancer

**Terraform Configuration:**

```hcl
# alb.tf
resource "aws_lb" "gorax" {
  name               = "gorax-alb"
  internal           = false
  load_balancer_type = "application"
  security_groups    = [aws_security_group.alb.id]
  subnets            = var.public_subnet_ids

  enable_deletion_protection = true
  enable_http2              = true
  enable_cross_zone_load_balancing = true

  tags = {
    Name        = "gorax-alb"
    Environment = "production"
  }
}

resource "aws_lb_target_group" "api" {
  name     = "gorax-api-tg"
  port     = 8080
  protocol = "HTTP"
  vpc_id   = var.vpc_id

  health_check {
    enabled             = true
    path                = "/health"
    protocol            = "HTTP"
    matcher             = "200"
    interval            = 30
    timeout             = 5
    healthy_threshold   = 2
    unhealthy_threshold = 3
  }

  stickiness {
    type            = "lb_cookie"
    cookie_duration = 86400
    enabled         = true
  }

  deregistration_delay = 30

  tags = {
    Name = "gorax-api-target-group"
  }
}

resource "aws_lb_listener" "https" {
  load_balancer_arn = aws_lb.gorax.arn
  port              = "443"
  protocol          = "HTTPS"
  ssl_policy        = "ELBSecurityPolicy-TLS-1-2-2017-01"
  certificate_arn   = var.certificate_arn

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.api.arn
  }
}

resource "aws_lb_listener" "http_redirect" {
  load_balancer_arn = aws_lb.gorax.arn
  port              = "80"
  protocol          = "HTTP"

  default_action {
    type = "redirect"

    redirect {
      port        = "443"
      protocol    = "HTTPS"
      status_code = "HTTP_301"
    }
  }
}
```

---

## Monitoring & Observability

### Prometheus Configuration

```yaml
# prometheus-config.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: prometheus-config
  namespace: observability
data:
  prometheus.yml: |
    global:
      scrape_interval: 15s
      evaluation_interval: 15s
      external_labels:
        cluster: 'gorax-production'
        environment: 'production'

    alerting:
      alertmanagers:
      - static_configs:
        - targets:
          - alertmanager:9093

    rule_files:
    - /etc/prometheus/rules/*.yml

    scrape_configs:
    # API Server metrics
    - job_name: 'gorax-api'
      kubernetes_sd_configs:
      - role: pod
        namespaces:
          names:
          - gorax
      relabel_configs:
      - source_labels: [__meta_kubernetes_pod_label_app]
        action: keep
        regex: gorax-api
      - source_labels: [__meta_kubernetes_pod_annotation_prometheus_io_scrape]
        action: keep
        regex: true
      - source_labels: [__meta_kubernetes_pod_annotation_prometheus_io_path]
        action: replace
        target_label: __metrics_path__
        regex: (.+)
      - source_labels: [__address__, __meta_kubernetes_pod_annotation_prometheus_io_port]
        action: replace
        regex: ([^:]+)(?::\d+)?;(\d+)
        replacement: $1:$2
        target_label: __address__

    # Worker metrics
    - job_name: 'gorax-worker'
      kubernetes_sd_configs:
      - role: pod
        namespaces:
          names:
          - gorax
      relabel_configs:
      - source_labels: [__meta_kubernetes_pod_label_app]
        action: keep
        regex: gorax-worker

    # PostgreSQL metrics (using postgres_exporter)
    - job_name: 'postgres'
      static_configs:
      - targets:
        - postgres-exporter.gorax.svc.cluster.local:9187

    # Redis metrics (using redis_exporter)
    - job_name: 'redis'
      static_configs:
      - targets:
        - redis-exporter.gorax.svc.cluster.local:9121

    # Node metrics
    - job_name: 'node-exporter'
      kubernetes_sd_configs:
      - role: node
      relabel_configs:
      - source_labels: [__address__]
        regex: '(.*):10250'
        replacement: '${1}:9100'
        target_label: __address__
```

### Alert Rules

```yaml
# alert-rules.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: prometheus-rules
  namespace: observability
data:
  gorax-alerts.yml: |
    groups:
    - name: gorax_api
      interval: 30s
      rules:
      # API Health
      - alert: APIDown
        expr: up{job="gorax-api"} == 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "API instance {{ $labels.instance }} is down"
          description: "API has been down for more than 1 minute"

      # High Error Rate
      - alert: HighErrorRate
        expr: |
          (
            rate(http_requests_total{job="gorax-api",status=~"5.."}[5m])
            /
            rate(http_requests_total{job="gorax-api"}[5m])
          ) > 0.05
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High error rate detected"
          description: "Error rate is {{ $value | humanizePercentage }} over the last 5 minutes"

      # High Latency
      - alert: HighLatency
        expr: |
          histogram_quantile(0.95,
            rate(http_request_duration_seconds_bucket{job="gorax-api"}[5m])
          ) > 1
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High API latency detected"
          description: "P95 latency is {{ $value }}s over the last 5 minutes"

      # Database Connection Pool
      - alert: DatabaseConnectionPoolExhausted
        expr: |
          (
            sum(database_connections_in_use{job="gorax-api"})
            /
            sum(database_connections_max{job="gorax-api"})
          ) > 0.9
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Database connection pool near exhaustion"
          description: "{{ $value | humanizePercentage }} of database connections are in use"

    - name: gorax_worker
      interval: 30s
      rules:
      # Worker Health
      - alert: WorkerDown
        expr: up{job="gorax-worker"} == 0
        for: 2m
        labels:
          severity: warning
        annotations:
          summary: "Worker instance {{ $labels.instance }} is down"

      # Queue Backlog
      - alert: QueueBacklog
        expr: sqs_messages_visible{queue="gorax-executions"} > 1000
        for: 10m
        labels:
          severity: warning
        annotations:
          summary: "High SQS queue backlog"
          description: "{{ $value }} messages waiting in queue"

      # Execution Failures
      - alert: HighExecutionFailureRate
        expr: |
          (
            rate(workflow_executions_total{status="failed"}[10m])
            /
            rate(workflow_executions_total[10m])
          ) > 0.1
        for: 10m
        labels:
          severity: warning
        annotations:
          summary: "High workflow execution failure rate"
          description: "{{ $value | humanizePercentage }} of executions are failing"

    - name: infrastructure
      interval: 30s
      rules:
      # Database
      - alert: PostgreSQLDown
        expr: up{job="postgres"} == 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "PostgreSQL is down"

      - alert: HighDatabaseConnections
        expr: |
          pg_stat_database_numbackends / pg_settings_max_connections > 0.8
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High database connection usage"

      # Redis
      - alert: RedisDown
        expr: up{job="redis"} == 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "Redis is down"

      - alert: RedisMemoryHigh
        expr: redis_memory_used_bytes / redis_memory_max_bytes > 0.9
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Redis memory usage is high"

      # Disk Space
      - alert: DiskSpaceLow
        expr: |
          (
            node_filesystem_avail_bytes{mountpoint="/"}
            /
            node_filesystem_size_bytes{mountpoint="/"}
          ) < 0.1
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Low disk space on {{ $labels.instance }}"
          description: "Only {{ $value | humanizePercentage }} of disk space remaining"
```

### Grafana Dashboards

**Import pre-built dashboards:**

- **Go Application Dashboard**: ID 14061
- **PostgreSQL Dashboard**: ID 9628
- **Redis Dashboard**: ID 11835
- **Kubernetes Cluster**: ID 7249
- **NGINX**: ID 11199

**Custom Gorax Dashboard** (JSON):

```json
{
  "dashboard": {
    "title": "Gorax Overview",
    "panels": [
      {
        "title": "Request Rate",
        "targets": [
          {
            "expr": "rate(http_requests_total{job=\"gorax-api\"}[5m])"
          }
        ]
      },
      {
        "title": "Error Rate",
        "targets": [
          {
            "expr": "rate(http_requests_total{job=\"gorax-api\",status=~\"5..\"}[5m])"
          }
        ]
      },
      {
        "title": "P95 Latency",
        "targets": [
          {
            "expr": "histogram_quantile(0.95, rate(http_request_duration_seconds_bucket{job=\"gorax-api\"}[5m]))"
          }
        ]
      },
      {
        "title": "Active Executions",
        "targets": [
          {
            "expr": "workflow_executions_active"
          }
        ]
      }
    ]
  }
}
```

### OpenTelemetry Collector

```yaml
# otel-collector-config.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: otel-collector-config
  namespace: observability
data:
  config.yaml: |
    receivers:
      otlp:
        protocols:
          grpc:
            endpoint: 0.0.0.0:4317
          http:
            endpoint: 0.0.0.0:4318

    processors:
      batch:
        timeout: 10s
        send_batch_size: 1024
      memory_limiter:
        check_interval: 1s
        limit_mib: 512

    exporters:
      # Jaeger for tracing
      jaeger:
        endpoint: jaeger-collector.observability.svc.cluster.local:14250
        tls:
          insecure: true

      # Prometheus for metrics
      prometheus:
        endpoint: 0.0.0.0:8889

      # Loki for logs
      loki:
        endpoint: http://loki.observability.svc.cluster.local:3100/loki/api/v1/push

    service:
      pipelines:
        traces:
          receivers: [otlp]
          processors: [memory_limiter, batch]
          exporters: [jaeger]
        metrics:
          receivers: [otlp]
          processors: [memory_limiter, batch]
          exporters: [prometheus]
        logs:
          receivers: [otlp]
          processors: [memory_limiter, batch]
          exporters: [loki]
```

---

## CI/CD Pipeline

### GitHub Actions Production Deploy

```yaml
# .github/workflows/deploy-production.yml
name: Deploy to Production

on:
  push:
    tags:
      - 'v*'
  workflow_dispatch:
    inputs:
      version:
        description: 'Version to deploy (e.g., v1.0.0)'
        required: true

env:
  REGISTRY: ghcr.io
  IMAGE_NAME: ${{ github.repository }}

jobs:
  build:
    runs-on: ubuntu-latest
    permissions:
      contents: read
      packages: write
    outputs:
      version: ${{ steps.meta.outputs.version }}
      tags: ${{ steps.meta.outputs.tags }}

    steps:
    - name: Checkout code
      uses: actions/checkout@v4

    - name: Set up Docker Buildx
      uses: docker/setup-buildx-action@v3

    - name: Log in to Container Registry
      uses: docker/login-action@v3
      with:
        registry: ${{ env.REGISTRY }}
        username: ${{ github.actor }}
        password: ${{ secrets.GITHUB_TOKEN }}

    - name: Extract metadata
      id: meta
      uses: docker/metadata-action@v5
      with:
        images: ${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}
        tags: |
          type=semver,pattern={{version}}
          type=semver,pattern={{major}}.{{minor}}
          type=semver,pattern={{major}}
          type=sha,prefix=prod-

    - name: Build and push API image
      uses: docker/build-push-action@v5
      with:
        context: .
        file: ./Dockerfile
        push: true
        tags: ${{ steps.meta.outputs.tags }}
        labels: ${{ steps.meta.outputs.labels }}
        cache-from: type=gha
        cache-to: type=gha,mode=max
        platforms: linux/amd64,linux/arm64

    - name: Build and push Worker image
      uses: docker/build-push-action@v5
      with:
        context: .
        file: ./deployments/docker/Dockerfile.worker
        push: true
        tags: ${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}-worker:${{ steps.meta.outputs.version }}
        cache-from: type=gha
        cache-to: type=gha,mode=max

  deploy:
    needs: build
    runs-on: ubuntu-latest
    environment:
      name: production
      url: https://app.gorax.io
    steps:
    - name: Checkout code
      uses: actions/checkout@v4

    - name: Configure AWS credentials
      uses: aws-actions/configure-aws-credentials@v4
      with:
        aws-access-key-id: ${{ secrets.AWS_ACCESS_KEY_ID }}
        aws-secret-access-key: ${{ secrets.AWS_SECRET_ACCESS_KEY }}
        aws-region: us-east-1

    - name: Configure kubectl
      run: |
        aws eks update-kubeconfig --name gorax-production --region us-east-1

    - name: Deploy to Kubernetes
      run: |
        kubectl set image deployment/gorax-api \
          api=${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}:${{ needs.build.outputs.version }} \
          -n gorax

        kubectl set image deployment/gorax-worker \
          worker=${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}-worker:${{ needs.build.outputs.version }} \
          -n gorax

        kubectl rollout status deployment/gorax-api -n gorax --timeout=5m
        kubectl rollout status deployment/gorax-worker -n gorax --timeout=5m

    - name: Run database migrations
      run: |
        kubectl run migrate-${{ github.run_number }} \
          --image=${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}:${{ needs.build.outputs.version }} \
          --restart=Never \
          --namespace=gorax \
          --env-from=configmap/gorax-config \
          --env-from=secret/gorax-secrets \
          --command -- /app/gorax migrate up

        kubectl wait --for=condition=complete --timeout=5m \
          job/migrate-${{ github.run_number }} -n gorax

    - name: Smoke tests
      run: |
        sleep 30  # Wait for pods to stabilize

        # Health check
        kubectl run smoke-test-${{ github.run_number }} \
          --image=curlimages/curl:latest \
          --restart=Never \
          --namespace=gorax \
          --command -- curl -f http://gorax-api.gorax.svc.cluster.local:8080/health

    - name: Notify deployment
      if: always()
      uses: slackapi/slack-github-action@v1
      with:
        webhook: ${{ secrets.SLACK_WEBHOOK_URL }}
        payload: |
          {
            "text": "Production Deployment ${{ job.status }}",
            "blocks": [
              {
                "type": "section",
                "text": {
                  "type": "mrkdwn",
                  "text": "*Production Deployment*\nStatus: ${{ job.status }}\nVersion: ${{ needs.build.outputs.version }}"
                }
              }
            ]
          }
```

---

## Production Checklist

### Pre-Deployment Checklist

#### Security
- [ ] All secrets rotated and stored securely (AWS Secrets Manager/Vault)
- [ ] Database password is strong (min 16 characters, random)
- [ ] Redis password is set
- [ ] Credential master key is unique (32 bytes, base64-encoded)
- [ ] TLS certificates are valid and not expiring soon
- [ ] Firewall rules configured (whitelist only necessary ports)
- [ ] VPC and security groups properly configured
- [ ] IAM roles follow least privilege principle
- [ ] No hardcoded credentials in code or config files
- [ ] Security headers configured in NGINX/Ingress
- [ ] CORS origins are production URLs only (no localhost)
- [ ] WebSocket origins validated

#### Database
- [ ] PostgreSQL version 14+ with SSL enabled
- [ ] DB_SSLMODE set to `require` or higher
- [ ] Connection pooling configured (PgBouncer)
- [ ] Automated backups enabled (daily at minimum)
- [ ] Point-in-time recovery enabled
- [ ] Read replicas configured for scaling
- [ ] Database monitoring enabled
- [ ] Query performance insights enabled

#### Infrastructure
- [ ] Kubernetes cluster is production-grade (3+ nodes)
- [ ] High availability configured for all services
- [ ] Resource limits and requests set for all pods
- [ ] Persistent volumes configured with backups
- [ ] Load balancer health checks configured
- [ ] Auto-scaling policies defined (HPA)
- [ ] Node auto-scaling configured
- [ ] Multi-AZ deployment for resilience

#### Observability
- [ ] Prometheus scraping API and worker metrics
- [ ] Grafana dashboards configured and accessible
- [ ] Alert rules defined and tested
- [ ] Alertmanager configured (email, Slack, PagerDuty)
- [ ] Distributed tracing enabled (Jaeger/Zipkin)
- [ ] Sentry error tracking enabled
- [ ] Log aggregation configured (ELK/Loki)
- [ ] Uptime monitoring configured (external service)

#### Application
- [ ] APP_ENV set to `production`
- [ ] CREDENTIAL_USE_KMS enabled
- [ ] AWS KMS key created and accessible
- [ ] SQS queue created with DLQ
- [ ] S3 bucket created for artifacts
- [ ] Retention policies configured
- [ ] Rate limiting enabled
- [ ] Session timeout configured
- [ ] Webhook signature verification enabled
- [ ] RBAC roles and permissions configured

#### Testing
- [ ] Load testing performed (e.g., k6, Locust)
- [ ] Failover testing completed
- [ ] Database backup/restore tested
- [ ] Rollback procedure documented and tested
- [ ] Disaster recovery plan documented
- [ ] Runbook created for common issues

#### Documentation
- [ ] API documentation published
- [ ] Deployment guide updated
- [ ] Runbook created
- [ ] Architecture diagram updated
- [ ] Security documentation reviewed
- [ ] Incident response plan documented

### Post-Deployment Checklist

- [ ] All pods are running and healthy
- [ ] Health check endpoints return 200 OK
- [ ] Database migrations completed successfully
- [ ] Metrics are being collected
- [ ] Logs are being aggregated
- [ ] Alerts are firing correctly (test with intentional failure)
- [ ] SSL certificates are valid
- [ ] DNS records are correct
- [ ] Load balancer is distributing traffic
- [ ] WebSocket connections work
- [ ] User authentication works (Kratos)
- [ ] Workflow execution works end-to-end
- [ ] Webhooks are receiving events
- [ ] Email notifications are being sent
- [ ] Performance is acceptable (check latency)
- [ ] No elevated error rates
- [ ] Team notified of successful deployment

---

## Scaling Strategies

### Horizontal Scaling

#### API Server Scaling

**Manual Scaling:**

```bash
# Kubernetes
kubectl scale deployment gorax-api --replicas=10 -n gorax

# Docker Compose
docker-compose up --scale api=5
```

**Auto-scaling (HPA):**

```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: gorax-api-hpa
  namespace: gorax
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: gorax-api
  minReplicas: 3
  maxReplicas: 20
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 80
  - type: Pods
    pods:
      metric:
        name: http_requests_per_second
      target:
        type: AverageValue
        averageValue: "1000"
```

#### Worker Scaling

**Based on Queue Depth:**

```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: gorax-worker-hpa
  namespace: gorax
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: gorax-worker
  minReplicas: 5
  maxReplicas: 50
  metrics:
  - type: External
    external:
      metric:
        name: sqs_messages_visible
        selector:
          matchLabels:
            queue_name: gorax-executions
      target:
        type: AverageValue
        averageValue: "20"  # Scale up when avg > 20 messages per worker
  behavior:
    scaleUp:
      stabilizationWindowSeconds: 60
      policies:
      - type: Percent
        value: 100
        periodSeconds: 60
    scaleDown:
      stabilizationWindowSeconds: 300
      policies:
      - type: Percent
        value: 50
        periodSeconds: 60
```

### Database Scaling

#### Read Replicas

**PostgreSQL Read Replica Setup:**

```bash
# On primary
ALTER SYSTEM SET wal_level = replica;
ALTER SYSTEM SET max_wal_senders = 10;
SELECT pg_reload_conf();

# Create replication user
CREATE USER replicator REPLICATION LOGIN ENCRYPTED PASSWORD 'password';

# On replica
pg_basebackup -h primary-host -D /var/lib/postgresql/data -U replicator -P --wal-method=stream

# Configure recovery
cat > /var/lib/postgresql/data/standby.signal << EOF
standby_mode = 'on'
primary_conninfo = 'host=primary-host port=5432 user=replicator password=password'
EOF
```

**Application Configuration:**

```go
// Use read replica for read operations
type DBConfig struct {
    Primary     *sql.DB  // Write operations
    ReadReplica *sql.DB  // Read operations
}

func (s *Service) GetWorkflow(ctx context.Context, id string) (*Workflow, error) {
    // Use read replica
    return s.repo.GetFromReplica(ctx, id)
}

func (s *Service) CreateWorkflow(ctx context.Context, input CreateInput) (*Workflow, error) {
    // Use primary
    return s.repo.Create(ctx, input)
}
```

#### Connection Pooling

```yaml
# PgBouncer for connection pooling
apiVersion: apps/v1
kind: Deployment
metadata:
  name: pgbouncer
  namespace: gorax
spec:
  replicas: 3
  template:
    spec:
      containers:
      - name: pgbouncer
        image: pgbouncer/pgbouncer:latest
        env:
        - name: DATABASES_HOST
          value: postgres.gorax.svc.cluster.local
        - name: DATABASES_PORT
          value: "5432"
        - name: DATABASES_USER
          valueFrom:
            secretKeyRef:
              name: gorax-secrets
              key: DB_USER
        - name: DATABASES_PASSWORD
          valueFrom:
            secretKeyRef:
              name: gorax-secrets
              key: DB_PASSWORD
        - name: PGBOUNCER_POOL_MODE
          value: "transaction"
        - name: PGBOUNCER_MAX_CLIENT_CONN
          value: "1000"
        - name: PGBOUNCER_DEFAULT_POOL_SIZE
          value: "25"
```

### Redis Scaling

#### Redis Cluster

```yaml
# Redis Cluster StatefulSet
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: redis-cluster
  namespace: gorax
spec:
  serviceName: redis-cluster
  replicas: 6
  selector:
    matchLabels:
      app: redis-cluster
  template:
    metadata:
      labels:
        app: redis-cluster
    spec:
      containers:
      - name: redis
        image: redis:7-alpine
        command:
        - redis-server
        - /conf/redis.conf
        - --cluster-enabled
        - "yes"
        - --cluster-config-file
        - /data/nodes.conf
        - --cluster-node-timeout
        - "5000"
        ports:
        - containerPort: 6379
          name: client
        - containerPort: 16379
          name: gossip
        volumeMounts:
        - name: conf
          mountPath: /conf
        - name: data
          mountPath: /data
      volumes:
      - name: conf
        configMap:
          name: redis-cluster-config
  volumeClaimTemplates:
  - metadata:
      name: data
    spec:
      accessModes: ["ReadWriteOnce"]
      resources:
        requests:
          storage: 20Gi
```

### CDN for Frontend Assets

**CloudFront Distribution (Terraform):**

```hcl
resource "aws_cloudfront_distribution" "frontend" {
  enabled             = true
  is_ipv6_enabled    = true
  comment            = "Gorax Frontend Distribution"
  default_root_object = "index.html"

  origin {
    domain_name = aws_s3_bucket.frontend.bucket_regional_domain_name
    origin_id   = "S3-gorax-frontend"

    s3_origin_config {
      origin_access_identity = aws_cloudfront_origin_access_identity.frontend.cloudfront_access_identity_path
    }
  }

  default_cache_behavior {
    allowed_methods  = ["GET", "HEAD", "OPTIONS"]
    cached_methods   = ["GET", "HEAD"]
    target_origin_id = "S3-gorax-frontend"

    forwarded_values {
      query_string = false
      cookies {
        forward = "none"
      }
    }

    viewer_protocol_policy = "redirect-to-https"
    min_ttl                = 0
    default_ttl            = 3600
    max_ttl                = 86400
    compress               = true
  }

  # Cache static assets longer
  ordered_cache_behavior {
    path_pattern     = "/static/*"
    allowed_methods  = ["GET", "HEAD"]
    cached_methods   = ["GET", "HEAD"]
    target_origin_id = "S3-gorax-frontend"

    forwarded_values {
      query_string = false
      cookies {
        forward = "none"
      }
    }

    viewer_protocol_policy = "https-only"
    min_ttl                = 31536000
    default_ttl            = 31536000
    max_ttl                = 31536000
    compress               = true
  }

  restrictions {
    geo_restriction {
      restriction_type = "none"
    }
  }

  viewer_certificate {
    acm_certificate_arn      = var.certificate_arn
    ssl_support_method       = "sni-only"
    minimum_protocol_version = "TLSv1.2_2021"
  }

  tags = {
    Name        = "gorax-frontend-cdn"
    Environment = "production"
  }
}
```

---

## Troubleshooting

### Common Deployment Issues

#### Issue: Pods in CrashLoopBackOff

**Symptoms:**
```bash
$ kubectl get pods -n gorax
NAME                         READY   STATUS             RESTARTS   AGE
gorax-api-xxx                0/1     CrashLoopBackOff   5          5m
```

**Diagnosis:**
```bash
# Check logs
kubectl logs -n gorax gorax-api-xxx

# Describe pod for events
kubectl describe pod -n gorax gorax-api-xxx

# Check previous container logs
kubectl logs -n gorax gorax-api-xxx --previous
```

**Common Causes:**
1. **Database connection failure**: Check DB_HOST, credentials
2. **Missing environment variables**: Verify ConfigMap and Secrets
3. **Resource limits too low**: Increase memory/CPU limits
4. **Application error**: Check application logs

**Solution:**
```bash
# Fix environment variables
kubectl edit configmap gorax-config -n gorax
kubectl edit secret gorax-secrets -n gorax

# Restart deployment
kubectl rollout restart deployment gorax-api -n gorax
```

#### Issue: Database Migration Fails

**Symptoms:**
```
Error: pq: SSL is not enabled on the server
```

**Solution:**
```bash
# Update DB_SSLMODE
kubectl set env deployment/gorax-api DB_SSLMODE=disable -n gorax  # Development only

# Or enable SSL on PostgreSQL
# Edit postgresql.conf:
ssl = on
ssl_cert_file = '/etc/ssl/certs/server.crt'
ssl_key_file = '/etc/ssl/private/server.key'
```

#### Issue: High API Latency

**Diagnosis:**
```bash
# Check resource usage
kubectl top pods -n gorax

# Check HPA status
kubectl get hpa -n gorax

# Check database connections
kubectl exec -it postgres-0 -n gorax -- psql -U gorax -c "SELECT count(*) FROM pg_stat_activity;"
```

**Solutions:**
1. **Scale API pods**: `kubectl scale deployment gorax-api --replicas=10 -n gorax`
2. **Add read replicas**: Direct read queries to replicas
3. **Enable connection pooling**: Deploy PgBouncer
4. **Optimize queries**: Check slow query log
5. **Add caching**: Increase Redis memory

#### Issue: Worker Not Processing Jobs

**Diagnosis:**
```bash
# Check worker logs
kubectl logs -f -n gorax -l app=gorax-worker

# Check SQS queue depth
aws sqs get-queue-attributes \
  --queue-url $QUEUE_URL \
  --attribute-names ApproximateNumberOfMessages

# Check worker health
kubectl exec -it gorax-worker-xxx -n gorax -- wget -qO- localhost:8081/health
```

**Solutions:**
1. **Scale workers**: `kubectl scale deployment gorax-worker --replicas=20 -n gorax`
2. **Check AWS credentials**: Verify IAM role/credentials
3. **Increase concurrency**: Update WORKER_CONCURRENCY env var
4. **Check dead-letter queue**: Inspect failed messages

#### Issue: SSL/TLS Certificate Errors

**Symptoms:**
```
x509: certificate signed by unknown authority
```

**Solution:**
```bash
# For Kubernetes with cert-manager
kubectl get certificate -n gorax
kubectl describe certificate gorax-tls -n gorax

# Renew certificate
kubectl delete certificate gorax-tls -n gorax
# Cert-manager will auto-recreate

# For manual certificates (Let's Encrypt)
certbot renew --nginx

# Verify certificate
openssl s_client -connect api.gorax.io:443 -servername api.gorax.io
```

#### Issue: Out of Memory (OOM) Kills

**Symptoms:**
```bash
$ kubectl get pods -n gorax
NAME                         READY   STATUS    RESTARTS   AGE
gorax-api-xxx                0/1     OOMKilled 10         20m
```

**Diagnosis:**
```bash
# Check memory usage
kubectl top pod gorax-api-xxx -n gorax

# Check resource limits
kubectl describe pod gorax-api-xxx -n gorax | grep -A 5 Limits
```

**Solution:**
```yaml
# Increase memory limits
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    spec:
      containers:
      - name: api
        resources:
          requests:
            memory: "1Gi"
          limits:
            memory: "4Gi"
```

### Health Check Failures

#### Readiness Probe Failing

**Diagnosis:**
```bash
# Check readiness endpoint
kubectl exec -it gorax-api-xxx -n gorax -- wget -qO- localhost:8080/health/ready

# Check dependencies
kubectl exec -it gorax-api-xxx -n gorax -- wget -qO- localhost:8080/health/db
kubectl exec -it gorax-api-xxx -n gorax -- wget -qO- localhost:8080/health/redis
```

**Solution:**
1. Database unreachable: Check network policies, DNS
2. Redis unavailable: Check Redis pod status
3. Kratos unavailable: Check Kratos deployment

#### Liveness Probe Failing

**Increase timeout:**
```yaml
livenessProbe:
  httpGet:
    path: /health
    port: 8080
  initialDelaySeconds: 60  # Increase if slow startup
  periodSeconds: 10
  timeoutSeconds: 10       # Increase timeout
  failureThreshold: 5      # Increase tolerance
```

### Performance Issues

#### Slow Database Queries

**Identify slow queries:**
```sql
-- Enable slow query logging
ALTER SYSTEM SET log_min_duration_statement = 1000;  -- Log queries > 1s
SELECT pg_reload_conf();

-- View slow queries
SELECT query, calls, total_exec_time, mean_exec_time
FROM pg_stat_statements
ORDER BY mean_exec_time DESC
LIMIT 10;

-- Missing indexes
SELECT schemaname, tablename, attname, n_distinct, correlation
FROM pg_stats
WHERE schemaname = 'public'
AND correlation < 0.1;
```

**Solutions:**
1. Add indexes for frequently queried columns
2. Optimize query structure
3. Use EXPLAIN ANALYZE to understand query plans
4. Consider materialized views for complex aggregations

#### Redis Performance

**Check Redis stats:**
```bash
kubectl exec -it redis-0 -n gorax -- redis-cli INFO stats
kubectl exec -it redis-0 -n gorax -- redis-cli SLOWLOG GET 10
```

**Optimize:**
1. Increase max memory
2. Tune eviction policy
3. Enable pipelining for bulk operations
4. Consider Redis Cluster for horizontal scaling

### Disaster Recovery

#### Database Restore

```bash
# Stop API and workers to prevent writes
kubectl scale deployment gorax-api --replicas=0 -n gorax
kubectl scale deployment gorax-worker --replicas=0 -n gorax

# Download backup from S3
aws s3 cp s3://gorax-backups/postgres/gorax_20250101_020000.sql.gz ./backup.sql.gz

# Restore database
gunzip backup.sql.gz
kubectl exec -i postgres-0 -n gorax -- psql -U gorax < backup.sql

# Restart services
kubectl scale deployment gorax-api --replicas=3 -n gorax
kubectl scale deployment gorax-worker --replicas=5 -n gorax

# Verify
kubectl logs -f -n gorax -l app=gorax-api
```

#### Complete Cluster Failure

**Recovery Steps:**
1. Provision new Kubernetes cluster
2. Restore database from backup
3. Deploy infrastructure (Redis, Kratos)
4. Deploy application (API, workers)
5. Update DNS to point to new load balancer
6. Verify functionality
7. Monitor for issues

---

## Additional Resources

### Documentation
- [Kubernetes Documentation](https://kubernetes.io/docs/)
- [Docker Documentation](https://docs.docker.com/)
- [PostgreSQL Documentation](https://www.postgresql.org/docs/)
- [Redis Documentation](https://redis.io/documentation)
- [Ory Kratos Documentation](https://www.ory.sh/docs/kratos/)
- [Prometheus Documentation](https://prometheus.io/docs/)
- [OpenTelemetry Documentation](https://opentelemetry.io/docs/)

### Tools
- [k9s](https://k9scli.io/) - Kubernetes CLI manager
- [kubectx/kubens](https://github.com/ahmetb/kubectx) - Context/namespace switcher
- [stern](https://github.com/stern/stern) - Multi-pod log tailing
- [kubectl-tree](https://github.com/ahmetb/kubectl-tree) - Show resource hierarchy
- [Lens](https://k8slens.dev/) - Kubernetes IDE
- [Terraform](https://www.terraform.io/) - Infrastructure as Code
- [Helm](https://helm.sh/) - Kubernetes package manager

### Monitoring SaaS
- [Datadog](https://www.datadoghq.com/)
- [New Relic](https://newrelic.com/)
- [Grafana Cloud](https://grafana.com/products/cloud/)
- [Sentry](https://sentry.io/)
- [BetterStack](https://betterstack.com/)

### Support
For deployment issues:
1. Check logs: `kubectl logs -n gorax -l app=gorax-api`
2. Review this documentation
3. Search GitHub Issues: https://github.com/stherrien/gorax/issues
4. Contact DevOps team

---

**Document Version:** 1.0
**Last Updated:** 2026-01-01
**Maintained By:** Gorax DevOps Team
