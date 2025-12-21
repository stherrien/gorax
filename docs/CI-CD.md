# CI/CD Pipeline Documentation

This document describes the CI/CD pipeline setup for the Gorax project using GitHub Actions.

## Overview

The Gorax project uses a comprehensive CI/CD pipeline with the following workflows:

1. **Continuous Integration (CI)** - Automated testing and building
2. **Security Scanning** - Vulnerability and security analysis
3. **Code Quality** - CodeQL analysis for Go and TypeScript
4. **Release** - Automated releases with multi-platform binaries
5. **Staging Deployment** - Automated deployment to staging environment

## Workflow Details

### 1. CI Workflow

**File:** `.github/workflows/ci.yml`

**Purpose:** Ensures code quality and functionality on every push and pull request.

**Triggers:**
- Push to any branch
- Pull requests to `main` or `dev`

**Jobs:**

#### Go Tests
- Runs on Ubuntu with PostgreSQL and Redis services
- Executes all Go tests with race detection
- Generates coverage reports (HTML and `.out` format)
- Uploads coverage to Codecov
- **Environment Variables:**
  - `DATABASE_URL`: postgres://gorax:gorax_test@localhost:5432/gorax_test?sslmode=disable
  - `REDIS_URL`: redis://localhost:6379
  - `ENV`: test

#### Go Lint
- Uses `golangci-lint` with configuration from `.golangci.yml`
- Checks code quality, security issues, and style
- Timeout: 5 minutes
- Configuration includes:
  - Cyclomatic complexity max: 15
  - Static analysis with `staticcheck`
  - Security analysis with `gosec`
  - Style checks with `revive`

#### Frontend Tests
- Runs Vitest tests with coverage
- Uses Node.js 18
- Caches npm dependencies for faster builds
- Uploads coverage to Codecov

#### Frontend Lint
- Runs ESLint with TypeScript support
- Checks React hooks rules
- Ensures code style consistency

#### Build Verification
- Builds backend binaries (`gorax` and `migrate`)
- Builds production frontend bundle
- Uploads artifacts for verification
- Only runs if all tests and lints pass

**Artifacts Generated:**
- `go-coverage`: Coverage reports
- `frontend-coverage`: Frontend coverage
- `build-artifacts`: Compiled binaries and frontend bundle

### 2. Security Workflow

**File:** `.github/workflows/security.yml`

**Purpose:** Identifies security vulnerabilities in code and dependencies.

**Triggers:**
- Pull requests to `main` or `dev`
- Push to `main` or `dev`
- Weekly schedule (Monday at 9:00 UTC)
- Manual dispatch

**Jobs:**

#### govulncheck
- Uses Go's official vulnerability scanner
- Checks all dependencies for known CVEs
- Fails if critical vulnerabilities found

#### gosec
- Static security analysis for Go code
- Generates SARIF report
- Uploads to GitHub Security tab
- Checks for:
  - SQL injection vulnerabilities
  - Hardcoded credentials
  - Weak cryptography
  - Unsafe file operations

#### npm audit
- Audits Node.js dependencies
- Fails on moderate or higher severity issues
- Generates JSON report artifact

#### Trivy Repository Scan
- Scans filesystem for vulnerabilities
- Checks dependencies and configuration files
- Reports critical and high severity issues

#### Trivy Docker Scan
- Scans Docker images for vulnerabilities
- Checks base images and installed packages
- Only runs if Dockerfile exists

#### Dependency Review
- Automated dependency change review
- Only runs on pull requests
- Comments summary in PR
- Fails on moderate or higher severity

### 3. CodeQL Workflow

**File:** `.github/workflows/codeql.yml`

**Purpose:** Advanced security and quality analysis.

**Triggers:**
- Push to `main` or `dev`
- Pull requests to `main` or `dev`
- Weekly schedule (Monday at 10:00 UTC)
- Manual dispatch

**Jobs:**

#### Analyze Go
- Uses GitHub's CodeQL engine
- Runs security-and-quality queries
- Identifies:
  - SQL injection
  - Command injection
  - Path traversal
  - Information exposure
  - Authentication issues

#### Analyze TypeScript
- Analyzes JavaScript/TypeScript code
- Checks frontend security issues
- Scans only `web/src` directory
- Excludes `node_modules` and `dist`

**Results:** Uploaded to GitHub Security tab under "Code scanning alerts"

### 4. Release Workflow

**File:** `.github/workflows/release.yml`

**Purpose:** Automates the release process with multi-platform builds.

**Triggers:**
- Version tags (e.g., `v1.0.0`, `v2.1.3`)
- Manual dispatch with version input

**Jobs:**

#### Build Binaries
- Cross-compiles for multiple platforms:
  - **Linux**: amd64, arm64
  - **macOS**: amd64, arm64
  - **Windows**: amd64
- Embeds version information:
  - Version number
  - Build timestamp
  - Git commit SHA
- Creates platform-specific archives:
  - `.tar.gz` for Linux/macOS
  - `.zip` for Windows

#### Build Frontend
- Production build with optimizations
- Minification and tree-shaking
- Creates tarball of `dist/` directory

#### Build Docker
- Multi-platform Docker images (amd64, arm64)
- Pushes to GitHub Container Registry (`ghcr.io`)
- Tags generated:
  - `v1.2.3` (semantic version)
  - `v1.2` (major.minor)
  - `v1` (major)
  - `main-abc1234` (branch-sha)
  - `latest` (on main branch)

#### Create Release
- Generates changelog from git commits
- Creates GitHub release
- Uploads all binaries and archives
- Generates SHA256 checksums

**Example Release Assets:**
```
gorax-linux-amd64.tar.gz
gorax-linux-arm64.tar.gz
gorax-darwin-amd64.tar.gz
gorax-darwin-arm64.tar.gz
gorax-windows-amd64.zip
gorax-frontend.tar.gz
SHA256SUMS.txt
```

### 5. Deploy Staging Workflow

**File:** `.github/workflows/deploy-staging.yml`

**Purpose:** Automated deployment to staging environment.

**Triggers:**
- Push to `dev` branch
- Manual dispatch

**Jobs:**

#### Build
- Builds Docker image
- Tags with:
  - `dev` (branch name)
  - `staging-abc1234` (staging-sha)
  - `staging-latest`
- Pushes to container registry

#### Deploy
- **Template for your infrastructure**
- Placeholder for deployment commands
- Common patterns:
  - **Kubernetes**: `kubectl set image ...`
  - **Docker Compose**: `docker-compose pull && up -d`
  - **AWS ECS**: Use `amazon-ecs-deploy-task-definition` action
  - **Terraform**: `terraform apply`
- Runs database migrations

#### Smoke Tests
- Waits 30 seconds for deployment
- Health check endpoints:
  - `/health` - API health
  - `/health/db` - Database connectivity
  - `/health/redis` - Redis connectivity
- Placeholder for additional tests:
  - API integration tests
  - End-to-end tests
  - Performance tests

**Environment:**
- Name: `staging`
- URL: `https://staging.gorax.dev` (configurable)

## Setup Instructions

### 1. Repository Secrets

Configure these in `Settings > Secrets and variables > Actions`:

#### Optional
- `CODECOV_TOKEN` - For Codecov uploads (can work without)
- `STAGING_URL` - Your staging environment URL

#### For Custom Deployments
- `AWS_ACCESS_KEY_ID` - If using AWS
- `AWS_SECRET_ACCESS_KEY` - If using AWS
- `KUBE_CONFIG` - If using Kubernetes
- `SSH_PRIVATE_KEY` - If using SSH deployments

### 2. Repository Settings

#### Actions Permissions
`Settings > Actions > General > Workflow permissions`
- ✅ Read and write permissions
- ✅ Allow GitHub Actions to create and approve pull requests

#### Branch Protection
`Settings > Branches > main`
- ✅ Require status checks to pass before merging
- ✅ Require branches to be up to date before merging
- Select required checks:
  - `Go Tests`
  - `Go Lint`
  - `Frontend Tests`
  - `Frontend Lint`
  - `Build Verification`

### 3. Customize Staging Deployment

Edit `.github/workflows/deploy-staging.yml`:

#### For Kubernetes:
```yaml
- name: Deploy to staging
  run: |
    echo "${{ secrets.KUBE_CONFIG }}" | base64 -d > kubeconfig
    export KUBECONFIG=./kubeconfig
    kubectl set image deployment/gorax gorax=${{ needs.build.outputs.image-tag }}
    kubectl rollout status deployment/gorax
```

#### For Docker Compose:
```yaml
- name: Deploy to staging
  uses: appleboy/ssh-action@master
  with:
    host: ${{ secrets.STAGING_HOST }}
    username: ${{ secrets.STAGING_USER }}
    key: ${{ secrets.SSH_PRIVATE_KEY }}
    script: |
      cd /app
      docker-compose pull
      docker-compose up -d
```

#### For AWS ECS:
```yaml
- name: Deploy to staging
  uses: aws-actions/amazon-ecs-deploy-task-definition@v1
  with:
    task-definition: task-definition.json
    service: gorax-staging
    cluster: gorax-cluster
```

### 4. Enable Security Features

#### Dependabot
Create `.github/dependabot.yml`:
```yaml
version: 2
updates:
  - package-ecosystem: "gomod"
    directory: "/"
    schedule:
      interval: "weekly"
  - package-ecosystem: "npm"
    directory: "/web"
    schedule:
      interval: "weekly"
  - package-ecosystem: "github-actions"
    directory: "/"
    schedule:
      interval: "weekly"
```

#### Code Scanning Alerts
`Settings > Code security and analysis`
- ✅ Dependency graph
- ✅ Dependabot alerts
- ✅ Dependabot security updates
- ✅ Code scanning (CodeQL)
- ✅ Secret scanning

## Usage

### Creating a Release

#### Option 1: Git Tag
```bash
git tag v1.0.0
git push origin v1.0.0
```

#### Option 2: Manual Dispatch
1. Go to Actions > Release
2. Click "Run workflow"
3. Enter version (e.g., `v1.0.0`)
4. Click "Run workflow"

### Deploying to Staging

Automatically triggers on push to `dev`:
```bash
git checkout dev
git merge feature-branch
git push origin dev
```

Or manually:
1. Go to Actions > Deploy to Staging
2. Click "Run workflow"
3. Select branch
4. Click "Run workflow"

### Viewing Test Coverage

1. Go to Actions > CI
2. Click on a workflow run
3. Download artifacts:
   - `go-coverage`
   - `frontend-coverage`

Or view on Codecov:
```
https://codecov.io/gh/stherrien/gorax
```

### Checking Security Issues

1. Go to Security > Code scanning
2. Review alerts by severity
3. Click on alert for details and remediation

## Best Practices

### 1. Pull Request Workflow
```bash
# Create feature branch
git checkout -b PROJ-123-feature-name

# Make changes and commit
git add .
git commit -m "Add feature"

# Push and create PR
git push origin PROJ-123-feature-name
```

CI runs automatically on PR:
- All tests must pass
- No security issues found
- Code quality checks pass

### 2. Versioning

Follow semantic versioning:
- **Major** (v2.0.0): Breaking changes
- **Minor** (v1.1.0): New features, backward compatible
- **Patch** (v1.0.1): Bug fixes

### 3. Commit Messages

Follow conventional commits:
```
feat: add webhook filtering
fix: resolve race condition in executor
docs: update API documentation
test: add integration tests for workflows
chore: update dependencies
```

### 4. Monitoring CI/CD

- Check Actions tab daily
- Set up notifications for failed workflows
- Review security alerts weekly
- Update dependencies regularly

### 5. Performance Optimization

**Caching:**
- Go modules cached by `go.sum` hash
- npm dependencies cached by `package-lock.json` hash
- Docker layers cached in GitHub Actions cache

**Concurrency:**
- Each workflow has concurrency group
- Cancels outdated runs on new push
- Saves CI minutes

**Job Dependencies:**
- Build only runs after tests pass
- Release only runs if all checks pass
- Parallel jobs where possible

## Troubleshooting

### CI Workflow Fails

**Problem:** Tests fail in CI but pass locally

**Solution:**
1. Check service health (PostgreSQL, Redis)
2. Verify environment variables
3. Run with same conditions locally:
   ```bash
   docker-compose up -d postgres redis
   export DATABASE_URL=postgres://gorax:gorax_test@localhost:5432/gorax_test?sslmode=disable
   export REDIS_URL=redis://localhost:6379
   go test -v -race ./...
   ```

### Security Workflow Issues

**Problem:** govulncheck reports vulnerabilities

**Solution:**
1. Update vulnerable package:
   ```bash
   go get -u github.com/vulnerable/package@latest
   go mod tidy
   ```
2. If no fix available, add to exclusions (temporary)

**Problem:** gosec false positives

**Solution:**
Add to `.golangci.yml`:
```yaml
linters-settings:
  gosec:
    excludes:
      - G104 # Audit errors not checked
```

### Docker Build Fails

**Problem:** Dockerfile build errors

**Solution:**
1. Test locally:
   ```bash
   docker build -t gorax:test .
   ```
2. Check `.dockerignore` isn't excluding required files
3. Verify all COPY paths exist

### Release Fails

**Problem:** Tag push doesn't trigger release

**Solution:**
1. Verify tag format: `v*` (e.g., `v1.0.0`)
2. Check workflow permissions
3. Manually trigger with workflow dispatch

### Coverage Not Uploading

**Problem:** Codecov upload fails

**Solution:**
1. Verify `CODECOV_TOKEN` is set (optional for public repos)
2. Check coverage file path
3. Ensure coverage was generated:
   ```bash
   go test -coverprofile=coverage.out ./...
   ```

## Maintenance

### Weekly Tasks
- [ ] Review security alerts
- [ ] Check failed workflow runs
- [ ] Update dependencies (Dependabot PRs)
- [ ] Review code scanning results

### Monthly Tasks
- [ ] Audit secrets and rotate if needed
- [ ] Review workflow performance and optimize
- [ ] Update GitHub Actions versions
- [ ] Check disk usage of artifacts

### Quarterly Tasks
- [ ] Review and update `.golangci.yml`
- [ ] Audit CI/CD costs and optimize
- [ ] Update documentation
- [ ] Review branch protection rules

## Resources

### Documentation
- [GitHub Actions](https://docs.github.com/en/actions)
- [golangci-lint](https://golangci-lint.run/)
- [CodeQL](https://codeql.github.com/docs/)
- [Trivy](https://aquasecurity.github.io/trivy/)
- [Codecov](https://docs.codecov.com/)

### Tools
- [act](https://github.com/nektos/act) - Run GitHub Actions locally
- [actionlint](https://github.com/rhysd/actionlint) - Lint workflow files

### Monitoring
- GitHub Actions usage: `Settings > Billing > Actions`
- Workflow runs: `Actions` tab
- Security alerts: `Security` tab

## Support

For CI/CD issues:
1. Check workflow logs in Actions tab
2. Review this documentation
3. Search GitHub Issues
4. Contact DevOps team

---

**Last Updated:** 2025-12-20
**Maintained By:** Gorax Team
