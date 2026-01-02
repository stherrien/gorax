# Gorax Operational Troubleshooting Guide

**Version:** 1.0
**Last Updated:** 2026-01-01
**Audience:** DevOps Engineers, SREs, On-Call Engineers

This comprehensive guide helps diagnose and resolve production issues in Gorax, a workflow automation platform built with Go and React.

---

## Table of Contents

1. [Quick Diagnostics](#quick-diagnostics)
2. [Application Issues](#application-issues)
3. [Database Issues](#database-issues)
4. [Performance Problems](#performance-problems)
5. [Worker/Queue Issues](#workerqueue-issues)
6. [Authentication/Authorization](#authenticationauthorization)
7. [Integration Failures](#integration-failures)
8. [Monitoring and Alerts](#monitoring-and-alerts)
9. [Logs and Debugging](#logs-and-debugging)
10. [Common kubectl/Docker Commands](#common-kubectldocker-commands)
11. [Emergency Procedures](#emergency-procedures)
12. [Post-Mortem Template](#post-mortem-template)

---

## Quick Diagnostics

### Health Check Endpoints

Gorax exposes multiple health check endpoints for monitoring:

```bash
# Basic liveness check (is process alive?)
curl http://localhost:8080/health

# Readiness check (can accept traffic?)
curl http://localhost:8080/ready

# Worker health (detailed worker status)
curl http://localhost:8081/health

# Worker liveness (Kubernetes probe)
curl http://localhost:8081/health/live

# Worker readiness (Kubernetes probe)
curl http://localhost:8081/health/ready
```

**Expected Response (Healthy):**
```json
{
  "status": "ok",
  "timestamp": "2026-01-01T10:30:00Z",
  "checks": {
    "database": "healthy",
    "redis": "healthy"
  }
}
```

**Expected Response (Degraded):**
```json
{
  "status": "degraded",
  "timestamp": "2026-01-01T10:30:00Z",
  "checks": {
    "database": "unhealthy: connection refused",
    "redis": "healthy"
  }
}
```

### System Status Commands

```bash
# Check if API is running
lsof -ti:8080

# Check if worker is running
lsof -ti:8081

# Check process status
ps aux | grep gorax

# View active connections
ss -tunap | grep -E ':(8080|8081|5432|6379)'

# Check resource usage
top -p $(pgrep gorax)

# View system load
uptime
```

### Log Locations and Viewing

**Development Environment:**
```bash
# API logs (stdout)
tail -f /var/log/gorax/api.log

# Worker logs (stdout)
tail -f /var/log/gorax/worker.log

# Docker logs
docker logs -f gorax-api
docker logs -f gorax-worker

# Kubernetes logs
kubectl logs -f deployment/gorax-api
kubectl logs -f deployment/gorax-worker
```

**Structured Log Query Examples:**
```bash
# Filter by log level
jq 'select(.level=="ERROR")' < api.log

# Filter by tenant
jq 'select(.tenant_id=="tenant-123")' < api.log

# Filter by execution
jq 'select(.execution_id=="exec-456")' < api.log

# Filter by time range (last 5 minutes)
jq 'select(.time > (now - 300 | strftime("%Y-%m-%dT%H:%M:%SZ")))' < api.log
```

### Common Error Patterns

**Pattern Recognition:**

| Error Message | Component | Severity | Common Cause |
|--------------|-----------|----------|--------------|
| `connection refused` | Database/Redis | CRITICAL | Service down or network issue |
| `too many open connections` | Database | HIGH | Connection pool exhausted |
| `context deadline exceeded` | Any | MEDIUM | Timeout or slow dependency |
| `tenant at capacity` | Worker | MEDIUM | Concurrent execution limit reached |
| `credential decryption failed` | Credential Service | HIGH | KMS/encryption key issue |
| `signature verification failed` | Webhook | LOW | Invalid webhook signature |
| `queue depth exceeds threshold` | Queue | MEDIUM | Worker backlog growing |

---

## Application Issues

### API Responding with 500 Errors

**Symptoms:**
- HTTP 5xx responses
- Client applications failing
- Error rate spike in Prometheus metrics
- Sentry alerts firing

**Diagnosis:**

1. Check error logs:
```bash
# Recent errors
kubectl logs deployment/gorax-api --tail=100 | jq 'select(.level=="ERROR")'

# Count errors by type
kubectl logs deployment/gorax-api --tail=1000 | jq -r '.error' | sort | uniq -c | sort -rn
```

2. Check Sentry dashboard:
```bash
# View recent errors grouped by type
# Navigate to: https://sentry.io/organizations/your-org/issues/
```

3. Check HTTP metrics:
```promql
# Error rate by endpoint
rate(gorax_http_requests_total{status=~"5.."}[5m])

# Top failing endpoints
topk(5, sum by (path) (rate(gorax_http_requests_total{status=~"5.."}[5m])))
```

4. Check dependent services:
```bash
# Database connectivity
curl http://localhost:8080/ready | jq '.checks.database'

# Redis connectivity
curl http://localhost:8080/ready | jq '.checks.redis'
```

**Solution:**

**For Database Connection Issues:**
```bash
# Restart API pods (rolling restart)
kubectl rollout restart deployment/gorax-api

# Scale up replicas temporarily
kubectl scale deployment/gorax-api --replicas=5

# Check database connection pool
psql -h localhost -U postgres -d gorax -c "SELECT count(*) FROM pg_stat_activity;"
```

**For Application Errors:**
```bash
# Review recent deployments
kubectl rollout history deployment/gorax-api

# Rollback to previous version if needed
kubectl rollout undo deployment/gorax-api

# Check for panics in logs
kubectl logs deployment/gorax-api | grep -i "panic"
```

**For Resource Exhaustion:**
```bash
# Check pod memory usage
kubectl top pods -l app=gorax-api

# Check for memory leaks (Go pprof)
curl http://localhost:9090/debug/pprof/heap > heap.out
go tool pprof -http=:8090 heap.out
```

**Prevention:**
- Set appropriate resource limits in Kubernetes
- Implement circuit breakers for external dependencies
- Add retries with exponential backoff
- Monitor error rate with alerts (threshold: >1% error rate for 5 min)

---

### Workflows Not Executing

**Symptoms:**
- Workflows stuck in "pending" state
- No execution records being created
- Users reporting workflows not triggered
- Queue depth not changing

**Diagnosis:**

1. Check execution creation:
```bash
# Recent executions
psql -h localhost -U postgres -d gorax -c \
  "SELECT id, workflow_id, status, created_at
   FROM workflow_executions
   ORDER BY created_at DESC
   LIMIT 10;"

# Count by status
psql -h localhost -U postgres -d gorax -c \
  "SELECT status, count(*)
   FROM workflow_executions
   WHERE created_at > NOW() - INTERVAL '1 hour'
   GROUP BY status;"
```

2. Check worker status:
```bash
# Worker health
curl http://localhost:8081/health

# Active executions
curl http://localhost:8081/health | jq '.worker_info.active_executions'

# Queue depth
redis-cli -h localhost LLEN workflow:executions:queue
```

3. Check for errors in executor:
```bash
# Worker logs
kubectl logs deployment/gorax-worker --tail=100 | jq 'select(.msg | contains("execution"))'

# Execution errors
kubectl logs deployment/gorax-api | jq 'select(.msg | contains("failed to execute"))'
```

**Solution:**

**If Workers Are Down:**
```bash
# Check worker pod status
kubectl get pods -l app=gorax-worker

# Restart workers
kubectl rollout restart deployment/gorax-worker

# Scale up workers
kubectl scale deployment/gorax-worker --replicas=3
```

**If Queue Is Stuck:**
```bash
# Check SQS queue attributes (if using SQS)
aws sqs get-queue-attributes \
  --queue-url $QUEUE_URL \
  --attribute-names All

# Check messages in flight
redis-cli -h localhost LLEN workflow:executions:queue

# Drain and requeue (if Redis)
redis-cli -h localhost LRANGE workflow:executions:queue 0 -1
```

**If Database Lock Issues:**
```bash
# Check for locks
psql -h localhost -U postgres -d gorax -c \
  "SELECT pid, usename, query, state
   FROM pg_stat_activity
   WHERE state = 'active' AND query NOT LIKE '%pg_stat_activity%';"

# Kill long-running queries (if necessary)
psql -h localhost -U postgres -d gorax -c "SELECT pg_terminate_backend(PID);"
```

**Prevention:**
- Monitor queue depth with alerts (threshold: >1000 messages for 10 min)
- Set execution timeouts (default: 5 minutes)
- Implement dead-letter queue for failed executions
- Add workflow execution metrics to dashboard

---

### Worker Queue Backlog Growing

**Symptoms:**
- `gorax_queue_depth` metric increasing
- Execution latency increasing
- Worker pods at max capacity
- Messages piling up in SQS/Redis

**Diagnosis:**

1. Check queue metrics:
```promql
# Queue depth over time
gorax_queue_depth{queue="default"}

# Message processing rate
rate(gorax_workflow_executions_total[5m])

# Queue depth rate of change
deriv(gorax_queue_depth[5m])
```

2. Check worker capacity:
```bash
# Worker pod count
kubectl get pods -l app=gorax-worker | wc -l

# Active executions per worker
kubectl logs deployment/gorax-worker | jq '.worker_info.active_executions'

# Worker concurrency config
kubectl get configmap gorax-config -o json | jq '.data.WORKER_CONCURRENCY'
```

3. Check for slow executions:
```promql
# P95 execution duration
histogram_quantile(0.95, gorax_workflow_execution_duration_seconds)

# Slowest workflows
topk(5, avg by (workflow_id) (gorax_workflow_execution_duration_seconds))
```

**Solution:**

**Scale Workers Horizontally:**
```bash
# Increase worker replicas
kubectl scale deployment/gorax-worker --replicas=10

# Enable HPA (Horizontal Pod Autoscaler)
kubectl autoscale deployment/gorax-worker \
  --cpu-percent=70 \
  --min=3 \
  --max=20
```

**Increase Worker Concurrency:**
```bash
# Update ConfigMap
kubectl edit configmap gorax-config
# Change WORKER_CONCURRENCY: "10" -> "20"

# Restart workers to apply
kubectl rollout restart deployment/gorax-worker
```

**Optimize Slow Workflows:**
```bash
# Identify slow workflows
psql -h localhost -U postgres -d gorax -c \
  "SELECT workflow_id, AVG(duration) as avg_duration, COUNT(*) as count
   FROM workflow_executions
   WHERE completed_at > NOW() - INTERVAL '1 hour'
   GROUP BY workflow_id
   ORDER BY avg_duration DESC
   LIMIT 10;"

# Review workflow definitions
curl -H "X-Tenant-ID: tenant-123" \
  http://localhost:8080/api/v1/workflows/{workflow_id}
```

**Prevention:**
- Monitor queue depth trend with alerts
- Set up auto-scaling based on queue depth
- Implement workflow timeout policies
- Optimize slow workflow actions

---

### High Memory Usage

**Symptoms:**
- Pods being OOMKilled
- Memory usage > 80%
- Slow response times
- Garbage collection pauses

**Diagnosis:**

1. Check memory usage:
```bash
# Kubernetes pod memory
kubectl top pods -l app=gorax-api

# System memory
free -h

# Go memory stats
curl http://localhost:9090/debug/pprof/heap > heap.out
go tool pprof heap.out
```

2. Identify memory leaks:
```bash
# Heap profile over time
curl http://localhost:9090/debug/pprof/heap?seconds=30 > heap-30s.out

# Compare with baseline
go tool pprof -base=heap-baseline.out heap-30s.out

# Top memory consumers
go tool pprof -top heap.out
```

3. Check for goroutine leaks:
```bash
# Goroutine count
curl http://localhost:9090/debug/pprof/goroutine?debug=1

# Growing goroutines
watch -n 5 'curl -s http://localhost:9090/debug/pprof/goroutine?debug=1 | grep "goroutine profile:" | wc -l'
```

**Solution:**

**Immediate Action:**
```bash
# Restart high-memory pods
kubectl delete pod <pod-name>

# Scale up to distribute load
kubectl scale deployment/gorax-api --replicas=5
```

**Identify Root Cause:**
```bash
# Capture heap dump
curl http://localhost:9090/debug/pprof/heap > heap-$(date +%s).out

# Analyze with pprof
go tool pprof -http=:8090 heap-*.out

# Look for:
# 1. Large allocations
# 2. Retained objects
# 3. Goroutine leaks
```

**Common Causes and Fixes:**

**Connection Pool Leaks:**
```go
// BAD: Not closing connections
resp, _ := http.Get(url)
// Missing: defer resp.Body.Close()

// GOOD: Always close
resp, err := http.Get(url)
if err != nil {
    return err
}
defer resp.Body.Close()
```

**Unbounded Caches:**
```bash
# Check Redis cache size
redis-cli -h localhost INFO memory

# Clear old cache entries
redis-cli -h localhost --scan --pattern "cache:*" | xargs redis-cli DEL
```

**Large Request Payloads:**
```bash
# Add request size limits in middleware
kubectl edit configmap gorax-config
# Add: MAX_REQUEST_SIZE: "10MB"
```

**Prevention:**
- Set memory limits: `resources.limits.memory: 512Mi`
- Monitor heap growth trend
- Implement cache eviction policies
- Add memory alerts (threshold: >80% for 10 min)
- Use `context.Context` with timeouts
- Close all resources with `defer`

---

### High CPU Usage

**Symptoms:**
- CPU throttling
- Slow response times
- CPU usage > 80%
- Increased latency

**Diagnosis:**

1. Check CPU usage:
```bash
# Kubernetes pod CPU
kubectl top pods -l app=gorax-api

# System CPU
top -b -n 1 | head -20

# CPU profile
curl http://localhost:9090/debug/pprof/profile?seconds=30 > cpu.out
go tool pprof cpu.out
```

2. Identify hot paths:
```bash
# CPU profile with flame graph
go tool pprof -http=:8090 cpu.out

# Top functions by CPU
go tool pprof -top cpu.out

# Call graph
go tool pprof -png cpu.out > cpu-graph.png
```

**Solution:**

**Immediate Action:**
```bash
# Scale horizontally
kubectl scale deployment/gorax-api --replicas=5

# Check for CPU-intensive workflows
psql -h localhost -U postgres -d gorax -c \
  "SELECT workflow_id, COUNT(*) as exec_count
   FROM workflow_executions
   WHERE created_at > NOW() - INTERVAL '5 minutes'
   GROUP BY workflow_id
   ORDER BY exec_count DESC
   LIMIT 5;"
```

**Optimize Code:**

Common CPU bottlenecks:
- Inefficient JSON parsing (use streaming for large payloads)
- Regex compilation in hot path (compile once, reuse)
- Unnecessary string allocations (use `strings.Builder`)
- N+1 database queries (use batch queries)

**Prevention:**
- Set CPU limits: `resources.limits.cpu: 1000m`
- Monitor CPU percentiles
- Add CPU alerts (threshold: >80% for 10 min)
- Profile production workloads regularly
- Optimize hot code paths

---

### Slow API Responses

**Symptoms:**
- API latency > 1 second
- User complaints about slow UI
- High P95/P99 latencies
- Timeout errors

**Diagnosis:**

1. Check latency metrics:
```promql
# P95 latency by endpoint
histogram_quantile(0.95,
  rate(gorax_http_request_duration_seconds_bucket[5m])
)

# Slowest endpoints
topk(5, histogram_quantile(0.95,
  rate(gorax_http_request_duration_seconds_bucket{path!="/health"}[5m])
))
```

2. Check database performance:
```bash
# Slow queries
psql -h localhost -U postgres -d gorax -c \
  "SELECT query, calls, mean_exec_time, max_exec_time
   FROM pg_stat_statements
   ORDER BY mean_exec_time DESC
   LIMIT 10;"

# Active connections
psql -h localhost -U postgres -d gorax -c \
  "SELECT state, count(*) FROM pg_stat_activity GROUP BY state;"
```

3. Check distributed traces:
```bash
# View trace in Jaeger UI
# Navigate to: http://localhost:16686
# Search for slow traces (duration > 1s)
```

**Solution:**

**Database Optimization:**
```sql
-- Add missing indexes
CREATE INDEX CONCURRENTLY idx_executions_workflow_created
ON workflow_executions(workflow_id, created_at);

-- Analyze query plans
EXPLAIN ANALYZE SELECT * FROM workflow_executions
WHERE tenant_id = 'tenant-123' AND created_at > NOW() - INTERVAL '1 day';

-- Update statistics
ANALYZE workflow_executions;
```

**Add Caching:**
```bash
# Cache frequently accessed workflows
redis-cli -h localhost SET workflow:123 "$(curl localhost:8080/api/v1/workflows/123)"

# Set expiration
redis-cli -h localhost EXPIRE workflow:123 300
```

**Optimize N+1 Queries:**
```go
// BAD: N+1 query
for _, wf := range workflows {
    executions := getExecutions(wf.ID) // N queries
}

// GOOD: Batch query
workflowIDs := extractIDs(workflows)
executions := getExecutionsBatch(workflowIDs) // 1 query
```

**Prevention:**
- Add indexes on frequently queried columns
- Implement query result caching
- Use database query explain plans
- Monitor P95/P99 latencies with alerts (threshold: >2s for 5 min)
- Add database connection pooling

---

### Database Connection Exhaustion

**Symptoms:**
- "too many connections" errors
- API returning 500 errors
- Database refusing new connections
- Connection pool timeout errors

**Diagnosis:**

1. Check connection count:
```bash
# Current connections
psql -h localhost -U postgres -d gorax -c \
  "SELECT count(*) FROM pg_stat_activity;"

# Max connections
psql -h localhost -U postgres -d gorax -c "SHOW max_connections;"

# Connections by state
psql -h localhost -U postgres -d gorax -c \
  "SELECT state, count(*) FROM pg_stat_activity GROUP BY state;"

# Long-running connections
psql -h localhost -U postgres -d gorax -c \
  "SELECT pid, usename, application_name, state,
          NOW() - query_start as duration, query
   FROM pg_stat_activity
   WHERE state != 'idle'
   ORDER BY duration DESC;"
```

2. Check connection pool settings:
```bash
# API connection pool config
kubectl get configmap gorax-config -o json | jq '.data | {
  DB_MAX_OPEN_CONNS,
  DB_MAX_IDLE_CONNS,
  DB_CONN_MAX_LIFETIME
}'
```

**Solution:**

**Immediate Action:**
```bash
# Kill idle connections
psql -h localhost -U postgres -d gorax -c \
  "SELECT pg_terminate_backend(pid)
   FROM pg_stat_activity
   WHERE state = 'idle' AND state_change < NOW() - INTERVAL '5 minutes';"

# Restart API pods (one at a time)
kubectl delete pod <pod-name>
```

**Adjust Connection Pool:**
```bash
# Update ConfigMap
kubectl edit configmap gorax-config

# Set appropriate values:
# DB_MAX_OPEN_CONNS: "25"   (per pod)
# DB_MAX_IDLE_CONNS: "5"    (per pod)
# DB_CONN_MAX_LIFETIME: "5m"

# Restart to apply
kubectl rollout restart deployment/gorax-api
```

**Increase Database Max Connections:**
```bash
# For PostgreSQL
# Edit postgresql.conf:
# max_connections = 200

# Restart PostgreSQL
docker restart gorax-postgres
# OR
kubectl rollout restart statefulset/postgres
```

**Prevention:**
- Calculate pool size: `max_connections = (pod_count * max_open_conns) + buffer`
- Monitor connection pool usage
- Set connection lifetime limits
- Use connection pooling middleware (PgBouncer)
- Add alerts for connection pool exhaustion

---

### Redis Connection Failures

**Symptoms:**
- "connection refused" errors
- Cache misses causing slow responses
- Quota tracking failures
- Session management broken

**Diagnosis:**

1. Check Redis connectivity:
```bash
# Ping Redis
redis-cli -h localhost -p 6379 ping

# Check connection count
redis-cli -h localhost INFO clients | grep connected_clients

# Check memory usage
redis-cli -h localhost INFO memory | grep used_memory_human

# Check for errors
redis-cli -h localhost INFO stats | grep rejected_connections
```

2. Check application logs:
```bash
# Redis connection errors
kubectl logs deployment/gorax-api | jq 'select(.error | contains("redis"))'
```

**Solution:**

**If Redis Is Down:**
```bash
# Check Redis pod/container status
kubectl get pods -l app=redis

# Check Redis logs
kubectl logs -l app=redis --tail=100

# Restart Redis
kubectl rollout restart deployment/redis
```

**If Out of Memory:**
```bash
# Check memory usage
redis-cli -h localhost INFO memory

# Clear expired keys
redis-cli -h localhost --scan --pattern "*" | xargs -L 100 redis-cli DEL

# Increase maxmemory
redis-cli -h localhost CONFIG SET maxmemory 2gb

# Set eviction policy
redis-cli -h localhost CONFIG SET maxmemory-policy allkeys-lru
```

**If Connection Pool Exhausted:**
```bash
# Check max clients
redis-cli -h localhost CONFIG GET maxclients

# Increase max clients
redis-cli -h localhost CONFIG SET maxclients 10000
```

**Prevention:**
- Monitor Redis memory usage with alerts
- Set appropriate maxmemory and eviction policies
- Use Redis Sentinel for high availability
- Implement application-level fallbacks for cache misses
- Monitor Redis connection pool usage

---

### WebSocket Disconnections

**Symptoms:**
- Real-time updates not working
- Collaboration features broken
- Frequent reconnection attempts
- "websocket: close 1006" errors

**Diagnosis:**

1. Check WebSocket connections:
```bash
# Active WebSocket connections
curl http://localhost:8080/api/v1/ws/stats

# Connection errors in logs
kubectl logs deployment/gorax-api | jq 'select(.msg | contains("websocket"))'
```

2. Check network/load balancer:
```bash
# Check load balancer timeout settings
kubectl get service gorax-api -o yaml | grep timeout

# Check for connection drops
netstat -an | grep :8080 | grep CLOSE_WAIT
```

**Solution:**

**Increase Timeouts:**
```bash
# Update load balancer timeout
kubectl annotate service gorax-api \
  service.beta.kubernetes.io/aws-load-balancer-connection-idle-timeout="300"

# Update ingress timeout
kubectl annotate ingress gorax \
  nginx.ingress.kubernetes.io/proxy-read-timeout="300"
  nginx.ingress.kubernetes.io/proxy-send-timeout="300"
```

**Configure WebSocket Keepalive:**
```bash
# Update WebSocket config
kubectl edit configmap gorax-config

# Add:
# WEBSOCKET_PING_INTERVAL: "30s"
# WEBSOCKET_PONG_WAIT: "60s"
```

**Prevention:**
- Set appropriate WebSocket timeouts
- Implement ping/pong keepalive
- Add reconnection logic in client
- Monitor WebSocket connection count
- Use sticky sessions for WebSocket connections

---

### Failed Workflow Executions

**Symptoms:**
- Workflows completing with "failed" status
- User-reported execution errors
- High failure rate in metrics
- Errors in workflow logs

**Diagnosis:**

1. Check execution failures:
```bash
# Recent failures
psql -h localhost -U postgres -d gorax -c \
  "SELECT id, workflow_id, status, error, created_at
   FROM workflow_executions
   WHERE status = 'failed'
   ORDER BY created_at DESC
   LIMIT 20;"

# Failure rate by workflow
psql -h localhost -U postgres -d gorax -c \
  "SELECT workflow_id,
          COUNT(*) FILTER (WHERE status = 'failed') as failures,
          COUNT(*) as total,
          ROUND(100.0 * COUNT(*) FILTER (WHERE status = 'failed') / COUNT(*), 2) as failure_rate
   FROM workflow_executions
   WHERE created_at > NOW() - INTERVAL '1 hour'
   GROUP BY workflow_id
   HAVING COUNT(*) FILTER (WHERE status = 'failed') > 0
   ORDER BY failure_rate DESC;"
```

2. Check execution steps:
```bash
# Failed steps for execution
psql -h localhost -U postgres -d gorax -c \
  "SELECT step_id, action_type, status, error
   FROM execution_steps
   WHERE execution_id = 'exec-123'
   ORDER BY step_order;"
```

3. Check for common errors:
```bash
# Group errors by type
kubectl logs deployment/gorax-worker | \
  jq -r 'select(.level=="ERROR") | .error' | \
  sort | uniq -c | sort -rn
```

**Solution:**

**For Integration Failures:**
```bash
# Check credential validity
curl -H "X-Tenant-ID: tenant-123" \
  http://localhost:8080/api/v1/credentials/{credential_id}

# Test external API connectivity
curl -v https://api.external-service.com/health
```

**For Action Errors:**
```bash
# Review workflow definition
curl -H "X-Tenant-ID: tenant-123" \
  http://localhost:8080/api/v1/workflows/{workflow_id}

# Dry-run workflow
curl -X POST -H "X-Tenant-ID: tenant-123" \
  http://localhost:8080/api/v1/workflows/{workflow_id}/dry-run \
  -d '{"trigger_data": {...}}'
```

**For Timeout Errors:**
```bash
# Increase execution timeout
kubectl edit configmap gorax-config
# EXECUTION_TIMEOUT: "600s"

# Restart workers
kubectl rollout restart deployment/gorax-worker
```

**Prevention:**
- Implement retry logic for transient failures
- Add circuit breakers for external services
- Monitor failure rate with alerts (threshold: >5% for 10 min)
- Validate workflow definitions before deployment
- Add comprehensive error messages

---

### Credential Decryption Failures

**Symptoms:**
- "credential decryption failed" errors
- Workflows failing with auth errors
- KMS permission denied errors
- Credentials returning null values

**Diagnosis:**

1. Check credential service:
```bash
# Credential service logs
kubectl logs deployment/gorax-api | \
  jq 'select(.msg | contains("credential"))'

# Check KMS configuration
kubectl get configmap gorax-config -o json | jq '.data | {
  CREDENTIAL_USE_KMS,
  CREDENTIAL_KMS_KEY_ID,
  CREDENTIAL_KMS_REGION
}'
```

2. Test KMS access:
```bash
# Test KMS encryption
aws kms encrypt \
  --key-id $KMS_KEY_ID \
  --plaintext "test" \
  --region $KMS_REGION

# Test KMS decryption
aws kms decrypt \
  --ciphertext-blob fileb://test.encrypted \
  --region $KMS_REGION
```

3. Check credential access logs:
```bash
# Recent credential access
psql -h localhost -U postgres -d gorax -c \
  "SELECT credential_id, action, success, created_at
   FROM credential_access_log
   WHERE success = false
   ORDER BY created_at DESC
   LIMIT 20;"
```

**Solution:**

**For KMS Permission Issues:**
```bash
# Check IAM role/policy
aws iam get-role --role-name gorax-api-role

# Add KMS permissions to IAM policy
cat > kms-policy.json <<EOF
{
  "Version": "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Action": [
      "kms:Decrypt",
      "kms:Encrypt",
      "kms:GenerateDataKey"
    ],
    "Resource": "arn:aws:kms:region:account:key/key-id"
  }]
}
EOF

aws iam put-role-policy \
  --role-name gorax-api-role \
  --policy-name KMSAccess \
  --policy-document file://kms-policy.json

# Restart pods to pick up new permissions
kubectl rollout restart deployment/gorax-api
```

**For Master Key Issues (Development):**
```bash
# Verify master key is set
kubectl get secret gorax-secrets -o json | \
  jq -r '.data.CREDENTIAL_MASTER_KEY' | base64 -d

# Regenerate master key if needed
openssl rand -base64 32 > master-key.txt

# Update secret
kubectl create secret generic gorax-secrets \
  --from-literal=CREDENTIAL_MASTER_KEY=$(cat master-key.txt) \
  --dry-run=client -o yaml | kubectl apply -f -
```

**For Corrupted Credentials:**
```bash
# Rotate affected credential
curl -X POST -H "X-Tenant-ID: tenant-123" \
  http://localhost:8080/api/v1/credentials/{credential_id}/rotate \
  -d '{"new_value": "new-secret-value"}'
```

**Prevention:**
- Use KMS in production (not simple encryption)
- Monitor KMS API calls and errors
- Implement credential rotation policies
- Add credential access audit logging
- Test credential decryption in health checks

---

## Database Issues

### Connection Pool Exhausted

See [Database Connection Exhaustion](#database-connection-exhaustion) above.

---

### Slow Queries

**Symptoms:**
- API responses slow
- High database CPU
- Long-running queries in pg_stat_activity
- Query timeout errors

**Diagnosis:**

1. Identify slow queries:
```sql
-- Enable pg_stat_statements extension
CREATE EXTENSION IF NOT EXISTS pg_stat_statements;

-- Top 10 slowest queries by average time
SELECT
  query,
  calls,
  mean_exec_time,
  max_exec_time,
  stddev_exec_time,
  rows
FROM pg_stat_statements
ORDER BY mean_exec_time DESC
LIMIT 10;

-- Currently running queries
SELECT
  pid,
  usename,
  query,
  state,
  NOW() - query_start AS duration
FROM pg_stat_activity
WHERE state != 'idle'
  AND query NOT LIKE '%pg_stat_activity%'
ORDER BY duration DESC;
```

2. Analyze query plans:
```sql
-- Get execution plan
EXPLAIN ANALYZE
SELECT * FROM workflow_executions
WHERE tenant_id = 'tenant-123'
  AND created_at > NOW() - INTERVAL '7 days'
ORDER BY created_at DESC
LIMIT 100;

-- Look for:
-- - Seq Scan (should be Index Scan)
-- - High cost values
-- - Nested Loop with large counts
```

3. Check index usage:
```sql
-- Unused indexes
SELECT
  schemaname,
  tablename,
  indexname,
  idx_scan
FROM pg_stat_user_indexes
WHERE idx_scan = 0
  AND indexname NOT LIKE 'pg_toast%';

-- Missing indexes
SELECT
  schemaname,
  tablename,
  seq_scan,
  seq_tup_read,
  idx_scan,
  seq_tup_read / seq_scan AS avg_seq_tup
FROM pg_stat_user_tables
WHERE seq_scan > 0
ORDER BY seq_tup_read DESC
LIMIT 10;
```

**Solution:**

**Add Missing Indexes:**
```sql
-- Index on tenant_id for all tenant-scoped tables
CREATE INDEX CONCURRENTLY idx_workflow_executions_tenant
ON workflow_executions(tenant_id);

-- Composite index for common queries
CREATE INDEX CONCURRENTLY idx_executions_tenant_created
ON workflow_executions(tenant_id, created_at DESC);

-- Index for workflow status queries
CREATE INDEX CONCURRENTLY idx_executions_workflow_status
ON workflow_executions(workflow_id, status);

-- Partial index for active executions only
CREATE INDEX CONCURRENTLY idx_executions_active
ON workflow_executions(tenant_id, created_at)
WHERE status IN ('pending', 'running');
```

**Optimize Queries:**
```sql
-- BAD: Using OFFSET for pagination (slow for large offsets)
SELECT * FROM workflow_executions
ORDER BY created_at DESC
OFFSET 10000 LIMIT 100;

-- GOOD: Using cursor-based pagination
SELECT * FROM workflow_executions
WHERE created_at < '2026-01-01 10:00:00'
ORDER BY created_at DESC
LIMIT 100;

-- BAD: N+1 query pattern
-- (multiple queries in application loop)

-- GOOD: Join or batch query
SELECT
  w.id,
  w.name,
  COUNT(e.id) as execution_count
FROM workflows w
LEFT JOIN workflow_executions e ON w.id = e.workflow_id
WHERE w.tenant_id = 'tenant-123'
GROUP BY w.id, w.name;
```

**Update Statistics:**
```sql
-- Analyze specific table
ANALYZE workflow_executions;

-- Analyze all tables
VACUUM ANALYZE;

-- Auto-vacuum configuration
ALTER TABLE workflow_executions
SET (autovacuum_vacuum_scale_factor = 0.1);
```

**Prevention:**
- Review query plans for all major queries
- Add indexes based on query patterns
- Monitor slow query log
- Set up query performance alerts
- Use EXPLAIN ANALYZE in development

---

### Lock Contention

**Symptoms:**
- Queries waiting for locks
- "deadlock detected" errors
- High lock wait time
- Transactions timing out

**Diagnosis:**

1. Check current locks:
```sql
-- View blocking queries
SELECT
  blocked_locks.pid AS blocked_pid,
  blocked_activity.usename AS blocked_user,
  blocking_locks.pid AS blocking_pid,
  blocking_activity.usename AS blocking_user,
  blocked_activity.query AS blocked_query,
  blocking_activity.query AS blocking_query
FROM pg_catalog.pg_locks blocked_locks
JOIN pg_catalog.pg_stat_activity blocked_activity ON blocked_activity.pid = blocked_locks.pid
JOIN pg_catalog.pg_locks blocking_locks
  ON blocking_locks.locktype = blocked_locks.locktype
  AND blocking_locks.database IS NOT DISTINCT FROM blocked_locks.database
  AND blocking_locks.relation IS NOT DISTINCT FROM blocked_locks.relation
  AND blocking_locks.page IS NOT DISTINCT FROM blocked_locks.page
  AND blocking_locks.tuple IS NOT DISTINCT FROM blocked_locks.tuple
  AND blocking_locks.virtualxid IS NOT DISTINCT FROM blocked_locks.virtualxid
  AND blocking_locks.transactionid IS NOT DISTINCT FROM blocked_locks.transactionid
  AND blocking_locks.classid IS NOT DISTINCT FROM blocked_locks.classid
  AND blocking_locks.objid IS NOT DISTINCT FROM blocked_locks.objid
  AND blocking_locks.objsubid IS NOT DISTINCT FROM blocked_locks.objsubid
  AND blocking_locks.pid != blocked_locks.pid
JOIN pg_catalog.pg_stat_activity blocking_activity ON blocking_activity.pid = blocking_locks.pid
WHERE NOT blocked_locks.granted;

-- Lock wait times
SELECT
  pg_stat_activity.pid,
  pg_stat_activity.query,
  NOW() - pg_stat_activity.query_start AS duration,
  pg_locks.mode,
  pg_locks.locktype
FROM pg_stat_activity
JOIN pg_locks ON pg_stat_activity.pid = pg_locks.pid
WHERE NOT pg_locks.granted
ORDER BY duration DESC;
```

2. Check deadlock history:
```bash
# Check PostgreSQL logs for deadlocks
grep "deadlock detected" /var/log/postgresql/postgresql-*.log
```

**Solution:**

**Terminate Blocking Query:**
```sql
-- Find blocking PID from query above
SELECT pg_cancel_backend(12345);  -- Try graceful cancel first
SELECT pg_terminate_backend(12345);  -- Force kill if needed
```

**Reduce Lock Contention:**
```sql
-- Use row-level locking instead of table-level
SELECT * FROM workflows
WHERE id = 'wf-123'
FOR UPDATE NOWAIT;  -- Fail immediately if locked

-- Use shorter transactions
BEGIN;
-- Keep transaction scope minimal
UPDATE workflows SET updated_at = NOW() WHERE id = 'wf-123';
COMMIT;

-- Use optimistic locking
UPDATE workflows
SET name = 'New Name', version = version + 1
WHERE id = 'wf-123' AND version = 5;  -- Check version hasn't changed
```

**Prevention:**
- Keep transactions short
- Acquire locks in consistent order
- Use row-level locks when possible
- Avoid long-running transactions
- Monitor lock wait times with alerts

---

### Disk Space Full

**Symptoms:**
- "No space left on device" errors
- Database writes failing
- WAL segments growing
- Disk usage > 90%

**Diagnosis:**

1. Check disk usage:
```bash
# Overall disk usage
df -h

# Database directory size
du -sh /var/lib/postgresql/data

# Largest tables
psql -h localhost -U postgres -d gorax -c \
  "SELECT
    schemaname,
    tablename,
    pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) AS size
   FROM pg_tables
   WHERE schemaname NOT IN ('pg_catalog', 'information_schema')
   ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC
   LIMIT 10;"

# WAL size
du -sh /var/lib/postgresql/data/pg_wal
```

**Solution:**

**Immediate Actions:**
```bash
# Clean up old logs
find /var/log -name "*.log.*" -mtime +7 -delete

# Vacuum to reclaim space
psql -h localhost -U postgres -d gorax -c "VACUUM FULL workflow_executions;"

# Remove old WAL files (PostgreSQL does this automatically, but check)
psql -h localhost -U postgres -d gorax -c "SELECT pg_switch_wal();"
```

**Archive Old Data:**
```sql
-- Move old executions to archive table
CREATE TABLE workflow_executions_archive (LIKE workflow_executions);

INSERT INTO workflow_executions_archive
SELECT * FROM workflow_executions
WHERE created_at < NOW() - INTERVAL '90 days';

DELETE FROM workflow_executions
WHERE created_at < NOW() - INTERVAL '90 days';

VACUUM workflow_executions;
```

**Increase Disk Space:**
```bash
# AWS EBS volume expansion
aws ec2 modify-volume --volume-id vol-xxx --size 100

# Wait for modification to complete
aws ec2 describe-volumes-modifications --volume-id vol-xxx

# Expand filesystem
sudo resize2fs /dev/xvdf

# Verify
df -h
```

**Prevention:**
- Set up disk space monitoring with alerts (threshold: >80%)
- Implement data retention policies
- Enable auto-vacuum
- Archive old data regularly
- Monitor table/index bloat

---

### Migration Failures

**Symptoms:**
- Migration script errors
- Database schema out of sync
- Application errors after deployment
- Missing tables/columns

**Diagnosis:**

1. Check migration status:
```bash
# Check applied migrations
psql -h localhost -U postgres -d gorax -c \
  "SELECT * FROM schema_migrations ORDER BY version;"

# Check migration errors
grep -i error /var/log/postgresql/postgresql-*.log
```

2. Compare schema:
```bash
# Dump schema
pg_dump -h localhost -U postgres -d gorax --schema-only > schema.sql

# Compare with expected schema
diff schema.sql expected-schema.sql
```

**Solution:**

**Retry Failed Migration:**
```bash
# Rollback failed migration
psql -h localhost -U postgres -d gorax < migrations/rollback/020_marketplace.sql

# Re-apply migration
psql -h localhost -U postgres -d gorax < migrations/020_marketplace.sql

# Verify
psql -h localhost -U postgres -d gorax -c "\d marketplace_templates"
```

**Manual Migration:**
```sql
-- Start transaction
BEGIN;

-- Apply migration steps manually
CREATE TABLE IF NOT EXISTS marketplace_templates (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  -- ... rest of schema
);

-- Mark as applied
INSERT INTO schema_migrations (version, applied_at)
VALUES ('020', NOW());

-- Commit if successful
COMMIT;
```

**Prevention:**
- Test migrations in staging first
- Use transactional migrations (BEGIN/COMMIT)
- Have rollback scripts ready
- Lock migrations during deployment
- Backup before major migrations

---

### Replication Lag (If Using Replicas)

**Symptoms:**
- Stale data in read queries
- Replication lag alerts firing
- Replica behind primary by > 1 minute
- Read replica connection errors

**Diagnosis:**

1. Check replication status:
```sql
-- On primary
SELECT
  client_addr,
  state,
  sent_lsn,
  write_lsn,
  flush_lsn,
  replay_lsn,
  sync_state,
  pg_wal_lsn_diff(sent_lsn, replay_lsn) AS lag_bytes
FROM pg_stat_replication;

-- On replica
SELECT
  NOW() - pg_last_xact_replay_timestamp() AS replication_lag;
```

2. Check replica load:
```bash
# CPU/memory on replica
kubectl top pod postgres-replica-0
```

**Solution:**

**If Replica Overloaded:**
```bash
# Scale out read replicas
kubectl scale statefulset postgres-replica --replicas=3

# Add connection pooling (PgBouncer)
kubectl apply -f pgbouncer.yaml
```

**If Network Issues:**
```bash
# Check network between primary and replica
kubectl exec postgres-0 -- ping postgres-replica-0

# Increase wal_sender_timeout
psql -h postgres-0 -U postgres -c \
  "ALTER SYSTEM SET wal_sender_timeout = '60s';"
```

**Prevention:**
- Monitor replication lag with alerts (threshold: >30s)
- Ensure replica has sufficient resources
- Use PgBouncer for connection pooling
- Distribute read queries across replicas
- Set up automatic failover (Patroni/Stolon)

---

### Table Bloat

**Symptoms:**
- Tables larger than expected
- Slow query performance
- Disk space usage growing
- Index size excessive

**Diagnosis:**

1. Check table bloat:
```sql
-- Table bloat
SELECT
  schemaname,
  tablename,
  pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) AS total_size,
  pg_size_pretty(pg_relation_size(schemaname||'.'||tablename)) AS table_size,
  round((pg_relation_size(schemaname||'.'||tablename)::float /
         NULLIF(pg_total_relation_size(schemaname||'.'||tablename), 0)) * 100, 2) AS table_percent
FROM pg_tables
WHERE schemaname NOT IN ('pg_catalog', 'information_schema')
ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC
LIMIT 10;

-- Dead tuples
SELECT
  schemaname,
  tablename,
  n_live_tup,
  n_dead_tup,
  round(n_dead_tup::float / NULLIF(n_live_tup + n_dead_tup, 0) * 100, 2) AS dead_percent,
  last_vacuum,
  last_autovacuum
FROM pg_stat_user_tables
WHERE n_dead_tup > 0
ORDER BY n_dead_tup DESC;
```

**Solution:**

**Run VACUUM:**
```sql
-- Regular vacuum (concurrent)
VACUUM workflow_executions;

-- Full vacuum (locks table, reclaims space)
VACUUM FULL workflow_executions;

-- Analyze after vacuum
ANALYZE workflow_executions;
```

**Tune Autovacuum:**
```sql
-- Per-table autovacuum settings
ALTER TABLE workflow_executions SET (
  autovacuum_vacuum_threshold = 50,
  autovacuum_vacuum_scale_factor = 0.1,
  autovacuum_analyze_threshold = 50,
  autovacuum_analyze_scale_factor = 0.05
);

-- Global settings (postgresql.conf)
ALTER SYSTEM SET autovacuum_max_workers = 4;
ALTER SYSTEM SET autovacuum_naptime = '10s';
```

**Prevention:**
- Enable autovacuum (default: on)
- Tune autovacuum for high-write tables
- Run manual VACUUM during maintenance windows
- Monitor dead tuple percentage
- Avoid long-running transactions

---

### Index Corruption

**Symptoms:**
- "invalid page in block" errors
- Query crashes
- Inconsistent query results
- Database restart failures

**Diagnosis:**

1. Check for corruption:
```sql
-- Check table integrity
SELECT pg_catalog.pg_check_visible('workflow_executions'::regclass);

-- Reindex with CONCURRENTLY to check for issues
REINDEX INDEX CONCURRENTLY idx_executions_tenant;
```

2. Check logs:
```bash
grep -i "corrupt\|invalid" /var/log/postgresql/postgresql-*.log
```

**Solution:**

**Rebuild Corrupted Index:**
```sql
-- Drop and recreate index
DROP INDEX CONCURRENTLY idx_executions_tenant;

CREATE INDEX CONCURRENTLY idx_executions_tenant
ON workflow_executions(tenant_id);

-- Or use REINDEX
REINDEX INDEX CONCURRENTLY idx_executions_tenant;
```

**Full Database Check:**
```bash
# Stop application
kubectl scale deployment/gorax-api --replicas=0

# Run integrity check
psql -h localhost -U postgres -d gorax -c \
  "SELECT datname FROM pg_database WHERE datallowconn = true;" | \
  grep -v "^-" | grep -v "datname" | \
  xargs -I {} psql -h localhost -U postgres -d {} -c "VACUUM FULL ANALYZE;"

# Restart application
kubectl scale deployment/gorax-api --replicas=3
```

**Prevention:**
- Regular REINDEX maintenance
- Monitor for corruption warnings
- Use checksums (initdb --data-checksums)
- Regular backups
- Avoid sudden power loss

---

## Performance Problems

### Identifying Bottlenecks

**Tools and Approaches:**

1. **Application Profiling:**
```bash
# CPU profile (30 seconds)
curl http://localhost:9090/debug/pprof/profile?seconds=30 > cpu.out

# Memory profile
curl http://localhost:9090/debug/pprof/heap > heap.out

# Goroutine profile
curl http://localhost:9090/debug/pprof/goroutine > goroutine.out

# Analyze with pprof web UI
go tool pprof -http=:8090 cpu.out
```

2. **Database Profiling:**
```sql
-- Enable query logging
ALTER SYSTEM SET log_min_duration_statement = 100;  -- Log queries > 100ms
SELECT pg_reload_conf();

-- Check slow queries
SELECT * FROM pg_stat_statements
ORDER BY mean_exec_time DESC
LIMIT 10;
```

3. **Distributed Tracing:**
```bash
# View trace in Jaeger
# Navigate to: http://localhost:16686
# Search for traces with high duration
# Look at trace waterfall to identify slow spans
```

**Common Bottlenecks:**
- Database queries (N+1 queries, missing indexes)
- External API calls (no timeout, no retry, no circuit breaker)
- Serialization/deserialization (large JSON payloads)
- Memory allocations (string concatenation in loops)
- Lock contention (global mutexes, database locks)

---

### Using Prometheus Metrics

**Key Metrics to Monitor:**

1. **HTTP Request Metrics:**
```promql
# Request rate by endpoint
rate(gorax_http_requests_total[5m])

# Error rate
rate(gorax_http_requests_total{status=~"5.."}[5m])

# P95 latency
histogram_quantile(0.95, rate(gorax_http_request_duration_seconds_bucket[5m]))

# Slow endpoints (P95 > 1s)
topk(10, histogram_quantile(0.95,
  rate(gorax_http_request_duration_seconds_bucket[5m])
)) > 1
```

2. **Workflow Metrics:**
```promql
# Workflow execution rate
rate(gorax_workflow_executions_total[5m])

# Success rate
sum(rate(gorax_workflow_executions_total{status="completed"}[5m])) /
sum(rate(gorax_workflow_executions_total[5m]))

# P95 execution duration
histogram_quantile(0.95, rate(gorax_workflow_execution_duration_seconds_bucket[5m]))

# Queue depth
gorax_queue_depth
```

3. **Resource Metrics:**
```promql
# CPU usage
rate(process_cpu_seconds_total[5m])

# Memory usage
process_resident_memory_bytes

# Goroutine count
go_goroutines

# GC pause time
rate(go_gc_duration_seconds_sum[5m])
```

**Viewing Metrics:**
```bash
# Query Prometheus API
curl 'http://localhost:9090/api/v1/query?query=gorax_queue_depth'

# View in Prometheus UI
# Navigate to: http://localhost:9090/graph
```

---

### Analyzing Grafana Dashboards

**Pre-built Dashboards:**

1. **Gorax Overview Dashboard:**
   - System health status
   - Request rate and error rate
   - Queue depth trends
   - Top workflows by execution count

2. **Performance Dashboard:**
   - P50/P95/P99 latency percentiles
   - Slow endpoint identification
   - Database query performance
   - API throughput

3. **Resource Dashboard:**
   - CPU/Memory usage by pod
   - Goroutine count
   - GC pause times
   - Connection pool usage

**Dashboard Navigation:**
```bash
# Access Grafana
# URL: http://localhost:3000
# Default credentials: admin/admin

# Import dashboard
# Dashboard > Import > Upload JSON file
```

---

### Database Query Analysis

See [Slow Queries](#slow-queries) above for detailed analysis.

---

### Memory Profiling (Go pprof)

**Capture Memory Profile:**
```bash
# Heap profile
curl http://localhost:9090/debug/pprof/heap > heap.out

# Heap profile with debug info
curl http://localhost:9090/debug/pprof/heap?debug=1 > heap-debug.txt

# Allocations profile
curl http://localhost:9090/debug/pprof/allocs > allocs.out
```

**Analyze Profile:**
```bash
# Interactive analysis
go tool pprof heap.out

# Commands in pprof:
# top       - Show top memory consumers
# list Func - Show source code for Func
# web       - Open graph in browser
# png       - Generate graph image

# Web UI
go tool pprof -http=:8090 heap.out
```

**Compare Profiles Over Time:**
```bash
# Baseline
curl http://localhost:9090/debug/pprof/heap > heap-baseline.out

# Wait and capture again
sleep 300
curl http://localhost:9090/debug/pprof/heap > heap-after.out

# Compare (shows growth)
go tool pprof -base=heap-baseline.out heap-after.out
```

---

### CPU Profiling

**Capture CPU Profile:**
```bash
# 30-second CPU profile
curl http://localhost:9090/debug/pprof/profile?seconds=30 > cpu.out

# Profile specific endpoint under load
ab -n 10000 -c 100 http://localhost:8080/api/v1/workflows &
curl http://localhost:9090/debug/pprof/profile?seconds=30 > cpu-under-load.out
```

**Analyze Profile:**
```bash
# Interactive
go tool pprof cpu.out

# Web UI with flame graph
go tool pprof -http=:8090 cpu.out

# Generate flame graph
go tool pprof -web cpu.out
```

**Interpreting Results:**
- Look for functions consuming > 10% CPU
- Check for inefficient loops
- Identify regex compilation in hot paths
- Look for excessive allocations

---

### Goroutine Leaks

**Detect Goroutine Leaks:**
```bash
# Current goroutine count
curl http://localhost:9090/debug/pprof/goroutine?debug=1

# Monitor goroutine growth
watch -n 5 'curl -s http://localhost:9090/debug/pprof/goroutine?debug=1 | head -1'

# Goroutine profile
curl http://localhost:9090/debug/pprof/goroutine > goroutine.out
go tool pprof -http=:8090 goroutine.out
```

**Common Causes:**
```go
// BAD: Goroutine leak - channel never closed
go func() {
    for msg := range ch {  // Blocks forever if ch never closed
        process(msg)
    }
}()

// GOOD: Use context for cancellation
go func() {
    for {
        select {
        case msg := <-ch:
            process(msg)
        case <-ctx.Done():
            return
        }
    }
}()

// BAD: HTTP client without timeout
resp, _ := http.Get(url)  // Can hang forever

// GOOD: Use context with timeout
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()
req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
resp, _ := http.DefaultClient.Do(req)
```

**Fix Goroutine Leak:**
```bash
# Restart affected pods
kubectl rollout restart deployment/gorax-api

# Deploy fix and monitor
kubectl apply -f deployment.yaml
watch -n 5 'curl -s http://localhost:9090/debug/pprof/goroutine?debug=1 | head -1'
```

---

## Worker/Queue Issues

### Messages Stuck in Queue

**Symptoms:**
- Queue depth not decreasing
- Messages not being processed
- Worker logs show no activity
- Execution status stuck in "pending"

**Diagnosis:**

1. Check queue status:
```bash
# SQS queue attributes
aws sqs get-queue-attributes \
  --queue-url $QUEUE_URL \
  --attribute-names All \
  --output json | jq '{
    ApproximateNumberOfMessages,
    ApproximateNumberOfMessagesNotVisible,
    ApproximateNumberOfMessagesDelayed
  }'

# Redis queue (if using Redis)
redis-cli -h localhost LLEN workflow:executions:queue
redis-cli -h localhost LRANGE workflow:executions:queue 0 10
```

2. Check worker status:
```bash
# Worker health
curl http://localhost:8081/health

# Worker logs
kubectl logs deployment/gorax-worker --tail=100
```

3. Check visibility timeout:
```bash
# SQS visibility timeout
aws sqs get-queue-attributes \
  --queue-url $QUEUE_URL \
  --attribute-names VisibilityTimeout
```

**Solution:**

**If Messages Are Invisible (Being Processed):**
```bash
# Wait for visibility timeout to expire
# Or restart workers to release messages
kubectl rollout restart deployment/gorax-worker
```

**If Messages Are Stuck:**
```bash
# Purge queue (CAUTION: deletes all messages)
aws sqs purge-queue --queue-url $QUEUE_URL

# Or move messages to DLQ for investigation
aws sqs receive-message \
  --queue-url $QUEUE_URL \
  --max-number-of-messages 10 \
  --wait-time-seconds 0

# Process DLQ messages manually
aws sqs receive-message --queue-url $DLQ_URL
```

**If Workers Not Polling:**
```bash
# Check worker config
kubectl get configmap gorax-config -o json | jq '.data | {
  QUEUE_ENABLED,
  QUEUE_URL,
  WORKER_CONCURRENCY
}'

# Restart workers
kubectl rollout restart deployment/gorax-worker
```

**Prevention:**
- Monitor queue age metric
- Set appropriate visibility timeout
- Implement dead-letter queue
- Add queue depth alerts

---

### DLQ (Dead Letter Queue) Analysis

**Diagnosis:**

1. Check DLQ messages:
```bash
# Count messages in DLQ
aws sqs get-queue-attributes \
  --queue-url $DLQ_URL \
  --attribute-names ApproximateNumberOfMessages

# Receive messages from DLQ
aws sqs receive-message \
  --queue-url $DLQ_URL \
  --max-number-of-messages 10 \
  --attribute-names All \
  --message-attribute-names All
```

2. Analyze failure reasons:
```bash
# Parse messages
aws sqs receive-message --queue-url $DLQ_URL | \
  jq -r '.Messages[].Body' | \
  jq -s 'group_by(.error_type) | map({error: .[0].error_type, count: length})'
```

**Solution:**

**Replay Messages:**
```bash
# Move message back to main queue
MSG=$(aws sqs receive-message --queue-url $DLQ_URL --max-number-of-messages 1)
BODY=$(echo $MSG | jq -r '.Messages[0].Body')
RECEIPT=$(echo $MSG | jq -r '.Messages[0].ReceiptHandle')

# Send to main queue
aws sqs send-message --queue-url $QUEUE_URL --message-body "$BODY"

# Delete from DLQ
aws sqs delete-message --queue-url $DLQ_URL --receipt-handle "$RECEIPT"
```

**Bulk Replay:**
```bash
# Use replay API endpoint
curl -X POST http://localhost:8080/api/v1/webhooks/{webhookID}/events/replay \
  -H "X-Tenant-ID: tenant-123" \
  -d '{
    "event_ids": ["event-1", "event-2"],
    "filter": {"status": "failed"}
  }'
```

**Prevention:**
- Monitor DLQ message count
- Review DLQ messages regularly
- Fix underlying issues causing failures
- Set appropriate retry limits

---

### Worker Crashes

**Symptoms:**
- Worker pods restarting frequently
- CrashLoopBackOff status
- Exit code 1 or 2
- Panic in logs

**Diagnosis:**

1. Check pod status:
```bash
# Pod status
kubectl get pods -l app=gorax-worker

# Restart count
kubectl get pods -l app=gorax-worker -o json | \
  jq -r '.items[] | "\(.metadata.name) \(.status.containerStatuses[0].restartCount)"'

# Recent events
kubectl get events --sort-by='.lastTimestamp' | grep worker
```

2. Check crash logs:
```bash
# Current logs
kubectl logs deployment/gorax-worker --tail=100

# Previous crash logs
kubectl logs deployment/gorax-worker --previous

# All restarts
kubectl logs deployment/gorax-worker --all-containers=true
```

3. Look for panics:
```bash
# Find panic stack traces
kubectl logs deployment/gorax-worker | grep -A 20 "panic:"
```

**Solution:**

**For Panic Errors:**
```bash
# Review panic stack trace
kubectl logs deployment/gorax-worker --previous | grep -A 30 "panic:"

# Common panics:
# - nil pointer dereference
# - index out of range
# - type assertion failed
# - close of closed channel
```

**For Resource Limits:**
```bash
# Check OOMKilled
kubectl describe pod <pod-name> | grep -i oom

# Increase memory limit
kubectl patch deployment gorax-worker -p \
  '{"spec":{"template":{"spec":{"containers":[{"name":"worker","resources":{"limits":{"memory":"2Gi"}}}]}}}}'
```

**For Database Connection Issues:**
```bash
# Test database connectivity
kubectl exec deployment/gorax-worker -- psql -h postgres -U postgres -d gorax -c "SELECT 1;"

# Check connection string
kubectl get secret gorax-secrets -o json | \
  jq -r '.data.DATABASE_URL' | base64 -d
```

**Prevention:**
- Add comprehensive error handling
- Use recover() for panic recovery
- Set appropriate resource limits
- Implement health checks
- Add readiness/liveness probes

---

### Concurrent Execution Limits

**Symptoms:**
- "tenant at capacity" errors
- Messages being requeued
- Workflows not starting
- Uneven tenant load distribution

**Diagnosis:**

1. Check tenant concurrency:
```bash
# Active executions per tenant
redis-cli -h localhost HGETALL tenant:active_executions

# Tenant quota limits
psql -h localhost -U postgres -d gorax -c \
  "SELECT tenant_id, concurrent_executions_limit
   FROM tenants
   WHERE concurrent_executions_limit IS NOT NULL;"
```

2. Check worker capacity:
```bash
# Worker concurrency config
kubectl get configmap gorax-config -o json | jq '.data.WORKER_CONCURRENCY'

# Active workers
kubectl get pods -l app=gorax-worker | grep Running | wc -l
```

**Solution:**

**Increase Tenant Limit:**
```bash
# Update tenant quota
curl -X PUT http://localhost:8080/api/v1/admin/tenants/{tenantID}/quotas \
  -H "X-User-ID: admin" \
  -d '{
    "concurrent_executions_limit": 50
  }'
```

**Scale Workers:**
```bash
# Increase worker replicas
kubectl scale deployment/gorax-worker --replicas=10

# Enable autoscaling
kubectl autoscale deployment/gorax-worker \
  --cpu-percent=70 \
  --min=5 \
  --max=20
```

**Adjust Worker Concurrency:**
```bash
# Increase per-worker concurrency
kubectl edit configmap gorax-config
# WORKER_CONCURRENCY: "20"

kubectl rollout restart deployment/gorax-worker
```

**Prevention:**
- Monitor tenant execution distribution
- Set appropriate tenant limits
- Implement queue prioritization
- Use fair scheduling algorithm

---

### Queue Depth Monitoring

**Set Up Monitoring:**

1. **Prometheus Alerts:**
```yaml
# alerts.yaml
- alert: HighQueueDepth
  expr: gorax_queue_depth > 1000
  for: 10m
  labels:
    severity: warning
  annotations:
    summary: Queue depth is high
    description: Queue {{ $labels.queue }} has {{ $value }} messages

- alert: QueueDepthGrowing
  expr: deriv(gorax_queue_depth[10m]) > 10
  for: 5m
  labels:
    severity: warning
  annotations:
    summary: Queue depth is growing rapidly
    description: Queue depth increasing at {{ $value }} messages/second
```

2. **Grafana Dashboard:**
```json
{
  "title": "Queue Depth",
  "targets": [{
    "expr": "gorax_queue_depth"
  }],
  "thresholds": [
    {"value": 500, "color": "yellow"},
    {"value": 1000, "color": "red"}
  ]
}
```

3. **Query Queue Metrics:**
```bash
# Current queue depth
curl -s http://localhost:9090/api/v1/query?query=gorax_queue_depth | jq

# Queue depth over time
curl -s 'http://localhost:9090/api/v1/query_range?query=gorax_queue_depth&start=1609459200&end=1609462800&step=60' | jq
```

---

## Authentication/Authorization

### Login Failures

**Symptoms:**
- Users cannot log in
- "unauthorized" errors
- Session cookie not set
- Kratos returning errors

**Diagnosis:**

1. Check Kratos status:
```bash
# Kratos health
curl http://localhost:4433/health/ready

# Kratos admin API
curl http://localhost:4434/admin/identities
```

2. Check authentication logs:
```bash
# API logs
kubectl logs deployment/gorax-api | jq 'select(.msg | contains("auth"))'

# Kratos logs
kubectl logs deployment/kratos
```

**Solution:**

**In Development (DevAuth):**
```bash
# Verify dev auth is enabled
kubectl get configmap gorax-config -o json | jq '.data.APP_ENV'
# Should be "development"

# Make requests with dev header
curl -H "X-User-ID: user-123" \
     -H "X-Tenant-ID: tenant-123" \
     http://localhost:8080/api/v1/workflows
```

**In Production (Kratos):**
```bash
# Restart Kratos
kubectl rollout restart deployment/kratos

# Check Kratos configuration
kubectl get configmap kratos-config -o yaml

# Test login flow
curl -X POST http://localhost:4433/self-service/login/api
```

**Prevention:**
- Monitor authentication failure rate
- Implement rate limiting on auth endpoints
- Add authentication health checks
- Use session management best practices

---

### JWT Token Issues

**Symptoms:**
- "invalid token" errors
- Token expired errors
- Token verification failures
- Signature verification failed

**Diagnosis:**

1. Decode JWT token:
```bash
# Decode token (header and payload)
TOKEN="eyJhbGc..."
echo $TOKEN | cut -d'.' -f2 | base64 -d | jq

# Check expiration
echo $TOKEN | cut -d'.' -f2 | base64 -d | jq '.exp' | \
  xargs -I {} date -d @{}
```

2. Verify token signature:
```bash
# Using jwt-cli tool
jwt decode $TOKEN

# Check issuer and audience
jwt decode $TOKEN | jq '{iss, aud, exp}'
```

**Solution:**

**For Expired Tokens:**
```bash
# Refresh token
curl -X POST http://localhost:4433/self-service/login/api \
  -d '{"refresh_token": "..."}'
```

**For Invalid Signature:**
```bash
# Check JWT secret
kubectl get secret gorax-secrets -o json | \
  jq -r '.data.JWT_SECRET' | base64 -d

# Verify secret matches Kratos config
kubectl get configmap kratos-config -o yaml | grep secret
```

**Prevention:**
- Use appropriate token expiration times
- Implement token refresh flow
- Rotate JWT secrets regularly
- Validate token claims properly

---

### RBAC Permission Errors

**Symptoms:**
- "forbidden" 403 errors
- Users cannot access resources
- Permission denied errors
- Role checks failing

**Diagnosis:**

1. Check user roles:
```bash
# Get user roles
psql -h localhost -U postgres -d gorax -c \
  "SELECT user_id, role, tenant_id
   FROM user_roles
   WHERE user_id = 'user-123';"

# Check role permissions
psql -h localhost -U postgres -d gorax -c \
  "SELECT * FROM role_permissions WHERE role = 'developer';"
```

2. Check RBAC logs:
```bash
# Permission denied logs
kubectl logs deployment/gorax-api | \
  jq 'select(.msg | contains("permission denied"))'
```

**Solution:**

**Grant Required Permissions:**
```bash
# Assign role to user
curl -X POST http://localhost:8080/api/v1/admin/tenants/{tenantID}/users/{userID}/roles \
  -H "X-User-ID: admin" \
  -d '{"role": "developer"}'

# Or update in database
psql -h localhost -U postgres -d gorax -c \
  "INSERT INTO user_roles (user_id, tenant_id, role)
   VALUES ('user-123', 'tenant-123', 'developer');"
```

**Check Role Hierarchy:**
```
admin > developer > viewer
- admin: full access
- developer: create/edit workflows
- viewer: read-only access
```

**Prevention:**
- Document role permissions clearly
- Use principle of least privilege
- Audit role assignments regularly
- Implement role-based UI hiding

---

### Tenant Isolation Violations

**Symptoms:**
- Users seeing other tenant's data
- Cross-tenant data leaks
- Tenant ID not being enforced
- Security audit failures

**Diagnosis:**

1. Check tenant isolation:
```bash
# Test cross-tenant access
curl -H "X-Tenant-ID: tenant-A" \
  http://localhost:8080/api/v1/workflows/{tenant-B-workflow-id}
# Should return 404 or 403

# Check database queries include tenant_id
kubectl logs deployment/gorax-api | \
  grep -i "SELECT.*workflows" | \
  grep -v "tenant_id"
```

2. Audit tenant ID usage:
```bash
# Find queries without tenant filter
grep -r "SELECT.*FROM.*workflows" internal/ | \
  grep -v "tenant_id"
```

**Solution:**

**Fix Missing Tenant Filter:**
```go
// BAD: No tenant filter
workflows, err := repo.GetAll(ctx)

// GOOD: Always filter by tenant
workflows, err := repo.GetAllByTenant(ctx, tenantID)
```

**Add Database Row-Level Security:**
```sql
-- Enable RLS
ALTER TABLE workflows ENABLE ROW LEVEL SECURITY;

-- Create policy
CREATE POLICY tenant_isolation_policy ON workflows
  USING (tenant_id = current_setting('app.current_tenant')::TEXT);

-- Set tenant ID in session
SET app.current_tenant = 'tenant-123';
```

**Prevention:**
- Always include tenant_id in WHERE clauses
- Use database row-level security
- Add integration tests for tenant isolation
- Review all queries for tenant filtering
- Use ORM with tenant scoping

---

### Session Management Problems

**Symptoms:**
- Users logged out unexpectedly
- Sessions not persisting
- "session expired" errors
- Multiple logins required

**Diagnosis:**

1. Check session storage:
```bash
# Redis session keys
redis-cli -h localhost KEYS "session:*"

# Session TTL
redis-cli -h localhost TTL session:abc123

# Session data
redis-cli -h localhost GET session:abc123
```

2. Check session configuration:
```bash
# Session settings
kubectl get configmap gorax-config -o json | jq '.data | {
  SESSION_TIMEOUT,
  SESSION_SECURE,
  SESSION_SAME_SITE
}'
```

**Solution:**

**Extend Session Timeout:**
```bash
# Update configuration
kubectl edit configmap gorax-config
# SESSION_TIMEOUT: "3600"  # 1 hour

# Restart API
kubectl rollout restart deployment/gorax-api
```

**Fix Session Cookie:**
```bash
# Check cookie settings in response
curl -v http://localhost:8080/api/v1/login

# Should have:
# Set-Cookie: session=...; HttpOnly; Secure; SameSite=Strict
```

**Prevention:**
- Set appropriate session timeout
- Use HttpOnly and Secure flags
- Implement sliding session expiration
- Store sessions in Redis (not in-memory)
- Monitor session count

---

## Integration Failures

### LLM API Errors (Rate Limits, Timeouts)

**Symptoms:**
- "rate limit exceeded" errors
- API timeout errors
- 429 Too Many Requests
- AI features not working

**Diagnosis:**

1. Check LLM integration logs:
```bash
# LLM errors
kubectl logs deployment/gorax-api | \
  jq 'select(.error | contains("rate limit"))'

# Request counts
kubectl logs deployment/gorax-api | \
  jq 'select(.msg | contains("llm")) | .provider' | \
  sort | uniq -c
```

2. Check quota usage:
```bash
# OpenAI usage
curl https://api.openai.com/v1/usage \
  -H "Authorization: Bearer $OPENAI_API_KEY"

# Anthropic usage
curl https://api.anthropic.com/v1/usage \
  -H "x-api-key: $ANTHROPIC_API_KEY"
```

**Solution:**

**Implement Rate Limiting:**
```go
// Add rate limiter
limiter := rate.NewLimiter(rate.Limit(10), 1)  // 10 req/min

// Before API call
if err := limiter.Wait(ctx); err != nil {
    return err
}
```

**Add Retry Logic:**
```go
// Exponential backoff
backoff := backoff.NewExponentialBackOff()
operation := func() error {
    return llmClient.Complete(ctx, prompt)
}
err := backoff.Retry(operation, backoff)
```

**Use Multiple Providers:**
```bash
# Configure fallback providers
kubectl edit configmap gorax-config

# AI_BUILDER_PROVIDER: "openai"
# AI_BUILDER_FALLBACK_PROVIDER: "anthropic"
```

**Prevention:**
- Monitor LLM API usage
- Implement circuit breaker
- Add request queuing
- Cache LLM responses
- Set rate limits per tenant

---

### External Service Connectivity

**Symptoms:**
- "connection refused" errors
- DNS resolution failures
- Timeout errors
- Certificate validation errors

**Diagnosis:**

1. Test connectivity:
```bash
# DNS resolution
kubectl exec deployment/gorax-api -- nslookup api.external-service.com

# HTTP connectivity
kubectl exec deployment/gorax-api -- \
  curl -v https://api.external-service.com/health

# Check network policies
kubectl get networkpolicies
```

2. Check firewall rules:
```bash
# AWS security groups
aws ec2 describe-security-groups --group-ids sg-xxx

# Check outbound rules
```

**Solution:**

**DNS Issues:**
```bash
# Update DNS servers
kubectl edit configmap coredns -n kube-system

# Add custom DNS entries
kubectl apply -f - <<EOF
apiVersion: v1
kind: ConfigMap
metadata:
  name: coredns-custom
  namespace: kube-system
data:
  example.server: |
    api.external-service.com:53 {
      forward . 8.8.8.8
    }
EOF
```

**Network Policies:**
```yaml
# Allow egress to external service
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: allow-external-api
spec:
  podSelector:
    matchLabels:
      app: gorax-api
  egress:
  - to:
    - podSelector: {}
    ports:
    - protocol: TCP
      port: 443
```

**Certificate Issues:**
```bash
# Update CA certificates
kubectl create configmap ca-certs \
  --from-file=/etc/ssl/certs/ca-bundle.crt

# Mount in deployment
```

**Prevention:**
- Monitor external service health
- Implement circuit breakers
- Add timeout configurations
- Use DNS caching
- Test connectivity in staging

---

### Credential Expiration

**Symptoms:**
- 401 Unauthorized from external APIs
- OAuth token expired
- Refresh token invalid
- Credential rotation needed

**Diagnosis:**

1. Check credential expiration:
```bash
# List credentials with expiration
psql -h localhost -U postgres -d gorax -c \
  "SELECT id, name, expires_at,
          expires_at < NOW() as is_expired
   FROM credentials
   WHERE expires_at IS NOT NULL
   ORDER BY expires_at;"
```

2. Check credential usage:
```bash
# Recent credential access failures
psql -h localhost -U postgres -d gorax -c \
  "SELECT credential_id, action, success, error, created_at
   FROM credential_access_log
   WHERE success = false
     AND created_at > NOW() - INTERVAL '1 hour'
   ORDER BY created_at DESC;"
```

**Solution:**

**Rotate Expired Credentials:**
```bash
# Rotate credential
curl -X POST http://localhost:8080/api/v1/credentials/{credentialID}/rotate \
  -H "X-Tenant-ID: tenant-123" \
  -d '{"new_value": "new-api-key"}'
```

**Implement Auto-Refresh:**
```go
// OAuth token refresh
if token.ExpiresAt.Before(time.Now()) {
    newToken, err := oauthClient.RefreshToken(token.RefreshToken)
    if err != nil {
        return err
    }
    // Update credential
    credentialService.Update(ctx, credID, newToken)
}
```

**Prevention:**
- Monitor credential expiration dates
- Send alerts before expiration (7 days, 1 day)
- Implement automatic token refresh
- Add credential rotation policies
- Log credential access attempts

---

### OAuth Token Refresh Failures

**Symptoms:**
- OAuth APIs returning 401
- Refresh token invalid/expired
- User re-authentication required
- OAuth flows broken

**Diagnosis:**

1. Check OAuth tokens:
```bash
# Get token details
psql -h localhost -U postgres -d gorax -c \
  "SELECT id, provider, expires_at, has_refresh_token
   FROM oauth_tokens
   WHERE user_id = 'user-123';"
```

2. Test token refresh:
```bash
# Attempt refresh
curl -X POST https://oauth.provider.com/token \
  -d "grant_type=refresh_token" \
  -d "refresh_token=$REFRESH_TOKEN" \
  -d "client_id=$CLIENT_ID" \
  -d "client_secret=$CLIENT_SECRET"
```

**Solution:**

**Refresh Token:**
```go
// Implement refresh logic
func refreshOAuthToken(ctx context.Context, token *OAuthToken) error {
    config := oauth2.Config{
        ClientID:     clientID,
        ClientSecret: clientSecret,
        Endpoint:     provider.Endpoint,
    }

    tokenSource := config.TokenSource(ctx, &oauth2.Token{
        RefreshToken: token.RefreshToken,
    })

    newToken, err := tokenSource.Token()
    if err != nil {
        return fmt.Errorf("refresh failed: %w", err)
    }

    // Save new token
    return tokenRepo.Update(ctx, token.ID, newToken)
}
```

**Re-authenticate User:**
```bash
# Initiate OAuth flow
curl http://localhost:8080/api/v1/oauth/{provider}/authorize

# Complete flow
curl http://localhost:8080/api/v1/oauth/{provider}/callback?code=xxx
```

**Prevention:**
- Proactively refresh tokens before expiration
- Store refresh tokens securely
- Handle refresh failures gracefully
- Prompt user for re-authentication
- Monitor OAuth token health

---

### Webhook Delivery Failures

**Symptoms:**
- Webhooks not received by external systems
- Delivery retry exhausted
- Signature verification failures
- External endpoint timeouts

**Diagnosis:**

1. Check webhook delivery status:
```bash
# Recent webhook deliveries
psql -h localhost -U postgres -d gorax -c \
  "SELECT id, webhook_id, status, attempts, last_error
   FROM webhook_deliveries
   WHERE created_at > NOW() - INTERVAL '1 hour'
   ORDER BY created_at DESC
   LIMIT 20;"

# Delivery success rate
psql -h localhost -U postgres -d gorax -c \
  "SELECT
     webhook_id,
     COUNT(*) as total,
     COUNT(*) FILTER (WHERE status = 'delivered') as delivered,
     COUNT(*) FILTER (WHERE status = 'failed') as failed
   FROM webhook_deliveries
   WHERE created_at > NOW() - INTERVAL '1 day'
   GROUP BY webhook_id;"
```

2. Check webhook endpoint:
```bash
# Test endpoint connectivity
curl -v -X POST https://customer-webhook.com/endpoint \
  -H "Content-Type: application/json" \
  -d '{"test": true}'
```

**Solution:**

**Retry Failed Deliveries:**
```bash
# Replay webhook events
curl -X POST http://localhost:8080/api/v1/webhooks/{webhookID}/events/replay \
  -H "X-Tenant-ID: tenant-123" \
  -d '{
    "event_ids": ["event-1", "event-2"],
    "filter": {"status": "failed"}
  }'
```

**Update Webhook Configuration:**
```bash
# Update webhook endpoint
curl -X PUT http://localhost:8080/api/v1/webhooks/{webhookID} \
  -H "X-Tenant-ID: tenant-123" \
  -d '{
    "url": "https://new-endpoint.com/webhook",
    "timeout_seconds": 30,
    "max_retries": 3
  }'
```

**Regenerate Secret:**
```bash
# Regenerate webhook secret
curl -X POST http://localhost:8080/api/v1/webhooks/{webhookID}/regenerate-secret \
  -H "X-Tenant-ID: tenant-123"
```

**Prevention:**
- Monitor webhook delivery success rate
- Implement exponential backoff retry
- Add webhook health checks
- Provide webhook testing tool
- Log delivery attempts

---

## Monitoring and Alerts

### Setting up Prometheus Alerts

**Alert Configuration File:**

```yaml
# /etc/prometheus/alerts.yaml
groups:
  - name: gorax_critical
    interval: 30s
    rules:
      - alert: APIDown
        expr: up{job="gorax-api"} == 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "Gorax API is down"
          description: "API has been down for more than 1 minute"

      - alert: HighErrorRate
        expr: rate(gorax_http_requests_total{status=~"5.."}[5m]) > 0.05
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "High HTTP error rate"
          description: "Error rate is {{ $value | humanize }} per second"

      - alert: DatabaseConnectionsHigh
        expr: |
          sum(pg_stat_activity_count) /
          sum(pg_settings_max_connections) > 0.8
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Database connections high"
          description: "Using {{ $value | humanizePercentage }} of connections"

  - name: gorax_performance
    interval: 1m
    rules:
      - alert: SlowAPIResponses
        expr: |
          histogram_quantile(0.95,
            rate(gorax_http_request_duration_seconds_bucket[5m])
          ) > 2
        for: 10m
        labels:
          severity: warning
        annotations:
          summary: "API responses are slow"
          description: "P95 latency is {{ $value }}s"

      - alert: QueueBacklog
        expr: gorax_queue_depth > 1000
        for: 10m
        labels:
          severity: warning
        annotations:
          summary: "Queue backlog growing"
          description: "Queue depth is {{ $value }}"

      - alert: HighMemoryUsage
        expr: |
          process_resident_memory_bytes /
          container_spec_memory_limit_bytes > 0.8
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High memory usage"
          description: "Using {{ $value | humanizePercentage }} of memory"
```

**Apply Alert Rules:**
```bash
# Update Prometheus ConfigMap
kubectl create configmap prometheus-alerts \
  --from-file=alerts.yaml \
  --dry-run=client -o yaml | kubectl apply -f -

# Reload Prometheus
kubectl exec prometheus-0 -- kill -HUP 1

# Verify rules loaded
curl http://localhost:9090/api/v1/rules
```

---

### Grafana Dashboard Setup

**Import Pre-built Dashboards:**

```bash
# Import via Grafana UI
# Navigate to: http://localhost:3000
# Dashboard > Import > Upload JSON

# Or via API
curl -X POST http://admin:admin@localhost:3000/api/dashboards/db \
  -H "Content-Type: application/json" \
  -d @dashboards/gorax-overview.json
```

**Create Custom Dashboard:**

```json
{
  "dashboard": {
    "title": "Gorax Operations",
    "panels": [
      {
        "title": "Request Rate",
        "targets": [{
          "expr": "rate(gorax_http_requests_total[5m])"
        }],
        "type": "graph"
      },
      {
        "title": "Error Rate",
        "targets": [{
          "expr": "rate(gorax_http_requests_total{status=~\"5..\"}[5m])"
        }],
        "type": "graph"
      },
      {
        "title": "Queue Depth",
        "targets": [{
          "expr": "gorax_queue_depth"
        }],
        "type": "gauge",
        "thresholds": [
          {"value": 500, "color": "yellow"},
          {"value": 1000, "color": "red"}
        ]
      }
    ]
  }
}
```

---

### Log Aggregation (Loki)

**Deploy Loki:**

```yaml
# loki-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: loki
spec:
  replicas: 1
  selector:
    matchLabels:
      app: loki
  template:
    metadata:
      labels:
        app: loki
    spec:
      containers:
      - name: loki
        image: grafana/loki:latest
        ports:
        - containerPort: 3100
        volumeMounts:
        - name: config
          mountPath: /etc/loki
      volumes:
      - name: config
        configMap:
          name: loki-config
```

**Configure Promtail:**

```yaml
# promtail-daemonset.yaml
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: promtail
spec:
  selector:
    matchLabels:
      app: promtail
  template:
    metadata:
      labels:
        app: promtail
    spec:
      containers:
      - name: promtail
        image: grafana/promtail:latest
        args:
        - -config.file=/etc/promtail/promtail.yaml
        volumeMounts:
        - name: config
          mountPath: /etc/promtail
        - name: varlog
          mountPath: /var/log
        - name: varlibdockercontainers
          mountPath: /var/lib/docker/containers
          readOnly: true
      volumes:
      - name: config
        configMap:
          name: promtail-config
      - name: varlog
        hostPath:
          path: /var/log
      - name: varlibdockercontainers
        hostPath:
          path: /var/lib/docker/containers
```

**Query Logs in Grafana:**

```
# LogQL queries
{app="gorax-api"} |= "error"
{app="gorax-api"} | json | level="ERROR"
{app="gorax-api"} | json | tenant_id="tenant-123"
```

---

### Error Tracking (Sentry)

**Configure Sentry:**

```bash
# Set Sentry DSN
kubectl create secret generic gorax-secrets \
  --from-literal=SENTRY_DSN="https://xxx@sentry.io/xxx"

# Enable in ConfigMap
kubectl edit configmap gorax-config
# SENTRY_ENABLED: "true"
# SENTRY_ENVIRONMENT: "production"
# SENTRY_SAMPLE_RATE: "1.0"
```

**View Errors in Sentry:**

```
Navigate to: https://sentry.io/organizations/your-org/issues/

Filter by:
- Environment: production
- Release: v1.2.3
- User: user-123
- Tags: tenant_id, workflow_id
```

---

### Distributed Tracing (Jaeger)

**Deploy Jaeger:**

```bash
# Using Jaeger Operator
kubectl create namespace observability
kubectl create -f https://github.com/jaegertracing/jaeger-operator/releases/download/v1.50.0/jaeger-operator.yaml -n observability

# Create Jaeger instance
kubectl apply -f - <<EOF
apiVersion: jaegertracing.io/v1
kind: Jaeger
metadata:
  name: gorax-jaeger
spec:
  strategy: production
  storage:
    type: elasticsearch
    options:
      es:
        server-urls: http://elasticsearch:9200
EOF
```

**Configure Tracing:**

```bash
# Enable tracing
kubectl edit configmap gorax-config
# TRACING_ENABLED: "true"
# TRACING_ENDPOINT: "jaeger-collector:4317"
# TRACING_SAMPLE_RATE: "0.1"  # Sample 10% of traces

# Restart pods
kubectl rollout restart deployment/gorax-api
kubectl rollout restart deployment/gorax-worker
```

**View Traces:**

```
Navigate to: http://localhost:16686

Search by:
- Service: gorax-api
- Operation: POST /api/v1/workflows/execute
- Tags: tenant_id=tenant-123
- Duration: >1s

Analyze:
- Trace waterfall
- Service dependencies
- Slow spans
```

---

## Logs and Debugging

### Log Format and Structure

**Structured Log Format:**

```json
{
  "time": "2026-01-01T10:30:00.000Z",
  "level": "INFO",
  "msg": "workflow execution completed",
  "trace_id": "abc123def456",
  "request_id": "req-789",
  "tenant_id": "tenant-123",
  "user_id": "user-456",
  "workflow_id": "wf-789",
  "execution_id": "exec-abc",
  "duration_ms": 1250,
  "status": "completed",
  "component": "executor"
}
```

**Log Fields:**

| Field | Description | Always Present |
|-------|-------------|----------------|
| `time` | ISO 8601 timestamp | Yes |
| `level` | DEBUG, INFO, WARN, ERROR | Yes |
| `msg` | Human-readable message | Yes |
| `trace_id` | Distributed trace ID | If tracing enabled |
| `request_id` | HTTP request ID | For API requests |
| `tenant_id` | Tenant identifier | For tenant operations |
| `user_id` | User identifier | For user operations |
| `workflow_id` | Workflow identifier | For workflow operations |
| `execution_id` | Execution identifier | For executions |
| `error` | Error message | For error logs |
| `stack` | Stack trace | For errors/panics |

---

### Structured Logging Fields

**Using Structured Fields:**

```bash
# Filter by level
jq 'select(.level=="ERROR")' < api.log

# Filter by tenant
jq 'select(.tenant_id=="tenant-123")' < api.log

# Filter by time range (last hour)
jq 'select(.time > (now - 3600 | todate))' < api.log

# Count errors by type
jq -r 'select(.level=="ERROR") | .msg' < api.log | sort | uniq -c | sort -rn

# Extract execution times
jq -r 'select(.duration_ms) | "\(.workflow_id) \(.duration_ms)"' < api.log | \
  awk '{sum[$1]+=$2; count[$1]++} END {for(w in sum) print w, sum[w]/count[w]}'
```

---

### Log Levels (DEBUG, INFO, WARN, ERROR)

**Log Level Guidelines:**

**ERROR:**
- System errors requiring immediate attention
- Failed operations that affect users
- Integration failures
- Database errors

```json
{
  "level": "ERROR",
  "msg": "failed to execute workflow",
  "error": "database connection failed",
  "tenant_id": "tenant-123",
  "workflow_id": "wf-456"
}
```

**WARN:**
- Degraded performance
- Retry attempts
- Configuration issues
- Deprecated API usage

```json
{
  "level": "WARN",
  "msg": "external API slow response",
  "duration_ms": 5000,
  "api": "https://api.example.com"
}
```

**INFO:**
- Successful operations
- State changes
- Business events
- Request/response logs

```json
{
  "level": "INFO",
  "msg": "workflow execution completed",
  "execution_id": "exec-123",
  "status": "completed"
}
```

**DEBUG:**
- Detailed diagnostic information
- Variable values
- Control flow
- Development debugging

```json
{
  "level": "DEBUG",
  "msg": "processing workflow step",
  "step_id": "step-1",
  "step_type": "http_request",
  "config": {...}
}
```

**Change Log Level:**

```bash
# Set log level via environment
kubectl set env deployment/gorax-api LOG_LEVEL=DEBUG

# Or update ConfigMap
kubectl edit configmap gorax-config
# LOG_LEVEL: "DEBUG"

kubectl rollout restart deployment/gorax-api
```

---

### Filtering Logs by Tenant/Workflow/Execution

**Command-Line Filtering:**

```bash
# By tenant
kubectl logs deployment/gorax-api | jq 'select(.tenant_id=="tenant-123")'

# By workflow
kubectl logs deployment/gorax-api | jq 'select(.workflow_id=="wf-456")'

# By execution
kubectl logs deployment/gorax-api | jq 'select(.execution_id=="exec-789")'

# Multiple filters
kubectl logs deployment/gorax-api | \
  jq 'select(.tenant_id=="tenant-123" and .level=="ERROR")'

# Time range + filters
kubectl logs --since=1h deployment/gorax-api | \
  jq 'select(.tenant_id=="tenant-123" and .workflow_id=="wf-456")'
```

**In Grafana/Loki:**

```
# LogQL queries
{app="gorax-api"} | json | tenant_id="tenant-123"
{app="gorax-api"} | json | workflow_id="wf-456" | level="ERROR"
{app="gorax-api"} | json | execution_id="exec-789"
```

---

### Correlation IDs for Tracing

**Trace ID Propagation:**

All logs include `trace_id` for correlation:

```bash
# Find all logs for a trace
TRACE_ID="abc123def456"
kubectl logs deployment/gorax-api | jq "select(.trace_id==\"$TRACE_ID\")"

# Trace request flow
kubectl logs deployment/gorax-api | \
  jq -r "select(.trace_id==\"$TRACE_ID\") | \"\(.time) \(.component) \(.msg)\"" | \
  sort
```

**Request ID Usage:**

```bash
# HTTP request tracking
REQUEST_ID="req-789"
kubectl logs deployment/gorax-api | jq "select(.request_id==\"$REQUEST_ID\")"

# Full request lifecycle
kubectl logs deployment/gorax-api | \
  jq -r "select(.request_id==\"$REQUEST_ID\") | \"\(.time) \(.msg)\"" | \
  sort
```

**In Jaeger UI:**

```
Search by trace_id: abc123def456
View complete trace waterfall
Navigate between spans
```

---

## Common kubectl/Docker Commands

### Viewing Pod Logs

```bash
# Current logs
kubectl logs deployment/gorax-api

# Follow logs (tail -f)
kubectl logs -f deployment/gorax-api

# Last 100 lines
kubectl logs deployment/gorax-api --tail=100

# Since time
kubectl logs deployment/gorax-api --since=1h
kubectl logs deployment/gorax-api --since=2023-01-01T10:00:00Z

# Previous container (after crash)
kubectl logs deployment/gorax-api --previous

# Multiple pods
kubectl logs -l app=gorax-api --all-containers=true

# Specific container in pod
kubectl logs pod-name -c container-name

# Save to file
kubectl logs deployment/gorax-api > api-logs.txt
```

---

### Exec into Containers

```bash
# Bash shell
kubectl exec -it deployment/gorax-api -- /bin/bash

# Run command
kubectl exec deployment/gorax-api -- ps aux

# Interactive psql
kubectl exec -it deployment/postgres -- psql -U postgres -d gorax

# Interactive redis-cli
kubectl exec -it deployment/redis -- redis-cli

# Run debug commands
kubectl exec deployment/gorax-api -- curl http://localhost:8080/health

# Copy files
kubectl cp deployment/gorax-api:/tmp/debug.log ./debug.log
kubectl cp ./config.yaml deployment/gorax-api:/tmp/config.yaml
```

---

### Port Forwarding

```bash
# Forward local port to pod
kubectl port-forward deployment/gorax-api 8080:8080

# Forward to specific pod
kubectl port-forward pod/gorax-api-abc123 8080:8080

# Multiple ports
kubectl port-forward deployment/gorax-api 8080:8080 9090:9090

# Background
kubectl port-forward deployment/gorax-api 8080:8080 &

# Access forwarded port
curl http://localhost:8080/health
```

---

### Resource Usage

```bash
# Pod resource usage
kubectl top pods

# Specific namespace
kubectl top pods -n production

# Node resource usage
kubectl top nodes

# Detailed pod metrics
kubectl describe pod gorax-api-abc123 | grep -A 5 "Limits:"

# All resources
kubectl get all
```

---

### Restart Strategies

```bash
# Rolling restart (zero downtime)
kubectl rollout restart deployment/gorax-api

# Rollout status
kubectl rollout status deployment/gorax-api

# Rollout history
kubectl rollout history deployment/gorax-api

# Rollback
kubectl rollout undo deployment/gorax-api

# Rollback to specific revision
kubectl rollout undo deployment/gorax-api --to-revision=3

# Delete pod (recreated automatically)
kubectl delete pod gorax-api-abc123

# Force recreate all pods
kubectl delete pods -l app=gorax-api
```

---

### Scaling Deployments

```bash
# Manual scaling
kubectl scale deployment/gorax-api --replicas=5

# Check replica status
kubectl get deployment gorax-api

# Autoscaling
kubectl autoscale deployment/gorax-api \
  --cpu-percent=70 \
  --min=3 \
  --max=10

# Check HPA status
kubectl get hpa

# Describe HPA
kubectl describe hpa gorax-api
```

---

## Emergency Procedures

### Service Degradation Response

**Incident Response Checklist:**

1. **Acknowledge Alert** (0-2 minutes)
   - Acknowledge in PagerDuty/OpsGenie
   - Check #incidents Slack channel
   - Assign incident commander

2. **Assess Impact** (2-5 minutes)
   ```bash
   # Check service health
   curl http://localhost:8080/health
   curl http://localhost:8080/ready

   # Check error rate
   curl -s 'http://localhost:9090/api/v1/query?query=rate(gorax_http_requests_total{status=~"5.."}[5m])' | jq

   # Check active pods
   kubectl get pods -l app=gorax-api
   ```

3. **Communicate** (5 minutes)
   - Post incident status to #incidents
   - Update status page
   - Notify stakeholders if user-facing

4. **Mitigate** (5-30 minutes)
   ```bash
   # Quick fixes:
   # - Scale up replicas
   kubectl scale deployment/gorax-api --replicas=10

   # - Restart unhealthy pods
   kubectl delete pods -l app=gorax-api --field-selector status.phase!=Running

   # - Drain problematic node
   kubectl drain node-xyz --ignore-daemonsets

   # - Rollback bad deployment
   kubectl rollout undo deployment/gorax-api
   ```

5. **Monitor** (Ongoing)
   - Watch error rate trend
   - Monitor user reports
   - Check dependent services

6. **Resolve** (Variable)
   - Apply permanent fix
   - Verify resolution
   - Update status page

7. **Post-Incident** (Within 48 hours)
   - Write incident report
   - Schedule post-mortem
   - Create action items

---

### Database Failover

**Automated Failover (Patroni/Stolon):**

```bash
# Check cluster status
kubectl exec patroni-0 -- patronictl list

# Current master
kubectl exec patroni-0 -- patronictl list | grep Leader

# Trigger failover
kubectl exec patroni-0 -- patronictl failover --candidate patroni-1
```

**Manual Failover:**

1. **Promote Replica:**
   ```bash
   # On replica
   kubectl exec postgres-replica-0 -- pg_ctl promote -D /var/lib/postgresql/data

   # Verify
   kubectl exec postgres-replica-0 -- psql -U postgres -c "SELECT pg_is_in_recovery();"
   # Should return: false
   ```

2. **Update Connection Strings:**
   ```bash
   # Update ConfigMap
   kubectl edit configmap gorax-config
   # DB_HOST: "postgres-replica-0"

   # Restart applications
   kubectl rollout restart deployment/gorax-api
   kubectl rollout restart deployment/gorax-worker
   ```

3. **Verify Connectivity:**
   ```bash
   # Test connection
   kubectl exec deployment/gorax-api -- \
     psql -h postgres-replica-0 -U postgres -d gorax -c "SELECT 1;"
   ```

---

### Rollback Deployments

**Fast Rollback:**

```bash
# Rollback to previous version
kubectl rollout undo deployment/gorax-api

# Check status
kubectl rollout status deployment/gorax-api

# Verify
curl http://localhost:8080/health
```

**Rollback to Specific Version:**

```bash
# View history
kubectl rollout history deployment/gorax-api

# Rollback to revision
kubectl rollout undo deployment/gorax-api --to-revision=5

# Verify version
kubectl get deployment gorax-api -o json | jq '.spec.template.spec.containers[0].image'
```

**Rollback Database Migrations:**

```bash
# Identify failed migration
psql -h localhost -U postgres -d gorax -c "SELECT * FROM schema_migrations ORDER BY version DESC LIMIT 5;"

# Run rollback script
psql -h localhost -U postgres -d gorax < migrations/rollback/020_marketplace.sql

# Verify
psql -h localhost -U postgres -d gorax -c "\d marketplace_templates"
```

---

### Circuit Breaker Activation

**Enable Circuit Breaker:**

```bash
# Update configuration
kubectl edit configmap gorax-config

# Add circuit breaker settings:
# CIRCUIT_BREAKER_ENABLED: "true"
# CIRCUIT_BREAKER_THRESHOLD: "5"  # Open after 5 failures
# CIRCUIT_BREAKER_TIMEOUT: "60s"  # Try again after 60s

# Restart
kubectl rollout restart deployment/gorax-api
```

**Check Circuit Breaker Status:**

```bash
# View metrics
curl http://localhost:9090/metrics | grep circuit_breaker

# Prometheus query
curl -s 'http://localhost:9090/api/v1/query?query=circuit_breaker_state' | jq
# States: 0=closed, 1=half_open, 2=open
```

**Manual Circuit Breaker Override:**

```go
// Temporarily disable external service
// Deploy config:
FEATURE_FLAG_EXTERNAL_SERVICE: "false"
```

---

### Rate Limiting Adjustments

**Increase Rate Limits:**

```bash
# Update rate limits
kubectl edit configmap gorax-config

# RATE_LIMIT_REQUESTS_PER_MINUTE: "1000"
# RATE_LIMIT_BURST: "100"

# Restart
kubectl rollout restart deployment/gorax-api
```

**Per-Tenant Rate Limits:**

```bash
# Update via API
curl -X PUT http://localhost:8080/api/v1/admin/tenants/{tenantID}/quotas \
  -H "X-User-ID: admin" \
  -d '{
    "rate_limit_per_minute": 500
  }'
```

**Disable Rate Limiting (Emergency):**

```bash
kubectl set env deployment/gorax-api RATE_LIMIT_ENABLED=false
```

**Monitor Rate Limiting:**

```promql
# Rate limit hits
rate(rate_limit_exceeded_total[5m])

# By tenant
rate(rate_limit_exceeded_total[5m]) by (tenant_id)
```

---

## Post-Mortem Template

**Incident Post-Mortem Template:**

```markdown
# Incident Post-Mortem: [Brief Title]

**Date:** YYYY-MM-DD
**Duration:** XX minutes
**Severity:** Critical/High/Medium/Low
**Incident Commander:** [Name]

---

## Summary

[1-2 paragraph summary of what happened]

---

## Impact

- **User Impact:** [Number of users/tenants affected]
- **Service Availability:** [Uptime percentage]
- **Data Loss:** [Any data loss or corruption]
- **Financial Impact:** [If applicable]

---

## Timeline (All times in UTC)

| Time | Event |
|------|-------|
| 10:00 | Alert fired: High error rate |
| 10:02 | On-call engineer acknowledged |
| 10:05 | Incident commander assigned |
| 10:10 | Root cause identified: Database connection pool exhausted |
| 10:15 | Mitigation applied: Increased connection pool size |
| 10:20 | Service recovered |
| 10:30 | Monitoring confirmed resolution |
| 11:00 | Incident closed |

---

## Root Cause Analysis

**What Happened:**
[Detailed explanation of the root cause]

**Why It Happened:**
[Underlying reasons/contributing factors]

**Detection:**
[How was the incident detected? Why didn't monitoring catch it earlier?]

---

## Resolution

**Immediate Actions:**
- [Action 1]
- [Action 2]

**Permanent Fix:**
- [Long-term solution]

---

## Action Items

| Action | Owner | Due Date | Status |
|--------|-------|----------|--------|
| Increase database connection pool | DevOps | 2026-01-05 | ✅ Done |
| Add connection pool monitoring | SRE | 2026-01-08 | 🔄 In Progress |
| Update runbook | On-call | 2026-01-10 | ⏳ Pending |
| Add pre-deployment validation | Dev | 2026-01-15 | ⏳ Pending |

---

## Lessons Learned

**What Went Well:**
- [Positive aspects]

**What Went Wrong:**
- [Issues during incident response]

**Where We Got Lucky:**
- [Factors that prevented worse outcome]

---

## Supporting Information

**Metrics:**
- [Relevant Prometheus queries]
- [Grafana dashboard links]

**Logs:**
- [Key log excerpts]
- [Trace IDs]

**Related Incidents:**
- [Link to similar past incidents]

---

**Reviewed By:** [Name, Date]
**Approved By:** [Name, Date]
```

---

## Additional Resources

### Documentation
- [Gorax Architecture Documentation](./architecture.md)
- [Deployment Guide](./DEPLOYMENT.md)
- [Observability Guide](./observability.md)
- [Security Guide](./WEBSOCKET_SECURITY.md)

### External Resources
- [Kubernetes Documentation](https://kubernetes.io/docs/)
- [PostgreSQL Documentation](https://www.postgresql.org/docs/)
- [Prometheus Documentation](https://prometheus.io/docs/)
- [Grafana Documentation](https://grafana.com/docs/)
- [Go Performance Profiling](https://go.dev/blog/pprof)

### On-Call Resources
- **Runbooks:** `/docs/runbooks/`
- **Incident Response:** `/docs/incident-response.md`
- **Escalation Path:** `/docs/escalation.md`
- **Team Contacts:** [Internal Wiki]

### Tools
- **Prometheus:** `http://localhost:9090`
- **Grafana:** `http://localhost:3000`
- **Jaeger:** `http://localhost:16686`
- **Sentry:** `https://sentry.io`

---

## Quick Reference Card

**Emergency Contacts:**
- On-Call: [PagerDuty/OpsGenie]
- #incidents Slack Channel
- Engineering Manager: [Contact]

**Critical Commands:**
```bash
# Health checks
curl http://localhost:8080/health
kubectl get pods -l app=gorax-api

# Scale up
kubectl scale deployment/gorax-api --replicas=10

# Restart
kubectl rollout restart deployment/gorax-api

# Rollback
kubectl rollout undo deployment/gorax-api

# View logs
kubectl logs -f deployment/gorax-api --tail=100

# Database status
psql -h localhost -U postgres -d gorax -c "SELECT version();"
```

**Key Metrics:**
- Error Rate: `rate(gorax_http_requests_total{status=~"5.."}[5m])`
- Queue Depth: `gorax_queue_depth`
- P95 Latency: `histogram_quantile(0.95, rate(gorax_http_request_duration_seconds_bucket[5m]))`

---

**Last Updated:** 2026-01-01
**Maintainer:** DevOps Team
**Version:** 1.0
