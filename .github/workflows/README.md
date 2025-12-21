# GitHub Actions Workflows

This directory contains all CI/CD workflows for the Gorax project.

## Workflows Overview

### 1. CI Workflow (`ci.yml`)

**Triggers:**
- Push to any branch
- Pull requests to `main` or `dev`

**Jobs:**
- **Go Tests**: Runs Go tests with PostgreSQL and Redis, generates coverage reports
- **Go Lint**: Runs golangci-lint for code quality checks
- **Frontend Tests**: Runs Vitest tests with coverage
- **Frontend Lint**: Runs ESLint on TypeScript/React code
- **Build Verification**: Builds both backend and frontend to ensure compilation

**Features:**
- PostgreSQL and Redis services for integration tests
- Coverage reports uploaded to Codecov
- Caching of Go modules and npm dependencies
- Artifacts uploaded for coverage reports and build outputs

### 2. Security Workflow (`security.yml`)

**Triggers:**
- Pull requests to `main` or `dev`
- Push to `main` or `dev`
- Weekly schedule (Monday at 9:00 UTC)
- Manual dispatch

**Jobs:**
- **govulncheck**: Scans Go dependencies for known vulnerabilities
- **gosec**: Static security analysis for Go code
- **npm audit**: Security audit for Node.js dependencies
- **Trivy Repository Scan**: Filesystem vulnerability scanning
- **Trivy Docker Scan**: Container image vulnerability scanning
- **Dependency Review**: Reviews dependency changes in PRs

**Features:**
- SARIF file upload to GitHub Security tab
- Weekly automated security scans
- Results uploaded as artifacts

### 3. Release Workflow (`release.yml`)

**Triggers:**
- Version tags (`v*`)
- Manual dispatch with version input

**Jobs:**
- **Build Binaries**: Builds for multiple platforms (Linux, macOS, Windows) and architectures (amd64, arm64)
- **Build Frontend**: Builds production-ready frontend bundle
- **Build Docker**: Multi-platform Docker images pushed to GitHub Container Registry
- **Create Release**: Generates changelog and creates GitHub release with artifacts

**Features:**
- Cross-platform binary compilation
- Multi-arch Docker images (amd64, arm64)
- Automatic changelog generation
- SHA256 checksums for all release assets
- Docker image tags: version, major.minor, major, sha, latest

### 4. Deploy Staging Workflow (`deploy-staging.yml`)

**Triggers:**
- Push to `dev` branch
- Manual dispatch

**Jobs:**
- **Build**: Builds and pushes Docker image to registry
- **Deploy**: Deploys to staging environment (placeholder for your deployment method)
- **Smoke Tests**: Runs health checks and API tests

**Features:**
- Automated deployment to staging
- Health checks for API, database, and Redis
- Placeholder for integration with your deployment platform (Kubernetes, ECS, etc.)

**Configuration Required:**
- Set `STAGING_URL` secret in GitHub repository settings
- Configure your deployment method in the deploy job
- Add smoke test scripts as needed

### 5. CodeQL Workflow (`codeql.yml`)

**Triggers:**
- Push to `main` or `dev`
- Pull requests to `main` or `dev`
- Weekly schedule (Monday at 10:00 UTC)
- Manual dispatch

**Jobs:**
- **Analyze Go**: CodeQL security and quality analysis for Go
- **Analyze TypeScript**: CodeQL security and quality analysis for TypeScript/JavaScript

**Features:**
- Advanced security analysis with security-and-quality query suite
- Results uploaded to GitHub Security tab
- Weekly automated scans

## Required GitHub Secrets

Configure these secrets in your GitHub repository settings (`Settings > Secrets and variables > Actions`):

### Optional Secrets
- `CODECOV_TOKEN`: For uploading coverage to Codecov (optional, can work without it)
- `STAGING_URL`: URL of your staging environment (default: https://staging.gorax.dev)

### Docker Registry (GitHub Container Registry)
No additional secrets needed. The workflows use `GITHUB_TOKEN` which is automatically provided.

## Required GitHub Permissions

Ensure your GitHub Actions have the following permissions:

**Repository Settings > Actions > General > Workflow permissions:**
- Read and write permissions (for creating releases, pushing Docker images)
- Allow GitHub Actions to create and approve pull requests (optional, for automated PRs)

## Caching Strategy

All workflows use caching to speed up builds:

- **Go**: Caches `~/.cache/go-build` and `~/go/pkg/mod`
- **Node.js**: Caches npm dependencies based on `package-lock.json`
- **Docker**: Uses GitHub Actions cache for Docker layer caching

## Badge URLs

Add these badges to your README:

```markdown
[![CI](https://github.com/stherrien/gorax/actions/workflows/ci.yml/badge.svg)](https://github.com/stherrien/gorax/actions/workflows/ci.yml)
[![Security](https://github.com/stherrien/gorax/actions/workflows/security.yml/badge.svg)](https://github.com/stherrien/gorax/actions/workflows/security.yml)
[![CodeQL](https://github.com/stherrien/gorax/actions/workflows/codeql.yml/badge.svg)](https://github.com/stherrien/gorax/actions/workflows/codeql.yml)
[![codecov](https://codecov.io/gh/stherrien/gorax/branch/main/graph/badge.svg)](https://codecov.io/gh/stherrien/gorax)
```

## Customization Guide

### Adding Environment Variables

For workflows that need environment variables, add them in the workflow file:

```yaml
env:
  MY_VAR: value
  ANOTHER_VAR: ${{ secrets.MY_SECRET }}
```

### Modifying Deploy Staging

The `deploy-staging.yml` workflow is a template. Customize it for your infrastructure:

#### For Kubernetes:
```yaml
- name: Deploy to staging
  run: |
    kubectl set image deployment/gorax gorax=${{ needs.build.outputs.image-tag }}
    kubectl rollout status deployment/gorax
```

#### For Docker Compose:
```yaml
- name: Deploy to staging
  run: |
    ssh user@staging-server << 'EOF'
      cd /app
      docker-compose pull
      docker-compose up -d
    EOF
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

### Customizing Smoke Tests

Add your own smoke tests in the `smoke-tests` job:

```yaml
- name: Run API tests
  run: |
    # Postman/Newman
    newman run postman-collection.json -e staging.json

    # Or pytest
    pytest tests/smoke/ --env=staging

    # Or custom Go tests
    go test -tags=smoke ./tests/...
```

## Troubleshooting

### CI Workflow Fails
- Check that PostgreSQL and Redis services are running
- Verify DATABASE_URL and REDIS_URL are correct
- Ensure migrations run successfully

### Security Workflow Issues
- govulncheck failures: Update vulnerable dependencies
- gosec issues: Review security findings or add exclusions in `.golangci.yml`
- npm audit: Run `npm audit fix` locally

### Docker Build Failures
- Ensure a `Dockerfile` exists in the repository root
- Check that all required build files are included (not in `.dockerignore`)

### Codecov Upload Failures
- Verify `CODECOV_TOKEN` is set (or remove if using without token)
- Check coverage file paths are correct

## Best Practices

1. **Branch Protection**: Enable required status checks for CI workflow
2. **CODEOWNERS**: Add a `.github/CODEOWNERS` file for automatic review assignments
3. **Dependabot**: Enable Dependabot for automated dependency updates
4. **Secrets Rotation**: Rotate secrets regularly and use least-privilege principles
5. **Workflow Monitoring**: Regularly review workflow run times and optimize caching
6. **Cost Optimization**: Use concurrency groups to cancel outdated runs

## Additional Resources

- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [golangci-lint Configuration](../.golangci.yml)
- [Codecov Documentation](https://docs.codecov.com/)
- [Docker Build Best Practices](https://docs.docker.com/develop/dev-best-practices/)
