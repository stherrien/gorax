# Post-Deployment Checklist

This checklist covers the medium-priority tasks that need to be completed after the recent platform enhancements.

## Table of Contents

1. [GitHub Configuration](#1-github-configuration)
2. [Database Migration](#2-database-migration)
3. [Performance Optimizations](#3-performance-optimizations)
4. [Verification Steps](#4-verification-steps)

---

## 1. GitHub Configuration

### 1.1 Configure GitHub Environments

Navigate to: **Settings → Environments**

#### Staging Environment
1. Click "New environment"
2. Name: `staging`
3. Environment URL: `https://staging.gorax.dev` (update with your URL)
4. Protection rules: **None** (auto-deploy from dev branch)
5. Environment secrets: Add `STAGING_URL` if needed

#### Production Environment
1. Click "New environment"
2. Name: `production`
3. Environment URL: `https://gorax.dev` (update with your URL)
4. Protection rules:
   - ✅ **Required reviewers**: Add at least 1 reviewer
   - ✅ **Wait timer**: 5 minutes (optional)
   - ✅ **Deployment branches**: Only `main` branch
5. Environment secrets: Add `PRODUCTION_URL`, `HEALTH_CHECK_TOKEN`

---

### 1.2 Add Required CI/CD Secrets

Navigate to: **Settings → Secrets and variables → Actions → New repository secret**

#### Required Secrets

| Secret Name | Description | Example Value |
|-------------|-------------|---------------|
| `CODECOV_TOKEN` | Codecov upload token (optional) | `abc123def456...` |
| `STAGING_URL` | Staging environment URL | `https://staging.gorax.dev` |
| `PRODUCTION_URL` | Production environment URL | `https://gorax.dev` |
| `HEALTH_CHECK_TOKEN` | API token for health checks | `Bearer your-token-here` |

#### Infrastructure-Specific Secrets

**If using Kubernetes:**
```
KUBECONFIG              Base64-encoded kubeconfig file
K8S_CLUSTER_URL         Kubernetes cluster API URL
K8S_TOKEN               Service account token
```

**If using AWS ECS:**
```
AWS_ACCESS_KEY_ID       AWS access key
AWS_SECRET_ACCESS_KEY   AWS secret key
AWS_REGION              AWS region (e.g., us-east-1)
ECS_CLUSTER            ECS cluster name
ECS_SERVICE            ECS service name
```

**If using SSH deployment:**
```
SSH_PRIVATE_KEY         Private key for SSH access
SSH_HOST                Server hostname or IP
SSH_USER                SSH username
```

---

### 1.3 Apply Branch Protection Rules

Navigate to: **Settings → Branches → Add branch protection rule**

#### Protection Rules for `main` Branch

**Branch name pattern:** `main`

**Settings:**

1. **Require a pull request before merging**
   - ✅ Require approvals: **1**
   - ✅ Dismiss stale pull request approvals when new commits are pushed
   - ✅ Require review from Code Owners (if CODEOWNERS file exists)

2. **Require status checks to pass before merging**
   - ✅ Require branches to be up to date before merging

   **Required status checks:**
   ```
   ✅ Go Tests
   ✅ Go Lint
   ✅ Frontend Tests
   ✅ Frontend Lint
   ✅ Coverage Threshold Check
   ✅ Build Verification
   ✅ Security Scanning / gosec
   ✅ Security Scanning / npm-audit
   ✅ CodeQL Analysis / analyze-go
   ✅ CodeQL Analysis / analyze-typescript
   ✅ Secrets Scanning / gitleaks
   ```

3. **Additional settings**
   - ✅ Require linear history
   - ✅ Require signed commits (recommended)
   - ✅ Include administrators (enforces rules on admins too)

4. **Restrict pushes**
   - ✅ Restrict who can push to matching branches
   - ✅ Do not allow bypassing the above settings

5. **Rules applied to everyone**
   - ✅ Allow force pushes: **DISABLED**
   - ✅ Allow deletions: **DISABLED**

#### Protection Rules for `dev` Branch

**Branch name pattern:** `dev`

Repeat the above settings with these differences:
- Require approvals: **1**
- Linear history: **Optional** (can be disabled for dev)
- Signed commits: **Optional** (can be disabled for dev)

---

## 2. Database Migration

### 2.1 Run Migration 022 (Webhook Retry)

The webhook retry feature requires a database migration to add retry-related fields.

**Migration File:** `migrations/022_webhook_retry.sql`

#### Development Environment

```bash
# Option 1: Using psql
psql -h localhost -U postgres -d gorax < migrations/022_webhook_retry.sql

# Option 2: Using make (if configured)
make migrate-up

# Option 3: Using Docker
docker exec -i gorax-postgres psql -U postgres -d gorax < migrations/022_webhook_retry.sql
```

#### Staging Environment

```bash
# Connect to staging database
psql -h staging-db.example.com -U gorax -d gorax_staging < migrations/022_webhook_retry.sql
```

#### Production Environment

**IMPORTANT: Follow these steps carefully**

1. **Backup Database First**
   ```bash
   pg_dump -h production-db -U gorax -d gorax > backup-pre-migration-022-$(date +%Y%m%d).sql
   ```

2. **Review Migration**
   - Verify migration file contents
   - Understand what changes will be made
   - Estimate impact (this migration adds columns and indexes, should be fast)

3. **Test in Staging**
   - Ensure migration runs successfully in staging
   - Verify application works after migration

4. **Run in Production**
   ```bash
   # Connect to production database
   psql -h production-db -U gorax -d gorax < migrations/022_webhook_retry.sql
   ```

5. **Verify Migration**
   ```sql
   -- Check that new columns exist
   \d webhook_events

   -- Check that new indexes exist
   \di idx_webhook_events_next_retry
   \di idx_webhook_events_permanently_failed
   \di idx_webhook_events_retry_count

   -- Check that functions exist
   \df calculate_next_retry_time
   \df mark_webhook_event_for_retry
   \df get_webhook_events_for_retry

   -- Check that views exist
   \dv webhook_retry_stats
   \dv webhook_health
   ```

6. **Deploy Application**
   - Deploy the application code that uses the new fields
   - Monitor logs for any issues

#### Rollback Plan

If issues occur, rollback script:

```sql
-- migrations/022_webhook_retry_rollback.sql

-- Drop views
DROP VIEW IF EXISTS webhook_health;
DROP VIEW IF EXISTS webhook_retry_stats;

-- Drop functions
DROP FUNCTION IF EXISTS get_webhook_events_for_retry(INTEGER);
DROP FUNCTION IF EXISTS mark_webhook_event_for_retry(UUID, TEXT);
DROP FUNCTION IF EXISTS calculate_next_retry_time(INTEGER, INTEGER, INTEGER, NUMERIC);

-- Drop indexes
DROP INDEX IF EXISTS idx_webhook_events_retry_count;
DROP INDEX IF EXISTS idx_webhook_events_permanently_failed;
DROP INDEX IF EXISTS idx_webhook_events_next_retry;

-- Drop columns
ALTER TABLE webhook_events DROP COLUMN IF EXISTS permanently_failed;
ALTER TABLE webhook_events DROP COLUMN IF EXISTS retry_error;
ALTER TABLE webhook_events DROP COLUMN IF EXISTS last_retry_at;
ALTER TABLE webhook_events DROP COLUMN IF EXISTS next_retry_at;
ALTER TABLE webhook_events DROP COLUMN IF EXISTS max_retries;
ALTER TABLE webhook_events DROP COLUMN IF EXISTS retry_count;
```

---

## 3. Performance Optimizations

### 3.1 Database Connection Pool Settings

**File to modify:** `internal/api/app.go` or where database connection is initialized

#### Current Settings
```go
db.SetMaxOpenConns(25)  // Max concurrent connections
db.SetMaxIdleConns(5)   // Idle connection pool
// Missing: SetConnMaxLifetime and SetConnMaxIdleTime
```

#### Recommended Settings
```go
// Connection pool configuration
db.SetMaxOpenConns(150)                        // Max concurrent connections
db.SetMaxIdleConns(25)                         // Idle connection pool
db.SetConnMaxLifetime(5 * time.Minute)         // Connection lifetime
db.SetConnMaxIdleTime(10 * time.Minute)        // Idle connection timeout
```

#### Calculation

```
Max Open Connections = (Expected Concurrent Requests × Avg Queries Per Request) / 2
Example: (100 concurrent requests × 3 queries) / 2 = 150 connections
```

**Environment Variables (recommended):**
```bash
# Add to .env.example
DB_MAX_OPEN_CONNS=150         # Max concurrent database connections
DB_MAX_IDLE_CONNS=25          # Idle connection pool size
DB_CONN_MAX_LIFETIME=5m       # Connection max lifetime
DB_CONN_MAX_IDLE_TIME=10m     # Idle connection timeout
```

**Implementation:**
```go
// In internal/api/app.go setupDatabase()
db.SetMaxOpenConns(cfg.Database.MaxOpenConns)
db.SetMaxIdleConns(cfg.Database.MaxIdleConns)
db.SetConnMaxLifetime(cfg.Database.ConnMaxLifetime)
db.SetConnMaxIdleTime(cfg.Database.ConnMaxIdleTime)
```

### 3.2 Fix GetTopFailures N+1 Query

**File:** `internal/workflow/metrics.go` (or similar)

**Current Query (Inefficient):**
```sql
-- Contains correlated subquery
SELECT workflow_id, COUNT(*) as failures,
       (SELECT error_message FROM executions
        WHERE workflow_id = e.workflow_id
        AND status = 'failed'
        ORDER BY completed_at DESC LIMIT 1) as error_preview
FROM executions e
WHERE tenant_id = $1 AND status = 'failed'
GROUP BY workflow_id
ORDER BY failures DESC
```

**Optimized Query (Using LATERAL Join):**
```sql
SELECT e.workflow_id, w.name, COUNT(*) as failure_count,
       MAX(e.completed_at) as last_failed_at,
       latest.error_message as error_preview
FROM executions e
INNER JOIN workflows w ON e.workflow_id = w.id
CROSS JOIN LATERAL (
    SELECT error_message
    FROM executions
    WHERE workflow_id = e.workflow_id
      AND status = 'failed'
      AND error_message IS NOT NULL
    ORDER BY completed_at DESC
    LIMIT 1
) latest
WHERE e.tenant_id = $1
  AND e.created_at >= $2
  AND e.created_at < $3
  AND e.status = 'failed'
GROUP BY e.workflow_id, w.name, latest.error_message
ORDER BY failure_count DESC
LIMIT $4
```

**Expected Improvement:** 70-90% query time reduction

### 3.3 Add Database Indexes

**Create migration:** `migrations/023_performance_indexes.sql`

```sql
-- Add specialized indexes for common queries

-- For execution trend queries
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_executions_tenant_hour_trunc
ON executions(tenant_id, DATE_TRUNC('hour', created_at));

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_executions_tenant_day_trunc
ON executions(tenant_id, DATE_TRUNC('day', created_at));

-- Covering index for workflow list with counts
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_workflows_tenant_status_updated
ON workflows(tenant_id, status, updated_at DESC)
INCLUDE (name, description);

-- Index for common workflow status queries
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_executions_workflow_status_created
ON executions(workflow_id, status, created_at DESC);
```

---

## 4. Verification Steps

### 4.1 Verify GitHub Configuration

```bash
# Check that environments are created
gh api repos/{owner}/{repo}/environments | jq '.environments[].name'
# Expected output: staging, production

# Check branch protection
gh api repos/{owner}/{repo}/branches/main/protection | jq '.required_status_checks'
# Should list all required checks

# Test secrets are accessible (from Actions)
# Trigger a workflow manually and check it can access secrets
```

### 4.2 Verify Database Migration

```sql
-- Verify schema changes
SELECT column_name, data_type
FROM information_schema.columns
WHERE table_name = 'webhook_events'
  AND column_name IN ('retry_count', 'max_retries', 'next_retry_at', 'last_retry_at', 'retry_error', 'permanently_failed');
-- Should return 6 rows

-- Verify indexes
SELECT indexname
FROM pg_indexes
WHERE tablename = 'webhook_events'
  AND indexname LIKE '%retry%';
-- Should return idx_webhook_events_next_retry, idx_webhook_events_retry_count, idx_webhook_events_permanently_failed

-- Test retry functions
SELECT calculate_next_retry_time(1);
-- Should return a timestamp ~1 second in the future

-- Check views
SELECT * FROM webhook_retry_stats LIMIT 1;
SELECT * FROM webhook_health LIMIT 1;
```

### 4.3 Verify Performance Optimizations

```bash
# Check connection pool settings
psql -h localhost -U postgres -d gorax -c "
SELECT
    setting as max_connections,
    (SELECT count(*) FROM pg_stat_activity) as current_connections,
    (SELECT count(*) FROM pg_stat_activity WHERE state = 'idle') as idle_connections
FROM pg_settings
WHERE name = 'max_connections';
"

# Monitor connection usage
watch -n 5 'psql -h localhost -U postgres -d gorax -c "SELECT state, count(*) FROM pg_stat_activity GROUP BY state;"'
```

### 4.4 Verify CI/CD Pipeline

```bash
# Create a test PR to trigger all checks
git checkout -b test-ci-pipeline
echo "# Test" >> README.md
git add README.md
git commit -m "test: verify CI pipeline"
git push origin test-ci-pipeline

# Create PR via GitHub CLI
gh pr create --title "Test: Verify CI Pipeline" --body "Testing all CI/CD enhancements" --base dev

# Monitor workflow runs
gh run list --limit 5

# Check status checks
gh pr checks <PR-number>
```

---

## 5. Post-Deployment Monitoring

### 5.1 Monitor for SSRF Attempts

```bash
# Check logs for blocked SSRF attempts
kubectl logs deployment/gorax-api | grep "SSRF protection blocked URL"

# Or with grep on log files
grep "SSRF protection blocked URL" /var/log/gorax/api.log | tail -20
```

### 5.2 Monitor Webhook Retries

```sql
-- Check retry statistics
SELECT * FROM webhook_retry_stats;

-- Check pending retries
SELECT webhook_id, COUNT(*) as pending_count
FROM webhook_events
WHERE next_retry_at IS NOT NULL
  AND permanently_failed = false
GROUP BY webhook_id;

-- Check permanently failed events
SELECT COUNT(*) as permanently_failed_count
FROM webhook_events
WHERE permanently_failed = true;
```

### 5.3 Monitor Performance

```bash
# Check API response times
curl -w "@curl-format.txt" -o /dev/null -s http://localhost:8080/api/v1/workflows

# Where curl-format.txt contains:
# time_namelookup:  %{time_namelookup}\n
# time_connect:  %{time_connect}\n
# time_appconnect:  %{time_appconnect}\n
# time_pretransfer:  %{time_pretransfer}\n
# time_redirect:  %{time_redirect}\n
# time_starttransfer:  %{time_starttransfer}\n
# ----------\n
# time_total:  %{time_total}\n
```

---

## 6. Rollback Procedures

### 6.1 Rollback GitHub Configuration

```bash
# Remove environment
# Settings → Environments → [environment name] → Delete environment

# Remove branch protection
# Settings → Branches → [rule name] → Delete rule

# Rotate compromised secrets
# Settings → Secrets and variables → Actions → [secret name] → Update
```

### 6.2 Rollback Database Migration

```bash
# Run rollback script
psql -h localhost -U postgres -d gorax < migrations/022_webhook_retry_rollback.sql

# Verify rollback
psql -h localhost -U postgres -d gorax -c "\d webhook_events"
# retry fields should be gone
```

### 6.3 Rollback Code Changes

```bash
# Revert to previous commit
git revert fb57399  # Security fixes
git push origin dev

# Or rollback deployment
kubectl rollout undo deployment/gorax-api
```

---

## 7. Completion Checklist

Use this checklist to track completion:

### GitHub Configuration
- [ ] Staging environment created
- [ ] Production environment created with protection rules
- [ ] All required secrets added
- [ ] Branch protection applied to `main`
- [ ] Branch protection applied to `dev`
- [ ] Test PR created to verify pipeline
- [ ] All status checks passing

### Database
- [ ] Backup created before migration
- [ ] Migration 022 run successfully in staging
- [ ] Migration 022 run successfully in production
- [ ] Migration verified (columns, indexes, functions, views)
- [ ] Application deployed and working with new schema

### Performance
- [x] Connection pool settings updated (recommended config documented)
- [ ] Application restarted with new settings (pending deployment)
- [ ] Connection usage monitored (pending deployment)
- [x] N+1 query fixed - GetTopFailures optimized with LATERAL join (10.13% improvement)
- [x] New indexes added - all performance indexes deployed and verified
- [x] Formula caching implemented (93.4% improvement at 100% hit rate)
- [x] Comprehensive benchmark suite created and executed
- [x] Performance benchmark results documented

### Verification
- [ ] CI/CD pipeline tested end-to-end
- [ ] SSRF protection verified in staging
- [ ] Webhook retry tested in staging
- [x] Performance metrics baseline established (see [PERFORMANCE_BENCHMARK_RESULTS.md](PERFORMANCE_BENCHMARK_RESULTS.md))
- [ ] Monitoring dashboards updated

### Documentation
- [ ] Team trained on new CI/CD process
- [ ] Runbook updated with new procedures
- [ ] Security team notified of SSRF protection
- [ ] Database team notified of migration

---

## 8. Estimated Timeline

| Task | Estimated Time | Dependencies |
|------|----------------|--------------|
| GitHub Environments | 15 minutes | None |
| CI/CD Secrets | 15 minutes | Environments created |
| Branch Protection | 20 minutes | None |
| Database Migration (Dev) | 10 minutes | None |
| Database Migration (Staging) | 15 minutes | Dev migration successful |
| Database Migration (Prod) | 30 minutes | Staging migration successful, approval |
| Connection Pool Settings | 30 minutes | None |
| Verification | 1 hour | All above complete |

**Total Estimated Time:** 3-4 hours

---

## 9. Contact & Support

**For issues:**
- GitHub Configuration: DevOps team
- Database Migration: Database team
- Performance Issues: Backend team
- CI/CD Problems: Check `.github/workflows/CICD_GUIDE.md`

**Emergency Contacts:**
- On-call: [PagerDuty/OpsGenie]
- Team Lead: [Contact info]
- #gorax-dev Slack channel

---

**Document Version:** 1.0
**Last Updated:** 2026-01-02
**Status:** Ready for Implementation
