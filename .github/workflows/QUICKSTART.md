# CI/CD Quick Start Guide

## Initial Setup (5 minutes)

### 1. Enable GitHub Actions
Already enabled by default for this repository.

### 2. Set Required Secrets (Optional)
Go to `Settings > Secrets and variables > Actions`:

- `CODECOV_TOKEN` (optional) - Get from https://codecov.io
- `STAGING_URL` (optional) - Your staging server URL

### 3. Configure Branch Protection
Go to `Settings > Branches > Add rule` for `main`:

```
✅ Require status checks to pass before merging
✅ Require branches to be up to date before merging

Required checks:
  - Go Tests
  - Go Lint
  - Frontend Tests
  - Frontend Lint
```

### 4. Customize Staging Deployment (if needed)
Edit `.github/workflows/deploy-staging.yml` lines 50-60 with your deployment method.

## Testing Your Setup

### 1. Test CI Pipeline
```bash
# Create a test branch
git checkout -b test-ci

# Make a small change
echo "# Test CI" >> README.md

# Push and create PR
git add .
git commit -m "test: verify CI pipeline"
git push origin test-ci
```

Visit Actions tab to watch the workflow run.

### 2. Test Security Scan
Manually trigger: Actions > Security Scanning > Run workflow

### 3. Test Release (Dry Run)
```bash
# Create a test tag locally (don't push yet)
git tag v0.0.1-test

# Preview what would be built
git show v0.0.1-test

# Delete test tag
git tag -d v0.0.1-test
```

## Common Tasks

### Create a Release
```bash
git tag v1.0.0
git push origin v1.0.0
```

### Deploy to Staging
```bash
git checkout dev
git merge your-feature
git push origin dev
```

### View Coverage
1. Go to Actions > CI > Latest run
2. Download `go-coverage` artifact
3. Open `coverage.html` in browser

### Check Security Issues
1. Go to Security tab
2. Click "Code scanning"
3. Review any alerts

## Workflow Status Badges

Add to README:
```markdown
[![CI](https://github.com/stherrien/gorax/actions/workflows/ci.yml/badge.svg)](https://github.com/stherrien/gorax/actions/workflows/ci.yml)
[![Security](https://github.com/stherrien/gorax/actions/workflows/security.yml/badge.svg)](https://github.com/stherrien/gorax/actions/workflows/security.yml)
```

## Troubleshooting

### Workflow Not Running
- Check Actions tab for errors
- Verify workflow file syntax
- Check branch filters in workflow triggers

### Tests Failing
- Check workflow logs
- Run tests locally with same environment
- Verify database/Redis services are healthy

### Docker Build Failing
- Test locally: `docker build -t test .`
- Check Dockerfile syntax
- Verify all required files exist

## Next Steps

1. ✅ Merge this PR to enable workflows
2. ✅ Enable branch protection
3. ✅ Configure staging deployment
4. ✅ Set up Codecov (optional)
5. ✅ Create first release

## Resources

- [Full Documentation](../../../docs/CI-CD.md)
- [Workflows README](./README.md)
- [GitHub Actions Docs](https://docs.github.com/en/actions)
